package plan9asm

import (
	"fmt"
	"strings"
)

type x86IntegerMinMaxForm struct {
	x86PackedBinaryForm
	spec amd64PackedIntegerMinMaxSpec
}

func x86IntegerMinMaxProperties(op Op, ins Instr, arch string) (x86IntegerMinMaxForm, bool, error) {
	base, _, _ := strings.Cut(strings.ToUpper(string(op)), ".")
	spec, known := amd64PackedIntegerMinMaxSpecs[base]
	form := x86IntegerMinMaxForm{spec: spec}
	if !known {
		return form, false, nil
	}
	if ins.Op != "" && !strings.EqualFold(string(ins.Op), string(op)) {
		return form, true, fmt.Errorf("min/max dispatch/source opcode mismatch")
	}
	grammar := x86PackedBinarySpec{encoding: x86PackedLogicalVEXOrEVEX, zeroing: true}
	if spec.laneBits == 64 {
		grammar.encoding = x86PackedLogicalEVEX
	}
	if spec.broadcast {
		grammar.broadcastBytes = spec.laneBits / 8
	}
	local := ins
	local.Op = op
	var err error
	form.x86PackedBinaryForm, err = x86ConflictFunnelOperandGrammar(local, arch, grammar, false)
	return form, true, err
}

func x86TypedIntegerMinMaxByteEffect(op Op, ins Instr, arch string) (x86TypedByteEffect, bool, error) {
	form, known, err := x86IntegerMinMaxProperties(op, ins, arch)
	if !known || err != nil {
		return x86TypedByteEffect{}, known, err
	}
	width, firstWidth := int64(form.byteWidth), int64(form.byteWidth)
	if form.memory >= 0 {
		firstWidth = int64(form.memoryBytes)
	}
	effect := x86TypedByteEffect{requiresResultProof: true, operands: []x86TypedByteOperand{
		{index: form.first, operand: ins.Args[form.first], width: firstWidth, read: true},
		{index: form.second, operand: ins.Args[form.second], width: width, read: true},
		{index: form.destination, operand: ins.Args[form.destination], width: width,
			read: form.mask >= 0 && !form.zeroing, write: true},
	}}
	if form.mask >= 0 {
		effect.operands = append(effect.operands,
			x86TypedByteOperand{index: form.mask, operand: ins.Args[form.mask], width: 8, read: true})
	}
	return effect, true, nil
}
