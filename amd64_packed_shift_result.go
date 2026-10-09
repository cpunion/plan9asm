package plan9asm

// One output byte can span two input bytes for an immediate bit shift. With an
// initialized but unknown count, it instead requires the union of every input
// bit that could reach it. Oversized logical counts contribute only zeros.
// The masks describe numerical dependence, never the native memory envelope.
func x86PackedShiftByteInputs(form x86PackedShiftForm, output int) (data, count uint64) {
	laneBytes := form.laneBits / 8
	if form.countMode == x86PackedShiftBytes {
		laneBytes = 16
	}
	lane, at := output/laneBytes*laneBytes, output%laneBytes
	first, last := at*8, at*8+7
	if form.countOperand.Kind == OpImm {
		shift := int(uint8(form.countOperand.Imm))
		if form.countMode == x86PackedShiftBytes {
			shift *= 8
		}
		if form.left {
			first, last = first-shift, last-shift
		} else {
			first, last = first+shift, last+shift
		}
	} else {
		if form.left {
			first = 0
		} else {
			last = laneBytes*8 - 1
		}
		count = 255 // The upper qword of a uniform register count is ignored.
		if form.countMode == x86PackedShiftPerLane {
			count = x86CallerVectorByteMask(int64(laneBytes)) << uint(lane)
		}
	}
	if first < 0 {
		first = 0
	}
	if last >= laneBytes*8 {
		last = laneBytes*8 - 1
	}
	if first <= last {
		data = x86CallerVectorByteMask(int64(last/8-first/8+1)) << uint(lane+first/8)
	}
	return data, count
}

func x86PackedShiftDefinedBytes(form x86PackedShiftForm, data, counts, old uint64) uint64 {
	var known uint64
	for output := 0; output < form.byteWidth; output++ {
		dataMask, countMask := x86PackedShiftByteInputs(form, output)
		if data&dataMask == dataMask && counts&countMask == countMask {
			known |= uint64(1) << uint(output)
		}
	}
	// A defined mask is not a known active mask. Every inactive merging byte
	// still needs its old destination; inactive zeroing bytes are constants.
	if form.mask >= 0 && !form.zeroing {
		known &= old
	}
	return known
}

func x86ResultPackedShiftStep(ins Instr, state *x86ResultState, owned []bool, arch string) (bool, bool) {
	form, known, err := x86PackedShiftProperties(ins.Op, ins, arch)
	if !known {
		return false, false
	}
	if err != nil || form.mmx || form.byteWidth != 16 {
		return false, true // Low X state does not stand in for physical MMX/Y/Z.
	}
	if form.mask >= 0 {
		index, valid := amd64ParseKReg(ins.Args[form.mask].Reg)
		if !valid || !state.masks[index] {
			return false, true
		}
	}
	if form.memory >= 0 {
		// Even a saturating count or zero mask cannot create memory authority.
		// Legacy m128 counts additionally need original native backing, which
		// this numerical observer cannot provide for logical FP/AUTO storage.
		if form.legacy || !x86ResultRead(ins.Args[form.memory], int64(form.memoryBytes), state, owned, arch) {
			return false, true
		}
	}
	bytes := func(index int) uint64 {
		if index == form.memory {
			return 0xffff // Full checked extent; includes replicated broadcasts.
		}
		register, _ := amd64ParseXReg(ins.Args[index].Reg)
		return uint64(state.xmm[register])
	}
	data, counts := bytes(form.data), uint64(0xffff)
	if form.countOperand.Kind != OpImm {
		counts = bytes(form.count)
	}
	destination, _ := amd64ParseXReg(ins.Args[form.destination].Reg)
	state.xmm[destination] = uint16(x86PackedShiftDefinedBytes(form, data, counts, uint64(state.xmm[destination])))
	return true, true
}
