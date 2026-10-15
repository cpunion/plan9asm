package plan9asm

import "strings"

// This is source-authored logical availability, not hardware Tag/TOP, a saved
// environment, physical stack capacity or control/rounding authority. Entry
// values/status start unknown regardless of the emitter's private zero slots.
type x86ResultX87State struct {
	slots   uint8
	depth   uint8
	invalid bool
	status  uint16
	gp      [16]uint16 // Captured known status bits, never whole NUM/canonical bytes.
}

const x86ResultX87StatusConditions uint16 = 0x4500 // C3, C2, C0 only.

func x86X87ValueFlowNeeded(fn Func) bool {
	instructions := fn.Instrs
	if fn.x86DirectTailRoot != nil {
		instructions = fn.x86DirectTailRoot.instructions
	} else if fn.x86ContinuationFlow != nil {
		instructions = fn.x86ContinuationFlow.instructions
	}
	for _, ins := range instructions {
		name, _, _ := strings.Cut(strings.ToUpper(string(ins.Op)), ".")
		if _, known := x86X87ValueSpecs[Op(name)]; known {
			return true
		}
	}
	return false
}

func (state *x86ResultX87State) read(operand Operand) bool {
	index, valid := x87OperandReg(operand)
	return valid && !state.invalid && index < int(state.depth) && state.slots&(1<<uint(index)) != 0
}

func (state *x86ResultX87State) write(operand Operand) bool {
	index, valid := x87OperandReg(operand)
	if !valid || state.invalid {
		return false
	}
	state.slots |= 1 << uint(index)
	if state.depth <= uint8(index) {
		state.depth = uint8(index + 1)
	}
	return true
}

func (state *x86ResultX87State) push() bool {
	if state.invalid || state.depth >= 8 {
		return false
	}
	state.slots = state.slots<<1 | 1
	state.depth++
	return true
}

func (state *x86ResultX87State) pop(count uint8) bool {
	if state.invalid || state.depth < count {
		return false
	}
	for i := uint8(0); i < count; i++ {
		if state.slots&1 == 0 {
			return false
		}
		state.slots >>= 1
		state.depth--
	}
	return true
}

// Only the private logical-FP observer installs this transfer. Existing
// bounded FP/AUTO or exact observed static bytes supply memory; no conditional
// native read or write follows from knowing an x87 instruction's width.
func x86ResultX87ValueStep(ins Instr, state *x86ResultState, owned []bool, resultOffset int64, arch string) (bool, bool) {
	form, known, err := x86X87ValueProperties(ins, arch)
	if !known {
		return false, false
	}
	if err != nil || form.unprovedEncoding || !x86ResultNamedPredicateInstruction(ins) || state.x87 == nil || state.x87.invalid {
		return false, true
	}
	x87 := state.x87
	read := func(operand Operand) bool {
		if isX87RegisterOperand(operand) {
			return state.shapeOnly || x87.read(operand)
		}
		if operand.Kind == OpImm {
			return true // The descriptor already validated the exact Go FLOAT rewrite.
		}
		return x86ResultRead(operand, form.memoryBytes(), state, owned, arch)
	}
	write := func(operand Operand) bool {
		if isX87RegisterOperand(operand) {
			return state.shapeOnly || x87.write(operand)
		}
		return x86ResultStore(operand, form.memoryBytes(), state, owned, resultOffset, arch)
	}
	f0 := Operand{Kind: OpReg, Reg: "F0"}
	status := uint16(0)
	switch form.spec.kind {
	case x86X87FloatMove:
		if !read(form.source) {
			return false, true
		}
		if form.push {
			if !state.shapeOnly && !x87.push() {
				return false, true
			}
		} else if !write(form.destination) {
			return false, true
		}
	case x86X87Constant:
		if !state.shapeOnly && !x87.push() {
			return false, true
		}
	case x86X87Unary:
		if !read(f0) {
			return false, true
		}
	case x86X87Exchange:
		if !read(form.source) || !read(form.destination) {
			return false, true
		}
		// Both exact cells were independently defined before either write.
	case x86X87Arithmetic:
		if !read(form.source) || !read(form.destination) || !write(form.destination) {
			return false, true
		}
	case x86X87Compare:
		if !read(form.left) || !read(form.right) {
			return false, true
		}
		if form.spec.directFlags {
			state.flags = x86ByteFlagsAll // CF/PF/ZF computed; OF/SF/AF cleared.
		} else {
			status = x86ResultX87StatusConditions
		}
	case x86X87Test:
		if !read(f0) {
			return false, true
		}
		status = x86ResultX87StatusConditions
	case x86X87Status:
		if form.destination.Kind == OpReg {
			index, mask, valid := x86ResultGP(form.destination.Reg, 2, arch)
			if !valid {
				return false, true
			}
			state.gp[index] &^= mask
			state.canonicalGP[index] &^= mask
			state.pointerGP[index], state.callerAddressGP[index] = false, false
			state.callerAddressOrigin[index] = 0
			state.resultAddressGP[index] = x86ResultSlotAddress{}
			if state.boolFrameGP != nil {
				state.boolFrameGP[index] = x86ResultSlotAddress{}
			}
			x87.gp[index] = x87.status
		} else if !state.shapeOnly && x87.status != 0xffff ||
			!x86ResultStore(form.destination, 2, state, owned, resultOffset, arch) {
			return false, true // A partially known status cannot become stored NUM bytes.
		}
	}
	if !state.shapeOnly && !x87.pop(form.spec.pop) {
		return false, true
	}
	// Never inherit stale C bits from a previous operation or private alloca.
	// FSTSW captures the previous status before its documented undefined
	// post-state; integer FLAGS and DF remain independent.
	x87.status = status
	return true, true
}

