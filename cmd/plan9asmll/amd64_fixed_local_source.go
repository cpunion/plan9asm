package main

import (
	"context"
	"errors"
	"fmt"
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
}

func (source *amd64FixedLocalCLISource) verify(ctx context.Context) error {
	pkg, selected := source.packageData, source.selection
	if pkg == nil || selected == nil || !selected.ready || selected.types != pkg.Types ||
		selected.identity != x86CallerSelectionIdentity(pkg) || !reflect.DeepEqual(selected.sizes, x86CallerSizes(pkg.TypesSizes)) {
		return fmt.Errorf("%w: fixed-local actual-load identity changed", plan9asm.ErrProbeNeedsContext)
	}
	return source.binding.Verify(ctx)
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
	if pkg == nil || selection == nil || !selection.ready || selection.target != target ||
		selection.types != pkg.Types || pkg.IllTyped || len(pkg.Errors) != 0 || pkg.TypesSizes == nil ||
		selection.identity == "" || selection.identity != x86CallerSelectionIdentity(pkg) ||
		!reflect.DeepEqual(selection.sizes, x86CallerSizes(pkg.TypesSizes)) ||
		len(pkg.Imports) != 0 || len(selection.imports) != 0 {
		return nil, nil, fmt.Errorf("%w: fixed-local needs unchanged actual numeric ordinary package selection", plan9asm.ErrProbeNeedsContext)
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
		plan9asm.GoPackage{Path: pkg.PkgPath, Types: pkg.Types, Syntax: pkg.Syntax}, file, filepath.Base(actualAsm),
		plan9asm.Options{Goarch: goarch, TargetTriple: triple, ResolveSym: resolve, Sigs: sigs})
	if err != nil {
		return nil, nil, err
	}
	source := &amd64FixedLocalCLISource{
		observation: observed, binding: binding, selection: selection, packageData: pkg,
	}
	if err := source.verify(ctx); err != nil {
		return nil, nil, err
	}
	return bound.Sigs, source, nil
}
