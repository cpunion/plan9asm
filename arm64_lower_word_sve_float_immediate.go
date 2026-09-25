package plan9asm

import (
	"fmt"
	"math"
)

// Go 1.27's ZFDUP encoder row has one 8-bit floating immediate, three
// element widths, and any Z destination. LLVM's FMOV spelling is an alias.
func decodeARM64RawSVEFloatImmediate(word uint32) (Instr, bool) {
	const variableFields = uint32(0x00c01fff)
	if word&^variableFields != 0x2539c000 {
		return Instr{}, false
	}
	width := map[uint32]string{1: "H", 2: "S", 3: "D"}[word>>22&3]
	if width == "" {
		return Instr{}, false
	}
	immediate := byte(word >> 5)
	value := arm64ExpandedFloatImmediate(immediate)
	destination := int(word) & 31
	return Instr{
		Op: "ZFDUP",
		Args: []Operand{
			{Kind: OpImm, Imm: int64(math.Float64bits(value)), ImmIsFloat: true},
			{Kind: OpReg, Reg: Reg(fmt.Sprintf("Z%d.%s", destination, width))},
		},
		Raw: fmt.Sprintf("WORD $%#08x", word),
	}, true
}
