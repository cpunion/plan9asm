package plan9asm

import (
	"fmt"
	"strings"
)

// lowerVariableBlend implements Go 1.27's complete yblendvpd and _yvblendvpd
// families. Each opcode tests the high bit of its mask lane: false selects the
// base/destination lane and true selects the register-or-memory source lane.
func (c *amd64Ctx) lowerVariableBlend(op Op, ins Instr) (ok bool, terminated bool, err error) {
	rawOp := strings.ToUpper(string(op))
	baseOp := rawOp
	if dot := strings.IndexByte(rawOp, '.'); dot >= 0 {
		baseOp = rawOp[:dot]
	}
	laneBits := 0
	vectorForm := false
	switch baseOp {
	case "BLENDVPS":
		laneBits = 32
	case "BLENDVPD":
		laneBits = 64
	case "PBLENDVB":
		laneBits = 8
	case "VBLENDVPS":
		laneBits, vectorForm = 32, true
	case "VBLENDVPD":
		laneBits, vectorForm = 64, true
	case "VPBLENDVB":
		laneBits, vectorForm = 8, true
	default:
		return false, false, nil
	}
	if rawOp != baseOp {
		return true, false, fmt.Errorf("amd64 %s does not accept instruction suffixes: %q", baseOp, ins.Raw)
	}
	wantArgs := 3
	if vectorForm {
		wantArgs = 4
	}
	if len(ins.Args) != wantArgs {
		return true, false, fmt.Errorf("amd64 %s expects mask, source, %s: %q", baseOp, map[bool]string{false: "X destination", true: "base, destination"}[vectorForm], ins.Raw)
	}

	maskArg, source := ins.Args[0], ins.Args[1]
	base, destination := ins.Args[2], ins.Args[2]
	if vectorForm {
		destination = ins.Args[3]
	}
	if maskArg.Kind != OpReg || base.Kind != OpReg || destination.Kind != OpReg {
		return true, false, fmt.Errorf("amd64 %s expects vector mask/base/destination registers: %q", baseOp, ins.Raw)
	}
	byteWidth := amd64VectorByteWidth(base.Reg)
	if !vectorForm {
		if maskArg.Reg != Reg("X0") || byteWidth != 16 || destination.Reg != base.Reg || !c.isGoLegacyXReg(base.Reg) {
			return true, false, fmt.Errorf("%s %s expects implicit X0 mask and an in-range X destination: %q", c.goarch, baseOp, ins.Raw)
		}
	} else {
		if byteWidth != 16 && byteWidth != 32 {
			return true, false, fmt.Errorf("amd64 %s only accepts X or Y forms: %q", baseOp, ins.Raw)
		}
		if !amd64VEXVectorRegister(maskArg, byteWidth) || !amd64VEXVectorRegister(base, byteWidth) || !amd64VEXVectorRegister(destination, byteWidth) {
			return true, false, fmt.Errorf("amd64 %s operands must be same-width VEX registers: %q", baseOp, ins.Raw)
		}
	}
	if source.Kind == OpReg {
		valid := false
		if vectorForm {
			valid = amd64VEXVectorRegister(source, byteWidth)
		} else {
			valid = c.isGoLegacyXReg(source.Reg)
		}
		if !valid {
			return true, false, fmt.Errorf("%s %s source register has the wrong width or range: %q", c.goarch, baseOp, ins.Raw)
		}
	} else if !isAMD64MemoryOperand(source) {
		return true, false, fmt.Errorf("amd64 %s source must be a same-width vector register or memory: %q", baseOp, ins.Raw)
	}

	maskBytes, err := c.loadPackedCompareBytes(maskArg, byteWidth)
	if err != nil {
		return true, false, err
	}
	sourceBytes, err := c.loadPackedCompareBytes(source, byteWidth)
	if err != nil {
		return true, false, err
	}
	baseBytes, err := c.loadPackedCompareBytes(base, byteWidth)
	if err != nil {
		return true, false, err
	}
	lanes := byteWidth * 8 / laneBits
	maskLanes := c.bitcastVectorBytesToIntegerLanes(byteWidth, lanes, laneBits, maskBytes)
	sourceLanes := c.bitcastVectorBytesToIntegerLanes(byteWidth, lanes, laneBits, sourceBytes)
	baseLanes := c.bitcastVectorBytesToIntegerLanes(byteWidth, lanes, laneBits, baseBytes)
	selectedMask := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = icmp slt <%d x i%d> %s, zeroinitializer\n", selectedMask, lanes, laneBits, maskLanes)
	result := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = select <%d x i1> %%%s, <%d x i%d> %s, <%d x i%d> %s\n", result, lanes, selectedMask, lanes, laneBits, sourceLanes, lanes, laneBits, baseLanes)
	out := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = bitcast <%d x i%d> %%%s to <%d x i8>\n", out, lanes, laneBits, result, byteWidth)
	return true, false, c.storeVectorBytes(destination.Reg, byteWidth, "%"+out)
}
