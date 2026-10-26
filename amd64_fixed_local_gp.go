package plan9asm

// Reuse the ordinary typed GP grammar. The descriptor only names byte and
// FLAGS dependencies; this consumer must still establish every private read,
// address origin and write. CMP/TEST retain their opposite immediate roles.
func fixedLocalGPByteEffect(ins Instr) (x86TypedByteEffect, bool, error) {
	if effect, known, err := x86TypedGPByteEffect(ins.Op, ins, "amd64"); known {
		return effect, true, err
	}
	form, known, err := x86ScalarCompareProperties(ins.Op, ins, "amd64")
	if !known || err != nil {
		return x86TypedByteEffect{}, known, err
	}
	effect := x86TypedByteEffect{flagsDefined: x86ByteFlagsAll, flagsKilled: x86ByteFlagsAll}
	if form.test {
		effect.flagsDefined &^= x86ByteFlagAF
	}
	for index, operand := range form.args {
		effect.operands = append(effect.operands, x86TypedByteOperand{
			index: index, operand: operand, width: int64(form.bits / 8), read: true,
		})
	}
	return effect, true, nil
}

// A parsed address is not necessarily a memory access. Numerical LEA and
// sized NOP use the same syntax without loading/storing it. Query the validated
// effect rather than granting them a fabricated private-storage certificate.
func fixedLocalGPAddressOnly(ins Instr, index int) bool {
	if index < 0 || index >= len(ins.Args) || ins.Args[index].Kind != OpMem {
		return false
	}
	effect, known, err := fixedLocalGPByteEffect(ins)
	if !known || err != nil || effect.dynamicMemory {
		return false
	}
	for _, operand := range effect.operands {
		if operand.index == index && (operand.operand.Kind == OpMem || operand.operand.Kind == OpSym) &&
			(operand.read || operand.write) {
			return false
		}
	}
	return true
}

func fixedLocalGPByteStep(ins *Instr, state *x86FixedLocalState, proof *x86FixedLocalProof, residue int,
	effect x86TypedByteEffect) error {
	fail := func() error {
		return goAMD64FrameContext("fixed local typed GP effect lacks numerical bytes, FLAGS or private storage: %s", ins.Op)
	}
	if effect.dynamicMemory || state.flags&effect.flagsRead != effect.flagsRead {
		return fail()
	}
	// Read all inputs before any write, including aliased outputs. RMW receives
	// one write-capable access certificate only after its old bytes are proved.
	for _, operand := range effect.operands {
		if !operand.read {
			continue
		}
		if operand.operand.Kind != OpMem && operand.operand.Kind != OpSym {
			if !fixedLocalNumericRead(operand.operand, operand.width, state) {
				return fail()
			}
			continue
		}
		if !operand.write {
			if _, _, err := fixedLocalReadMemory(ins, operand.index, state, proof, residue, operand.width, 1); err != nil {
				return err
			}
			continue
		}
		offset, err := fixedLocalMemory(ins, operand.index, state, proof, residue, operand.width, 1, true, false)
		if err != nil {
			return err
		}
		for index := offset; index < offset+operand.width; index++ {
			if !state.bytes[index] {
				return fail()
			}
		}
	}
	for _, operand := range effect.operands {
		if !operand.write {
			continue
		}
		switch operand.operand.Kind {
		case OpMem:
			offset, err := fixedLocalMemory(ins, operand.index, state, proof, residue, operand.width, 1, true, false)
			if err != nil {
				return err
			}
			fixedLocalNumericalStore(state, offset, operand.width, true)
		case OpReg:
			root, valid := x86CallerGP(operand.operand.Reg)
			mask := x86CallerByteMask(operand.operand.Reg, operand.width)
			_, private := state.pointers[root]
			_, opaque := state.opaque[root]
			if !valid || root == SP || root == BP || x86SourceGRegister(root) || mask == 0 ||
				(private || opaque) && operand.width != 8 {
				return fail()
			}
			if operand.width == 4 {
				mask = 255 // Architectural dword writes define the upper half.
			}
			delete(state.pointers, root)
			delete(state.opaque, root)
			delete(state.constants, root)
			fixedLocalWriteBooleans(state, operand.operand.Reg, operand.width, 0)
			state.known[root] |= mask
		default:
			return fail() // FP transport and external stores have separate owners.
		}
	}
	state.flags = state.flags&^effect.flagsKilled | effect.flagsDefined
	return nil
}
