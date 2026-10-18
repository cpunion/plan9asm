package plan9asm

import (
	"fmt"
	"strings"
)

// These seven names are the whole-128-bit subset of Go's yxmov table.
// Scalar and half-register moves share the table, not this byte effect.
var x86LegacyPackedMoveNames = map[Op]bool{
	"MOVOU": true, "MOVOA": true, "MOVO": true,
	"MOVUPS": true, "MOVAPS": true, "MOVUPD": true, "MOVAPD": true,
}

type x86LegacyPackedMoveForm struct {
	source, destination Operand
	memory              int // -1 for a register copy
	memoryWrite         bool
	aligned             bool // Native m128 requires a separate actual-address proof.
}

// This is Go operand shape only. In particular, accepting a memory operand
// does not prove backing, initialization, alignment or caller permissions.
// These legacy rows reject high X registers in mode32; VEX/EVEX's different
// named-bank acceptance cannot be borrowed by this grammar.
func x86LegacyPackedMoveProperties(op Op, ins Instr, arch string) (x86LegacyPackedMoveForm, bool, error) {
	name, suffix, dotted := strings.Cut(strings.ToUpper(string(op)), ".")
	if !x86LegacyPackedMoveNames[Op(name)] {
		return x86LegacyPackedMoveForm{}, false, nil
	}
	form := x86LegacyPackedMoveForm{memory: -1}
	switch Op(name) {
	case "MOVOA", "MOVO", "MOVAPS", "MOVAPD":
		form.aligned = true
	}
	fail := func() (x86LegacyPackedMoveForm, bool, error) {
		return form, true, fmt.Errorf("%s %s outside Go whole-128 legacy move grammar: %q", arch, op, ins.Raw)
	}
	if arch != "amd64" && arch != "386" || dotted || suffix != "" || len(ins.Args) != 2 ||
		ins.Op != "" && !strings.EqualFold(string(ins.Op), name) {
		return fail()
	}
	vector := func(operand Operand) bool {
		index, valid := amd64ParseXReg(operand.Reg)
		return operand.Kind == OpReg && valid && index < 16 && (arch != "386" || index < 8)
	}
	form.source, form.destination = ins.Args[0], ins.Args[1]
	if vector(form.source) && vector(form.destination) {
		return form, true, nil
	}
	if vector(form.source) && isAMD64MemoryOperand(form.destination) {
		form.memory, form.memoryWrite = 1, true
	} else if isAMD64MemoryOperand(form.source) && vector(form.destination) {
		form.memory = 0
	} else {
		return fail()
	}
	memory := ins.Args[form.memory]
	if memory.Kind == OpMem {
		if err := x86SymbolIndexedMemoryError(memory.Mem, arch); err != nil {
			return form, true, err
		}
		if memory.Mem.OffRaw != "" || memory.Mem.IndexExt != "" {
			return fail()
		}
		if arch == "386" && strings.HasSuffix(memory.Mem.Sym, "(SB)") {
			memory.Mem.Base = "" // Go ignores this annotation, not a value input.
		}
	}
	if !x86PackedShiftMemoryValid(memory, arch) {
		return fail()
	}
	return form, true, nil
}

func x86TypedLegacyPackedMoveByteEffect(op Op, ins Instr, arch string) (x86TypedByteEffect, bool, error) {
	form, recognized, err := x86LegacyPackedMoveProperties(op, ins, arch)
	if !recognized || err != nil {
		return x86TypedByteEffect{}, recognized, err
	}
	if err := x86Mode32VectorRegisterContext(ins, arch); err != nil {
		return x86TypedByteEffect{}, true, err
	}
	return x86TypedByteEffect{operands: []x86TypedByteOperand{
		{index: 0, operand: form.source, width: 16, read: true},
		{index: 1, operand: form.destination, width: 16, write: true},
	}}, true, nil
}

func x86TypedInstructionByteEffect(op Op, ins Instr, arch string) (x86TypedByteEffect, bool, error) {
	if effect, known, err := x86TypedRegisterQueryByteEffect(op, ins, arch); known {
		return effect, true, err
	}
	if effect, known, err := x86TypedCompactExpandByteEffect(op, ins, arch); known {
		return effect, known, err
	}
	if effect, known, err := x86TypedMXCSRByteEffect(op, ins, arch); known {
		return effect, true, err
	}
	if effect, known, err := x86TypedByteShuffleEffect(op, ins, arch); known {
		return effect, known, err
	}
	if effect, known, err := x86TypedPackedConversionByteEffect(op, ins, arch); known {
		return effect, known, err
	}
	if effect, known, err := x86TypedCOMIByteEffect(op, ins, arch); known {
		return effect, known, err
	}
	if effect, known, err := x86TypedScalarFloatIntegerByteEffect(op, ins, arch); known {
		return effect, known, err
	}
	if effect, known, err := x86TypedAESByteEffect(op, ins, arch); known {
		return effect, known, err
	}
	if _, known, err := x86OrderingHintProperties(op, ins, arch); known {
		return x86TypedByteEffect{}, true, err
	}
	if effect, recognized, err := x86TypedGPByteEffect(op, ins, arch); recognized {
		return effect, recognized, err
	}
	if effect, known, err := x86TypedImplicitMultiplyByteEffect(op, ins, arch); known {
		return effect, known, err
	}
	if effect, known, err := x86TypedCRC32ByteEffect(op, ins, arch); known {
		return effect, known, err
	}
	if effect, recognized, err := x86TypedLegacyIntegerArithmeticByteEffect(op, ins, arch); recognized {
		return effect, recognized, err
	}
	if effect, recognized, err := x86TypedPackedLowMultiplyByteEffect(op, ins, arch); recognized {
		return effect, recognized, err
	}
	if effect, recognized, err := x86TypedPackedWordMultiplyByteEffect(op, ins, arch); recognized {
		return effect, recognized, err
	}
	if effect, recognized, err := x86TypedDwordProductByteEffect(op, ins, arch); recognized {
		return effect, recognized, err
	}
	if effect, recognized, err := x86TypedPackedAverageByteEffect(op, ins, arch); recognized {
		return effect, recognized, err
	}
	if effect, recognized, err := x86TypedPackedIntegerCompareByteEffect(op, ins, arch); recognized {
		return effect, recognized, err
	}
	if effect, recognized, err := x86TypedPackedMultiplyAddByteEffect(op, ins, arch); recognized {
		return effect, recognized, err
	}
	return x86TypedLegacyPackedMoveByteEffect(op, ins, arch)
}

// The old direct emitter API returns unhandled for a non-X destination.
// Retaining that rejection shape grants no effect and emits no partial IR.
func (c *amd64Ctx) validateLegacyPackedMove(op Op, ins Instr) (bool, error) {
	arch := c.goarch
	if arch == "" {
		arch = "amd64"
	}
	_, recognized, err := x86LegacyPackedMoveProperties(op, ins, arch)
	if !recognized {
		return false, nil
	}
	if err != nil && ins.Op == "" && len(ins.Args) == 2 && ins.Args[1].Kind == OpReg && !isAMD64XReg(ins.Args[1].Reg) {
		return false, nil
	}
	if err == nil {
		err = x86Mode32VectorRegisterContext(ins, arch)
	}
	return true, err
}
