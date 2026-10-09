package plan9asm

import (
	"fmt"
	"strings"
	"testing"
)

type x86DefinedZeroAFCase struct {
	op, instruction string
	amd64Only       bool
}

// These independent ISA witnesses cover every current non-arithmetic writer
// whose AF result is defined to be zero. Go encoding tables establish syntax;
// they do not describe the status effects. ENQCMD/S currently use raw source.
func x86DefinedZeroAFWriters() []x86DefinedZeroAFCase {
	cases := []x86DefinedZeroAFCase{
		{"UMWAIT", "UMWAIT BX", false},
		{"TPAUSE", "TPAUSE BX", false},
		{"PTEST", "PTEST X0,X1", false},
		{"VPTEST", "VPTEST X0,X1", false},
		{"VTESTPD", "VTESTPD X0,X1", false},
		{"VTESTPS", "VTESTPS X0,X1", false},
		{"XTEST", "XTEST", false},
		{"ENQCMD", "BYTE $0xf2\nBYTE $0x0f\nBYTE $0x38\nBYTE $0xf8\nBYTE $0x38", false},
		{"ENQCMDS", "BYTE $0xf3\nBYTE $0x0f\nBYTE $0x38\nBYTE $0xf8\nBYTE $0x38", false},
	}
	for _, stem := range []string{"RDRAND", "RDSEED", "POPCNT"} {
		for _, width := range []string{"W", "L", "Q"} {
			op := stem + width
			operand := "BX"
			if stem == "POPCNT" {
				operand = "AX,BX"
			}
			cases = append(cases, x86DefinedZeroAFCase{op, op + " " + operand, width == "Q"})
		}
	}
	for _, stem := range []string{"KTEST", "KORTEST"} {
		for _, width := range []string{"B", "W", "D", "Q"} {
			op := stem + width
			cases = append(cases, x86DefinedZeroAFCase{op, op + " K1,K2", false})
		}
	}
	for _, stem := range []string{"PCMPESTRI", "PCMPESTRM", "PCMPISTRI", "PCMPISTRM"} {
		for _, prefix := range []string{"", "V"} {
			op := prefix + stem
			cases = append(cases, x86DefinedZeroAFCase{op, op + " $0,X0,X1", false})
		}
	}
	return cases
}

func x86DefinedZeroAFSource(arch, name, instruction string) string {
	word := "Q"
	if arch == "386" {
		word = "L"
	}
	return fmt.Sprintf("TEXT %s(SB),4,$0-0\nMOVL $0,AX\nMOVL $0,DX\nMOVL $0,BX\n"+
		"PUSH%s $0x8d7\nPOPF%s\n%s\nPUSHF%s\nPOP%s CX\nRET\n",
		name, word, word, instruction, word, word)
}

func TestX86DefinedZeroAFWritersClearAfterOperation(t *testing.T) {
	for _, arch := range []string{"386", "amd64"} {
		for _, row := range x86DefinedZeroAFWriters() {
			if arch == "386" && row.amd64Only {
				continue // Go's explicit 64-bit operand row is not a 386 form.
			}
			t.Run(arch+"/"+row.op, func(t *testing.T) {
				source := x86DefinedZeroAFSource(arch, "status", row.instruction)
				requireX86GoAssemblerResult(t, arch, source, true)
				file, err := Parse(ArchAMD64, source)
				if err != nil {
					t.Fatal(err)
				}
				triple := "x86_64-unknown-linux-gnu"
				if arch == "386" {
					triple = "i386-unknown-linux-gnu"
				}
				ir, err := Translate(file, Options{Goarch: arch, TargetTriple: triple,
					Sigs: map[string]FuncSig{"status": {Name: "status", Ret: Void}},
				})
				if err != nil {
					t.Fatal(err)
				}
				// One false store initializes the slot before the explicit AF=1
				// seed. A defined-zero writer needs its own subsequent store.
				if strings.Count(ir, "store i1 false, ptr %flags_af") != 2 {
					t.Fatalf("%s retained seeded AF instead of its defined-zero status result", row.op)
				}
			})
		}
	}
}

func TestX86DefinedZeroAFExistingCompareWriterControls(t *testing.T) {
	for _, arch := range []string{"386", "amd64"} {
		triple := "x86_64-unknown-linux-gnu"
		if arch == "386" {
			triple = "i386-unknown-linux-gnu"
		}
		for _, op := range []string{"COMISS", "COMISD", "UCOMISS", "UCOMISD", "VCOMISS", "VCOMISD", "VUCOMISS", "VUCOMISD"} {
			t.Run(arch+"/"+op, func(t *testing.T) {
				source := x86DefinedZeroAFSource(arch, "status", op+" X0,X1")
				requireX86GoAssemblerResult(t, arch, source, true)
				file, err := Parse(ArchAMD64, source)
				if err != nil {
					t.Fatal(err)
				}
				ir, err := Translate(file, Options{Goarch: arch, TargetTriple: triple,
					Sigs: map[string]FuncSig{"status": {Name: "status", Ret: Void}},
				})
				if err != nil || strings.Count(ir, "store i1 false, ptr %flags_af") != 2 {
					t.Fatal("already correct compare writer lost AF clearing", err)
				}
			})
		}
	}
}
