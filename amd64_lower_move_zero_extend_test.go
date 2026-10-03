package plan9asm

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"strings"
	"testing"

	"github.com/xgo-dev/llvm"
)

func TestTranslateX86PackedZeroExtendMoveCompleteGoAssemblerForms(t *testing.T) {
	for _, target := range []struct {
		name   string
		goarch string
		triple string
	}{
		{name: "darwin-amd64", goarch: "amd64", triple: "x86_64-apple-darwin"},
		{name: "linux-amd64", goarch: "amd64", triple: "x86_64-unknown-linux-gnu"},
		{name: "windows-amd64", goarch: "amd64", triple: "x86_64-pc-windows-msvc"},
		{name: "linux-386", goarch: "386", triple: "i386-unknown-linux-gnu"},
		{name: "windows-386", goarch: "386", triple: "i686-pc-windows-msvc"},
	} {
		t.Run(target.name, func(t *testing.T) {
			var source strings.Builder
			source.WriteString("TEXT packedzeroextendmoveforms(SB),4,$0-0\n")
			for _, ratio := range x86ZeroExtendTestRatios {
				for _, prefix := range []string{"P", "VP"} {
					for _, row := range x86ZeroExtendAcceptedForms(prefix+"MOVZX"+ratio.suffix, target.goarch) {
						source.WriteString(row + "\n")
					}
				}
			}
			source.WriteString("\tRET\n")
			requireX86GoAssemblerResult(t, target.goarch, source.String(), true)

			file, err := Parse(ArchAMD64, source.String())
			if err != nil {
				t.Fatal(err)
			}
			ll, err := Translate(file, Options{
				TargetTriple: target.triple,
				Goarch:       target.goarch,
				Sigs: map[string]FuncSig{
					"packedzeroextendmoveforms": {Name: "packedzeroextendmoveforms", Ret: Void},
				},
			})
			if err != nil {
				t.Fatal(err)
			}
			llc := findLLVM22Tool("llc")
			if llc == "" {
				t.Fatal("LLVM 22 llc not found")
			}
			compileLLVMToObject(t, llc, target.triple, "packed-zero-extend-move-"+target.name+".ll", "packed-zero-extend-move-"+target.name+".o", ll)

			ctx := llvm.NewContext()
			defer ctx.Dispose()
			module, err := TranslateModuleInContext(ctx, file, Options{
				TargetTriple: target.triple,
				Goarch:       target.goarch,
				Sigs:         map[string]FuncSig{"packedzeroextendmoveforms": {Ret: Void}},
			})
			if err != nil {
				t.Fatal(err)
			}
			defer module.Dispose()
			compileLLVMToObject(t, llc, target.triple, "zero-context.ll", "zero-context.o", module.String())

			fset := token.NewFileSet()
			declaration, err := parser.ParseFile(fset, "forms.go", "package forms\nfunc packedzeroextendmoveforms()\n", 0)
			if err != nil {
				t.Fatal(err)
			}
			typed, err := (&types.Config{Sizes: types.SizesFor("gc", target.goarch)}).Check("forms", fset, []*ast.File{declaration}, nil)
			if err != nil {
				t.Fatal(err)
			}
			bound, err := translateGoModuleInContext(ctx, GoPackage{Path: "forms", Types: typed, Syntax: []*ast.File{declaration}}, []byte(source.String()), GoModuleOptions{
				GOARCH: target.goarch, TargetTriple: target.triple, FileName: "forms.s",
			})
			if err != nil {
				t.Fatal(err)
			}
			defer bound.Module.Dispose()
			compileLLVMToObject(t, llc, target.triple, "zero-go-module.ll", "zero-go-module.o", bound.Module.String())
		})
	}
}

func TestTranslateX86PackedZeroExtendMoveRejectsFormsOutsideGoAssemblerTables(t *testing.T) {
	for _, instruction := range []string{
		"PMOVZXBW.Z X0, X1",
		"PMOVZXBD X0",
		"PMOVZXBQ M0, X1",
		"PMOVZXWD Y0, X1",
		"PMOVZXWQ X0, Y1",
		"PMOVZXDQ X0, K1, X1",
		"VPMOVZXBW Y0, Y1",
		"VPMOVZXWD X0, Z1",
		"VPMOVZXDQ Z0, Z1",
		"VPMOVZXBD Y0, Z1",
		"VPMOVZXBQ X0, K0, X1",
		"VPMOVZXWQ.Z X0, X1",
		"VPMOVZXBD.BCST 8(AX), X1",
		"VPMOVZXWQ X0, (AX)",
	} {
		t.Run(strings.NewReplacer(" ", "_", "(", "_", ")", "_").Replace(instruction), func(t *testing.T) {
			assertX86PackedZeroExtendMoveRejected(t, "amd64", "x86_64-unknown-linux-gnu", instruction)
		})
	}
	for _, instruction := range []string{
		"PMOVZXBW X8, X0",
		"PMOVZXDQ X0, X8",
		"VPMOVZXBW Y20, Z8",
		"VPMOVZXBQ X20, Z8",
	} {
		t.Run("386_"+strings.NewReplacer(" ", "_").Replace(instruction), func(t *testing.T) {
			assertX86PackedZeroExtendMoveRejected(t, "386", "i386-unknown-linux-gnu", instruction)
		})
	}
}

func assertX86PackedZeroExtendMoveRejected(t *testing.T, goarch, triple, instruction string) {
	t.Helper()
	file, err := Parse(ArchAMD64, "TEXT bad(SB),NOSPLIT,$0-0\n\t"+instruction+"\n\tRET\n")
	if err != nil {
		return
	}
	if _, err := Translate(file, Options{
		TargetTriple: triple,
		Goarch:       goarch,
		Sigs:         map[string]FuncSig{"bad": {Name: "bad", Ret: Void}},
	}); err == nil {
		t.Fatalf("Translate accepted %q outside Go 1.27's packed zero-extension forms for %s", instruction, goarch)
	}
}
