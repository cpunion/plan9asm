package plan9asm

// Each bit records initialization of one bit in private numerical storage.
// This retains independently unknown status bits through PUSHF/POPF without
// turning an entire packed byte into a numerical or pointer witness.
type x86ResultValueStack struct {
	bytes   []uint8 // Low byte is at the end, matching a descending byte stack.
	invalid bool
	id      bool // ID is separate from the six arithmetic flags and DF.
}

func x86ResultValueStackLimit(arch string) int {
	// AMD64 lowering reserves one of its 64 virtual slots for an unproved
	// entry continuation. This observer never reads that slot. The 386 byte
	// stack uses the same bounded number of native words.
	return 63 * int(x86ResultPointerWord(arch))
}

func x86ResultValueStackContinuation(ins Instr, state *x86ResultState, arch string) bool {
	if state.valueStack == nil || state.shapeOnly || len(state.valueStack.bytes) == 0 && !state.valueStack.invalid {
		return true
	}
	if state.valueStack.invalid || ins.Op == OpCALL || !x86ResultNamedPredicateInstruction(ins) {
		return false
	}
	// On 386 the private byte stack moves the simulated SP. An AUTO access
	// cannot borrow the independent unchanged-SP local-byte proof while live.
	if arch == "386" {
		for _, arg := range ins.Args {
			if arg.Kind == OpMem && (arg.Mem.Base == SP || arg.Mem.Base == BP) {
				return false
			}
		}
	}
	return true
}

func x86ResultValueStackStep(ins Instr, state *x86ResultState, owned []bool, resultOffset int64, arch string) (bool, bool) {
	spec, known, err := x86StackWidthProperties(ins, arch)
	if !known {
		return false, false
	}
	if err != nil || !x86ResultNamedPredicateInstruction(ins) || state.valueStack == nil || state.valueStack.invalid {
		return false, true
	}
	width := int64(spec.bits / 8)
	storage := int(width)
	if arch == "amd64" {
		storage = 8 // Existing virtual AMD64 slots zero-extend narrow pushes.
	}
	stack := state.valueStack
	push := spec.mode == amd64StackPushValue || spec.mode == amd64StackPushFlags
	if push {
		bits := ^uint64(0)
		if spec.mode == amd64StackPushValue {
			arg := ins.Args[0]
			if x86SegmentRegisterOperand(arg) || !x86ResultRead(arg, width, state, owned, arch) {
				return false, true
			}
		} else {
			for index, physical := range []uint{0, 2, 4, 6, 7, 11} {
				if state.flags&(1<<uint(index)) == 0 {
					bits &^= 1 << physical
				}
			}
			if state.direction == x86ResultDirectionUnknown {
				bits &^= 1 << 10
			}
			if !stack.id && spec.bits > 16 {
				bits &^= 1 << 21
			}
		}
		if state.shapeOnly {
			return true, true
		}
		if len(stack.bytes) > x86ResultValueStackLimit(arch)-storage {
			return false, true
		}
		for index := storage - 1; index >= 0; index-- {
			stack.bytes = append(stack.bytes, uint8(bits>>uint(8*index)))
		}
		return true, true
	}
	bits := ^uint64(0)
	if !state.shapeOnly {
		if len(stack.bytes) < storage {
			return false, true
		}
		bits = 0
		for index := 0; index < storage; index++ {
			bits |= uint64(stack.bytes[len(stack.bytes)-1-index]) << uint(8*index)
		}
		stack.bytes = stack.bytes[:len(stack.bytes)-storage]
	}
	if spec.mode == amd64StackPopFlags {
		state.flags = 0
		for index, physical := range []uint{0, 2, 4, 6, 7, 11} {
			if bits&(1<<physical) != 0 {
				state.flags |= 1 << uint(index)
			}
		}
		state.direction = x86ResultDirectionUnknown
		if bits&(1<<10) != 0 {
			state.direction = x86ResultDirectionEither
		}
		if spec.bits > 16 {
			stack.id = bits&(1<<21) != 0
		}
		return true, true
	}
	arg := ins.Args[0]
	if x86SegmentRegisterOperand(arg) {
		return false, true
	}
	if arg.Kind == OpReg {
		index, mask, valid := x86ResultGP(arg.Reg, width, arch)
		if !valid {
			return false, true
		}
		var initialized uint8
		for b := int64(0); b < width; b++ {
			if uint8(bits>>uint(8*b)) == 0xff {
				initialized |= 1 << uint(b)
			}
		}
		state.gp[index] = state.gp[index]&^mask | initialized
		if width == 4 {
			state.gp[index] |= 0xf0
		}
		state.pointerGP[index], state.callerAddressGP[index] = false, false
		state.canonicalGP[index] &^= mask
		return true, true
	}
	for b := int64(0); b < width; b++ {
		if uint8(bits>>uint(8*b)) != 0xff {
			return false, true
		}
	}
	return x86ResultStore(arg, width, state, owned, resultOffset, arch), true
}
