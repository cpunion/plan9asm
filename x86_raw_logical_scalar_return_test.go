package plan9asm

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/xgo-dev/llvm"
)

func TestX86Mode32LogicalScalarResultsComposeWithBoundZeroCleanupReturn(t *testing.T) {
	for _, target := range x86RawReturnTargets {
		if target.arch != "386" {
			continue
		}
		for _, result := range []struct {
			typ  LLVMType
			move string
		}{{I8, "MOVB $17,AX"}, {I16, "MOVW $4660,AX"}, {I32, "MOVL $2309737967,AX"}} {
			for index, encoding := range []string{"BYTE $0xc3", "BYTE $0xc2\nWORD $0"} {
				t.Run(fmt.Sprintf("%s/%s/%d", target.triple, result.typ, index), func(t *testing.T) {
					source := "TEXT value(SB),4,$0-0\n" + result.move + "\n" + encoding + "\nRET\n"
					requireX86GoAssemblerResult(t, target.arch, source, true)
					file, err := Parse(ArchAMD64, source)
					if err != nil {
						t.Fatal(err)
					}
					sig := FuncSig{Name: "value", Ret: result.typ}
					if err := CheckX86ABI0ResultInitialization(file.Funcs[0], sig, target.arch); !errors.Is(err, ErrProbeNeedsContext) {
						t.Fatal("standalone declaration query borrowed a raw return contract", err)
					}
					for _, api := range []string{"text", "module"} {
						t.Run(api, func(t *testing.T) {
							options := Options{Goarch: target.arch, TargetTriple: target.triple,
								Sigs: map[string]FuncSig{"value": sig}}
							var ir string
							if api == "text" {
								ir, err = Translate(file, options)
							} else {
								context := llvm.NewContext()
								defer context.Dispose()
								module, moduleErr := TranslateModuleInContext(context, file, options)
								err = moduleErr
								if !module.IsNil() {
									defer module.Dispose()
									ir = module.String()
								}
							}
							if err != nil || ir == "" {
								t.Fatal("complete original zero-cleanup leaf lost its logical AX result", err)
							}
							for _, optimized := range []bool{false, true} {
								text := ir
								if optimized {
									text = x86DwordProductIROptimize(t, text)
								}
								compileLLVMToObject(t, findLLVM22Tool("llc"), target.triple, "raw-logical-return.ll", "raw-logical-return.o", text)
							}
						})
					}
				})
			}
		}
	}
}

func TestX86Mode32RawLogicalReturnsCannotSupplyMissingBytesOrPhysicalEntry(t *testing.T) {
	for _, source := range []string{
		"TEXT value(SB),4,$0-0\nBYTE $0xc3\nRET\n",
		"TEXT value(SB),4,$0-0\nMOVB $17,AX\nBYTE $0xc3\nRET\n",
		"TEXT value(SB),4,$0-0\nMOVL SP,BX\nMOVL $17,AX\nBYTE $0xc3\nRET\n",
		"TEXT value(SB),4,$8-0\nMOVL $17,AX\nBYTE $0xc3\nRET\n",
		"TEXT value(SB),4,$0-0\nMOVL $17,AX\nBYTE $0xc2\nWORD $4\nRET\n",
	} {
		requireX86GoAssemblerResult(t, "386", source, true)
		file, err := Parse(ArchAMD64, source)
		if err != nil {
			t.Fatal(err)
		}
		ir, err := Translate(file, Options{Goarch: "386", TargetTriple: "i386-unknown-linux-gnu",
			Sigs: map[string]FuncSig{"value": {Name: "value", Ret: I32}}})
		if ir != "" || !errors.Is(err, ErrProbeNeedsContext) {
			t.Fatal("raw RET supplied missing data, stack or entry authority", strings.ReplaceAll(source, "\n", ";"), err)
		}
	}
}
