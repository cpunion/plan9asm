package plan9asm

func x86ResultTernaryLogicStep(ins Instr, state *x86ResultState, owned []bool, arch string) (bool, bool) {
	form, recognized, err := x86TernaryLogicProperties(ins.Op, ins, arch)
	if !recognized {
		return false, false
	}
	if err != nil {
		return false, true
	}
	if form.mask >= 0 {
		index, _ := amd64ParseKReg(ins.Args[form.mask].Reg)
		if !state.masks[index] {
			return false, true
		}
	}
	// Fault suppression and truth-table independence never grant authority to
	// read an unknown address or a partially described memory tuple.
	if form.memory >= 0 && !x86ResultRead(ins.Args[form.memory], int64(form.memoryBytes), state, owned, arch) {
		return false, true
	}
	known := x86CallerVectorByteMask(int64(form.byteWidth))
	dependencies := form.dependencies()
	if form.mask >= 0 && !form.zeroing {
		dependencies |= 1 // Inactive merging lanes preserve A.
	}
	for role, argument := range form.inputs {
		if dependencies>>uint(role)&1 == 0 || argument == form.memory {
			continue
		}
		index, _ := amd64VectorRegisterIndex(ins.Args[argument].Reg, form.byteWidth)
		known &= state.vectorBytes(index)
	}
	destination, _ := amd64VectorRegisterIndex(ins.Args[form.destination].Reg, form.byteWidth)
	state.writeVector(destination, form.byteWidth, known, false)
	return true, true
}
