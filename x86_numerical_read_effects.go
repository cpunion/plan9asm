package plan9asm

// The actual SB source bytes are supplied only by a checked live certificate.
// All consumers use this same typed shape for the remaining numerical GP and
// flags effects; no observer turns the global into a pointer or numeric literal.
func x86NumericalReadEffect(ins Instr, arch string) (x86TypedByteEffect, bool) {
	_, form, valid := numericalCertifiedRead(ins, arch)
	if !valid {
		return x86TypedByteEffect{}, false
	}
	partner := x86TypedByteOperand{index: form.partner, operand: ins.Args[form.partner], width: form.spec.gpBytes,
		read: form.spec.compare, write: !form.spec.compare}
	effect := x86TypedByteEffect{operands: []x86TypedByteOperand{partner}}
	if form.spec.compare {
		effect.flagsDefined, effect.flagsKilled = x86ByteFlagsAll, x86ByteFlagsAll
	}
	return effect, true
}

func x86ResultNumericalReadStep(ins Instr, state *x86ResultState, owned []bool, offset int64, arch string) (bool, bool) {
	if ins.x86NumericalRead == nil {
		return false, false
	}
	effect, valid := x86NumericalReadEffect(ins, arch)
	if !valid {
		return false, true
	}
	for _, operand := range effect.operands {
		if operand.read && !x86ResultRead(operand.operand, operand.width, state, owned, arch) ||
			operand.write && !x86ResultStore(operand.operand, operand.width, state, owned, offset, arch) {
			return false, true
		}
	}
	state.flags = state.flags&^effect.flagsKilled | effect.flagsDefined
	return true, true
}

func x86CallerNumericalReadStep(ins Instr, state *x86SourceStackState) (bool, bool) {
	if ins.x86NumericalRead == nil {
		return false, false
	}
	effect, valid := x86NumericalReadEffect(ins, "amd64")
	if !valid {
		return false, true
	}
	for _, operand := range effect.operands {
		if operand.read && !x86CallerNumericRead(operand.operand, operand.width, state) ||
			operand.write && !x86CallerNumericWrite(operand.operand.Reg, operand.width, state) {
			return false, true
		}
	}
	if effect.flagsDefined != 0 {
		state.numericFlags = true
	}
	return true, true
}
