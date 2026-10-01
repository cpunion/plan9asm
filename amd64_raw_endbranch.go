package plan9asm

// ENDBR32/ENDBR64 have fixed, operand-free encodings (Intel SDM, volume 2).
// x86asm does not decode them yet. Preserve the actual instruction rather than
// replacing the CET state transition with a NOP. Go names only ENDBR64, so
// x86Encoded also keeps ENDBR32 confined to validated raw input.
func decodedX86EndBranchInstruction(code []byte) (Instr, int, bool) {
	if len(code) < 4 || code[0] != 0xf3 || code[1] != 0x0f || code[2] != 0x1e {
		return Instr{}, 0, false
	}
	var op Op
	switch code[3] {
	case 0xfa:
		op = "ENDBR64"
	case 0xfb:
		op = "ENDBR32"
	default:
		return Instr{}, 0, false
	}
	return Instr{Op: op, Raw: string(op), x86Encoded: true}, 4, true
}
