package plan9asm

// Comparison writes either complete predicate lanes or a fully defined K
// result. A K destination is never an old merge input; its optional source
// mask must still be initialized, even when source and destination alias.
func x86CallerFloatCompareStep(ins *Instr, form x86FloatCompareForm, state *x86SourceStackState, proof *x86GoCallerRegionProof) bool {
	if form.mask >= 0 && !state.masks[ins.Args[form.mask].Reg] {
		return false
	}
	width := form.byteWidth
	if form.scalar {
		width = form.bits / 8
	}
	required := x86CallerVectorByteMask(int64(width))
	if form.memory >= 0 && !x86CallerOrdinaryMemoryStep(ins, form.memory, state, proof, int64(form.memoryBytes), false) {
		return false
	}
	for _, index := range []int{form.first, form.second} {
		if index != form.memory && state.vectors[x86CallerVectorRoot(ins.Args[index].Reg)]&required != required {
			return false
		}
	}
	if form.maskResult {
		state.masks[ins.Args[form.destination].Reg] = true
		return true
	}
	destination := x86CallerVectorRoot(ins.Args[form.destination].Reg)
	known := ^uint64(0)
	if form.legacy {
		known = state.vectors[destination] | required
	} else if form.scalar {
		second := state.vectors[x86CallerVectorRoot(ins.Args[form.second].Reg)]
		upper := x86CallerVectorByteMask(16) &^ required
		known = ^x86CallerVectorByteMask(16) | required | second&upper
	}
	state.vectors[destination] = known
	return true
}
