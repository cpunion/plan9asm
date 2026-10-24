package plan9asm

// Consume only the actual emitted module's numeric storage and address values.
// The complete shared scalar grammar owns widths, byte views and FLAGS effects;
// the unchanged emitter owns paired volatile accesses. No external symbol or
// declaration-only query gains storage, address or initialized-result facts.
func (c *amd64Ctx) scalarStaticRMWTransfer(ins Instr, state *x86ResultState, owned []bool, arch string) (bool, bool) {
	form, known, err := x86LockScalarRMWProperties(ins.Op, ins, arch)
	if !known || err != nil || form.spec.bitString ||
		!x86ResultNamedPredicateInstruction(ins) || ins.x86NumericalRead != nil {
		return false, false
	}
	destination, static := x86StaticSBReadAddress(form.destination, arch)
	relocation := x86AbsoluteRelocationSource(form.source, arch, form.spec.bits)
	if !static && (!relocation || form.destination.Kind != OpReg) {
		return false, false
	}
	values := c.fpStaticAddressValues
	if values == nil || state == nil {
		return false, true
	}
	width := int64(form.spec.bits / 8)
	if static {
		name := resolveAssemblyDataSymbol(destination.symbol, c.resolve)
		if !values.writable[name] || c.fpScalarData[destination] < width {
			return false, true // A numeric extent alone is not writable storage.
		}
	} else if !x86ResultRead(form.destination, width, state, owned, arch) {
		return false, true
	}
	if relocation {
		address, immediate, valid := x86StaticAddressValueSource(form.source, arch)
		if !valid || !immediate || !values.objects[resolveAssemblyDataSymbol(address.symbol, c.resolve)] {
			return false, true
		}
	} else if !form.spec.unary && !x86ResultRead(form.source, width, state, owned, arch) {
		return false, true
	}
	if (form.spec.stem == "ADC" || form.spec.stem == "SBB") && state.flags&x86ByteFlagCF == 0 {
		return false, true
	}
	if !static {
		index, _, valid := x86ResultGP(form.destination.Reg, width, arch)
		if !valid || state.pointerGP[index] || !x86ResultStore(form.destination, width, state, owned, 0, arch) {
			return false, true
		}
		// Relocation arithmetic is mode32/L only. Its full numerical result
		// and architectural upper zeroes are not a canonical bool or pointer.
		state.canonicalGP[index] = 0xf0
		state.callerAddressGP[index], state.callerAddressOrigin[index] = false, 0
		state.resultAddressGP[index] = x86ResultSlotAddress{}
		if state.boolFrameGP != nil {
			state.boolFrameGP[index] = x86ResultSlotAddress{}
		}
	}
	state.flags = state.flags&^form.spec.flagsKilled | form.spec.flagsDefined
	return true, true // No new memory-read capability or unrelated state escapes.
}
