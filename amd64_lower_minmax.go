package plan9asm

import (
	"fmt"
	"strings"
)

type amd64PackedIntegerMinMaxSpec struct {
	laneBits  int
	signed    bool
	minimum   bool
	broadcast bool
	mapNumber int
	opcode    int
}

// lowerPackedIntegerMinMax implements the complete 16-opcode min/max family
// in Go 1.27. B/W/D forms use _yvandnpd, Q forms use _yvblendmpd, and only
// D/Q encodings enable scalar-memory broadcast.
func (c *amd64Ctx) lowerPackedIntegerMinMax(op Op, ins Instr) (ok bool, terminated bool, err error) {
	form, known, err := x86IntegerMinMaxProperties(op, ins, c.goarch)
	if !known || err != nil {
		return known, false, err
	}
	if err := c.validateConflictFunnelMemory(ins, form.x86PackedBinaryForm); err != nil {
		return true, false, err
	}
	spec := form.spec
	destination := ins.Args[form.destination]
	byteWidth := form.byteWidth
	lanes := byteWidth * 8 / spec.laneBits
	masked := form.mask >= 0
	mask := ""
	if masked {
		mask, err = c.loadK(ins.Args[form.mask].Reg)
		if err != nil {
			return true, false, err
		}
	}
	firstValue, err := c.loadX86PackedShiftLanes(ins.Args[form.first], byteWidth, spec.laneBits, form.broadcast, mask)
	if err != nil {
		return true, false, err
	}
	secondValue, err := c.loadPackedCompareLanes(ins.Args[form.second], byteWidth, spec.laneBits, false)
	if err != nil {
		return true, false, err
	}
	predicate := "ult"
	if spec.signed {
		predicate = "slt"
	}
	if !spec.minimum {
		predicate = strings.Replace(predicate, "lt", "gt", 1)
	}
	compared := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = icmp %s <%d x i%d> %s, %s\n", compared, predicate, lanes, spec.laneBits, secondValue, firstValue)
	computed := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = select <%d x i1> %%%s, <%d x i%d> %s, <%d x i%d> %s\n", computed, lanes, compared, lanes, spec.laneBits, secondValue, lanes, spec.laneBits, firstValue)
	result := "%" + computed
	if masked {
		oldBytes, err := c.loadPackedCompareBytes(destination, byteWidth)
		if err != nil {
			return true, false, err
		}
		old := c.bitcastVectorBytesToIntegerLanes(byteWidth, lanes, spec.laneBits, oldBytes)
		result = amd64ApplyIntegerLaneMask(c, lanes, spec.laneBits, result, old, mask, form.zeroing)
	}
	out := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = bitcast <%d x i%d> %s to <%d x i8>\n", out, lanes, spec.laneBits, result, byteWidth)
	return true, false, c.storePackedMoveOperand(destination, byteWidth, "%"+out)
}

var amd64PackedIntegerMinMaxSpecs = map[string]amd64PackedIntegerMinMaxSpec{
	"VPMINSB": {laneBits: 8, signed: true, minimum: true, mapNumber: 2, opcode: 0x38},
	"VPMINUB": {laneBits: 8, minimum: true, mapNumber: 1, opcode: 0xda},
	"VPMAXSB": {laneBits: 8, signed: true, mapNumber: 2, opcode: 0x3c},
	"VPMAXUB": {laneBits: 8, mapNumber: 1, opcode: 0xde},
	"VPMINSW": {laneBits: 16, signed: true, minimum: true, mapNumber: 1, opcode: 0xea},
	"VPMINUW": {laneBits: 16, minimum: true, mapNumber: 2, opcode: 0x3a},
	"VPMAXSW": {laneBits: 16, signed: true, mapNumber: 1, opcode: 0xee},
	"VPMAXUW": {laneBits: 16, mapNumber: 2, opcode: 0x3e},

	"VPMINSD": {laneBits: 32, signed: true, minimum: true, broadcast: true, mapNumber: 2, opcode: 0x39},
	"VPMINUD": {laneBits: 32, minimum: true, broadcast: true, mapNumber: 2, opcode: 0x3b},
	"VPMAXSD": {laneBits: 32, signed: true, broadcast: true, mapNumber: 2, opcode: 0x3d},
	"VPMAXUD": {laneBits: 32, broadcast: true, mapNumber: 2, opcode: 0x3f},
	"VPMINSQ": {laneBits: 64, signed: true, minimum: true, broadcast: true, mapNumber: 2, opcode: 0x39},
	"VPMINUQ": {laneBits: 64, minimum: true, broadcast: true, mapNumber: 2, opcode: 0x3b},
	"VPMAXSQ": {laneBits: 64, signed: true, broadcast: true, mapNumber: 2, opcode: 0x3d},
	"VPMAXUQ": {laneBits: 64, broadcast: true, mapNumber: 2, opcode: 0x3f},
}
