package plan9asm

import "strconv"

// This initial test-only seam retains the existing driver's single C string
// layer. The Linux object/caller regression observes the GNU assembler defect
// before the corrected encoding is shared by the discovering fixtures.
const runtimeCAsmLabelPreamble = ""

func runtimeCAsmLabel(symbol string) string {
	return strconv.Quote(symbol)
}
