package plan9asm

// These byte transfers reuse the same grammar as lowering. Native reads are
// conditional numerical availability only: they grant no source, pointer,
// alignment, initialized pointee, extent or later-read fact. A native=true
// caller must pair the instruction with an observable architectural read.
func x86ResultAlternatingFloatStep(ins Instr, state *x86ResultState, owned []bool, arch string, native bool) (bool, bool) {
	form, known, err := x86AlternatingFloatProperties(ins, arch)
	if !known || err != nil {
		return false, known
	}
	if !x86ResultNamedPredicateInstruction(ins) {
		return false, true
	}
	return x86ScalarFPFloatBinaryTransferWithRead(ins, form, state, owned, arch, native && !form.legacy), true
}

func x86ResultRoundStep(ins Instr, state *x86ResultState, owned []bool, arch string, native bool) (bool, bool) {
	form, known, err := x86RoundProperties(ins, arch)
	if !known || err != nil {
		return false, known
	}
	if !x86ResultNamedPredicateInstruction(ins) || !x86ResultRead(ins.Args[0], 1, state, owned, arch) {
		return false, true
	}
	low := x86CallerVectorByteMask(int64(form.memoryBytes))
	if form.memory {
		if !x86ResultReadWithNativeContinuation(ins, form.source, int64(form.memoryBytes), state, owned, arch, native && form.spec.vector) {
			return false, true
		}
	} else {
		source, valid := amd64VectorRegisterIndex(form.source.Reg, form.byteWidth)
		if !valid || arch == "386" && source >= 8 || state.vectorBytes(source)&low != low {
			return false, true
		}
	}
	destination, valid := amd64VectorRegisterIndex(form.destination.Reg, form.byteWidth)
	if !valid || arch == "386" && destination >= 8 {
		return false, true
	}
	defined := low
	if form.spec.scalar {
		upper, valid := amd64ParseXReg(form.upper.Reg)
		if !valid || arch == "386" && upper >= 8 {
			return false, true
		}
		defined |= state.vectorBytes(upper) & (0xffff &^ low)
	}
	state.writeVector(destination, form.byteWidth, defined, !form.spec.vector)
	return true, true // All unrelated vector/GP/FLAGS/mask/storage facts survive.
}
