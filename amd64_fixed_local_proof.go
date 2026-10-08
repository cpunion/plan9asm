package plan9asm

import (
	"fmt"
	"math"
	"reflect"
)

// Local offsets are from the observed post-prologue Go SP, not incoming FP,
// saved BP/return slots, a caller allocation or an arbitrary native address.
type x86FixedLocalProof struct {
	binding  *x86FixedLocalBinding
	frame    int64
	function [32]byte
	accesses map[int]map[int]x86FixedLocalAccess
	ordinal  int // private builder cursor; never trusted by a consumer
}

type x86FixedLocalAccess struct {
	owner   *x86FixedLocalProof
	shape   [32]byte
	width   int64
	write   bool
	offset  [4]int64
	ordinal int
}

type x86FixedLocalState struct {
	pointers  map[Reg]int64
	constants map[Reg]int64
	known     map[Reg]uint8
	vectors   map[Reg]uint64
	bytes     []bool
	addresses map[int64]int64
	results   map[int]bool
	flags     bool
}

func cloneFixedLocalState(input x86FixedLocalState) x86FixedLocalState {
	clone := input
	clone.pointers = make(map[Reg]int64, len(input.pointers))
	clone.constants = make(map[Reg]int64, len(input.constants))
	clone.addresses = make(map[int64]int64, len(input.addresses))
	clone.results = make(map[int]bool, len(input.results))
	for key, value := range input.pointers {
		clone.pointers[key] = value
	}
	for key, value := range input.constants {
		clone.constants[key] = value
	}
	for key, value := range input.addresses {
		clone.addresses[key] = value
	}
	for key, value := range input.results {
		clone.results[key] = value
	}
	clone.known = x86CloneCallerKnown(input.known)
	clone.vectors, _ = x86CloneCallerVectorState(input.vectors, nil)
	clone.bytes = append([]bool(nil), input.bytes...)
	return clone
}

func proveX86FixedLocal(fn Func, sig FuncSig) (Func, error) {
	voidReturn := sig.Ret == Void && len(sig.Frame.Results) == 0
	valueReturn := sig.Ret != "" && sig.Ret != Void && len(sig.Frame.Results) == 1
	if fn.FrameSize <= 0 || fn.FrameSize > 64<<10 || fn.FrameSize%8 != 0 ||
		len(fn.Instrs) == 0 || len(fn.Instrs) > 8192 || fn.FrameSize > (1<<20)/int64(len(fn.Instrs)) ||
		len(fn.X86RawText) != 0 || len(sig.ArgRegs) != 0 || sig.ARMEntry != nil ||
		sig.ARM64GoRegisterABI != nil || sig.x86CallerRegions != nil || !voidReturn && !valueReturn {
		return fn, goAMD64FrameContext("fixed local source has an unbound frame or entry contract")
	}
	copyFn := fn
	copyFn.Instrs = append([]Instr(nil), fn.Instrs...)
	proof := &x86FixedLocalProof{frame: fn.FrameSize, accesses: make(map[int]map[int]x86FixedLocalAccess)}
	for index := range copyFn.Instrs {
		copyFn.Instrs[index].Args = append([]Operand(nil), fn.Instrs[index].Args...)
		copyFn.Instrs[index].x86FixedLocalMemory = nil
		ins := copyFn.Instrs[index]
		if ins.x86Encoded || ins.x86SourceRaw != nil || ins.x86AddressBits != 0 || ins.x86SegmentPrefix != 0 || x86SourceGOperand(ins) {
			return fn, goAMD64FrameContext("fixed local source contains raw/native/g effects")
		}
		if !fixedLocalInstructionShape(ins) {
			return fn, goAMD64FrameContext("fixed local complete source has an unregistered instruction shape: %s", ins.Op)
		}
	}
	for residue := 0; residue < 4; residue++ {
		if err := replayX86FixedLocal(&copyFn, sig, proof, residue); err != nil {
			return fn, err
		}
	}
	// Independent final replay must recreate exactly the same access set and
	// offset in every possible word-aligned original Go SP congruence class.
	final := copyFn
	final.Instrs = append([]Instr(nil), copyFn.Instrs...)
	for index := range final.Instrs {
		final.Instrs[index].x86FixedLocalMemory = nil
	}
	for residue := 0; residue < 4; residue++ {
		if err := replayX86FixedLocal(&final, sig, proof, residue); err != nil {
			return fn, err
		}
	}
	for index := range final.Instrs {
		if !reflect.DeepEqual(final.Instrs[index].x86FixedLocalMemory, copyFn.Instrs[index].x86FixedLocalMemory) {
			return fn, goAMD64FrameContext("fixed local final access replay changed")
		}
		if len(final.Instrs[index].x86FixedLocalMemory) != 0 {
			proof.accesses[index] = make(map[int]x86FixedLocalAccess)
			for operand, access := range final.Instrs[index].x86FixedLocalMemory {
				proof.accesses[index][operand] = access
			}
		}
	}
	if err := fixedLocalRelocationAliases(proof.accesses); err != nil {
		return fn, err
	}
	var err error
	proof.function, err = x86CallerSourceFingerprint(final)
	if err != nil {
		return fn, goAMD64FrameContext("fixed local normalized function cannot be bound: %v", err)
	}
	final.x86FixedLocal = proof
	return final, nil
}

