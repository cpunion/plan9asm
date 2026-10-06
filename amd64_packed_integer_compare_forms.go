package plan9asm

import (
	"fmt"
	"strings"
)

type x86PackedIntegerCompareSpec struct {
	laneBits int
	greater  bool
	legacy   bool
	mmx      bool
}

var x86PackedIntegerCompareSpecs = map[Op]x86PackedIntegerCompareSpec{
	"PCMPEQB":  {8, false, true, true},
	"PCMPEQW":  {16, false, true, true},
	"PCMPEQL":  {32, false, true, true},
	"PCMPEQQ":  {64, false, true, false},
	"PCMPGTB":  {8, true, true, true},
	"PCMPGTW":  {16, true, true, true},
	"PCMPGTL":  {32, true, true, true},
	"PCMPGTQ":  {64, true, true, false},
	"VPCMPEQB": {8, false, false, false},
	"VPCMPEQW": {16, false, false, false},
	"VPCMPEQD": {32, false, false, false},
	"VPCMPEQQ": {64, false, false, false},
	"VPCMPGTB": {8, true, false, false},
	"VPCMPGTW": {16, true, false, false},
	"VPCMPGTD": {32, true, false, false},
	"VPCMPGTQ": {64, true, false, false},
}

type x86PackedIntegerCompareForm struct {
	x86PackedBinaryForm
	spec       x86PackedIntegerCompareSpec
	source     Operand
	maskResult bool
}

// Properties describe the Go operand rows and physical mode, not initialized
// inputs, native backing, memory bounds, or caller authority. The raw decoder's
// eight-entry encoding registry remains distinct from these sixteen Go names.
func x86PackedIntegerCompareProperties(op Op, ins Instr, arch string) (x86PackedIntegerCompareForm, bool, error) {
	base, suffix, dotted := strings.Cut(strings.ToUpper(string(op)), ".")
	spec, known := x86PackedIntegerCompareSpecs[Op(base)]
	if !known {
		return x86PackedIntegerCompareForm{}, false, nil
	}
	form := x86PackedIntegerCompareForm{spec: spec}
	if ins.Op != "" && !strings.EqualFold(string(ins.Op), string(op)) {
		return form, true, fmt.Errorf("packed integer compare dispatch/source opcode mismatch: %s/%s", op, ins.Op)
	}
	local := ins
	local.Op = op
	local.Args = append([]Operand(nil), ins.Args...)
	fail := func() (x86PackedIntegerCompareForm, bool, error) {
		return form, true, fmt.Errorf("%s %s operands/suffix outside Go packed integer compare grammar: %q", arch, op, ins.Raw)
	}
	if arch != "amd64" && arch != "386" {
		return fail()
	}
	if spec.legacy || len(local.Args) == 3 && local.Args[2].Kind == OpReg && !amd64IsKOperand(local.Args[2]) {
		encoding := x86PackedLogicalVEX
		if spec.legacy {
			encoding = x86PackedLogicalLegacyFP
			if spec.mmx {
				encoding = x86PackedLogicalLegacyInteger
				var err error
				local, err = x86LegacyMMXReadInstruction(local, arch)
				if err != nil {
					return form, true, err
				}
			}
		}
		binary, err := x86PackedBinaryProperties(local, arch, x86PackedBinarySpec{encoding: encoding})
		form.x86PackedBinaryForm = binary
		if err != nil {
			return form, true, err
		}
	} else {
		// _yvpcmpeqb EVEX rows produce K0-7, not an X/Y/Z destination;
		// their optional input mask is K1-7. There is no .Z form.
		form.maskResult = true
		form.x86PackedBinaryForm = x86PackedBinaryForm{
			first: 0, second: 1, destination: len(local.Args) - 1,
			mask: -1, memory: -1,
		}
		if len(local.Args) != 3 && len(local.Args) != 4 || dotted && suffix != "BCST" {
			return fail()
		}
		if !amd64IsKOperand(local.Args[form.destination]) {
			return fail()
		}
		width, valid := amd64PackedCompareWidth(local.Args[1])
		if !valid || !amd64GoPackedVectorMoveRegister(local.Args[1], width, "amd64") {
			return fail()
		}
		form.byteWidth = width
		first := local.Args[0]
		if first.Kind == OpReg {
			if !amd64GoPackedVectorMoveRegister(first, width, "amd64") {
				return fail()
			}
		} else if !x86PackedShiftMemoryValid(first, arch) {
			return fail()
		} else {
			form.memory, form.memoryBytes = 0, width
		}
		if dotted {
			if spec.laneBits < 32 || form.memory < 0 {
				return fail()
			}
			form.broadcast, form.memoryBytes = true, spec.laneBits/8
		}
		if len(local.Args) == 4 {
			if !amd64NonzeroKOperand(local.Args[2]) || arch == "386" && !local.x86Encoded {
				return fail()
			}
			form.mask = 2
		}
	}
	form.source = local.Args[form.first]
	if err := x86Mode32VectorRegisterContext(local, arch); err != nil {
		return form, true, err
	}
	if form.memory >= 0 && (ins.x86AddressBits != 0 || ins.x86UnboundAddressOverride) {
		return form, true, fmt.Errorf("%w: packed integer compare has no bound address-size override", ErrProbeNeedsContext)
	}
	return form, true, nil
}

func x86TypedPackedIntegerCompareByteEffect(op Op, ins Instr, arch string) (x86TypedByteEffect, bool, error) {
	form, known, err := x86PackedIntegerCompareProperties(op, ins, arch)
	if !known || err != nil {
		return x86TypedByteEffect{}, known, err
	}
	firstWidth := form.byteWidth
	if form.memory >= 0 {
		firstWidth = form.memoryBytes
	}
	operands := []x86TypedByteOperand{
		{index: form.first, operand: form.source, width: int64(firstWidth), read: true},
		{index: form.second, operand: ins.Args[form.second], width: int64(form.byteWidth), read: true},
	}
	if form.mask >= 0 {
		operands = append(operands, x86TypedByteOperand{index: form.mask, operand: ins.Args[form.mask], width: 8, read: true})
	}
	width := form.byteWidth
	if form.maskResult {
		width = 8
	}
	operands = append(operands, x86TypedByteOperand{
		index: form.destination, operand: ins.Args[form.destination], width: int64(width), write: true,
	})
	return x86TypedByteEffect{operands: operands}, true, nil
}
