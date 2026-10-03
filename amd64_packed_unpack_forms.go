package plan9asm

import (
	"fmt"
	"strings"
)

type x86PackedUnpackSpec struct {
	laneBits int
	high     bool
	grammar  x86PackedBinarySpec
}

// Go's six ymm rows include MMX. Its other six legacy rows use yxm.
// All twelve vector rows use _yvandnpd, but only D/Q/PS/PD permit BCST.
// Their EVEX memory exception class is E4NF: masking does not suppress faults.
var x86PackedUnpackSpecs = map[Op]x86PackedUnpackSpec{
	"PUNPCKLBW":   {8, false, x86PackedBinarySpec{x86PackedLogicalLegacyInteger, 0, false}},
	"PUNPCKHBW":   {8, true, x86PackedBinarySpec{x86PackedLogicalLegacyInteger, 0, false}},
	"PUNPCKLWL":   {16, false, x86PackedBinarySpec{x86PackedLogicalLegacyInteger, 0, false}},
	"PUNPCKHWL":   {16, true, x86PackedBinarySpec{x86PackedLogicalLegacyInteger, 0, false}},
	"PUNPCKLLQ":   {32, false, x86PackedBinarySpec{x86PackedLogicalLegacyInteger, 0, false}},
	"PUNPCKHLQ":   {32, true, x86PackedBinarySpec{x86PackedLogicalLegacyInteger, 0, false}},
	"PUNPCKLQDQ":  {64, false, x86PackedBinarySpec{x86PackedLogicalLegacyFP, 0, false}},
	"PUNPCKHQDQ":  {64, true, x86PackedBinarySpec{x86PackedLogicalLegacyFP, 0, false}},
	"UNPCKLPS":    {32, false, x86PackedBinarySpec{x86PackedLogicalLegacyFP, 0, false}},
	"UNPCKHPS":    {32, true, x86PackedBinarySpec{x86PackedLogicalLegacyFP, 0, false}},
	"UNPCKLPD":    {64, false, x86PackedBinarySpec{x86PackedLogicalLegacyFP, 0, false}},
	"UNPCKHPD":    {64, true, x86PackedBinarySpec{x86PackedLogicalLegacyFP, 0, false}},
	"VPUNPCKLBW":  {8, false, x86PackedBinarySpec{x86PackedLogicalVEXOrEVEX, 0, true}},
	"VPUNPCKHBW":  {8, true, x86PackedBinarySpec{x86PackedLogicalVEXOrEVEX, 0, true}},
	"VPUNPCKLWD":  {16, false, x86PackedBinarySpec{x86PackedLogicalVEXOrEVEX, 0, true}},
	"VPUNPCKHWD":  {16, true, x86PackedBinarySpec{x86PackedLogicalVEXOrEVEX, 0, true}},
	"VPUNPCKLDQ":  {32, false, x86PackedBinarySpec{x86PackedLogicalVEXOrEVEX, 4, true}},
	"VPUNPCKHDQ":  {32, true, x86PackedBinarySpec{x86PackedLogicalVEXOrEVEX, 4, true}},
	"VPUNPCKLQDQ": {64, false, x86PackedBinarySpec{x86PackedLogicalVEXOrEVEX, 8, true}},
	"VPUNPCKHQDQ": {64, true, x86PackedBinarySpec{x86PackedLogicalVEXOrEVEX, 8, true}},
	"VUNPCKLPS":   {32, false, x86PackedBinarySpec{x86PackedLogicalVEXOrEVEX, 4, true}},
	"VUNPCKHPS":   {32, true, x86PackedBinarySpec{x86PackedLogicalVEXOrEVEX, 4, true}},
	"VUNPCKLPD":   {64, false, x86PackedBinarySpec{x86PackedLogicalVEXOrEVEX, 8, true}},
	"VUNPCKHPD":   {64, true, x86PackedBinarySpec{x86PackedLogicalVEXOrEVEX, 8, true}},
}

type x86PackedUnpackForm struct {
	x86PackedBinaryForm
	spec x86PackedUnpackSpec
}

func x86PackedUnpackProperties(ins Instr, goarch string) (x86PackedUnpackForm, bool, error) {
	base, _, _ := strings.Cut(strings.ToUpper(string(ins.Op)), ".")
	spec, recognized := x86PackedUnpackSpecs[Op(base)]
	if !recognized {
		return x86PackedUnpackForm{}, false, nil
	}
	form, err := x86PackedBinaryProperties(ins, goarch, spec.grammar)
	// Intel's three MMX low-unpack rows are mm/m32, unlike the high
	// rows' mm/m64. The unused high source half is not a memory access.
	if err == nil && form.mmx && form.memory >= 0 && !spec.high {
		form.memoryBytes = 4
	}
	if err == nil && goarch == "386" && !form.legacy {
		for _, operand := range ins.Args {
			index, valid := amd64VectorRegisterIndex(operand.Reg, form.byteWidth)
			if operand.Kind == OpReg && valid && index >= 8 {
				return x86PackedUnpackForm{x86PackedBinaryForm: form, spec: spec}, true,
					fmt.Errorf("%w: 386 unpack register %s has no proved physical mode32 binding", ErrProbeNeedsContext, operand.Reg)
			}
		}
	}
	return x86PackedUnpackForm{x86PackedBinaryForm: form, spec: spec}, true, err
}
