package plan9asm

import (
	"fmt"
	"strings"
)

type x86DwordProductSpec struct {
	x86PackedBinarySpec
	signed       bool
	nativeOpcode string
}

// Go names the legacy unsigned instruction PMULULQ, not Intel's PMULUDQ.
// Only the unsigned legacy ymm table has an MMX/m64 row. Both AVX tables
// permit qword broadcast and qword writemasking, never rounding or SAE.
var x86DwordProductSpecs = map[Op]x86DwordProductSpec{
	"PMULDQ":   {x86PackedBinarySpec{encoding: x86PackedLogicalLegacyFP}, true, "pmuldq"},
	"PMULULQ":  {x86PackedBinarySpec{encoding: x86PackedLogicalLegacyInteger}, false, "pmuludq"},
	"VPMULDQ":  {x86PackedBinarySpec{encoding: x86PackedLogicalVEXOrEVEX, broadcastBytes: 8, zeroing: true}, true, ""},
	"VPMULUDQ": {x86PackedBinarySpec{encoding: x86PackedLogicalVEXOrEVEX, broadcastBytes: 8, zeroing: true}, false, ""},
}

type x86DwordProductForm struct {
	x86PackedBinaryForm
	spec   x86DwordProductSpec
	source Operand
}

// The operand model grants no initialized bank, native extent or caller role.
// In particular a GP spelling admitted to Ymm still names the physical MMX
// bank, and a named high mode32 vector still needs its physical binding.
func x86DwordProductProperties(op Op, ins Instr, arch string) (x86DwordProductForm, bool, error) {
	base, _, _ := strings.Cut(strings.ToUpper(string(op)), ".")
	spec, known := x86DwordProductSpecs[Op(base)]
	if !known {
		return x86DwordProductForm{}, false, nil
	}
	if ins.Op != "" && !strings.EqualFold(string(ins.Op), string(op)) {
		return x86DwordProductForm{}, true, fmt.Errorf("dword product dispatch/source opcode mismatch: %s/%s", op, ins.Op)
	}
	local := ins
	local.Op = op
	local.Args = append([]Operand(nil), ins.Args...)
	if spec.encoding == x86PackedLogicalLegacyInteger {
		var err error
		local, err = x86LegacyMMXReadInstruction(local, arch)
		if err != nil {
			return x86DwordProductForm{}, true, err
		}
	}
	grammarArch := arch
	if arch == "386" && local.x86Encoded && spec.encoding == x86PackedLogicalVEXOrEVEX {
		grammarArch = "amd64"
	}
	binary, err := x86PackedBinaryProperties(local, grammarArch, spec.x86PackedBinarySpec)
	form := x86DwordProductForm{x86PackedBinaryForm: binary, spec: spec}
	if err == nil {
		form.source = local.Args[binary.first]
		if arch == "386" && grammarArch != arch && binary.memory >= 0 && !x86PackedShiftMemoryValid(form.source, arch) {
			err = fmt.Errorf("386 %s physical source has an unbound address", op)
		}
		if err == nil {
			err = x86Mode32VectorRegisterContext(local, arch)
		}
		if err == nil && binary.memory >= 0 && (ins.x86AddressBits != 0 || ins.x86UnboundAddressOverride) {
			err = fmt.Errorf("%w: dword product has no bound address-size override", ErrProbeNeedsContext)
		}
	}
	return form, true, err
}

func x86TypedDwordProductByteEffect(op Op, ins Instr, arch string) (x86TypedByteEffect, bool, error) {
	form, known, err := x86DwordProductProperties(op, ins, arch)
	if !known || err != nil || !form.legacy {
		return x86TypedByteEffect{}, known && (err != nil || form.legacy), err
	}
	return x86TypedByteEffect{operands: []x86TypedByteOperand{
		{index: form.first, operand: form.source, width: int64(form.byteWidth), read: true},
		{index: form.destination, operand: ins.Args[form.destination], width: int64(form.byteWidth), read: true, write: true},
	}}, true, nil
}
