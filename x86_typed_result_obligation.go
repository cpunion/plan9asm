package plan9asm

// Unlike a MOV-store cost selector, a registered numerical family has its own
// whole-result obligation. Missing result stores, unused bool inputs and other
// storage operations cannot substitute for initialized source/result bytes.
// Independently bound entry/storage roles are consumed by the complete cached
// proof; this syntax-only routing decision never grants one.
func x86TypedFPResultNeedsProof(fn Func, sig FuncSig, arch string) bool {
	if arch != "386" && arch != "amd64" {
		return false
	}
	for _, ins := range x86X87FlowInstructions(fn) {
		if _, known, _ := x86X87EnvironmentReadProperties(ins, arch); known {
			return true // Dead environment reads retain input and stack obligations for Void.
		}
		if form, known, err := x86X87ValueProperties(ins, arch); known && err == nil &&
			(form.spec.kind == x86X87IntegerMove || form.spec.kind == x86X87Round) && x86ResultNamedPredicateInstruction(ins) {
			return true // Input/control/stack obligations also apply to Void; private slots cannot supply unknown facts.
		}
	}
	if sig.Ret == Void || len(sig.Frame.Results) == 0 {
		return false // Other families retain their existing result-only routing.
	}
	for _, ins := range fn.Instrs {
		if _, known, err := x86X87WordProperties(ins, arch); known && err == nil && x86ResultNamedPredicateInstruction(ins) {
			return true // Word results need the same whole proof without a GP-to-FP MOV.
		}
		if effect, known, err := x86TypedInstructionByteEffect(ins.Op, ins, arch); known && err == nil &&
			effect.requiresResultProof && x86ResultNamedPredicateInstruction(ins) {
			return true
		}
	}
	return false
}