// Every reachable access, including initialization stores and private address
// spills, contributes all its bytes. Each original residue must induce a
// bijection on those bytes: equal bytes remain equal and distinct bytes stay
// distinct. Different distances between permanently disjoint accesses are
// unobservable because the source cannot expose numeric addresses or flags.
func fixedLocalRelocationAliases(accesses map[int]map[int]x86FixedLocalAccess) error {
	var forward, reverse [4]map[int64]int64
	for residue := range forward {
		forward[residue] = make(map[int64]int64)
		reverse[residue] = make(map[int64]int64)
	}
	var total int64
	for _, operands := range accesses {
		for _, access := range operands {
			if access.width <= 0 || access.width > (1<<20)-total {
				return goAMD64FrameContext("fixed local complete access-byte budget exceeded")
			}
			total += access.width
			for byteIndex := int64(0); byteIndex < access.width; byteIndex++ {
				original := access.offset[0] + byteIndex
				for residue := range forward {
					current := access.offset[residue] + byteIndex
					if mapped, found := forward[residue][original]; found && mapped != current {
						return goAMD64FrameContext("fixed local overlapping-byte identity depends on original SP residue")
					}
					if mapped, found := reverse[residue][current]; found && mapped != original {
						return goAMD64FrameContext("fixed local distinct-byte identity depends on original SP residue")
					}
					forward[residue][original] = current
					reverse[residue][current] = original
				}
			}
		}
	}
	return nil
}

func fixedLocalInstructionShape(ins Instr) bool {
	if ins.x86NumericalRead != nil {
		_, valid := x86NumericalReadEffect(ins, "amd64")
		return valid
	}
	for _, operand := range ins.Args {
		if operand.Kind == OpReg && amd64VectorByteWidth(operand.Reg) != 0 {
			index, valid := amd64VectorRegisterIndex(operand.Reg, 16)
			if !valid || index > 15 {
				return false
			}
		}
	}
	if ins.Op == OpTEXT || ins.Op == OpLABEL {
		return true // Parser/complete CFG separately binds TEXT and label operands.
	}
	if ins.Op == OpRET {
		return len(ins.Args) == 0
	}
	if isAMD64ConditionalBranch(ins.Op) || ins.Op == OpJMP {
		_, named := fixedLocalBranchLabel(ins)
		return named
	}
	if x86RawReturnScalarMoveBytes(ins.Op) != 0 {
		return len(ins.Args) == 2
	}
	if ins.Op == "NOTQ" {
		return len(ins.Args) == 1 && ins.Args[0].Kind == OpReg
	}
	if ins.Op == "ANDQ" || ins.Op == "ADDQ" || ins.Op == "CMPQ" {
		return len(ins.Args) == 2
	}
	if _, known, err := x86TypedConditionalMoveByteEffect(ins.Op, ins, "amd64"); known {
		return err == nil
	}
	if form, known, err := x86AESProperties(ins.Op, ins, "amd64"); known {
		return err == nil && form.byteWidth == 16
	}
	if form, known, err := x86PackedWrappingAddProperties(ins, "amd64"); known {
		return err == nil && form.legacy && !form.mmx && form.byteWidth == 16
	}
	if form, known, err := x86WordMultiplyAddProperties(ins, "amd64"); known {
		return err == nil && form.legacy && !form.mmx && form.byteWidth == 16
	}
	if form, known, err := x86FloatCompareProperties(ins, "amd64"); known {
		return err == nil && form.legacy && !form.scalar && form.byteWidth == 16
	}
	return (ins.Op == "MOVOU" || ins.Op == "MOVO") && len(ins.Args) == 2
}

