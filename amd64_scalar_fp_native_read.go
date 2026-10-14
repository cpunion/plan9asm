package plan9asm

import "strings"

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
	if fn.x86FixedLocal != nil {
		return c.checkFixedLocalResultInitialization()
	}
	if calls := fn.x86BoolCalls; calls != nil && calls.tail != nil {
		return calls.checkTailTerminal(fn, sig, arch)
	}
	strictErr := checkX86ScalarFPResultInitialization(fn, sig, arch)
	customRegisterBool := arch == "386" && sig.Ret == I1 && len(sig.ArgRegs) != 0
	if customRegisterBool {
		strictErr = checkX86CustomRegisterBoolResultInitialization(fn, sig, arch,
			func(_ int, ins Instr, state *x86ResultState, owned []bool, resultOffset int64, arch string) bool {
				return x86ScalarFPResultTransfer(ins, state, owned, resultOffset, arch)
			})
	}
	registerEntry := arch == "amd64" && strings.HasSuffix(fn.Sym, "<ABIInternal>")
	if registerEntry {
		strictErr = checkAMD64RegisterResultInitialization(fn, sig,
			func(_ int, ins Instr, state *x86ResultState, owned []bool, resultOffset int64, arch string) bool {
				return x86ScalarFPResultTransfer(ins, state, owned, resultOffset, arch)
			})
	}
	if strictErr == nil {
		return nil
	}
	// Native/raw entry contracts are established before ordinary lowering.
	// Do not let this conditional byte rule create or supplement one.
	if len(fn.X86RawText) != 0 || fn.x86SourceStack != nil || fn.x86FixedLocal != nil {
		return strictErr
	}
	for _, ins := range fn.Instrs {
		if ins.x86Encoded || ins.x86RawNearReturn != nil ||
			ins.x86AddressBits != 0 || ins.x86UnboundAddressOverride {
			return strictErr
		}
		if origin := ins.x86SourceRaw; origin != nil {
			// The physical decoder already validated the complete NOP and
			// its prefixes before discarding inert ModRM/SIB address fields.
			// Ordinary padding has no memory, register, FLAGS or entry effect.
			// Preserve the sealed original/normalized identity; no other raw
			// instruction may borrow a named load's conditional result rule.
			if origin.op != "NOP" || len(origin.args) != 0 || ins.Op != "NOP" || len(ins.Args) != 0 {
				return strictErr
			}
		}
	}
	transfer := func(ordinal int, ins Instr, state *x86ResultState, owned []bool, resultOffset int64, arch string) bool {
		state.loweredStaticReads = c.fpScalarData
		defer func() { state.loweredStaticReads = nil }()
		if accepted, known := x86ResultSignedIntegerFloatStepWithRead(ins, state, owned, arch, true); known {
			return accepted // The same typed emitter retains an exact volatile source read.
		}
		if customRegisterBool {
			if accepted, known := x86CustomRegisterBoolCompareExchangeStep(fn, ordinal, ins, state, arch); known {
				return accepted
			}
		}
		if accepted, known := x86ResultPackedStringStepWithRead(ins, state, owned, arch, true); known {
			return accepted // Paired emitter retains the full volatile m128 read.
		}
		if accepted, known := x86ResultPackedZeroExtendStep(ins, state, owned, arch, true); known {
			return accepted
		}
		if accepted, recognized := x86LoweredBlockInsertRead(ins, state, owned, resultOffset, arch); recognized {
			return accepted
		}
		if accepted, recognized := x86LoweredScalarExtensionRead(ins, state, owned, resultOffset, arch); recognized {
			return accepted
		}
		if accepted, recognized := x86LoweredScalarNativeLoad(ins, state, owned, resultOffset, arch); recognized {
			return accepted
		}
		if accepted, recognized := x86LoweredScalarArithmeticNativeRead(ins, state, owned, resultOffset, arch); recognized {
			return accepted
		}
		if accepted, recognized := x86ResultScalarLaneStepWithRead(ins, state, owned, resultOffset, arch, true); recognized {
			return accepted
		}
		if form, known, err := x86GPSIMDTransferProperties(ins, arch); known && err == nil && form.memory == 0 {
			return x86ResultGPSIMDTransferWithRead(ins, form, state, owned, resultOffset, arch, true)
		}
		if accepted, known := c.scalarCompareNativeResultTransfer(ins, state, owned, arch); known {
			return accepted
		}
		if form, known, err := amd64PackedVectorMoveProperties(ins, arch); known && err == nil &&
			form.memory == 0 && form.ordinaryMemory(ins, arch) &&
			x86ResultIncomingPointerAddressIdentity(ins.Args[0], state, arch) {
			// The emitter preserves the original aligned opcode, or the
			// unaligned full/guarded tuple with volatile accesses. If it continues,
			// active bytes are numerical; inactive bytes still require the
			// typed merge/zero rule. No pointer validity is asserted.
			return x86ResultPackedVectorMoveTransfer(ins, form, state, owned, resultOffset, arch, true)
		}
		form, known, err := x86LegacyPackedMoveProperties(ins.Op, ins, arch)
		if known && err == nil && form.memory == 0 && form.source.Kind == OpMem &&
			form.source.Mem.Sym == "" && x86OrdinaryNativeMemoryOperand(form.source, arch) {
			// This same typed form selects the original sideeffect opcode
			// in lowerLegacyPackedMove. Only reaching incoming-pointer
			// identity and initialized numerical indexes supply an address.
			// No native store, static symbol, virtual frame or unknown base
			// gains a numerical input from this load-only rule.
			return x86ResultIncomingPointerAddressIdentity(form.source, state, arch) &&
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
		return x86LoweredPureGPUnknownTransfer(ins, state, arch)
	}
	if registerEntry {
		return checkAMD64RegisterResultInitialization(fn, sig, transfer)
	}
	if customRegisterBool {
		return checkX86CustomRegisterBoolResultInitialization(fn, sig, arch, transfer)
	}
	return checkX86ABI0ResultInitializationAtSource(fn, sig, arch, transfer, true)
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
		x86OrdinaryNativeMemoryOperand(source, arch) && x86ResultIncomingPointerAddressIdentity(source, state, arch)
}
