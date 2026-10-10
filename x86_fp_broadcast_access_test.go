package plan9asm

import (
	"fmt"
	"strings"
	"testing"
)

func TestX86FPAccessBytesBroadcastCompleteGoForms(t *testing.T) {
	for _, arch := range []string{"386", "amd64"} {
		t.Run(arch, func(t *testing.T) {
			var source strings.Builder
			type expected struct {
				instruction string
				width       int64
			}
			var rows []expected
			// This independent test directory is checked against every original
			// Go broadcast opcode/table, not derived from the production grammar.
			for _, spec := range x86BroadcastTestDirectory {
				if spec.source == "K" {
					continue // Mask-data forms have no FP/memory row; reject below.
				}
				for _, bank := range spec.destinations {
					for _, masking := range []string{"", "K1,", "K7,"} {
						if strings.HasSuffix(spec.name, "128") && masking != "" {
							continue
						}
						op := spec.name
						if masking == "K7," {
							op += ".Z"
						}
						instruction := fmt.Sprintf("%s a+0(FP),%s%c7", op, masking, bank)
						fmt.Fprintf(&source, "TEXT form%d(SB),4,$0-32\n%s\nRET\n", len(rows), instruction)
						rows = append(rows, expected{instruction, int64(spec.bytes)})
					}
				}
			}
			goAssemblerOracleOutput(t, arch, source.String(), "", true)
			file, err := Parse(ArchAMD64, source.String())
			if err != nil {
				t.Fatal(err)
			}
			for index, fn := range file.Funcs {
				ins := fn.Instrs[1]
				width, write, known := X86FPAccessBytes(ins, 0, arch)
				if !known || write || width != rows[index].width {
					t.Errorf("%s = %d/%t/%t, want %d/false/true", rows[index].instruction, width, write, known, rows[index].width)
				}
				for operand := 1; operand < len(ins.Args); operand++ {
					if width, write, known := X86FPAccessBytes(ins, operand, arch); known || write || width != 0 {
						t.Fatal("vector/mask acquired FP byte ownership", ins.Raw, operand)
					}
				}
			}
		})
	}
}

func TestX86FPAccessBytesBroadcastRejectedForms(t *testing.T) {
	for _, arch := range []string{"386", "amd64"} {
		for _, row := range []string{
			"VPBROADCASTMB2Q a+0(FP),X0", "VPBROADCASTMW2D a+0(FP),Y0",
			"VPBROADCASTQ a+0(FP),AX", "VPBROADCASTQ X0,a+0(FP)",
			"VPBROADCASTQ.Z a+0(FP),Y0", "VPBROADCASTQ a+0(FP),K0,Y0",
			"VBROADCASTSD a+0(FP),X0", "VBROADCASTF128 a+0(FP),X0",
			"VBROADCASTI128 a+0(FP),K1,Y0", "VBROADCASTF32X8 a+0(FP),Y0",
		} {
			t.Run(arch+"/"+row, func(t *testing.T) {
				source := "TEXT f(SB),4,$0-32\n" + row + "\nRET\n"
				goAssemblerOracleOutput(t, arch, source, "", false)
				file, err := Parse(ArchAMD64, source)
				if err != nil {
					t.Fatal(err)
				}
				ins := file.Funcs[0].Instrs[1]
				for index := range ins.Args {
					if width, write, known := X86FPAccessBytes(ins, index, arch); known || write || width != 0 {
						t.Fatal("invalid broadcast acquired FP bytes", width, write, known)
					}
				}
			})
		}
	}
}
