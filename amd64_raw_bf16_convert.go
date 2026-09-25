package plan9asm

import (
	"fmt"
	"strings"
)

// VCVTNEPS2BF16 is emitted as raw bytes by external GoAT modules because
// Go 1.27 has no named encoder row for AVX-512 BF16. LLVM 22 defines three
// input widths with a narrowing destination, plus masks and memory broadcast.
func decodedX86RawBF16ConvertInstruction(code []byte, mode int) (Instr, int, bool, error) {
	p, ok := decodeX86RawVectorEncoding(code)
	if !ok || !p.evex || p.mapNumber != 2 || p.pp != 2 || p.opcode != 0x72 {
		return Instr{}, 0, false, nil
	}
	fail := func(message string) (Instr, int, bool, error) {
		return Instr{}, 0, true, fmt.Errorf("VCVTNEPS2BF16: %s", message)
	}
	if mode != 32 && mode != 64 {
		return fail("unsupported x86 mode")
	}
	if p.addressOverride {
		return fail("address-size override is not source-layout safe")
	}
	if !p.fixed || p.w || p.upper != 0 || p.vectorLength > 2 {
		return fail("reserved EVEX width, source or vector-length field")
	}
	if p.zero && p.mask == 0 {
		return fail("zeroing requires a nonzero K mask")
	}
	if len(code) <= p.modRM {
		return fail("missing ModRM byte")
	}
	registerSource := code[p.modRM]>>6 == 3
	if p.broadcast && registerSource {
		return fail("broadcast requires a memory source")
	}
	if mode == 32 && (!registerSource && (p.b != 0 || p.x != 0) || p.r != 0) {
		return fail("extended address or destination register in 32-bit mode")
	}

	sourceBytes := 16 << p.vectorLength
	sourcePrefix := [...]string{"X", "Y", "Z"}[p.vectorLength]
	destinationPrefix := [...]string{"X", "X", "Y"}[p.vectorLength]
	accessBytes := sourceBytes
	if p.broadcast {
		accessBytes = 4
	}
	source, consumed, err := decodedX86EVEXRMOperand(code[p.modRM:], mode, p.b, p.x, p.segment, sourcePrefix, accessBytes)
	if err != nil {
		return Instr{}, 0, true, err
	}
	if mode == 32 && registerSource {
		index, _ := amd64VectorRegisterIndex(source.Reg, sourceBytes)
		if index > 7 {
			return fail("extended source register in 32-bit mode")
		}
	}
	destination := int(code[p.modRM]>>3&7) + p.r*8
	op := [...]Op{"VCVTNEPS2BF16X", "VCVTNEPS2BF16Y", "VCVTNEPS2BF16"}[p.vectorLength]
	if p.broadcast {
		op += ".BCST"
	}
	if p.zero {
		op += ".Z"
	}
	args := []Operand{source}
	if p.mask != 0 {
		args = append(args, Operand{Kind: OpReg, Reg: Reg(fmt.Sprintf("K%d", p.mask))})
	}
	args = append(args, Operand{Kind: OpReg, Reg: Reg(fmt.Sprintf("%s%d", destinationPrefix, destination))})
	parts := make([]string, len(args))
	for i, arg := range args {
		parts[i] = arg.String()
	}
	return Instr{
		Op:         op,
		Args:       args,
		Raw:        fmt.Sprintf("%s %s", op, strings.Join(parts, ", ")),
		x86Encoded: true,
	}, p.modRM + consumed, true, nil
}
