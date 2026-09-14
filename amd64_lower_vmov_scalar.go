package plan9asm

import (
	"fmt"
	"strings"
)

// lowerVectorScalarMove implements all Go 1.27 _yvmovsd forms shared by
// VMOVSD and VMOVSS: scalar loads/stores, three-register upper merges, and
// their EVEX masked variants.
func (c *amd64Ctx) lowerVectorScalarMove(op Op, ins Instr) (ok bool, terminated bool, err error) {
	rawOp := strings.ToUpper(string(op))
	baseOp := rawOp
	suffix := ""
	if dot := strings.IndexByte(rawOp, '.'); dot >= 0 {
		baseOp, suffix = rawOp[:dot], rawOp[dot+1:]
	}
	elemBits := 0
	switch baseOp {
	case "VMOVSD":
		elemBits = 64
	case "VMOVSS":
		elemBits = 32
	default:
		return false, false, nil
	}
	if c.goarch != "amd64" {
		return true, false, fmt.Errorf("%s %s is absent from Go 1.27's 386 assembler forms: %q", c.goarch, baseOp, ins.Raw)
	}
	if suffix != "" && suffix != "Z" {
		return true, false, fmt.Errorf("amd64 %s suffix is absent from Go 1.27's _yvmovsd table: %q", baseOp, ins.Raw)
	}
	zeroing := suffix == "Z"
	switch len(ins.Args) {
	case 2:
		if zeroing {
			return true, false, fmt.Errorf("amd64 %s zeroing requires a K1-K7 mask and register destination: %q", baseOp, ins.Raw)
		}
		return c.lowerVectorScalarMoveTwoOperand(baseOp, elemBits, ins)
	case 3:
		if amd64ScalarXOperand(ins.Args[0]) && amd64ScalarXOperand(ins.Args[1]) && amd64ScalarXOperand(ins.Args[2]) {
			if zeroing {
				return true, false, fmt.Errorf("amd64 %s zeroing requires a K1-K7 mask: %q", baseOp, ins.Raw)
			}
			return c.lowerVectorScalarMoveRegisterMerge(elemBits, ins.Args[0], ins.Args[1], Operand{}, ins.Args[2], false)
		}
		return c.lowerVectorScalarMoveMaskedMemory(baseOp, elemBits, zeroing, ins)
	case 4:
		if !amd64ScalarXOperand(ins.Args[0]) || !amd64ScalarXOperand(ins.Args[1]) || !amd64ScalarXOperand(ins.Args[3]) {
			return true, false, fmt.Errorf("amd64 %s four-operand form expects X, X, K1-K7, X: %q", baseOp, ins.Raw)
		}
		if !amd64NonzeroKOperand(ins.Args[2]) {
			return true, false, fmt.Errorf("amd64 %s masked form expects K1-K7: %q", baseOp, ins.Raw)
		}
		return c.lowerVectorScalarMoveRegisterMerge(elemBits, ins.Args[0], ins.Args[1], ins.Args[2], ins.Args[3], zeroing)
	default:
		return true, false, fmt.Errorf("amd64 %s operand list is absent from Go 1.27's _yvmovsd table: %q", baseOp, ins.Raw)
	}
}

func (c *amd64Ctx) lowerVectorScalarMoveTwoOperand(baseOp string, elemBits int, ins Instr) (bool, bool, error) {
	src, dst := ins.Args[0], ins.Args[1]
	if amd64ScalarXOperand(src) && isAMD64MemoryOperand(dst) {
		value, err := c.loadXLowInteger(src.Reg, elemBits)
		if err != nil {
			return true, false, err
		}
		return true, false, c.storeVectorScalarMemory(dst, elemBits, value, "")
	}
	if isAMD64MemoryOperand(src) && amd64ScalarXOperand(dst) {
		value, err := c.evalIntSized(src, amd64IntegerTypeForBits(elemBits))
		if err != nil {
			return true, false, err
		}
		return true, false, c.storeVectorScalarRegister(dst.Reg, elemBits, value, "zeroinitializer", "", false)
	}
	return true, false, fmt.Errorf("amd64 %s two-operand form requires X-to-memory or memory-to-X: %q", baseOp, ins.Raw)
}

