package gooriginalsrc

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"golang.org/x/mod/modfile"
	"golang.org/x/mod/module"
	"golang.org/x/mod/sumdb/dirhash"
)

var ErrUnclosedModuleDeclaration = errors.New("original downloaded module declaration/language is unclosed")

// DownloadReceipt is ordinary data decoded from an actual driver's stdout.
// Supplying it cannot create an object/frame token; each factory owns its command.
type DownloadReceipt struct {
	Path, Version, Dir, Zip, GoMod, Sum, GoModSum, Error string
}

// Module separates the exact archive identity from the original go.mod namespace.
// It is a value-copy observation, not a source permission or mutation interface.
type Module struct {
	SourceModule, SourceVersion string
	Root, GoMod, ModulePath     string
	GoVersion                   string
}

// OriginSet shares the original graph-wide inventory bounds across exact modules.
// Unrelated cache additions are allowed; observed parent identities are not.
type OriginSet struct {
	mu          sync.Mutex
	cache       string
	files       *Snapshot
	directories map[string]os.FileInfo
	modules     map[string]Module
	revoked     bool
}

func NewOriginSet(cache, goBinary string) (*OriginSet, error) {
	if !filepath.IsAbs(cache) || !filepath.IsAbs(goBinary) {
		return nil, fmt.Errorf("exact source guard requires bounded absolute cache/tool roles")
	}
	set := &OriginSet{cache: cache, files: NewSnapshot(),
		directories: make(map[string]os.FileInfo), modules: make(map[string]Module)}
	if err := set.retainCacheDirectory(cache); err != nil {
		return nil, err
	}
	if err := set.files.Retain(goBinary); err != nil {
		return nil, err
	}
	return set, nil
}

