package plan9asm

// This common dependency consumer supplies numerical values only. Operation
// specs independently identify whether old destination bytes are data inputs.
func x86ResultFixedVectorInputs(ins Instr, form x86PackedBinaryForm, oldInput bool, state *x86ResultState, owned []bool, arch string) bool {
	return x86ResultFixedVectorInputsWithRead(ins, form, oldInput, state, owned, arch, false)
}

func x86ResultFixedVectorInputsWithRead(ins Instr, form x86PackedBinaryForm, oldInput bool,
	state *x86ResultState, owned []bool, arch string, native bool) bool {
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
	first := ins.Args[form.first]
	firstKnown := x86ResultPackedVectorRead(first, firstWidth, state, owned, arch)
	if !firstKnown && native && first.Kind != OpReg {
		firstKnown = x86ResultReadWithNativeContinuation(ins, first, firstWidth, state, owned, arch, true)
	}
	if !firstKnown ||
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

// Only the paired exact-width native emitter consumes this conditional read.
// Complete masks, second inputs and I2/T2/merge destination bytes stay required.
// No pointer backing, bounds, caller, frame or declaration authority follows.
func x86FixedVectorValueProperties(ins Instr, arch string) (x86PackedBinaryForm, bool, bool, error) {
	if parsed, spec, recognized, parseErr := x86IndexedPermuteProperties(ins.Op, ins, arch); recognized {
		return parsed, spec.mode != amd64IndexedPermuteSingle, recognized, parseErr
	}
	if parsed, recognized, parseErr := x86GFNIProperties(ins.Op, ins, arch); recognized {
		return parsed.x86PackedBinaryForm, false, recognized, parseErr
	}
	if parsed, recognized, parseErr := x86PackedBitCountProperties(ins.Op, ins, arch); recognized {
		return parsed.x86PackedBinaryForm, false, recognized, parseErr
	}
	return x86PackedBinaryForm{}, false, false, nil
}

func x86LoweredFixedVectorNativeRead(ins Instr, state *x86ResultState, owned []bool, arch string) (bool, bool) {
	form, oldInput, known, err := x86FixedVectorValueProperties(ins, arch)
	if !known || err != nil {
		return false, known
	}
	if form.memory < 0 || ins.Args[form.memory].Kind != OpMem {
		return false, false // Logical fields and bound DATA retain their existing consumers.
	}
	return x86ResultFixedVectorInputsWithRead(ins, form, oldInput, state, owned, arch, true), true
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
