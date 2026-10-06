package plan9asm

// This transfer consumes existing typed arithmetic syntax solely for numeric
// byte definedness. Native reads, caller certificates, ISA, exceptions and
// MXCSR remain the lowerer's independent contracts. It deliberately represents
// only X bytes; a complete Y/Z effect cannot be inferred from a low-X mask.
func x86ScalarFPFloatBinaryTransfer(ins Instr, form x86FloatBinaryForm, state *x86ResultState, owned []bool, arch string) bool {
	if form.byteWidth != 16 || form.bits != 32 && form.bits != 64 {
		return false
	}
	width := int64(form.byteWidth)
	if form.scalar {
		width = int64(form.bits / 8)
	}
	required := uint16(1)<<uint(width) - 1
	xreg := func(operand Operand) (int, bool) {
		index, valid := amd64ParseXReg(operand.Reg)
		return index, operand.Kind == OpReg && valid && index < len(state.xmm) &&
			(arch != "386" || index < 8) && (!form.legacy || index < 16)
	}
	if form.mask >= 0 {
		index, valid := amd64ParseKReg(ins.Args[form.mask].Reg)
		if !valid || !state.masks[index] {
			return false
		}
	}
	for _, input := range []int{form.first, form.second} {
		if form.literal && input == form.first {
			continue
		}
		if input == form.memory {
			// FP/AUTO have independently initialized logical byte envelopes.
			// A GP/SB native address cannot obtain authority from this query.
			if !x86ResultRead(ins.Args[input], int64(form.memoryBytes), state, owned, arch) {
				return false
			}
			continue
		}
		index, valid := xreg(ins.Args[input])
		if !valid || state.xmm[index]&required != required {
			return false
		}
	}
	destination, valid := xreg(ins.Args[form.destination])
	if !valid {
		return false
	}
	old := state.xmm[destination]
	if form.mask >= 0 && !form.zeroing && old&required != required {
		return false
	}
	known := uint16(0xffff)
	if form.legacy {
		known = old | required
	} else if form.scalar {
		second, valid := xreg(ins.Args[form.second])
		if !valid {
			return false
		}
		known = required | state.xmm[second]&^required
	}
	state.xmm[destination] = known
	return true
}
