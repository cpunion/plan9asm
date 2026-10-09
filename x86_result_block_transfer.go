package plan9asm

// A block transfer selects a 128/256-bit chunk. Unknown bytes elsewhere in
// the source are not dependencies; masked merge bytes retain both possibilities.
// This numerical effect never creates native memory or initialized pointers.
func x86ResultBlockTransfer(ins Instr, form x86InsertExtractForm, state *x86ResultState,
	owned []bool, resultOffset int64, arch string) bool {
	return x86ResultBlockTransferWithRead(ins, form, state, owned, resultOffset, arch, false)
}

// This private counterpart is paired with loadInsertExtractSource's full-tuple
// volatile E6NF read. Continuing past that read provides numerical bytes, not
// initialized Go pointee, bounds, alignment, write or raw/caller permission.
func x86LoweredBlockInsertRead(ins Instr, state *x86ResultState, owned []bool, resultOffset int64, arch string) (bool, bool) {
	form, recognized, err := x86InsertExtractProperties(ins, arch)
	if !recognized || err != nil || form.spec.scalar || !form.spec.insert || form.memory < 0 {
		return false, false
	}
	source := ins.Args[form.memory]
	if source.Kind != OpMem || source.Mem.Sym != "" || !x86OrdinaryNativeMemoryOperand(source, arch) {
		return false, false
	}
	if !x86ResultIncomingPointerAddressIdentity(source, state, arch) {
		return false, true
	}
	return x86ResultBlockTransferWithRead(ins, form, state, owned, resultOffset, arch, true), true
}

func x86ResultBlockTransferWithRead(ins Instr, form x86InsertExtractForm, state *x86ResultState,
	owned []bool, resultOffset int64, arch string, conditionalRead bool) bool {
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
		if !conditionalRead && !x86ResultRead(ins.Args[form.memory], int64(form.spec.chunk), state, owned, arch) {
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