func fixedLocalBranchLabel(ins Instr) (string, bool) {
	if len(ins.Args) != 1 {
		return "", false
	}
	operand := ins.Args[0]
	if operand.Kind == OpLabel && operand.Sym != "" {
		return operand.Sym, true
	}
	if operand.Kind == OpIdent && operand.Ident != "" {
		return operand.Ident, true
	}
	return "", false
}

func fixedLocalEdges(fn Func) (map[int][]int, []int, error) {
	labels := make(map[string]int)
	var instructions []int
	for index, ins := range fn.Instrs {
		if ins.Op == OpLABEL {
			if len(ins.Args) != 1 || ins.Args[0].Kind != OpLabel {
				return nil, nil, goAMD64FrameContext("malformed local label")
			}
			name := ins.Args[0].Sym
			if _, exists := labels[name]; exists {
				return nil, nil, goAMD64FrameContext("duplicate local label")
			}
			labels[name] = len(instructions)
		} else {
			instructions = append(instructions, index)
		}
	}
	if !x86CallerReachableAcyclic(fn, instructions, labels) {
		return nil, nil, goAMD64FrameContext("fixed local CFG is not a closed reachable DAG")
	}
	edges := make(map[int][]int)
	for pc, index := range instructions {
		ins := fn.Instrs[index]
		if ins.Op == OpRET {
			continue
		}
		next := []int{pc + 1}
		if ins.Op == OpJMP || isAMD64ConditionalBranch(ins.Op) {
			label, named := fixedLocalBranchLabel(ins)
			if !named {
				return nil, nil, goAMD64FrameContext("fixed local branch needs a complete named target")
			}
			target, found := labels[label]
			if !found {
				return nil, nil, goAMD64FrameContext("fixed local branch target is missing")
			}
			if ins.Op == OpJMP {
				next = []int{target}
			} else {
				next = append(next, target)
			}
		}
		edges[pc] = next
	}
	return edges, instructions, nil
}

