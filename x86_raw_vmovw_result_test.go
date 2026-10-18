package plan9asm

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

// Intel AVX512-FP16 section 5.61: both W values, two directions and
// register/m16 forms. GP destinations zero-extend; X writes clear wider aliases.
// https://cdrdv2-public.intel.com/678970/intel-avx512-fp16.pdf
func x86RawVMOVWResultFixture(name, arch string, load, wide, memory bool) (string, string) {
	move, word, result := "MOVQ", 8, 16
	if arch == "386" {
		move, word, result = "MOVL", 4, 8
	}
	general := 1 // CX is numerical, not the incoming pointer in AX.
	if memory {
		general = 0
	}
	source := fmt.Sprintf("TEXT %s(SB),4,$0-%d\n"+
		"%s p+0(FP),AX\nMOVL value+%d(FP),CX\nMOVL CX,X1\n",
		name, result+4, move, word)
	source += x86RawDisplacementDirectives(rawVMOVW(load, wide, 1, general, memory))
	if load || memory {
		source += "MOVL X1,BX\n"
	} else {
		source += "MOVL CX,BX\n"
	}
	source += fmt.Sprintf("MOVL BX,ret+%d(FP)\nRET\n", result)
	return source, fmt.Sprintf("func %s(p *byte,value uint32) uint32\n", name)
}

func TestX86RawVMOVWInitializedResultsFiveTargets(t *testing.T) {
	llc := findLLVM22Tool("llc")
	if llc == "" {
		t.Fatal("LLVM 22 llc required")
	}
	for _, target := range x86RawReturnTargets {
		t.Run(target.triple, func(t *testing.T) {
			var source, declarations strings.Builder
			count := 0
			for _, load := range []bool{false, true} {
				for _, wide := range []bool{false, true} {
					for _, memory := range []bool{false, true} {
						asm, declaration := x86RawVMOVWResultFixture(
							fmt.Sprintf("Value%d", count), target.arch, load, wide, memory)
						source.WriteString(asm)
						declarations.WriteString(declaration)
						count++
					}
				}
			}
			if count != 8 {
				t.Fatal("incomplete raw VMOVW format inventory", count)
			}
			x86LEAActualTargetGoObject(t, target.arch, target.triple, source.String())
			for _, api := range []string{"text", "module", "Go"} {
				t.Run(api, func(t *testing.T) {
					ir, err := x86ScalarFPBitViewAPI(t, source.String(), declarations.String(), target.arch, target.triple, api)
					if err != nil {
						t.Fatal("initialized raw GP/X/m16 result sequence", err)
					}
					for _, text := range []string{ir, x86DwordProductIROptimize(t, ir)} {
						for _, level := range []string{"-O0", "-O2"} {
							x86PackedStringCompileObject(t, llc, target.triple, text, level)
						}
					}
				})
			}
		})
	}
}

func TestX86RawVMOVWRequiresInputsWithoutResults(t *testing.T) {
	for _, target := range x86RawReturnTargets {
		for _, load := range []bool{false, true} {
			for _, wide := range []bool{false, true} {
				for _, memory := range []bool{false, true} {
					name := fmt.Sprintf("%s/load=%t/wide=%t/memory=%t", target.triple, load, wide, memory)
					t.Run(name, func(t *testing.T) {
						source := "TEXT Value(SB),4,$0-0\n" +
							x86RawDisplacementDirectives(rawVMOVW(load, wide, 1, 0, memory)) + "RET\n"
						x86LEAActualTargetGoObject(t, target.arch, target.triple, source)
						for _, api := range []string{"text", "module", "Go"} {
							ir, err := x86ScalarFPBitViewAPI(t, source, "func Value()", target.arch, target.triple, api)
							if ir != "" || !errors.Is(err, ErrProbeNeedsContext) {
								t.Fatal("raw VMOVW borrowed implementation zeros or unknown addresses", api, err)
							}
						}
					})
				}
			}
		}
	}
}

func TestX86RawVMOVWRetainsPointerAndValueBoundaries(t *testing.T) {
	for _, target := range x86RawReturnTargets {
		for _, load := range []bool{false, true} {
			source, declaration := x86RawVMOVWResultFixture("Value", target.arch, load, false, true)
			move := "MOVQ"
			if target.arch == "386" {
				move = "MOVL"
			}
			for name, bad := range map[string]string{
				"unknown pointer": strings.Replace(source, move+" p+0(FP),AX", "", 1),
				"numeric pointer": strings.Replace(source, move+" p+0(FP),AX", "MOVL $4096,AX", 1),
			} {
				t.Run(fmt.Sprintf("%s/load=%t/%s", target.triple, load, name), func(t *testing.T) {
					for _, api := range []string{"text", "module", "Go"} {
						ir, err := x86ScalarFPBitViewAPI(t, bad, declaration, target.arch, target.triple, api)
						if ir != "" || !errors.Is(err, ErrProbeNeedsContext) {
							t.Fatal("raw VMOVW invented incoming pointer identity", api, err)
						}
					}
				})
			}
		}
	}
}
