package plan9asm

import (
	"fmt"
	"strings"
	"testing"
)

// Independent of lowering: all legacy names/Go aliases and both complete AVX
// tables, including native dead loads and the legacy Yml register #UD rows.
var x86NonTemporalNativeOracle = []struct {
	name, native string
	vector, load bool
}{
	{"MOVNTO", "movntdq", false, false},
	{"MOVNTDQ", "movntdq", false, false},
	{"MOVNTPD", "movntpd", false, false},
	{"MOVNTPS", "movntps", false, false},
	{"MOVNTDQA", "movntdqa", false, true},
	{"VMOVNTDQ", "vmovntdq", true, false},
	{"VMOVNTPD", "vmovntpd", true, false},
	{"VMOVNTPS", "vmovntps", true, false},
	{"VMOVNTDQA", "vmovntdqa", true, true},
}

func x86NonTemporalNativeProgram(arch, name, bank string, register int, vector, load, dead bool) (string, string) {
	word, move := 8, "MOVQ"
	if arch == "386" {
		word, move = 4, "MOVL"
	}
	reg := fmt.Sprint(bank, register)
	body := move + " p+0(FP),AX\n"
	if load {
		body += name + " (AX)," + reg + "\n"
	} else {
		zero := fmt.Sprintf("PXOR X%d,X%d\n", register, register)
		if vector {
			zero = fmt.Sprintf("VXORPS %s,%s,%s\n", reg, reg, reg)
			if bank == "Z" || register >= 16 {
				zero = fmt.Sprintf("VPXORD Z%d,Z%d,Z%d\n", register, register, register)
			}
		}
		body += zero + name + " " + reg + ",(AX)\nSFENCE\n"
	}
	if load && !dead {
		if register != 0 {
			copy := "MOVOU"
			if vector {
				copy = "VMOVUPS"
			}
			body += fmt.Sprintf("%s X%d,X0\n", copy, register)
		}
		body += "MOVL X0,CX\n"
	} else {
		body += "MOVL $7,CX\n"
	}
	body += fmt.Sprintf("MOVL CX,ret+%d(FP)\nRET\n", word)
	return fmt.Sprintf("TEXT ·Value(SB),4,$0-%d\n%s", word+4, body),
		"func Value(p *[64]byte) uint32"
}

func TestX86NonTemporalVectorNativeCompleteFamily(t *testing.T) {
	llc := findLLVM22Tool("llc")
	if llc == "" {
		t.Fatal("LLVM22 llc required")
	}
	for _, target := range x86RawReturnTargets {
		t.Run(target.triple, func(t *testing.T) {
			var source, declarations strings.Builder
			var names []string
			for _, row := range x86NonTemporalNativeOracle {
				banks := []string{"X"}
				if row.vector {
					banks = []string{"X", "Y", "Z"}
				}
				for _, bank := range banks {
					last := 7
					if target.arch == "amd64" {
						last = 15
						if row.vector {
							last = 31
						}
					}
					for _, register := range []int{0, last} {
						deadCases := []bool{false}
						if row.load {
							deadCases = append(deadCases, true)
						}
						for _, dead := range deadCases {
							body, declaration := x86NonTemporalNativeProgram(target.arch, row.name, bank,
								register, row.vector, row.load, dead)
							name := fmt.Sprint("Stream", len(names))
							names = append(names, name)
							source.WriteString(strings.ReplaceAll(body, "Value", name))
							declarations.WriteString(strings.ReplaceAll(declaration, "Value", name) + "\n")
						}
					}
				}
			}
			if len(names) != 42 {
				t.Fatal("incomplete independent non-temporal matrix", len(names))
			}
			x86LEAActualTargetGoObject(t, target.arch, target.triple, source.String())
			for _, api := range []string{"text", "module", "Go"} {
				t.Run(api, func(t *testing.T) {
					ir, err := x86ScalarFPBitViewAPI(t, source.String(), declarations.String(), target.arch, target.triple, api)
					if err != nil {
						t.Fatal("complete non-temporal program lost its native inputs", api, err)
					}
					for _, text := range []string{ir, x86DwordProductIROptimize(t, ir)} {
						x86CompactExpandRequireAllDefinitions(t, text, names)
						if strings.Count(text, "asm sideeffect") < len(names) {
							t.Fatal("native alignment/dead-access effects became optional LLVM cache hints")
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

func TestX86NonTemporalLegacyRegisterDestinationTrapsBeforeInputs(t *testing.T) {
	for _, target := range x86RawReturnTargets {
		for _, name := range []string{"MOVNTO", "MOVNTDQ", "MOVNTPD", "MOVNTPS"} {
			for _, destination := range []string{"AX", "DI", "SP"} {
				t.Run(target.triple+"/"+name+"/"+destination, func(t *testing.T) {
					source := "TEXT ·Value(SB),4,$0-0\n" + name + " X0," + destination + "\nRET\n"
					x86LEAActualTargetGoObject(t, target.arch, target.triple, source)
					for _, api := range []string{"text", "module", "Go"} {
						ir, err := x86ScalarFPBitViewAPI(t, source, "func Value()", target.arch, target.triple, api)
						if err != nil || !strings.Contains(ir, "ud2") || !strings.Contains(ir, "unreachable") {
							t.Fatal("invalid native ModRM register destination did not fault before inputs", api, err)
						}
					}
				})
			}
		}
	}
}
