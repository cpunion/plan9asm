package plan9asm

import (
	"fmt"
	"strings"
	"testing"
)

func x86RawNumericalInstruction(t *testing.T, arch, instruction string) string {
	t.Helper()
	code := x86OrderingHintGoBytes(t, arch, "TEXT original(SB),4,$0-0\n"+instruction+"\nRET\n")
	if len(code) < 2 || code[len(code)-1] != 0xc3 {
		t.Fatal("original numerical instruction lost its exact terminal near RET", instruction)
	}
	return x86DwordProductDirectives(code[:len(code)-1])
}

func TestX86RawScalarCountCompleteFamilyNativeObjects(t *testing.T) {
	for _, target := range x86RawReturnTargets {
		t.Run(target.triple, func(t *testing.T) {
			word, move := 8, "MOVQ"
			widths := []string{"W", "L", "Q"}
			if target.arch == "386" {
				word, move = 4, "MOVL"
				widths = widths[:2] // Q has no Go mode32 encoder row.
			}
			var source, declarations strings.Builder
			count := 0
			for _, stem := range []string{"POPCNT", "LZCNT", "TZCNT"} {
				for _, width := range widths {
					for _, memory := range []bool{false, true} {
						for _, raw := range []bool{false, true} {
							name, input := fmt.Sprintf("Count%d", count), "AX"
							if memory {
								input = "(BX)"
							}
							line := stem + width + " " + input + ",CX"
							if raw {
								line = x86RawNumericalInstruction(t, target.arch, line)
							}
							fmt.Fprintf(&declarations, "func %s(value uintptr,input *[8]byte) uintptr\n", name)
							fmt.Fprintf(&source, "TEXT ·%s(SB),4,$0-%d\n%s value+0(FP),AX\n%s input+%d(FP),BX\n%s $-1,CX\n%s\n%s CX,ret+%d(FP)\nRET\n",
								name, 3*word, move, move, word, move, line, move, 2*word)
							count++
						}
					}
				}
			}
			if count != 12*len(widths) {
				t.Fatal("incomplete scalar count matrix", count)
			}
			x86RawNumericalFamilyObjects(t, source.String(), declarations.String(), target.arch, target.triple)
		})
	}
}

func TestX86RawMoveMaskCompleteFamilyNativeObjects(t *testing.T) {
	for _, target := range x86RawReturnTargets {
		t.Run(target.triple, func(t *testing.T) {
			word, move := 8, "MOVQ"
			if target.arch == "386" {
				word, move = 4, "MOVL"
			}
			var source, declarations strings.Builder
			count := 0
			for _, op := range []string{"PMOVMSKB", "MOVMSKPS", "MOVMSKPD", "VPMOVMSKB", "VMOVMSKPS", "VMOVMSKPD"} {
				banks := []string{"X"}
				if strings.HasPrefix(op, "V") {
					banks = append(banks, "Y")
				}
				for _, bank := range banks {
					for _, raw := range []bool{false, true} {
						name := fmt.Sprintf("Mask%d", count)
						line := op + " " + bank + "1,AX"
						if raw {
							line = x86RawNumericalInstruction(t, target.arch, line)
						}
						fmt.Fprintf(&declarations, "func %s(input *[32]byte) uintptr\n", name)
						fmt.Fprintf(&source, "TEXT ·%s(SB),4,$0-%d\n%s input+0(FP),BX\nVMOVDQU (BX),Y1\n%s\n%s AX,ret+%d(FP)\nRET\n",
							name, 2*word, move, line, move, word)
						count++
					}
				}
			}
			if count != 18 {
				t.Fatal("incomplete legacy/VEX X/Y move-mask matrix", count)
			}
			x86RawNumericalFamilyObjects(t, source.String(), declarations.String(), target.arch, target.triple)
		})
	}
}

func x86RawNumericalFamilyObjects(t *testing.T, source, declarations, arch, triple string) {
	t.Helper()
	llc := findLLVM22Tool("llc")
	if llc == "" {
		t.Fatal("LLVM22 llc required")
	}
	x86LEAActualTargetGoObject(t, arch, triple, source)
	for _, api := range []string{"text", "module", "Go"} {
		ir, err := x86ScalarFPBitViewAPI(t, source, declarations, arch, triple, api)
		if err != nil {
			t.Fatal("complete raw numerical family lost its paired native inputs", api, err)
		}
		for _, text := range []string{ir, x86DwordProductIROptimize(t, ir)} {
			for _, level := range []string{"-O0", "-O2"} {
				x86PackedStringCompileObject(t, llc, triple, text, level)
			}
		}
	}
}
