package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go/types"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/xgo-dev/plan9asm"
	"github.com/xgo-dev/plan9asm/internal/gotoolprofile"
	"golang.org/x/tools/go/packages"
)

// This adapter owns one complete-file observation. It never reconstructs a
// permission from the diagnostic readback and never lends it to file chunks.
type x86SDKNumericalCLISource struct {
	numerical *plan9asm.GoX86NumericalReadSource
	selected  *x86CallerPackageSelection
	pkg       *packages.Package
	goPackage plan9asm.GoPackage
	target    string
	sources   map[string]string
	cleanup   []func() error
	original  *packages.Package
	origin    *x86CallerPackageSelection
}

type x86SDKNumericalRetryError struct{ cause error }

func (err *x86SDKNumericalRetryError) Error() string { return err.cause.Error() }
func (err *x86SDKNumericalRetryError) Unwrap() error { return err.cause }

// The independent SDK binder pins native object names, including private <>
// suffixes. Ordinary CLI $local aliases remain unchanged outside this exact
// observed complete-file route. This mapping grants no storage/read identity.
func (source *x86SDKNumericalCLISource) resolve(symbol string) string {
	symbol = strings.ReplaceAll(strings.ReplaceAll(symbol, "∕", "/"), "·", ".")
	if strings.HasPrefix(symbol, ".") {
		return source.pkg.PkgPath + symbol
	}
	return symbol
}

// Deferred source guards may join another error after translation. Such a
// failure is authoritative: retry must never hide it behind a new source load.
func x86SDKNumericalRetryCause(err error) (error, bool) {
	switch value := err.(type) {
	case *x86SDKNumericalRetryError:
		return value.cause, true
	case interface{ Unwrap() []error }:
		children := value.Unwrap()
		if len(children) == 1 {
			return x86SDKNumericalRetryCause(children[0])
		}
	case interface{ Unwrap() error }:
		return x86SDKNumericalRetryCause(value.Unwrap())
	}
	return nil, false
}

func (source *x86SDKNumericalCLISource) close() error {
	var result error
	for i := len(source.cleanup) - 1; i >= 0; i-- {
		result = errors.Join(result, source.cleanup[i]())
	}
	source.cleanup = nil
	return result
}

func (source *x86SDKNumericalCLISource) verify(ctx context.Context) error {
	if err := verifyAMD64FixedLocalSelection(source.original, source.origin, source.target); err != nil {
		return err
	}
	if err := verifyAMD64FixedLocalSelection(source.pkg, source.selected, source.target); err != nil {
		return err
	}
	for name, expected := range source.sources {
		actual, err := gotoolprofile.FileSHA256(name)
		if err != nil {
			return err
		}
		if actual != expected {
			return fmt.Errorf("%w: SDK numerical compiled Go source changed", plan9asm.ErrProbeNeedsContext)
		}
	}
	return source.numerical.Verify(ctx)
}

// The symbol query chooses whether to pay for an independent observer. It is
// not a read-shape, offset, layout, source-role or caller certificate. The
// supported SDK owner producer, rather than a package-name exemption, checks
// the complete actual internal/cpu variable and every native read separately.
func x86SDKCPUOwnerRequested(pkg *packages.Package, file *plan9asm.File, target string) bool {
	if pkg == nil || pkg.Module != nil || file == nil || file.Arch != plan9asm.ArchAMD64 ||
		(target != "linux/amd64" && target != "linux/386") {
		return false
	}
	resolve := resolveSymFunc(pkg.PkgPath)
	for _, fn := range file.Funcs {
		for _, instruction := range fn.Instrs {
			for _, operand := range instruction.Args {
				if operand.Kind != plan9asm.OpSym || !strings.HasSuffix(operand.Sym, "(SB)") {
					continue
				}
				name := strings.TrimSuffix(operand.Sym, "(SB)")
				if offset := strings.IndexAny(name, "+-"); offset >= 0 {
					name = name[:offset]
				}
				if resolve(name) == "internal/cpu.X86" {
					return true
				}
			}
		}
	}
	return false
}

