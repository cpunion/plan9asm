package plan9asm

import (
	"fmt"
	"strings"
)

// lowerPackedAndNot implements every Go 1.27 operand form in the packed
// AND-NOT family: PANDN (MMX/XMM), VPANDN (X/Y VEX), and VPANDND/Q
// (X/Y/Z EVEX, including masking, zeroing, and memory broadcast).
func (c *amd64Ctx) lowerPackedAndNot(op Op, ins Instr) (ok bool, terminated bool, err error) {
	rawOp := strings.ToUpper(string(op))
	baseOp := rawOp
	suffix := ""
	if dot := strings.IndexByte(rawOp, '.'); dot >= 0 {
		baseOp, suffix = rawOp[:dot], rawOp[dot+1:]
	}
	switch baseOp {
	case "PANDN":
		if suffix != "" {
			return true, false, fmt.Errorf("amd64 PANDN does not accept instruction suffixes: %q", ins.Raw)
		}
		return c.lowerLegacyPackedAndNot(ins)
	case "VPANDN":
		if suffix != "" {
			return true, false, fmt.Errorf("amd64 VPANDN does not accept instruction suffixes: %q", ins.Raw)
		}
		return c.lowerVEXPackedAndNot(ins)
	case "VPANDND", "VPANDNQ":
		return c.lowerEVEXPackedLogical(op, ins)
	default:
		return false, false, nil
	}
}

func (c *amd64Ctx) lowerLegacyPackedAndNot(ins Instr) (bool, bool, error) {
	if len(ins.Args) != 2 || ins.Args[1].Kind != OpReg {
		return true, false, fmt.Errorf("amd64 PANDN expects MMX/m64, MMX or X/m128, X: %q", ins.Raw)
	}
	dst := ins.Args[1].Reg
	if _, ok := amd64ParseMReg(dst); ok {
		var src string
		var err error
		if ins.Args[0].Kind == OpReg {
			if _, ok := amd64ParseMReg(ins.Args[0].Reg); !ok {
				return true, false, fmt.Errorf("amd64 PANDN MMX form requires an MMX source: %q", ins.Raw)
			}
			src, err = c.loadReg(ins.Args[0].Reg)
		} else if isAMD64MemoryOperand(ins.Args[0]) {
			src, err = c.evalIntSized(ins.Args[0], I64)
		} else {
			return true, false, fmt.Errorf("amd64 PANDN MMX form requires MMX or memory source: %q", ins.Raw)
		}
		if err != nil {
			return true, false, err
		}
		old, err := c.loadReg(dst)
		if err != nil {
			return true, false, err
		}
		inverted := c.newTmp()
		result := c.newTmp()
		fmt.Fprintf(c.b, "  %%%s = xor i64 %s, -1\n", inverted, old)
		fmt.Fprintf(c.b, "  %%%s = and i64 %%%s, %s\n", result, inverted, src)
		return true, false, c.storeReg(dst, "%"+result)
	}
	if !isAMD64XReg(dst) {
		return true, false, fmt.Errorf("amd64 PANDN destination must be MMX or X: %q", ins.Raw)
	}
	if ins.Args[0].Kind == OpReg && !isAMD64XReg(ins.Args[0].Reg) {
		return true, false, fmt.Errorf("amd64 PANDN XMM form requires an X source: %q", ins.Raw)
	}
	src, err := c.loadXVecOperand(ins.Args[0])
	if err != nil {
		return true, false, err
	}
	old, err := c.loadX(dst)
	if err != nil {
		return true, false, err
	}
	result := c.emitPackedAndNotBytes(16, src, old)
	return true, false, c.storeX(dst, result)
}

func (c *amd64Ctx) lowerVEXPackedAndNot(ins Instr) (bool, bool, error) {
	if len(ins.Args) != 3 || ins.Args[2].Kind != OpReg {
		return true, false, fmt.Errorf("amd64 VPANDN expects X/m, X, X or Y/m, Y, Y: %q", ins.Raw)
	}
	dst := ins.Args[2].Reg
	byteWidth := amd64VectorByteWidth(dst)
	if byteWidth != 16 && byteWidth != 32 {
		return true, false, fmt.Errorf("amd64 VPANDN destination must be X or Y: %q", ins.Raw)
	}
	if !amd64VectorRegisterHasWidth(ins.Args[1], byteWidth) {
		return true, false, fmt.Errorf("amd64 VPANDN second source must match its destination width: %q", ins.Raw)
	}
	if ins.Args[0].Kind == OpReg && !amd64VectorRegisterHasWidth(ins.Args[0], byteWidth) {
		return true, false, fmt.Errorf("amd64 VPANDN first source must match its destination width: %q", ins.Raw)
	}
	first, err := c.loadPackedCompareBytes(ins.Args[0], byteWidth)
	if err != nil {
		return true, false, err
	}
	second, err := c.loadPackedCompareBytes(ins.Args[1], byteWidth)
	if err != nil {
		return true, false, err
	}
	result := c.emitPackedAndNotBytes(byteWidth, first, second)
	return true, false, c.storeVectorBytes(dst, byteWidth, result)
}

