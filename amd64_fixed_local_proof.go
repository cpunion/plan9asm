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
	owner       *x86FixedLocalProof
	shape       [32]byte
	width       int64
	write       bool
	nativeInput bool // Typed source address, not a private extent/alignment proof.
	offset      [8]int64
	ordinal     int
}

type x86FixedLocalState struct {
	pointers     map[Reg]int64
	opaque       map[Reg]x86FixedLocalOrigin
	opaqueCells  map[int64]x86FixedLocalOrigin
	booleans     map[Reg]uint8
	booleanBytes []bool
	constants    map[Reg]int64
	known        map[Reg]uint8
	vectors      map[Reg]uint64
	masks        map[Reg]bool
	bytes        []bool
	addresses    map[int64]int64
	results      map[int]bool
	flags        uint8
}

func cloneFixedLocalState(input x86FixedLocalState) x86FixedLocalState {
	clone := input
	clone.pointers = make(map[Reg]int64, len(input.pointers))
	clone.opaque = make(map[Reg]x86FixedLocalOrigin, len(input.opaque))
	clone.opaqueCells = make(map[int64]x86FixedLocalOrigin, len(input.opaqueCells))
	clone.booleans = x86CloneCallerKnown(input.booleans)
	clone.booleanBytes = append([]bool(nil), input.booleanBytes...)
	clone.constants = make(map[Reg]int64, len(input.constants))
	clone.addresses = make(map[int64]int64, len(input.addresses))
	clone.results = make(map[int]bool, len(input.results))
	for key, value := range input.pointers {
		clone.pointers[key] = value
	}
	for key, value := range input.opaque {
		clone.opaque[key] = value
	}
	for key, value := range input.opaqueCells {
		clone.opaqueCells[key] = value
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
	clone.vectors, clone.masks = x86CloneCallerVectorState(input.vectors, input.masks)
	clone.bytes = append([]bool(nil), input.bytes...)
	return clone
}

func proveX86FixedLocal(fn Func, sig FuncSig) (Func, error) {
	voidReturn := sig.Ret == Void && len(sig.Frame.Results) == 0
	valueReturn := sig.Ret != "" && sig.Ret != Void && len(sig.Frame.Results) > 0 && len(sig.Frame.Results) <= 4096
	if fn.FrameSize <= 0 || fn.FrameSize > 64<<10 || fn.FrameSize%8 != 0 ||
		len(fn.Instrs) == 0 || len(fn.Instrs) > 8192 || fn.FrameSize > (1<<20)/int64(len(fn.Instrs)) ||
		len(fn.X86RawText) != 0 || len(sig.ArgRegs) != 0 || sig.ARMEntry != nil ||
		sig.ARM64GoRegisterABI != nil || sig.x86CallerRegions != nil || !voidReturn && !valueReturn {
		return fn, goAMD64FrameContext("fixed local source has an unbound frame or entry contract")
	}
	copyFn := fn
	copyFn.Instrs = append([]Instr(nil), fn.Instrs...)
	proof := &x86FixedLocalProof{binding: sig.x86FixedLocal, frame: fn.FrameSize, accesses: make(map[int]map[int]x86FixedLocalAccess)}
	budget := x86FixedLocalBudget{steps: 1 << 18, bytes: 64 << 20}
	for index := range copyFn.Instrs {
		copyFn.Instrs[index].Args = append([]Operand(nil), fn.Instrs[index].Args...)
		copyFn.Instrs[index].x86FixedLocalMemory = nil
	}
	effects, err := fixedLocalEffectPlan(copyFn)
	if err != nil {
		return fn, err
	}
	for residue := 0; residue < 8; residue++ {
		if err := replayX86FixedLocal(&copyFn, sig, proof, residue, &budget, effects); err != nil {
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
	effects, err = fixedLocalEffectPlan(final)
	if err != nil {
		return fn, err
	}
	for residue := 0; residue < 8; residue++ {
		if err := replayX86FixedLocal(&final, sig, proof, residue, &budget, effects); err != nil {
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
	proof.function, err = x86CallerSourceFingerprint(final)
	if err != nil {
		return fn, goAMD64FrameContext("fixed local normalized function cannot be bound: %v", err)
	}
	final.x86FixedLocal = proof
	return final, nil
}

// This call-local plan holds immutable syntax effects, never a proof result.
// Original ordinals bind it to the private function copy. Each independent
// eight-residue replay builds its own plan, avoiding repeated family parsing
// in a loop without caching any input, storage or source authority.
func fixedLocalEffectPlan(fn Func) (map[int]x86TypedByteEffect, error) {
	var effects map[int]x86TypedByteEffect
	for index, ins := range fn.Instrs {
		if ins.x86Encoded || ins.x86SourceRaw != nil || ins.x86AddressBits != 0 || ins.x86SegmentPrefix != 0 {
			return nil, goAMD64FrameContext("fixed local source contains raw/native effects")
		}
		if effect, known, err := fixedLocalVectorByteEffect(ins); known {
			if err != nil {
				return nil, err
			}
			if effects == nil {
				effects = make(map[int]x86TypedByteEffect)
			}
			effects[index] = effect
		} else if !fixedLocalInstructionShape(ins) {
			return nil, goAMD64FrameContext("fixed local complete source has an unregistered instruction shape: %s", ins.Op)
		}
	}
	return effects, nil
}

// Every reachable access, including initialization stores and private address
// spills, contributes all its bytes. Each original residue must induce a
// bijection on those bytes: equal bytes remain equal and distinct bytes stay
// distinct. Different distances between permanently disjoint accesses are
// unobservable because the source cannot expose numeric addresses or flags.
func fixedLocalRelocationAliases(accesses map[int]map[int]x86FixedLocalAccess) error {
	var forward, reverse [8]map[int64]int64
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
			if access.nativeInput {
				continue // Native input addresses are never relocated.
			}
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
	if form, known, err := fixedLocalConversionProperties(ins); known {
		return err == nil && !form.mmx
	}
	if _, known, err := x86ConditionalSetProperties(ins.Op, ins, "amd64"); known {
		return err == nil
	}
	if ins.x86NumericalRead != nil {
		_, valid := x86NumericalReadEffect(ins, "amd64")
		return valid
	}
	if _, known, err := amd64PackedVectorMoveProperties(ins, "amd64"); known {
		return err == nil
	}
	if form, known, err := x86NonTemporalVectorProperties(ins.Op, ins, "amd64"); known {
		return err == nil && !form.trap
	}
	if _, known, err := x86OrderingHintProperties(ins.Op, ins, "amd64"); known {
		return err == nil
	}
	if _, known, err := x86VectorZeroProperties(ins, "amd64"); known {
		return err == nil
	}
	if _, known, err := x86LegacyPackedMoveProperties(ins.Op, ins, "amd64"); known {
		return err == nil
	}
	if _, known := amd64MaskMoveSpecs[ins.Op]; known {
		return amd64MaskMoveOperandsValid(ins, "amd64")
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
	if x86RawReturnScalarMoveBytes(x86GoMOVQSourceInstruction(ins).Op) != 0 {
		return len(ins.Args) == 2
	}
	if _, address := fixedLocalStaticAddress(ins); address {
		return true
	}
	if effect, known, err := fixedLocalGPByteEffect(ins); known {
		return err == nil && !effect.dynamicMemory
	}
	if form, known, err := x86AESProperties(ins.Op, ins, "amd64"); known {
		return err == nil && form.byteWidth == 16
	}
	if _, known, err := fixedLocalVectorByteEffect(ins); known {
		return err == nil
	}
	if form, known, err := x86FloatCompareProperties(ins, "amd64"); known {
		return err == nil && form.legacy && !form.scalar && form.byteWidth == 16
	}
	return false
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
	if len(instructions) == 0 || len(instructions) > 8192 {
		return nil, nil, goAMD64FrameContext("fixed local CFG lacks a bounded instruction inventory")
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
			if !found || target >= len(instructions) {
				return nil, nil, goAMD64FrameContext("fixed local branch target is missing an instruction")
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

// All congruence classes and the independent final replay share these limits.
// A cyclic source cannot obtain an unbounded analysis or a partial certificate.
type x86FixedLocalBudget struct {
	steps int
	bytes int64
}

func (budget *x86FixedLocalBudget) consume(frame int64) error {
	if budget.steps <= 0 || frame > budget.bytes {
		return goAMD64FrameContext("fixed local CFG work/byte bound exceeded")
	}
	budget.steps--
	budget.bytes -= frame
	return nil
}

func replayX86FixedLocal(fn *Func, sig FuncSig, proof *x86FixedLocalProof, residue int, budget *x86FixedLocalBudget,
	effects map[int]x86TypedByteEffect) error {
	edges, instructions, err := fixedLocalEdges(*fn)
	if err != nil {
		return err
	}
	state := x86FixedLocalState{
		pointers: map[Reg]int64{SP: 0}, constants: make(map[Reg]int64), known: make(map[Reg]uint8),
		opaque: make(map[Reg]x86FixedLocalOrigin), opaqueCells: make(map[int64]x86FixedLocalOrigin), booleans: make(map[Reg]uint8),
		booleanBytes: make([]bool, proof.frame),
		vectors:      make(map[Reg]uint64), masks: make(map[Reg]bool),
		bytes: make([]bool, proof.frame), addresses: make(map[int64]int64), results: make(map[int]bool),
	}
	// All GP entry bytes except the private SP origin remain unknown. ABI0
	// does not supply an initialized R14/g value; the observed nonzero-frame
	// prologue/epilogue independently preserves BP. Explicit definitions may
	// use either register without importing its entry value or native role.
	states := map[int]x86FixedLocalState{0: state}
	work := []int{0}
	queued := map[int]bool{0: true}
	accesses := make(map[int]map[int]x86FixedLocalAccess)
	returned := false
	for head := 0; head < len(work); head++ {
		if err := budget.consume(proof.frame); err != nil {
			return err
		}
		pc := work[head]
		queued[pc] = false
		if pc < 0 || pc >= len(instructions) {
			return goAMD64FrameContext("fixed local fallthrough lacks a source return")
		}
		current := cloneFixedLocalState(states[pc])
		// Transfer a private instruction copy: provisional loop states must
		// never publish accesses, including certificates from another residue.
		ins := fn.Instrs[instructions[pc]]
		ins.x86FixedLocalMemory = nil
		proof.ordinal = instructions[pc]
		if effect, found := effects[instructions[pc]]; found {
			err = fixedLocalTypedByteStep(&ins, &current, proof, residue, effect)
		} else {
			err = fixedLocalStep(&ins, &current, sig, proof, residue)
		}
		if err != nil {
			return fmt.Errorf("%s: %q: %w", fn.Sym, ins.Raw, err)
		}
		accesses[pc] = ins.x86FixedLocalMemory
		returned = returned || ins.Op == OpRET
		for _, target := range edges[pc] {
			if target < 0 || target >= len(instructions) {
				return goAMD64FrameContext("fixed local fallthrough lacks a source return")
			}
			changed := true
			if previous, found := states[target]; found {
				changed, err = mergeFixedLocalState(&previous, current)
				if err != nil {
					return err
				}
				states[target] = previous
			} else {
				states[target] = cloneFixedLocalState(current)
			}
			if changed && !queued[target] {
				queued[target] = true
				work = append(work, target)
			}
		}
	}
	if !returned {
		return goAMD64FrameContext("fixed local source has no reachable return")
	}
	// An empty worklist means each retained transfer consumed the final input
	// state: every changed meet requeued its consumer. Publish those local
	// annotations only now, without repeating every instruction in a DAG.
	// Entry state remains a predecessor of PC zero, including backedges.
	for pc, index := range instructions {
		current, reachable := accesses[pc]
		if !reachable {
			continue
		}
		previous := fn.Instrs[index].x86FixedLocalMemory
		if residue != 0 && len(previous) != len(current) {
			return goAMD64FrameContext("fixed local access inventory depends on original SP residue")
		}
		for operand, access := range current {
			if residue == 0 {
				continue
			}
			prior, found := previous[operand]
			if !found || prior.owner != access.owner || prior.shape != access.shape || prior.width != access.width ||
				prior.write != access.write || prior.nativeInput != access.nativeInput || prior.ordinal != access.ordinal {
				return goAMD64FrameContext("fixed local access shape depends on original SP residue")
			}
			prior.offset[residue] = access.offset[residue]
			current[operand] = prior
		}
		fn.Instrs[index].x86FixedLocalMemory = current
	}
	return nil
}

func mergeFixedLocalState(previous *x86FixedLocalState, current x86FixedLocalState) (bool, error) {
	// Private relocation identities must agree. Incoming origins are a separate
	// must-domain: losing one leaves unknown bytes, never numerical knowledge.
	if !reflect.DeepEqual(previous.pointers, current.pointers) || !reflect.DeepEqual(previous.addresses, current.addresses) {
		return false, goAMD64FrameContext("fixed local CFG address identity disagrees")
	}
	changed := intersectFixedLocalOrigins(previous.constants, current.constants)
	changed = intersectFixedLocalOrigins(previous.opaque, current.opaque) || changed
	changed = intersectFixedLocalOrigins(previous.opaqueCells, current.opaqueCells) || changed
	flags := previous.flags & current.flags
	changed = changed || flags != previous.flags
	previous.flags = flags
	for reg, bits := range previous.known {
		value := bits & current.known[reg]
		changed = changed || value != bits
		previous.known[reg] = value
	}
	for reg, bits := range previous.booleans {
		value := bits & current.booleans[reg]
		changed = changed || value != bits
		previous.booleans[reg] = value
	}
	for reg, bits := range previous.vectors {
		value := bits & current.vectors[reg]
		changed = changed || value != bits
		previous.vectors[reg] = value
	}
	for reg, known := range previous.masks {
		value := known && current.masks[reg]
		changed = changed || value != known
		previous.masks[reg] = value
	}
	for index, known := range previous.bytes {
		value := known && current.bytes[index]
		boolean := previous.booleanBytes[index] && current.booleanBytes[index]
		changed = changed || value != known || boolean != previous.booleanBytes[index]
		previous.bytes[index], previous.booleanBytes[index] = value, boolean
	}
	for index, written := range previous.results {
		value := written && current.results[index]
		changed = changed || value != written
		previous.results[index] = value
	}
	return changed, nil
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
	return offset, fixedLocalRecordAccess(ins, index, proof, residue, width, write, false, offset)
}

// Native input addresses retain ordinary faulting memory semantics. Their Go
// pointer type proves only provenance, never a pointee extent or alignment.
func fixedLocalMemoryAccess(ins *Instr, index int, state *x86FixedLocalState, proof *x86FixedLocalProof,
	residue int, width, alignment int64, write, pointer bool) (int64, x86FixedLocalMemoryRole, error) {
	if index >= 0 && index < len(ins.Args) && ins.Args[index].Kind == OpMem {
		memory := ins.Args[index].Mem
		if origin, input := state.opaque[memory.Base]; input {
			root, valid := x86CallerGP(memory.Base)
			if !valid || root != memory.Base || root == SP || width <= 0 || alignment <= 0 || pointer || write && origin.global != "" ||
				memory.OffRaw != "" || memory.Sym != "" || memory.Segment != "" ||
				memory.Index != "" || memory.Scale != 0 || memory.IndexExt != "" {
				return 0, x86FixedLocalNativeInput, goAMD64FrameContext("fixed local native input requires a plain typed base and displacement")
			}
			if origin.literal.symbol != "" {
				shifted, valid := fixedLocalShiftLiteral(origin, memory.Off, proof)
				if write || !valid || !fixedLocalLiteralRange(proof.binding, shifted.literal, width, alignment) {
					return 0, x86FixedLocalLiteralMemory, goAMD64FrameContext("fixed local indirect DATA read exceeds readonly identity, extent or alignment")
				}
			}
			return 0, x86FixedLocalNativeInput, fixedLocalRecordAccess(ins, index, proof, residue, width, write, true, 0)
		}
	}
	offset, err := fixedLocalMemory(ins, index, state, proof, residue, width, alignment, write, pointer)
	return offset, x86FixedLocalPrivateMemory, err
}

func fixedLocalRecordAccess(ins *Instr, index int, proof *x86FixedLocalProof, residue int,
	width int64, write, nativeInput bool, offset int64) error {
	if ins.x86FixedLocalMemory == nil {
		ins.x86FixedLocalMemory = make(map[int]x86FixedLocalAccess)
	}
	shape, err := fixedLocalInstructionDigest(*ins)
	if err != nil {
		return err
	}
	access, found := ins.x86FixedLocalMemory[index]
	if !found {
		access = x86FixedLocalAccess{owner: proof, shape: shape, width: width, write: write, nativeInput: nativeInput, ordinal: proof.ordinal}
	}
	if access.owner != proof || access.shape != shape || access.width != width || access.write != write || access.nativeInput != nativeInput {
		return goAMD64FrameContext("fixed local access shape changed")
	}
	access.offset[residue] = offset
	ins.x86FixedLocalMemory[index] = access
	return nil
}

func fixedLocalStep(ins *Instr, state *x86FixedLocalState, sig FuncSig, proof *x86FixedLocalProof, residue int) error {
	fail := func() error {
		return goAMD64FrameContext("unproved fixed local width/initialization/address/control effect: %s", ins.Op)
	}
	if ins.Op == OpTEXT {
		return nil
	}
	if _, known, err := x86OrderingHintProperties(ins.Op, *ins, "amd64"); known {
		return err // Exact native fences change no byte, pointer or FLAGS facts.
	}
	if form, known, err := x86VectorZeroProperties(*ins, "amd64"); known {
		if err != nil {
			return err
		}
		x86VectorZeroDefinedBytes(form, state.vectors)
		return nil
	}
	if ins.Op == OpRET {
		sp, found := state.pointers[SP]
		if len(ins.Args) != 0 || !found || sp != 0 {
			return fail()
		}
		for index := range sig.Frame.Results {
			if !state.results[index] {
				return fail()
			}
		}
		return nil
	}
	if isAMD64ConditionalBranch(ins.Op) {
		needed := x86ByteBranchFlags(ins.Op)
		if needed == 0 || state.flags&needed != needed {
			return fail()
		}
		return nil
	}
	if ins.Op == OpJMP {
		return nil
	}
	if known, err := fixedLocalNumericalReadStep(ins, state, proof, residue); known {
		return err
	}
	if known, err := fixedLocalStaticAddressStep(*ins, state, proof); known {
		return err
	}
	if width := x86RawReturnScalarMoveBytes(x86GoMOVQSourceInstruction(*ins).Op); width != 0 {
		return fixedLocalMove(ins, state, sig, proof, residue, width)
	}
	if known, err := fixedLocalConditionalSetStep(ins, state, sig, proof, residue); known {
		return err
	}
	if ins.Op == "NOTQ" && len(ins.Args) == 1 && ins.Args[0].Kind == OpReg {
		reg := ins.Args[0].Reg
		if _, pointer := state.pointers[reg]; pointer {
			return fail()
		}
		value, known := state.constants[reg]
		if known {
			state.constants[reg] = ^value
			fixedLocalWriteBooleans(state, reg, 8, fixedLocalConstantBooleans(^value, 8))
			return nil
		}
	}
	if (ins.Op == "ADDQ" || ins.Op == "ANDQ") && len(ins.Args) == 2 && ins.Args[1].Kind == OpReg {
		reg := ins.Args[1].Reg
		value := ins.Args[0].Imm
		known := ins.Args[0].Kind == OpImm && ins.Args[0].ImmRaw == "" && !ins.Args[0].ImmIsFloat
		if ins.Args[0].Kind == OpReg {
			value, known = state.constants[ins.Args[0].Reg]
		}
		if offset, pointer := state.pointers[reg]; pointer {
			if !known {
				return fail()
			}
			if ins.Op == "ANDQ" {
				// Only an exact power-of-two residue projection preserves the
				// same private address origin. No pointer bit becomes numeric.
				if value != -8 && value != -16 && value != -32 && value != -64 {
					return fail()
				}
				offset = (offset+int64(residue*8))&value - int64(residue*8)
			} else {
				if value > 0 && offset > math.MaxInt64-value || value < 0 && offset < math.MinInt64-value {
					return fail()
				}
				offset += value
			}
			if offset < -64 || offset > proof.frame {
				return fail()
			}
			state.pointers[reg] = offset
			state.flags = 0 // Relocated pointer arithmetic flags are opaque.
			return nil
		}
	}
	if known, err := fixedLocalConditionalMoveStep(ins, state, proof, residue); known {
		return err
	}
	if known, err := fixedLocalInputAddressStep(ins, state, proof, residue); known {
		return err
	}
	if effect, known, err := fixedLocalGPByteEffect(*ins); known {
		if err != nil {
			return err
		}
		return fixedLocalTypedByteStep(ins, state, proof, residue, effect)
	}
	if known, err := fixedLocalAESByteStep(ins, state, proof, residue); known {
		return err
	}
	if known, err := fixedLocalPackedMoveStep(ins, state, proof, residue); known {
		return err
	}
	if known, err := fixedLocalMaskMoveStep(ins, state, proof, residue); known {
		return err
	}
	if known, err := fixedLocalConversionStep(ins, state, proof, residue); known {
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
	_, opaque := state.opaque[root]
	mask := x86CallerByteMask(operand.Reg, width)
	return valid && root != SP && !pointer && !opaque && mask != 0 && state.known[root]&mask == mask
}

func fixedLocalMove(ins *Instr, state *x86FixedLocalState, sig FuncSig, proof *x86FixedLocalProof, residue int, width int64) error {
	fail := func() error { return goAMD64FrameContext("unproved fixed local scalar move") }
	if len(ins.Args) != 2 {
		return fail()
	}
	source, destination := ins.Args[0], ins.Args[1]
	var address int64
	var pointer bool
	var opaque bool
	var identity x86FixedLocalOrigin
	var booleans uint8
	var constant int64
	var constantKnown bool
	switch source.Kind {
	case OpReg:
		address, pointer = state.pointers[source.Reg]
		identity, opaque = state.opaque[source.Reg]
		if pointer || opaque {
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
			booleans = fixedLocalBooleanRead(source, width, state)
		}
	case OpImm:
		if !fixedLocalNumericRead(source, width, state) {
			return fail()
		}
		constant, constantKnown = source.Imm, width == 8
		booleans = fixedLocalBooleanRead(source, width, state)
	case OpFP:
		sp, original := state.pointers[SP]
		slot, _, found := fixedLocalTransportSlot(source.FPOffset, width, sig.Frame.Params)
		if !original || sp != 0 || !found {
			return fail()
		}
		if slot.Type == Ptr {
			opaque, identity = true, x86FixedLocalOrigin{parameter: slot.Offset}
		} else if slot.Type == I1 {
			booleans = 1
		}
	case OpSym:
		if _, _, err := fixedLocalReadMemory(ins, 0, state, proof, residue, width, 1); err != nil {
			return err
		}
	case OpMem:
		memory, err := fixedLocalMemoryAddress(*ins, 0, proof.frame)
		if err != nil {
			return err
		}
		base, found := state.pointers[memory.Base]
		if found && memory.Off >= -proof.frame && memory.Off <= proof.frame {
			address, pointer = state.addresses[base+memory.Off]
			identity, opaque = state.opaqueCells[base+memory.Off]
		}
		if (pointer || opaque) && width != 8 {
			return fail()
		}
		offset, role, err := fixedLocalMemoryAccess(ins, 0, state, proof, residue, width, 1, false, pointer || opaque)
		if err != nil {
			return err
		}
		if role == x86FixedLocalPrivateMemory && !pointer && !opaque {
			booleans = fixedLocalMemoryBooleans(state, offset, width)
		}
	default:
		return fail()
	}
	switch destination.Kind {
	case OpReg:
		root, valid := x86CallerGP(destination.Reg)
		if !valid || (pointer || opaque) && (width != 8 || destination.Reg != root) {
			return fail()
		}
		if root == SP && !pointer {
			return fail()
		}
		if _, oldPointer := state.pointers[root]; oldPointer && width != 8 {
			return fail()
		}
		delete(state.pointers, root)
		delete(state.opaque, root)
		delete(state.constants, root)
		fixedLocalWriteBooleans(state, destination.Reg, width, booleans)
		if pointer {
			state.pointers[root] = address
			state.known[root] = 0
		} else if opaque {
			state.opaque[root] = identity
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
		offset, role, err := fixedLocalMemoryAccess(ins, 1, state, proof, residue, width, 1, true, false)
		if err != nil {
			return err
		}
		if role == x86FixedLocalNativeInput {
			if pointer || opaque {
				return fail() // Neither private nor input pointers escape as bytes.
			}
			return nil
		}
		fixedLocalRevokeSpillRoles(state, offset, width)
		for index := offset; index < offset+width; index++ {
			state.bytes[index] = !pointer && !opaque
			state.booleanBytes[index] = booleans&(1<<uint(index-offset)) != 0
		}
		if pointer {
			state.addresses[offset] = address
		}
		if opaque {
			state.opaqueCells[offset] = identity
		}
	case OpFP:
		sp, original := state.pointers[SP]
		slot, index, found := fixedLocalTransportSlot(destination.FPOffset, width, sig.Frame.Results)
		if !original || sp != 0 || pointer || identity.literal.symbol != "" || !found || source.Kind != OpReg && source.Kind != OpImm ||
			opaque != (slot.Type == Ptr) || slot.Type == I1 && booleans&1 == 0 {
			return fail()
		}
		state.results[index] = true
	default:
		return fail()
	}
	return nil
}

func fixedLocalVectorStep(ins *Instr, state *x86FixedLocalState, proof *x86FixedLocalProof, residue int) error {
	if effect, known, err := fixedLocalVectorByteEffect(*ins); known {
		if err != nil {
			return err
		}
		return fixedLocalTypedByteStep(ins, state, proof, residue, effect)
	}
	width := uint64(65535)
	first, second, destination, memory := 0, 1, 1, -1
	if form, known, err := x86FloatCompareProperties(*ins, "amd64"); known {
		if err != nil || !form.legacy || form.scalar || form.byteWidth != 16 {
			return goAMD64FrameContext("fixed local compare shape is not closed")
		}
		first, second, destination, memory = form.first, form.second, form.destination, form.memory
	} else {
		return goAMD64FrameContext("fixed local source has an unregistered typed vector effect: %s", ins.Op)
	}
	if state.vectors[x86CallerVectorRoot(ins.Args[second].Reg)]&width != width {
		return goAMD64FrameContext("fixed local vector operation reads undefined destination bytes")
	}
	if memory >= 0 {
		if _, _, err := fixedLocalReadMemory(ins, memory, state, proof, residue, 16, 16); err != nil {
			return err
		}
	} else if state.vectors[x86CallerVectorRoot(ins.Args[first].Reg)]&width != width {
		return goAMD64FrameContext("fixed local vector operation reads undefined source bytes")
	}
	state.vectors[x86CallerVectorRoot(ins.Args[destination].Reg)] |= width
	return nil
}

func (c *amd64Ctx) fixedLocalMemoryCertified(ins Instr, index int) bool {
	access, valid := c.fixedLocalMemoryAccess(ins, index)
	return valid && !access.nativeInput
}

func (c *amd64Ctx) fixedLocalMemoryAccess(ins Instr, index int) (x86FixedLocalAccess, bool) {
	access, found := ins.x86FixedLocalMemory[index]
	shape, err := fixedLocalInstructionDigest(ins)
	if !found || c.fixedLocal == nil || access.owner != c.fixedLocal || ins.x86Encoded ||
		ins.x86SourceRaw != nil || ins.x86AddressBits != 0 || ins.x86SegmentPrefix != 0 ||
		err != nil || access.shape != shape || access.width <= 0 {
		return x86FixedLocalAccess{}, false
	}
	expected, bound := c.fixedLocal.accesses[access.ordinal][index]
	return access, bound && reflect.DeepEqual(access, expected)
}
