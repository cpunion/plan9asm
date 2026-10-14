package plan9asm

import "reflect"

// These are source ordinals of complete, same-block prefix/body pairs. This
// is not an inert LOCK effect or a memory capability. The following opcode
// still needs its own independent reaching input/result transfer.
func x86ResultSourceLockPrefixes(fn Func, arch string) map[int]Instr {
	bound := make(map[int]Instr)
	if fn.x86DirectTailRoot != nil {
		return bound // A one-node continuation block cannot borrow source adjacency.
	}
	var ordinals []int
	for ordinal, ins := range fn.Instrs {
		if ins.Op != OpLABEL {
			ordinals = append(ordinals, ordinal)
		}
	}
	context := amd64Ctx{goarch: arch, blocks: amd64SplitBlocks(fn)}
	at := 0
	for bi, block := range context.blocks {
		for ii, ins := range block.instrs {
			if at >= len(ordinals) || !reflect.DeepEqual(ins, fn.Instrs[ordinals[at]]) {
				return nil
			}
			if ins.Op == "LOCK" && x86ResultNamedPredicateInstruction(ins) && ii+1 < len(block.instrs) &&
				x86ResultNamedPredicateInstruction(block.instrs[ii+1]) && context.namedLockPrefixBinding(bi, ii, ins) == nil {
				bound[ordinals[at]] = ins
			}
			at++
		}
	}
	if at != len(ordinals) {
		return nil
	}
	return bound
}

func x86CompareExchangeResultOutputs(form x86CompareExchangeForm) (int64, []Reg) {
	width, outputs := int64(form.spec.bits/8), []Reg{AX}
	if form.spec.pair {
		return width / 2, append(outputs, DX)
	}
	if form.destination.Kind == OpReg {
		outputs = append(outputs, form.destination.Reg)
	}
	return width, outputs
}

// This is only the numerical continuation of the already paired native CAS
// emitter. It supplies neither a readable pointee nor reusable write/extent,
// native-entry, alignment, frame or caller permission. The original complete
// sideeffect opcode still performs its read/write cycle, faults and ordering.
func x86LoweredCompareExchangeResultStep(ins Instr, state *x86ResultState, owned []bool, arch string) (bool, bool) {
	form, known, err := x86CompareExchangeProperties(ins.Op, ins, arch)
	if !known {
		return false, false
	}
	if err != nil || !x86ResultNamedPredicateInstruction(ins) {
		return false, true
	}
	width, outputs := x86CompareExchangeResultOutputs(form)
	inputs := []Operand{form.source, {Kind: OpReg, Reg: AX}}
	if form.spec.pair {
		inputs = []Operand{{Kind: OpReg, Reg: AX}, {Kind: OpReg, Reg: DX}, {Kind: OpReg, Reg: BX}, {Kind: OpReg, Reg: CX}}
	}
	for _, operand := range inputs {
		if !x86ResultRead(operand, width, state, owned, arch) {
			return false, true
		}
	}
	if form.destination.Kind == OpReg {
		if !x86ResultRead(form.destination, width, state, owned, arch) {
			return false, true
		}
	} else if !x86ResultIncomingPointerAddressIdentity(form.destination, state, arch) {
		return false, true // No SB, FP/AUTO, segment or numerical base is inferred.
	}
	for _, reg := range outputs {
		root, mask, valid := x86ResultGP(reg, width, arch)
		if !valid {
			return false, true
		}
		// The accumulator write is conditional. In particular, a successful
		// dword CAS preserves its old high half; no unconditional zero extension.
		state.gp[root] |= mask
		state.pointerGP[root] = false
	}
	if form.spec.pair {
		state.flags |= x86ByteFlagZF // 8B/16B leave the other status flags unchanged.
	} else {
		state.flags = x86ByteFlagsAll
	}
	return true, true
}
