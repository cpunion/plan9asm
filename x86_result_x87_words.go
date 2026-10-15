package plan9asm

import "strings"

// Only explicitly initialized control-word knownness is mutable state. The
// parameter offsets are immutable actual typed-layout metadata, shared by all
// CFG copies like pointerFP, not native backing or generic input-write rights.
type x86ResultX87WordState struct {
	control uint16
	layout  *x86X87WordLayout
}

type x86X87WordLayout struct {
	parameters map[int64]bool
}

func x86X87WordStateNeeded(fn Func) bool {
	for _, ins := range x86X87FlowInstructions(fn) {
		name, _, _ := strings.Cut(strings.ToUpper(string(ins.Op)), ".")
		if kind, known := x86X87WordSpecs[Op(name)]; known && kind != x86X87StoreStatus {
			return true // No target permission follows from this allocation predicate.
		}
	}
	return false
}

func x86ResultX87WordStep(ins Instr, state *x86ResultState, owned []bool, resultOffset int64, arch string) (bool, bool) {
	form, known, err := x86X87WordProperties(ins, arch)
	if !known {
		return false, false
	}
	if err != nil || !x86ResultNamedPredicateInstruction(ins) || state.x87 == nil || state.x87.invalid {
		return false, true
	}
	x87 := state.x87
	switch form.kind {
	case x86X87Initialize:
		if x87.word == nil {
			return false, true
		}
		if !state.shapeOnly {
			// Empty tags are not zero-valued F0..F7. Existing GP captures
			// remain physically unchanged; no GP or integer FLAGS are written.
			x87.slots, x87.depth = 0, 0
			x87.status, x87.word.control = 0xffff, 0xffff
		}
		return true, true
	case x86X87Clear:
		if !state.shapeOnly {
			// Only B/ES/SF and the exception bits become known zero.
			// C3/TOP/C2/C1/C0 retain precisely their previous knownness.
			x87.status = x87.status&0x7f00 | 0x80ff
		}
		return true, true
	case x86X87LoadControl:
		return false, true // NUM m16 alone closes no control/rounding/trapping role.
	}
	bits := x87.status
	if form.kind == x86X87StoreControl {
		bits = 0
		if x87.word != nil {
			bits = x87.word.control
		}
	}
	if form.operand.Kind == OpReg {
		index, mask, valid := x86ResultGP(form.operand.Reg, 2, arch)
		if !valid {
			return false, true
		}
		state.gp[index] &^= mask
		if bits&0xff == 0xff {
			state.gp[index] |= 1
		}
		if bits&0xff00 == 0xff00 {
			state.gp[index] |= 2
		}
		state.canonicalGP[index] &^= mask // Defined bytes are not canonical values.
		state.pointerGP[index], state.callerAddressGP[index] = false, false
		state.callerAddressOrigin[index] = 0
		state.resultAddressGP[index] = x86ResultSlotAddress{}
		if state.boolFrameGP != nil {
			state.boolFrameGP[index] = x86ResultSlotAddress{}
		}
		x87.gp[index] = bits
	} else {
		if !state.shapeOnly && bits != 0xffff {
			return false, true
		}
		parameter := form.operand.Kind == OpFP && form.operand.FPOffset < resultOffset
		if parameter {
			offset := form.operand.FPOffset
			if x87.word == nil || x87.word.layout == nil || !x87.word.layout.parameters[offset] || offset < 0 ||
				offset > int64(len(owned))-2 || !owned[offset] || !owned[offset+1] {
				return false, true
			}
			// Exactly the pre-existing natural I16 input shadow, not an
			// arbitrary span of owned scalar, padding, pointer or bool bytes.
			state.fp[offset], state.fp[offset+1] = true, true
		} else if !x86ResultStore(form.operand, 2, state, owned, resultOffset, arch) {
			return false, true
		}
	}
	// Preserve the existing conservative undefined post-condition-code policy.
	// A second store cannot borrow stale full-status facts from its first read.
	x87.status = 0
	return true, true
}
