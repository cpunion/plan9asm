package plan9asm

import (
	"fmt"
	"strings"
)

// _yvblendmpd is the existing complete EVEX binary grammar. The mode adds
// old-destination dependencies, not another register/memory/suffix parser.
func x86IndexedPermuteProperties(op Op, ins Instr, arch string) (x86PackedBinaryForm, amd64IndexedPermuteSpec, bool, error) {
	base, _, _ := strings.Cut(strings.ToUpper(string(op)), ".")
	spec, known := amd64IndexedPermuteSpecs[Op(base)]
	if !known {
		return x86PackedBinaryForm{}, spec, false, nil
	}
	if ins.Op != "" && !strings.EqualFold(string(ins.Op), string(op)) {
		return x86PackedBinaryForm{}, spec, true, fmt.Errorf("indexed permute dispatch/source opcode mismatch: %s/%s", op, ins.Op)
	}
	local := ins
	local.Op = op
	grammar := x86PackedBinarySpec{encoding: x86PackedLogicalEVEX, zeroing: true}
	if spec.broadcast {
		grammar.broadcastBytes = spec.laneBits / 8
	}
	form, err := x86PackedBinaryProperties(local, arch, grammar)
	return form, spec, true, err
}

func x86ResultIndexedPermuteStep(ins Instr, state *x86ResultState, owned []bool, arch string) (bool, bool) {
	form, spec, known, err := x86IndexedPermuteProperties(ins.Op, ins, arch)
	if !known {
		return false, false
	}
	if err != nil || !x86ResultNamedPredicateInstruction(ins) || x86Mode32VectorRegisterContext(ins, arch) != nil {
		return false, true
	}
	if form.mask >= 0 {
		mask, _ := amd64ParseKReg(ins.Args[form.mask].Reg)
		if !state.masks[mask] {
			return false, true
		}
	}
	width := int64(form.byteWidth)
	firstWidth := width
	if form.memory >= 0 {
		firstWidth = int64(form.memoryBytes)
	}
	if !x86ResultPackedVectorRead(ins.Args[form.first], firstWidth, state, owned, arch) ||
		!x86ResultPackedVectorRead(ins.Args[form.second], width, state, owned, arch) {
		return false, true
	}
	// I2 reads old destination indices; T2 reads its first table, even with
	// zero masking. Single-table operations need old bytes only for merging.
	if (spec.mode != amd64IndexedPermuteSingle || form.mask >= 0 && !form.zeroing) &&
		!x86ResultPackedVectorRead(ins.Args[form.destination], width, state, owned, arch) {
		return false, true
	}
	destination, _ := amd64VectorRegisterIndex(ins.Args[form.destination].Reg, form.byteWidth)
	state.writeVector(destination, form.byteWidth, x86CallerVectorByteMask(width), false)
	return true, true // No FLAGS/DF/mask, memory or pointer facts change.
}

// This is only a numerical vector view. Keep the older scalar/legacy-X read
// domain unchanged; neither native pointers nor wider logical storage follow.
func x86ResultPackedVectorRead(operand Operand, width int64, state *x86ResultState, owned []bool, arch string) bool {
	if operand.Kind != OpReg {
		return x86ResultRead(operand, width, state, owned, arch)
	}
	index, valid := amd64VectorRegisterIndex(operand.Reg, int(width))
	mask := x86CallerVectorByteMask(width)
	return valid && (arch != "386" || index < 8) && mask != 0 && state.vectorBytes(index)&mask == mask
}
