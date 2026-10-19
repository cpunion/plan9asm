package plan9asm

import (
	"fmt"
	"strings"
)

type x86FloatMinMaxSpec struct {
	bits                    int
	legacy, scalar, minimum bool
}

// Go's eight yxm opcodes and eight _yvaddpd/_yvaddsd members share operand
// axes, not arithmetic's ER attributes. MIN/MAX enable SAE only.
var x86FloatMinMaxSpecs = map[Op]x86FloatMinMaxSpec{
	"MINPS": {32, true, false, true}, "MINPD": {64, true, false, true},
	"MINSS": {32, true, true, true}, "MINSD": {64, true, true, true},
	"MAXPS": {32, true, false, false}, "MAXPD": {64, true, false, false},
	"MAXSS": {32, true, true, false}, "MAXSD": {64, true, true, false},
	"VMINPS": {32, false, false, true}, "VMINPD": {64, false, false, true},
	"VMINSS": {32, false, true, true}, "VMINSD": {64, false, true, true},
	"VMAXPS": {32, false, false, false}, "VMAXPD": {64, false, false, false},
	"VMAXSS": {32, false, true, false}, "VMAXSD": {64, false, true, false},
}

type x86FloatMinMaxForm struct {
	x86FloatBinaryForm
	spec x86FloatMinMaxSpec
	sae  bool
}

func x86FloatMinMaxNativeEVEX(ins Instr, f x86FloatMinMaxForm) bool {
	if f.legacy {
		return false
	}
	if ins.x86FloatMinMaxEVEX || f.byteWidth == 64 || f.mask >= 0 || f.broadcast || f.sae {
		return true
	}
	for _, operand := range ins.Args {
		if index, valid := amd64VectorRegisterIndex(operand.Reg, f.byteWidth); operand.Kind == OpReg && valid && index >= 16 {
			return true
		}
	}
	return false
}

func x86FloatMinMaxProperties(ins Instr, arch string) (x86FloatMinMaxForm, bool, error) {
	base, suffix, dotted := strings.Cut(strings.ToUpper(string(ins.Op)), ".")
	spec, known := x86FloatMinMaxSpecs[Op(base)]
	if !known {
		return x86FloatMinMaxForm{}, false, nil
	}
	controls, valid := parseAMD64BinaryFloatingSuffix(suffix)
	f := x86FloatMinMaxForm{spec: spec, sae: controls.sae}
	fail := func() (x86FloatMinMaxForm, bool, error) {
		return f, true, fmt.Errorf("%s %s has invalid MIN/MAX operands or controls: %q", arch, ins.Op, ins.Raw)
	}
	if !valid || dotted && suffix == "" || controls.rounding != "" || spec.legacy && dotted {
		return fail()
	}
	// Remove only the separately checked SAE control from a private shape.
	// Original opcode/operands stay unchanged for native emission and identity.
	shape := ins
	if controls.sae {
		shape.Op = Op(base)
		if controls.zeroing {
			shape.Op += ".Z"
		}
	}
	var err error
	// Go's float-constant rewrite has no MIN/MAX entries in obj6.go.
	f.x86FloatBinaryForm, err = x86FloatBinaryOperandProperties(shape, arch, spec.bits, spec.legacy, spec.scalar, false, "min/max")
	if err != nil {
		return f, true, err
	}
	if controls.sae && (ins.Args[0].Kind != OpReg || !spec.scalar && f.byteWidth != 64) {
		return fail()
	}
	return f, true, x86Mode32VectorRegisterContext(ins, arch)
}
