package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go/types"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"github.com/xgo-dev/plan9asm"
	"github.com/xgo-dev/plan9asm/gotoolprofile"
	"github.com/xgo-dev/plan9asm/internal/sourcepath"
	"golang.org/x/tools/go/packages"
)

// Reload under a real owned invocation, rather than changing the environment
// attached to the original packages.Load result. Both selections must describe
// the same complete source/import inventory. Only the new load's types are used
// by signatures and the independent SDK factories.
func loadX86SDKNumericalPackage(ctx context.Context, binary string, pkg *packages.Package,
	selected *x86CallerPackageSelection,
) (_ *packages.Package, _ *x86CallerPackageSelection, _ func() error, result error) {
	observed, stderr, err := gotoolprofile.RunBounded(ctx, pkg.Dir, selected.env, binary, "env", "-json",
		"GOOS", "GOARCH", "GOAMD64", "GO386", "GOFLAGS", "GOEXPERIMENT", "GOROOT", "GOCACHE")
	if err != nil || len(stderr) != 0 {
		return nil, nil, nil, errors.Join(err, numericalCLIStderr(stderr))
	}
	var actual map[string]string
	if err := json.Unmarshal(observed, &actual); err != nil {
		return nil, nil, nil, err
	}
	arch := strings.TrimPrefix(selected.target, "linux/")
	baselineCPU := arch == "amd64" && actual["GOAMD64"] == "v1" || arch == "386" && actual["GO386"] == "sse2"
	if actual["GOOS"] != "linux" || actual["GOARCH"] != arch || !baselineCPU ||
		actual["GOFLAGS"] != "" || actual["GOEXPERIMENT"] != "" {
		return nil, nil, nil, fmt.Errorf("%w: SDK numerical source requires the actual baseline x86 compiler profile", plan9asm.ErrProbeNeedsContext)
	}
	if selected.mode != nil && *selected.mode != actual["GOFLAGS"] {
		return nil, nil, nil, fmt.Errorf("%w: SDK numerical effective source flags changed after selection", plan9asm.ErrProbeNeedsContext)
	}
	cache, err := sourcepath.Resolve(actual["GOCACHE"])
	if err != nil || !filepath.IsAbs(cache) {
		return nil, nil, nil, fmt.Errorf("%w: SDK numerical build cache is not an existing absolute directory", plan9asm.ErrProbeNeedsContext)
	}
	before, err := x86SDKLoadGraph(pkg)
	if err != nil {
		return nil, nil, nil, err
	}
	inputs, err := x86SDKSelectedSourceHashes(pkg)
	if err != nil {
		return nil, nil, nil, err
	}
	workspace, err := os.MkdirTemp("", "plan9asm-sdk-numerical-load-")
	if err != nil {
		return nil, nil, nil, err
	}
	ownedDirectory := workspace
	close := func() error { return os.RemoveAll(ownedDirectory) }
	defer func() {
		if result != nil {
			result = errors.Join(result, close())
		}
	}()
	workspace, err = sourcepath.Resolve(workspace)
	if err != nil {
		return nil, nil, nil, err
	}
	values := make(map[string]string)
	for _, entry := range selected.env {
		key, value, valid := strings.Cut(entry, "=")
		if !valid || key == "" {
			return nil, nil, nil, fmt.Errorf("malformed actual SDK loading environment")
		}
		values[key] = value
	}
	for key, value := range map[string]string{
		"GOENV": "off", "GOWORK": "off", "GOTOOLCHAIN": "local", "GOPROXY": "off",
		"GOFLAGS": "", "GOEXPERIMENT": "", "CGO_ENABLED": "0", "GOPACKAGESDRIVER": "off",
		"GOOS": "linux", "GOARCH": arch, "GOAMD64": "v1", "GO386": "sse2",
		"GOROOT": actual["GOROOT"], "GOCACHE": cache,
	} {
		values[key] = value
	}
	for _, key := range []string{"GOPATH", "GOMODCACHE", "TMPDIR", "GOTMPDIR"} {
		path := filepath.Join(workspace, strings.ToLower(key))
		if err := os.Mkdir(path, 0700); err != nil {
			return nil, nil, nil, err
		}
		values[key] = path
	}
	var environment []string
	for key, value := range values {
		environment = append(environment, key+"="+value)
	}
	sort.Strings(environment)
	loaded, err := packages.Load(&packages.Config{Context: ctx, Dir: pkg.Dir, Env: environment,
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles |
			packages.NeedModule | packages.NeedDeps | packages.NeedImports | packages.NeedTypes | packages.NeedTypesSizes,
	}, ".")
	if err != nil {
		return nil, nil, nil, err
	}
	if len(loaded) != 1 || loaded[0].Module != nil || loaded[0].IllTyped || len(loaded[0].Errors) != 0 ||
		loaded[0].PkgPath != pkg.PkgPath || loaded[0].Types == nil || isTestVariantPackage(loaded[0]) {
		return nil, nil, nil, fmt.Errorf("%w: SDK numerical actual reload is not the same ordinary SDK package", plan9asm.ErrProbeNeedsContext)
	}
	reloaded := loaded[0]
	after, err := x86SDKLoadGraph(reloaded)
	if err != nil || !reflect.DeepEqual(before, after) ||
		!reflect.DeepEqual(x86CallerSizes(pkg.TypesSizes), x86CallerSizes(reloaded.TypesSizes)) {
		return nil, nil, nil, fmt.Errorf("%w: SDK numerical source/import graph differs after actual reload: %v", plan9asm.ErrProbeNeedsContext, err)
	}
	afterInputs, err := x86SDKSelectedSourceHashes(reloaded)
	if err != nil || !reflect.DeepEqual(inputs, afterInputs) {
		return nil, nil, nil, fmt.Errorf("%w: SDK numerical source bytes changed during actual reload: %v", plan9asm.ErrProbeNeedsContext, err)
	}
	imports := make(map[string]*types.Package, len(reloaded.Imports))
	for path, dependency := range reloaded.Imports {
		imports[path] = dependency.Types
	}
	snapshot := &x86CallerPackageSelection{
		identity: x86CallerSelectionIdentity(reloaded), target: selected.target, types: reloaded.Types,
		sizes: x86CallerSizes(reloaded.TypesSizes), imports: imports, env: environment,
		ready: reflect.DeepEqual(reloaded.GoFiles, reloaded.CompiledGoFiles),
	}
	return reloaded, snapshot, close, nil
}

