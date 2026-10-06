package plan9asm

import "fmt"

// Go 1.27's four VPACK rows share the same three-operand grammar. The two
// dword-to-word operations also permit EVEX 32-bit memory broadcast.
var x86RawPackedNarrowOps = map[[2]int]struct {
	op        Op
	broadcast bool
}{
	{1, 0x6b}: {op: "VPACKSSDW", broadcast: true},
	{1, 0x63}: {op: "VPACKSSWB"},
	{2, 0x2b}: {op: "VPACKUSDW", broadcast: true},
	{1, 0x67}: {op: "VPACKUSWB"},
}

func decodedX86VectorPackedSaturatingNarrowInstruction(code []byte, mode int) (Instr, int, bool, error) {
	p, matched := decodeX86RawVectorEncoding(code)
	if !matched {
		return Instr{}, 0, false, nil
	}
	spec, known := x86RawPackedNarrowOps[[2]int{p.mapNumber, p.opcode}]
	if !known {
		return Instr{}, 0, false, nil
	}
	// Go emits W0 for these rows. Do not infer support for additional
	// physical WIG encodings from the shared named-operand grammar.
	if p.w {
		return Instr{}, 0, true, fmt.Errorf("packed saturating narrow: unsupported W bit")
	}
	bits := 16
	if spec.broadcast {
		bits = 32
	}
	return decodedX86BinaryVectorOperands(code, p, mode, spec.op, bits)
}
