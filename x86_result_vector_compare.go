package plan9asm

// Legacy/VEX comparisons write vector lanes; EVEX comparisons write K bits.
// Their common operand grammar must not erase these different destination
// roles. Complete numerical inputs confer no native/caller memory permission.
func x86ResultVectorCompareStep(ins Instr, state *x86ResultState, owned []bool, arch string) (bool, bool) {
	form, known, err := x86PackedIntegerCompareProperties(ins.Op, ins, arch)
	if !known {
		immediate, recognized, immediateErr := x86ImmediatePackedCompareProperties(ins.Op, ins, arch)
		form, known, err = immediate.x86PackedIntegerCompareForm, recognized, immediateErr
	}
	if !known {
		return false, false
	}
	if err != nil || form.mmx {
		return false, true
	}
	if form.mask >= 0 {
		index, valid := amd64ParseKReg(ins.Args[form.mask].Reg)
		if !valid || !state.masks[index] {
			return false, true
		}
	}
	required := x86CallerVectorByteMask(int64(form.byteWidth))
	for _, input := range []int{form.first, form.second} {
		operand := ins.Args[input]
		if input == form.first {
			operand = form.source
		}
		if input == form.memory {
			if !x86ResultRead(operand, int64(form.memoryBytes), state, owned, arch) {
				return false, true
			}
			continue
		}
		index, valid := amd64VectorRegisterIndex(operand.Reg, form.byteWidth)
		if !valid || state.vectorBytes(index)&required != required {
			return false, true
		}
	}
	destination := ins.Args[form.destination].Reg
	if form.maskResult {
		index, _ := amd64ParseKReg(destination)
		state.masks[index] = true
	} else {
		index, _ := amd64VectorRegisterIndex(destination, form.byteWidth)
		state.writeVector(index, form.byteWidth, required, form.legacy)
	}
	return true, true
}
