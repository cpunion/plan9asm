package plan9asm

// This is a lowering-only numerical availability check. An exact side-effecting
// native load either faults, or produces all 16 destination bytes. It does not
// assert that the pointer is valid, aligned, in bounds, nonaliasing or backed by
// initialized Go storage. Those conditions must not be optimizer assumptions.
//
// The declaration-only observers cannot use this rule: without the paired
// original native opcode emission, their query supplies no such conditional
// result. In particular, this never creates a caller-memory or raw-RET token.
func (c *amd64Ctx) checkLoweredScalarFPResultInitialization() error {
	fn, sig, arch := c.fpScalarSource, c.sig, c.goarch
	strictErr := checkX86ScalarFPResultInitialization(fn, sig, arch)
	if strictErr == nil {
		return nil
	}
	// Native/raw entry contracts are established before ordinary lowering.
	// Do not let this conditional byte rule create or supplement one.
	if len(fn.X86RawText) != 0 || fn.x86SourceStack != nil || fn.x86FixedLocal != nil {
		return strictErr
	}
	for _, ins := range fn.Instrs {
		if ins.x86Encoded || ins.x86SourceRaw != nil || ins.x86RawNearReturn != nil ||
			ins.x86AddressBits != 0 || ins.x86UnboundAddressOverride {
			return strictErr
		}
	}
	return checkX86ABI0ResultInitializationAtSource(fn, sig, arch,
		func(_ int, ins Instr, state *x86ResultState, owned []bool, resultOffset int64, arch string) bool {
			state.loweredStaticReads = c.fpScalarData
			defer func() { state.loweredStaticReads = nil }()
			form, known, err := x86LegacyPackedMoveProperties(ins.Op, ins, arch)
			if known && err == nil && form.memory == 0 && form.source.Kind == OpMem &&
				x86LegacyPackedMoveOrdinaryNativeMemory(form.source) {
				// This same typed form selects the original sideeffect opcode
				// in lowerLegacyPackedMove. Only reaching incoming-pointer
				// identity and initialized numerical indexes supply an address.
				// No native store, static symbol, virtual frame or unknown base
				// gains a numerical input from this load-only rule.
				return x86ResultIncomingPointerAddressIdentity(form.source, state, arch) &&
					x86ResultStore(form.destination, 16, state, owned, resultOffset, arch)
			}
			return x86ScalarFPResultTransfer(ins, state, owned, resultOffset, arch)
		}, true)
}

// The ordinary native branch is shared with the actual emitter. Certificates
// can additionally bind physical SP/BP in that emitter; they are not admitted
// by the independent incoming-pointer result rule above.
func x86LegacyPackedMoveOrdinaryNativeMemory(memory Operand) bool {
	if memory.Kind == OpSym {
		return true
	}
	return memory.Kind == OpMem && memory.Mem.Base != SP && memory.Mem.Base != BP &&
		memory.Mem.Index != SP && memory.Mem.Index != BP
}
