//go:build go1.27

package plan9asm

import (
	"crypto/sha256"
	"errors"
	"go/types"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xgo-dev/llvm"
)

func amd64SDKArgumentTranslate(t *testing.T, api string, pkg GoPackage, file *File, options Options) (string, error) {
	t.Helper()
	if api == "text" {
		return Translate(file, options)
	}
	context := llvm.NewContext()
	defer context.Dispose()
	var module llvm.Module
	var err error
	if api == "module" {
		module, err = TranslateModuleInContext(context, file, options)
	} else {
		source := options.NumericalReads.source
		original, failure := os.ReadFile(filepath.Join(source.request.Directory, source.request.AssemblyFile))
		if failure != nil {
			t.Fatal(failure)
		}
		goos, _, _ := strings.Cut(source.request.Target, "/")
		result, failure := translateGoModuleInContext(context, pkg, original, GoModuleOptions{
			GOARCH: options.Goarch, GOOS: goos, TargetTriple: options.TargetTriple, FileName: source.request.AssemblyFile,
			NumericalReads: options.NumericalReads, ResolveSym: options.ResolveSym,
		})
		err = failure
		if result != nil {
			module = result.Module
		}
	}
	if !module.IsNil() {
		defer module.Dispose()
		if err == nil {
			return module.String(), nil
		}
	}
	return "", err
}

func amd64SDKTypedArgumentsWholeSource(t *testing.T, member string) {
	t.Helper()
	ctx, request, metadata, _ := numericalSDKAssemblyFixture(t)
	request.AssemblyFile = member
	original, err := os.ReadFile(filepath.Join(request.Directory, member))
	if err != nil {
		t.Fatal(err)
	}
	source, err := CaptureGoX86SDKNumericalReadSource(ctx, request)
	if err != nil {
		t.Fatal("unchanged complete original SDK source/compiler/CPU join", err)
	}
	defer closeNumericalConsumerSource(t, source)
	pkg, file, options := numericalSDKOriginalTypedPackage(t, metadata, member)
	if len(file.Funcs) != 3 || len(file.Data)+len(file.Globl) != 0 || len(source.reads) != 1 ||
		len(metadata.observation.Objects.Sources) != len(metadata.guard.selection.SFiles) {
		t.Fatal("complete original three TEXT/all selected ASM/CPU inventory changed")
	}
	sizes := types.SizesFor("gc", "amd64")
	options.Sigs, err = goSigsForAsmFile(pkg, file, options.ResolveSym, "amd64", sizes, sizes, nil)
	if err != nil || len(options.Sigs) != 2 {
		t.Fatal("actual full original declaration inventory", err)
	}
	before, err := x86CallerSourceFingerprint(file)
	if err != nil {
		t.Fatal(err)
	}
	bound, err := source.BindNumericalReads(ctx, pkg, file, options)
	if err != nil {
		t.Fatal("complete original typed carrier/compiler/source binding", err)
	}
	for _, api := range []string{"text", "module", "go-module"} {
		t.Run(api, func(t *testing.T) {
			ir, err := amd64SDKArgumentTranslate(t, api, pkg, file, bound)
			if member == "equal_amd64.s" {
				if ir != "" || !errors.Is(err, ErrProbeNeedsContext) || !strings.Contains(strings.ReplaceAll(err.Error(), " ", ""), "8(DX)") {
					t.Fatal("whole equal source must retain its separate hidden closure extent Context", err)
				}
				t.Log("complete original equal source remains Context (NOT file pass):", err)
				return
			}
			if err != nil || ir == "" {
				t.Fatal("whole original Compare source still needs an independent contract", err)
			}
			if !strings.Contains(ir, "define i64 "+llvmGlobal("internal/bytealg.Compare")) ||
				!strings.Contains(ir, "define i64 "+llvmGlobal("runtime.cmpstring")) ||
				!strings.Contains(ir, "{ ptr, i64, i64 }") || !strings.Contains(ir, "{ ptr, i64 }") {
				t.Fatal("complete two original declared roots/grouped string+slice transport were replaced", ir)
			}
		})
	}
	after, err := x86CallerSourceFingerprint(file)
	if err != nil || after != before {
		t.Fatal("consumer changed the original source", err)
	}
	if err := source.Verify(ctx); err != nil {
		t.Fatal("final original numerical source Verify", err)
	}
	if err := metadata.Verify(ctx); err != nil {
		t.Fatal("final original complete header/archive Verify", err)
	}
	afterBytes, err := os.ReadFile(filepath.Join(request.Directory, member))
	if err != nil || sha256.Sum256(afterBytes) != sha256.Sum256(original) {
		t.Fatal("original SDK bytes changed", err)
	}
	t.Logf("member=%s completeTEXTs=3 declaredEntries=2 allSelectedASM=%d sourceSHA=%x; no object/native/closure/stack proof", member, len(metadata.guard.selection.SFiles), sha256.Sum256(original))
}

func TestGoX86SDKAMD64CompareTypedArgumentsWholeSource(t *testing.T) {
	amd64SDKTypedArgumentsWholeSource(t, "compare_amd64.s")
}

func TestGoX86SDKAMD64EqualTypedArgumentsWholeSource(t *testing.T) {
	amd64SDKTypedArgumentsWholeSource(t, "equal_amd64.s")
}
