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
		// Indexing/segment/symbol roles are independently validated by the
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

// This query resolves only a complete numerical constant index. It neither
// changes the emitted address nor supplies a memory origin, extent or access.
// All arithmetic is checked before an original private/literal bound is used.
func fixedLocalConstantDisplacement(memory MemRef, state *x86FixedLocalState) (int64, bool) {
	if memory.OffRaw != "" || memory.Sym != "" || memory.Segment != "" || memory.IndexExt != "" {
		return 0, false
	}
	if memory.Index == "" {
		return memory.Off, memory.Scale == 0
	}
	root, valid := x86CallerGP(memory.Index)
	value, known := state.constants[root]
	scale := memory.Scale
	if !valid || root != memory.Index || !known ||
		!fixedLocalNumericRead(Operand{Kind: OpReg, Reg: root}, 8, state) ||
		scale != 1 && scale != 2 && scale != 4 && scale != 8 {
		return 0, false
	}
	if value > math.MaxInt64/scale || value < math.MinInt64/scale {
		return 0, false
	}
	delta := value * scale
	if delta > 0 && memory.Off > math.MaxInt64-delta || delta < 0 && memory.Off < math.MinInt64-delta {
		return 0, false
	}
	return memory.Off + delta, true
}

// Reads, writes and complete pointer-spill lookups use the same resolved
// byte address. An indexed read must not borrow the unindexed slot's role.
func fixedLocalPrivateOffset(memory MemRef, state *x86FixedLocalState) (int64, bool) {
	base, private := state.pointers[memory.Base]
	displacement, known := fixedLocalConstantDisplacement(memory, state)
	if !private || !known || displacement > 0 && base > math.MaxInt64-displacement ||
		displacement < 0 && base < math.MinInt64-displacement {
		return 0, false
	}
	return base + displacement, true
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
	if err != nil || state.flags&effect.flagsRead != effect.flagsRead || len(effect.operands) != 2 {
		return fail()
	}
	source, destination := effect.operands[0], effect.operands[1]
	if !fixedLocalNumericRead(destination.operand, destination.width, state) {
		return fail()
	}
	role := x86FixedLocalPrivateMemory
	if source.operand.Kind == OpMem || source.operand.Kind == OpSym {
		var err error
		_, role, err = fixedLocalReadMemory(ins, source.index, state, proof, residue, source.width, 1)
		if err != nil {
			return true, err
		}
	} else if !fixedLocalNumericRead(source.operand, source.width, state) {
		return fail() // FP/caller slots and private pointer bits are not imports.
	}
	root, _ := x86CallerGP(destination.operand.Reg)
	booleans := fixedLocalBooleanRead(destination.operand, destination.width, state)
	if source.operand.Kind == OpReg {
		booleans &= fixedLocalBooleanRead(source.operand, source.width, state)
	} else if role != x86FixedLocalPrivateMemory {
		booleans = 0 // Numerical input/literal bytes are not a boolean proof.
	} else {
		// The source read was proved above; its actual private offset is
		// retained in this same residue's instruction-bound certificate.
		access := ins.x86FixedLocalMemory[source.index]
		booleans &= fixedLocalMemoryBooleans(state, access.offset[residue], source.width)
	}
	fixedLocalWriteBooleans(state, destination.operand.Reg, destination.width, booleans)
	mask := x86CallerByteMask(destination.operand.Reg, destination.width)
	if destination.width == 4 {
		mask = 255 // False CMOVL still zero-defines the upper half.
	}
	state.known[root] |= mask
	delete(state.constants, root) // Flags are unchanged, but selected values are not constants.
	return true, nil
}
