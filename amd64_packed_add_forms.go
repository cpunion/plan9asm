package plan9asm

import "strings"

type x86PackedAddSpec struct {
	x86PackedBinarySpec
	laneBits int
}

// The eight wrapping Go spellings, plus cmd/asm's explicit PADDD -> APADDL
// alias. Saturating adds have a distinct arithmetic contract and do not gain
// this family's private caller permission.
var x86PackedWrappingAddSpecs = map[Op]x86PackedAddSpec{
	"PADDB":  {x86PackedBinarySpec{encoding: x86PackedLogicalLegacyInteger}, 8},
	"PADDW":  {x86PackedBinarySpec{encoding: x86PackedLogicalLegacyInteger}, 16},
	"PADDL":  {x86PackedBinarySpec{encoding: x86PackedLogicalLegacyInteger}, 32},
	"PADDD":  {x86PackedBinarySpec{encoding: x86PackedLogicalLegacyInteger}, 32},
	"PADDQ":  {x86PackedBinarySpec{encoding: x86PackedLogicalLegacyFP}, 64},
	"VPADDB": {x86PackedBinarySpec{encoding: x86PackedLogicalVEXOrEVEX, zeroing: true}, 8},
	"VPADDW": {x86PackedBinarySpec{encoding: x86PackedLogicalVEXOrEVEX, zeroing: true}, 16},
	"VPADDD": {x86PackedBinarySpec{encoding: x86PackedLogicalVEXOrEVEX, broadcastBytes: 4, zeroing: true}, 32},
	"VPADDQ": {x86PackedBinarySpec{encoding: x86PackedLogicalVEXOrEVEX, broadcastBytes: 8, zeroing: true}, 64},
}

type x86PackedAddForm struct {
	x86PackedBinaryForm
	spec x86PackedAddSpec
}

func x86PackedWrappingAddProperties(ins Instr, goarch string) (x86PackedAddForm, bool, error) {
	base, _, _ := strings.Cut(strings.ToUpper(string(ins.Op)), ".")
	spec, known := x86PackedWrappingAddSpecs[Op(base)]
	if !known {
		return x86PackedAddForm{}, false, nil
	}
	if spec.encoding == x86PackedLogicalLegacyInteger || spec.encoding == x86PackedLogicalLegacyFP {
		form, _, err := x86LegacyIntegerArithmeticProperties(ins.Op, ins, goarch)
		if err == nil {
			err = x86LegacyIntegerArithmeticEmitterContext(form, ins)
		}
		return x86PackedAddForm{form.x86PackedBinaryForm, spec}, true, err
	}
	form, err := x86PackedBinaryProperties(ins, goarch, spec.x86PackedBinarySpec)
	if err == nil {
		err = x86Mode32VectorRegisterContext(ins, goarch)
	}
	return x86PackedAddForm{form, spec}, true, err
}
