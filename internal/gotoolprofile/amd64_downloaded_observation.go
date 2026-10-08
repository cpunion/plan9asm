package gotoolprofile

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path/filepath"

	"github.com/xgo-dev/plan9asm"
	"github.com/xgo-dev/plan9asm/internal/gooriginalsrc"
	"golang.org/x/mod/module"
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
	origins *gooriginalsrc.OriginSet
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
	boundary, err := gooriginalsrc.NewDownloadBoundary(originalObservationEnvironment(original.Environment, "GOTMPDIR"),
		"amd64-original-download-", "example.invalid/plan9asm-original-download")
	if err != nil {
		return nil, err
	}
	defer func() {
		result = errors.Join(result, boundary.Close())
	}()
	// An empty directory can still inherit an ancestor module. Give the
	// download command its own module boundary, without changing source inputs.
	if err := boundary.Prepare(); err != nil {
		return nil, err
	}
	marker := boundary.Directory()
	origins, err := gooriginalsrc.NewOriginSet(cache, original.GoBinary)
	if err != nil {
		return nil, err
	}
	sources := &amd64DownloadedSources{modules: make(map[string]*amd64DownloadedModule), origins: origins}
	primary, err := sources.captureModule(ctx, original, request.SourceModule, request.SourceVersion, marker, run)
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
		source, err := sources.captureModule(ctx, original, pkg.Module.Path, pkg.Module.Version, marker, run)
		if err != nil || !source.matches(pkg.Module, false) {
			return nil, fmt.Errorf("downloaded dependency differs from its independently observed source: %v", err)
		}
	}
	if !selected {
		return nil, fmt.Errorf("downloaded graph omitted the requested original assembly package")
	}
	return sources, sources.verify()
}

func (sources *amd64DownloadedSources) captureModule(ctx context.Context, request AMD64OriginalObservationRequest, path, version, marker string, run Runner) (*amd64DownloadedModule, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if module.Check(path, version) != nil || module.CanonicalVersion(version) != version {
		return nil, fmt.Errorf("downloaded source requires a canonical exact module version")
	}
	key := path + "@" + version
	if source := sources.modules[key]; source != nil {
		return source, sources.verify()
	}
	if len(sources.modules) >= 64 {
		return nil, fmt.Errorf("downloaded source module graph exceeds its explicit bound")
	}
	if err := sources.verify(); err != nil {
		return nil, err
	}
	data, stderr, err := run(ctx, marker, request.Environment, request.GoBinary, "mod", "download", "-json", key)
	if err != nil {
		return nil, fmt.Errorf("actual exact module download failed: %w; %s", err, stderr)
	}
	if len(stderr) != 0 || len(data) == 0 || len(data) > commandOutputLimit {
		return nil, fmt.Errorf("actual exact module download returned incomplete output: %s", stderr)
	}
	var receipt gooriginalsrc.DownloadReceipt
	if err := json.Unmarshal(data, &receipt); err != nil {
		return nil, fmt.Errorf("actual module download returned malformed exact identity: %w", err)
	}
	original, err := sources.origins.Observe(ctx, path, version, receipt)
	if errors.Is(err, gooriginalsrc.ErrUnclosedModuleDeclaration) {
		return nil, fmt.Errorf("%w: %v", plan9asm.ErrProbeNeedsContext, err)
	}
	if err != nil {
		return nil, err
	}
	source := &amd64DownloadedModule{sourceModule: original.SourceModule, sourceVersion: original.SourceVersion,
		root: original.Root, goMod: original.GoMod, modulePath: original.ModulePath, goVersion: original.GoVersion}
	sources.modules[key] = source
	return source, sources.verify()
}

func (sources *amd64DownloadedSources) verify() error {
	if sources == nil {
		return fmt.Errorf("missing exact downloaded source observation")
	}
	return sources.origins.Verify()
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
