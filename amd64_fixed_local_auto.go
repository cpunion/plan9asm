package plan9asm

import (
	"math"
	"reflect"
	"strings"
)

// NAME_AUTO is a source address role, not an input/storage permission. Go
// obj6 adds deltasp-bpsize to its explicit offset. This bounded observer has
// no CALL/PUSH/POP/ADJSP producer, so the lexical bias is exactly TEXT's local
// size; separately proved SP relocation still contributes its native offset.
// Replay the original operand because parsing constant expressions alone
// loses the difference between tmp-8(SP) and physical -8(SP).
func fixedLocalMemoryAddress(ins Instr, index int, frame int64) (MemRef, error) {
	fail := func() (MemRef, error) {
		return MemRef{}, goAMD64FrameContext("fixed local original memory address role is not closed")
	}
	if index < 0 || index >= len(ins.Args) || ins.Args[index].Kind != OpMem {
		return fail()
	}
	memory := ins.Args[index].Mem
	if memory.Base != SP {
		if memory.OffRaw != "" {
			return fail()
		}
		return memory, nil
	}
	op, rest := splitOpcode(ins.Raw)
	literals := splitTopLevelCSV(rest)
	if !strings.EqualFold(op, string(ins.Op)) || len(literals) != len(ins.Args) {
		return fail()
	}
	parsed, err := parseOperandForArch(ArchAMD64, literals[index])
	if err != nil || !reflect.DeepEqual(parsed, ins.Args[index]) {
		return fail()
	}
	literal := x86ResultAUTOLiteral.FindStringSubmatch(strings.TrimSpace(literals[index]))
	if literal == nil {
		// Indexing/segment/symbol roles are independently excluded by the
		// native observer; none may be relabeled as a named local offset.
		if memory.OffRaw != "" {
			return fail()
		}
		return memory, nil
	}
	prefix := strings.TrimSpace(literal[1])
	offset, named := parseNamedStackConstantOffset(prefix)
	if _, bare := parseIdent(prefix); bare {
		offset, named = 0, true
	}
	if !named {
		if memory.OffRaw != "" {
			return fail()
		}
		return memory, nil // Numeric, unnamed (SP) remains post-prologue SP.
	}
	if offset != memory.Off || frame <= 0 || offset > 0 && frame > math.MaxInt64-offset {
		return fail()
	}
	memory.Off += frame
	memory.OffRaw = ""
	return memory, nil
}

// Descriptive bare NAME_AUTO is offset zero, not an undefined preprocessor
// expression. This syntax query does not grant a native memory certificate.
func x86ConditionalMoveOffsetKnown(memory MemRef) bool {
	if memory.OffRaw == "" {
		return true
	}
	if memory.Base != SP || memory.Sym != "" {
		return false
	}
	if offset, named := parseNamedStackConstantOffset(memory.OffRaw); named {
		return offset == memory.Off
	}
	_, bare := parseIdent(memory.OffRaw)
	return bare && memory.Off == 0
}

func fixedLocalConditionalMoveStep(ins *Instr, state *x86FixedLocalState, proof *x86FixedLocalProof, residue int) (bool, error) {
	effect, known, err := x86TypedConditionalMoveByteEffect(ins.Op, *ins, "amd64")
	if !known {
		return false, nil
	}
	fail := func() (bool, error) {
		return true, goAMD64FrameContext("fixed local CMOV flags or potential numerical source bytes are not defined")
	}
	if err != nil || !state.flags || len(effect.operands) != 2 {
		return fail()
	}
	source, destination := effect.operands[0], effect.operands[1]
	if !fixedLocalNumericRead(destination.operand, destination.width, state) || x86SourceGRegister(destination.operand.Reg) {
		return fail()
	}
	if source.operand.Kind == OpMem {
		if _, err := fixedLocalMemory(ins, source.index, state, proof, residue, source.width, 1, false, false); err != nil {
			return true, err
		}
	} else if !fixedLocalNumericRead(source.operand, source.width, state) {
		return fail() // FP/caller slots and private pointer bits are not imports.
	}
	root, _ := x86CallerGP(destination.operand.Reg)
	mask := x86CallerByteMask(destination.operand.Reg, destination.width)
	if destination.width == 4 {
		mask = 255 // False CMOVL still zero-defines the upper half.
	}
	state.known[root] |= mask
	delete(state.constants, root) // Flags are unchanged, but selected values are not constants.
	return true, nil
}
