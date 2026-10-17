package plan9asm

import (
	"fmt"
	"strings"
)

type x86BEXTRForm struct {
	bits                         int
	control, source, destination Operand
}

// Both complete Go _ybextrl rows are Yrl, Yml, Yrl with VEX.LZ/0F38/F7.
// W selects the qword only in mode64; the original Q spelling stays intact.
func x86BEXTRProperties(op Op, ins Instr, arch string) (x86BEXTRForm, bool, error) {
	name, _, dotted := strings.Cut(strings.ToUpper(string(op)), ".")
	form := x86BEXTRForm{bits: 32}
	switch name {
	case "BEXTRL":
	case "BEXTRQ":
		if arch == "amd64" {
			form.bits = 64
		}
	default:
		return form, false, nil
	}
	fail := func() (x86BEXTRForm, bool, error) {
		return form, true, fmt.Errorf("%s %s outside complete Go BEXTR grammar: %q", arch, op, ins.Raw)
	}
	if arch != "386" && arch != "amd64" || dotted || len(ins.Args) != 3 ||
		ins.Op != "" && !strings.EqualFold(string(ins.Op), name) {
		return fail()
	}
	form.control, form.source, form.destination = ins.Args[0], ins.Args[1], ins.Args[2]
	for _, operand := range []Operand{form.control, form.destination} {
		if operand.Kind != OpReg || !isX86YrlRegisterForArch(operand.Reg, arch) {
			return fail()
		}
	}
	if form.source.Kind == OpReg {
		if !isX86YrlRegisterForArch(form.source.Reg, arch) {
			return fail()
		}
	} else if !isAMD64MemoryOperand(form.source) || !x86PackedShiftMemoryValid(form.source, arch) ||
		x86NamedMultiplyMemoryError(form.source, arch) != nil {
		return fail()
	}
	return form, true, nil
}

func x86TypedBEXTRByteEffect(op Op, ins Instr, arch string) (x86TypedByteEffect, bool, error) {
	form, known, err := x86BEXTRProperties(op, ins, arch)
	if !known || err != nil {
		return x86TypedByteEffect{}, known, err
	}
	width := int64(form.bits / 8)
	return x86TypedByteEffect{
		operands: []x86TypedByteOperand{
			{index: 0, operand: form.control, width: 2, read: true},
			{index: 1, operand: form.source, width: width, read: true},
			{index: 2, operand: form.destination, width: width, write: true},
		},
		flagsDefined:        x86ByteFlagCF | x86ByteFlagOF | x86ByteFlagZF,
		flagsKilled:         x86ByteFlagsAll,
		requiresResultProof: true,
	}, true, nil
}

// A paired native read supplies numerical source bytes only if it continues.
// It grants no caller, extent, raw-entry or reusable native-memory capability.
func x86ResultBEXTRStep(ins Instr, state *x86ResultState, owned []bool, resultOffset int64,
	arch string, nativeRead bool) (bool, bool) {
	effect, known, err := x86TypedBEXTRByteEffect(ins.Op, ins, arch)
	if !known || err != nil {
		return false, known
	}
	if !x86ResultNamedPredicateInstruction(ins) {
		return false, true
	}
	for _, operand := range effect.operands {
		if !operand.read {
			continue
		}
		if operand.operand.Kind == OpReg {
			index, _, valid := x86ResultGP(operand.operand.Reg, operand.width, arch)
			if !valid || state.pointerGP[index] {
				return false, true
			}
		}
		if !x86ResultReadWithNativeContinuation(ins, operand.operand, operand.width, state, owned, arch, nativeRead) {
			return false, true
		}
	}
	output := effect.operands[2]
	if !x86ResultStore(output.operand, output.width, state, owned, resultOffset, arch) {
		return false, true
	}
	state.flags = state.flags&^effect.flagsKilled | effect.flagsDefined
	return true, true
}
