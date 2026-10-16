package plan9asm

// Unknown mask values can select any source lane and any merge byte. Require
// the complete source and, for merging, the complete old destination. A mask's
// definedness is not a constant-value or memory-access certificate.
func x86ResultCompactExpandStep(ins Instr, state *x86ResultState, arch string) (bool, bool) {
	form, known, err := x86PackedCompactProperties(ins.Op, ins, arch)
	if !known {
		return false, false
	}
	if err != nil || form.memory >= 0 || !x86ResultNamedPredicateInstruction(ins) {
		return false, true
	}
	source, sourceOK := amd64VectorRegisterIndex(ins.Args[0].Reg, form.byteWidth)
	destination, destinationOK := amd64VectorRegisterIndex(ins.Args[form.destination].Reg, form.byteWidth)
	if !sourceOK || !destinationOK || arch == "386" && (source >= 8 || destination >= 8) {
		return false, true // Go's accepted mode32 high-bank spellings are not physical value inputs.
	}
	required := x86CallerVectorByteMask(int64(form.byteWidth))
	if state.vectorBytes(source)&required != required {
		return false, true
	}
	if form.mask >= 0 {
		mask, _ := amd64ParseKReg(ins.Args[form.mask].Reg)
		if !state.masks[mask] || !form.zeroing && state.vectorBytes(destination)&required != required {
			return false, true
		}
	}
	state.writeVector(destination, form.byteWidth, required, false)
	return true, true // EVEX clears wider aliases; FLAGS, DF and K are unchanged.
}
