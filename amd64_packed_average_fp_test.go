package plan9asm

import (
	"errors"
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"strings"
	"testing"

	"github.com/xgo-dev/llvm"
)

func TestX86PackedAverageFPSpansThreeAPIsFiveObjects(t *testing.T) {
	llc := findLLVM22Tool("llc")
	if llc == "" {
		t.Fatal("LLVM22 required")
	}
	for _, triple := range []string{"x86_64-apple-darwin", "x86_64-unknown-linux-gnu", "x86_64-pc-windows-msvc", "i386-unknown-linux-gnu", "i686-pc-windows-msvc"} {
		arch := "amd64"
		if strings.HasPrefix(triple, "i") {
			arch = "386"
		}
		t.Run(triple, func(t *testing.T) {
			var source, declarations strings.Builder
			declarations.WriteString("package oracle\n")
			count := 0
			for _, op := range x86AverageNames[2:] {
				for _, bank := range []string{"X", "Y", "Z"} {
					for _, mode := range []int{0, 1, 2} {
						if arch == "386" && mode != 0 {
							continue // Actual named Go rejection is tested independently.
						}
						name := fmt.Sprintf("F%d", count)
						count++
						row := op
						if mode == 2 {
							row += ".Z"
						}
						row += " input+0(FP)," + bank + "1"
						if mode != 0 {
							row += ",K1"
						}
						row += "," + bank + "2"
						fmt.Fprintf(&source, "TEXT ·%s(SB),4,$0-64\n%s\nRET\n", name, row)
						fmt.Fprintf(&declarations, "func %s(input [64]byte)\n", name)
					}
				}
			}
			if arch == "amd64" {
				for _, op := range x86AverageNames[:2] {
					fmt.Fprintf(&source, "TEXT ·F%d(SB),4,$0-64\n%s input+0(FP),M1\nRET\n", count, op)
					fmt.Fprintf(&declarations, "func F%d(input [64]byte)\n", count)
					count++
				}
			}
			goAssemblerOracleOutput(t, arch, source.String(), "", true)
			pkg := mustGoPackage(t, "oracle", declarations.String())
			file, err := Parse(ArchAMD64, source.String())
			if err != nil {
				t.Fatal(err)
			}
			sigs := map[string]FuncSig{}
			for _, fn := range file.Funcs {
				name := strings.TrimPrefix(fn.Sym, "·")
				sig, err := DeriveGoFuncABI0Signature(name, pkg.Types.Scope().Lookup(name).(*types.Func), arch, types.SizesFor("gc", arch))
				if err != nil {
					t.Fatal(err)
				}
				sigs[name] = sig
			}
			options := Options{Goarch: arch, TargetTriple: triple, Sigs: sigs,
				ResolveSym: func(symbol string) string { return strings.TrimPrefix(symbol, "·") }}
			ir, err := Translate(file, options)
			if err != nil {
				t.Fatal(err)
			}
			x86ResultInitializationObject(t, llc, triple, ir, "average-fp-text")
			ctx := llvm.NewContext()
			defer ctx.Dispose()
			module, err := TranslateModuleInContext(ctx, file, options)
			if !module.IsNil() {
				defer module.Dispose()
			}
			if err != nil {
				t.Fatal(err)
			}
			x86ResultInitializationObject(t, llc, triple, module.String(), "average-fp-module")
			translated, err := translateGoModuleInContext(ctx, pkg, []byte(source.String()), GoModuleOptions{
				GOARCH: arch, Sizes: types.SizesFor("gc", arch), TargetTriple: triple, ResolveSym: options.ResolveSym})
			if translated != nil && !translated.Module.IsNil() {
				defer translated.Module.Dispose()
			}
			if err != nil {
				t.Fatal(err)
			}
			x86ResultInitializationObject(t, llc, triple, translated.Module.String(), "average-fp-go")
		})
	}
}

func TestX86PackedAverageFPHolesCannotBeMaskedAway(t *testing.T) {
	for _, op := range x86AverageNames[2:] {
		for index, bank := range []string{"X", "Y", "Z"} {
			width := 16 << index
			for _, start := range []int{0, 1, -1, 16} {
				for _, mode := range []string{"", ".Z"} {
					row := fmt.Sprintf("%s%s input%+d(FP),%s1,K1,%s2", op, mode, start, bank, bank)
					source := "TEXT f(SB),4,$0-8\nKMOVQ AX,K1\n" + row + "\nRET\n"
					goAssemblerOracleOutput(t, "amd64", source, "", true)
					file, err := Parse(ArchAMD64, source)
					if err != nil {
						t.Fatal(err)
					}
					for _, frame := range []FrameLayout{
						{Params: []FrameSlot{{Offset: 0, Type: I64, Index: 0, Field: -1}}},
						{Params: []FrameSlot{{Offset: 0, Type: I8, Index: 0, Field: -1}, {Offset: int64(width - 1), Type: I8, Index: 1, Field: -1}}},
					} {
						ctx, output := newAMD64CtxWithFuncForTest(t, file.Funcs[0], FuncSig{Ret: Void, Args: []LLVMType{I64, I64}, Frame: frame}, nil)
						before := output.String()
						ins := file.Funcs[0].Instrs[2]
						if _, _, err := ctx.lowerPackedUnsignedAverage(ins.Op, ins); !errors.Is(err, ErrProbeNeedsContext) || output.String() != before {
							t.Fatalf("%s incomplete logical byte span emitted IR: %v", row, err)
						}
					}
				}
			}
		}
	}
}

func TestX86PackedAverageRequiredLinuxOriginalVEXRegistration(t *testing.T) {
	const name = "amd64_packed_average_original_linux_test.go"
	for _, goos := range []string{"linux", "darwin", "windows"} {
		ctx := build.Default
		ctx.GOOS, ctx.GOARCH = goos, "amd64"
		selected, err := ctx.MatchFile(".", name)
		if err != nil || selected != (goos == "linux") {
			t.Fatal("required original AVX axis has wrong actual Go selection", goos, selected, err)
		}
	}
	file, err := parser.ParseFile(token.NewFileSet(), name, nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(file.Decls) != 2 {
		t.Fatal("required wrapper declaration changed")
	}
	fn, ok := file.Decls[1].(*ast.FuncDecl)
	if !ok || fn.Name.Name != "TestCrossLinuxRuntimeMatrixX86PackedAverageOriginalVEXAliases" || fn.Body == nil || len(fn.Body.List) != 1 {
		t.Fatal("required original AVX wrapper omitted or bypassed")
	}
	expression, ok := fn.Body.List[0].(*ast.ExprStmt)
	if !ok {
		t.Fatal("required wrapper is not an actual unconditional call")
	}
	call, ok := expression.X.(*ast.CallExpr)
	if !ok || len(call.Args) != 2 {
		t.Fatal("required wrapper lost its original Go helper")
	}
	callee, ok := call.Fun.(*ast.Ident)
	wide, wideOK := call.Args[1].(*ast.Ident)
	if !ok || callee.Name != "x86AverageOriginalGo" || !wideOK || wide.Name != "true" {
		t.Fatal("required original wide alias scope was narrowed")
	}
	workflow, err := os.ReadFile(".github/workflows/go-ci.yml")
	if err != nil || !strings.Contains(string(workflow), "-ci-selector '^(TestCrossLinuxRuntimeMatrix|") {
		t.Fatal("required Linux CI selector does not select this family", err)
	}
}
