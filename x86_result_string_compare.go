package plan9asm

import (
	"fmt"
	"reflect"
	"strings"
)

// These are facts about the full native repeat counter, not numerical bytes
// or an entry convention. Partial writes cannot establish either fact.
type x86ResultStringCount uint8

const (
	x86StringCountUnknown x86ResultStringCount = iota
	x86StringCountZero
	x86StringCountNonzero
)

func x86ResultStringCompareProperties(ins Instr, arch string) (x86StringSpec, bool, error) {
	form, known, err := x86StringInstructionProperties(ins, arch)
	return form, known && (form.kind == "scas" || form.kind == "cmps"), err
}

type x86ResultStringCompareSite struct {
	instruction Instr
	prefix      string
}

type x86ResultStringComparePlan struct {
	prefixes map[int]Instr
	bodies   map[int]x86ResultStringCompareSite
}

// Bind the complete original source before the result proof. The exact same
// CFG verifier runs again at emission; a lexical prefix cannot stand in for
// an incoming edge which bypasses its source node.
func x86ResultStringCompareSourcePlan(fn Func, arch string) (*x86ResultStringComparePlan, error) {
	hasCompare := false
	for _, ins := range fn.Instrs {
		if _, known, _ := x86ResultStringCompareProperties(ins, arch); known {
			hasCompare = true
			break
		}
	}
	if !hasCompare {
		return nil, nil
	}
	fail := func(reason string) error {
		return fmt.Errorf("%w: string comparison needs its original source CFG: %s", ErrProbeNeedsContext, reason)
	}
	if fn.x86DirectTailRoot != nil || fn.x86ContinuationFlow != nil {
		return nil, fail("another source-graph role cannot supply prefix ordinals")
	}
	context := amd64Ctx{goarch: arch, blocks: amd64SplitBlocks(fn), blockByIdx: make(map[int]int)}
	base := 0
	for index, block := range context.blocks {
		context.blockBase = append(context.blockBase, base)
		context.blockByIdx[base] = index
		base += len(block.instrs)
	}
	pairs, err := context.stringRepeatSourceBindings()
	if err != nil {
		return nil, err
	}
	var ordinals []int
	for ordinal, ins := range fn.Instrs {
		if ins.Op != OpLABEL {
			ordinals = append(ordinals, ordinal)
		}
	}
	positions := make([][]int, len(context.blocks))
	plan := &x86ResultStringComparePlan{prefixes: make(map[int]Instr), bodies: make(map[int]x86ResultStringCompareSite)}
	at := 0
	for block, body := range context.blocks {
		for _, ins := range body.instrs {
			if at >= len(ordinals) || !reflect.DeepEqual(ins, fn.Instrs[ordinals[at]]) {
				return nil, fail("split source changed an instruction identity")
			}
			ordinal := ordinals[at]
			positions[block] = append(positions[block], ordinal)
			if _, known, err := x86ResultStringCompareProperties(ins, arch); known {
				if err != nil || !x86ResultNamedPredicateInstruction(ins) {
					return nil, fail("unbound string comparison shape")
				}
				plan.bodies[ordinal] = x86ResultStringCompareSite{instruction: ins}
			}
			at++
		}
	}
	if at != len(ordinals) {
		return nil, fail("incomplete original instruction inventory")
	}
	for _, pair := range pairs {
		prefixOrdinal := positions[pair.prefixBlock][pair.prefixInstruction]
		bodyOrdinal := positions[pair.bodyBlock][pair.bodyInstruction]
		if body, participant := plan.bodies[bodyOrdinal]; participant {
			prefix := fn.Instrs[prefixOrdinal]
			if !x86ResultNamedPredicateInstruction(prefix) {
				return nil, fail("unbound repeat prefix")
			}
			body.prefix = strings.ToUpper(string(prefix.Op))
			plan.prefixes[prefixOrdinal], plan.bodies[bodyOrdinal] = prefix, body
		}
	}
	return plan, nil
}

