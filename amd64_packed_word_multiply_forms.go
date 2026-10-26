package plan9asm

import "strings"

type x86PackedWordMultiplySpec struct {
	mode     amd64PackedWordMultiplyMode
	encoding x86PackedLogicalEncoding
}

var x86PackedWordMultiplySpecs = map[Op]x86PackedWordMultiplySpec{
	"PMULLW":   {amd64PackedWordMultiplyLow, x86PackedLogicalLegacyInteger},
	"PMULHW":   {amd64PackedWordMultiplySignedHigh, x86PackedLogicalLegacyInteger},
	"PMULHUW":  {amd64PackedWordMultiplyUnsignedHigh, x86PackedLogicalLegacyInteger},
	"PMULHRSW": {amd64PackedWordMultiplyRoundedSignedHigh, x86PackedLogicalLegacyFP},
	"VPMULLW":  {amd64PackedWordMultiplyLow, x86PackedLogicalVEXOrEVEX},
	"VPMULHW":  {amd64PackedWordMultiplySignedHigh, x86PackedLogicalVEXOrEVEX},
	"VPMULHUW": {amd64PackedWordMultiplyUnsignedHigh, x86PackedLogicalVEXOrEVEX},
	"VPMULHRSW": {amd64PackedWordMultiplyRoundedSignedHigh,
		x86PackedLogicalVEXOrEVEX},
}

type x86PackedWordMultiplyForm struct {
	x86PackedBinaryForm
	spec   x86PackedWordMultiplySpec
	source Operand // Actual encoded bank, not a new initialized input.
}

// This shared grammar records the real native read/modify/write shape only.
// It grants neither input initialization nor native backing/caller authority.
func x86PackedWordMultiplyProperties(op Op, ins Instr, arch string) (x86PackedWordMultiplyForm, bool, error) {
	base, _, _ := strings.Cut(strings.ToUpper(string(op)), ".")
	spec, recognized := x86PackedWordMultiplySpecs[Op(base)]
	if !recognized {
		return x86PackedWordMultiplyForm{}, false, nil
	}
	grammar := x86PackedBinarySpec{encoding: spec.encoding, zeroing: spec.encoding == x86PackedLogicalVEXOrEVEX}
	binary, source, started, err := x86PackedBinaryPhysicalSource(op, ins, arch, grammar, "packed word multiply")
	if !started {
		return x86PackedWordMultiplyForm{}, true, err
	}
	return x86PackedWordMultiplyForm{binary, spec, source}, true, err
}

func x86ResultPackedWordMultiplyStep(ins Instr, state *x86ResultState, owned []bool, arch string) (bool, bool) {
	form, known, err := x86PackedWordMultiplyProperties(ins.Op, ins, arch)
	if !known || err != nil {
		return false, known
	}
	if !form.mmx && form.spec.mode == amd64PackedWordMultiplyLow {
		// W/D/Q low products share per-lane numerical dependencies. Missing
		// source bytes leave only their dependent result lanes unknown; they
		// never become initialized implementation-slot values. Whole-source
		// word input obligations remain in the separate scalar-FP consumer.
		return x86ResultPackedLaneArithmeticStep(ins, form.x86PackedBinaryForm, 16, state, owned, arch), true
	}
	return x86ResultFixedVectorInputs(ins, form.x86PackedBinaryForm, false, state, owned, arch), true
}