func replayX86FixedLocal(fn *Func, sig FuncSig, proof *x86FixedLocalProof, residue int) error {
	edges, instructions, err := fixedLocalEdges(*fn)
	if err != nil {
		return err
	}
	state := x86FixedLocalState{
		pointers: map[Reg]int64{SP: 0}, constants: make(map[Reg]int64), known: make(map[Reg]uint8),
		vectors: make(map[Reg]uint64), bytes: make([]bool, proof.frame), addresses: make(map[int64]int64), results: make(map[int]bool),
	}
	states := map[int]x86FixedLocalState{0: state}
	// Wait for every reachable predecessor before transferring a block. This
	// topological DAG replay creates no provisional initialization certificate,
	// visits each instruction once, and keeps total byte-state storage bounded.
	reachable := map[int]bool{}
	find := []int{0}
	for len(find) != 0 {
		pc := find[0]
		find = find[1:]
		if reachable[pc] {
			continue
		}
		reachable[pc] = true
		find = append(find, edges[pc]...)
	}
	pending := make(map[int]int)
	for pc := range reachable {
		for _, target := range edges[pc] {
			pending[target]++
		}
	}
	work := []int{0}
	returned := false
	for visits := 0; len(work) != 0; visits++ {
		if visits >= len(instructions) {
			return goAMD64FrameContext("fixed local CFG work bound exceeded")
		}
		pc := work[0]
		work = work[1:]
		if pc < 0 || pc >= len(instructions) {
			return goAMD64FrameContext("fixed local fallthrough lacks a source return")
		}
		current := cloneFixedLocalState(states[pc])
		ins := &fn.Instrs[instructions[pc]]
		proof.ordinal = instructions[pc]
		if err := fixedLocalStep(ins, &current, sig, proof, residue); err != nil {
			return fmt.Errorf("%s: %q: %w", fn.Sym, ins.Raw, err)
		}
		if ins.Op == OpRET {
			returned = true
			continue
		}
		for _, target := range edges[pc] {
			if previous, found := states[target]; found {
				// Acyclic replay permits only all-path initialization to decrease.
				// Address identity and saved private pointer cells remain exact.
				if !reflect.DeepEqual(previous.pointers, current.pointers) || !reflect.DeepEqual(previous.addresses, current.addresses) ||
					!reflect.DeepEqual(previous.constants, current.constants) || previous.flags != current.flags {
					return goAMD64FrameContext("fixed local CFG address/constant/flags identity disagrees")
				}
				merged := cloneFixedLocalState(previous)
				for reg, bits := range merged.known {
					merged.known[reg] = bits & current.known[reg]
				}
				for reg, bits := range merged.vectors {
					merged.vectors[reg] = bits & current.vectors[reg]
				}
				for index := range merged.bytes {
					merged.bytes[index] = merged.bytes[index] && current.bytes[index]
				}
				for index, written := range merged.results {
					merged.results[index] = written && current.results[index]
				}
				states[target] = merged
			} else {
				states[target] = cloneFixedLocalState(current)
			}
			pending[target]--
			if pending[target] == 0 {
				work = append(work, target)
			}
		}
	}
	if !returned {
		return goAMD64FrameContext("fixed local source has no reachable return")
	}
	return nil
}

func fixedLocalMemory(ins *Instr, index int, state *x86FixedLocalState, proof *x86FixedLocalProof, residue int, width, alignment int64, write, pointer bool) (int64, error) {
	if index < 0 || index >= len(ins.Args) || ins.Args[index].Kind != OpMem || width <= 0 {
		return 0, goAMD64FrameContext("fixed local access lacks native memory")
	}
	memory, err := fixedLocalMemoryAddress(*ins, index, proof.frame)
	if err != nil {
		return 0, err
	}
	base, found := state.pointers[memory.Base]
	if !found || memory.Sym != "" || memory.Segment != "" || memory.Index != "" ||
		memory.Off > 0 && base > math.MaxInt64-memory.Off || memory.Off < 0 && base < math.MinInt64-memory.Off {
		return 0, goAMD64FrameContext("fixed local access has unknown/escaping address provenance")
	}
	offset := base + memory.Off
	if offset < 0 || offset > proof.frame || width > proof.frame-offset || alignment <= 0 ||
		(int64(residue*8)+offset)%alignment != 0 {
		return 0, goAMD64FrameContext("fixed local access crosses storage or has residue-dependent native alignment")
	}
	if !write && !pointer {
		for byteIndex := offset; byteIndex < offset+width; byteIndex++ {
			if !state.bytes[byteIndex] {
				return 0, goAMD64FrameContext("fixed local native read contains uninitialized or address bytes")
			}
		}
	}
	if ins.x86FixedLocalMemory == nil {
		ins.x86FixedLocalMemory = make(map[int]x86FixedLocalAccess)
	}
	shape, err := fixedLocalInstructionDigest(*ins)
	if err != nil {
		return 0, err
	}
	access, found := ins.x86FixedLocalMemory[index]
	if !found {
		access = x86FixedLocalAccess{owner: proof, shape: shape, width: width, write: write, ordinal: proof.ordinal}
	}
	if access.owner != proof || access.shape != shape || access.width != width || access.write != write {
		return 0, goAMD64FrameContext("fixed local access shape changed")
	}
	access.offset[residue] = offset
	ins.x86FixedLocalMemory[index] = access
	return offset, nil
}

