package gotoolprofile

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/xgo-dev/plan9asm"
	"golang.org/x/mod/modfile"
	"golang.org/x/mod/module"
	"golang.org/x/mod/semver"
	"golang.org/x/mod/sumdb/dirhash"
	modzip "golang.org/x/mod/zip"
)

// AMD64DownloadedObservationRequest names an exact actual Go download, never
// an owned-main role, signature, frame or verification token. The ZIP's module
// path may differ from its original go.mod declaration (a historical mirror).
type AMD64DownloadedObservationRequest struct {
	Original                    AMD64OriginalObservationRequest
	SourceModule, SourceVersion string
}

type amd64DownloadedModule struct {
	sourceModule, sourceVersion string
	root, goMod, modulePath     string
	goVersion                   string
}

type amd64DownloadedSources struct {
	primary *amd64DownloadedModule
	modules map[string]*amd64DownloadedModule
	files   *amd64OriginalObservationFiles
}

// CaptureAMD64DownloadedPackageObservation retains the original cache bytes.
// Every nonstandard dependency gets its own exact Go-download/ZIP observation.
// Hash agreement proves byte consistency with that configured Go download, not
// a standalone SumDB/signature certificate. Its environment and actual tools
// remain part of the ordinary observation; no external initialization runs.
func CaptureAMD64DownloadedPackageObservation(ctx context.Context, request AMD64DownloadedObservationRequest, run Runner) (*AMD64OriginalPackageObservation, error) {
	if err := validateAMD64OriginalObservationRequest(ctx, request.Original, run); err != nil {
		return nil, err
	}
	sources, err := captureAMD64DownloadedSources(ctx, request, run)
	if err != nil {
		return nil, err
	}
	return captureAMD64OriginalPackageObservation(ctx, request.Original, sources, run)
}

