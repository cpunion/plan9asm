package plan9asm

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestX86PairedNativeDeadInputsNeedActualSourceClosure(t *testing.T) {
	rows := []struct{ name, instruction string }{
		{"ternary", "VPTERNLOGD.Z $0,(AX),X1,K7,X2"},
		{"integer-minmax", "VPMAXSD.Z (AX),X1,K7,X2"},
		{"horizontal-reduction", "VHADDPS (AX),X1,X2"},
		{"broadcast", "VPBROADCASTD.Z (AX),K7,X2"},
	}
	// A real mask value of zero still needs a source-bound address role.
	// The complete positive controls also retain an initialized pointer before
	// the jump, so a repair cannot simply reject branches or unreachable code.
	prefix := "MOVQ p+0(FP),AX\nMOVL $0,CX\nKMOVD CX,K7\n" +
		"VPXORD Z0,Z0,Z0\nVPXORD Z1,Z1,Z1\nVPXORD Z2,Z2,Z2\n"
	for _, row := range rows {
		for _, dead := range []bool{false, true} {
			body := row.instruction + "\nRET\n"
			if dead {
				body = "JMP finish\n" + row.instruction + "\nfinish:\nRET\n"
			}
			control := "TEXT ·Value(SB),4,$0-8\n" + prefix + body
			changed := strings.Replace(control, "MOVQ p+0(FP),AX", "MOVQ $0,AX", 1)
			for _, api := range []string{"text", "module", "Go"} {
				t.Run(fmt.Sprintf("%s/dead=%t/%s", row.name, dead, api), func(t *testing.T) {
					if _, err := x86ScalarFPBitViewAPI(t, control, "func Value(p *byte)", "amd64", "x86_64-unknown-linux-gnu", api); err != nil {
						t.Fatal("genuine pointer/mask positive control failed before dead-source comparison", err)
					}
					if ir, err := x86ScalarFPBitViewAPI(t, changed, "func Value(p *byte)", "amd64", "x86_64-unknown-linux-gnu", api); ir != "" || !errors.Is(err, ErrProbeNeedsContext) {
						t.Fatal("shape-only all-known GP values supplied an unreachable native pointer", err)
					}
				})
			}
		}
	}
}

func TestX86TernaryDeadNamedRawMasksAndAddressesRemainObligations(t *testing.T) {
	for _, row := range x86TernaryCompleteOracleRows() {
		if !row.memory && row.mask < 0 {
			continue // A dead unmasked bitwise result may legitimately stay unknown.
		}
		for _, spelling := range []string{"named", "raw"} {
			instruction := strings.Replace(row.row, "$150", "$0", 1)
			if spelling == "raw" {
				code := x86TernaryIndependentFixtureBytes(row)
				code[len(code)-1] = 0
				instruction = x86RawDisplacementDirectives(code)
			}
			instruction = "JMP finish\n" + instruction + "\nfinish:"
			source, declaration := x86TernaryInitializedFixture("Value", "amd64", instruction)
			// Establish a real zero mask, not an assumed value in the source
			// shape state. The unknown base or K role remains independently
			// required even when no result or active lane consumes the read.
			source = strings.Replace(source, "MOVL $85,DX", "MOVL $0,DX", 1)
			changed := source
			if row.memory {
				changed = strings.Replace(changed, "MOVQ p+0(FP),AX", "MOVQ $0,AX", 1)
			} else {
				changed = strings.Replace(changed, "KMOVD DX,K7\n", "", 1)
			}
			for _, api := range []string{"text", "module", "Go"} {
				t.Run(spelling+"/"+row.row+"/"+api, func(t *testing.T) {
					if _, err := x86ScalarFPBitViewAPI(t, source, declaration, "amd64", "x86_64-unknown-linux-gnu", api); err != nil {
						t.Fatal("complete initialized dead-source control lost a legitimate typed role", err)
					}
					if ir, err := x86ScalarFPBitViewAPI(t, changed, declaration, "amd64", "x86_64-unknown-linux-gnu", api); ir != "" || !errors.Is(err, ErrProbeNeedsContext) {
						t.Fatal("constant selector, zero mask or dead CFG hid an unbound native input", err)
					}
				})
			}
		}
	}
}
