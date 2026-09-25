package plan9asm

import (
	"encoding/binary"
	"testing"

	"golang.org/x/arch/x86/x86asm"
)

func TestX86RawLegacyScalarSourceLocalRIPData(t *testing.T) {
	for _, op := range []string{
		"MOVD", "UCOMISD", "UCOMISS", "COMISD", "COMISS",
		"CMPSS", "CMPSD",
		"ADDSS", "ADDSD", "SUBSS", "SUBSD", "MULSS", "MULSD",
		"DIVSS", "DIVSD", "MINSS", "MINSD", "MAXSS", "MAXSD",
		"SQRTSS", "SQRTSD",
	} {
		for _, destination := range []string{"X0", "X10"} {
			t.Run(op+"/"+destination, func(t *testing.T) {
				instruction := op + " pool(SB), " + destination
				if op == "CMPSS" || op == "CMPSD" {
					instruction += ", $1"
				}
				code := assembleX87ControlBytes(t, "amd64",
					"TEXT scalarRIP(SB),4,$0-0\n\t"+instruction+"\n\tRET\n")
				inst, err := x86asm.Decode(code, 64)
				if err != nil || inst.PCRel != 4 || inst.MemBytes <= 0 || inst.Len >= len(code) {
					t.Fatalf("Go assembler %s %s: %x, decode %+v, %v", op, destination, code, inst, err)
				}
				binary.LittleEndian.PutUint32(code[inst.PCRelOff:], 1)
				for value := 0; value < inst.MemBytes; value++ {
					code = append(code, byte(value))
				}
				decoded, err := decodeX86RawDirectiveGroup(code, 64, 0, "scalar RIP", map[string]bool{})
				if err != nil {
					t.Fatal(err)
				}
				if len(decoded) != 2 || len(decoded[0].x86RIPLiteralData) != inst.MemBytes {
					t.Fatalf("%s %s decoded as %#v", op, destination, decoded)
				}
			})
		}
	}
}
