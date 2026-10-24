package plan9asm

// Only an exact private read certificate supplies the SB bytes. The local
// proof learns numerical definedness, never a constant, pointer or address.
func fixedLocalNumericalReadStep(ins Instr, state *x86FixedLocalState) (bool, error) {
	if ins.x86NumericalRead == nil {
		return false, nil
	}
	if _, form, valid := numericalCertifiedRead(ins, "amd64"); valid && form.spec.vector != x86NumericalSBScalar {
		if !numericalVectorMapStep(ins, "amd64", form, state.vectors, state.masks) {
			return true, goAMD64FrameContext("fixed local numerical vector read needs its complete typed inputs")
		}
		return true, nil
	}
	effect, valid := x86NumericalReadEffect(ins, "amd64")
	if !valid {
		return true, goAMD64FrameContext("fixed local numerical read certificate changed")
	}
	for _, operand := range effect.operands {
		if operand.read && !fixedLocalNumericRead(operand.operand, operand.width, state) {
			return true, goAMD64FrameContext("fixed local numerical compare needs initialized numeric GP bytes")
		}
		if !operand.write {
			continue
		}
		root, valid := x86CallerGP(operand.operand.Reg)
		mask := x86CallerByteMask(operand.operand.Reg, operand.width)
		_, pointer := state.pointers[root]
		if !valid || root == SP || root == BP || x86SourceGRegister(root) || mask == 0 || pointer && operand.width != 8 {
			return true, goAMD64FrameContext("fixed local numerical read cannot replace a private address or stack role")
		}
		if operand.width == 4 {
			mask = 255 // Architectural dword writes zero the upper GP bytes.
		}
		delete(state.pointers, root)
		delete(state.opaque, root)
		delete(state.constants, root)
		fixedLocalWriteBooleans(state, operand.operand.Reg, operand.width, 0)
		state.known[root] |= mask
	}
	if effect.flagsDefined != 0 {
		state.flags = true
	}
	return true, nil
}
