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

// The twelve legacy and sixteen VEX/EVEX members share typed operand and
// numerical rules. Legacy retains its RMW/upper aliases and original m128
// opcode; only VEX/EVEX D/Q has scalar-memory broadcast.
func (c *amd64Ctx) lowerPackedIntegerMinMax(op Op, ins Instr) (ok bool, terminated bool, err error) {
	form, known, err := x86IntegerMinMaxProperties(op, ins, c.goarch)
	if !known || err != nil {
		return known, false, err
	}
	if err := c.validateConflictFunnelMemory(ins, form.x86PackedBinaryForm); err != nil {
		return true, false, err
	}
	first := ins.Args[form.first]
	if form.legacy && form.memory >= 0 &&
		(first.Kind == OpFP || first.Kind == OpMem && !x86OrdinaryNativeMemoryOperand(first, c.goarch) ||
			!x86ScalarFloatIntegerNativeTarget(c.goarch, c.targetTriple)) {
		return true, false, fmt.Errorf("%w: legacy integer min/max m128 needs its original native storage/target", ErrProbeNeedsContext)
	}
	spec := form.spec
	destination := ins.Args[form.destination]
	byteWidth := form.byteWidth
	lanes := byteWidth * 8 / spec.laneBits
	if form.legacy && form.memory >= 0 {
		old, err := c.loadPackedCompareLanes(ins.Args[form.second], 16, spec.laneBits, false)
		if err != nil {
			return true, false, err
		}
		value, err := c.emitPackedBinaryNativeMemory(first, spec.laneBits, old, strings.ToLower(string(op)))
		if err != nil {
			return true, false, err
		}
		bytes := c.newTmp()
		fmt.Fprintf(c.b, "  %%%s = bitcast <%d x i%d> %s to <16 x i8>\n", bytes, lanes, spec.laneBits, value)
		return true, false, c.storeLegacyPackedLogical(destination, "%"+bytes)
	}
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
	if form.legacy {
		return true, false, c.storeLegacyPackedLogical(destination, "%"+out)
	}
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
