package plan9asm

import (
	"errors"
	"strings"
	"testing"
)

func TestX86X87ConditionalOriginalRuntimeSourcePortableBoundary(t *testing.T) {
	source := x86X87ENVLiteral(t, "amd64_lower_x87_conditional_move_test.go", "TestAMD64X87ConditionalMoveRuntimeConditions")
	// Actual Go syntax/object acceptance is independent of a native pointer
	// extent, caller write envelope, or a portable numerical proof.
	goAssemblerOracleOutput(t, "amd64", "#include \"textflag.h\"\n"+source, "", true)
	for _, triple := range []string{"x86_64-unknown-linux-gnu", "x86_64-apple-darwin", "x86_64-pc-windows-msvc"} {
		for _, api := range []string{"text", "module", "Go"} {
			t.Run(triple+"/"+api, func(t *testing.T) {
				ir, err := x86X87InitializedWordAPI(t, source, "func x87conditionalmovesemantics(data *byte)",
					"amd64", triple, api, X87Software)
				if !errors.Is(err, ErrProbeNeedsContext) || ir != "" || !strings.Contains(err.Error(), "FMOVD 64(DI)") {
					t.Fatal("complete original conditional source acquired an unproved native read/write region", err)
				}
			})
		}
	}
}
