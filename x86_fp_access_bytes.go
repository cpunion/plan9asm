package plan9asm

// X86FPAccessBytes observes a registered instruction's full native FP access
// width. It is not a source, initialization, alignment, caller or translation
// certificate. Unknown forms return no width; no opcode-suffix inference is
// used. The actual original assembler and selected declaration remain the
// diagnostic consumer's independent obligations.
func X86FPAccessBytes(ins Instr, index int, arch string) (width int64, write, known bool) {
	if arch != "amd64" && arch != "386" || index < 0 || index >= len(ins.Args) || ins.Args[index].Kind != OpFP {
		return 0, false, false
	}
	// Generic source-stack diagnostics normalize suffixes, but that fallback
	// cannot certify a constructed scalar opcode which bypassed this grammar.
	if op := Op(normalizeInstructionOpcode(ins.Op)); x86RawReturnScalarMoveBytes(op) != 0 && ins.Op != op {
		return 0, false, false
	}
	if ins.Op == "MOVB" || ins.Op == "MOVW" || ins.Op == "MOVL" || ins.Op == "MOVQ" {
		if len(ins.Args) != 2 {
			return 0, false, false
		}
		width = x86RawReturnScalarMoveBytes(ins.Op)
		otherOperand := ins.Args[1-index]
		if index == 1 && otherOperand.Kind == OpImm && otherOperand.ImmRaw == "" && !otherOperand.ImmIsFloat &&
			(width != 8 || arch == "amd64" && otherOperand.Imm == int64(int32(otherOperand.Imm))) {
			return width, true, true
		}
		if otherOperand.Kind != OpReg {
			return 0, false, false
		}
		other := ins.Args[1-index].Reg
		valid := isX86YrlRegisterForArch(other, arch)
		if width == 1 {
			valid = isGoYmbRegisterForArch(other, arch)
		}
		if !valid || width == 8 && arch == "386" {
			return 0, false, false
		}
		return width, index == 1, true
	}
	switch ins.Op {
	case "MOVOU", "MOVOA", "MOVO", "MOVUPS", "MOVAPS", "MOVUPD", "MOVAPD":
		if len(ins.Args) != 2 || ins.Args[1-index].Kind != OpReg {
			return 0, false, false
		}
		bank, valid := amd64ParseXReg(ins.Args[1-index].Reg)
		if !valid || bank >= 16 || arch == "386" && bank >= 8 {
			return 0, false, false
		}
		return 16, index == 1, true // no native alignment or initialized vector permission
	}
	// The source-stack consumer shares these ordered typed forms. Its own
	// numerical/pointer/mask authority is not exported by this observation.
	return x86InstructionMemoryShape(ins, index, arch)
}
