package plan9asm

import "fmt"

// These bases are the predicated integer-unary encoder rows in Go 1.27's
// ARM64 inst_gen.go, with size and register fields cleared. Bit 20 selects
// merge versus zero: both modes share the typed lowering grammar.
var arm64RawSVEIntegerUnaryMergeBases = map[uint32]Op{
	0x0416a000: "ZABS",
	0x0418a000: "ZCLS",
	0x0419a000: "ZCLZ",
	0x041ba000: "ZCNOT",
	0x041aa000: "ZCNT",
	0x0417a000: "ZNEG",
	0x041ea000: "ZNOT",
	0x0410a000: "ZSXTB",
	0x0412a000: "ZSXTH",
	0x0414a000: "ZSXTW",
	0x0411a000: "ZUXTB",
	0x0413a000: "ZUXTH",
	0x0415a000: "ZUXTW",
}

func decodeARM64RawSVEIntegerUnary(word uint32) (Instr, bool) {
	const variableFields = uint32(0x00c01fff) // size, Pg, Zn, Zd
	base := word &^ variableFields
	mode := "M"
	op, ok := arm64RawSVEIntegerUnaryMergeBases[base]
	if !ok {
		base ^= 1 << 20
		op, ok = arm64RawSVEIntegerUnaryMergeBases[base]
		if !ok {
			return Instr{}, false
		}
		mode = "Z"
	}

	bits := 8 << ((word >> 22) & 3)
	spec := arm64SVEIntegerUnarySpecs[op]
	if bits < spec.minBits || (spec.maxBits != 0 && bits > spec.maxBits) {
		return Instr{}, false
	}
	width := map[int]string{8: "B", 16: "H", 32: "S", 64: "D"}[bits]
	source := int(word>>5) & 31
	predicate := int(word>>10) & 7
	destination := int(word) & 31
	return Instr{
		Op: op,
		Args: []Operand{
			{Kind: OpReg, Reg: Reg(fmt.Sprintf("Z%d.%s", source, width))},
			{Kind: OpReg, Reg: Reg(fmt.Sprintf("P%d.%s", predicate, mode))},
			{Kind: OpReg, Reg: Reg(fmt.Sprintf("Z%d.%s", destination, width))},
		},
		Raw: fmt.Sprintf("WORD $%#08x", word),
	}, true
}
