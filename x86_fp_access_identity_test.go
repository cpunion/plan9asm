package plan9asm

import (
	"strings"
	"testing"
)

func TestX86FPAccessBytesCompleteSourceIdentity(t *testing.T) {
	for _, opcode := range []string{
		"MOVB", "MOVW", "MOVL", "MOVQ",
		"MOVOU", "MOVOA", "MOVO", "MOVUPS", "MOVAPS", "MOVUPD", "MOVAPD",
		"MOVSS", "MOVSD", "VMOVSS", "VMOVSD",
	} {
		register := "AX"
		if opcode == "MOVB" {
			register = "AL"
		} else if len(opcode) > 4 || opcode == "MOVO" {
			register = "X0"
		}
		canonical := "TEXT ·probe(SB),4,$0-64\n" + opcode + " value+0(FP)," + register + "\nRET\n"
		file, err := Parse(ArchAMD64, canonical)
		if err != nil {
			t.Fatal(err)
		}
		for _, modified := range []string{opcode + ".BAD", strings.ToLower(opcode)} {
			t.Run(modified, func(t *testing.T) {
				// These are actual original assembler failures, not inferred
				// rejection from a framework which omits old SDK execution.
				source := strings.Replace(canonical, "\n"+opcode+" ", "\n"+modified+" ", 1)
				goAssemblerOracleOutput(t, "amd64", source, "", false)
				ins := file.Funcs[0].Instrs[1]
				ins.Op = Op(modified) // explicit constructed API mutation
				if width, _, known := X86FPAccessBytes(ins, 0, "amd64"); known || width != 0 {
					t.Fatalf("unclosed original source identity %q inferred %d/%v", modified, width, known)
				}
			})
		}
	}
}
