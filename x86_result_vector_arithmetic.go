package plan9asm

// Numerical result effects consume the same forms as the lowerer. Legacy
// arithmetic keeps its existing typed consumer; no caller or native-memory
// permission follows from knowing the bytes of a vector arithmetic result.
func x86ResultVectorArithmeticStep(ins Instr, state *x86ResultState, owned []bool, arch string) (bool, bool) {
	add, known, err := x86PackedIntegerAddProperties(ins, arch)
	form, bits := add.x86PackedBinaryForm, add.spec.laneBits
	if !known {
		subtract, subKnown, subErr := x86PackedSubtractProperties(ins, arch)
		if !subKnown {
			return false, false
		}
		form, bits, err = subtract.x86PackedBinaryForm, subtract.laneBits, subErr
	}
	if err != nil {
		return false, true
	}
	if form.legacy {
		return false, false
	}
	return x86ResultPackedLaneArithmeticStep(ins, form, bits, state, owned, arch), true
}
