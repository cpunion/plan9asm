package plan9asm

import (
	"fmt"
	"strings"
)

// Go calls Intel's legacy PMADDWD instruction PMADDWL. The raw decoder keeps
// that canonical spelling. ymm supplies MMX/m64 and X/m128; _yvandnpd supplies
// all VEX/EVEX widths. Each VPMADDWD row permits zeroing, never broadcasting.
var x86WordMultiplyAddSpecs = map[Op]x86PackedBinarySpec{
	"PMADDWL":  {encoding: x86PackedLogicalLegacyInteger},
	"VPMADDWD": {encoding: x86PackedLogicalVEXOrEVEX, zeroing: true},
}

type x86PackedMultiplyAddSpec struct {
	x86PackedBinarySpec
	inputBits, outputBits int
	nativeOpcode          string
}

var x86PackedMultiplyAddSpecs = map[Op]x86PackedMultiplyAddSpec{
	"PMADDWL":   {x86PackedBinarySpec{encoding: x86PackedLogicalLegacyInteger}, 16, 32, "pmaddwd"},
	"PMADDUBSW": {x86PackedBinarySpec{encoding: x86PackedLogicalLegacyFP}, 8, 16, "pmaddubsw"},
	"VPMADDWD":  {x86PackedBinarySpec{encoding: x86PackedLogicalVEXOrEVEX, zeroing: true}, 16, 32, ""},
	"VPMADDUBSW": {x86PackedBinarySpec{encoding: x86PackedLogicalVEXOrEVEX, zeroing: true},
		8, 16, ""},
}

type x86PackedMultiplyAddForm struct {
	x86PackedBinaryForm
	spec   x86PackedMultiplyAddSpec
	source Operand
}

// This describes Go's actual ymm/yxm_q4/_yvandnpd operand rows only. WIG
// and the decoder's private physical marker do not grant a native address,
// initialized input, caller frame, or a new MMX transport.
func x86PackedMultiplyAddProperties(op Op, ins Instr, arch string) (x86PackedMultiplyAddForm, bool, error) {
	base, _, _ := strings.Cut(strings.ToUpper(string(op)), ".")
	spec, known := x86PackedMultiplyAddSpecs[Op(base)]
	if !known {
		return x86PackedMultiplyAddForm{}, false, nil
	}
	if ins.Op != "" && !strings.EqualFold(string(ins.Op), string(op)) {
		return x86PackedMultiplyAddForm{}, true, fmt.Errorf("packed multiply-add dispatch/source opcode mismatch: %s/%s", op, ins.Op)
	}
	local := ins
	local.Op = op
	local.Args = append([]Operand(nil), ins.Args...)
	if spec.encoding == x86PackedLogicalLegacyInteger && arch == "amd64" && len(local.Args) == 2 {
		if _, mmx := amd64ParseMReg(local.Args[1].Reg); local.Args[1].Kind == OpReg && mmx {
			local.Args[0] = x86WordMultiplyMMXSource(local.Args[0])
		}
	}
	grammarArch := arch
	if arch == "386" && local.x86Encoded && spec.encoding == x86PackedLogicalVEXOrEVEX {
		grammarArch = "amd64"
	}
	binary, err := x86PackedBinaryProperties(local, grammarArch, spec.x86PackedBinarySpec)
	form := x86PackedMultiplyAddForm{x86PackedBinaryForm: binary, spec: spec}
	if err == nil {
		form.source = local.Args[binary.first]
		if arch == "386" && grammarArch != arch && binary.memory >= 0 && !x86PackedShiftMemoryValid(form.source, arch) {
			err = fmt.Errorf("386 %s physical source has an unbound address", op)
		}
		if err == nil {
			err = x86Mode32VectorRegisterContext(local, arch)
		}
		if err == nil && binary.memory >= 0 && (ins.x86AddressBits != 0 || ins.x86UnboundAddressOverride) {
			err = fmt.Errorf("%w: packed multiply-add has no bound address-size override", ErrProbeNeedsContext)
		}
	}
	return form, true, err
}

func x86WordMultiplyAddProperties(ins Instr, goarch string) (x86PackedBinaryForm, bool, error) {
	base, _, _ := strings.Cut(strings.ToUpper(string(ins.Op)), ".")
	_, recognized := x86WordMultiplyAddSpecs[Op(base)]
	if !recognized {
		return x86PackedBinaryForm{}, false, nil
	}
	// Keep this word-only consumer boundary for the existing frame proof.
	form, known, err := x86PackedMultiplyAddProperties(ins.Op, ins, goarch)
	return form.x86PackedBinaryForm, known, err
}

func x86TypedPackedMultiplyAddByteEffect(op Op, ins Instr, arch string) (x86TypedByteEffect, bool, error) {
	form, known, err := x86PackedMultiplyAddProperties(op, ins, arch)
	if !known || err != nil || !form.legacy {
		return x86TypedByteEffect{}, known && (err != nil || form.legacy), err
	}
	return x86TypedByteEffect{operands: []x86TypedByteOperand{
		{index: form.first, operand: form.source, width: int64(form.byteWidth), read: true},
		{index: form.destination, operand: ins.Args[form.destination], width: int64(form.byteWidth), read: true, write: true},
	}}, true, nil
}
