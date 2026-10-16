package plan9asm

// This common dependency consumer supplies numerical values only. Operation
// specs independently identify whether old destination bytes are data inputs.
func x86ResultFixedVectorInputs(ins Instr, form x86PackedBinaryForm, oldInput bool, state *x86ResultState, owned []bool, arch string) bool {
	if !x86ResultNamedPredicateInstruction(ins) || x86Mode32VectorRegisterContext(ins, arch) != nil {
		return false
	}
	if form.mask >= 0 {
		mask, _ := amd64ParseKReg(ins.Args[form.mask].Reg)
		if !state.masks[mask] {
			return false
		}
	}
	width := int64(form.byteWidth)
	firstWidth := width
	if form.memory >= 0 {
		firstWidth = int64(form.memoryBytes)
	}
	if !x86ResultPackedVectorRead(ins.Args[form.first], firstWidth, state, owned, arch) ||
		form.second >= 0 && !x86ResultPackedVectorRead(ins.Args[form.second], width, state, owned, arch) {
		return false
	}
	// An operation may read its old destination as data, independently of
	// the optional masked merge. Neither dependency is supplied by .Z.
	if (oldInput || form.mask >= 0 && !form.zeroing) &&
		!x86ResultPackedVectorRead(ins.Args[form.destination], width, state, owned, arch) {
		return false
	}
	destination, _ := amd64VectorRegisterIndex(ins.Args[form.destination].Reg, form.byteWidth)
	state.writeVector(destination, form.byteWidth, x86CallerVectorByteMask(width), false)
	return true // No FLAGS/DF/mask, memory or pointer facts change.
}

// This is only a numerical vector view. Keep the older scalar/legacy-X read
// domain unchanged; neither native pointers nor wider logical storage follow.
func x86ResultPackedVectorRead(operand Operand, width int64, state *x86ResultState, owned []bool, arch string) bool {
	if operand.Kind != OpReg {
		return x86ResultRead(operand, width, state, owned, arch)
	}
	index, valid := amd64VectorRegisterIndex(operand.Reg, int(width))
	mask := x86CallerVectorByteMask(width)
	return valid && (arch != "386" || index < 8) && mask != 0 && state.vectorBytes(index)&mask == mask
}
