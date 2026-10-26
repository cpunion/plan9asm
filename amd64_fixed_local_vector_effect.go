package plan9asm

// Only descriptors with explicit vector storage/alias metadata participate.
// MMX, unregistered upper writes and wider banks keep their separate owners.
func fixedLocalVectorByteEffect(ins Instr) (x86TypedByteEffect, bool, error) {
	if ins.x86NumericalRead != nil {
		return x86TypedByteEffect{}, false, nil // Preserve the existing owner-specific consumer.
	}
	effect, known, err := x86PackedBinaryByteEffect(ins.Op, ins, "amd64")
	if !known {
		return effect, false, nil
	}
	if err != nil {
		return x86TypedByteEffect{}, true, goAMD64FrameContext("fixed local vector form is not closed: %v", err)
	}
	valid := effect.vector.upperWrite == x86TypedVectorUpperPreserve && effect.vector.memoryAlignment == 16
	for _, operand := range effect.operands {
		valid = valid && operand.width == 16
		switch operand.operand.Kind {
		case OpReg:
			index, xmm := amd64VectorRegisterIndex(operand.operand.Reg, 16)
			valid = valid && xmm && index < 16
		case OpMem, OpSym:
		default:
			valid = false
		}
	}
	if !valid {
		return x86TypedByteEffect{}, true, goAMD64FrameContext("fixed local vector effect has an unbound storage or alias role: %s", ins.Op)
	}
	return effect, true, nil
}
