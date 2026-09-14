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
}

// lowerPackedIntegerMinMax implements the complete 16-opcode min/max family
// in Go 1.27. B/W/D forms use _yvandnpd, Q forms use _yvblendmpd, and only
// D/Q encodings enable scalar-memory broadcast.
func (c *amd64Ctx) lowerPackedIntegerMinMax(op Op, ins Instr) (ok bool, terminated bool, err error) {
	rawOp := strings.ToUpper(string(op))
	baseOp, suffix := rawOp, ""
	if dot := strings.IndexByte(rawOp, '.'); dot >= 0 {
		baseOp, suffix = rawOp[:dot], rawOp[dot+1:]
	}
	spec, found := amd64PackedIntegerMinMaxSpecs[baseOp]
	if !found {
		return false, false, nil
	}
	properties, validSuffix := parseAMD64BinaryFloatingSuffix(suffix)
	if !validSuffix || properties.sae || properties.rounding != "" {
		return true, false, fmt.Errorf("%s %s suffix is absent from its Go 1.27 encodings: %q", c.goarch, baseOp, ins.Raw)
	}
	if properties.broadcast && !spec.broadcast {
		return true, false, fmt.Errorf("%s %s does not enable EVEX broadcast: %q", c.goarch, baseOp, ins.Raw)
	}
	if len(ins.Args) != 3 && len(ins.Args) != 4 {
		return true, false, fmt.Errorf("%s %s expects source2, source1, [K mask,] destination: %q", c.goarch, baseOp, ins.Raw)
	}
	masked := len(ins.Args) == 4
	if c.goarch == "386" && masked {
		return true, false, fmt.Errorf("386 %s mask forms exceed the Go assembler frontend's three-operand limit: %q", baseOp, ins.Raw)
	}
	if properties.zeroing && !masked {
		return true, false, fmt.Errorf("%s %s zeroing requires a K1-K7 mask: %q", c.goarch, baseOp, ins.Raw)
	}
	if masked && !amd64NonzeroKOperand(ins.Args[2]) {
		return true, false, fmt.Errorf("amd64 %s masked form expects K1-K7 as its third operand: %q", baseOp, ins.Raw)
	}
	destination := ins.Args[len(ins.Args)-1]
	if destination.Kind != OpReg {
		return true, false, fmt.Errorf("%s %s destination must be X, Y, or Z: %q", c.goarch, baseOp, ins.Raw)
	}
	byteWidth := amd64VectorByteWidth(destination.Reg)
	if byteWidth != 16 && byteWidth != 32 && byteWidth != 64 ||
		!c.isGoPackedVectorMoveRegister(destination, byteWidth) ||
		!c.isGoPackedVectorMoveRegister(ins.Args[1], byteWidth) {
		return true, false, fmt.Errorf("%s %s source1 and destination must be matching Go vector registers: %q", c.goarch, baseOp, ins.Raw)
	}
	first := ins.Args[0]
	if properties.broadcast {
		if !isAMD64MemoryOperand(first) {
			return true, false, fmt.Errorf("%s %s.BCST requires a scalar memory source2: %q", c.goarch, baseOp, ins.Raw)
		}
	} else if first.Kind == OpReg {
		if !c.isGoPackedVectorMoveRegister(first, byteWidth) {
			return true, false, fmt.Errorf("%s %s source2 register must match the destination width: %q", c.goarch, baseOp, ins.Raw)
		}
	} else if !isAMD64MemoryOperand(first) {
		return true, false, fmt.Errorf("%s %s source2 must be a matching vector register or memory: %q", c.goarch, baseOp, ins.Raw)
	}

	lanes := byteWidth * 8 / spec.laneBits
	firstValue, err := c.loadPackedCompareLanes(first, byteWidth, spec.laneBits, properties.broadcast)
	if err != nil {
		return true, false, err
	}
	secondValue, err := c.loadPackedCompareLanes(ins.Args[1], byteWidth, spec.laneBits, false)
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
		mask, err := c.loadK(ins.Args[2].Reg)
		if err != nil {
			return true, false, err
		}
		oldBytes, err := c.loadPackedCompareBytes(destination, byteWidth)
		if err != nil {
			return true, false, err
		}
		old := c.bitcastVectorBytesToIntegerLanes(byteWidth, lanes, spec.laneBits, oldBytes)
		result = amd64ApplyIntegerLaneMask(c, lanes, spec.laneBits, result, old, mask, properties.zeroing)
	}
	out := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = bitcast <%d x i%d> %s to <%d x i8>\n", out, lanes, spec.laneBits, result, byteWidth)
	return true, false, c.storeVectorBytes(destination.Reg, byteWidth, "%"+out)
}

var amd64PackedIntegerMinMaxSpecs = map[string]amd64PackedIntegerMinMaxSpec{
	"VPMINSB": {laneBits: 8, signed: true, minimum: true},
	"VPMINUB": {laneBits: 8, minimum: true},
	"VPMAXSB": {laneBits: 8, signed: true},
	"VPMAXUB": {laneBits: 8},
	"VPMINSW": {laneBits: 16, signed: true, minimum: true},
	"VPMINUW": {laneBits: 16, minimum: true},
	"VPMAXSW": {laneBits: 16, signed: true},
	"VPMAXUW": {laneBits: 16},
	"VPMINSD": {laneBits: 32, signed: true, minimum: true, broadcast: true},
	"VPMINUD": {laneBits: 32, minimum: true, broadcast: true},
	"VPMAXSD": {laneBits: 32, signed: true, broadcast: true},
	"VPMAXUD": {laneBits: 32, broadcast: true},
	"VPMINSQ": {laneBits: 64, signed: true, minimum: true, broadcast: true},
	"VPMINUQ": {laneBits: 64, minimum: true, broadcast: true},
	"VPMAXSQ": {laneBits: 64, signed: true, broadcast: true},
	"VPMAXUQ": {laneBits: 64, broadcast: true},
}
