package plan9asm

// This domain observes only opaque incoming pointer identities. Copying or
// computing an address does not initialize any numerical bytes or its pointee:
// all native reads remain unproved without an independent caller-memory fact.
// It grants no bounds, noalias, escape, lifetime, frame or translator permission.

func x86ResultPointerWord(arch string) int64 {
	if arch == "386" {
		return 4
	}
	return 8
}

func x86ResultIncomingPointerAddressIdentity(arg Operand, state *x86ResultState, arch string) bool {
	if arg.Kind != OpMem {
		return false
	}
	memory := arg.Mem
	if memory.Sym != "" || memory.OffRaw != "" || memory.Segment != "" || memory.IndexExt != "" ||
		memory.Off < -1<<31 || memory.Off > 1<<31-1 {
		return false
	}
	word := x86ResultPointerWord(arch)
	base, _, valid := x86ResultGP(memory.Base, word, arch)
	if !valid || !state.shapeOnly && !state.pointerGP[base] {
		return false
	}
	if memory.Index == "" {
		return memory.Scale == 0
	}
	if memory.Scale != 1 && memory.Scale != 2 && memory.Scale != 4 && memory.Scale != 8 {
		return false
	}
	index, mask, valid := x86ResultGP(memory.Index, word, arch)
	return valid && state.gp[index]&mask == mask
}

func x86ResultIncomingPointerTransfer(ins Instr, state *x86ResultState, owned []bool, arch string) (bool, bool) {
	word := x86ResultPointerWord(arch)
	move, lea, add, sub := Op("MOVQ"), Op("LEAQ"), Op("ADDQ"), Op("SUBQ")
	if arch == "386" {
		move, lea, add, sub = "MOVL", "LEAL", "ADDL", "SUBL"
	}
	if len(ins.Args) != 2 {
		return false, false
	}
	first, last := ins.Args[0], ins.Args[1]
	writePointer := func() bool {
		if last.Kind != OpReg {
			return false
		}
		index, _, valid := x86ResultGP(last.Reg, word, arch)
		if valid {
			state.pointerGP[index], state.gp[index] = true, 0
		}
		return valid
	}
	if ins.Op == move {
		if first.Kind == OpFP && state.pointerFP[first.FPOffset] {
			return writePointer(), true
		}
		// The source-wide top state represents numerical and pointer inputs
		// together; a reaching state never carries that union as authority.
		if first.Kind == OpReg && !state.shapeOnly {
			index, _, valid := x86ResultGP(first.Reg, word, arch)
			if valid && state.pointerGP[index] {
				return writePointer(), true
			}
		}
	}
	if ins.Op == lea && first.Kind == OpMem {
		return x86ResultIncomingPointerAddressIdentity(first, state, arch) && writePointer(), true
	}
	if (ins.Op == add || ins.Op == sub) && last.Kind == OpReg && !state.shapeOnly {
		index, _, valid := x86ResultGP(last.Reg, word, arch)
		if !valid || !state.pointerGP[index] {
			return false, false
		}
		effect, recognized, err := x86TypedInstructionByteEffect(ins.Op, ins, arch)
		if !recognized || err != nil || effect.dynamicMemory ||
			!x86ResultRead(first, word, state, owned, arch) {
			return false, true
		}
		state.flags = state.flags&^effect.flagsKilled | effect.flagsDefined
		return true, true
	}
	return false, false
}
