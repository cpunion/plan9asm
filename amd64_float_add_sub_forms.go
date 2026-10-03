package plan9asm

import "strings"

type x86FloatAddSubSpec struct {
	bits                     int
	legacy, scalar, subtract bool
}

// Go's eight yxm rows and eight _yvaddpd/_yvaddsd opcodes. Horizontal,
// alternating and raw-only FP16 arithmetic have separate contracts.
var x86FloatAddSubSpecs = map[Op]x86FloatAddSubSpec{
	"ADDPS": {32, true, false, false}, "ADDPD": {64, true, false, false},
	"ADDSS": {32, true, true, false}, "ADDSD": {64, true, true, false},
	"SUBPS": {32, true, false, true}, "SUBPD": {64, true, false, true},
	"SUBSS": {32, true, true, true}, "SUBSD": {64, true, true, true},
	"VADDPS": {32, false, false, false}, "VADDPD": {64, false, false, false},
	"VADDSS": {32, false, true, false}, "VADDSD": {64, false, true, false},
	"VSUBPS": {32, false, false, true}, "VSUBPD": {64, false, false, true},
	"VSUBSS": {32, false, true, true}, "VSUBSD": {64, false, true, true},
}

type x86FloatAddSubForm struct {
	spec x86FloatAddSubSpec
	x86FloatBinaryForm
}

func x86FloatAddSubProperties(ins Instr, goarch string) (x86FloatAddSubForm, bool, error) {
	base, _, _ := strings.Cut(strings.ToUpper(string(ins.Op)), ".")
	spec, recognized := x86FloatAddSubSpecs[Op(base)]
	if !recognized {
		return x86FloatAddSubForm{}, false, nil
	}
	form, err := x86FloatBinaryOperandProperties(ins, goarch, spec.bits, spec.legacy, spec.scalar, "add/sub")
	return x86FloatAddSubForm{spec: spec, x86FloatBinaryForm: form}, true, err
}
