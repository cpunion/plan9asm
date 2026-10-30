package plan9asm

import (
	"errors"
	"testing"
)

// A dead ten-byte load remains an actual memory/input effect. FINIT or a
// different typed instruction must not be needed to select its obligation.
// This semantic RED is independent of the original-table formatter feature.
func TestX86OfficialLegacyX87TenByteDeadInputsNeedContext(t *testing.T) {
	for _, arch := range []string{"amd64", "386"} {
		for _, op := range []string{"FMOVB", "FMOVX"} {
			source := "TEXT ·TenByte(SB),4,$0-0\n" + op + " (AX),F0\nRET\n"
			goAssemblerOracleOutput(t, arch, source, "", true)
			for _, api := range []string{"text", "module", "Go"} {
				t.Run(arch+"/"+op+"/"+api, func(t *testing.T) {
					ir, err := x86ScalarFPBitViewAPI(t, source, "func TenByte()", arch, testTargetTriple("linux", arch), api)
					if ir != "" || !errors.Is(err, ErrProbeNeedsContext) {
						t.Fatal("X87_TEN_BYTE_INPUT_NOT_PROVED: dead BCD/extended load borrowed an unknown native address or x87 entry", err)
					}
				})
			}
		}
	}
}
