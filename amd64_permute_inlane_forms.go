package plan9asm

import (
	"fmt"
	"strings"
)

// Both members share Go's complete eighteen-row _yvpermilpd grammar. The
// immediate selects data bytes; variable control reads a separate vector.
var x86InLanePermuteBits = map[Op]int{"VPERMILPS": 32, "VPERMILPD": 64}

type x86InLanePermuteForm struct {
	bits, byteWidth, source, data, destination, mask, memory, memoryBytes int
	immediate, broadcast, zeroing                                         bool
	selector                                                              uint8
}

func x86InLanePermuteProperties(op Op, ins Instr, arch string) (x86InLanePermuteForm, bool, error) {
	base, suffix, dotted := strings.Cut(strings.ToUpper(string(op)), ".")
	bits, known := x86InLanePermuteBits[Op(base)]
	f := x86InLanePermuteForm{bits: bits, source: 0, data: 1, destination: len(ins.Args) - 1, mask: -1, memory: -1}
	if !known {
		return f, false, nil
	}
	fail := func() (x86InLanePermuteForm, bool, error) {
		return f, true, fmt.Errorf("%s %s outside Go's complete in-lane permute grammar: %q", arch, op, ins.Raw)
	}
	if arch == "" {
		arch = "amd64"
	}
	if arch != "amd64" && arch != "386" || ins.Op != "" && !strings.EqualFold(string(ins.Op), string(op)) ||
		len(ins.Args) != 3 && len(ins.Args) != 4 {
		return fail()
	}
	switch suffix {
	case "":
		if dotted {
			return fail()
		}
	case "Z":
		f.zeroing = true
	case "BCST":
		f.broadcast = true
	case "BCST.Z":
		f.broadcast, f.zeroing = true, true
	default:
		return fail()
	}
	if len(ins.Args) == 4 {
		if arch == "386" || !amd64NonzeroKOperand(ins.Args[2]) {
			return fail() // Named mode32 and its existing raw decoder have no masked row.
		}
		f.mask = 2
	} else if f.zeroing {
		return fail()
	}
	f.byteWidth = amd64VectorByteWidth(ins.Args[f.destination].Reg)
	register := func(operand Operand) bool {
		if !amd64EVEXVectorRegister(operand, f.byteWidth) {
			return false
		}
		index, _ := amd64VectorRegisterIndex(operand.Reg, f.byteWidth)
		return arch != "386" || f.byteWidth != 64 || index < 8
	}
	if f.byteWidth != 16 && f.byteWidth != 32 && f.byteWidth != 64 || !register(ins.Args[f.destination]) {
		return fail()
	}
	f.immediate = ins.Args[0].Kind == OpImm
	if f.immediate {
		constant, integer := x86ImmediateShuffleConstant(ins.Args[0], arch, ins.x86Encoded)
		if !integer || constant.Imm < -128 || constant.Imm > 255 {
			return fail()
		}
		f.source, f.data, f.selector = 1, -1, uint8(constant.Imm)
		if constant.Imm < 0 && (f.broadcast || amd64PackedWordShuffleRequiresEVEX(
			ins.Args[f.source], ins.Args[f.destination], f.byteWidth, f.mask >= 0)) {
			return fail()
		}
	} else if !register(ins.Args[f.data]) {
		return fail()
	}
	source := ins.Args[f.source]
	if isAMD64MemoryOperand(source) && x86PackedShiftMemoryValid(source, arch) {
		f.memory, f.memoryBytes = f.source, f.byteWidth
		if f.broadcast {
			f.memoryBytes = bits / 8
		}
	} else if f.broadcast || !register(source) {
		return fail()
	}
	return f, true, x86Mode32VectorRegisterContext(ins, arch)
}

func (f x86InLanePermuteForm) sourceByte(output int) int {
	if f.broadcast {
		return output % (f.bits / 8)
	}
	bytes := f.bits / 8
	lane := output / bytes
	control, mask := uint(lane), uint8(1)
	if f.bits == 32 {
		control, mask = uint(2*(lane%4)), 3
	}
	return output/16*16 + int(f.selector>>control&mask)*bytes + output%bytes
}

func x86TypedInLanePermuteByteEffect(op Op, ins Instr, arch string) (x86TypedByteEffect, bool, error) {
	f, known, err := x86InLanePermuteProperties(op, ins, arch)
	if !known || err != nil {
		return x86TypedByteEffect{}, known, err
	}
	width := f.byteWidth
	if f.memory >= 0 {
		width = f.memoryBytes
	}
	effect := x86TypedByteEffect{requiresResultProof: true, operands: []x86TypedByteOperand{
		{index: f.source, operand: ins.Args[f.source], width: int64(width), read: true},
		{index: f.destination, operand: ins.Args[f.destination], width: int64(f.byteWidth), write: true},
	}}
	if f.data >= 0 {
		effect.operands = append(effect.operands, x86TypedByteOperand{
			index: f.data, operand: ins.Args[f.data], width: int64(f.byteWidth), read: true})
	}
	if f.mask >= 0 {
		effect.operands = append(effect.operands, x86TypedByteOperand{index: f.mask, operand: ins.Args[f.mask], width: 8, read: true})
		if !f.zeroing {
			effect.operands[1].read = true
		}
	}
	return effect, true, nil
}
