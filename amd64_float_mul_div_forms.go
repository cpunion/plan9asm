package plan9asm

import (
	"fmt"
	"strings"
)

type x86FloatMulDivSpec struct {
	bits                   int
	legacy, scalar, divide bool
}

// Every legacy yxm and vector _yvaddpd/_yvaddsd MUL/DIV opcode. Sharing
// operand shapes does not share ADD/SUB's special-value or rounding rules.
var x86FloatMulDivSpecs = map[Op]x86FloatMulDivSpec{
	"MULPS": {32, true, false, false}, "MULPD": {64, true, false, false},
	"MULSS": {32, true, true, false}, "MULSD": {64, true, true, false},
	"DIVPS": {32, true, false, true}, "DIVPD": {64, true, false, true},
	"DIVSS": {32, true, true, true}, "DIVSD": {64, true, true, true},
	"VMULPS": {32, false, false, false}, "VMULPD": {64, false, false, false},
	"VMULSS": {32, false, true, false}, "VMULSD": {64, false, true, false},
	"VDIVPS": {32, false, false, true}, "VDIVPD": {64, false, false, true},
	"VDIVSS": {32, false, true, true}, "VDIVSD": {64, false, true, true},
}

type x86FloatMulDivForm struct {
	spec x86FloatMulDivSpec
	x86FloatBinaryForm
}

func x86FloatMulDivProperties(ins Instr, goarch string) (x86FloatMulDivForm, bool, error) {
	base, _, _ := strings.Cut(strings.ToUpper(string(ins.Op)), ".")
	spec, recognized := x86FloatMulDivSpecs[Op(base)]
	if !recognized {
		return x86FloatMulDivForm{}, false, nil
	}
	shape, err := x86FloatBinaryOperandProperties(ins, goarch, spec.bits, spec.legacy, spec.scalar, "mul/div")
	form := x86FloatMulDivForm{spec: spec, x86FloatBinaryForm: shape}
	if err != nil {
		return form, true, err
	}
	if goarch == "386" && !spec.legacy {
		for _, operand := range ins.Args {
			width := amd64VectorByteWidth(operand.Reg)
			if index, valid := amd64VectorRegisterIndex(operand.Reg, width); operand.Kind == OpReg && valid && index >= 8 {
				return form, true, fmt.Errorf("mode32 floating mul/div has no physical SIMD register %s: %w", operand.Reg, ErrProbeNeedsContext)
			}
		}
	}
	return form, true, nil
}
