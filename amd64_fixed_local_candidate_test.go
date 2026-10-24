package plan9asm

import (
	"errors"
	"strings"
	"testing"
)

func TestGoAMD64FixedLocalFrameCandidateQueryCannotSupplyPermission(t *testing.T) {
	for _, operation := range []string{
		"PADDB 16(SP),X0", "PADDW 16(SP),X0", "PADDL 16(SP),X0", "PADDD 16(SP),X0", "PADDQ 16(SP),X0",
		"PMADDWL 16(SP),X0", "CMPPS 16(SP),X0,$0", "CMPPD 16(SP),X0,$0",
	} {
		t.Run(strings.Fields(operation)[0], func(t *testing.T) {
			source, sig, pkg := x86PrivateAlignedFrameAuditFixture(t, operation)
			file, err := Parse(ArchAMD64, source)
			if err != nil {
				t.Fatal(err)
			}
			if !GoAMD64FixedLocalFrameBindingNeeded(file, "amd64") || GoAMD64FixedLocalFrameBindingNeeded(file, "386") {
				t.Fatal("candidate query lost its exact native family/target role")
			}
			x86PrivateAlignedFrameAuditAPIs(t, source, sig, pkg, func(t *testing.T, err error) {
				if !errors.Is(err, ErrProbeNeedsContext) {
					t.Fatalf("query supplied a forged native permission: %v", err)
				}
			})
		})
	}
	for _, source := range []string{
		"TEXT ·Value(SB),4,$0-0\nPADDL 16(SP),X0\nRET\n",
		"TEXT ·Value(SB),4,$160-0\nPADDL X1,X0\nRET\n",
		"TEXT ·Value(SB),4,$160-0\nVPADDD X1,X2,X0\nRET\n",
		"TEXT ·Value(SB),4,$160-0\nPADDL (AX),M0\nRET\n",
		"TEXT ·Value(SB),4,$160-0\nPADDL X1\nRET\n",
		"TEXT ·Value(SB),4,$160-0\nPADDL constant<>(SB),X0\nRET\n",
		"TEXT ·Value(SB),4,$160-0\nPADDL (AX),X0\nRET\n",
		"TEXT ·Value(SB),4,$160-0\nPADDL (AX)(CX*2),X0\nRET\n",
		"TEXT ·Value(SB),4,$160-16\nPADDL input+0(FP),X0\nRET\n",
	} {
		file, err := Parse(ArchAMD64, source)
		if err != nil {
			t.Fatal(err)
		}
		if GoAMD64FixedLocalFrameBindingNeeded(file, "amd64") {
			t.Fatalf("query widened unrelated source grammar: %s", source)
		}
	}
}
