package plan9asm

import (
	"errors"
	"fmt"
	"testing"
)

func TestX86PoppedMixedWidthsRetainOriginalFlagBitPositions(t *testing.T) {
	for _, triple := range []string{"i386-unknown-linux-gnu", "i686-pc-windows-msvc"} {
		for _, row := range []struct {
			name   string
			mask   int
			accept bool
		}{{"unmodeled-AC", 4, false}, {"modeled-ID", 32, true}} {
			source := fmt.Sprintf("TEXT ·Value(SB),4,$0-4\nPUSHL $0\nPOPFL\nMOVL $3,AX\nMULL AX\nCLD\nPUSHFL\nPOPW BX\nPOPW AX\nANDW $%d,AX\nMOVW AX,ret+0(FP)\nMOVW $0,ret+2(FP)\nRET\n", row.mask)
			goAssemblerOracleOutput(t, "386", source, "", true)
			for _, api := range []string{"text", "module", "Go"} {
				t.Run(triple+"/"+row.name+"/"+api, func(t *testing.T) {
					ir, err := x86ScalarFPBitViewAPI(t, source, "func Value() uint32", "386", triple, api)
					if row.accept && (err != nil || ir == "") || !row.accept && (!errors.Is(err, ErrProbeNeedsContext) || ir != "") {
						t.Fatal("mixed-width flags transport changed an original bit role", row.accept, err)
					}
				})
			}
		}
	}
}