// A saved status may expose only independently defined bits. The complete
// existing AND B/W/L/Q grammar provides its physical aliases, encoded mask and
// FLAGS effects; unknown status bits selected by that mask remain Context.
func x86ResultX87StatusProjection(ins Instr, state *x86ResultState, owned []bool, resultOffset int64, arch string) (bool, bool) {
	if state.x87 == nil || state.x87.invalid || state.shapeOnly || !x86ResultNamedPredicateInstruction(ins) {
		return false, false
	}
	form, known, err := x86LockScalarRMWProperties(ins.Op, ins, arch)
	if !known || err != nil || form.spec.stem != "AND" || form.source.Kind != OpImm || form.destination.Kind != OpReg {
		return false, false
	}
	width := int64(form.spec.bits / 8)
	index, bytes, valid := x86ResultGP(form.destination.Reg, width, arch)
	if !valid || state.x87.gp[index] == 0 {
		return false, false
	}
	var initialized uint64
	for b := uint(0); b < 8; b++ {
		if state.gp[index]&(1<<b) != 0 {
			initialized |= uint64(0xff) << (8 * b)
		}
	}
	initialized |= uint64(state.x87.gp[index])
	mask := uint64(form.source.Imm)
	if width != 8 {
		mask &= uint64(1)<<uint(8*width) - 1
	}
	if bytes == 2 && width == 1 {
		mask <<= 8 // AH aliases the same captured AX word, not a second GP.
	}
	if initialized&mask != mask || !x86ResultStore(form.destination, width, state, owned, resultOffset, arch) {
		return false, true
	}
	state.canonicalGP[index] &^= bytes
	state.pointerGP[index], state.callerAddressGP[index] = false, false
	state.callerAddressOrigin[index] = 0
	state.resultAddressGP[index] = x86ResultSlotAddress{}
	if state.boolFrameGP != nil {
		state.boolFrameGP[index] = x86ResultSlotAddress{}
	}
	state.x87.gp[index] &^= x86X87CapturedByteMask(bytes)
	state.flags = state.flags&^form.spec.flagsKilled | form.spec.flagsDefined
	return true, true
}

func x86X87CapturedByteMask(bytes uint8) uint16 {
	var mask uint16
	if bytes&1 != 0 {
		mask |= 0xff
	}
	if bytes&2 != 0 {
		mask |= 0xff00
	}
	return mask
}

// Called after an ordinary transfer. Overwriting GP bytes revokes captured
// status-bit knowledge even when numerical definedness remains unchanged.
func x86ResultX87After(ins Instr, state *x86ResultState, arch string) {
	if state.x87 == nil {
		return
	}
	if ins.Op == OpCALL {
		*state.x87 = x86ResultX87State{invalid: true}
		return // An independently accepted call supplies no empty-stack or captured GP fact.
	}
	if _, known, _ := x86X87ValueProperties(ins, arch); known {
		return // The typed value transfer already applied its exact write/capture.
	}
	x86ResultX87CapturedAfter(ins, state, arch)
}

func x86ResultX87CapturedAfter(ins Instr, state *x86ResultState, arch string) {
	if state.x87 == nil {
		return
	}
	clear := func(operand Operand, width int64) {
		if operand.Kind == OpReg {
			if index, bytes, valid := x86ResultGP(operand.Reg, width, arch); valid {
				state.x87.gp[index] &^= x86X87CapturedByteMask(bytes)
			}
		}
	}
	if effect, known, err := x86TypedInstructionByteEffect(ins.Op, ins, arch); known && err == nil {
		for _, operand := range effect.operands {
			if operand.write {
				clear(operand.operand, operand.width)
			}
		}
		return
	}
	if width := x86RawReturnScalarMoveBytes(ins.Op); width != 0 && len(ins.Args) == 2 {
		clear(ins.Args[1], width)
		return
	}
	if form, known, err := x86ConditionalSetProperties(ins.Op, ins, arch); known && err == nil {
		clear(form.destination, 1)
		return
	}
	var captures [16]bool
	for index, bits := range state.x87.gp {
		captures[index] = bits != 0
	}
	after := x86ResultCallerAddressesAfter(ins, captures, nil, arch)
	for index, live := range after {
		if !live {
			state.x87.gp[index] = 0
		}
	}
}
