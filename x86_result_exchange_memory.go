package plan9asm

import "strings"

// This is the numerical continuation of the already paired, complete native
// XADD/XCHG sideeffect opcode. The read/write cycle retains native faults and
// ordering. It grants no reusable pointee, write envelope, alignment, frame,
// caller, raw-entry or declaration-query permission.
func x86LoweredExchangeResultStep(ins Instr, state *x86ResultState, owned []bool, arch string) (bool, bool) {
	name := Op(strings.ToUpper(string(ins.Op)))
	spec, known := amd64ExchangeSpecs[name]
	if !known {
		return false, false
	}
	form, err := parseAMD64ExchangeForm(arch, name, spec, ins)
	if err != nil || !x86ResultNamedPredicateInstruction(ins) {
		return false, true
	}
	if form.other.Kind == OpReg && spec.kind == amd64ExchangeSwap {
		return false, false // Keep the independent register/opaque/NOP consumer.
	}
	width := int64(spec.bits / 8)
	if !x86ResultRead(form.register, width, state, owned, arch) {
		return false, true
	}
	outputs := []Operand{form.register}
	if form.other.Kind == OpReg {
		if !x86ResultRead(form.other, width, state, owned, arch) {
			return false, true
		}
		outputs = append(outputs, form.other)
	} else if !x86ResultIncomingPointerAddressIdentity(form.other, state, arch) {
		return false, true // No FP/AUTO, SB, segment or numerical base witness.
	}
	for _, operand := range outputs {
		if _, _, valid := x86ResultGP(operand.Reg, width, arch); !valid {
			return false, true
		}
	}
	// Snapshot every dependency before any partial or aliased writeback.
	for _, operand := range outputs {
		index, mask, _ := x86ResultGP(operand.Reg, width, arch)
		x86ResultStore(operand, width, state, owned, 0, arch)
		state.canonicalGP[index] &^= mask
		state.callerAddressGP[index] = false
		state.callerAddressOrigin[index] = 0
		state.resultAddressGP[index] = x86ResultSlotAddress{}
		if state.boolFrameGP != nil {
			state.boolFrameGP[index] = x86ResultSlotAddress{}
		}
	}
	if spec.kind == amd64ExchangeAdd {
		state.flags = x86ByteFlagsAll
	}
	return true, true // XCHG preserves arithmetic FLAGS; both preserve DF.
}
