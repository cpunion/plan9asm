package plan9asm

import (
	"fmt"
	"strings"
)

type x86LegacyIntegerArithmeticForm struct {
	x86PackedBinaryForm
	source, destination Operand
}

// Reuse the arithmetic specifications and Go operand rows already owned by
// lowering. This descriptor adds no frame, source, caller or initialized bank.
func x86LegacyIntegerArithmeticProperties(op Op, ins Instr, arch string) (x86LegacyIntegerArithmeticForm, bool, error) {
	name, _, _ := strings.Cut(strings.ToUpper(string(op)), ".")
	spec, add := amd64PackedAddSpecs[Op(name)]
	_, _, vector, subtract := amd64PackedSubtractProperties(name)
	if add && spec.form == amd64PackedAddVEXEVEX || subtract && vector || !add && !subtract {
		return x86LegacyIntegerArithmeticForm{}, false, nil
	}
	if ins.Op != "" && !strings.EqualFold(string(ins.Op), string(op)) {
		return x86LegacyIntegerArithmeticForm{}, true, fmt.Errorf("legacy integer arithmetic dispatch/source mismatch: %s/%s", op, ins.Op)
	}
	local := ins
	local.Op = op
	local.Args = append([]Operand(nil), ins.Args...)
	encoding := x86PackedLogicalLegacyFP
	if add && spec.form == amd64PackedAddLegacyYMM {
		encoding = x86PackedLogicalLegacyInteger
	}
	// Go's Yax/Ycx/Yrx/Yrl -> Ymm compatibility emits a physical MMX
	// ModRM bank. It does not read the named GP value. Keep that identity in
	// the pure descriptor; emitters without this transport remain Context.
	if encoding == x86PackedLogicalLegacyInteger && arch == "amd64" && len(local.Args) == 2 {
		if _, mmx := amd64ParseMReg(local.Args[1].Reg); local.Args[1].Kind == OpReg && mmx &&
			local.Args[0].Kind == OpReg && isX86YrlRegisterForArch(local.Args[0].Reg, arch) {
			for index, reg := range []Reg{AX, CX, DX, BX, SP, BP, SI, DI, "R8", "R9", "R10", "R11", "R12", "R13", "R14", "R15"} {
				if local.Args[0].Reg == reg {
					local.Args[0].Reg = Reg(fmt.Sprintf("M%d", index&7))
					break
				}
			}
		}
	}
	form, err := x86PackedBinaryProperties(local, arch, x86PackedBinarySpec{encoding: encoding})
	result := x86LegacyIntegerArithmeticForm{x86PackedBinaryForm: form}
	if err == nil {
		result.source, result.destination = local.Args[0], local.Args[1]
	}
	return result, true, err
}

func x86TypedLegacyIntegerArithmeticByteEffect(op Op, ins Instr, arch string) (x86TypedByteEffect, bool, error) {
	form, recognized, err := x86LegacyIntegerArithmeticProperties(op, ins, arch)
	if !recognized || err != nil {
		return x86TypedByteEffect{}, recognized, err
	}
	width := int64(form.byteWidth)
	return x86TypedByteEffect{operands: []x86TypedByteOperand{
		{index: 0, operand: form.source, width: width, read: true},
		{index: 1, operand: form.destination, width: width, read: true, write: true},
	}}, true, nil
}

func x86LegacyIntegerArithmeticEmitterContext(form x86LegacyIntegerArithmeticForm, ins Instr) error {
	if form.mmx && ins.Args[0].Kind == OpReg && ins.Args[0].Reg != form.source.Reg {
		return fmt.Errorf("%w: legacy integer arithmetic GP spelling encodes an MMX source without its physical bank transport", ErrProbeNeedsContext)
	}
	return nil
}
