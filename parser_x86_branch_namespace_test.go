package plan9asm

import (
	"fmt"
	"strings"
	"testing"
)

func TestParseX86BranchLabelsRespectGoRegisterCase(t *testing.T) {
	// Go's shared jump parser classifies its original, case-sensitive token
	// before distinguishing JMP, CALL, Jcc, LOOP and transactional branches.
	var labels []string
	for _, name := range x86SymbolIndexedActualRegisterNames(t) {
		labels = append(labels, strings.ToLower(name))
	}
	// Cross-architecture names and Intel aliases are not Go x86 registers.
	labels = append(labels, "R0", "V1", "ZR", "LR", "RAX", "EAX", "Xmm0", "aX", "G", "pc", "fp", "sb")
	for _, arch := range []string{"amd64", "386"} {
		for _, op := range []string{"JMP", "CALL", "JEQ", "LOOP", "XBEGIN"} {
			t.Run(arch+"/"+op, func(t *testing.T) {
				var source strings.Builder
				for index, label := range labels {
					fmt.Fprintf(&source, "TEXT ·branch%d(SB),4,$0-0\n%s %s\n%s:\nRET\n", index, op, label, label)
				}
				goAssemblerOracleOutput(t, arch, source.String(), "", true)
				file, err := Parse(ArchAMD64, source.String())
				if err != nil || len(file.Funcs) != len(labels) {
					t.Fatal("complete original label inventory did not parse", err)
				}
				for index, function := range file.Funcs {
					operand := function.Instrs[1].Args[0]
					if operand.Kind != OpIdent || operand.Ident != labels[index] {
						t.Errorf("Go label %q became a register or changed case: %+v", labels[index], operand)
					}
				}
			})
		}
	}
}

func TestParseX86BranchOperandsKeepNativeRegisterIdentity(t *testing.T) {
	for _, op := range []string{"JMP", "CALL", "JEQ", "LOOP", "XBEGIN"} {
		for _, name := range []string{"AX", "BX", "R8", "X1", "K1", "SP", "BP", "g"} {
			operands, err := parseOperandsCSV(ArchAMD64, Op(op), name)
			if err != nil || len(operands) != 1 || operands[0].Kind != OpReg {
				t.Errorf("original register token became a local label: %s %s: %+v / %v", op, name, operands, err)
			}
		}
	}
	for _, op := range []string{"JMP", "CALL"} {
		for _, spelling := range []string{"*AX", "(AX)", "·target(SB)", "4(PC)"} {
			operands, err := parseOperandsCSV(ArchAMD64, Op(op), spelling)
			if err != nil || len(operands) != 1 || operands[0].Kind == OpIdent || operands[0].Kind == OpLabel {
				t.Errorf("indirect, external or PC-relative source became a bare label: %s %s: %+v / %v", op, spelling, operands, err)
			}
		}
	}
	for _, arch := range []string{"amd64", "386"} {
		for _, op := range []string{"JMP", "CALL", "JEQ", "LOOP", "XBEGIN"} {
			// A label with the same spelling does not override Go's register
			// namespace. JMP/CALL retain indirect AX; the others reject it.
			source := fmt.Sprintf("TEXT ·Value(SB),4,$0-0\n%s AX\nAX:\nRET\n", op)
			goAssemblerOracleOutput(t, arch, source, "", op == "JMP" || op == "CALL")
		}
	}
}
