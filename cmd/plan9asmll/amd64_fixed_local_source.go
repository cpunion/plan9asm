package main

import (
	"context"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"github.com/xgo-dev/plan9asm"
	"github.com/xgo-dev/plan9asm/internal/gotoolprofile"
	"golang.org/x/tools/go/packages"
)

type amd64FixedLocalCLISource struct {
	observation *gotoolprofile.AMD64OriginalPackageObservation
	binding     *gotoolprofile.AMD64OriginalFixedLocalBinding
	selection   *x86CallerPackageSelection
	packageData *packages.Package
	target      string
	sources     map[string]string
}

func (source *amd64FixedLocalCLISource) verify(ctx context.Context) error {
	if err := verifyAMD64FixedLocalSelection(source.packageData, source.selection, source.target); err != nil {
		return err
	}
	for name, expected := range source.sources {
		actual, err := gotoolprofile.FileSHA256(name)
		if err != nil || actual != expected {
			return fmt.Errorf("%w: fixed-local actual compiled Go source changed", plan9asm.ErrProbeNeedsContext)
		}
	}
	return source.binding.Verify(ctx)
}

func verifyAMD64FixedLocalSelection(pkg *packages.Package, selected *x86CallerPackageSelection, target string) error {
	if pkg == nil || selected == nil || !selected.ready || selected.types != pkg.Types ||
		selected.target != target || pkg.Types == nil || pkg.TypesSizes == nil || pkg.IllTyped || len(pkg.Errors) != 0 ||
		selected.identity == "" || selected.identity != x86CallerSelectionIdentity(pkg) ||
		!reflect.DeepEqual(selected.sizes, x86CallerSizes(pkg.TypesSizes)) || len(selected.imports) != len(pkg.Imports) {
		return fmt.Errorf("%w: fixed-local actual-load identity changed", plan9asm.ErrProbeNeedsContext)
	}
	for path, dependency := range pkg.Imports {
		if dependency == nil || dependency.Types == nil || dependency.Types != selected.imports[path] {
			return fmt.Errorf("%w: fixed-local actual-load dependency types changed", plan9asm.ErrProbeNeedsContext)
		}
	}
	return nil
}

// packages.Load intentionally omits syntax for ordinary packages. Read every
// actual compiled Go file only for this candidate; the independent root factory
// typechecks those original bytes against the actual SDK export graph.
func amd64FixedLocalActualSyntax(ctx context.Context, pkg *packages.Package) ([]*ast.File, map[string]string, error) {
	if len(pkg.CompiledGoFiles) == 0 || len(pkg.CompiledGoFiles) > 256 {
		return nil, nil, fmt.Errorf("%w: fixed-local compiled Go inventory is empty or exceeds its bound", plan9asm.ErrProbeNeedsContext)
	}
	set, total := token.NewFileSet(), 0
	var syntax []*ast.File
	sources := make(map[string]string, len(pkg.CompiledGoFiles))
	for _, name := range pkg.CompiledGoFiles {
		if _, duplicate := sources[name]; duplicate {
			return nil, nil, fmt.Errorf("%w: fixed-local duplicate compiled Go source", plan9asm.ErrProbeNeedsContext)
		}
		data, err := x86CallerReadSource(ctx, name, &total)
		if err != nil {
			return nil, nil, err
		}
		file, err := parser.ParseFile(set, name, data, parser.ParseComments)
		if err != nil {
			return nil, nil, err
		}
		syntax = append(syntax, file)
		sources[name] = featureBytesSHA256(data)
	}
	return syntax, sources, nil
}

func (source *amd64FixedLocalCLISource) close() error {
	return source.observation.Close()
}

// This complete-file seam precedes splitting. The source-role query supplies
// no permission; the existing independently captured object and numeric-only
// binder must close the full source, source/normalized CFG and every access.
func prepareAMD64FixedLocalSource(pkg *packages.Package, file *plan9asm.File, asm, target, triple, goarch string,
	config compileConfig, resolve func(string) string, sigs map[string]plan9asm.FuncSig,
) (_ map[string]plan9asm.FuncSig, _ *amd64FixedLocalCLISource, result error) {
	if !plan9asm.GoAMD64FixedLocalFrameBindingNeeded(file, goarch) {
		return sigs, nil, nil
	}
	selection := config.x86CallerSelections[pkg]
	if err := verifyAMD64FixedLocalSelection(pkg, selection, target); err != nil {
		return nil, nil, err
	}
	ctx := config.Context
	if ctx == nil || ctx.Err() != nil {
		return nil, nil, fmt.Errorf("%w: fixed-local needs a bounded live compilation context", plan9asm.ErrProbeNeedsContext)
	}
	if _, bounded := ctx.Deadline(); !bounded {
		return nil, nil, fmt.Errorf("%w: fixed-local compilation context is unbounded", plan9asm.ErrProbeNeedsContext)
	}
	binary, err := exec.LookPath("go")
	if err != nil {
		return nil, nil, err
	}
	binary, err = filepath.Abs(binary)
	if err != nil {
		return nil, nil, err
	}
	binary, err = filepath.EvalSymlinks(binary)
	if err != nil {
		return nil, nil, err
	}
	dir, err := filepath.EvalSymlinks(pkg.Dir)
	if err != nil {
		return nil, nil, err
	}
	actualAsm, err := filepath.EvalSymlinks(asm)
	if err != nil || filepath.Dir(actualAsm) != dir {
		return nil, nil, fmt.Errorf("%w: fixed-local original assembly origin differs", plan9asm.ErrProbeNeedsContext)
	}
	// Execute with the actual captured effective environment. Deduplication
	// follows exec's last-value rule; no flags, CGO, SDK or cache are replaced.
	values := make(map[string]string)
	for _, item := range selection.env {
		key, value, valid := strings.Cut(item, "=")
		if !valid || key == "" {
			return nil, nil, fmt.Errorf("malformed actual fixed-local environment")
		}
		values[key] = value
	}
	var env []string
	for key, value := range values {
		env = append(env, key+"="+value)
	}
	sort.Strings(env)
	syntax, sources, err := amd64FixedLocalActualSyntax(ctx, pkg)
	if err != nil {
		return nil, nil, err
	}
	observed, err := gotoolprofile.CaptureAMD64OriginalPackageObservation(ctx,
		gotoolprofile.AMD64OriginalObservationRequest{
			GoBinary: binary, Directory: dir, PackagePath: pkg.PkgPath, Target: target, Environment: env,
		}, gotoolprofile.RunBounded)
	if err != nil {
		return nil, nil, err
	}
	defer func() {
		if result != nil {
			if err := observed.Close(); err != nil {
				result = errors.Join(result, err)
			}
		}
	}()
	bound, binding, err := observed.BindFixedLocal(ctx,
		plan9asm.GoPackage{Path: pkg.PkgPath, Types: pkg.Types, Syntax: syntax}, file, filepath.Base(actualAsm),
		plan9asm.Options{Goarch: goarch, TargetTriple: triple, ResolveSym: resolve, Sigs: sigs})
	if err != nil {
		return nil, nil, err
	}
	source := &amd64FixedLocalCLISource{
		observation: observed, binding: binding, selection: selection, packageData: pkg,
		target: target, sources: sources,
	}
	if err := source.verify(ctx); err != nil {
		return nil, nil, err
	}
	return bound.Sigs, source, nil
}
