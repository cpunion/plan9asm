package plan9asm

import (
	"fmt"
	"strings"
)

func (c *amd64Ctx) lowerWordMultiplyAddFamily(op Op, ins Instr) (bool, bool, error) {
	base, _, _ := strings.Cut(strings.ToUpper(string(op)), ".")
	if _, known := x86WordMultiplyAddSpecs[Op(base)]; !known {
		return false, false, nil
	}
	if ins.Op != "" && strings.ToUpper(string(ins.Op)) != strings.ToUpper(string(op)) {
		return true, false, fmt.Errorf("packed word multiply-add dispatch/source opcode mismatch: %s/%s", op, ins.Op)
	}
	ins.Op = op
	form, recognized, err := x86WordMultiplyAddProperties(ins, c.goarch)
	if err != nil || !recognized {
		return recognized, false, err
	}
	first, second, destination := ins.Args[form.first], ins.Args[form.second], ins.Args[form.destination]
	if form.legacy && !form.mmx && form.memory >= 0 {
		virtual := first.Kind == OpFP
		if first.Kind == OpMem && !x86CallerMemoryCertified(ins, form.memory) && !c.fixedLocalMemoryCertified(ins, form.memory) {
			virtual = first.Mem.Base == SP || first.Mem.Base == BP || first.Mem.Index == SP || first.Mem.Index == BP
		}
		if virtual {
			return true, false, fmt.Errorf("%w: PMADDWL m128 needs original native storage, not virtual FP/SP/BP alignment", ErrProbeNeedsContext)
		}
	}
	if form.memory >= 0 && first.Kind == OpFP {
		if err := c.validateFPVectorSpan(first.FPOffset, form.memoryBytes); err != nil {
			return true, false, err
		}
	}
	if form.mmx {
		var a string
		if form.memory < 0 {
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
		result := c.emitPackedWordMultiplyAdd(4, c.bitcastI64ToIntegerLanes(4, 16, a), c.bitcastI64ToIntegerLanes(4, 16, b))
		bits := c.newTmp()
		fmt.Fprintf(c.b, "  %%%s = bitcast <2 x i32> %s to i64\n", bits, result)
		return true, false, c.storeReg(destination.Reg, "%"+bits)
	}
	b, err := c.loadPackedCompareLanes(second, form.byteWidth, 16, false)
	if err != nil {
		return true, false, err
	}
	var result string
	if form.legacy && form.memory >= 0 && first.Kind != OpFP {
		result, err = c.emitWordMultiplyAddNativeMemory(first, b)
	} else {
		// Intel SDM's E4NF.nb is non-fault-suppressing. A plain eager LLVM
		// load can disappear under a constant zero mask, so use the reviewed
		// complete volatile native reader, never logical masked loads.
		a, loadErr := c.loadPackedUnpackSource(first, x86PackedUnpackForm{
			x86PackedBinaryForm: form, spec: x86PackedUnpackSpec{laneBits: 16},
		})
		if loadErr != nil {
			return true, false, loadErr
		}
		result = c.emitPackedWordMultiplyAdd(form.byteWidth/2, a, b)
	}
	if err != nil {
		return true, false, err
	}
	lanes := form.byteWidth / 4
	if form.mask >= 0 {
		mask, err := c.loadK(ins.Args[form.mask].Reg)
		if err != nil {
			return true, false, err
		}
		old := "zeroinitializer"
		if !form.zeroing {
			old, err = c.loadPackedCompareLanes(destination, form.byteWidth, 32, false)
			if err != nil {
				return true, false, err
			}
		}
		result = amd64ApplyIntegerLaneMask(c, lanes, 32, result, old, mask, form.zeroing)
	}
	out := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = bitcast <%d x i32> %s to <%d x i8>\n", out, lanes, result, form.byteWidth)
	if form.legacy {
		return true, false, c.storeLegacyPackedLogical(destination, "%"+out)
	}
	return true, false, c.storePackedMoveOperand(destination, form.byteWidth, "%"+out)
}

func (c *amd64Ctx) emitWordMultiplyAddNativeMemory(source Operand, data string) (string, error) {
	pointer, pointerType, _, err := c.packedVectorMemoryPointer(source)
	if err != nil {
		return "", err
	}
	segment := ""
	switch pointerType {
	case "ptr addrspace(256)":
		segment = "%gs:"
	case "ptr addrspace(257)":
		segment = "%fs:"
	}
	carrier, value := c.newTmp(), c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = bitcast <8 x i16> %s to <4 x i32>\n", carrier, data)
	fmt.Fprintf(c.b, "  %%%s = call <4 x i32> asm sideeffect %q, %q(<4 x i32> %%%s, %s %s)\n",
		value, "pmaddwd "+segment+"($2), $0", "=x,0,r,~{memory}", carrier, pointerType, pointer)
	return "%" + value, nil
}
