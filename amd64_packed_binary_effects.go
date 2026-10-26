package plan9asm

import "strings"

// The six legacy binary families share a read/modify/write dependency. The
// caller supplies the validated physical source (including the MMX bank),
// rather than reinterpreting its original source spelling here.
func x86LegacyPackedBinaryByteEffect(ins Instr, form x86PackedBinaryForm, source Operand) x86TypedByteEffect {
	width := int64(form.byteWidth)
	sourceWidth := width
	if form.memory >= 0 {
		sourceWidth = int64(form.memoryBytes)
	}
	effect := x86TypedByteEffect{operands: []x86TypedByteOperand{
		{index: form.first, operand: source, width: sourceWidth, read: true},
		{index: form.destination, operand: ins.Args[form.destination], width: width, read: true, write: true},
	}}
	if form.legacy && !form.mmx {
		effect.vector = x86TypedVectorEffect{x86TypedVectorUpperPreserve, width}
	}
	return effect
}

// Keep the existing family parsers as the only opcode/form authority. This
// grouping adds no VEX, mask, broadcast, initialized input or memory backing.
func x86PackedBinaryByteEffect(op Op, ins Instr, arch string) (x86TypedByteEffect, bool, error) {
	name := strings.ToUpper(string(op))
	if !strings.HasPrefix(name, "P") && !strings.HasPrefix(name, "VP") {
		return x86TypedByteEffect{}, false, nil // Negative filter only; the specs below remain authoritative.
	}
	if effect, known, err := x86TypedLegacyIntegerArithmeticByteEffect(op, ins, arch); known {
		return effect, known, err
	}
	if effect, known, err := x86TypedPackedLowMultiplyByteEffect(op, ins, arch); known {
		return effect, known, err
	}
	if effect, known, err := x86TypedPackedWordMultiplyByteEffect(op, ins, arch); known {
		return effect, known, err
	}
	if effect, known, err := x86TypedDwordProductByteEffect(op, ins, arch); known {
		return effect, known, err
	}
	if effect, known, err := x86TypedPackedAverageByteEffect(op, ins, arch); known {
		return effect, known, err
	}
	return x86TypedPackedMultiplyAddByteEffect(op, ins, arch)
}
