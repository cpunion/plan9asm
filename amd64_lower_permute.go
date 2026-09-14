package plan9asm

import (
	"fmt"
	"strings"
)

// lowerVariableDwordPermute implements the complete Go 1.27 _yvpermd
// operand family shared by VPERMD and VPERMPS.
func (c *amd64Ctx) lowerVariableDwordPermute(op Op, ins Instr) (ok bool, terminated bool, err error) {
	rawOp := strings.ToUpper(string(op))
	baseOp := rawOp
	suffix := ""
	if dot := strings.IndexByte(rawOp, '.'); dot >= 0 {
		baseOp, suffix = rawOp[:dot], rawOp[dot+1:]
	}
	switch baseOp {
	case "VPERMD", "VPERMPS":
	default:
		return false, false, nil
	}
	if c.goarch != "amd64" {
		return true, false, fmt.Errorf("%s %s is absent from Go 1.27's 386 assembler forms: %q", c.goarch, baseOp, ins.Raw)
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
		return true, false, fmt.Errorf("amd64 %s suffix is absent from Go 1.27's _yvpermd encodings: %q", baseOp, ins.Raw)
	}
	if len(ins.Args) != 3 && len(ins.Args) != 4 {
		return true, false, fmt.Errorf("amd64 %s expects data, indices, [K mask,] destination: %q", baseOp, ins.Raw)
	}
	dstArg := ins.Args[len(ins.Args)-1]
	if dstArg.Kind != OpReg {
		return true, false, fmt.Errorf("amd64 %s destination must be Y or Z: %q", baseOp, ins.Raw)
	}
	byteWidth := amd64VectorByteWidth(dstArg.Reg)
	if byteWidth != 32 && byteWidth != 64 {
		return true, false, fmt.Errorf("amd64 %s destination must be Y or Z: %q", baseOp, ins.Raw)
	}
	if !amd64EVEXVectorRegister(dstArg, byteWidth) || !amd64EVEXVectorRegister(ins.Args[1], byteWidth) {
		return true, false, fmt.Errorf("amd64 %s indices and destination must be matching Y or Z registers: %q", baseOp, ins.Raw)
	}
	if ins.Args[0].Kind == OpReg {
		if !amd64EVEXVectorRegister(ins.Args[0], byteWidth) {
			return true, false, fmt.Errorf("amd64 %s data register must match the destination width: %q", baseOp, ins.Raw)
		}
	} else if !isAMD64MemoryOperand(ins.Args[0]) {
		return true, false, fmt.Errorf("amd64 %s data must be a matching vector register or memory: %q", baseOp, ins.Raw)
	}
	if broadcast && !isAMD64MemoryOperand(ins.Args[0]) {
		return true, false, fmt.Errorf("amd64 %s.BCST requires a memory data source: %q", baseOp, ins.Raw)
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
		index, ok := amd64ParseKReg(maskArg.Reg)
		if !ok || index == 0 {
			return true, false, fmt.Errorf("amd64 %s masked form expects K1-K7: %q", baseOp, ins.Raw)
		}
		mask, err = c.loadK(maskArg.Reg)
		if err != nil {
			return true, false, err
		}
	}

	evex := masked || suffix != "" || byteWidth == 64 || !amd64VEXVectorRegister(dstArg, byteWidth) || !amd64VEXVectorRegister(ins.Args[1], byteWidth)
	if ins.Args[0].Kind == OpReg && !amd64VEXVectorRegister(ins.Args[0], byteWidth) {
		evex = true
	}
	if !evex && byteWidth != 32 {
		return true, false, fmt.Errorf("amd64 %s VEX form is Y-only in Go 1.27's _yvpermd table: %q", baseOp, ins.Raw)
	}

	lanes := byteWidth / 4
	var data string
	if broadcast {
		scalar, err := c.evalIntSized(ins.Args[0], I32)
		if err != nil {
			return true, false, err
		}
		data = amd64SplatInteger(c, lanes, 32, scalar)
	} else {
		dataBytes, err := c.loadPackedCompareBytes(ins.Args[0], byteWidth)
		if err != nil {
			return true, false, err
		}
		data = c.bitcastVectorBytesToIntegerLanes(byteWidth, lanes, 32, dataBytes)
	}
	indicesBytes, err := c.loadPackedCompareBytes(ins.Args[1], byteWidth)
	if err != nil {
		return true, false, err
	}
	indices := c.bitcastVectorBytesToIntegerLanes(byteWidth, lanes, 32, indicesBytes)
	result := c.emitVariableDwordPermute(data, indices, lanes)
	if masked {
		oldBytes, err := c.loadPackedCompareBytes(dstArg, byteWidth)
		if err != nil {
			return true, false, err
		}
		old := c.bitcastVectorBytesToIntegerLanes(byteWidth, lanes, 32, oldBytes)
		result = amd64ApplyIntegerLaneMask(c, lanes, 32, result, old, mask, zeroing)
	}
	out := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = bitcast <%d x i32> %s to <%d x i8>\n", out, lanes, result, byteWidth)
	return true, false, c.storeVectorBytes(dstArg.Reg, byteWidth, "%"+out)
}

func (c *amd64Ctx) emitVariableDwordPermute(data, indices string, lanes int) string {
	result := "zeroinitializer"
	for lane := 0; lane < lanes; lane++ {
		index := c.newTmp()
		fmt.Fprintf(c.b, "  %%%s = extractelement <%d x i32> %s, i32 %d\n", index, lanes, indices, lane)
		maskedIndex := c.newTmp()
		fmt.Fprintf(c.b, "  %%%s = and i32 %%%s, %d\n", maskedIndex, index, lanes-1)
		value := c.newTmp()
		fmt.Fprintf(c.b, "  %%%s = extractelement <%d x i32> %s, i32 %%%s\n", value, lanes, data, maskedIndex)
		inserted := c.newTmp()
		fmt.Fprintf(c.b, "  %%%s = insertelement <%d x i32> %s, i32 %%%s, i32 %d\n", inserted, lanes, result, value, lane)
		result = "%" + inserted
	}
	return result
}
