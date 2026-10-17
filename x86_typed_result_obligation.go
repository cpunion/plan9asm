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
	if x86StaticIntegerResultSourceCandidate(fn, sig, arch) {
		return true
	}
	for _, ins := range x86X87FlowInstructions(fn) {
		if _, known, err := x86X87EnvironmentReadProperties(ins, arch); known && err == nil {
			return true // Dead environment reads retain input and stack obligations for Void.
		}
		if form, known, err := x86X87ValueProperties(ins, arch); known && err == nil &&
			(form.spec.kind == x86X87IntegerMove || form.spec.kind == x86X87Round ||
				form.spec.kind == x86X87ConditionalMove) && x86ResultNamedPredicateInstruction(ins) {
			return true // Input/control/stack obligations also apply to Void; private slots cannot supply unknown facts.
		}
	}
	opaqueAddressResult := x86OpaqueAddressPointerResultRole(fn, sig, arch)
	for ordinal, ins := range fn.Instrs {
		if _, known, err := x86PackedIntegerAddProperties(ins, arch); known && err == nil {
			return true // Native add inputs/masks remain obligations without a result or unrelated KMOV.
		}
		if _, known, err := x86PackedSubtractProperties(ins, arch); known && err == nil {
			return true // Subtract shares the same whole-source obligation, not add's caller authority.
		}
		if _, known, err := x86BEXTRProperties(ins.Op, ins, arch); known && err == nil {
			return true // START, LEN and source bytes remain required even for dead BEXTR results.
		}
		if _, known, err := x86BroadcastProperties(ins.Op, ins, arch); known && err == nil {
			return true // Dead scalar/block/mask broadcasts still require their actual input bytes.
		}
		if _, known, err := x86FloatCompareProperties(ins, arch); known && err == nil {
			return true // Dead comparisons still consume real vector/mask inputs and native FP effects.
		}
		if form, known, err := x86NonTemporalVectorProperties(ins.Op, ins, arch); known && err == nil && !form.trap {
			return true // Dead non-temporal accesses still need real memory and complete inputs, including Void.
		}
		if _, known, err := x86ScalarBitCountProperties(ins.Op, ins, arch); known && err == nil &&
			isAMD64MemoryOperand(ins.Args[0]) && x86ResultNamedPredicateInstruction(ins) {
			return true // A dead count still performs its exact native read, including Void.
		}
		if _, known, err := x86NativeMMXInstruction(ins, arch); known && err == nil && x86ResultNamedPredicateInstruction(ins) {
			return true // Actual MMX inputs and the closed x87-mode continuation remain obligations for Void.
		}
		if _, known, err := x86PackedShiftProperties(ins.Op, ins, arch); known && err == nil {
			return true // Native accesses and masks stay obligations even with dead numerical outputs.
		}
		if _, known, err := x86PackedUnpackProperties(ins, arch); known && err == nil {
			return true // Named/physical permutations cannot borrow implementation entry bytes.
		}
		if _, known, err := x86PackedNarrowProperties(ins, arch); known && err == nil {
			return true // Named or physical PACK retains address, mask and result obligations even for Void.
		}
		if _, known, err := x86StackWidthProperties(ins, arch); known && err == nil && x86ResultNamedPredicateInstruction(ins) {
			return true // Private transport must not hide unknown result bytes or unbalanced storage.
		}
		if _, known, err := x86ScalarMemoryRMWProperties(ins.Op, ins, arch); known && err == nil &&
			x86ResultNamedPredicateInstruction(ins) && (ordinal == 0 || fn.Instrs[ordinal-1].Op != "LOCK") {
			return true // Ordinary memory RMW inputs remain obligations even for Void or missing results.
		}
		if _, known, err := x86AHFlagsProperties(ins.Op, ins, arch); known && err == nil && x86ResultNamedPredicateInstruction(ins) {
			return true // Dead AH/FLAGS transfers still need their actual source byte or five flags.
		}
		if _, known, err := x86SignedPackedFloatProperties(ins, arch); known && err == nil && x86ResultNamedPredicateInstruction(ins) {
			return true // Complete conversion inputs stay required for Void without a KMOV routing prerequisite.
		}
		if _, known, err := x86ScalarNativeStoreProperties(ins.Op, ins, arch); known && err == nil &&
			x86ResultNamedPredicateInstruction(ins) {
			return true // Register and immediate stores retain data/address obligations, including Void.
		}
		if _, _, known, err := x86FixedVectorValueProperties(ins, arch); known && err == nil && x86ResultNamedPredicateInstruction(ins) {
			return true // Complete value inputs and native addresses stay required without results or masks.
		}
		if _, known, err := x86ConditionalSetProperties(ins.Op, ins, arch); known && err == nil &&
			x86ResultNamedPredicateInstruction(ins) {
			return true // Every conditional byte write consumes actual predicate FLAGS, even for Void.
		}
		if _, known, err := x86MaskValueProperties(ins.Op, ins, arch); known && err == nil && x86ResultNamedPredicateInstruction(ins) {
			return true // KADD/KSHIFT/KUNPCK retain their complete input obligation for Void.
		}
		if amd64MaskMoveOperandsValid(ins, arch) && x86ResultNamedPredicateInstruction(ins) {
			return true // Mask inputs/results cannot borrow zeroed implementation slots, even for Void.
		}
		if form, known, err := x86PackedCompactProperties(ins.Op, ins, arch); known && err == nil &&
			!form.undefinedOpcode() && x86ResultNamedPredicateInstruction(ins) {
			return true // Dead named compact/expand inputs remain obligations even for Void.
		}
		if _, _, _, known := x86StaticAddressValueMove(ins, arch); known && !opaqueAddressResult {
			return true // The address source cannot hide an unwritten result.
		}
		if x86NamedSegmentMOVNeedsProof(ins, arch) {
			return true // Source/input effects remain mandatory even with no result.
		}
		if _, known, err := x86StringInstructionProperties(ins, arch); known && err == nil && x86ResultNamedPredicateInstruction(ins) {
			return true // Implicit data/address/count/DF inputs remain mandatory, including Void.
		}
		if sig.Ret == Void || len(sig.Frame.Results) == 0 {
			continue
		}
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
