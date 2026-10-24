package plan9asm

import (
	"fmt"
	"strings"
)

func (c *amd64Ctx) lowerPackedLowMultiply(op Op, ins Instr) (bool, bool, error) {
	form, known, err := x86PackedLowMultiplyProperties(op, ins, c.goarch)
	if !known || err != nil {
		return known, false, err
	}
	first, second, destination := ins.Args[form.first], ins.Args[form.second], ins.Args[form.destination]
	if form.legacy && form.memory >= 0 {
		virtual := first.Kind == OpFP
		if first.Kind == OpMem && !x86CallerMemoryCertified(ins, form.memory) && !c.fixedLocalMemoryCertified(ins, form.memory) {
			virtual = !x86OrdinaryNativeMemoryOperand(first, c.goarch)
		}
		if virtual {
			return true, false, fmt.Errorf("%w: legacy packed low multiply m128 needs original native storage, not virtual FP/SP/BP", ErrProbeNeedsContext)
		}
	}
	if form.memory >= 0 && first.Kind == OpFP {
		if err := c.validateX86PackedShiftFPSpan(first.FPOffset, form.memoryBytes); err != nil {
			return true, false, err
		}
	}
	bits := form.spec.laneBits
	lanes := form.byteWidth * 8 / bits
	b, err := c.loadPackedCompareLanes(second, form.byteWidth, bits, false)
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
		// Keep the original SSE4.1 opcode and its alignment/fault behavior,
		// including a dead result. An align-1 portable load is not equivalent.
		value, err = c.emitPackedBinaryNativeMemory(first, bits, b, "pmulld")
	} else {
		// Intel's E4 class suppresses inactive source lanes. The caller still
		// proves the entire possible read envelope independently of the mask.
		a, loadErr := c.loadX86PackedShiftLanes(first, form.byteWidth, bits, form.broadcast, mask)
		err = loadErr
		if err == nil {
			product := c.newTmp()
			fmt.Fprintf(c.b, "  %%%s = mul <%d x i%d> %s, %s\n", product, lanes, bits, b, a)
			value = "%" + product
		}
	}
	if err != nil {
		return true, false, err
	}
	if form.mask >= 0 {
		old := "zeroinitializer"
		if !form.zeroing {
			old, err = c.loadPackedCompareLanes(destination, form.byteWidth, bits, false)
			if err != nil {
				return true, false, err
			}
		}
		value = amd64ApplyIntegerLaneMask(c, lanes, bits, value, old, mask, form.zeroing)
	}
	out := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = bitcast <%d x i%d> %s to <%d x i8>\n", out, lanes, bits, value, form.byteWidth)
	if form.legacy {
		return true, false, c.storeLegacyPackedLogical(destination, "%"+out)
	}
	return true, false, c.storePackedMoveOperand(destination, form.byteWidth, "%"+out)
}

func isAMD64MemoryOperand(op Operand) bool {
	switch op.Kind {
	case OpFP, OpMem:
		return true
	case OpSym:
		// Go TYPE_ADDR is an address, not a TYPE_MEM read.
		return !strings.HasPrefix(strings.TrimSpace(op.Sym), "$")
	default:
		return false
	}
}
