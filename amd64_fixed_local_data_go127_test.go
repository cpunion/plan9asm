//go:build go1.27

package plan9asm

import (
	"go/types"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xgo-dev/llvm"
)

const fixedLocalLiteralData = "DATA literal<>+0(SB)/8,$0x0102030405060708\n" +
	"DATA literal<>+8(SB)/8,$0x1112131415161718\nGLOBL literal<>(SB),8,$32\n"

func fixedLocalDataFixture(t *testing.T, operation, data string) (GoPackage, GoAMD64FrameCaptureRequest, *File, Options, string) {
	t.Helper()
	pkg, request, _, options := fixedLocalFrameFixture(t, operation)
	name := filepath.Join(request.Directory, request.AssemblyFile)
	original, err := os.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	assembly := data + string(original)
	if err := os.WriteFile(name, []byte(assembly), 0600); err != nil {
		t.Fatal(err)
	}
	file, err := Parse(ArchAMD64, assembly)
	if err != nil {
		t.Fatal(err)
	}
	return pkg, request, file, options, assembly
}

func TestGoAMD64FixedLocalReadonlyDataThreeTargetObjects(t *testing.T) {
	t.Setenv("GOCACHE", t.TempDir())
	llc := findLLVM22Tool("llc")
	if llc == "" {
		t.Fatal("LLVM 22 llc required")
	}
	for _, target := range []struct{ name, triple string }{
		{"linux/amd64", "x86_64-unknown-linux-gnu"},
		{"darwin/amd64", "x86_64-apple-darwin"},
		{"windows/amd64", "x86_64-pc-windows-msvc"},
	} {
		t.Run(target.name, func(t *testing.T) {
			pkg, request, file, options, assembly := fixedLocalDataFixture(t,
				"MOVOU literal<>(SB),X0\nPADDQ 16(SP),X0", fixedLocalLiteralData)
			request.Target, options.TargetTriple = target.name, target.triple
			x86LEAActualTargetGoObject(t, "amd64", target.triple, assembly)
			source := requireGoAMD64FrameCapture(t, pkg, request)
			ctx, cancel := runtimeOracleStageContext(t)
			defer cancel()
			bound, err := source.BindFixedLocal(ctx, pkg, file, options)
			if err != nil {
				t.Fatal("original private readonly DATA cannot join the fixed-local numeric proof", err)
			}
			fixedLocalDataAPIs(t, pkg, file, assembly, bound, func(t *testing.T, ir string, err error) {
				if err != nil || !strings.Contains(ir, "internal constant [32 x i8]") || !strings.Contains(ir, "fixed_local_native") {
					t.Fatal("missing joined private DATA/frame translation", err)
				}
				compileLLVMToObject(t, llc, target.triple, "data-frame.ll", "data-frame.o", ir)
				compileLLVMToObject(t, llc, target.triple, "data-frame-opt.ll", "data-frame-opt.o", x86MulDivRoundOptimize(t, ir))
			})
		})
	}
}

func fixedLocalDataAPIs(t *testing.T, pkg GoPackage, file *File, assembly string, options Options, check func(*testing.T, string, error)) {
	t.Helper()
	t.Run("Translate", func(t *testing.T) {
		ir, err := Translate(file, options)
		check(t, ir, err)
	})
	t.Run("Module", func(t *testing.T) {
		context := llvm.NewContext()
		defer context.Dispose()
		module, err := TranslateModuleInContext(context, file, options)
		ir := ""
		if !module.IsNil() {
			defer module.Dispose()
			ir = module.String()
		}
		check(t, ir, err)
	})
	t.Run("GoModule", func(t *testing.T) {
		goos := "linux"
		if strings.Contains(options.TargetTriple, "apple") {
			goos = "darwin"
		} else if strings.Contains(options.TargetTriple, "windows") {
			goos = "windows"
		}
		translation, err := TranslateGoModule(pkg, []byte(assembly), GoModuleOptions{
			GOOS: goos, GOARCH: "amd64", TargetTriple: options.TargetTriple,
			Sizes: types.SizesFor("gc", "amd64"), ResolveSym: options.ResolveSym,
			ManualSig: func(name string) (FuncSig, bool) { sig, found := options.Sigs[name]; return sig, found },
		})
		ir := ""
		if translation != nil && !translation.Module.IsNil() {
			defer translation.Module.Dispose()
			ir = translation.Module.String()
		}
		check(t, ir, err)
	})
}
