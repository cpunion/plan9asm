package plan9asm

import (
	"fmt"
	"strings"
)

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
	form, err := x86PackedBinaryProperties(ins, goarch, spec.x86PackedBinarySpec)
	if err == nil && goarch == "386" && !form.legacy {
		for _, operand := range ins.Args {
			index, valid := amd64VectorRegisterIndex(operand.Reg, form.byteWidth)
			if operand.Kind == OpReg && valid && index >= 8 {
				return x86PackedAddForm{form, spec}, true, fmt.Errorf("%w: 386 packed add register %s has no proved physical mode32 binding", ErrProbeNeedsContext, operand.Reg)
			}
		}
	}
	return x86PackedAddForm{form, spec}, true, err
}
