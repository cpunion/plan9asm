package plan9asm

import (
	"fmt"
	"strings"
)

// Keep the complete opcode set in a package-level Op map so corpus coverage
// extraction sees this family without depending on spelling tests in the
// lowering body.
var amd64PackedUnsignedWordMinimumPositionVector = map[Op]bool{
	"PHMINPOSUW":  false,
	"VPHMINPOSUW": true,
}

// lowerPackedUnsignedWordMinimumPosition implements every Go 1.27 form of
// PHMINPOSUW (yxm_q4) and VPHMINPOSUW (_yvaesimc). Both tables are restricted
// to a single X/m128 source and an X destination; the VEX spelling is also a
// two-operand, destructive-looking Plan 9 form.
func (c *amd64Ctx) lowerPackedUnsignedWordMinimumPosition(op Op, ins Instr) (ok bool, terminated bool, err error) {
	rawOp := strings.ToUpper(string(op))
	baseOp := rawOp
	if dot := strings.IndexByte(rawOp, '.'); dot >= 0 {
		baseOp = rawOp[:dot]
	}
	vector, recognized := amd64PackedUnsignedWordMinimumPositionVector[Op(baseOp)]
	if !recognized {
		return false, false, nil
	}
	if rawOp != baseOp {
		return true, false, fmt.Errorf("%s %s has no suffixed forms in Go 1.27: %q", c.goarch, baseOp, ins.Raw)
	}
	if len(ins.Args) != 2 || ins.Args[1].Kind != OpReg {
		return true, false, fmt.Errorf("%s %s expects X/m128, X: %q", c.goarch, baseOp, ins.Raw)
	}
	validRegister := func(arg Operand, allowMemory bool) bool {
		if vector {
			return c.isGoVEXVectorRegister(arg, 16, allowMemory)
		}
		if allowMemory && isAMD64MemoryOperand(arg) {
			return true
		}
		return arg.Kind == OpReg && c.isGoLegacyXReg(arg.Reg)
	}
	if !validRegister(ins.Args[0], true) || !validRegister(ins.Args[1], false) {
		return true, false, fmt.Errorf("%s %s requires an in-range X/m128 source and X destination: %q", c.goarch, baseOp, ins.Raw)
	}
	bytesValue, err := c.loadPackedCompareBytes(ins.Args[0], 16)
	if err != nil {
		return true, false, err
	}
	words := c.bitcastVectorBytesToIntegerLanes(16, 8, 16, bytesValue)
	result := c.emitPackedUnsignedWordMinimumPosition(words)
	out := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = bitcast <8 x i16> %s to <16 x i8>\n", out, result)
	return true, false, c.storeX(ins.Args[1].Reg, "%"+out)
}

func (c *amd64Ctx) emitPackedUnsignedWordMinimumPosition(words string) string {
	minimum := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = extractelement <8 x i16> %s, i32 0\n", minimum, words)
	minimumValue := "%" + minimum
	indexValue := "0"
	for lane := 1; lane < 8; lane++ {
		value := c.newTmp()
		less := c.newTmp()
		selectedMinimum := c.newTmp()
		selectedIndex := c.newTmp()
		fmt.Fprintf(c.b, "  %%%s = extractelement <8 x i16> %s, i32 %d\n", value, words, lane)
		fmt.Fprintf(c.b, "  %%%s = icmp ult i16 %%%s, %s\n", less, value, minimumValue)
		fmt.Fprintf(c.b, "  %%%s = select i1 %%%s, i16 %%%s, i16 %s\n", selectedMinimum, less, value, minimumValue)
		fmt.Fprintf(c.b, "  %%%s = select i1 %%%s, i16 %d, i16 %s\n", selectedIndex, less, lane, indexValue)
		minimumValue = "%" + selectedMinimum
		indexValue = "%" + selectedIndex
	}
	withMinimum := c.newTmp()
	withIndex := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = insertelement <8 x i16> zeroinitializer, i16 %s, i32 0\n", withMinimum, minimumValue)
	fmt.Fprintf(c.b, "  %%%s = insertelement <8 x i16> %%%s, i16 %s, i32 1\n", withIndex, withMinimum, indexValue)
	return "%" + withIndex
}
