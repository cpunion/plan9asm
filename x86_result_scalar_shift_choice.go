package plan9asm

import "reflect"

func x86RawLegacyScalarShiftSource(ins Instr, arch string) bool {
	form, known, err := x86ScalarShiftProperties(ins.Op, ins, arch)
	origin := ins.x86SourceRaw
	return known && err == nil && !form.bmi2 && origin != nil && origin.op == ins.Op &&
		reflect.DeepEqual(origin.args, ins.Args) && origin.vectorBytes == ins.x86VectorBytes &&
		ins.x86RawNearReturn == nil && ins.x86AddressBits == 0 && !ins.x86UnboundAddressOverride &&
		ins.x86SegmentPrefix == 0 && !ins.x86RIPLiteral && ins.x86NumericalRead == nil
}

// The paired scalar emitter materializes only producer-specific ISA choices.
// It grants neither public-query definedness nor native FLAGS, entry, caller,
// frame or memory-backing authority. Native RMW retains its exact volatile load
// and store, including zero count and otherwise unobservable chosen outputs.
func x86LoweredScalarShiftChoiceStep(ins Instr, state *x86ResultState, owned []bool,
	resultOffset int64, arch string) (bool, bool) {
	form, known, err := x86ScalarShiftProperties(ins.Op, ins, arch)
	if !known || form.bmi2 {
		return false, false
	}
	if err != nil {
		return false, true
	}
	local := ins
	if !x86ResultNamedPredicateInstruction(local) {
		if !x86RawLegacyScalarShiftSource(ins, arch) {
			return false, true
		}
		local.x86Encoded, local.x86SourceRaw = false, nil
	}
	next := state.clone()
	next.loweredStaticReads = state.loweredStaticReads
	if x86LoweredScalarShiftTransfer(local, form, &next, owned, resultOffset, arch) {
		*state = next
		return true, true
	}
	// Retain only the already closed nontrapping GP fallback. It does not
	// materialize choices or initialize missing dependent bytes/FLAGS. Memory
	// accesses and W-double forms outside that old domain cannot borrow it.
	return x86LoweredPureGPUnknownTransfer(local, state, arch), true
}
