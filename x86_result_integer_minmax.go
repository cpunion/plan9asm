package plan9asm

func x86RawIntegerMinMaxSource(ins Instr, arch string) bool {
	_, known, err := x86IntegerMinMaxProperties(ins.Op, ins, arch)
	return known && err == nil && ins.x86Encoded && x86RawOrdinaryVectorSource(ins, arch)
}

func x86ResultIntegerMinMaxStep(ins Instr, state *x86ResultState, owned []bool, arch string, native bool) (bool, bool) {
	form, known, err := x86IntegerMinMaxProperties(ins.Op, ins, arch)
	if !known || err != nil {
		return false, known
	}
	if !x86ResultNamedPredicateInstruction(ins) {
		if !native || !x86RawIntegerMinMaxSource(ins, arch) {
			return false, true
		}
		// Preserve the already-validated physical form. This local numerical
		// projection does not reparse mode32 bytes as named source operands.
		ins.x86Encoded, ins.x86SourceRaw = false, nil
	}
	return x86ResultPackedLaneArithmeticStepWithRead(ins, form.x86PackedBinaryForm,
		form.spec.laneBits, state, owned, arch, native), true
}
