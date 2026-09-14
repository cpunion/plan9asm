package plan9asm

import (
	"fmt"
	"strings"
)

var amd64PackedPopcountLaneBits = map[Op]int{
	"VPOPCNTB": 8,
	"VPOPCNTW": 16,
	"VPOPCNTD": 32,
	"VPOPCNTQ": 64,
}

// lowerPackedPopcount implements Go 1.27's complete _yvexpandpd form table
// for VPOPCNTB/W/D/Q: X/Y/Z register or memory sources, K1-K7 merge and .Z,
// and the D/Q scalar-memory broadcast forms.
func (c *amd64Ctx) lowerPackedPopcount(op Op, ins Instr) (ok bool, terminated bool, err error) {
	rawOp := strings.ToUpper(string(op))
	baseOp := rawOp
	suffix := ""
	if dot := strings.IndexByte(rawOp, '.'); dot >= 0 {
		baseOp, suffix = rawOp[:dot], rawOp[dot+1:]
	}
	laneBits, ok := amd64PackedPopcountLaneBits[Op(baseOp)]
	if !ok {
		return false, false, nil
	}
	broadcast, zeroing := false, false
	switch suffix {
	case "":
	case "Z":
		zeroing = true
	case "BCST":
		broadcast = true
	case "BCST.Z":
		broadcast, zeroing = true, true
	default:
		return true, false, fmt.Errorf("%s %s has a suffix outside Go 1.27's _yvexpandpd form: %q", c.goarch, baseOp, ins.Raw)
	}
	if broadcast && laneBits < 32 {
		return true, false, fmt.Errorf("%s %s has no broadcast form: %q", c.goarch, baseOp, ins.Raw)
	}
	if len(ins.Args) != 2 && len(ins.Args) != 3 {
		return true, false, fmt.Errorf("%s %s expects source, [K1-K7,] destination: %q", c.goarch, baseOp, ins.Raw)
	}
	masked := len(ins.Args) == 3
	if zeroing && !masked {
		return true, false, fmt.Errorf("%s %s.Z requires a K1-K7 mask: %q", c.goarch, baseOp, ins.Raw)
	}
	destination := ins.Args[len(ins.Args)-1]
	if destination.Kind != OpReg {
		return true, false, fmt.Errorf("%s %s requires an X, Y, or Z destination: %q", c.goarch, baseOp, ins.Raw)
	}
	byteWidth := amd64VectorByteWidth(destination.Reg)
	if byteWidth == 0 || !c.isGoEVEXVectorRegister(destination, byteWidth) {
		return true, false, fmt.Errorf("%s %s destination is outside Go's EVEX register range: %q", c.goarch, baseOp, ins.Raw)
	}
	source := ins.Args[0]
	if broadcast {
		if !isAMD64MemoryOperand(source) {
			return true, false, fmt.Errorf("%s %s.BCST requires a scalar memory source: %q", c.goarch, baseOp, ins.Raw)
		}
	} else if !isAMD64MemoryOperand(source) && !c.isGoEVEXVectorRegister(source, byteWidth) {
		return true, false, fmt.Errorf("%s %s source must be matching X/Y/Z register or memory: %q", c.goarch, baseOp, ins.Raw)
	}

	mask := ""
	if masked {
		maskArg := ins.Args[1]
		maskIndex, valid := amd64ParseKReg(maskArg.Reg)
		if maskArg.Kind != OpReg || !valid || maskIndex == 0 {
			return true, false, fmt.Errorf("%s %s masked form requires K1-K7: %q", c.goarch, baseOp, ins.Raw)
		}
		mask, err = c.loadK(maskArg.Reg)
		if err != nil {
			return true, false, err
		}
	}

	lanes := byteWidth * 8 / laneBits
	sourceValue, err := c.loadPackedCompareLanes(source, byteWidth, laneBits, broadcast)
	if err != nil {
		return true, false, err
	}
	computed := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = call <%d x i%d> @llvm.ctpop.v%di%d(<%d x i%d> %s)\n", computed, lanes, laneBits, lanes, laneBits, lanes, laneBits, sourceValue)
	result := "%" + computed
	if masked {
		oldBytes, err := c.loadPackedCompareBytes(destination, byteWidth)
		if err != nil {
			return true, false, err
		}
		old := c.bitcastVectorBytesToIntegerLanes(byteWidth, lanes, laneBits, oldBytes)
		result = amd64ApplyIntegerLaneMask(c, lanes, laneBits, result, old, mask, zeroing)
	}
	out := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = bitcast <%d x i%d> %s to <%d x i8>\n", out, lanes, laneBits, result, byteWidth)
	return true, false, c.storeVectorBytes(destination.Reg, byteWidth, "%"+out)
}
