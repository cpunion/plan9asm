package plan9asm

import (
	"fmt"
	"strings"
)

func (c *arm64Ctx) lowerData(op Op, postInc bool, ins Instr) (ok bool, terminated bool, err error) {
	if ok, terminated, err := c.lowerARM64RegisterAddress(op, ins); ok {
		return ok, terminated, err
	}
	if ok, terminated, err := c.lowerARM64ScalarMemoryWriteback(op, ins); ok {
		return ok, terminated, err
	}
	if ok, terminated, err := c.lowerARM64SymbolScalarMove(op, ins); ok {
		return ok, terminated, err
	}
	if ok, terminated, err := c.lowerARM64ScalarExtend(op, ins); ok {
		return ok, terminated, err
	}
	if ok, terminated, err := c.lowerARM64IntegerPair(op, ins); ok {
		return ok, terminated, err
	}
	switch op {
	case "MOVK", "MOVKW":
		return true, false, c.lowerARM64MoveKeep(op, ins)

	case "MOVD":
		if len(ins.Args) != 2 {
			return true, false, fmt.Errorf("arm64 MOVD expects 2 operands: %q", ins.Raw)
		}
		src, dst := ins.Args[0], ins.Args[1]
		v, err := c.eval64(src, postInc)
		if err != nil {
			return true, false, err
		}
		switch dst.Kind {
		case OpReg:
			return true, false, c.storeReg(dst.Reg, v)
		case OpMem:
			return true, false, c.storeMem(dst.Mem, 64, postInc, v)
		case OpFP:
			return true, false, c.storeFPResult64(dst.FPOffset, v)
		default:
			return true, false, nil
		}

	case "MOVB", "MOVBU", "MOVH", "MOVHU", "MOVW", "MOVWU":
		bits := 8
		if op == "MOVH" || op == "MOVHU" {
			bits = 16
		} else if op == "MOVW" || op == "MOVWU" {
			bits = 32
		}
		signed := op == "MOVB" || op == "MOVH" || op == "MOVW"
		return true, false, c.lowerNarrowMove(op, ins, bits, signed, postInc)

	case "MOVBW", "MOVHW":
		bits := 8
		if op == "MOVHW" {
			bits = 16
		}
		return true, false, c.lowerSignedNarrowLoadToWord(op, ins, bits, postInc)
	}
	return false, false, nil
}

func (c *arm64Ctx) lowerARM64MoveKeep(op Op, ins Instr) error {
	if strings.ToUpper(string(ins.Op)) != string(op) || len(ins.Args) != 2 ||
		ins.Args[0].Kind != OpImm || ins.Args[0].ImmRaw != "" ||
		ins.Args[1].Kind != OpReg || !isARM64GeneralOrZeroReg(ins.Args[1].Reg) {
		return fmt.Errorf("arm64 %s expects $imm16[<<16...], dstReg: %q", op, ins.Raw)
	}
	if ins.Args[0].Imm == 0 {
		return fmt.Errorf("arm64 %s cannot encode a zero immediate: %q", op, ins.Raw)
	}
	value := uint64(ins.Args[0].Imm)
	width := 64
	if op == "MOVKW" {
		if ins.Args[0].Imm < 0 || value > uint64(^uint32(0)) {
			return fmt.Errorf("arm64 MOVKW immediate is outside 32 bits: %q", ins.Raw)
		}
		width = 32
	}
	shift := -1
	for candidate := 0; candidate < width; candidate += 16 {
		fieldMask := uint64(0xffff) << candidate
		if value & ^fieldMask == 0 {
			shift = candidate
			break
		}
	}
	if shift < 0 {
		return fmt.Errorf("arm64 %s immediate is not a shifted unsigned 16-bit value: %q", op, ins.Raw)
	}
	old, err := c.loadReg(ins.Args[1].Reg)
	if err != nil {
		return err
	}
	if width == 64 {
		keepMask := ^(uint64(0xffff) << shift)
		kept := c.newTmp()
		fmt.Fprintf(c.b, "  %%%s = and i64 %s, %d\n", kept, old, keepMask)
		merged := c.newTmp()
		fmt.Fprintf(c.b, "  %%%s = or i64 %%%s, %d\n", merged, kept, value)
		return c.storeReg(ins.Args[1].Reg, "%"+merged)
	}
	old32 := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = trunc i64 %s to i32\n", old32, old)
	keepMask := ^(uint32(0xffff) << shift)
	kept := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = and i32 %%%s, %d\n", kept, old32, keepMask)
	merged := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = or i32 %%%s, %d\n", merged, kept, uint32(value))
	wide := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = zext i32 %%%s to i64\n", wide, merged)
	return c.storeReg(ins.Args[1].Reg, "%"+wide)
}

func (c *arm64Ctx) lowerSignedNarrowLoadToWord(op Op, ins Instr, bits int, postInc bool) error {
	if len(ins.Args) != 2 || ins.Args[0].Kind != OpMem || ins.Args[1].Kind != OpReg || !isARM64GeneralOrZeroReg(ins.Args[1].Reg) {
		return fmt.Errorf("arm64 %s expects memory, C_ZREG: %q", op, ins.Raw)
	}
	value, err := c.loadMem(ins.Args[0].Mem, bits, postInc)
	if err != nil {
		return err
	}
	narrow := c.newTmp()
	signed := c.newTmp()
	wide := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = trunc i64 %s to i%d\n", narrow, value, bits)
	fmt.Fprintf(c.b, "  %%%s = sext i%d %%%s to i32\n", signed, bits, narrow)
	fmt.Fprintf(c.b, "  %%%s = zext i32 %%%s to i64\n", wide, signed)
	return c.storeReg(ins.Args[1].Reg, "%"+wide)
}

func (c *arm64Ctx) lowerNarrowMove(op Op, ins Instr, bits int, signed, postInc bool) error {
	if handled, _, err := c.lowerARM64SymbolScalarMove(op, ins); handled {
		return err
	}
	if len(ins.Args) != 2 {
		return fmt.Errorf("arm64 %s expects 2 operands: %q", op, ins.Raw)
	}
	src, dst := ins.Args[0], ins.Args[1]
	var value string
	var err error
	switch src.Kind {
	case OpMem:
		value, err = c.loadMem(src.Mem, bits, postInc)
	default:
		value, err = c.eval64(src, false)
	}
	if err != nil {
		return err
	}

	switch dst.Kind {
	case OpReg:
		// A MOVW immediate is materialized through the 32-bit register view,
		// which clears the upper half even though register and memory sources
		// for MOVW are sign-extended.
		value = c.arm64ExtendNarrow(value, bits, signed && src.Kind != OpImm)
		return c.storeReg(dst.Reg, value)
	case OpMem:
		return c.storeMem(dst.Mem, bits, postInc, value)
	case OpFP:
		value = c.arm64ExtendNarrow(value, bits, signed)
		return c.storeFPResult64(dst.FPOffset, value)
	default:
		return fmt.Errorf("arm64 %s unsupported destination: %q", op, ins.Raw)
	}
}

func (c *arm64Ctx) arm64ExtendNarrow(value string, bits int, signed bool) string {
	narrow := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = trunc i64 %s to i%d\n", narrow, value, bits)
	extended := c.newTmp()
	extendOp := "zext"
	if signed {
		extendOp = "sext"
	}
	fmt.Fprintf(c.b, "  %%%s = %s i%d %%%s to i64\n", extended, extendOp, bits, narrow)
	return "%" + extended
}
