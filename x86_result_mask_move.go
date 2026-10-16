package plan9asm

// Native continuation belongs only to lowering paired with the exact volatile
// access. Standalone declaration, caller and frame observers retain their own
// storage contracts; an incoming pointer never asserts backing or bounds.
func x86MaskMoveResultStep(ins Instr, state *x86ResultState, owned []bool,
	resultOffset int64, arch string, native bool) (bool, bool) {
	if !amd64MaskMoveOperandsValid(ins, arch) {
		return false, false
	}
	width := int64(amd64MaskMoveSpecs[ins.Op] / 8)
	first, last := ins.Args[0], ins.Args[1]
	if amd64IsKOperand(first) {
		index, _ := amd64ParseKReg(first.Reg)
		if !state.masks[index] {
			return false, true
		}
	} else if !x86ResultReadWithNativeContinuation(ins, first, width, state, owned, arch, native) {
		return false, true
	}
	if amd64IsKOperand(last) {
		index, _ := amd64ParseKReg(last.Reg)
		state.masks[index] = true
		return true, true
	}
	if native && x86ResultNamedPredicateInstruction(ins) && ins.x86NumericalRead == nil &&
		x86OrdinaryNativeMemoryOperand(last, arch) && x86ResultIncomingPointerAddressIdentity(last, state, arch) {
		return true, true // Native stores establish no reusable memory facts.
	}
	if last.Kind == OpReg {
		// KMOV's GP encoding zero-extends B/W/D, unlike partial MOVB/W.
		// Preserve the existing mode32 GP-Q boundary, not a new grant.
		word := x86ResultPointerWord(arch)
		if width < word {
			width = word
		}
	}
	return x86ResultStore(last, width, state, owned, resultOffset, arch), true
}
