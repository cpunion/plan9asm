package plan9asm

import (
	"fmt"
	"testing"
)

func x86OrdinaryMoveFPAddressFixture(arch, width, address string) (string, string) {
	word, move := goWordSize(arch), "MOVQ"
	if word == 4 {
		move = "MOVL"
	}
	return fmt.Sprintf("TEXT ·FPAddress(SB),4,$0-%d\n%s input+0(FP),AX\nMOV%s %s,BX\n"+
			"%s $0,BX\n%s AX,ret+%d(FP)\nRET\n", 2*word, move, width, address, move, move, word),
		"func FPAddress(input uintptr) uintptr\n"
}

func TestX86OrdinaryMoveNamedFPAddressObjects(t *testing.T) {
	llc := findLLVM22Tool("llc")
	if llc == "" {
		t.Fatal("LLVM22 llc required")
	}
	observations := 0
	for _, target := range x86RawReturnTargets {
		for _, width := range []string{"W", "L", "Q"} {
			if target.arch == "386" && width == "Q" {
				continue // Independently rejected by the grammar owner below.
			}
			t.Run(target.triple+"/"+width, func(t *testing.T) {
				source, declaration := x86OrdinaryMoveFPAddressFixture(target.arch, width, "$input+0(FP)")
				x86LEAActualTargetGoObject(t, target.arch, target.triple, source)
				for _, api := range []string{"text", "module", "Go"} {
					ir, err := x86ScalarFPBitViewAPI(t, source, declaration, target.arch, target.triple, api)
					if err != nil || ir == "" {
						t.Fatal("dead address of an independently typed scalar field rejected", api, err)
					}
					for _, level := range []string{"-O0", "-O2"} {
						x86PackedStringCompileObject(t, llc, target.triple, ir, level)
					}
				}
			})
			observations++
		}
	}
	if observations != 13 {
		t.Fatal("complete mode/target FP-address object partition changed", observations)
	}
}

func TestX86OrdinaryMoveNamedFPAddressSourceBoundaries(t *testing.T) {
	for _, arch := range []string{"amd64", "386"} {
		for _, width := range []string{"W", "L", "Q"} {
			for _, control := range []string{"numeric-FP", "outside-field", "changed-Raw", "changed-operand", "physical-origin"} {
				t.Run(arch+"/"+width+"/"+control, func(t *testing.T) {
					address := "$input+0(FP)"
					if control == "numeric-FP" {
						address = "$0(FP)"
					} else if control == "outside-field" {
						address = "$input+32(FP)"
					}
					source, declaration := x86OrdinaryMoveFPAddressFixture(arch, width, address)
					accepted := control != "numeric-FP" && !(arch == "386" && width == "Q")
					goAssemblerOracleOutput(t, arch, source, "", accepted)
					ins := x86ScalarLaneInstruction(t, "MOV"+width+" "+address+",BX")
					original := fmt.Sprintf("%#v", ins)
					switch control {
					case "changed-Raw":
						ins.Raw = "MOV" + width + " $input+8(FP),BX"
					case "changed-operand":
						ins.Args[0].FPOffset = 8
					case "physical-origin":
						ins.x86Encoded = true
					}
					view, known, err := x86NamedFrameMoveAddressInstruction(ins, arch)
					if control != "outside-field" && known && err == nil {
						t.Error("invalid/changed FP spelling or physical role became a named address", view)
					}
					if control == "numeric-FP" || control == "outside-field" || arch == "386" && width == "Q" {
						for _, api := range []string{"text", "module", "Go"} {
							if ir, err := x86ScalarFPBitViewAPI(t, source, declaration, arch, x86AddressMemoryTargets(arch)[0], api); err == nil || ir != "" {
								t.Fatal("invalid/unmapped/mode32-Q FP address produced public IR", api, ir, err)
							}
						}
					}
					if control == "outside-field" && original != fmt.Sprintf("%#v", ins) {
						t.Fatal("source-only address projection mutated original instruction")
					}
				})
			}
		}
	}
}
