package plan9asm

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

// Independent syntax axes from Go's yxm, _yvaddsd and _yvcvtdq2ps tables.
// Scalar VEX/EVEX has a separate upper-lane source; packed sqrt is unary.
func x86SqrtInputRows(arch string) []string {
	var rows []string
	for _, opcode := range []string{"SQRTSS", "SQRTSD", "SQRTPS", "SQRTPD", "VSQRTSS", "VSQRTSD", "VSQRTPS", "VSQRTPD"} {
		vector := strings.HasPrefix(opcode, "V")
		scalar := strings.HasSuffix(opcode, "SS") || strings.HasSuffix(opcode, "SD")
		for _, bank := range []string{"X", "Y", "Z"} {
			if (!vector || scalar) && bank != "X" {
				continue
			}
			for _, memory := range []bool{false, true} {
				for _, control := range []string{"", "BCST", "RN_SAE", "RD_SAE", "RU_SAE", "RZ_SAE"} {
					if control != "" && !vector || control == "BCST" && (!memory || scalar) ||
						strings.HasSuffix(control, "SAE") && (memory || !scalar && bank != "Z") {
						continue
					}
					for _, mode := range []string{"plain", "merge", "zero"} {
						if mode != "plain" && (!vector || scalar && arch == "386") {
							continue
						}
						name, first := opcode, bank+"0"
						if control != "" {
							name += "." + control
						}
						if mode == "zero" {
							name += ".Z"
						}
						if memory {
							first = "(AX)"
						}
						args := first
						if vector && scalar {
							args += ",X1"
						}
						if mode != "plain" {
							args += ",K1"
						}
						rows = append(rows, name+" "+args+","+bank+"2")
					}
				}
			}
		}
	}
	return rows
}

func TestX86SqrtUnknownVoidInputs(t *testing.T) {
	for _, arch := range []string{"386", "amd64"} {
		triple := "x86_64-unknown-linux-gnu"
		if arch == "386" {
			triple = "i386-unknown-linux-gnu"
		}
		for _, instruction := range x86SqrtInputRows(arch) {
			t.Run(arch+"/"+instruction, func(t *testing.T) {
				source := "TEXT ·Value(SB),4,$0\n" + instruction + "\nRET\n"
				x86LEAActualTargetGoObject(t, arch, triple, source)
				for _, api := range []string{"text", "module", "Go"} {
					ir, err := x86ScalarFPBitViewAPI(t, source, "func Value()", arch, triple, api)
					if ir != "" || !errors.Is(err, ErrProbeNeedsContext) {
						t.Fatal("dead sqrt invented numerical inputs or an address", api, err)
					}
				}
			})
		}
	}
}

func TestX86SqrtInitializedFiveTargetObjects(t *testing.T) {
	llc := findLLVM22Tool("llc")
	if llc == "" {
		t.Fatal("LLVM 22 llc required")
	}
	for _, target := range x86RawReturnTargets {
		t.Run(target.triple, func(t *testing.T) {
			word, move := int(goWordSize(target.arch)), "MOVQ"
			if word == 4 {
				move = "MOVL"
			}
			var source, declarations strings.Builder
			for index, instruction := range x86SqrtInputRows(target.arch) {
				name := fmt.Sprintf("Sqrt%d", index)
				fmt.Fprintf(&declarations, "func %s(p *[64]byte,mask uintptr) uint32\n", name)
				fmt.Fprintf(&source, "TEXT ·%s(SB),4,$0-%d\n%s p+0(FP),AX\n%s mask+%d(FP),CX\nKMOVD CX,K1\n"+
					"VPXORD Z0,Z0,Z0\nVPXORD Z1,Z1,Z1\nVPXORD Z2,Z2,Z2\n%s\nPEXTRD $0,X2,AX\nMOVL AX,ret+%d(FP)\nRET\n",
					name, 2*word+4, move, move, word, instruction, 2*word)
			}
			x86LEAActualTargetGoObject(t, target.arch, target.triple, source.String())
			for _, api := range []string{"text", "module", "Go"} {
				ir, err := x86ScalarFPBitViewAPI(t, source.String(), declarations.String(), target.arch, target.triple, api)
				if err != nil {
					t.Fatal("initialized sqrt lacks a numerical continuation", api, err)
				}
				for _, text := range []string{ir, x86DwordProductIROptimize(t, ir)} {
					for _, level := range []string{"-O0", "-O2"} {
						x86PackedStringCompileObject(t, llc, target.triple, text, level)
					}
				}
			}
		})
	}
}
