package plan9asm

// Sqrt computes only from its first operand. Scalar VEX/EVEX copies the
// second operand's upper X bytes without requiring its unused low scalar.
// Unknown copied bytes stay unknown; no pointer or storage facts are created.
func x86ResultSqrtStep(ins Instr, state *x86ResultState, owned []bool, arch string, nativeRead bool) (bool, bool) {
	f, known, err := x86SqrtProperties(ins, arch)
	if !known || err != nil {
		return false, known
	}
	if f.mask >= 0 {
		mask, _ := amd64ParseKReg(ins.Args[f.mask].Reg)
		if !state.masks[mask] {
			return false, true
		}
	}
	width := f.byteWidth
	if f.scalar {
		width = f.bits / 8
	}
	required := x86CallerVectorByteMask(int64(width))
	if f.memory >= 0 {
		if !f.scalar && ins.Args[0].Kind == OpFP {
			return false, true // Packed original memory needs contiguous native backing.
		}
		if !x86ResultReadWithNativeContinuation(ins, ins.Args[0], int64(f.memoryBytes), state, owned, arch, nativeRead) {
			return false, true
		}
	} else {
		source, _ := amd64VectorRegisterIndex(ins.Args[0].Reg, f.byteWidth)
		if state.vectorBytes(source)&required != required {
			return false, true
		}
	}
	destination, _ := amd64VectorRegisterIndex(ins.Args[f.destination].Reg, f.byteWidth)
	old := state.vectorBytes(destination)
	if f.mask >= 0 && !f.zeroing && old&required != required {
		return false, true
	}
	value := required
	if f.legacy && f.scalar {
		value |= old
	} else if f.scalar {
		upper, _ := amd64VectorRegisterIndex(ins.Args[f.second].Reg, 16)
		value |= state.vectorBytes(upper) & (0xffff &^ required)
	}
	state.writeVector(destination, f.byteWidth, value, f.legacy)
	return true, true
}
