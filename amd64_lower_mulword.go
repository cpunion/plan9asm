package plan9asm

import (
	"fmt"
	"strings"
)

type amd64PackedWordMultiplyMode uint8

const (
	amd64PackedWordMultiplyLow amd64PackedWordMultiplyMode = iota
	amd64PackedWordMultiplySignedHigh
	amd64PackedWordMultiplyUnsignedHigh
	amd64PackedWordMultiplyRoundedSignedHigh
)

func (c *amd64Ctx) lowerPackedWordMultiply(op Op, ins Instr) (bool, bool, error) {
	form, recognized, err := x86PackedWordMultiplyProperties(op, ins, c.goarch)
	if !recognized || err != nil {
		return recognized, false, err
	}
	first, second, destination := form.source, ins.Args[form.second], ins.Args[form.destination]
	if form.legacy && !form.mmx && form.memory >= 0 {
		virtual := first.Kind == OpFP
		if first.Kind == OpMem && !x86CallerMemoryCertified(ins, form.memory) && !c.fixedLocalMemoryCertified(ins, form.memory) {
			virtual = first.Mem.Base == SP || first.Mem.Base == BP || first.Mem.Index == SP || first.Mem.Index == BP
		}
		if virtual {
			return true, false, fmt.Errorf("%w: legacy packed word multiply m128 needs original native storage, not virtual FP/SP/BP", ErrProbeNeedsContext)
		}
	}
	if form.memory >= 0 && first.Kind == OpFP {
		if err := c.validateFPVectorSpan(first.FPOffset, form.memoryBytes); err != nil {
			return true, false, err
		}
	}
	lanes := form.byteWidth / 2
	if form.mmx {
		// A GP source accepted through ycover encodes an MMX register; it
		// must never consume the unrelated GP shadow as its input value.
		if first.Kind == OpReg && first.Reg != ins.Args[form.first].Reg {
			return true, false, fmt.Errorf("%w: Go GP-spelled packed multiply source names the MMX bank", ErrProbeNeedsContext)
		}
		var a string
		if first.Kind == OpReg {
			a, err = c.loadReg(first.Reg)
		} else {
			a, err = c.loadPackedUnpackMMXSource(first, 8)
		}
		if err != nil {
			return true, false, err
		}
		b, err := c.loadReg(second.Reg)
		if err != nil {
			return true, false, err
		}
		value := c.emitPackedWordMultiply(lanes, c.bitcastI64ToIntegerLanes(lanes, 16, a), c.bitcastI64ToIntegerLanes(lanes, 16, b), form.spec.mode)
		out := c.newTmp()
		fmt.Fprintf(c.b, "  %%%s = bitcast <%d x i16> %s to i64\n", out, lanes, value)
		return true, false, c.storeReg(destination.Reg, "%"+out)
	}
	b, err := c.loadPackedCompareLanes(second, form.byteWidth, 16, false)
	if err != nil {
		return true, false, err
	}
	var mask, value string
	if form.mask >= 0 {
		mask, err = c.loadK(ins.Args[form.mask].Reg)
		if err != nil {
			return true, false, err
		}
	}
	if form.legacy && form.memory >= 0 {
		value, err = c.emitPackedBinaryNativeMemory(first, 16, b, strings.ToLower(string(op)))
	} else {
		// These four EVEX opcodes are E4.nb: masked-off word lanes do not
		// read native memory. The full possible envelope still belongs to
		// the independent source/caller proof, never to the mask value.
		a, loadErr := c.loadX86PackedShiftLanes(first, form.byteWidth, 16, false, mask)
		err = loadErr
		if err == nil {
			value = c.emitPackedWordMultiply(lanes, a, b, form.spec.mode)
		}
	}
	if err != nil {
		return true, false, err
	}
	if form.mask >= 0 {
		old := "zeroinitializer"
		if !form.zeroing {
			old, err = c.loadPackedCompareLanes(destination, form.byteWidth, 16, false)
			if err != nil {
				return true, false, err
			}
		}
		value = amd64ApplyIntegerLaneMask(c, lanes, 16, value, old, mask, form.zeroing)
	}
	out := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = bitcast <%d x i16> %s to <%d x i8>\n", out, lanes, value, form.byteWidth)
	if form.legacy {
		return true, false, c.storeLegacyPackedLogical(destination, "%"+out)
	}
	return true, false, c.storePackedMoveOperand(destination, form.byteWidth, "%"+out)
}

func (c *amd64Ctx) emitPackedWordMultiply(lanes int, first, second string, mode amd64PackedWordMultiplyMode) string {
	wordType := fmt.Sprintf("<%d x i16>", lanes)
	if mode == amd64PackedWordMultiplyLow {
		result := c.newTmp()
		fmt.Fprintf(c.b, "  %%%s = mul %s %s, %s\n", result, wordType, second, first)
		return "%" + result
	}

	wideType := fmt.Sprintf("<%d x i32>", lanes)
	extend := "sext"
	shift := int64(16)
	if mode == amd64PackedWordMultiplyUnsignedHigh {
		extend = "zext"
	}
	firstWide := c.newTmp()
	secondWide := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = %s %s %s to %s\n", firstWide, extend, wordType, first, wideType)
	fmt.Fprintf(c.b, "  %%%s = %s %s %s to %s\n", secondWide, extend, wordType, second, wideType)
	product := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = mul %s %%%s, %%%s\n", product, wideType, secondWide, firstWide)
	value := "%" + product
	if mode == amd64PackedWordMultiplyRoundedSignedHigh {
		rounded := c.newTmp()
		fmt.Fprintf(c.b, "  %%%s = add %s %s, %s\n", rounded, wideType, value, llvmSplatSignedInteger(lanes, 32, 16384))
		value = "%" + rounded
		shift = 15
	}
	shifted := c.newTmp()
	shiftOp := "ashr"
	if mode == amd64PackedWordMultiplyUnsignedHigh {
		shiftOp = "lshr"
	}
	fmt.Fprintf(c.b, "  %%%s = %s %s %s, %s\n", shifted, shiftOp, wideType, value, llvmSplatSignedInteger(lanes, 32, shift))
	value = "%" + shifted
	result := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = trunc %s %s to %s\n", result, wideType, value, wordType)
	return "%" + result
}
