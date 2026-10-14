//go:build go1.27

package plan9asm

import (
	"errors"
	"go/types"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xgo-dev/llvm"
)

func TestGoX86PackedStringOriginalSDKIndexCompleteSourceThreeAPIs(t *testing.T) {
	ctx, request, metadata, _ := numericalSDKAssemblyFixture(t)
	request.AssemblyFile = "index_amd64.s"
	source, err := CaptureGoX86SDKNumericalReadSource(ctx, request)
	if err != nil {
		t.Fatal("actual complete SDK source/CPU/compiler precondition", err)
	}
	defer closeNumericalConsumerSource(t, source)
	pkg, _, _ := numericalSDKOriginalTypedPackage(t, metadata, request.AssemblyFile)
	sizes := types.SizesFor("gc", "amd64")
	file := source.file
	if len(file.Funcs) != 3 || len(file.Data) != 0 || len(file.Globl) != 0 {
		t.Fatal("complete original three-TEXT member changed")
	}
	resolve := func(symbol string) string { return numericalNativeName(goTextSymbolForResolution(symbol), pkg.Path) }
	sigs, err := goSigsForAsmFile(pkg, file, resolve, "amd64", sizes, sizes, nil)
	if err != nil {
		t.Fatal("complete actual SDK declarations and private graph inventory", err)
	}
	options := Options{Goarch: "amd64", TargetTriple: "x86_64-unknown-linux-gnu", ResolveSym: resolve, Sigs: sigs}
	bound, err := source.BindNumericalReads(ctx, pkg, file, options)
	if err != nil {
		t.Fatal("original CPU/read/source join", err)
	}
	original, err := os.ReadFile(filepath.Join(request.Directory, request.AssemblyFile))
	if err != nil {
		t.Fatal(err)
	}
	for _, api := range []string{"text", "module", "Go"} {
		t.Run(api, func(t *testing.T) {
			var ir string
			var result error
			if api == "text" {
				ir, result = Translate(file, bound)
			} else {
				context := llvm.NewContext()
				defer context.Dispose()
				var module llvm.Module
				if api == "module" {
					module, result = TranslateModuleInContext(context, file, bound)
				} else {
					var translated *GoModuleTranslation
					translated, result = translateGoModuleInContext(context, pkg, original, GoModuleOptions{GOOS: "linux", GOARCH: "amd64", TargetTriple: bound.TargetTriple,
						FileName: request.AssemblyFile, ResolveSym: bound.ResolveSym, NumericalReads: bound.NumericalReads})
					if translated != nil {
						module = translated.Module
					}
				}
				if !module.IsNil() {
					defer module.Dispose()
					ir = module.String()
				}
			}
			if result != nil {
				if !errors.Is(result, ErrProbeNeedsContext) || ir != "" || strings.Contains(result.Error(), "PCMPESTR") || strings.Contains(result.Error(), "PCMPISTR") {
					t.Fatal("complete original source did not advance beyond the packed-string effect", result)
				}
				t.Logf("complete unchanged SDK next Context: %v", result)
				return
			}
			if ir == "" || strings.Count(ir, "load volatile <16 x i8>") < 2 {
				t.Fatal("complete original full m128 reads missing")
			}
			t.Log("complete unchanged SDK numerical/IR source succeeded; no object or native execution claim")
		})
	}
	if err := source.Verify(ctx); err != nil {
		t.Fatal("original complete source changed after all consumers", err)
	}
}
