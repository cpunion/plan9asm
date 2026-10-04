package plan9asm

import (
	"fmt"
	"strings"
	"testing"
)

// The arithmetic carrier must retain the same lane type as the horizontal
// shuffles. These are the complete register rows of Go's yxm/_yvaddsubpd
// tables, independently enumerated rather than derived from the emitter.
func TestX86HorizontalFloatArithmeticPreservesTypedLaneWidth(t *testing.T) {
	for _, opcode := range []string{
		"HADDPS", "HADDPD", "HSUBPS", "HSUBPD",
		"VHADDPS", "VHADDPD", "VHSUBPS", "VHSUBPD",
	} {
		banks := []string{"X"}
		vector := strings.HasPrefix(opcode, "V")
		if vector {
			banks = append(banks, "Y")
		}
		for _, bank := range banks {
			t.Run(opcode+"/"+bank, func(t *testing.T) {
				row := fmt.Sprintf("%s %s0,%s1", opcode, bank, bank)
				if vector {
					row += "," + bank + "2"
				}
				source := "TEXT f(SB),4,$0-0\n" + row + "\nRET\n"
				requireX86GoAssemblerResult(t, "amd64", source, true)
				file, err := Parse(ArchAMD64, source)
				if err != nil {
					t.Fatal(err)
				}
				ir, err := Translate(file, Options{
					Goarch: "amd64", TargetTriple: "x86_64-unknown-linux-gnu",
					Sigs: map[string]FuncSig{"f": {Name: "f", Ret: Void}},
				})
				if err != nil {
					t.Fatal(err)
				}
				bits, element, width := 32, "float", 16
				if strings.HasSuffix(opcode, "PD") {
					bits, element = 64, "double"
				}
				if bank == "Y" {
					width = 32
				}
				operation := "fadd"
				if strings.Contains(opcode, "SUB") {
					operation = "fsub"
				}
				lanes := width * 8 / bits
				call := fmt.Sprintf("call <%d x %s> @llvm.experimental.constrained.%s.v%df%d(", lanes, element, operation, lanes, bits)
				if strings.Count(ir, call) != 1 {
					t.Fatalf("%s lost its complete typed arithmetic carrier %q", row, call)
				}
			})
		}
	}
}
