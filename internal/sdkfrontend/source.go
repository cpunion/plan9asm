package sdkfrontend

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/xgo-dev/plan9asm"
	"github.com/xgo-dev/plan9asm/internal/sourcepath"
)

// Source owns the independently captured whole-file numerical/readonly join.
// Parsed CPP readback supplies no permission without its live numerical binder.
type Source struct {
	Numerical *plan9asm.GoX86NumericalReadSource
	cleanup   []func() error
}

func (source *Source) Close() error {
	var result error
	for i := len(source.cleanup) - 1; i >= 0; i-- {
		result = errors.Join(result, source.cleanup[i]())
	}
	source.cleanup = nil
	return result
}

func (source *Source) Verify(ctx context.Context) error { return source.Numerical.Verify(ctx) }

func Resolve(packagePath, symbol string) string {
	// The selector chooses a source entry role; LLVM's logical symbol keeps
	// the independently observed Go declaration name. Private <> is distinct.
	symbol = strings.TrimSuffix(strings.TrimSuffix(symbol, "<ABIInternal>"), "<ABI0>")
	symbol = strings.ReplaceAll(strings.ReplaceAll(symbol, "∕", "/"), "·", ".")
	if strings.HasPrefix(symbol, ".") {
		return packagePath + symbol
	}
	return symbol
}

// CPUOwnerRequested is only a cost query. The independent storage/source
// producers authenticate package role, variable layout and every native read.
func CPUOwnerRequested(pkg *Package, file *plan9asm.File, target string) bool {
	if pkg == nil || pkg.Module || file == nil || file.Arch != plan9asm.ArchAMD64 ||
		(target != "linux/amd64" && target != "linux/386") {
		return false
	}
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
				if Resolve(pkg.Path, name) == "internal/cpu.X86" {
					return true
				}
			}
		}
	}
	return false
}

type CaptureRequest struct {
	GoBinary, Root, Directory, PackagePath, AssemblyFile, Target string
	Environment                                                  []string
	Package                                                      *Package
}

func Capture(ctx context.Context, request CaptureRequest, file *plan9asm.File) (_ *plan9asm.File, _ *Source, result error) {
	directory, err := sourcepath.Resolve(request.Directory)
	if err != nil {
		return nil, nil, err
	}
	asm, err := sourcepath.Resolve(request.AssemblyFile)
	if err != nil || filepath.Dir(asm) != directory {
		return nil, nil, contextError(fmt.Sprintf("assembly origin differs: %v", err))
	}
	owners, close, err := captureOwners(ctx, request, file)
	if err != nil {
		return nil, nil, err
	}
	source := &Source{cleanup: []func() error{close}}
	defer func() {
		if result != nil {
			result = errors.Join(result, source.Close())
		}
	}()
	numerical := plan9asm.GoX86NumericalReadCaptureRequest{
		GoBinary: request.GoBinary, Directory: directory, PackagePath: request.PackagePath, AssemblyFile: filepath.Base(asm),
		Target: request.Target, Environment: request.Environment, Owners: owners,
	}
	if len(file.Data)+len(file.Globl) != 0 {
		data, err := plan9asm.CaptureGoX86SDKASMDataSource(ctx, plan9asm.GoX86SDKASMDataCaptureRequest{
			GoBinary: request.GoBinary, Directory: directory, PackagePath: request.PackagePath, Target: request.Target, Environment: request.Environment,
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
		numerical.ReadonlyLiterals = literal
	}
	source.Numerical, err = plan9asm.CaptureGoX86SDKNumericalReadSource(ctx, numerical)
	if err != nil {
		return nil, nil, err
	}
	source.cleanup = append(source.cleanup, source.Numerical.Close)
	raw, err := source.Numerical.Readback(ctx)
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
	if err := source.Verify(ctx); err != nil {
		return nil, nil, err
	}
	return actual, source, nil
}