func fixedLocalStep(ins *Instr, state *x86FixedLocalState, sig FuncSig, proof *x86FixedLocalProof, residue int) error {
	fail := func() error {
		return goAMD64FrameContext("unproved fixed local width/initialization/address/control effect: %s", ins.Op)
	}
	if ins.Op == OpTEXT {
		return nil
	}
	if ins.Op == OpRET {
		sp, found := state.pointers[SP]
		if len(ins.Args) != 0 || !found || sp != 0 || sig.Ret != Void && !state.results[0] {
			return fail()
		}
		return nil
	}
	if isAMD64ConditionalBranch(ins.Op) {
		if !state.flags || isAMD64CounterZeroBranch(ins.Op) {
			return fail()
		}
		return nil
	}
	if ins.Op == OpJMP {
		return nil
	}
	if known, err := fixedLocalNumericalReadStep(*ins, state); known {
		return err
	}
	if width := x86RawReturnScalarMoveBytes(ins.Op); width != 0 {
		return fixedLocalMove(ins, state, sig, proof, residue, width)
	}
	if ins.Op == "NOTQ" && len(ins.Args) == 1 && ins.Args[0].Kind == OpReg {
		reg := ins.Args[0].Reg
		if _, pointer := state.pointers[reg]; pointer {
			return fail()
		}
		value, known := state.constants[reg]
		if !known {
			return fail()
		}
		state.constants[reg] = ^value
		return nil
	}
	if (ins.Op == "ADDQ" || ins.Op == "ANDQ") && len(ins.Args) == 2 && ins.Args[1].Kind == OpReg {
		reg := ins.Args[1].Reg
		value := ins.Args[0].Imm
		known := ins.Args[0].Kind == OpImm && ins.Args[0].ImmRaw == "" && !ins.Args[0].ImmIsFloat
		if ins.Args[0].Kind == OpReg {
			value, known = state.constants[ins.Args[0].Reg]
		}
		if !known {
			return fail()
		}
		if offset, pointer := state.pointers[reg]; pointer {
			if ins.Op == "ANDQ" {
				// Only an exact power-of-two residue projection preserves the
				// same private address origin. No pointer bit becomes numeric.
				if value != -8 && value != -16 && value != -32 {
					return fail()
				}
				offset = (offset+int64(residue*8))&value - int64(residue*8)
			} else {
				if value > 0 && offset > math.MaxInt64-value || value < 0 && offset < math.MinInt64-value {
					return fail()
				}
				offset += value
			}
			if offset < -32 || offset > proof.frame {
				return fail()
			}
			state.pointers[reg] = offset
			state.flags = false // Relocated pointer arithmetic flags are opaque.
			return nil
		}
		return fail() // Ordinary numeric arithmetic needs its complete family.
	}
	if ins.Op == "CMPQ" && len(ins.Args) == 2 {
		for _, operand := range ins.Args {
			if !fixedLocalNumericRead(operand, 8, state) {
				return fail()
			}
		}
		state.flags = true
		return nil
	}
	if known, err := fixedLocalConditionalMoveStep(ins, state, proof, residue); known {
		return err
	}
	if known, err := fixedLocalAESByteStep(ins, state, proof, residue); known {
		return err
	}
	return fixedLocalVectorStep(ins, state, proof, residue)
}

func fixedLocalNumericRead(operand Operand, width int64, state *x86FixedLocalState) bool {
	if operand.Kind == OpImm {
		return operand.ImmRaw == "" && !operand.ImmIsFloat
	}
	if operand.Kind != OpReg {
		return false
	}
	root, valid := x86CallerGP(operand.Reg)
	_, pointer := state.pointers[root]
	mask := x86CallerByteMask(operand.Reg, width)
	return valid && root != SP && root != BP && !pointer && mask != 0 && state.known[root]&mask == mask
}

