package plan9asm

func x86CallerPackedNarrowStep(ins *Instr, form x86PackedNarrowForm, state *x86SourceStackState, proof *x86GoCallerRegionProof) bool {
	if form.mmx || form.mask >= 0 && !state.masks[ins.Args[form.mask].Reg] {
		return false
	}
	width := x86CallerVectorByteMask(int64(form.byteWidth))
	first := state.vectors[x86CallerVectorRoot(ins.Args[form.first].Reg)] & width
	if form.memory >= 0 {
		// E4NF/.nb requires the complete tuple, even when every output mask
		// bit is zero. The form supplies a width, never a pointer permission.
		if !x86CallerOrdinaryMemoryStep(ins, form.memory, state, proof, int64(form.memoryBytes), false) {
			return false
		}
		first = width
	}
	second := state.vectors[x86CallerVectorRoot(ins.Args[form.second].Reg)] & width
	destination := x86CallerVectorRoot(ins.Args[form.destination].Reg)
	old := state.vectors[destination]
	inputBytes, outputBytes := form.spec.inputBits/8, form.spec.outputBits/8
	var known uint64
	for output := 0; output < form.byteWidth; output += outputBytes {
		// PACK concatenates separately within each 128-bit chunk. Go's
		// second operand supplies its low half; first supplies its high half.
		chunk, local := output/16*16, output%16
		input, source := chunk+(local%8)/outputBytes*inputBytes, second
		if local >= 8 {
			source = first
		}
		needed := x86CallerVectorByteMask(int64(inputBytes)) << uint(input)
		if source&needed == needed {
			known |= x86CallerVectorByteMask(int64(outputBytes)) << uint(output)
		}
	}
	if form.mask >= 0 && !form.zeroing {
		// An initialized mask is not a known mask value: a byte must be
		// initialized both when computed and when the old destination wins.
		known &= old
	}
	if form.legacy {
		state.vectors[destination] = old&^width | known
	} else {
		state.vectors[destination] = ^width | known
	}
	return true
}
