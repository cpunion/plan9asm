package plan9asm

// The existing complete CMOV grammar supplies widths and exact predicate
// dependencies. Its paired lowerer always emits the full volatile read,
// including false predicates and dead outputs. Continuing past that access
// produces NUM, not backing, bounds, pointer or native-entry authority.
func x86LoweredCMOVNativeRead(ins Instr, state *x86ResultState, owned []bool, resultOffset int64, arch string) (bool, bool) {
	form, known, err := x86ConditionalMoveProperties(ins.Op, ins, arch)
	if !known || err != nil {
		return false, known
	}
	if form.source.Kind != OpMem || form.source.Mem.Sym != "" ||
		!x86OrdinaryNativeMemoryOperand(form.source, arch) {
		return false, false // Logical fields and independently bound sources keep their own consumers.
	}
	width := int64(form.bits / 8)
	required := x86ByteConditionFlags(form.condition)
	if !x86ResultNamedPredicateInstruction(ins) || required == 0 || state.flags&required != required ||
		!x86ResultRead(form.destination, width, state, owned, arch) ||
		!x86ResultReadWithNativeContinuation(ins, form.source, width, state, owned, arch, true) {
		return false, true
	}
	return x86ResultStore(form.destination, width, state, owned, resultOffset, arch), true
}
