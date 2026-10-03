package plan9asm

// Availability is separate from encoding. No count, mask, or zero-valued
// shadow register is evidence that an incoming vector contains numeric data.
func x86CallerPackedShiftStep(ins *Instr, form x86PackedShiftForm, state *x86SourceStackState, proof *x86GoCallerRegionProof) bool {
	if form.mmx || form.mask >= 0 && !state.masks[ins.Args[form.mask].Reg] {
		return false
	}
	if form.memory >= 0 {
		// Legacy XMM uniform count loads require native alignment, absent
		// from the slice-region contract. VEX/EVEX are unaligned accesses.
		if form.legacy || !x86CallerOrdinaryMemoryStep(ins, form.memory, state, proof, int64(form.memoryBytes), false) {
			return false
		}
	}
	width := x86CallerVectorByteMask(int64(form.byteWidth))
	data := width
	if form.data != form.memory {
		data = state.vectors[x86CallerVectorRoot(ins.Args[form.data].Reg)] & width
	}
	if data != width {
		return false
	}
	if ins.Args[form.count].Kind != OpImm && form.count != form.memory {
		countWidth := int64(form.byteWidth)
		if form.countMode == x86PackedShiftUniform {
			countWidth = 8 // The native register's upper count qword is ignored.
		}
		needed := x86CallerVectorByteMask(countWidth)
		if state.vectors[x86CallerVectorRoot(ins.Args[form.count].Reg)]&needed != needed {
			return false
		}
	}
	// Counts can move any source bit across bytes. An independently bound
	// count value would be needed to accept only partially defined data.
	known := width
	destination := x86CallerVectorRoot(ins.Args[form.destination].Reg)
	old := state.vectors[destination]
	if form.mask >= 0 && !form.zeroing {
		known &= old
	}
	if form.legacy {
		state.vectors[destination] = old&^width | known
	} else {
		state.vectors[destination] = ^width | known
	}
	return true
}