// Observe validates the complete original ZIP, metadata and extracted tree.
// The caller must independently execute its actual exact-download command.
func (set *OriginSet) Observe(ctx context.Context, path, version string, receipt DownloadReceipt) (_ Module, result error) {
	if set == nil || ctx == nil {
		return Module{}, fmt.Errorf("exact source observation requires a live origin set and context")
	}
	set.mu.Lock()
	defer set.mu.Unlock()
	defer func() {
		if result != nil {
			set.revoked = true
		}
	}()
	if err := ctx.Err(); err != nil {
		return Module{}, err
	}
	if err := set.verifyLocked(); err != nil {
		return Module{}, err
	}
	if module.Check(path, version) != nil || module.CanonicalVersion(version) != version {
		return Module{}, fmt.Errorf("downloaded source requires a canonical exact module version")
	}
	if receipt.Error != "" || receipt.Path != path || receipt.Version != version {
		return Module{}, fmt.Errorf("actual module download returned a different or incomplete exact identity")
	}
	key := path + "@" + version
	if _, exists := set.modules[key]; !exists && len(set.modules) >= 64 {
		return Module{}, fmt.Errorf("downloaded source module graph exceeds its explicit bound")
	}
	escapedPath, err := module.EscapePath(path)
	if err != nil {
		return Module{}, err
	}
	escapedVersion, err := module.EscapeVersion(version)
	if err != nil {
		return Module{}, err
	}
	root := filepath.Join(set.cache, escapedPath+"@"+escapedVersion)
	prefix := filepath.Join(set.cache, "cache", "download", escapedPath, "@v", escapedVersion)
	roles := []struct{ actual, expected string }{
		{receipt.Dir, root}, {receipt.Zip, prefix + ".zip"}, {receipt.GoMod, prefix + ".mod"},
	}
	for _, role := range roles {
		physical, err := filepath.EvalSymlinks(role.actual)
		if err != nil || physical != role.expected {
			return Module{}, fmt.Errorf("actual exact download escaped its canonical cache source role")
		}
		for directory := filepath.Dir(role.expected); ; directory = filepath.Dir(directory) {
			if err := set.retainCacheDirectory(directory); err != nil {
				return Module{}, err
			}
			if directory == set.cache {
				break
			}
			relative, err := filepath.Rel(set.cache, directory)
			if err != nil || !filepath.IsLocal(relative) {
				return Module{}, fmt.Errorf("downloaded cache directory role escaped its observed root")
			}
		}
	}
	if err := set.files.Directory(root, true); err != nil {
		return Module{}, err
	}
	for _, name := range []string{prefix + ".zip", prefix + ".mod"} {
		if err := set.files.Retain(name); err != nil {
			return Module{}, err
		}
	}
	zipSum, dirSum, err := set.moduleHashes(ctx, root, path, version, prefix+".zip")
	if err != nil {
		return Module{}, fmt.Errorf("original module ZIP/tree did not close: %w", err)
	}
	if zipSum != receipt.Sum {
		return Module{}, fmt.Errorf("original ZIP bytes differ from actual Go download h1")
	}
	if dirSum != zipSum {
		return Module{}, fmt.Errorf("complete original cache directory differs from original ZIP h1")
	}
	goMod := filepath.Join(root, "go.mod")
	if set.files.hashes[goMod] == "" || set.files.hashes[goMod] != set.files.hashes[prefix+".mod"] {
		return Module{}, fmt.Errorf("original ZIP go.mod differs from exact downloaded module metadata")
	}
	modData, err := set.files.ReadRetained(goMod)
	if err != nil {
		return Module{}, err
	}
	modSum, err := dirhash.Hash1([]string{"go.mod"}, func(string) (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(modData)), nil
	})
	if err != nil || modSum != receipt.GoModSum {
		return Module{}, fmt.Errorf("original module metadata differs from actual Go download h1")
	}
	parsed, err := modfile.Parse(goMod, modData, nil)
	if err != nil || parsed.Module == nil || parsed.Go == nil {
		return Module{}, ErrUnclosedModuleDeclaration
	}
	observed := Module{SourceModule: path, SourceVersion: version, Root: root,
		GoMod: prefix + ".mod", ModulePath: parsed.Module.Mod.Path, GoVersion: parsed.Go.Version}
	set.modules[key] = observed
	if err := ctx.Err(); err != nil {
		return Module{}, err
	}
	return observed, set.verifyLocked()
}

func (set *OriginSet) retainCacheDirectory(name string) error {
	if !filepath.IsAbs(name) {
		return fmt.Errorf("downloaded cache directory requires an absolute physical role")
	}
	if _, retained := set.directories[name]; !retained && len(set.directories) >= directoryLimit {
		return fmt.Errorf("downloaded cache directory inventory exceeds its count bound")
	}
	physical, err := filepath.EvalSymlinks(name)
	info, statErr := os.Lstat(name)
	if err != nil || statErr != nil || physical != name || !info.IsDir() {
		return fmt.Errorf("downloaded source cache directory is missing, redirected or over-bound")
	}
	if previous := set.directories[name]; previous != nil && !os.SameFile(previous, info) {
		return fmt.Errorf("downloaded source cache directory identity changed during capture")
	}
	set.directories[name] = info
	return nil
}

func (set *OriginSet) Verify() error {
	if set == nil {
		return fmt.Errorf("missing exact original source guard")
	}
	set.mu.Lock()
	defer set.mu.Unlock()
	err := set.verifyLocked()
	if err != nil {
		set.revoked = true
	}
	return err
}

func (set *OriginSet) verifyLocked() error {
	if set.revoked {
		return fmt.Errorf("exact original source guard is permanently revoked")
	}
	for name, expected := range set.directories {
		physical, err := filepath.EvalSymlinks(name)
		actual, statErr := os.Lstat(name)
		if err != nil || statErr != nil || physical != name || !actual.IsDir() || !os.SameFile(expected, actual) {
			return fmt.Errorf("downloaded source cache directory identity changed")
		}
	}
	return set.files.Verify()
}
