package plan9asm

import (
	"reflect"
	"strings"
)

// This is a lowering-only numerical availability check. A paired original
// opcode or volatile unaligned load either faults or produces its active
// destination bytes; masked inactive bytes retain their separate rule. It does not
// assert that the pointer is valid, aligned, in bounds, nonaliasing or backed by
// initialized Go storage. Those conditions must not be optimizer assumptions.
//
// The declaration-only observers cannot use this rule: without the paired
// original native opcode emission, their query supplies no such conditional
// result. In particular, this never creates a caller-memory or raw-RET token.
func (c *amd64Ctx) checkLoweredScalarFPResultInitialization() error {
	fn, sig, arch := c.fpScalarSource, c.sig, c.goarch
	var nativeMMX *x86NativeMMXRegion
	named := true
	for _, ins := range fn.Instrs {
		named = named && x86ResultNamedPredicateInstruction(ins)
	}
	if named {
		var err error
		nativeMMX, err = newX86NativeMMXRegion(fn, c.fpSourceSig, arch, c.targetTriple)
		if err != nil {
			return err
		}
	}
	if owned, err := c.checkCustomHelperBoolSource(); owned || err != nil {
		return err
	}
	if x86StaticIntegerResultSourceCandidate(fn, sig, arch) {
		return c.checkStaticIntegerResultInitialization()
	}
	if fn.x86ContinuationFlow != nil {
		if err := c.checkContinuationResultSource(); err != nil {
			return err
		}
		sig = c.fpSourceSig
	}
	if fn.x86FixedLocal != nil {
		return c.checkFixedLocalResultInitialization()
	}
	if fn.x86SourceStack != nil && fn.x86SourceStack.scalarResult != nil {
		return c.checkSourceStackScalarResultInitialization()
	}
	if calls := fn.x86BoolCalls; calls != nil && calls.registerTail != nil {
		return calls.checkCustomRegisterTail(fn, sig, arch)
	}
	if calls := fn.x86BoolCalls; calls != nil && calls.tail != nil {
		return calls.checkTailTerminal(fn, sig, arch)
	}
	strictErr := checkX86ScalarFPResultInitialization(fn, sig, arch)
	customRegisterBool := arch == "386" && sig.Ret == I1 && len(sig.ArgRegs) != 0
	logicalScalar := x86LogicalScalarResultNeedsProof(sig, arch)
	if customRegisterBool || logicalScalar {
		strictErr = checkX86LogicalRegisterResultInitialization(fn, sig, arch,
			func(_ int, ins Instr, state *x86ResultState, owned []bool, resultOffset int64, arch string) bool {
				return x86ScalarFPResultTransfer(ins, state, owned, resultOffset, arch)
			}, customRegisterBool, false)
	}
	registerEntry := arch == "amd64" && strings.HasSuffix(fn.Sym, "<ABIInternal>")
	if registerEntry {
		strictErr = checkAMD64RegisterResultInitialization(fn, sig,
			func(_ int, ins Instr, state *x86ResultState, owned []bool, resultOffset int64, arch string) bool {
				return x86ScalarFPResultTransfer(ins, state, owned, resultOffset, arch)
			})
	}
	if strictErr == nil && nativeMMX == nil {
		return nil
	}
	if selected, err := c.checkRawLiteralResults(); selected {
		return err
	}
	if x86TrapOnlyLogicalFPSourceRole(fn, sig, arch) {
		if err := c.checkTrapOnlyLogicalFPResultInitialization(); err == nil {
			c.fpScalarTrapOnly = true
			return nil
		}
	}
	// Native/raw entry contracts are established before ordinary lowering.
	// Do not let this conditional byte rule create or supplement one.
	if len(fn.X86RawText) != 0 || fn.x86SourceStack != nil || fn.x86FixedLocal != nil {
		return strictErr
	}
	boundRawClear := false
	for _, ins := range fn.Instrs {
		if x86RawADXSource(ins, arch) {
			continue // Only the bound default address mode shares the typed carry chain.
		}
		if x86RawScalarObserverSource(ins, arch) {
			continue // Existing typed count/mask effects retain all numerical inputs.
		}
		if x86RawVectorZeroSource(ins, arch) {
			continue // Exact implicit SIMD writes preserve independent native input roles.
		}
		if x86RawBMI2ShiftSource(ins, arch) {
			continue // Exact shifts/rotates retain their complete count/data obligations.
		}
		if x86RawAccumulatorSignExtendSource(ins, arch) {
			continue // Exact implicit GP arithmetic preserves its independent input obligations.
		}
		if x86RawDwordProductResultSource(ins, arch) {
			continue // Exact decoded integer inputs retain their own mask/address obligations.
		}
		if _, valid := x86RawVMOVWResultForm(ins, arch); valid {
			continue // Exact GP/X/m16 effects share no raw entry or native frame permission.
		}
		if _, valid := x86RawFP16ConversionSourceForm(ins, arch); valid {
			continue // Exact conversion identity retains independent input and mask obligations.
		}
		if x86RawFP16SqrtSource(ins, arch) {
			continue // Square-root inputs remain guarded by the shared typed byte effect.
		}
		if _, valid := x86RawFP16FMASourceForm(ins, arch); valid {
			continue // FMA retains all three inputs, including a zeroing destination.
		}
		if _, valid := x86RawFP16BinarySourceForm(ins, arch); valid {
			continue // Binary input/merge/upper bytes retain their own typed effect.
		}
		if _, valid := x86EnqueueSourceForm(ins, arch); valid {
			continue // Exact device submission retains its own two-address native effect.
		}
		if x86EndBranchPreservesSourceState(ins) {
			continue // Exact native CET tracking preserves inputs, not memory or entry permissions.
		}
		if ins.x86Encoded || ins.x86RawNearReturn != nil ||
			ins.x86AddressBits != 0 || ins.x86UnboundAddressOverride {
			return strictErr
		}
		if ins.x86SourceRaw != nil {
			if x86MaskLogicalPreservesStorage(ins) {
				continue // Exact K-only source has independent mask/FLAGS effects, not memory or entry authority.
			}
			if x86RegisterQueryPreservesSourceState(ins) {
				// Exact original/normalized register producers share their
				// checked inputs/outputs. They add no memory,
				// raw-return, stack or native-entry authority to this continuation.
				continue
			}
			if x86X87RawClearInstruction(ins) {
				if !boundRawClear && (fn.x86RawReplay == nil || fn.x86RawReplay.validate(fn, arch) != nil) {
					return strictErr
				}
				boundRawClear = true
				continue
			}
			// The physical decoder already validated the complete NOP and
			// its prefixes before discarding inert ModRM/SIB address fields.
			// Ordinary padding has no memory, register, FLAGS or entry effect.
			// Preserve the sealed original/normalized identity; no other raw
			// instruction may borrow a named load's conditional result rule.
			if !x86NOPPreservesSourceState(ins) {
				return strictErr
			}
		}
	}
	lockPrefixes := x86ResultSourceLockPrefixes(fn, arch)
	addressBitStores := x86LeafNativeAddressBitStoreScope(fn, c.fpSourceSig, arch)
	stringMemory, err := x86ResultStringMemorySourcePlan(fn, arch)
	if err != nil {
		return err
	}
	transfer := func(ordinal int, ins Instr, state *x86ResultState, owned []bool, resultOffset int64, arch string) bool {
		state.loweredStaticReads = c.fpScalarData
		defer func() { state.loweredStaticReads = nil }()
		if accepted, known := x86ResultSDKLiteralStep(ins, state, owned, resultOffset, arch); known {
			return accepted // Consume the exact bound source before an overlapping native-memory family.
		}
		if accepted, known := x86ResultNumericalReadStep(ins, state, owned, resultOffset, arch); known {
			return accepted // A checked mutable SB source keeps its own lifetime and byte effects.
		}
		if accepted, known := x86LoweredADXResultStep(ins, state, owned, resultOffset, arch); known {
			return accepted
		}
		if accepted, known := x86LoweredRawScalarObserverStep(ins, state, owned, resultOffset, arch); known {
			return accepted
		}
		if accepted, known := x86LoweredVectorZeroResultStep(ins, state, owned, resultOffset, arch); known {
			return accepted
		}
		if accepted, known := x86LoweredBMI2ShiftResultStep(ins, state, owned, resultOffset, arch); known {
			return accepted
		}
		if accepted, known := x86LoweredAccumulatorSignExtendResultStep(ins, state, owned, resultOffset, arch); known {
			return accepted
		}
		if accepted, known := x86LoweredDwordProductResultStep(ins, state, owned, arch); known {
			return accepted
		}
		if accepted, known := x86LoweredRawVMOVWResultStep(ins, state, owned, resultOffset, arch); known {
			return accepted
		}
		if accepted, known := x86LoweredFP16ConversionStep(ins, state, owned, arch); known {
			return accepted
		}
		if accepted, known := x86LoweredFP16FMAStep(ins, state, owned, arch); known {
			return accepted
		}
		if accepted, known := x86LoweredFP16BinaryStep(ins, state, owned, arch); known {
			return accepted
		}
		if accepted, known := x86LoweredEnqueueResultStep(ins, state, arch); known {
			return accepted
		}
		if accepted, known := nativeMMX.transfer(ordinal, ins, state, owned, resultOffset, arch); known {
			return accepted
		}
		if accepted, known := x86LoweredNativeMXCSRStep(ins, state, arch); known {
			return accepted
		}
		if accepted, known := c.scalarStaticAddressValueTransfer(ins, state, owned, arch); known {
			return accepted
		}
		if accepted, known := c.scalarStaticRMWTransfer(ins, state, owned, arch); known {
			return accepted
		}
		if accepted, known := stringMemory.transfer(ordinal, ins, state, arch); known {
			return accepted
		}
		if ins.Op == "LOCK" {
			prefix, bound := lockPrefixes[ordinal]
			return bound && reflect.DeepEqual(prefix, ins)
		}
		if boundRawClear && x86X87RawClearInstruction(ins) {
			return x86ResultX87BoundRawClear(ins, state, owned, resultOffset, arch)
		}
		if accepted, known := c.x87EnvironmentReadResultStep(ins, state, arch); known {
			return accepted
		}
		if accepted, known := c.x87NativeFloatingMoveResultStep(ins, state, owned, resultOffset, arch); known {
			return accepted
		}
		if accepted, known := x86LoweredStackNativeRead(ins, state, owned, resultOffset, arch); known {
			return accepted
		}
		if accepted, known := c.x87HardwareControlResultStep(ins, state, owned, resultOffset, arch); known {
			return accepted // The same actual hardware emitter writes the full i16.
		}
		if accepted, known := c.x87HardwareRoundingResultStep(ins, state, owned, resultOffset, arch); known {
			return accepted // No ambient environment fact reaches the standalone query.
		}
		if accepted, known := x86LoweredCompareExchangeResultStepWithAddressBits(ins, state, owned, arch, addressBitStores); known {
			return accepted
		}
		if accepted, known := x86LoweredExchangeResultStep(ins, state, owned, arch); known {
			return accepted
		}
		if accepted, known := x86LoweredCRC32NativeRead(ins, state, owned, resultOffset, arch); known {
			return accepted // The same typed emitter retains the exact volatile source read.
		}
		if accepted, known := x86LoweredScalarBitCountNativeRead(ins, state, owned, resultOffset, arch); known {
			return accepted
		}
		if accepted, known := x86ResultScalarBitFieldStep(ins, state, owned, resultOffset, arch, true); known {
			return accepted
		}
		if accepted, known := x86LoweredBLSNativeRead(ins, state, owned, resultOffset, arch); known {
			return accepted
		}
		if accepted, known := x86ResultNonTemporalVectorStep(ins, state, owned, resultOffset, arch, true); known {
			return accepted
		}
		if accepted, known := x86ResultFloatCompareStep(ins, state, owned, arch, true); known {
			return accepted
		}
		if accepted, known := x86ResultSqrtStep(ins, state, owned, arch, true); known {
			return accepted // The paired native opcode retains scalar/packed MXCSR effects.
		}
		if accepted, known := x86ResultCOMIStepWithRead(ins, state, owned, arch, true); known {
			return accepted // The original scalar comparison retains exact read and MXCSR effects.
		}
		if accepted, known := x86ResultSignedIntegerFloatStepWithRead(ins, state, owned, arch, true); known {
			return accepted // The same typed emitter retains an exact volatile source read.
		}
		if accepted, known := x86ResultSignedPackedFloatStep(ins, state, owned, arch, true); known {
			return accepted // The original native opcode retains rounding, exception and memory effects.
		}
		if accepted, known := x86ResultPackedConversionStepWithRead(ins, state, owned, arch, true); known {
			return accepted // The paired conversion retains its exact access and MXCSR effects.
		}
		if form, known, err := x86FloatAddSubProperties(ins, arch); known {
			return err == nil && x86ScalarFPFloatBinaryTransferWithRead(ins, form.x86FloatBinaryForm, state, owned, arch, true)
		}
		if form, known, err := x86FloatMulDivProperties(ins, arch); known {
			return err == nil && x86ScalarFPFloatBinaryTransferWithRead(ins, form.x86FloatBinaryForm, state, owned, arch, true)
		}
		if accepted, known := x86ResultPackedStringStepWithRead(ins, state, owned, arch, true); known {
			return accepted // Paired emitter retains the full volatile m128 read.
		}
		if accepted, known := x86ResultPackedZeroExtendStep(ins, state, owned, arch, true); known {
			return accepted
		}
		if accepted, known := x86ResultByteShuffleStepWithRead(ins, state, owned, arch, true); known {
			return accepted // Same full-tuple E4NF read or original legacy m128 opcode.
		}
		if accepted, known := x86ResultLegacyIntegerNativeReadStep(ins, state, owned, resultOffset, arch); known {
			return accepted
		}
		if accepted, recognized := x86LoweredBlockNativeMemory(ins, state, owned, resultOffset, arch); recognized {
			return accepted
		}
		if accepted, recognized := x86LoweredScalarExtensionRead(ins, state, owned, resultOffset, arch); recognized {
			return accepted
		}
		if accepted, recognized := x86LoweredScalarNativeLoad(ins, state, owned, resultOffset, arch); recognized {
			return accepted
		}
		if accepted, recognized := x86LoweredScalarNativeStoreWithAddressBits(ins, state, owned, arch, addressBitStores); recognized {
			return accepted
		}
		if accepted, recognized := x86LoweredScalarMemoryRMW(ins, state, owned, arch); recognized {
			return accepted
		}
		if accepted, known := x86LoweredScalarShiftMemory(ins, state, owned, resultOffset, arch); known {
			return accepted
		}
		if accepted, known := x86MaskMoveResultStep(ins, state, owned, resultOffset, arch, true); known {
			return accepted // The same KMOV emitter retains the complete volatile native width.
		}
		if accepted, known := x86LoweredNativeConditionalSet(ins, state, arch); known {
			return accepted // The same typed SETcc emitter preserves its exact volatile byte write.
		}
		if accepted, known := x86LoweredFixedVectorNativeRead(ins, state, owned, arch); known {
			return accepted // The typed emitter retains full E4NF or guarded E4 native accesses.
		}
		if accepted, known := x86ResultSADStep(ins, state, owned, arch, true); known {
			return accepted // The same SAD emitter retains the full tuple and legacy m128 opcode.
		}
		if accepted, known := x86ResultPackedNarrowStep(ins, state, owned, arch, true); known {
			return accepted // The same emitter retains the complete E4NF tuple or original legacy opcode.
		}
		if accepted, known := x86ResultPackedShiftStepWithRead(ins, state, owned, arch, true); known {
			return accepted // Uniform m128, E4NF tuples and guarded E4 reads retain their independent shapes.
		}
		if accepted, known := x86ResultPackedUnpackStepWithRead(ins, state, owned, arch, true); known {
			return accepted // Unpack retains the full E4NF tuple even with zero masks or dead results.
		}
		if accepted, known := x86ResultVectorArithmeticStepWithRead(ins, state, owned, arch, true); known {
			return accepted // The same add/sub emitter retains exact E4 reads and inactive-lane suppression.
		}
		if accepted, known := x86ResultVectorCompareStepWithRead(ins, state, owned, arch, true); known {
			return accepted // Complete vector/K-result compares retain their exact native source accesses.
		}
		if len(ins.Args) != 0 && ins.Args[0].Kind == OpMem && ins.Args[0].Mem.Sym == "" {
			if accepted, known := x86ResultBroadcastStepWithRead(ins, state, owned, arch, true); known {
				return accepted // Only scalar/block native reads need this paired E6 continuation.
			}
		}
		if form, known, err := x86PackedLogicalProperties(ins, arch); known {
			// Only the paired emitter supplies these conditional native values.
			// Standalone queries and physical MMX keep their separate contracts.
			return err == nil && x86ScalarFPLogicalTransferWithRead(ins, form, state, owned, arch, true)
		}
		if accepted, recognized := x86LoweredScalarArithmeticNativeRead(ins, state, owned, resultOffset, arch); recognized {
			return accepted
		}
		if accepted, known := c.scalarPackedLogicalFPWriteTransfer(ins, state, owned, arch); known {
			return accepted
		}
		if accepted, known := c.scalarPSLogicalFPWriteTransfer(ins, state, owned, arch); known {
			return accepted
		}
		if accepted, known := c.scalarBitTestFPShadowTransfer(ins, state, owned, arch); known {
			return accepted
		}
		if accepted, known := x86LoweredNativeBitTest(ins, state, owned, arch); known {
			return accepted
		}
		if accepted, recognized := x86ResultScalarLaneStepWithRead(ins, state, owned, resultOffset, arch, true); recognized {
			return accepted
		}
		if form, known, err := x86GPSIMDTransferProperties(ins, arch); known && err == nil && form.memory >= 0 {
			return x86ResultGPSIMDTransferWithNativeMemory(ins, form, state, owned, resultOffset, arch, true)
		}
		if accepted, known := c.scalarCompareNativeResultTransfer(ins, state, owned, arch); known {
			return accepted
		}
		if accepted, known := x86ResultCompactExpandMemoryStep(ins, state, owned, resultOffset, arch, true); known {
			return accepted // Same guarded native emitter; public observers keep logical-only memory.
		}
		if form, known, err := amd64PackedVectorMoveProperties(ins, arch); known && err == nil &&
			form.ordinaryMemory(ins, arch) && x86ResultNamedPredicateInstruction(ins) &&
			(form.memory == 0 && x86ResultIncomingPointerReadAddressIdentity(ins.Args[0], state, arch) ||
				form.memoryWrite && x86ResultIncomingPointerAddressIdentity(ins.Args[form.memory], state, arch)) {
			// The emitter preserves the original aligned opcode, or the
			// unaligned full/guarded tuple with volatile accesses. If it continues,
			// active bytes are numerical; inactive bytes still require the
			// typed merge/zero rule. No pointer validity is asserted.
			return x86ResultPackedVectorMoveTransfer(ins, form, state, owned, resultOffset, arch, true)
		}
		form, known, err := x86LegacyPackedMoveProperties(ins.Op, ins, arch)
		if known && err == nil && form.memoryWrite && x86ResultNamedPredicateInstruction(ins) &&
			form.destination.Kind == OpMem && form.destination.Mem.Sym == "" &&
			x86OrdinaryNativeMemoryOperand(form.destination, arch) {
			// All seven whole128 legacy names retain their exact native
			// store opcode and its alignment/fault contract. This consumes
			// numerical source bytes plus an incoming-pointer address only;
			// no logical storage, validity, extent or later-read fact follows.
			return x86ResultRead(form.source, 16, state, owned, arch) &&
				x86ResultIncomingPointerAddressIdentity(form.destination, state, arch)
		}
		if known && err == nil && form.memory == 0 && form.source.Kind == OpMem &&
			form.source.Mem.Sym == "" && x86OrdinaryNativeMemoryOperand(form.source, arch) {
			// This same typed form selects the original sideeffect opcode
			// in lowerLegacyPackedMove. Only reaching incoming-pointer
			// identity and initialized numerical indexes supply an address.
			// No native store, static symbol, virtual frame or unknown base
			// gains a numerical input from this load-only rule.
			return x86ResultIncomingPointerReadAddressIdentity(form.source, state, arch) &&
				x86ResultStore(form.destination, 16, state, owned, resultOffset, arch)
		}
		next := state.clone()
		// A reaching-state clone deliberately drops temporary readers.
		// This isolated attempt is still inside the same lowering transfer.
		next.loweredStaticReads = state.loweredStaticReads
		if x86ScalarFPResultTransfer(ins, &next, owned, resultOffset, arch) {
			*state = next
			return true
		}
		if accepted, known := x86ResultUnusedFPAddressStep(ins, state, owned, resultOffset, arch); known {
			return accepted // Mapped logical storage only; no initialized address bits.
		}
		if accepted, known := x86ResultUnknownSBAddressStep(ins, state, arch); known {
			return accepted // Address calculation only; written address bits stay unknown.
		}
		return x86LoweredPureGPUnknownTransfer(ins, state, arch)
	}
	if registerEntry {
		return checkAMD64RegisterResultInitialization(fn, sig, transfer)
	}
	if customRegisterBool || logicalScalar {
		return checkX86LogicalRegisterResultInitialization(fn, sig, arch, transfer, customRegisterBool, true)
	}
	if err := checkX86ABI0ResultInitializationWithTrapOnly(fn, sig, arch, transfer, true, false, true); err != nil {
		return err
	}
	c.nativeMMXRegion = nativeMMX
	return nil
}

// Only the lowering observer may consume a paired exact-width volatile load
// from a typed emitter. Continued execution supplies active numerical source
// bytes, even if a later selector discards them; masked inactive bytes keep
// the family's independent merge/zero rule.
// Family-specific lane/zero/alias effects remain independent. This grants no
// validity, bounds, write, frame, caller or raw-entry permission.
func x86ResultReadWithNativeContinuation(ins Instr, source Operand, width int64, state *x86ResultState, owned []bool, arch string, conditionalNativeRead bool) bool {
	if x86ResultRead(source, width, state, owned, arch) {
		return true
	}
	return conditionalNativeRead && source.Kind == OpMem && source.Mem.Sym == "" &&
		ins.x86NumericalRead == nil && !ins.x86Encoded && ins.x86SourceRaw == nil && ins.x86RawNearReturn == nil &&
		ins.x86AddressBits == 0 && !ins.x86UnboundAddressOverride && ins.x86SegmentPrefix == 0 &&
		x86OrdinaryNativeMemoryOperand(source, arch) && x86ResultIncomingPointerReadAddressIdentity(source, state, arch)
}
