package plan9asm

// Unlike a MOV-store cost selector, a registered numerical family has its own
// whole-result obligation. Missing result stores, unused bool inputs and other
// storage operations cannot substitute for initialized source/result bytes.
// Independently bound entry/storage roles are consumed by the complete cached
// proof; this syntax-only routing decision never grants one.
func x86TypedFPResultNeedsProof(fn Func, sig FuncSig, arch string) bool {
	if arch != "386" && arch != "amd64" || sig.Ret == Void || len(sig.Frame.Results) == 0 {
		return false
	}
	for _, ins := range fn.Instrs {
		if _, _, _, known := x86StaticAddressValueMove(ins, arch); known {
			return true // The address source cannot hide an unwritten result.
		}
		if effect, known, err := x86TypedInstructionByteEffect(ins.Op, ins, arch); known && err == nil &&
			effect.requiresResultProof && x86ResultNamedPredicateInstruction(ins) {
			return true
		}
	}
	return false
}