func captureX86SDKNumericalSource(pkg *packages.Package, file *plan9asm.File, asm, target string,
	config compileConfig,
) (_ *plan9asm.File, _ *x86SDKNumericalCLISource, result error) {
	if !config.acquireSDKNumerical {
		return file, nil, nil
	}
	if config.Feature != nil || !x86SDKCPUOwnerRequested(pkg, file, target) {
		return nil, nil, fmt.Errorf("%w: SDK numerical retry no longer has its original candidate role", plan9asm.ErrProbeNeedsContext)
	}
	selected := config.x86CallerSelections[pkg]
	if err := verifyAMD64FixedLocalSelection(pkg, selected, target); err != nil {
		return nil, nil, err
	}
	ctx := config.Context
	if ctx == nil {
		return nil, nil, fmt.Errorf("%w: SDK numerical source requires a bounded context", plan9asm.ErrProbeNeedsContext)
	}
	if err := ctx.Err(); err != nil {
		return nil, nil, err
	}
	if _, bounded := ctx.Deadline(); !bounded {
		return nil, nil, fmt.Errorf("%w: SDK numerical source context is unbounded", plan9asm.ErrProbeNeedsContext)
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
	original, origin := pkg, selected
	pkg, selected, release, err := loadX86SDKNumericalPackage(ctx, binary, pkg, selected)
	if err != nil {
		return nil, nil, err
	}
	source := &x86SDKNumericalCLISource{selected: selected, pkg: pkg, target: target,
		original: original, origin: origin, cleanup: []func() error{release}}
	defer func() {
		if result != nil {
			result = errors.Join(result, source.close())
		}
	}()
	directory, err := filepath.EvalSymlinks(pkg.Dir)
	if err != nil {
		return nil, nil, err
	}
	actualAsm, err := filepath.EvalSymlinks(asm)
	if err != nil || filepath.Dir(actualAsm) != directory {
		return nil, nil, fmt.Errorf("%w: SDK numerical assembly origin differs: %v", plan9asm.ErrProbeNeedsContext, err)
	}
	// Deduplicate only: exec uses the last value. Never replace a loaded
	// profile with more convenient flags, cache paths or environment defaults.
	values := make(map[string]string)
	for _, entry := range selected.env {
		key, value, valid := strings.Cut(entry, "=")
		if !valid || key == "" {
			return nil, nil, fmt.Errorf("malformed SDK numerical actual-load environment")
		}
		values[key] = value
	}
	var environment []string
	for key, value := range values {
		environment = append(environment, key+"="+value)
	}
	sort.Strings(environment)
	root, stderr, err := gotoolprofile.RunBounded(ctx, directory, environment, binary, "env", "GOROOT")
	if err != nil || len(stderr) != 0 {
		return nil, nil, fmt.Errorf("observe actual SDK root: %w", errors.Join(err, numericalCLIStderr(stderr)))
	}
	rootPath, err := filepath.EvalSymlinks(strings.TrimSpace(string(root)))
	if err != nil || !filepath.IsAbs(rootPath) {
		return nil, nil, fmt.Errorf("%w: SDK numerical root is not canonical: %v", plan9asm.ErrProbeNeedsContext, err)
	}
	syntax, sources, err := amd64FixedLocalActualSyntax(ctx, pkg)
	if err != nil {
		return nil, nil, err
	}
	imports := make(map[string]*types.Package, len(pkg.Imports))
	for path, dependency := range pkg.Imports {
		imports[path] = dependency.Types
	}
	source.sources = sources
	// Preserve linkname discovery for the newly loaded package as well as the
	// complete declaration syntax required by the independent numerical binder.
	pkg.Syntax = syntax
	source.goPackage = plan9asm.GoPackage{Path: pkg.PkgPath, Types: pkg.Types, Syntax: syntax, Imports: imports}
	owner, err := plan9asm.CaptureGoNumericalStorageOwner(ctx, plan9asm.GoNumericalStorageCaptureRequest{
		GoBinary: binary, Directory: filepath.Join(rootPath, "src", "internal", "cpu"),
		PackagePath: "internal/cpu", Target: target, Environment: environment, SDKCPU: true,
		Variables: []plan9asm.GoNumericalStorageReference{{PackagePath: "internal/cpu", Name: "X86"}},
	})
	if err != nil {
		return nil, nil, err
	}
	source.cleanup = append(source.cleanup, func() error {
		owner.Close()
		return nil
	})
	request := plan9asm.GoX86NumericalReadCaptureRequest{
		GoBinary: binary, Directory: directory, PackagePath: pkg.PkgPath, AssemblyFile: filepath.Base(actualAsm),
		Target: target, Environment: environment, Owners: []*plan9asm.GoNumericalStorageOwner{owner},
	}
	if len(file.Data)+len(file.Globl) != 0 {
		data, err := plan9asm.CaptureGoX86SDKASMDataSource(ctx, plan9asm.GoX86SDKASMDataCaptureRequest{
			GoBinary: binary, Directory: directory, PackagePath: pkg.PkgPath, Target: target, Environment: environment,
		})
		if err != nil {
			return nil, nil, err
		}
		source.cleanup = append(source.cleanup, data.Close)
		literal, err := plan9asm.CaptureGoX86SDKReadonlyLiteralDataOwner(ctx, data)
		if err != nil {
			return nil, nil, err
		}
		source.cleanup = append(source.cleanup, literal.Close)
		request.ReadonlyLiterals = literal
	}
	source.numerical, err = plan9asm.CaptureGoX86SDKNumericalReadSource(ctx, request)
	if err != nil {
		return nil, nil, err
	}
	source.cleanup = append(source.cleanup, source.numerical.Close)
	raw, err := source.numerical.Readback(ctx)
	if err != nil {
		return nil, nil, err
	}
	var observed struct {
		CPP plan9asm.AssemblyCPPObservation
	}
	if err := json.Unmarshal(raw, &observed); err != nil {
		return nil, nil, err
	}
	actual, err := plan9asm.Parse(plan9asm.ArchAMD64, observed.CPP.Output)
	if err != nil {
		return nil, nil, err
	}
	if err := source.verify(ctx); err != nil {
		return nil, nil, err
	}
	return actual, source, nil
}

func numericalCLIStderr(stderr []byte) error {
	if len(stderr) == 0 {
		return nil
	}
	return fmt.Errorf("actual Go stderr: %s", strings.TrimSpace(string(stderr)))
}