func (c *amd64Ctx) lowerVectorScalarMoveMaskedMemory(baseOp string, elemBits int, zeroing bool, ins Instr) (bool, bool, error) {
	if !amd64NonzeroKOperand(ins.Args[1]) {
		return true, false, fmt.Errorf("amd64 %s three-operand masked memory form expects K1-K7 in the middle: %q", baseOp, ins.Raw)
	}
	mask, err := c.loadK(ins.Args[1].Reg)
	if err != nil {
		return true, false, err
	}
	if amd64ScalarXOperand(ins.Args[0]) && isAMD64MemoryOperand(ins.Args[2]) {
		if zeroing {
			return true, false, fmt.Errorf("amd64 %s cannot use .Z with a memory destination: %q", baseOp, ins.Raw)
		}
		value, err := c.loadXLowInteger(ins.Args[0].Reg, elemBits)
		if err != nil {
			return true, false, err
		}
		return true, false, c.storeVectorScalarMemory(ins.Args[2], elemBits, value, mask)
	}
	if isAMD64MemoryOperand(ins.Args[0]) && amd64ScalarXOperand(ins.Args[2]) {
		value, err := c.evalIntSized(ins.Args[0], amd64IntegerTypeForBits(elemBits))
		if err != nil {
			return true, false, err
		}
		return true, false, c.storeVectorScalarRegister(ins.Args[2].Reg, elemBits, value, "zeroinitializer", mask, zeroing)
	}
	return true, false, fmt.Errorf("amd64 %s masked memory form requires X, K, memory or memory, K, X: %q", baseOp, ins.Raw)
}

func (c *amd64Ctx) lowerVectorScalarMoveRegisterMerge(elemBits int, lowSource, upperSource, maskArg, dst Operand, zeroing bool) (bool, bool, error) {
	low, err := c.loadXLowInteger(lowSource.Reg, elemBits)
	if err != nil {
		return true, false, err
	}
	upperBytes, err := c.loadX(upperSource.Reg)
	if err != nil {
		return true, false, err
	}
	lanes := 128 / elemBits
	upper := c.bitcastVectorBytesToIntegerLanes(16, lanes, elemBits, upperBytes)
	mask := ""
	if maskArg.Kind == OpReg {
		mask, err = c.loadK(maskArg.Reg)
		if err != nil {
			return true, false, err
		}
	}
	return true, false, c.storeVectorScalarRegister(dst.Reg, elemBits, low, upper, mask, zeroing)
}

func amd64ScalarXOperand(op Operand) bool {
	if op.Kind != OpReg {
		return false
	}
	index, ok := amd64ParseXReg(op.Reg)
	return ok && index < 32
}

func amd64NonzeroKOperand(op Operand) bool {
	if op.Kind != OpReg {
		return false
	}
	index, ok := amd64ParseKReg(op.Reg)
	return ok && index != 0
}

func (c *amd64Ctx) loadXLowInteger(reg Reg, elemBits int) (string, error) {
	bytesValue, err := c.loadX(reg)
	if err != nil {
		return "", err
	}
	lanes := 128 / elemBits
	values := c.bitcastVectorBytesToIntegerLanes(16, lanes, elemBits, bytesValue)
	low := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = extractelement <%d x i%d> %s, i32 0\n", low, lanes, elemBits, values)
	return "%" + low, nil
}

func (c *amd64Ctx) storeVectorScalarRegister(dst Reg, elemBits int, low, base, mask string, zeroing bool) error {
	lanes := 128 / elemBits
	if mask != "" {
		fallback := "0"
		if !zeroing {
			old, err := c.loadXLowInteger(dst, elemBits)
			if err != nil {
				return err
			}
			fallback = old
		}
		maskBit := amd64MaskBitI1(c, mask, 0)
		selected := c.newTmp()
		fmt.Fprintf(c.b, "  %%%s = select i1 %s, i%d %s, i%d %s\n", selected, maskBit, elemBits, low, elemBits, fallback)
		low = "%" + selected
	}
	updated := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = insertelement <%d x i%d> %s, i%d %s, i32 0\n", updated, lanes, elemBits, base, elemBits, low)
	bytesValue := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = bitcast <%d x i%d> %%%s to <16 x i8>\n", bytesValue, lanes, elemBits, updated)
	return c.storeX(dst, "%"+bytesValue)
}

func (c *amd64Ctx) storeVectorScalarMemory(dst Operand, elemBits int, value, mask string) error {
	if mask != "" {
		old, err := c.evalIntSized(dst, amd64IntegerTypeForBits(elemBits))
		if err != nil {
			return err
		}
		maskBit := amd64MaskBitI1(c, mask, 0)
		selected := c.newTmp()
		fmt.Fprintf(c.b, "  %%%s = select i1 %s, i%d %s, i%d %s\n", selected, maskBit, elemBits, value, elemBits, old)
		value = "%" + selected
	}
	switch dst.Kind {
	case OpMem:
		ptr, ptrType, err := c.ptrFromMem(dst.Mem)
		if err != nil {
			return err
		}
		fmt.Fprintf(c.b, "  store i%d %s, %s %s, align 1\n", elemBits, value, ptrType, ptr)
		return nil
	case OpSym:
		ptr, err := c.ptrFromSB(dst.Sym)
		if err != nil {
			return err
		}
		fmt.Fprintf(c.b, "  store i%d %s, ptr %s, align 1\n", elemBits, value, ptr)
		return nil
	case OpFP:
		return c.storeFPResult(dst.FPOffset, amd64IntegerTypeForBits(elemBits), value)
	default:
		return fmt.Errorf("expected scalar memory destination, got %s", dst.String())
	}
}
