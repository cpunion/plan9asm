package plan9asm

import "fmt"

func x86CallerVectorMoveKnown(ins Instr) bool {
	if _, recognized, err := x86FloatCompareProperties(ins, "amd64"); recognized {
		return err == nil
	}
	if form, known, err := x86WordMultiplyAddProperties(ins, "amd64"); known {
		return err == nil && !form.mmx
	}
	if _, recognized, err := x86FloatMulDivProperties(ins, "amd64"); recognized {
		return err == nil
	}
	if _, recognized, err := x86PackedZeroExtendProperties(ins, "amd64"); recognized {
		return err == nil
	}
	if _, recognized, err := x86VectorZeroProperties(ins, "amd64"); recognized {
		return err == nil
	}
	if _, recognized, err := x86SignedIntegerFloatProperties(ins, "amd64"); recognized {
		return err == nil
	}
	if f, recognized, err := x86InsertExtractProperties(ins, "amd64"); recognized {
		return err == nil && !(f.mask >= 0 && f.memoryWrite)
	}
	if form, recognized, err := x86PackedShiftProperties(ins.Op, ins, "amd64"); recognized {
		return err == nil && !form.mmx
	}
	if _, recognized, err := x86HorizontalFloatProperties(ins, "amd64"); recognized {
		return err == nil
	}
	if form, recognized, err := x86PackedUnpackProperties(ins, "amd64"); recognized {
		return err == nil && !form.mmx
	}
	if _, recognized, err := x86FloatAddSubProperties(ins, "amd64"); recognized {
		return err == nil
	}
	if form, recognized, err := x86PackedLogicalProperties(ins, "amd64"); recognized {
		return err == nil && !form.mmx
	}
	if amd64MaskMoveOperandsValid(ins, "amd64") {
		return true
	}
	if form, recognized, err := amd64PackedVectorMoveProperties(ins, "amd64"); recognized {
		return err == nil && !form.undefinedOpcode()
	}
	if form, recognized, err := amd64FMA3Properties(ins, "amd64"); recognized {
		return err == nil && !form.spec.rawOnly
	}
	form, recognized, err := amd64ScalarVectorMoveProperties(ins, "amd64")
	// The shared lowerer also models raw-only FP16 moves. Their caller-bank
	// permission needs its own complete source/runtime family, not inheritance
	// from the reviewed Go-named single/double-precision forms.
	return recognized && err == nil && !form.spec.rawOnly
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
	if form, recognized, err := x86FloatCompareProperties(*ins, "amd64"); recognized {
		return err == nil && x86CallerFloatCompareStep(ins, form, state, proof)
	}
	if form, known, err := x86WordMultiplyAddProperties(*ins, "amd64"); known {
		return err == nil && x86CallerWordMultiplyAddStep(ins, form, state, proof)
	}
	if form, recognized, err := x86FloatMulDivProperties(*ins, "amd64"); recognized {
		return err == nil && x86CallerFloatBinaryStep(ins, form.x86FloatBinaryForm, state, proof)
	}
	if form, recognized, err := x86PackedZeroExtendProperties(*ins, "amd64"); recognized {
		return err == nil && x86CallerPackedZeroExtendStep(ins, form, state, proof)
	}
	if form, recognized, err := x86VectorZeroProperties(*ins, "amd64"); recognized {
		return err == nil && x86CallerVectorZeroStep(form, state)
	}
	if f, recognized, err := x86SignedIntegerFloatProperties(*ins, "amd64"); recognized {
		return err == nil && x86CallerSignedIntegerFloatStep(ins, f, state, proof)
	}
	if f, recognized, err := x86InsertExtractProperties(*ins, "amd64"); recognized {
		return err == nil && x86CallerInsertExtractStep(ins, f, state, proof)
	}
	if form, recognized, err := x86PackedShiftProperties(ins.Op, *ins, "amd64"); recognized {
		return err == nil && x86CallerPackedShiftStep(ins, form, state, proof)
	}
	if form, recognized, err := x86HorizontalFloatProperties(*ins, "amd64"); recognized {
		return err == nil && x86CallerHorizontalFloatStep(ins, form, state, proof)
	}
	if form, recognized, err := x86PackedUnpackProperties(*ins, "amd64"); recognized {
		return err == nil && x86CallerPackedUnpackStep(ins, form, state, proof)
	}
	if form, recognized, err := x86FloatAddSubProperties(*ins, "amd64"); recognized {
		return err == nil && x86CallerFloatAddSubStep(ins, form, state, proof)
	}
	if form, recognized, err := x86PackedLogicalProperties(*ins, "amd64"); recognized {
		return err == nil && x86CallerPackedLogicalStep(ins, form, state, proof)
	}
	if amd64MaskMoveOperandsValid(*ins, "amd64") {
		return x86CallerMaskMoveStep(ins, state, proof)
	}
	if form, recognized, err := amd64PackedVectorMoveProperties(*ins, "amd64"); recognized {
		return err == nil && x86CallerPackedVectorMoveStep(ins, form, state, proof)
	}
	if form, recognized, err := amd64FMA3Properties(*ins, "amd64"); recognized {
		return err == nil && x86CallerFMA3Step(ins, form, state, proof)
	}
	form, recognized, err := amd64ScalarVectorMoveProperties(*ins, "amd64")
	if !recognized || err != nil || form.spec.rawOnly || form.mask >= 0 && !state.masks[ins.Args[form.mask].Reg] {
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
	if form.spec.legacy || form.legacyZero {
		if form.memory >= 0 || form.literal {
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
	if form.memory < 0 && !form.literal {
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
	if form, known, err := x86WordMultiplyAddProperties(ins, "amd64"); known {
		return int64(form.memoryBytes), false, err == nil && !form.mmx && form.memory == index
	}
	if form, recognized, err := x86PackedZeroExtendProperties(ins, "amd64"); recognized {
		return int64(form.memoryBytes), false, err == nil && form.memory == index
	}
	if form, recognized, err := x86FloatCompareProperties(ins, "amd64"); recognized {
		return int64(form.memoryBytes), false, err == nil && form.memory == index
	}
	if f, recognized, err := x86SignedIntegerFloatProperties(ins, "amd64"); recognized {
		return int64(f.sourceBits / 8), false, err == nil && f.memory == index
	}
	if form, recognized, err := x86FloatMulDivProperties(ins, "amd64"); recognized {
		return int64(form.memoryBytes), false, err == nil && form.memory == index
	}
	if f, recognized, err := x86InsertExtractProperties(ins, "amd64"); recognized {
		return int64(f.spec.chunk), f.memoryWrite, err == nil && f.memory == index && !(f.mask >= 0 && f.memoryWrite)
	}
	if form, recognized, err := x86PackedShiftProperties(ins.Op, ins, "amd64"); recognized {
		return int64(form.memoryBytes), false, err == nil && !form.mmx && form.memory == index
	}
	if form, recognized, err := x86HorizontalFloatProperties(ins, "amd64"); recognized {
		return int64(form.byteWidth), false, err == nil && form.spec.vector && form.memory == index
	}
	if form, recognized, err := x86PackedUnpackProperties(ins, "amd64"); recognized {
		return int64(form.memoryBytes), false, err == nil && !form.mmx && !form.legacy && form.memory == index
	}
	if form, recognized, err := x86FloatAddSubProperties(ins, "amd64"); recognized {
		return int64(form.memoryBytes), false, err == nil && form.memory == index
	}
	if form, recognized, err := x86PackedLogicalProperties(ins, "amd64"); recognized {
		return int64(form.memoryBytes), false, err == nil && !form.mmx && form.memory == index
	}
	if width = x86RawReturnScalarMoveBytes(Op(normalizeInstructionOpcode(ins.Op))); width != 0 && len(ins.Args) == 2 {
		return width, index == 1, true
	}
	if form, recognized, err := x86ScalarExtensionProperties(ins.Op, ins, "amd64"); recognized {
		return int64(form.spec.sourceBits / 8), false, err == nil && index == 0
	}
	if amd64MaskMoveOperandsValid(ins, "amd64") {
		return int64(amd64MaskMoveSpecs[ins.Op] / 8), index == 1, true
	}
	if form, recognized, err := amd64PackedVectorMoveProperties(ins, "amd64"); recognized {
		return int64(form.byteWidth), form.memoryWrite, err == nil && !form.undefinedOpcode() && form.memory == index
	}
	if form, recognized, err := amd64FMA3Properties(ins, "amd64"); recognized {
		return int64(form.memoryBytes), false, err == nil && !form.spec.rawOnly && form.memory == index
	}
	form, recognized, err := amd64ScalarVectorMoveProperties(ins, "amd64")
	if !recognized || err != nil || form.memory != index {
		return 0, false, false
	}
	return int64(form.spec.bits / 8), form.memoryWrite, true
}

func x86CallerPackedVectorMoveStep(ins *Instr, form amd64PackedVectorMoveForm, state *x86SourceStackState, proof *x86GoCallerRegionProof) bool {
	if form.undefinedOpcode() || form.mask >= 0 && !state.masks[ins.Args[form.mask].Reg] {
		return false
	}
	widthMask := x86CallerVectorByteMask(int64(form.byteWidth))
	var known uint64
	if form.memory >= 0 {
		// A slice-data/nonalias witness is not a SIMD alignment promise.
		// Aligned register copies are valid, but aligned memory requires a
		// separate actual-address proof before borrowing this continuation.
		if form.spec.aligned || !x86CallerOrdinaryMemoryStep(ins, form.memory, state, proof, int64(form.byteWidth), form.memoryWrite) {
			return false
		}
		if form.memoryWrite {
			return state.vectors[x86CallerVectorRoot(ins.Args[0].Reg)]&widthMask == widthMask
		}
		known = widthMask
	} else {
		known = state.vectors[x86CallerVectorRoot(ins.Args[0].Reg)] & widthMask
	}
	destination := x86CallerVectorRoot(ins.Args[form.destination].Reg)
	if form.mask >= 0 && !form.zeroing {
		// No initialized mask is assumed all-one. Each byte must be numeric
		// in both the incoming and retained old lanes for an unknown mask.
		known &= state.vectors[destination]
	}
	state.vectors[destination] = ^widthMask | known
	return true
}
