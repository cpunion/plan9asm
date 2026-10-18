package plan9asm

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

func x86MultiplyAddInputRows(arch string) []string {
	var rows []string
	for _, opcode := range []string{"PMADDWL", "PMADDUBSW"} {
		rows = append(rows, opcode+" X0,X1", opcode+" (BX),X1")
	}
	for _, opcode := range []string{"VPMADDWD", "VPMADDUBSW"} {
		for _, bank := range []string{"X", "Y", "Z"} {
			for _, input := range []string{bank + "0", "(BX)"} {
				rows = append(rows, opcode+" "+input+","+bank+"1,"+bank+"2")
				if arch == "amd64" {
					rows = append(rows, opcode+" "+input+","+bank+"1,K1,"+bank+"2",
						opcode+".Z "+input+","+bank+"1,K1,"+bank+"2")
				}
			}
		}
	}
	return rows
}

func TestX86MultiplyAddVoidCannotInventInputs(t *testing.T) {
	for _, arch := range []string{"amd64", "386"} {
		triple, count := "x86_64-unknown-linux-gnu", 40
		if arch == "386" {
			triple, count = "i386-unknown-linux-gnu", 16
		}
		rows := x86MultiplyAddInputRows(arch)
		if len(rows) != count {
			t.Fatal("incomplete independent multiply-add input inventory", arch, len(rows))
		}
		var original strings.Builder
		for i, row := range rows {
			fmt.Fprintf(&original, "TEXT f%d(SB),4,$0-0\n%s\nRET\n", i, row)
		}
		x86LEAActualTargetGoObject(t, arch, triple, original.String())
		for _, row := range rows {
			for _, unused := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/%s/unused-bool=%t", arch, row, unused), func(t *testing.T) {
					declaration, extent := "func Value()", 0
					if unused {
						declaration, extent = "func Value(unused bool)", 1
					}
					source := fmt.Sprintf("TEXT ·Value(SB),4,$0-%d\n%s\nRET\n", extent, row)
					for _, api := range []string{"text", "module", "Go"} {
						ir, err := x86ScalarFPBitViewAPI(t, source, declaration, arch, triple, api)
						if ir != "" || !errors.Is(err, ErrProbeNeedsContext) {
							t.Fatal("Void or unused bool supplied absent vector, address or mask inputs", api, err)
						}
					}
				})
			}
		}
	}
}

func TestX86MultiplyAddInitializedInputsFiveTargetObjects(t *testing.T) {
	llc := findLLVM22Tool("llc")
	if llc == "" {
		t.Fatal("LLVM22 llc required")
	}
	for _, target := range x86RawReturnTargets {
		t.Run(target.triple, func(t *testing.T) {
			word, move := 8, "MOVQ"
			if target.arch == "386" {
				word, move = 4, "MOVL"
			}
			var source, declarations strings.Builder
			for index, row := range x86MultiplyAddInputRows(target.arch) {
				name := fmt.Sprint("MultiplyAdd", index)
				fmt.Fprintf(&declarations, "func %s(p *[64]byte,a,b [64]byte,mask uint64)\n", name)
				fmt.Fprintf(&source, "TEXT ·%s(SB),4,$0-%d\n%s p+0(FP),BX\n", name, word+136, move)
				fmt.Fprintf(&source, "VMOVUPS a+%d(FP),Z0\nVMOVUPS b+%d(FP),Z1\nVMOVUPS b+%d(FP),Z2\n",
					word, word+64, word+64)
				if target.arch == "amd64" {
					fmt.Fprintf(&source, "KMOVQ mask+%d(FP),K1\n", word+128)
				}
				fmt.Fprintf(&source, "%s\nRET\n", row)
			}
			x86LEAActualTargetGoObject(t, target.arch, target.triple, source.String())
			for _, api := range []string{"text", "module", "Go"} {
				t.Run(api, func(t *testing.T) {
					ir, err := x86ScalarFPBitViewAPI(t, source.String(), declarations.String(), target.arch, target.triple, api)
					if err != nil {
						t.Fatal("complete initialized multiply-add family rejected", err)
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