func (c *amd64Ctx) lowerEVEXPackedAndNot(baseOp, suffix string, ins Instr) (bool, bool, error) {
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
		return true, false, fmt.Errorf("amd64 %s accepts only .Z, .BCST, or .BCST.Z as enabled by its Go 1.27 optab: %q", baseOp, ins.Raw)
	}
	if len(ins.Args) != 3 && len(ins.Args) != 4 {
		return true, false, fmt.Errorf("amd64 %s expects src1, src2, [K mask,] destination: %q", baseOp, ins.Raw)
	}
	if ins.Args[len(ins.Args)-1].Kind != OpReg {
		return true, false, fmt.Errorf("amd64 %s expects a vector destination: %q", baseOp, ins.Raw)
	}
	dst := ins.Args[len(ins.Args)-1].Reg
	byteWidth := amd64VectorByteWidth(dst)
	if byteWidth == 0 {
		return true, false, fmt.Errorf("amd64 %s destination must be X, Y, or Z: %q", baseOp, ins.Raw)
	}
	if !amd64VectorRegisterHasWidth(ins.Args[1], byteWidth) {
		return true, false, fmt.Errorf("amd64 %s second source must match its destination width: %q", baseOp, ins.Raw)
	}
	if broadcast {
		if !isAMD64MemoryOperand(ins.Args[0]) {
			return true, false, fmt.Errorf("amd64 %s.BCST requires a memory first source: %q", baseOp, ins.Raw)
		}
	} else if ins.Args[0].Kind == OpReg && !amd64VectorRegisterHasWidth(ins.Args[0], byteWidth) {
		return true, false, fmt.Errorf("amd64 %s first source must match its destination width: %q", baseOp, ins.Raw)
	}

	masked := len(ins.Args) == 4
	if zeroing && !masked {
		return true, false, fmt.Errorf("amd64 %s zeroing requires a K1-K7 mask: %q", baseOp, ins.Raw)
	}
	var mask string
	if masked {
		maskArg := ins.Args[2]
		if maskArg.Kind != OpReg {
			return true, false, fmt.Errorf("amd64 %s masked form expects K1-K7: %q", baseOp, ins.Raw)
		}
		maskIndex, ok := amd64ParseKReg(maskArg.Reg)
		if !ok || maskIndex == 0 {
			return true, false, fmt.Errorf("amd64 %s masked form expects K1-K7: %q", baseOp, ins.Raw)
		}
		var err error
		mask, err = c.loadK(maskArg.Reg)
		if err != nil {
			return true, false, err
		}
	}

	laneBits := 32
	if baseOp == "VPANDNQ" {
		laneBits = 64
	}
	lanes := byteWidth * 8 / laneBits
	vecType := fmt.Sprintf("<%d x i%d>", lanes, laneBits)
	var first string
	var err error
	if broadcast {
		first, err = c.evalIntSized(ins.Args[0], amd64IntegerTypeForBits(laneBits))
		if err == nil {
			first = amd64SplatInteger(c, lanes, laneBits, first)
		}
	} else {
		var firstBytes string
		firstBytes, err = c.loadPackedCompareBytes(ins.Args[0], byteWidth)
		if err == nil {
			first = c.bitcastVectorBytesToIntegerLanes(byteWidth, lanes, laneBits, firstBytes)
		}
	}
	if err != nil {
		return true, false, err
	}
	secondBytes, err := c.loadPackedCompareBytes(ins.Args[1], byteWidth)
	if err != nil {
		return true, false, err
	}
	second := c.bitcastVectorBytesToIntegerLanes(byteWidth, lanes, laneBits, secondBytes)
	inverted := c.newTmp()
	computed := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = xor %s %s, %s\n", inverted, vecType, second, llvmSplatInteger(lanes, laneBits, ^uint64(0)))
	fmt.Fprintf(c.b, "  %%%s = and %s %%%s, %s\n", computed, vecType, inverted, first)
	result := "%" + computed
	if masked {
		oldBytes, err := c.loadPackedCompareBytes(Operand{Kind: OpReg, Reg: dst}, byteWidth)
		if err != nil {
			return true, false, err
		}
		old := c.bitcastVectorBytesToIntegerLanes(byteWidth, lanes, laneBits, oldBytes)
		result = amd64ApplyIntegerLaneMask(c, lanes, laneBits, result, old, mask, zeroing)
	}
	out := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = bitcast %s %s to <%d x i8>\n", out, vecType, result, byteWidth)
	return true, false, c.storeVectorBytes(dst, byteWidth, "%"+out)
}

func amd64VectorByteWidth(r Reg) int {
	switch {
	case isAMD64XReg(r):
		return 16
	case isAMD64YReg(r):
		return 32
	case isAMD64ZReg(r):
		return 64
	default:
		return 0
	}
}

func amd64VectorRegisterHasWidth(op Operand, byteWidth int) bool {
	return op.Kind == OpReg && amd64VectorByteWidth(op.Reg) == byteWidth
}

func (c *amd64Ctx) emitPackedAndNotBytes(byteWidth int, first, second string) string {
	inverted := c.newTmp()
	result := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = xor <%d x i8> %s, %s\n", inverted, byteWidth, second, llvmAllOnesI8Vec(byteWidth))
	fmt.Fprintf(c.b, "  %%%s = and <%d x i8> %%%s, %s\n", result, byteWidth, inverted, first)
	return "%" + result
}

func (c *amd64Ctx) bitcastVectorBytesToIntegerLanes(byteWidth, lanes, laneBits int, value string) string {
	cast := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = bitcast <%d x i8> %s to <%d x i%d>\n", cast, byteWidth, value, lanes, laneBits)
	return "%" + cast
}

func (c *amd64Ctx) storeVectorBytes(dst Reg, byteWidth int, value string) error {
	switch byteWidth {
	case 16:
		return c.storeX(dst, value)
	case 32:
		return c.storeY(dst, value)
	case 64:
		return c.storeZ(dst, value)
	default:
		return fmt.Errorf("unsupported packed AND-NOT vector width %d", byteWidth)
	}
}
