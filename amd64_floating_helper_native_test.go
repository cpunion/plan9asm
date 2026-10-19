package plan9asm

import (
	"fmt"
	"strings"
	"testing"
)

// This is an independent inventory, not an iteration over the implementation
// registration. FPCLASS's X/Y/Z opcode spellings differ from LLVM's mnemonic.
type x86FloatingHelperOracleMember struct {
	name, mnemonic string
	immediate      bool
	second         bool
	scalar         bool
	maskResult     bool
	bank           string
}

func x86FloatingHelperOracleMembers() []x86FloatingHelperOracleMember {
	var members []x86FloatingHelperOracleMember
	for _, family := range []struct {
		name      string
		immediate bool
		second    bool
	}{
		{"VGETEXP", false, false},
		{"VGETMANT", true, false},
		{"VSCALEF", false, true},
		{"VRANGE", true, true},
		{"VFIXUPIMM", true, true},
		{"VRNDSCALE", true, false},
		{"VREDUCE", true, false},
	} {
		for _, suffix := range []string{"PS", "PD", "SS", "SD"} {
			members = append(members, x86FloatingHelperOracleMember{
				name: family.name + suffix, mnemonic: strings.ToLower(family.name + suffix),
				immediate: family.immediate, second: family.second,
				scalar: suffix == "SS" || suffix == "SD", bank: "X",
			})
		}
	}
	for _, precision := range []string{"PS", "PD"} {
		for _, bank := range []string{"X", "Y", "Z"} {
			members = append(members, x86FloatingHelperOracleMember{
				name: "VFPCLASS" + precision + bank, mnemonic: "vfpclass" + strings.ToLower(precision),
				immediate: true, maskResult: true, bank: bank,
			})
		}
	}
	for _, precision := range []string{"SS", "SD"} {
		members = append(members, x86FloatingHelperOracleMember{
			name: "VFPCLASS" + precision, mnemonic: "vfpclass" + strings.ToLower(precision),
			immediate: true, scalar: true, maskResult: true, bank: "X",
		})
	}
	return members
}

func x86FloatingHelperOracleInstruction(member x86FloatingHelperOracleMember, memory bool) string {
	var args []string
	if member.immediate {
		args = append(args, "$255")
	}
	source := member.bank + "0"
	if memory {
		source = "8(AX)"
	}
	args = append(args, source)
	if !member.maskResult && (member.second || member.scalar) {
		args = append(args, member.bank+"1")
	}
	destination := member.bank + "2"
	if member.maskResult {
		destination = "K2"
	}
	return member.name + " " + strings.Join(append(args, destination), ",")
}

func TestX86FloatingHelperCompleteNativeOpcodes(t *testing.T) {
	members := x86FloatingHelperOracleMembers()
	if len(members) != 36 {
		t.Fatal("independent floating-helper family denominator changed", len(members))
	}
	for _, target := range x86RawReturnTargets {
		t.Run(target.triple, func(t *testing.T) {
			var source strings.Builder
			signatures := make(map[string]FuncSig)
			var expected []string
			for _, member := range members {
				// Go's named mode32 frontend has no four-operand forms. Its
				// physical encodings are a separate decoder test obligation.
				if target.arch == "386" && (member.immediate && (member.second || member.scalar) && !member.maskResult) {
					continue
				}
				for _, memory := range []bool{false, true} {
					name := fmt.Sprintf("F%d", len(signatures))
					fmt.Fprintf(&source, "TEXT ·%s(SB),4,$0-0\n%s\nRET\n", name,
						x86FloatingHelperOracleInstruction(member, memory))
					signatures[name] = FuncSig{Name: name, Ret: Void}
					expected = append(expected, member.mnemonic)
				}
			}
			x86LEAActualTargetGoObject(t, target.arch, target.triple, source.String())
			file, err := Parse(ArchAMD64, source.String())
			if err != nil {
				t.Fatal(err)
			}
			// These deliberately unbound original formats own only emitter
			// objects. They do not claim typed inputs or a public translation.
			ir := x86ScalarFormatObjectUnit(t, file, signatures, target.arch, target.triple, 0)
			if got := strings.Count(ir, "asm sideeffect"); got != len(expected) {
				t.Fatalf("floating helpers lost original native MXCSR/exception/access effects: %d native instructions, want %d", got, len(expected))
			}
			for ordinal, mnemonic := range expected {
				start := strings.Index(ir, fmt.Sprintf("@InstructionFormat%d(", ordinal))
				if start < 0 {
					t.Fatal("missing original format function", ordinal)
				}
				body, _, _ := strings.Cut(ir[start:], "\n}")
				if !strings.Contains(body, mnemonic+" ") || !strings.Contains(body, "~{fpsr}") ||
					!strings.Contains(body, "~{memory}") {
					t.Fatal("native source opcode or observable clobbers lost", ordinal, mnemonic)
				}
			}
			llc := findLLVM22Tool("llc")
			if llc == "" {
				t.Fatal("LLVM 22 llc not found")
			}
			compileLLVMToObject(t, llc, target.triple, "floating-helper-native.ll", "floating-helper-native.o", ir)
		})
	}
}
