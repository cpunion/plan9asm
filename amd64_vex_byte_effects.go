package plan9asm

import "strings"

// This is a typed dependency adapter, not a source/storage authority or an
// opcode whitelist. Each family retains its complete existing operand parser.
// The current consumer explicitly closes only non-masked VEX; EVEX storage,
// writemasks and broadcasts remain with their separately validated owners.
func x86UnmaskedVEXByteEffect(op Op, ins Instr, arch string) (x86TypedByteEffect, bool, error) {
	if !strings.HasPrefix(strings.ToUpper(string(op)), "V") {
		return x86TypedByteEffect{}, false, nil
	}
	if form, known, err := x86PackedIntegerAddProperties(ins, arch); known {
		return x86UnmaskedVEXBinaryEffect(ins, form.x86PackedBinaryForm, true, err)
	}
	if form, known, err := x86PackedSubtractProperties(ins, arch); known {
		return x86UnmaskedVEXBinaryEffect(ins, form.x86PackedBinaryForm, true, err)
	}
	if form, known, err := x86PackedBinaryEffectProperties(op, ins, arch); known {
		return x86UnmaskedVEXBinaryEffect(ins, form.x86PackedBinaryForm, form.vex, err)
	}
	if form, known, err := x86IntegerMinMaxProperties(op, ins, arch); known {
		return x86UnmaskedVEXBinaryEffect(ins, form.x86PackedBinaryForm, form.spec.laneBits <= 32, err)
	}
	if form, known, err := x86PackedLogicalProperties(ins, arch); known {
		return x86UnmaskedVEXBinaryEffect(ins, form.x86PackedBinaryForm,
			form.spec.encoding != x86PackedLogicalEVEX, err)
	}
	if form, known, err := x86ByteShuffleProperties(op, ins, arch); known {
		return x86UnmaskedVEXBinaryEffect(ins, form, true, err)
	}
	if form, known, err := x86PackedAlignProperties(op, ins, arch); known {
		return x86UnmaskedVEXBinaryEffect(ins, form.x86PackedBinaryForm,
			form.spec.encoding != x86PackedLogicalEVEX, err)
	}
	if form, known, err := x86ImmediateBlendProperties(op, ins, arch); known {
		binary := x86PackedBinaryForm{
			byteWidth: form.byteWidth, first: form.source, second: form.base, destination: form.destination,
			legacy: form.legacy, mask: -1, memory: -1,
		}
		if form.memory {
			binary.memory, binary.memoryBytes = form.source, form.byteWidth
		}
		return x86UnmaskedVEXBinaryEffect(ins, binary, true, err)
	}
	if form, known, err := x86Permute128Properties(op, ins, arch); known {
		return x86UnmaskedVEXBinaryEffect(ins, form.x86PackedBinaryForm, true, err)
	}
	if form, known, err := x86PackedShiftProperties(op, ins, arch); known {
		return x86UnmaskedVEXShiftEffect(ins, form, err)
	}
	return x86TypedByteEffect{}, false, nil
}

// VEX has only the low physical X/Y bank. This check does not infer an
// encoding from the opcode prefix: opcode-specific EVEX-only forms are also
// excluded by the validated family descriptor passed by each adapter.
func x86LowVEXVectorOperands(ins Instr) bool {
	for _, operand := range ins.Args {
		if operand.Kind != OpReg {
			continue
		}
		width := amd64VectorByteWidth(operand.Reg)
		if width > 32 || !amd64VEXVectorRegister(operand, width) {
			return false
		}
	}
	return true
}

func x86UnmaskedVEXBinaryEffect(ins Instr, form x86PackedBinaryForm, vex bool, err error) (x86TypedByteEffect, bool, error) {
	if err != nil {
		return x86TypedByteEffect{}, true, err
	}
	if !vex || form.legacy || form.mmx || form.mask >= 0 || form.broadcast || form.zeroing || !x86LowVEXVectorOperands(ins) {
		return x86TypedByteEffect{}, true, goAMD64FrameContext("vector effect requires its separate non-VEX consumer: %s", ins.Op)
	}
	return x86PackedBinaryDependencyEffect(ins, form, ins.Args[form.first]), true, nil
}

func x86UnmaskedVEXShiftEffect(ins Instr, form x86PackedShiftForm, err error) (x86TypedByteEffect, bool, error) {
	if err != nil {
		return x86TypedByteEffect{}, true, err
	}
	// The grammar's numerical dimensions distinguish VEX from the EVEX-only
	// word-per-lane and arithmetic-qword forms without a second opcode list.
	if form.legacy || form.mmx || form.mask >= 0 || form.broadcast || form.zeroing || !x86LowVEXVectorOperands(ins) ||
		form.arithmetic && form.laneBits == 64 || form.countMode == x86PackedShiftPerLane && form.laneBits == 16 ||
		form.countOperand.Kind == OpImm && form.memory == form.data {
		return x86TypedByteEffect{}, true, goAMD64FrameContext("shift effect requires its separate non-VEX consumer: %s", ins.Op)
	}
	width, countWidth := int64(form.byteWidth), int64(form.byteWidth)
	if form.countOperand.Kind == OpImm {
		countWidth = 1
	} else if form.countMode == x86PackedShiftUniform {
		countWidth = 16 // The full m128 access is observable, not just its low qword.
	}
	return x86TypedByteEffect{
		vector: x86TypedVectorEffect{x86TypedVectorUpperZero, 1},
		operands: []x86TypedByteOperand{
			{index: form.count, operand: form.countOperand, width: countWidth, read: true},
			{index: form.data, operand: ins.Args[form.data], width: width, read: true},
			{index: form.destination, operand: ins.Args[form.destination], width: width, write: true},
		},
	}, true, nil
}
