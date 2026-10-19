package plan9asm

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

// These are independent numerical programs, not replacements for the original
// eighteen-TEXT Windows fixture. Two copies of the same FLAGS bits must cancel
// after XOR even when their entry values are unknown. No unknown bit becomes
// available to an unrelated read, branch, pointer or return.
func x86StackFlagsCorrelationBody(width, register string, mask uint32) string {
	return fmt.Sprintf("PUSHF%s\nPUSHF%s\nXOR%s $%d,0(SP)\nPOPF%s\n"+
		"PUSHF%s\nPOP%s %s\nXOR%s 0(SP),%s\nPOPF%s\n",
		width, width, width, mask, width, width, width, register, width, register, width)
}

func TestX86Mode32PrivateFlagsCopiesCancelUnknownInputs(t *testing.T) {
	for _, row := range []struct {
		width, typ string
		bytes      int
		masks      []uint32
	}{
		{"W", "uint16", 2, []uint32{0, 1, 0x40, 0x400, 0x8d5}},
		{"L", "uint32", 4, []uint32{0, 1, 0x400, 0x200000, 0x200cd5}},
	} {
		for _, register := range []string{"AX", "BX", "CX", "DX", "SI", "DI"} {
			for _, mask := range row.masks {
				body := x86StackFlagsCorrelationBody(row.width, register, mask)
				source := fmt.Sprintf("TEXT ·Value(SB),4,$0-%d\n%sMOV%s %s,ret+0(FP)\nRET\n",
					row.bytes, body, row.width, register)
				for _, api := range []string{"text", "module", "Go"} {
					t.Run(fmt.Sprintf("%s/%s/%x/%s", row.width, register, mask, api), func(t *testing.T) {
						ir, err := x86ScalarFPBitViewAPI(t, source, "func Value() "+row.typ,
							"386", "i386-unknown-linux-gnu", api)
						if err != nil || ir == "" {
							t.Fatal("balanced private FLAGS copies did not cancel their same unknown bits", err)
						}
					})
				}
			}
		}
	}
}

func TestX86Mode32PrivateFlagsCorrelationCannotInventValues(t *testing.T) {
	const prefix = "TEXT ·Value(SB),4,$0-4\n"
	const suffix = "MOVL AX,ret+0(FP)\nRET\n"
	base := x86StackFlagsCorrelationBody("L", "AX", 0x200000)
	for _, row := range []struct{ name, before, after string }{
		{"uncancelled", "XORL 0(SP),AX\n", ""},
		{"different-unknown", "XORL 0(SP),AX", "XORL BX,AX"},
		{"unwritten-upper", "POPL AX", "POPW AX\nPOPW BX"},
		{"outside-live-stack", "XORL 0(SP),AX", "XORL 4(SP),AX"},
		{"partial-outside-live-stack", "XORL 0(SP),AX", "XORL 1(SP),AX"},
		{"before-live-stack", "XORL 0(SP),AX", "XORL -4(SP),AX"},
		{"indexed-stack", "XORL 0(SP),AX", "XORL 0(SP)(CX*1),AX"},
		{"symbolic-stack", "XORL 0(SP),AX", "XORL saved+0(SP),AX"},
		{"changed-stack-pointer", "XORL 0(SP),AX", "MOVL $0,SP\nXORL 0(SP),AX"},
		{"unbalanced-return", "XORL 0(SP),AX\nPOPFL", "XORL 0(SP),AX"},
	} {
		if strings.Count(base, row.before) != 1 {
			t.Fatal("independent mutation is not unique", row.name)
		}
		source := prefix + strings.Replace(base, row.before, row.after, 1) + suffix
		for _, api := range []string{"text", "module", "Go"} {
			t.Run(row.name+"/"+api, func(t *testing.T) {
				ir, err := x86ScalarFPBitViewAPI(t, source, "func Value() uint32",
					"386", "i386-unknown-linux-gnu", api)
				if ir != "" || !errors.Is(err, ErrProbeNeedsContext) {
					t.Fatal("private FLAGS identity invented a value or backing stack extent", err)
				}
			})
		}
	}
}
