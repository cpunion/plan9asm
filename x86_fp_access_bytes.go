package plan9asm

import "strings"

// X86FPAccessBytes observes a registered instruction's full native FP access
// width. It is not a source, initialization, alignment, caller or translation
// certificate. Unknown forms return no width; no opcode-suffix inference is
// used. The actual original assembler and selected declaration remain the
// diagnostic consumer's independent obligations.
func X86FPAccessBytes(ins Instr, index int, arch string) (width int64, write, known bool) {
	if arch != "amd64" && arch != "386" || index < 0 || index >= len(ins.Args) || ins.Args[index].Kind != OpFP {
		return 0, false, false
	}
	// Actual Parse/CPP output uses canonical source opcodes. The observation
	// does not promote a constructed case-folded identity to original source.
	if name := string(ins.Op); name != strings.ToUpper(name) || name != strings.TrimSpace(name) {
		return 0, false, false
	}
	// Generic source-stack diagnostics normalize suffixes, but that fallback
	// cannot certify a constructed scalar opcode which bypassed this grammar.
	if op := Op(normalizeInstructionOpcode(ins.Op)); x86RawReturnScalarMoveBytes(op) != 0 && ins.Op != op {
		return 0, false, false
	}
	if effect, recognized, err := x86TypedInstructionByteEffect(ins.Op, ins, arch); recognized {
		if err != nil {
			return 0, false, false
		}
		for _, operand := range effect.operands {
			if operand.index == index {
				return operand.width, operand.write, true
			}
		}
		return 0, false, false
	}
	// MOVL/MOVD and the VEX forms have complete Go GP/X memory rows in the
	// same typed transfer grammar as register copies. Keep the FP observer on
	// those rows instead of falling through to the GP-only scalar shortcut;
	// otherwise MOVL X0,value(FP) and its load direction are silently reported
	// as unknown despite being official formats. MOVQ stays out of this
	// observer: its scalar/vector alias is intentionally not an FP-width proof,
	// even though the caller grammar models its complete register/address rows.
	if ins.Op != "MOVQ" {
		if form, recognized, err := x86GPSIMDTransferProperties(ins, arch); recognized {
			if err != nil || form.memory < 0 || form.memory != index {
				return 0, false, false
			}
			return int64(form.bits / 8), form.memoryWrite, true
		}
	}
	// The remaining ymovl rows are the legacy MMX MOVD forms. They are
	// deliberately separate from the X-register transfer table because the
	// physical MMX bank has distinct state and is not a SIMD caller proof.
	if ins.Op == "MOVL" && len(ins.Args) == 2 {
		other := ins.Args[1-index]
		if other.Kind == OpReg && isAMD64MOVDMRegister(other) {
			return 4, index == 1, true
		}
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
	// The source-stack consumer shares these ordered typed forms. Its own
	// numerical/pointer/mask authority is not exported by this observation.
	return x86InstructionMemoryShape(ins, index, arch)
}