func captureAMD64DownloadedSources(ctx context.Context, request AMD64DownloadedObservationRequest, run Runner) (_ *amd64DownloadedSources, result error) {
	original := request.Original
	cache, err := filepath.EvalSymlinks(originalObservationEnvironment(original.Environment, "GOMODCACHE"))
	if err != nil {
		return nil, err
	}
	created, err := os.MkdirTemp(originalObservationEnvironment(original.Environment, "GOTMPDIR"), "amd64-original-download-")
	if err != nil {
		return nil, err
	}
	owned, err := os.Lstat(created)
	if err != nil {
		return nil, err
	}
	defer func() {
		info, err := os.Lstat(created)
		if err != nil || !info.IsDir() || !os.SameFile(owned, info) {
			result = errors.Join(result, fmt.Errorf("owned download marker directory identity changed"))
		} else if err := os.RemoveAll(created); err != nil {
			result = errors.Join(result, err)
		}
	}()
	marker, err := filepath.EvalSymlinks(created)
	if err != nil {
		return nil, err
	}
	sources := &amd64DownloadedSources{modules: make(map[string]*amd64DownloadedModule), files: newAMD64OriginalObservationFiles()}
	if err := sources.files.retain(original.GoBinary); err != nil {
		return nil, err
	}
	primary, err := sources.captureModule(ctx, original, request.SourceModule, request.SourceVersion, marker, cache, run)
	if err != nil {
		return nil, err
	}
	sources.primary = primary
	data, stderr, err := run(ctx, original.Directory, original.Environment, original.GoBinary,
		"list", "-deps", "-compiled", "-json", "-mod=readonly", original.PackagePath)
	if err != nil || len(stderr) != 0 || len(data) == 0 || len(data) > commandOutputLimit {
		return nil, fmt.Errorf("actual downloaded source graph failed: %v; %s", err, stderr)
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	seen := make(map[string]bool)
	selected := false
	for {
		var pkg amd64OriginalListPackage
		err := decoder.Decode(&pkg)
		if err == io.EOF {
			break
		}
		if err != nil || len(seen) >= 4096 || pkg.ImportPath == "" || seen[pkg.ImportPath] {
			return nil, fmt.Errorf("invalid, repeated or over-bound downloaded source graph: %v", err)
		}
		seen[pkg.ImportPath] = true
		if pkg.ImportPath == original.PackagePath {
			selected = true
			if !primary.matches(pkg.Module, true) {
				return nil, fmt.Errorf("downloaded primary source differs from actual original package role")
			}
		}
		if pkg.Module == nil {
			if !pkg.Standard {
				return nil, fmt.Errorf("%w: downloaded dependency has no stock SDK/module role", plan9asm.ErrProbeNeedsContext)
			}
			continue
		}
		if pkg.Module.Main {
			if !primary.matches(pkg.Module, true) {
				return nil, fmt.Errorf("downloaded graph acquired a different main source")
			}
			continue
		}
		if pkg.Module.Replace != nil || pkg.Standard {
			return nil, fmt.Errorf("%w: replaced/foreign dependency needs a separate exact source role", plan9asm.ErrProbeNeedsContext)
		}
		source, err := sources.captureModule(ctx, original, pkg.Module.Path, pkg.Module.Version, marker, cache, run)
		if err != nil || !source.matches(pkg.Module, false) {
			return nil, fmt.Errorf("downloaded dependency differs from its independently observed source: %v", err)
		}
	}
	if !selected {
		return nil, fmt.Errorf("downloaded graph omitted the requested original assembly package")
	}
	return sources, sources.files.verify()
}

func (sources *amd64DownloadedSources) captureModule(ctx context.Context, request AMD64OriginalObservationRequest, path, version, marker, cache string, run Runner) (*amd64DownloadedModule, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if module.Check(path, version) != nil || semver.Canonical(version) != version {
		return nil, fmt.Errorf("downloaded source requires a canonical exact module version")
	}
	key := path + "@" + version
	if source := sources.modules[key]; source != nil {
		return source, nil
	}
	if len(sources.modules) >= 64 {
		return nil, fmt.Errorf("downloaded source module graph exceeds its explicit bound")
	}
	data, stderr, err := run(ctx, marker, request.Environment, request.GoBinary, "mod", "download", "-json", key)
	if err != nil || len(stderr) != 0 || len(data) == 0 || len(data) > commandOutputLimit {
		return nil, fmt.Errorf("actual exact module download failed: %v; %s", err, stderr)
	}
	var downloaded struct{ Path, Version, Dir, Zip, GoMod, Sum, GoModSum, Error string }
	if err := json.Unmarshal(data, &downloaded); err != nil || downloaded.Error != "" || downloaded.Path != path || downloaded.Version != version {
		return nil, fmt.Errorf("actual module download returned a different or incomplete exact identity")
	}
	escapedPath, err := module.EscapePath(path)
	if err != nil {
		return nil, err
	}
	escapedVersion, err := module.EscapeVersion(version)
	if err != nil {
		return nil, err
	}
	root := filepath.Join(cache, escapedPath+"@"+escapedVersion)
	prefix := filepath.Join(cache, "cache", "download", escapedPath, "@v", escapedVersion)
	for actual, expected := range map[string]string{downloaded.Dir: root, downloaded.Zip: prefix + ".zip", downloaded.GoMod: prefix + ".mod"} {
		physical, err := filepath.EvalSymlinks(actual)
		if err != nil || physical != expected {
			return nil, fmt.Errorf("actual exact download escaped its canonical cache source role")
		}
	}
	if err := sources.files.directory(root, true); err != nil {
		return nil, err
	}
	for _, name := range []string{prefix + ".zip", prefix + ".mod"} {
		if err := sources.files.retain(name); err != nil {
			return nil, err
		}
	}
	if _, err := modzip.CheckZip(module.Version{Path: path, Version: version}, prefix+".zip"); err != nil {
		return nil, fmt.Errorf("original module ZIP has an invalid complete member inventory: %w", err)
	}
	zipSum, err := dirhash.HashZip(prefix+".zip", dirhash.Hash1)
	if err != nil || zipSum != downloaded.Sum {
		return nil, fmt.Errorf("original ZIP bytes differ from actual Go download h1")
	}
	dirSum, err := dirhash.HashDir(root, key, dirhash.Hash1)
	if err != nil || dirSum != zipSum {
		return nil, fmt.Errorf("complete original cache directory differs from original ZIP h1")
	}
	goMod := filepath.Join(root, "go.mod")
	if sources.files.hashes[goMod] == "" || sources.files.hashes[goMod] != sources.files.hashes[prefix+".mod"] {
		return nil, fmt.Errorf("original ZIP go.mod differs from exact downloaded module metadata")
	}
	modData, err := os.ReadFile(goMod)
	if err != nil {
		return nil, err
	}
	modSum, err := dirhash.Hash1([]string{"go.mod"}, func(string) (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(modData)), nil
	})
	if err != nil || modSum != downloaded.GoModSum {
		return nil, fmt.Errorf("original module metadata differs from actual Go download h1")
	}
	parsed, err := modfile.Parse(goMod, modData, nil)
	if err != nil || parsed.Module == nil || parsed.Go == nil {
		return nil, fmt.Errorf("%w: original downloaded module declaration/language is unclosed", plan9asm.ErrProbeNeedsContext)
	}
	source := &amd64DownloadedModule{sourceModule: path, sourceVersion: version, root: root,
		goMod: prefix + ".mod", modulePath: parsed.Module.Mod.Path, goVersion: parsed.Go.Version}
	sources.modules[key] = source
	return source, sources.files.verify()
}

func (source *amd64DownloadedModule) matches(actual *amd64OriginalListModule, primary bool) bool {
	if source == nil || actual == nil || actual.Replace != nil || actual.Main != primary ||
		actual.Path != source.modulePath || actual.GoVersion != source.goVersion {
		return false
	}
	wantedVersion := source.sourceVersion
	if primary {
		wantedVersion = "" // Preserve Go's actual cached-CWD main role, not a version fiction.
	}
	root, err := filepath.EvalSymlinks(actual.Dir)
	goMod, modErr := filepath.EvalSymlinks(actual.GoMod)
	return err == nil && modErr == nil && root == source.root && actual.Version == wantedVersion &&
		(goMod == source.goMod || goMod == filepath.Join(source.root, "go.mod"))
}

func (sources *amd64DownloadedSources) module(actual *amd64OriginalListModule) (*amd64DownloadedModule, error) {
	if actual != nil && actual.Main && sources.primary.matches(actual, true) {
		return sources.primary, nil
	}
	if actual != nil && !actual.Main {
		source := sources.modules[actual.Path+"@"+actual.Version]
		if source.matches(actual, false) {
			return source, nil
		}
	}
	return nil, fmt.Errorf("%w: module source lacks its separate exact original download handoff", plan9asm.ErrProbeNeedsContext)
}
