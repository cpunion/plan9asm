package plan9asm

import (
	"fmt"
	"strings"
)

type x86RawHalfConversionForm struct {
	op        Op
	mapNumber int
	opcode    int
	immediate bool
}

// These are the complete Go 1.27 F16C/AVX-512 half-conversion opcode rows.
// Both share the same X/Y/Z width axis, but VCVTPS2PH reverses the ModRM
// source/destination roles and appends an imm8 rounding control.
var x86RawHalfConversionForms = [...]x86RawHalfConversionForm{
	{op: "VCVTPH2PS", mapNumber: 2, opcode: 0x13},
	{op: "VCVTPS2PH", mapNumber: 3, opcode: 0x1d, immediate: true},
}

func decodedX86PackedHalfConversionInstruction(code []byte, mode int) (Instr, int, bool, error) {
	p, recognized := decodeX86RawVectorEncoding(code)
	if !recognized || p.pp != 1 {
		return Instr{}, 0, false, nil
	}
	var form x86RawHalfConversionForm
	for _, candidate := range x86RawHalfConversionForms {
		if p.mapNumber == candidate.mapNumber && p.opcode == candidate.opcode {
			form = candidate
			break
		}
	}
	if form.op == "" {
		return Instr{}, 0, false, nil
	}
	fail := func(message string) (Instr, int, bool, error) {
		return Instr{}, 0, true, fmt.Errorf("packed half conversion: %s", message)
	}
	if mode != 32 && mode != 64 {
		return fail("unsupported x86 mode")
	}
	if p.addressOverride {
		return fail("address-size override is not source-layout safe")
	}
	if p.upper != 0 || p.w || p.evex && !p.fixed {
		return fail("invalid reserved vvvv/V', W or EVEX fixed bit")
	}
	if p.evex && p.zero && p.mask == 0 {
		return fail("zeroing requires a nonzero mask")
	}
	if p.vectorLength > 2 {
		return fail("reserved vector length")
	}
	if len(code) <= p.modRM {
		return fail("missing ModRM byte")
	}
	if !p.evex && mode == 32 && (p.r != 0 || p.b != 0 || p.x != 0) {
		return fail("extended VEX register in 32-bit mode")
	}
	if p.evex && mode == 32 && code[p.modRM]>>6 != 3 && (p.b != 0 || p.x != 0) {
		return fail("extended memory address in 32-bit mode")
	}
	if p.broadcast {
		if !p.evex || p.vectorLength != 2 || form.op == "VCVTPH2PS" && code[p.modRM]>>6 != 3 {
			return fail("SAE requires the EVEX Z-width register form")
		}
	}

	vector := [...]string{"X", "Y", "Z"}[p.vectorLength]
	shorter := "X"
	if p.vectorLength == 2 {
		shorter = "Y"
	}
	op := form.op
	if p.broadcast {
		op += ".SAE"
	}
	if p.zero {
		op += ".Z"
	}
	modRM := code[p.modRM]
	regNumber := int(modRM>>3&7) + p.r*8
	register := func(prefix string) Operand {
		return Operand{Kind: OpReg, Reg: Reg(fmt.Sprintf("%s%d", prefix, regNumber))}
	}
	decodeRM := func(prefix string) (Operand, int, error) {
		if p.evex {
			return decodedX86EVEXRMOperand(code[p.modRM:], mode, p.b, p.x, p.segment, prefix, 8<<p.vectorLength)
		}
		return decodedX86VEXVectorRMOperand(code[p.modRM:], mode, p.b, p.x, p.segment, prefix)
	}

	args := make([]Operand, 0, 4)
	var consumed int
	if !form.immediate {
		source, n, err := decodeRM(shorter)
		if err != nil {
			return Instr{}, 0, true, err
		}
		consumed = n
		args = append(args, source)
		if p.mask != 0 {
			args = append(args, Operand{Kind: OpReg, Reg: Reg(fmt.Sprintf("K%d", p.mask))})
		}
		args = append(args, register(vector))
	} else {
		destination, n, err := decodeRM(shorter)
		if err != nil {
			return Instr{}, 0, true, err
		}
		consumed = n
		if len(code) <= p.modRM+consumed {
			return fail("missing imm8 rounding control")
		}
		args = append(args, Operand{Kind: OpImm, Imm: int64(code[p.modRM+consumed])})
		args = append(args, register(vector))
		if p.mask != 0 {
			args = append(args, Operand{Kind: OpReg, Reg: Reg(fmt.Sprintf("K%d", p.mask))})
		}
		args = append(args, destination)
		consumed++
	}
	printed := make([]string, len(args))
	for index, arg := range args {
		printed[index] = arg.String()
	}
	return Instr{Op: op, Args: args, Raw: fmt.Sprintf("%s %s", op, strings.Join(printed, ", "))}, p.modRM + consumed, true, nil
}