func fixedLocalMove(ins *Instr, state *x86FixedLocalState, sig FuncSig, proof *x86FixedLocalProof, residue int, width int64) error {
	fail := func() error { return goAMD64FrameContext("unproved fixed local scalar move") }
	if len(ins.Args) != 2 {
		return fail()
	}
	source, destination := ins.Args[0], ins.Args[1]
	var address int64
	var pointer bool
	var constant int64
	var constantKnown bool
	switch source.Kind {
	case OpReg:
		address, pointer = state.pointers[source.Reg]
		if pointer {
			if width != 8 {
				return fail()
			}
		} else if amd64VectorByteWidth(source.Reg) == 16 {
			if width != 8 || state.vectors[x86CallerVectorRoot(source.Reg)]&255 != 255 {
				return fail()
			}
		} else {
			if !fixedLocalNumericRead(source, width, state) {
				return fail()
			}
			constant, constantKnown = state.constants[source.Reg]
			// Narrow moves truncate or preserve a destination's old upper
			// bytes. They cannot transport a complete 64-bit address mask.
			constantKnown = constantKnown && width == 8
		}
	case OpImm:
		if !fixedLocalNumericRead(source, width, state) {
			return fail()
		}
		constant, constantKnown = source.Imm, width == 8
	case OpFP:
		sp, original := state.pointers[SP]
		if !original || sp != 0 || !fixedLocalFPBound(source.FPOffset, width, sig.Frame.Params) {
			return fail()
		}
	case OpMem:
		memory, err := fixedLocalMemoryAddress(*ins, 0, proof.frame)
		if err != nil {
			return err
		}
		base, found := state.pointers[memory.Base]
		if found && memory.Off >= -proof.frame && memory.Off <= proof.frame {
			address, pointer = state.addresses[base+memory.Off]
		}
		if pointer && width != 8 {
			return fail()
		}
		if _, err := fixedLocalMemory(ins, 0, state, proof, residue, width, 1, false, pointer); err != nil {
			return err
		}
	default:
		return fail()
	}
	switch destination.Kind {
	case OpReg:
		root, valid := x86CallerGP(destination.Reg)
		if !valid || root == BP || x86SourceGRegister(root) || pointer && (width != 8 || destination.Reg != root) {
			return fail()
		}
		if root == SP && !pointer {
			return fail()
		}
		if _, oldPointer := state.pointers[root]; oldPointer && width != 8 {
			return fail()
		}
		delete(state.pointers, root)
		delete(state.constants, root)
		if pointer {
			state.pointers[root] = address
			state.known[root] = 0
		} else {
			mask := x86CallerByteMask(destination.Reg, width)
			if mask == 0 {
				return fail()
			}
			if width == 4 {
				mask = 255
			}
			state.known[root] |= mask
			if constantKnown {
				state.constants[root] = constant
			}
		}
	case OpMem:
		if source.Kind == OpMem {
			return fail()
		}
		offset, err := fixedLocalMemory(ins, 1, state, proof, residue, width, 1, true, false)
		if err != nil {
			return err
		}
		for start := range state.addresses {
			if offset < start+8 && start < offset+width {
				delete(state.addresses, start)
			}
		}
		for index := offset; index < offset+width; index++ {
			state.bytes[index] = !pointer
		}
		if pointer {
			state.addresses[offset] = address
		}
	case OpFP:
		sp, original := state.pointers[SP]
		if !original || sp != 0 || pointer || !fixedLocalFPBound(destination.FPOffset, width, sig.Frame.Results) || source.Kind != OpReg {
			return fail()
		}
		state.results[0] = true
	default:
		return fail()
	}
	return nil
}

func fixedLocalFPBound(offset, width int64, slots []FrameSlot) bool {
	for _, slot := range slots {
		if len(frameSlotFields(slot)) == 0 && slot.Offset == offset && frameTypeSize(slot.Type, 8) == width {
			return true
		}
	}
	return false
}

