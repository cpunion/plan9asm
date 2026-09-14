package plan9asm

import (
	"fmt"
	"strings"
)

// lowerPackedVectorMove implements the complete Go 1.27 packed vector-move
// tables. VMOVAPD/APS/UPD/UPS share _yvmovapd, while VMOVDQA32/DQA64 and
// VMOVDQU8/DQU16/DQU32/DQU64 share _yvmovdqa32. All have bit-preserving move
// semantics; the element width only controls the granularity of an EVEX mask.
func (c *amd64Ctx) lowerPackedVectorMove(op Op, ins Instr) (ok bool, terminated bool, err error) {
	rawOp := strings.ToUpper(string(op))
	baseOp := rawOp
	suffix := ""
	if dot := strings.IndexByte(rawOp, '.'); dot >= 0 {
		baseOp, suffix = rawOp[:dot], rawOp[dot+1:]
	}
	elemBits := 0
	switch baseOp {
	case "VMOVAPD", "VMOVUPD":
		elemBits = 64
	case "VMOVAPS", "VMOVUPS":
		elemBits = 32
	case "VMOVDQA32", "VMOVDQU32":
		elemBits = 32
	case "VMOVDQA64", "VMOVDQU64":
		elemBits = 64
	case "VMOVDQU8":
		elemBits = 8
	case "VMOVDQU16":
		elemBits = 16
	default:
		return false, false, nil
	}
	zeroing := false
	switch suffix {
	case "":
	case "Z":
		zeroing = true
	default:
		return true, false, fmt.Errorf("%s %s %s suffix is absent from its Go 1.27 packed-move encodings: %q", c.goarch, baseOp, suffix, ins.Raw)
	}
	if len(ins.Args) != 2 && len(ins.Args) != 3 {
		return true, false, fmt.Errorf("%s %s expects source, [K mask,] destination: %q", c.goarch, baseOp, ins.Raw)
	}
	masked := len(ins.Args) == 3
	if zeroing && !masked {
		return true, false, fmt.Errorf("%s %s zeroing requires a K1-K7 mask: %q", c.goarch, baseOp, ins.Raw)
	}
	if masked && !amd64NonzeroKOperand(ins.Args[1]) {
		return true, false, fmt.Errorf("%s %s masked form expects K1-K7 in the middle: %q", c.goarch, baseOp, ins.Raw)
	}

	src := ins.Args[0]
	dst := ins.Args[len(ins.Args)-1]
	byteWidth := 0
	switch {
	case src.Kind == OpReg:
		byteWidth = amd64VectorByteWidth(src.Reg)
		if byteWidth == 0 || !c.isGoPackedVectorMoveRegister(src, byteWidth) {
			return true, false, fmt.Errorf("%s %s source must be an in-range X, Y, or Z register: %q", c.goarch, baseOp, ins.Raw)
		}
		if dst.Kind == OpReg {
			if !c.isGoPackedVectorMoveRegister(dst, byteWidth) {
				return true, false, fmt.Errorf("%s %s register widths must match: %q", c.goarch, baseOp, ins.Raw)
			}
		} else if !isAMD64MemoryOperand(dst) {
			return true, false, fmt.Errorf("%s %s register source requires a matching vector or memory destination: %q", c.goarch, baseOp, ins.Raw)
		}
	case isAMD64MemoryOperand(src):
		if dst.Kind != OpReg {
			return true, false, fmt.Errorf("%s %s memory source requires a vector destination: %q", c.goarch, baseOp, ins.Raw)
		}
		byteWidth = amd64VectorByteWidth(dst.Reg)
		if byteWidth == 0 || !c.isGoPackedVectorMoveRegister(dst, byteWidth) {
			return true, false, fmt.Errorf("%s %s destination must be an in-range X, Y, or Z register: %q", c.goarch, baseOp, ins.Raw)
		}
	default:
		return true, false, fmt.Errorf("%s %s source must be a vector register or memory: %q", c.goarch, baseOp, ins.Raw)
	}

	value, err := c.loadPackedCompareBytes(src, byteWidth)
	if err != nil {
		return true, false, err
	}
	if !masked {
		return true, false, c.storeVectorBytesOperand(dst, byteWidth, value)
	}
	mask, err := c.loadK(ins.Args[1].Reg)
	if err != nil {
		return true, false, err
	}
	oldBytes, err := c.loadPackedCompareBytes(dst, byteWidth)
	if err != nil {
		return true, false, err
	}
	lanes := byteWidth * 8 / elemBits
	computed := c.bitcastVectorBytesToIntegerLanes(byteWidth, lanes, elemBits, value)
	old := c.bitcastVectorBytesToIntegerLanes(byteWidth, lanes, elemBits, oldBytes)
	// EVEX zeroing only changes a register destination. For a masked memory
	// store, inactive lanes remain untouched even though Go accepts the .Z
	// spelling for this table.
	result := amd64ApplyIntegerLaneMask(c, lanes, elemBits, computed, old, mask, zeroing && dst.Kind == OpReg)
	out := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = bitcast <%d x i%d> %s to <%d x i8>\n", out, lanes, elemBits, result, byteWidth)
	return true, false, c.storeVectorBytesOperand(dst, byteWidth, "%"+out)
}

func (c *amd64Ctx) isGoPackedVectorMoveRegister(arg Operand, byteWidth int) bool {
	if !amd64EVEXVectorRegister(arg, byteWidth) {
		return false
	}
	if c.goarch == "386" && byteWidth == 64 {
		index, _ := amd64VectorRegisterIndex(arg.Reg, byteWidth)
		return index < 8
	}
	return true
}
