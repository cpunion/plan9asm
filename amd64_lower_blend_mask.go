package plan9asm

import (
	"fmt"
	"strings"
)

var amd64MaskBlendLaneBits = map[string]int{
	"VPBLENDMB": 8,
	"VPBLENDMW": 16,
	"VPBLENDMD": 32,
	"VPBLENDMQ": 64,
}

// lowerMaskBlend implements Go 1.27's complete _yvblendmpd operand table for
// VPBLENDM{B,W,D,Q}. A set K bit selects the first Plan 9 (Intel r/m) source;
// an unset bit selects the second source, or zero for a .Z form.
func (c *amd64Ctx) lowerMaskBlend(op Op, ins Instr) (ok bool, terminated bool, err error) {
	rawOp := strings.ToUpper(string(op))
	baseOp := rawOp
	if dot := strings.IndexByte(rawOp, '.'); dot >= 0 {
		baseOp = rawOp[:dot]
	}
	laneBits, ok := amd64MaskBlendLaneBits[baseOp]
	if !ok {
		return false, false, nil
	}
	suffixes, err := amd64ParseTwoSourcePermuteSuffixes(rawOp, baseOp)
	if err != nil {
		return true, false, fmt.Errorf("%w: %q", err, ins.Raw)
	}
	if suffixes.broadcast && laneBits != 32 && laneBits != 64 {
		return true, false, fmt.Errorf("amd64 %s does not enable broadcast in Go 1.27: %q", baseOp, ins.Raw)
	}
	if len(ins.Args) != 3 && len(ins.Args) != 4 {
		return true, false, fmt.Errorf("amd64 %s expects first source, second source, [K mask,] destination: %q", baseOp, ins.Raw)
	}
	masked := len(ins.Args) == 4
	if masked && c.goarch == "386" {
		return true, false, fmt.Errorf("386 %s masked form exceeds the Go assembler frontend's operand limit: %q", baseOp, ins.Raw)
	}
	if suffixes.zeroing && !masked {
		return true, false, fmt.Errorf("amd64 %s .Z requires a K1-K7 mask: %q", baseOp, ins.Raw)
	}

	destination := ins.Args[len(ins.Args)-1]
	byteWidth := 0
	if destination.Kind == OpReg {
		byteWidth = amd64VectorByteWidth(destination.Reg)
	}
	if byteWidth == 0 || !c.isGoPackedVectorMoveRegister(destination, byteWidth) {
		return true, false, fmt.Errorf("amd64 %s expects an in-range X, Y, or Z destination: %q", baseOp, ins.Raw)
	}
	second := ins.Args[1]
	if !c.isGoPackedVectorMoveRegister(second, byteWidth) {
		return true, false, fmt.Errorf("amd64 %s second source must match the destination width: %q", baseOp, ins.Raw)
	}
	first := ins.Args[0]
	if first.Kind == OpReg {
		if !c.isGoPackedVectorMoveRegister(first, byteWidth) {
			return true, false, fmt.Errorf("amd64 %s first source register must match the destination width: %q", baseOp, ins.Raw)
		}
	} else if !isAMD64MemoryOperand(first) {
		return true, false, fmt.Errorf("amd64 %s first source must be a matching vector register or memory: %q", baseOp, ins.Raw)
	}
	if suffixes.broadcast && !isAMD64MemoryOperand(first) {
		return true, false, fmt.Errorf("amd64 %s.BCST requires a memory first source: %q", baseOp, ins.Raw)
	}

	firstValue, err := c.loadPackedCompareLanes(first, byteWidth, laneBits, suffixes.broadcast)
	if err != nil {
		return true, false, err
	}
	if !masked {
		out := c.newTmp()
		lanes := byteWidth * 8 / laneBits
		fmt.Fprintf(c.b, "  %%%s = bitcast <%d x i%d> %s to <%d x i8>\n", out, lanes, laneBits, firstValue, byteWidth)
		return true, false, c.storeVectorBytes(destination.Reg, byteWidth, "%"+out)
	}

	maskArg := ins.Args[2]
	if maskArg.Kind != OpReg {
		return true, false, fmt.Errorf("amd64 %s masked form expects K1-K7: %q", baseOp, ins.Raw)
	}
	maskIndex, validMask := amd64ParseKReg(maskArg.Reg)
	if !validMask || maskIndex == 0 {
		return true, false, fmt.Errorf("amd64 %s masked form expects K1-K7: %q", baseOp, ins.Raw)
	}
	mask, err := c.loadK(maskArg.Reg)
	if err != nil {
		return true, false, err
	}
	secondValue, err := c.loadPackedCompareLanes(second, byteWidth, laneBits, false)
	if err != nil {
		return true, false, err
	}
	lanes := byteWidth * 8 / laneBits
	result := amd64ApplyIntegerLaneMask(c, lanes, laneBits, firstValue, secondValue, mask, suffixes.zeroing)
	out := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = bitcast <%d x i%d> %s to <%d x i8>\n", out, lanes, laneBits, result, byteWidth)
	return true, false, c.storeVectorBytes(destination.Reg, byteWidth, "%"+out)
}
