package plan9asm

import "fmt"

func x86CallerVectorMoveKnown(ins Instr) bool {
	if amd64MaskMoveOperandsValid(ins, "amd64") {
		return true
	}
	form, recognized, err := amd64ScalarVectorMoveProperties(ins, "amd64")
	// The shared lowerer also models raw-only FP16 moves. Their caller-bank
	// permission needs its own complete source/runtime family, not inheritance
	// from the reviewed Go-named single/double-precision forms.
	return recognized && err == nil && !form.spec.rawOnly && !form.literal
}

func x86CloneCallerVectorState(vectors map[Reg]uint64, masks map[Reg]bool) (map[Reg]uint64, map[Reg]bool) {
	vectorCopy := make(map[Reg]uint64, len(vectors))
	for reg, known := range vectors {
		vectorCopy[reg] = known
	}
	maskCopy := make(map[Reg]bool, len(masks))
	for reg, known := range masks {
		maskCopy[reg] = known
	}
	return vectorCopy, maskCopy
}

func x86CallerVectorRoot(reg Reg) Reg {
	index, _ := amd64VectorRegisterIndex(reg, amd64VectorByteWidth(reg))
	return Reg(fmt.Sprintf("Z%d", index))
}

func x86CallerVectorByteMask(width int64) uint64 {
	if width <= 0 || width > 64 {
		return 0
	}
	if width == 64 {
		return ^uint64(0)
	}
	return uint64(1)<<uint(width) - 1
}

func x86CallerVectorMoveStep(ins *Instr, state *x86SourceStackState, proof *x86GoCallerRegionProof) bool {
	if amd64MaskMoveOperandsValid(*ins, "amd64") {
		return x86CallerMaskMoveStep(ins, state, proof)
	}
	form, recognized, err := amd64ScalarVectorMoveProperties(*ins, "amd64")
	if !recognized || err != nil || form.spec.rawOnly || form.literal || form.mask >= 0 && !state.masks[ins.Args[form.mask].Reg] {
		return false
	}
	width := int64(form.spec.bits / 8)
	lowMask := x86CallerVectorByteMask(width)
	if form.memory >= 0 {
		if !x86CallerOrdinaryMemoryStep(ins, form.memory, state, proof, width, form.memoryWrite) {
			return false
		}
		if form.memoryWrite {
			source := x86CallerVectorRoot(ins.Args[form.low].Reg)
			return state.vectors[source]&lowMask == lowMask
		}
	}
	destination := x86CallerVectorRoot(ins.Args[form.destination].Reg)
	old := state.vectors[destination]
	if form.spec.legacy {
		if form.memory >= 0 {
			// A legacy memory load zeroes the rest of XMM but preserves the
			// upper Y/Z aliases. Register copies preserve the old upper XMM.
			state.vectors[destination] = old | x86CallerVectorByteMask(16)
		} else {
			source := x86CallerVectorRoot(ins.Args[form.low].Reg)
			state.vectors[destination] = old&^lowMask | state.vectors[source]&lowMask
		}
		return true
	}
	known := ^uint64(0)
	if form.memory < 0 {
		low := state.vectors[x86CallerVectorRoot(ins.Args[form.low].Reg)] & lowMask
		upperMask := x86CallerVectorByteMask(16) &^ lowMask
		upper := state.vectors[x86CallerVectorRoot(ins.Args[form.upper].Reg)] & upperMask
		known = ^x86CallerVectorByteMask(16) | low | upper
	}
	if form.mask >= 0 && !form.zeroing {
		// The mask is initialized, not assumed all-one: either active data
		// or the old low element may survive. Higher lanes follow the move
		// form's upper merge/zero rules independently of that mask bit.
		known = known&^lowMask | known&old&lowMask
	}
	state.vectors[destination] = known
	return true
}

func x86CallerMaskMoveStep(ins *Instr, state *x86SourceStackState, proof *x86GoCallerRegionProof) bool {
	bits := amd64MaskMoveSpecs[ins.Op]
	width := int64(bits / 8)
	source, destination := ins.Args[0], ins.Args[1]
	if source.Kind == OpMem {
		if !x86CallerOrdinaryMemoryStep(ins, 0, state, proof, width, false) {
			return false
		}
	} else if amd64IsKOperand(source) {
		if !state.masks[source.Reg] {
			return false
		}
	} else if !x86CallerNumericRead(source, width, state) {
		return false // In particular, GP pointer bits cannot become mask data.
	}
	if destination.Kind == OpMem {
		return x86CallerOrdinaryMemoryStep(ins, 1, state, proof, width, true)
	}
	if amd64IsKOperand(destination) {
		state.masks[destination.Reg] = true
		return true
	}
	// All KMOV GP destinations zero-extend the transferred value, including
	// B/W/D, so the complete GP result is numeric; SP/BP/g stay forbidden.
	return x86CallerNumericWrite(destination.Reg, 8, state)
}

func x86CallerMoveMemoryShape(ins Instr, index int) (width int64, write, valid bool) {
	if index < 0 || index >= len(ins.Args) || !isAMD64MemoryOperand(ins.Args[index]) {
		return 0, false, false
	}
	if width = x86RawReturnScalarMoveBytes(Op(normalizeInstructionOpcode(ins.Op))); width != 0 && len(ins.Args) == 2 {
		return width, index == 1, true
	}
	if amd64MaskMoveOperandsValid(ins, "amd64") {
		return int64(amd64MaskMoveSpecs[ins.Op] / 8), index == 1, true
	}
	form, recognized, err := amd64ScalarVectorMoveProperties(ins, "amd64")
	if !recognized || err != nil || form.memory != index {
		return 0, false, false
	}
	return int64(form.spec.bits / 8), form.memoryWrite, true
}