// This is the numerical continuation of the existing exact volatile helper,
// not a readable region, extent, pointer-to-number conversion or native-entry
// capability. An unsuccessful native access still faults in the paired emitter.
func (plan *x86ResultStringComparePlan) transfer(ordinal int, ins Instr, state *x86ResultState, arch string) (bool, bool) {
	if plan == nil {
		return false, false
	}
	if prefix, bound := plan.prefixes[ordinal]; bound {
		return state != nil && reflect.DeepEqual(prefix, ins), true
	}
	form, known, err := x86ResultStringCompareProperties(ins, arch)
	if !known {
		return false, false
	}
	site, bound := plan.bodies[ordinal]
	if err != nil || state == nil || !bound || !reflect.DeepEqual(site.instruction, ins) || !x86ResultNamedPredicateInstruction(ins) {
		return false, true
	}
	if state.shapeOnly {
		return true, true // Complete shape only; never a reaching input fact.
	}
	repeat := site.prefix != ""
	word := x86ResultPointerWord(arch)
	if repeat {
		if !x86ResultRead(Operand{Kind: OpReg, Reg: CX}, word, state, nil, arch) {
			return false, true
		}
		if state.stringCount == x86StringCountZero {
			return true, true // No read, index update, comparison or DF observation.
		}
	}
	if state.direction == x86ResultDirectionUnknown {
		return false, true
	}
	indexes := []Reg{DI}
	if form.kind == "cmps" {
		indexes = append(indexes, SI)
	} else if !x86ResultRead(Operand{Kind: OpReg, Reg: AX}, int64(form.width), state, nil, arch) {
		return false, true
	}
	for _, index := range indexes {
		if !x86ResultIncomingPointerReadAddressIdentity(Operand{Kind: OpMem, Mem: MemRef{Base: index}}, state, arch) {
			return false, true
		}
	}
	for _, register := range indexes {
		index, _, _ := x86ResultGP(register, word, arch)
		state.gp[index], state.canonicalGP[index] = 0, 0
		state.pointerGP[index], state.callerAddressGP[index], state.callerAddressOrigin[index] = false, false, 0
		state.resultAddressGP[index] = x86ResultSlotAddress{}
		if state.boolFrameGP != nil {
			state.boolFrameGP[index] = x86ResultSlotAddress{}
		}
	}
	if !repeat || state.stringCount == x86StringCountNonzero {
		state.flags = x86ByteFlagsAll
	} // Otherwise the possible zero path retains only the old six must-bits.
	if repeat {
		if !x86ResultStore(Operand{Kind: OpReg, Reg: CX}, word, state, nil, 0, arch) {
			return false, true
		}
		state.canonicalGP[1] = 0
		if arch == "386" {
			state.canonicalGP[1] = 0xf0
		}
		state.callerAddressGP[1], state.callerAddressOrigin[1] = false, 0
		state.resultAddressGP[1] = x86ResultSlotAddress{}
		if state.boolFrameGP != nil {
			state.boolFrameGP[1] = x86ResultSlotAddress{}
		}
		state.stringCount = x86StringCountUnknown
	}
	return true, true
}

// Observe only independently accepted typed transfers. The single native
// counter fact is revoked by any overlapping write, unknown effect or CALL;
// clone/join and the existing shared work budget account for it separately.
func x86ResultStringCountAfter(ins Instr, before, current x86ResultStringCount, arch string) x86ResultStringCount {
	if !x86ResultNamedPredicateInstruction(ins) || ins.Op == OpCALL {
		return x86StringCountUnknown
	}
	if _, known, err := x86ResultStringCompareProperties(ins, arch); known {
		if err == nil {
			return current
		}
		return x86StringCountUnknown
	}
	if (ins.Op == "REP" || ins.Op == "REPN") && len(ins.Args) == 0 {
		return before // Only the independently bound source transfer accepts it.
	}
	ins = x86GoMOVQSourceInstruction(ins)
	word := x86ResultPointerWord(arch)
	if width := x86RawReturnScalarMoveBytes(ins.Op); width != 0 && len(ins.Args) == 2 {
		if ins.Args[1].Kind != OpReg {
			return before
		}
		destination, _, valid := x86ResultGP(ins.Args[1].Reg, width, arch)
		if !valid {
			return x86StringCountUnknown
		}
		if destination != 1 {
			return before
		}
		first := ins.Args[0]
		if width >= word || arch == "amd64" && width == 4 {
			if first.Kind == OpImm && first.ImmRaw == "" && !first.ImmIsFloat {
				value := uint64(first.Imm)
				if width == 4 {
					value = uint64(uint32(value))
				}
				if value == 0 {
					return x86StringCountZero
				}
				return x86StringCountNonzero
			}
		}
		return x86StringCountUnknown
	}
	if form, known, err := x86LockScalarRMWProperties(ins.Op, ins, arch); known && err == nil {
		width := int64(form.spec.bits / 8)
		if form.destination.Kind == OpReg {
			index, _, valid := x86ResultGP(form.destination.Reg, width, arch)
			if valid && index == 1 && (width >= word || arch == "amd64" && width == 4) && x86ScalarXORZeroRegister(form, width) {
				return x86StringCountZero
			}
		}
	}
	if effect, known, err := x86TypedInstructionByteEffect(ins.Op, ins, arch); known {
		if err != nil {
			return x86StringCountUnknown
		}
		for _, operand := range effect.operands {
			if operand.write && operand.operand.Kind == OpReg {
				if index, _, valid := x86ResultGP(operand.operand.Reg, operand.width, arch); !valid || index == 1 {
					return x86StringCountUnknown
				}
			}
		}
		return before
	}
	if _, known, err := x86ScalarCompareProperties(ins.Op, ins, arch); known && err == nil {
		return before
	}
	if _, known, err := x86DirectionControlProperties(ins.Op, ins, arch); known && err == nil {
		return before
	}
	if _, known, err := x86DirectiveProperties(ins); known && err == nil {
		return before
	}
	if ins.Op == OpTEXT || ins.Op == OpLABEL || ins.Op == OpJMP || isAMD64ConditionalBranch(ins.Op) || ins.Op == "NOP" {
		return before
	}
	return x86StringCountUnknown
}
