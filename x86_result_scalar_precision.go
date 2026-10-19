package plan9asm

// Active floating input and possible merge bytes are mandatory, including
// Void: implementation zeros cannot stand in for their native exceptions.
// Independently copied upper bytes may stay unknown and supply no pointer,
// backing, frame, caller or raw-entry fact.
func x86ResultScalarPrecisionStep(ins Instr, state *x86ResultState, owned []bool, arch string, native bool) (bool, bool) {
	form, known, err := x86ScalarPrecisionProperties(ins.Op, ins, arch)
	if !known || err != nil {
		return false, known
	}
	if !x86ResultNamedPredicateInstruction(ins) {
		return false, true
	}
	if form.mask >= 0 {
		mask, _ := amd64ParseKReg(ins.Args[form.mask].Reg)
		if !state.masks[mask] {
			return false, true
		}
	}
	input := int64(form.spec.inputBits / 8)
	if form.memory >= 0 {
		if !x86ResultReadWithNativeContinuation(ins, ins.Args[0], input, state, owned, arch, native) {
			return false, true
		}
	} else {
		source, _ := amd64ParseXReg(ins.Args[0].Reg)
		required := x86CallerVectorByteMask(input)
		if state.vectorBytes(source)&required != required {
			return false, true
		}
	}
	destination, _ := amd64ParseXReg(ins.Args[form.destination].Reg)
	output := x86CallerVectorByteMask(int64(form.spec.outputBits / 8))
	if form.mask >= 0 && !form.zeroing && state.vectorBytes(destination)&output != output {
		return false, true
	}
	base, _ := amd64ParseXReg(ins.Args[form.second].Reg)
	defined := output | state.vectorBytes(base)&(0xffff&^output)
	state.writeVector(destination, 16, defined, form.legacy)
	return true, true
}
