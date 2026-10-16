package plan9asm

// Conditional continuation of the complete native BIT emitter. The exact
// volatile word read/write and signed dynamic bit-string addressing remain
// observable. This supplies CF, not reusable memory or address authority.
func x86LoweredNativeBitTest(ins Instr, state *x86ResultState, owned []bool, arch string) (bool, bool) {
	form, known, err := x86BitTestProperties(ins.Op, ins, arch)
	if !known || err != nil || form.destination.Kind != OpMem {
		return false, false
	}
	if !x86ResultNamedPredicateInstruction(ins) ||
		!x86ResultIncomingPointerAddressIdentity(form.destination, state, arch) ||
		!x86ResultRead(form.index, int64(form.bits/8), state, owned, arch) {
		return false, true
	}
	effect, _, _ := x86TypedBitTestByteEffect(ins.Op, ins, arch)
	state.flags = state.flags&^effect.flagsKilled | effect.flagsDefined
	return true, true // No GP, vector, DF or logical storage changes.
}
