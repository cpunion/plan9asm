//go:build go1.27

package plan9asm

import "testing"

// Mixed DATA uses the existing audited Go 1.27 object producer. Older SDKs
// retain the native-only object gate and an explicit producer rejection test.
func TestGoAMD64FixedLocalPointerChoiceObjects(t *testing.T) {
	rows := fixedLocalPointerChoiceAllPositiveCases()
	if len(rows) != 23 {
		t.Fatal("expected 19 bounded-alternative and 4 dead/overwritten CFG functions", len(rows))
	}
	assembly, declarations := fixedLocalPointerChoiceProgram(t, fixedLocalChoiceData, rows)
	fixedLocalCompleteVectorObjects(t, assembly, declarations, len(rows))
}
