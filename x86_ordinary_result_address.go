package plan9asm

// This is a cost predicate, not storage or initialization authority. A named
// result-field FP LEA must not bypass the cached whole-source
// proof merely because no GP/X transfer or scalar-vector probe follows it.
func x86OrdinaryResultAddressNeedsProof(fn Func, sig FuncSig, arch string) bool {
	if arch != "386" && arch != "amd64" {
		return false
	}
	for _, ins := range fn.Instrs {
		if (ins.Op == "LEAW" || ins.Op == "LEAL" || ins.Op == "LEAQ") && len(ins.Args) == 2 && ins.Args[0].Kind == OpFP {
			for _, slot := range sig.Frame.Results {
				width := x86ResultScalarBytes(slot.Type, arch)
				if x86IntegerResultSlot(slot) && slot.Offset >= 0 && slot.Offset <= 4096-width &&
					ins.Args[0].FPOffset >= slot.Offset && ins.Args[0].FPOffset < slot.Offset+width {
					return true
				}
			}
		}
	}
	return false
}

func x86IntegerResultSlot(slot FrameSlot) bool {
	return slot.Type == I8 || slot.Type == I16 || slot.Type == I32 || slot.Type == I64
}

// Only the private lowering observer has logical result backing. Its existing
// typed layout validation and bounded must-state proof remain mandatory. No
// native incoming frame, parameter address, padding, CALL, raw entry, caller
// extent, or canonical bool role is inferred here.
func x86OrdinaryResultAddressRole(fn Func, sig FuncSig, arch string) bool {
	if !x86OrdinaryResultAddressNeedsProof(fn, sig, arch) || fn.FrameSize != 0 ||
		len(fn.X86RawText) != 0 || fn.x86SourceStack != nil || fn.x86FixedLocal != nil ||
		fn.x86RawReplay != nil || len(fn.x86ContinuationAddresses) != 0 || len(fn.x86IndirectLabels) != 0 {
		return false
	}
	for _, slot := range sig.Frame.Results {
		if !x86IntegerResultSlot(slot) {
			return false
		}
	}
	for _, ins := range fn.Instrs {
		if ins.Op == "CALL" || isX86RawDirective(ins) || ins.x86Encoded || ins.x86SourceRaw != nil ||
			ins.x86RawNearReturn != nil || ins.x86AddressBits != 0 || ins.x86UnboundAddressOverride || ins.x86SegmentPrefix != 0 {
			return false
		}
	}
	return true
}
