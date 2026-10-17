package plan9asm

import (
	"fmt"
	"strings"
	"testing"
)

func TestX86BEXTRTypedByteEffectExactInputs(t *testing.T) {
	for _, arch := range []string{"386", "amd64"} {
		for _, op := range []string{"BEXTRL", "BEXTRQ"} {
			t.Run(arch+"/"+op, func(t *testing.T) {
				width := int64(4)
				if arch == "amd64" && op == "BEXTRQ" {
					width = 8
				}
				ins := x86ScalarLaneInstruction(t, op+" AX,BX,CX")
				effect, known, err := x86TypedGPByteEffect(ins.Op, ins, arch)
				if err != nil || !known || len(effect.operands) != 3 {
					t.Fatal("complete BEXTR family lacks its typed input/result effect", known, err)
				}
				for ordinal, operand := range effect.operands {
					bytes := width
					if ordinal == 0 {
						bytes = 2 // Only START and LEN, not the rest of the GP.
					}
					if operand.index != ordinal || operand.operand != ins.Args[ordinal] ||
						operand.width != bytes || operand.read != (ordinal < 2) || operand.write != (ordinal == 2) {
						t.Fatal("BEXTR changed an input, output or physical mode width", ordinal, operand)
					}
				}
				if effect.flagsRead != 0 || effect.flagsDefined != x86ByteFlagCF|x86ByteFlagOF|x86ByteFlagZF ||
					effect.flagsKilled != x86ByteFlagsAll || effect.dynamicMemory || !effect.requiresResultProof {
					t.Fatal("BEXTR invented preserved/defined flags or waived its input obligation", effect)
				}
			})
		}
	}
}

func TestX86BEXTRCompleteInitializedSourceObjects(t *testing.T) {
	llc := findLLVM22Tool("llc")
	if llc == "" {
		t.Fatal("complete BEXTR objects require LLVM22 llc")
	}
	for _, target := range x86RawReturnTargets {
		for _, op := range []string{"BEXTRL", "BEXTRQ"} {
			for _, native := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/%s/native%t", target.triple, op, native), func(t *testing.T) {
					width, typ, move := 4, "uint32", "MOVL"
					if target.arch == "amd64" && op == "BEXTRQ" {
						width, typ, move = 8, "uint64", "MOVQ"
					}
					offset, result, input, load := width, 2*width, "BX", move
					parameter := typ
					if native {
						word := int(goWordSize(target.arch))
						offset, result, input = word, 2*word, "(BX)"
						parameter = "*" + typ
						if word == 8 {
							load = "MOVQ"
						} else {
							load = "MOVL"
						}
					}
					declaration := fmt.Sprintf("func Value(control uint16, source %s) %s", parameter, typ)
					body := fmt.Sprintf("MOVW control+0(FP),AX\n%s source+%d(FP),BX\n%s AX,%s,CX\n%s CX,ret0+%d(FP)\nRET\n",
						load, offset, op, input, move, result)
					_, _, _, source := x86ResultInitializationFixture(t, target.arch, declaration, body)
					x86LEAActualTargetGoObject(t, target.arch, target.triple, source)
					for _, api := range []string{"text", "module", "Go"} {
						t.Run(api, func(t *testing.T) {
							ir, err := x86ScalarFPBitViewAPI(t, source, declaration, target.arch, target.triple, api)
							if err != nil || ir == "" {
								t.Fatal("initialized complete BEXTR source remains unclosed", err)
							}
							for _, text := range []string{ir, x86MulDivRoundOptimize(t, ir)} {
								if native && !strings.Contains(text, fmt.Sprintf("load volatile i%d, ptr", width*8)) {
									t.Fatal("BEXTR lost its exact observable native read, including mode32 Q", width)
								}
								for _, level := range []string{"-O0", "-O2"} {
									x86PackedStringCompileObject(t, llc, target.triple, text, level)
								}
							}
						})
					}
				})
			}
		}
	}
}

func TestX86BEXTRVoidCannotInventInputs(t *testing.T) {
	for _, target := range x86RawReturnTargets {
		for _, op := range []string{"BEXTRL", "BEXTRQ"} {
			for name, prefix := range map[string]string{
				"missing START and LEN": "MOVL $7,BX\n",
				"missing LEN":           "MOVB $4,AL\nMOVL $7,BX\n",
				"missing source":        "MOVW $0x0804,AX\n",
			} {
				t.Run(target.triple+"/"+op+"/"+name, func(t *testing.T) {
					source := "TEXT f(SB),4,$0-0\n" + prefix + op + " AX,BX,CX\nRET\n"
					x86UnboundVoidFormatSource(t, source, "f", target.arch, target.triple)
				})
			}
		}
	}
}
