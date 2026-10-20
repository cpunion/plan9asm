package plan9asm

// Ordinary named FMA3 consumes three numerical sources in every order. Its
// destructive destination remains an arithmetic input even with zero masking.
// The existing raw-only FP16 consumer owns a separate exact-source contract.
func x86ResultFMA3Step(ins Instr, state *x86ResultState, owned []bool, arch string, nativeRead bool) (bool, bool) {
	form, known, err := amd64FMA3Properties(ins, arch)
	if !known || form.spec.rawOnly {
		return false, false
	}
	if err != nil || !x86ResultNamedPredicateInstruction(ins) {
		return false, true
	}
	if form.mask >= 0 {
		mask, _ := amd64ParseKReg(ins.Args[form.mask].Reg)
		if !state.masks[mask] {
			return false, true
		}
	}
	active := form.byteWidth
	if form.spec.scalar {
		active = form.spec.laneBits / 8
	}
	required := x86CallerVectorByteMask(int64(active))
	for _, index := range []int{0, 1, form.destination} {
		if index == form.memory {
			if !x86ResultReadWithNativeContinuation(ins, ins.Args[index], int64(form.memoryBytes), state, owned, arch, nativeRead) {
				return false, true
			}
			continue
		}
		register, _ := amd64VectorRegisterIndex(ins.Args[index].Reg, form.byteWidth)
		if state.vectorBytes(register)&required != required {
			return false, true
		}
	}
	destination, _ := amd64VectorRegisterIndex(ins.Args[form.destination].Reg, form.byteWidth)
	knownBytes := ^uint64(0)
	if form.spec.scalar {
		// Unlike scalar binary arithmetic, FMA's upper-X passthrough comes
		// from the old destructive destination, not either explicit source.
		knownBytes = required | state.vectorBytes(destination)&(0xffff&^required)
	}
	state.writeVector(destination, form.byteWidth, knownBytes, false)
	return true, true
}