func fixedLocalVectorStep(ins *Instr, state *x86FixedLocalState, proof *x86FixedLocalProof, residue int) error {
	width := uint64(65535)
	first, second, destination, memory := 0, 1, 1, -1
	if form, known, err := x86PackedWrappingAddProperties(*ins, "amd64"); known {
		if err != nil || !form.legacy || form.mmx || form.byteWidth != 16 {
			return goAMD64FrameContext("fixed local packed add shape is not closed")
		}
		first, second, destination, memory = form.first, form.second, form.destination, form.memory
	} else if form, known, err := x86WordMultiplyAddProperties(*ins, "amd64"); known {
		if err != nil || !form.legacy || form.mmx || form.byteWidth != 16 {
			return goAMD64FrameContext("fixed local multiply-add shape is not closed")
		}
		first, second, destination, memory = form.first, form.second, form.destination, form.memory
	} else if form, known, err := x86FloatCompareProperties(*ins, "amd64"); known {
		if err != nil || !form.legacy || form.scalar || form.byteWidth != 16 {
			return goAMD64FrameContext("fixed local compare shape is not closed")
		}
		first, second, destination, memory = form.first, form.second, form.destination, form.memory
	} else if ins.Op == "MOVOU" || ins.Op == "MOVO" {
		if len(ins.Args) != 2 {
			return goAMD64FrameContext("fixed local vector copy shape is not closed")
		}
		alignment := int64(1)
		if ins.Op == "MOVO" {
			alignment = 16
		}
		if ins.Args[0].Kind == OpMem {
			if amd64VectorByteWidth(ins.Args[1].Reg) != 16 || ins.Args[1].Kind != OpReg {
				return goAMD64FrameContext("fixed local vector destination is not X")
			}
			if _, err := fixedLocalMemory(ins, 0, state, proof, residue, 16, alignment, false, false); err != nil {
				return err
			}
			state.vectors[x86CallerVectorRoot(ins.Args[1].Reg)] |= width
			return nil
		}
		if ins.Args[0].Kind != OpReg || amd64VectorByteWidth(ins.Args[0].Reg) != 16 || state.vectors[x86CallerVectorRoot(ins.Args[0].Reg)]&width != width {
			return goAMD64FrameContext("fixed local vector copy source is incomplete")
		}
		if ins.Args[1].Kind == OpMem {
			offset, err := fixedLocalMemory(ins, 1, state, proof, residue, 16, alignment, true, false)
			if err != nil {
				return err
			}
			for start := range state.addresses {
				if offset < start+8 && start < offset+16 {
					delete(state.addresses, start)
				}
			}
			for index := offset; index < offset+16; index++ {
				state.bytes[index] = true
			}
			return nil
		}
		if ins.Args[1].Kind != OpReg || amd64VectorByteWidth(ins.Args[1].Reg) != 16 {
			return goAMD64FrameContext("fixed local vector copy destination is not X")
		}
		state.vectors[x86CallerVectorRoot(ins.Args[1].Reg)] |= width
		return nil
	} else {
		return goAMD64FrameContext("fixed local source has an unregistered typed vector effect: %s", ins.Op)
	}
	if state.vectors[x86CallerVectorRoot(ins.Args[second].Reg)]&width != width {
		return goAMD64FrameContext("fixed local vector operation reads undefined destination bytes")
	}
	if memory >= 0 {
		if _, err := fixedLocalMemory(ins, memory, state, proof, residue, 16, 16, false, false); err != nil {
			return err
		}
	} else if state.vectors[x86CallerVectorRoot(ins.Args[first].Reg)]&width != width {
		return goAMD64FrameContext("fixed local vector operation reads undefined source bytes")
	}
	state.vectors[x86CallerVectorRoot(ins.Args[destination].Reg)] |= width
	return nil
}

func (c *amd64Ctx) fixedLocalMemoryCertified(ins Instr, index int) bool {
	access, found := ins.x86FixedLocalMemory[index]
	shape, err := fixedLocalInstructionDigest(ins)
	if !found || c.fixedLocal == nil || access.owner != c.fixedLocal || ins.x86Encoded ||
		ins.x86SourceRaw != nil || ins.x86AddressBits != 0 || ins.x86SegmentPrefix != 0 ||
		err != nil || access.shape != shape || access.width <= 0 {
		return false
	}
	expected, bound := c.fixedLocal.accesses[access.ordinal][index]
	return bound && reflect.DeepEqual(access, expected)
}
