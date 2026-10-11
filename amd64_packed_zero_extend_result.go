package plan9asm

func x86ResultPackedZeroExtendStep(ins Instr, state *x86ResultState, owned []bool, arch string, conditionalNativeRead bool) (bool, bool) {
	form, known, err := x86PackedZeroExtendProperties(ins, arch)
	if !known || err != nil {
		return false, known
	}
	if form.mask >= 0 {
		index, valid := amd64ParseKReg(ins.Args[form.mask].Reg)
		if !valid || !state.masks[index] {
			return false, true
		}
	}
	var input uint64
	if form.memory >= 0 {
		// Only the lowering observer may use the paired native E5 read.
		// The full potential tuple remains required for FP/AUTO/caller proofs;
		// a numerical mask cannot manufacture a smaller owned envelope.
		if !x86ResultReadWithNativeContinuation(ins, ins.Args[0], int64(form.memoryBytes), state, owned, arch, conditionalNativeRead) {
			return false, true
		}
		input = x86CallerVectorByteMask(int64(form.memoryBytes))
	} else {
		index, _ := amd64VectorRegisterIndex(ins.Args[0].Reg, form.sourceRegisterBytes)
		input = state.vectorBytes(index)
	}
	destination, _ := amd64VectorRegisterIndex(ins.Args[form.destination].Reg, form.byteWidth)
	initialized := x86PackedZeroExtendDefinedBytes(form, input, state.vectorBytes(destination))
	state.writeVector(destination, form.byteWidth, initialized, !form.spec.vector)
	return true, true
}
