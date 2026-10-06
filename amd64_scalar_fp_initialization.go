package plan9asm

// This private transfer extends the existing must-initialized result analysis
// only for logical scalar views. It does not expand the public asmdecl-filter
// query, caller admission, frame backing, or pointer permissions.
func checkX86ScalarFPResultInitialization(fn Func, sig FuncSig, arch string) error {
	return checkX86ABI0ResultInitialization(fn, sig, arch, x86ScalarFPResultTransfer, true)
}

func x86ScalarFPResultTransfer(ins Instr, state *x86ResultState, owned []bool, resultOffset int64, arch string) bool {
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
