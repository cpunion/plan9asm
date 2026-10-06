package plan9asm

import (
	"fmt"
	"strings"
)

type x86PackedWordMultiplySpec struct {
	mode     amd64PackedWordMultiplyMode
	encoding x86PackedLogicalEncoding
}

var x86PackedWordMultiplySpecs = map[Op]x86PackedWordMultiplySpec{
	"PMULLW":   {amd64PackedWordMultiplyLow, x86PackedLogicalLegacyInteger},
	"PMULHW":   {amd64PackedWordMultiplySignedHigh, x86PackedLogicalLegacyInteger},
	"PMULHUW":  {amd64PackedWordMultiplyUnsignedHigh, x86PackedLogicalLegacyInteger},
	"PMULHRSW": {amd64PackedWordMultiplyRoundedSignedHigh, x86PackedLogicalLegacyFP},
	"VPMULLW":  {amd64PackedWordMultiplyLow, x86PackedLogicalVEXOrEVEX},
	"VPMULHW":  {amd64PackedWordMultiplySignedHigh, x86PackedLogicalVEXOrEVEX},
	"VPMULHUW": {amd64PackedWordMultiplyUnsignedHigh, x86PackedLogicalVEXOrEVEX},
	"VPMULHRSW": {amd64PackedWordMultiplyRoundedSignedHigh,
		x86PackedLogicalVEXOrEVEX},
}

type x86PackedWordMultiplyForm struct {
	x86PackedBinaryForm
	spec   x86PackedWordMultiplySpec
	source Operand // Actual encoded bank, not a new initialized input.
}

// Go's ymm source row covers Yax/Ycx/Yrx/Yrl. Its ModRM field names the
// MMX bank even when the source spelling is a GP register. No GP value is
// transported into that bank by the assembler.
func x86WordMultiplyMMXSource(operand Operand) Operand {
	if operand.Kind != OpReg {
		return operand
	}
	for index, register := range [...]Reg{AX, CX, DX, BX, SP, BP, SI, DI,
		"R8", "R9", "R10", "R11", "R12", "R13", "R14", "R15"} {
		if operand.Reg == register {
			operand.Reg = Reg(fmt.Sprintf("M%d", index&7))
			return operand
		}
	}
	return operand
}

// This shared grammar records the real native read/modify/write shape only.
// It grants neither input initialization nor native backing/caller authority.
func x86PackedWordMultiplyProperties(op Op, ins Instr, arch string) (x86PackedWordMultiplyForm, bool, error) {
	base, _, _ := strings.Cut(strings.ToUpper(string(op)), ".")
	spec, recognized := x86PackedWordMultiplySpecs[Op(base)]
	if !recognized {
		return x86PackedWordMultiplyForm{}, false, nil
	}
	if ins.Op != "" && !strings.EqualFold(string(ins.Op), string(op)) {
		return x86PackedWordMultiplyForm{}, true, fmt.Errorf("packed word multiply dispatch/source opcode mismatch: %s/%s", op, ins.Op)
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
	// Only the actual physical decoder sets this private marker. A physical
	// mode32 EVEX instruction is not subject to Go's named 3-operand limit.
	if arch == "386" && local.x86Encoded && spec.encoding == x86PackedLogicalVEXOrEVEX {
		grammarArch = "amd64"
	}
	binary, err := x86PackedBinaryProperties(local, grammarArch, x86PackedBinarySpec{
		encoding: spec.encoding, zeroing: spec.encoding == x86PackedLogicalVEXOrEVEX,
	})
	form := x86PackedWordMultiplyForm{x86PackedBinaryForm: binary, spec: spec}
	if err == nil {
		form.source = local.Args[binary.first]
		if arch == "386" && grammarArch != arch && binary.memory >= 0 && !x86PackedShiftMemoryValid(form.source, arch) {
			err = fmt.Errorf("386 %s physical source has an unbound address", op)
		}
		if err == nil {
			err = x86Mode32VectorRegisterContext(local, arch)
		}
		if err == nil && binary.memory >= 0 && (ins.x86AddressBits != 0 || ins.x86UnboundAddressOverride) {
			err = fmt.Errorf("%w: packed word multiply has no bound address-size override", ErrProbeNeedsContext)
		}
	}
	return form, true, err
}

func x86TypedPackedWordMultiplyByteEffect(op Op, ins Instr, arch string) (x86TypedByteEffect, bool, error) {
	form, recognized, err := x86PackedWordMultiplyProperties(op, ins, arch)
	if !recognized || err != nil || !form.legacy {
		return x86TypedByteEffect{}, recognized && (err != nil || form.legacy), err
	}
	return x86TypedByteEffect{operands: []x86TypedByteOperand{
		{index: form.first, operand: form.source, width: int64(form.byteWidth), read: true},
		{index: form.destination, operand: ins.Args[form.destination], width: int64(form.byteWidth), read: true, write: true},
	}}, true, nil
}