func x86SDKLoadGraph(pkg *packages.Package) (map[string]string, error) {
	graph := make(map[string]string)
	var visit func(*packages.Package) error
	visit = func(current *packages.Package) error {
		if current == nil || current.IllTyped || len(current.Errors) != 0 || current.Types == nil || current.Module != nil {
			return fmt.Errorf("%w: SDK numerical import graph is incomplete or not standard", plan9asm.ErrProbeNeedsContext)
		}
		identity := x86CallerSelectionIdentity(current)
		if existing, seen := graph[current.PkgPath]; seen {
			if existing != identity {
				return fmt.Errorf("%w: SDK numerical import graph has conflicting package identities", plan9asm.ErrProbeNeedsContext)
			}
			return nil
		}
		if len(graph) >= 256 || identity == "" {
			return fmt.Errorf("%w: SDK numerical import graph is unbounded", plan9asm.ErrProbeNeedsContext)
		}
		graph[current.PkgPath] = identity
		for _, dependency := range current.Imports {
			if err := visit(dependency); err != nil {
				return err
			}
		}
		return nil
	}
	return graph, visit(pkg)
}

func x86SDKSelectedSourceHashes(pkg *packages.Package) (map[string]string, error) {
	files := append(append([]string(nil), pkg.CompiledGoFiles...), pkg.OtherFiles...)
	if len(files) == 0 || len(files) > 512 {
		return nil, fmt.Errorf("%w: SDK numerical selected source inventory is unbounded", plan9asm.ErrProbeNeedsContext)
	}
	result := make(map[string]string, len(files))
	for _, path := range files {
		digest, err := gotoolprofile.FileSHA256(path)
		if err != nil {
			return nil, err
		}
		result[path] = digest
	}
	return result, nil
}
