package plan9asm

import (
	"fmt"
	"strings"
)

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
	if ins.Op != "" && !strings.EqualFold(string(ins.Op), string(op)) {
		return x86PackedAverageForm{}, true, fmt.Errorf("packed average dispatch/source opcode mismatch: %s/%s", op, ins.Op)
	}
	local := ins
	local.Op = op
	local.Args = append([]Operand(nil), ins.Args...)
	if spec.encoding == x86PackedLogicalLegacyInteger {
		var err error
		local, err = x86LegacyMMXReadInstruction(local, arch)
		if err != nil {
			return x86PackedAverageForm{}, true, err
		}
	}
	grammarArch := arch
	if arch == "386" && local.x86Encoded && spec.encoding == x86PackedLogicalVEXOrEVEX {
		grammarArch = "amd64" // Only the complete physical decoder sets this.
	}
	binary, err := x86PackedBinaryProperties(local, grammarArch, spec.x86PackedBinarySpec)
	form := x86PackedAverageForm{x86PackedBinaryForm: binary, spec: spec}
	if err == nil {
		form.source = local.Args[binary.first]
		if arch == "386" && grammarArch != arch && binary.memory >= 0 && !x86PackedShiftMemoryValid(form.source, arch) {
			err = fmt.Errorf("386 %s physical source has an unbound address", op)
		}
		if err == nil {
			err = x86Mode32VectorRegisterContext(local, arch)
		}
		if err == nil && binary.memory >= 0 && (ins.x86AddressBits != 0 || ins.x86UnboundAddressOverride) {
			err = fmt.Errorf("%w: packed average has no bound address-size override", ErrProbeNeedsContext)
		}
	}
	return form, true, err
}

func x86TypedPackedAverageByteEffect(op Op, ins Instr, arch string) (x86TypedByteEffect, bool, error) {
	form, known, err := x86PackedAverageProperties(op, ins, arch)
	if !known || err != nil || !form.legacy {
		return x86TypedByteEffect{}, known && (err != nil || form.legacy), err
	}
	return x86TypedByteEffect{operands: []x86TypedByteOperand{
		{index: form.first, operand: form.source, width: int64(form.byteWidth), read: true},
		{index: form.destination, operand: ins.Args[form.destination], width: int64(form.byteWidth), read: true, write: true},
	}}, true, nil
}

func x86ResultPackedAverageStep(ins Instr, state *x86ResultState, owned []bool, arch string) (bool, bool) {
	form, known, err := x86PackedAverageProperties(ins.Op, ins, arch)
	if !known || err != nil {
		return false, known
	}
	return x86ResultFixedVectorInputs(ins, form.x86PackedBinaryForm, false, state, owned, arch), true
}
