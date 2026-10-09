package plan9asm

// This private transfer extends the existing must-initialized result analysis
// for logical scalar views and existing exact same-Void-source whole128 caller
// reads. It does not expand the public asmdecl-filter query, caller admission,
// frame backing, or pointer permissions.
func checkX86ScalarFPResultInitialization(fn Func, sig FuncSig, arch string) error {
	reads := x86ScalarFPCallerReads(fn, sig, arch)
	return checkX86ABI0ResultInitializationAtSource(fn, sig, arch,
		func(ordinal int, ins Instr, state *x86ResultState, owned []bool, resultOffset int64, arch string) bool {
			if read, present := reads[ordinal]; present {
				return read.transfer(ins, state, owned, resultOffset, arch)
			}
			return x86ScalarFPResultTransfer(ins, state, owned, resultOffset, arch)
		}, true)
}

func x86ScalarFPResultTransfer(ins Instr, state *x86ResultState, owned []bool, resultOffset int64, arch string) bool {
	if form, known, err := x86PackedLogicalProperties(ins, arch); known {
		return err == nil && x86ScalarFPLogicalTransfer(ins, form, state, owned, arch)
	}
	if f, known, err := x86GPSIMDTransferProperties(ins, arch); known {
		return err == nil && x86ResultGPSIMDTransfer(ins, f, state, owned, resultOffset, arch)
	}
	if form, recognized, err := x86FloatAddSubProperties(ins, arch); recognized {
		return err == nil && x86ScalarFPFloatBinaryTransfer(ins, form.x86FloatBinaryForm, state, owned, arch)
	}
	if form, recognized, err := x86FloatMulDivProperties(ins, arch); recognized {
		return err == nil && x86ScalarFPFloatBinaryTransfer(ins, form.x86FloatBinaryForm, state, owned, arch)
	}
	if amd64MaskMoveOperandsValid(ins, arch) {
		width := int64(amd64MaskMoveSpecs[ins.Op] / 8)
		first, last := ins.Args[0], ins.Args[1]
		if amd64IsKOperand(first) {
			index, _ := amd64ParseKReg(first.Reg)
			if !state.masks[index] {
				return false
			}
		} else if !x86ResultRead(first, width, state, owned, arch) {
			return false
		}
		if amd64IsKOperand(last) {
			index, _ := amd64ParseKReg(last.Reg)
			state.masks[index] = true
			return true
		}
		return x86ResultStore(last, width, state, owned, resultOffset, arch)
	}
	form, recognized, err := amd64ScalarVectorMoveProperties(ins, arch)
	if !recognized || form.spec.legacy {
		return x86ResultTransfer(ins, state, owned, resultOffset, arch)
	}
	if err != nil || form.spec.rawOnly || form.spec.bits != 32 && form.spec.bits != 64 {
		return false
	}
	width := int64(form.spec.bits / 8)
	lowMask := uint16(1)<<uint(width) - 1
	xreg := func(operand Operand) (int, bool) {
		index, valid := amd64ParseXReg(operand.Reg)
		return index, operand.Kind == OpReg && valid && index < len(state.xmm) && (arch != "386" || index < 8)
	}
	if form.mask >= 0 {
		index, valid := amd64ParseKReg(ins.Args[form.mask].Reg)
		if !valid || !state.masks[index] {
			return false
		}
	}
	first, last := ins.Args[form.low], ins.Args[form.destination]
	if source, valid := xreg(first); valid {
		if state.xmm[source]&lowMask != lowMask {
			return false
		}
	} else if !form.literal && !x86ResultRead(first, width, state, owned, arch) {
		return false
	}
	if form.memoryWrite {
		if form.mask >= 0 {
			// A defined mask is not necessarily one. Only already initialized
			// bytes survive every possible inactive-store path.
			return x86ResultRead(last, width, state, owned, arch)
		}
		return x86ResultStore(last, width, state, owned, resultOffset, arch)
	}
	destination, valid := xreg(last)
	if !valid {
		return false
	}
	if form.mask >= 0 && !form.zeroing && state.xmm[destination]&lowMask != lowMask {
		return false // An inactive merge reads the old scalar, even for void.
	}
	known := uint16(0xffff)
	if form.upper >= 0 {
		upper, valid := xreg(ins.Args[form.upper])
		if !valid {
			return false
		}
		known = state.xmm[upper]&^lowMask | lowMask
	}
	if form.mask >= 0 && !form.zeroing {
		known = known&^lowMask | known&state.xmm[destination]&lowMask
	}
	state.xmm[destination] = known
	return true
}

// The private scalar observer tracks only low X bytes. Wider and physical MMX
// effects need their own state model; the ordinary lowerer/caller grammar still
// supports its independently proved widths. No native pointer becomes FP data.
func x86ScalarFPLogicalTransfer(ins Instr, form x86PackedLogicalForm, state *x86ResultState, owned []bool, arch string) bool {
	if form.mmx || form.byteWidth != 16 {
		return false
	}
	xreg := func(operand Operand) (int, bool) {
		index, valid := amd64ParseXReg(operand.Reg)
		return index, operand.Kind == OpReg && valid && index < len(state.xmm) && (arch != "386" || index < 8)
	}
	second, valid := xreg(ins.Args[form.second])
	if !valid {
		return false
	}
	destination, valid := xreg(ins.Args[form.destination])
	if !valid {
		return false
	}
	if form.mask >= 0 {
		index, valid := amd64ParseKReg(ins.Args[form.mask].Reg)
		if !valid || !state.masks[index] {
			return false
		}
	}
	first := uint64(0xffff)
	if form.memory >= 0 {
		if !x86ResultRead(form.source, int64(form.memoryBytes), state, owned, arch) {
			return false // Even a zero mask supplies no missing native envelope.
		}
	} else {
		index, valid := xreg(ins.Args[form.first])
		if !valid {
			return false
		}
		first = uint64(state.xmm[index])
	}
	state.xmm[destination] = uint16(x86PackedLogicalDefinedBytes(ins, form, first,
		uint64(state.xmm[second]), uint64(state.xmm[destination])))
	return true
}
