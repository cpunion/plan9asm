package plan9asm

import "strings"

type x86PackedAverageSpec struct {
	x86PackedBinarySpec
	laneBits int
}

// Go's complete unsigned rounded-average family: ymm/Py1 for legacy and
// _yvandnpd for VEX/EVEX. E4.nb has mask suppression but no broadcast or SAE.
var x86PackedAverageSpecs = map[Op]x86PackedAverageSpec{
	"PAVGB":  {x86PackedBinarySpec{encoding: x86PackedLogicalLegacyInteger}, 8},
	"PAVGW":  {x86PackedBinarySpec{encoding: x86PackedLogicalLegacyInteger}, 16},
	"VPAVGB": {x86PackedBinarySpec{encoding: x86PackedLogicalVEXOrEVEX, zeroing: true}, 8},
	"VPAVGW": {x86PackedBinarySpec{encoding: x86PackedLogicalVEXOrEVEX, zeroing: true}, 16},
}

type x86PackedAverageForm struct {
	x86PackedBinaryForm
	spec   x86PackedAverageSpec
	source Operand
}

// A shape/effect observation supplies no input initialization, caller, frame
// or native storage authority. All consumers retain their own proof gates.
func x86PackedAverageProperties(op Op, ins Instr, arch string) (x86PackedAverageForm, bool, error) {
	base, _, _ := strings.Cut(strings.ToUpper(string(op)), ".")
	spec, known := x86PackedAverageSpecs[Op(base)]
	if !known {
		return x86PackedAverageForm{}, false, nil
	}
	binary, source, started, err := x86PackedBinaryPhysicalSource(op, ins, arch, spec.x86PackedBinarySpec, "packed average")
	if !started {
		return x86PackedAverageForm{}, true, err
	}
	return x86PackedAverageForm{binary, spec, source}, true, err
}

func x86ResultPackedAverageStep(ins Instr, state *x86ResultState, owned []bool, arch string) (bool, bool) {
	form, known, err := x86PackedAverageProperties(ins.Op, ins, arch)
	if !known || err != nil {
		return false, known
	}
	return x86ResultFixedVectorInputs(ins, form.x86PackedBinaryForm, false, state, owned, arch), true
}
