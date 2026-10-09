package plan9asm

// A block transfer selects a 128/256-bit chunk. Unknown bytes elsewhere in
// the source are not dependencies; masked merge bytes retain both possibilities.
// This numerical effect never creates native memory or initialized pointers.
func x86ResultBlockTransfer(ins Instr, form x86InsertExtractForm, state *x86ResultState,
	owned []bool, resultOffset int64, arch string) bool {
	if form.spec.scalar || form.memory >= 0 && (ins.x86AddressBits != 0 || ins.x86UnboundAddressOverride) {
		return false
	}
	if form.mask >= 0 {
		index, valid := amd64ParseKReg(ins.Args[form.mask].Reg)
		if !valid || !state.masks[index] {
			return false
		}
	}
	chunk := x86CallerVectorByteMask(int64(form.spec.chunk))
	source := chunk
	if form.spec.insert && form.memory >= 0 {
		// E6NF reads its entire native tuple even for a zero mask, an
		// unobserved chunk or an independently initialized return value.
		if !x86ResultRead(ins.Args[form.memory], int64(form.spec.chunk), state, owned, arch) {
			return false
		}
	} else {
		operand := ins.Args[form.source]
		index, valid := amd64VectorRegisterIndex(operand.Reg, amd64VectorByteWidth(operand.Reg))
		if !valid {
			return false
		}
		source = state.vectorBytes(index)
	}
	known := source >> uint(form.offset) & chunk
	if form.spec.insert {
		index, _ := amd64VectorRegisterIndex(ins.Args[form.base].Reg, form.width)
		mask := chunk << uint(form.offset)
		known = state.vectorBytes(index)&^mask | (source&chunk)<<uint(form.offset)
	}
	destination := ins.Args[form.destination]
	if form.memoryWrite {
		if known&chunk != chunk {
			return false
		}
		if form.mask >= 0 {
			// A defined mask need not write any byte. The destination must
			// already be numerical on every inactive-store path. Its read
			// authority still cannot expand the separate write envelope.
			if !x86ResultRead(destination, int64(form.output), state, owned, arch) {
				return false
			}
		}
		return x86ResultStore(destination, int64(form.output), state, owned, resultOffset, arch)
	}
	index, valid := amd64VectorRegisterIndex(destination.Reg, form.output)
	if !valid {
		return false
	}
	if form.mask >= 0 && !form.zeroing {
		known &= state.vectorBytes(index)
	}
	state.writeVector(index, form.output, known, false)
	return true
}
