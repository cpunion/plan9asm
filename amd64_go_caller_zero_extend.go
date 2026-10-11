package plan9asm

func x86CallerPackedZeroExtendStep(ins *Instr, form x86PackedZeroExtendForm, state *x86SourceStackState, proof *x86GoCallerRegionProof) bool {
	if form.mask >= 0 && !state.masks[ins.Args[form.mask].Reg] {
		return false
	}
	var input uint64
	if form.memory >= 0 {
		// Mask knowledge grants no address/alignment/envelope permission.
		// Keep the complete potential read bound even when lanes can fault-
		// suppress. Input regions never become writable scalar results.
		if !x86CallerOrdinaryMemoryStep(ins, form.memory, state, proof, int64(form.memoryBytes), false) {
			return false
		}
		input = x86CallerVectorByteMask(int64(form.memoryBytes))
	} else {
		input = state.vectors[x86CallerVectorRoot(ins.Args[0].Reg)]
	}
	destination := x86CallerVectorRoot(ins.Args[form.destination].Reg)
	old := state.vectors[destination]
	known := x86PackedZeroExtendDefinedBytes(form, input, old)
	width := x86CallerVectorByteMask(int64(form.byteWidth))
	if form.spec.vector {
		state.vectors[destination] = ^width | known
	} else {
		state.vectors[destination] = old&^width | known
	}
	return true
}

// This pure byte dependency grants no memory or caller permission. Upper
// zero bits are known independently of input bytes; mask/alias writes remain
// distinct from the register and memory observations that precede this step.
func x86PackedZeroExtendDefinedBytes(form x86PackedZeroExtendForm, input, old uint64) uint64 {
	var known uint64
	inputBytes, outputBytes := form.spec.inputBits/8, form.spec.outputBits/8
	for output := 0; output < form.byteWidth; output++ {
		part := output % outputBytes
		if part >= inputBytes || input>>uint(output/outputBytes*inputBytes+part)&1 != 0 {
			known |= uint64(1) << uint(output)
		}
	}
	if form.mask >= 0 && !form.zeroing {
		known &= old // An initialized mask is not an all-active mask promise.
	}
	return known
}
