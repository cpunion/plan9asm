package plan9asm

import (
	"fmt"
	"reflect"
	"strings"
)

// CheckX86ABI0ResultInitialization checks only byte initialization of one
// ordinary scalar ABI0 result. It supplies no source authentication, native
// frame, caller-memory, return-continuation or translation permission. The
// caller must derive sig from the actual selected Go declaration and bind the
// complete original CPP source independently. Unknown effects fail closed.
func CheckX86ABI0ResultInitialization(fn Func, sig FuncSig, goarch string) error {
	fail := func(reason string) error {
		return fmt.Errorf("%w: result initialization: %s", ErrProbeNeedsContext, reason)
	}
	if goarch != "amd64" && goarch != "386" || strings.Contains(fn.Sym, "<") ||
		len(sig.ArgRegs) != 0 || sig.ARMEntry != nil || sig.WASMContext != "" || len(sig.Frame.Results) != 1 {
		return fail("requires an ordinary x86 scalar ABI0 declaration")
	}
	result := sig.Frame.Results[0]
	width := x86ResultScalarBytes(result.Type, goarch)
	if width == 0 || result.Type != sig.Ret || result.Index != 0 || result.Field != -1 || len(result.Fields) != 0 ||
		result.Offset < 0 || result.Offset > 4096-width {
		return fail("result layout is not one bounded scalar")
	}
	frame := int(result.Offset + width)
	initial := x86ResultState{fp: make([]bool, frame)}
	autoExtent, autoErr := x86ResultAUTOExtent(fn, goarch)
	if autoErr != nil {
		return autoErr
	}
	initial.auto = make([]bool, autoExtent)
	owned := make([]bool, frame)
	for offset := int(result.Offset); offset < frame; offset++ {
		owned[offset] = true
	}
	for _, slot := range sig.Frame.Params {
		// Actual gc layout includes pointer storage even when no numeric
		// instruction consumes it. Occupancy prevents overlap, but pointer
		// bytes are deliberately not a numerical input or pointee witness.
		bytes := x86FrameScalarBytes(slot.Type, goarch)
		if bytes == 0 || slot.Offset < 0 || slot.Offset > int64(frame)-bytes || slot.Offset+bytes > result.Offset {
			return fail("parameter bytes or result boundary are not closed")
		}
		for offset := int(slot.Offset); offset < int(slot.Offset+bytes); offset++ {
			if owned[offset] {
				return fail("overlapping typed frame slots")
			}
			owned[offset], initial.fp[offset] = true, slot.Type != Ptr
		}
	}
	if fn.ArgSize != 0 && (fn.ArgSize < int64(frame) || fn.ArgSize > 4096) {
		return fail("TEXT argument extent disagrees with the typed result")
	}
	labels := make(map[string]int)
	var instructions []Instr
	for _, ins := range fn.Instrs {
		if ins.Op == OpTEXT {
			continue
		}
		if ins.Op == OpLABEL {
			if len(ins.Args) != 1 || ins.Args[0].Kind != OpLabel {
				return fail("malformed label")
			}
			if _, duplicate := labels[ins.Args[0].Sym]; duplicate {
				return fail("duplicate label")
			}
			labels[ins.Args[0].Sym] = len(instructions)
			continue
		}
		instructions = append(instructions, ins)
	}
	if len(instructions) == 0 || len(instructions) > 4096 {
		return fail("missing or oversized source CFG")
	}
	// One bounded byte-work envelope includes both logical domains. AUTO
	// states do not obtain a second independent frame or iteration budget.
	byteState := len(initial.fp) + len(initial.auto) + 64
	const byteWorkLimit = 64 << 20
	byteWork := len(instructions) * byteState
	if autoExtent != 0 && byteWork > byteWorkLimit/8 {
		return fail("combined FP/AUTO source-CFG byte budget exhausted")
	}
	// Validate effects and all static successors source-wide. Definedness is
	// proved below only for reachable paths; dead instructions must not hide
	// unknown effects, malformed branches or external entry requirements.
	top := initial.clone()
	for offset := range top.fp {
		top.fp[offset] = initial.fp[offset] || int64(offset) >= result.Offset
	}
	for offset := range top.auto {
		top.auto[offset] = true
	}
	for index := range top.gp {
		top.gp[index], top.xmm[index] = 0xff, 0xffff
	}
	top.flags = true
	edges := make([][]int, len(instructions))
	for pc, ins := range instructions {
		if ins.Op == OpRET {
			if len(ins.Args) != 0 {
				return fail("nonordinary return")
			}
			continue
		}
		shape := top.clone()
		if !x86ResultTransfer(ins, &shape, owned, result.Offset, goarch) {
			return fail(fmt.Sprintf("unclosed source effect at instruction %d: %s", pc, ins.Raw))
		}
		next := []int{pc + 1}
		if ins.Op == OpJMP || isAMD64ConditionalBranch(ins.Op) {
			if len(ins.Args) != 1 {
				return fail("unbound branch operands")
			}
			arg := ins.Args[0]
			label := arg.Sym
			if arg.Kind == OpIdent {
				label = arg.Ident
			}
			target, found := labels[label]
			if arg.Kind != OpLabel && arg.Kind != OpIdent || !found {
				return fail("indirect, external or unresolved branch")
			}
			if ins.Op == OpJMP {
				next = []int{target}
			} else {
				next = append(next, target)
			}
		}
		for _, target := range next {
			if target < 0 || target >= len(instructions) {
				return fail("source falls outside its continuation")
			}
		}
		edges[pc] = next
	}
	states := map[int]x86ResultState{0: initial}
	queue := []int{0}
	returned := false
	for visits := 0; len(queue) != 0; visits++ {
		if visits >= 16384 {
			return fail("must-defined fixed-point budget exhausted")
		}
		if autoExtent != 0 {
			byteWork += 8 * byteState // conservative bound for copies and joins
			if byteWork > byteWorkLimit {
				return fail("combined FP/AUTO must-defined byte budget exhausted")
			}
		}
		pc := queue[0]
		queue = queue[1:]
		state := states[pc].clone()
		ins := instructions[pc]
		if ins.Op == OpRET {
			if len(ins.Args) != 0 {
				return fail("nonordinary return")
			}
			returned = true
			continue // Check results after the complete fixed point, never early.
		}
		if !x86ResultTransfer(ins, &state, owned, result.Offset, goarch) {
			return fail(fmt.Sprintf("unknown effect or undefined input at instruction %d: %s", pc, ins.Raw))
		}
		for _, target := range edges[pc] {
			if previous, visited := states[target]; visited {
				merged := previous.intersect(state)
				if reflect.DeepEqual(previous, merged) {
					continue
				}
				states[target] = merged
			} else {
				states[target] = state.clone()
			}
			queue = append(queue, target)
		}
	}
	if !returned {
		return fail("no ordinary reachable RET")
	}
	for pc, state := range states {
		if instructions[pc].Op != OpRET {
			continue
		}
		for offset := int(result.Offset); offset < frame; offset++ {
			if !state.fp[offset] {
				return fail(fmt.Sprintf("result byte %d is not written on every path to RET %d", offset, pc))
			}
		}
	}
	return nil
}

func x86ResultScalarBytes(typ LLVMType, arch string) int64 {
	if typ == Ptr {
		return 0 // Initialized pointer bits do not close a Go result/GC contract.
	}
	return x86FrameScalarBytes(typ, arch)
}

type x86ResultState struct {
	fp    []bool
	auto  []bool
	gp    [16]uint8
	xmm   [16]uint16
	flags bool
}

func (state x86ResultState) clone() x86ResultState {
	state.fp = append([]bool(nil), state.fp...)
	state.auto = append([]bool(nil), state.auto...)
	return state
}

func (state x86ResultState) intersect(other x86ResultState) x86ResultState {
	merged := state.clone()
	for index := range merged.fp {
		merged.fp[index] = merged.fp[index] && other.fp[index]
	}
	for index := range merged.auto {
		merged.auto[index] = merged.auto[index] && other.auto[index]
	}
	for index := range merged.gp {
		merged.gp[index] &= other.gp[index]
		merged.xmm[index] &= other.xmm[index]
	}
	merged.flags = state.flags && other.flags
	return merged
}

func x86ResultGP(reg Reg, width int64, arch string) (int, uint8, bool) {
	root, valid := x86CallerGP(reg)
	// This observer admits only ABI0. Its R14 entry bytes are undefined,
	// unlike ABIInternal's opaque g identity; only explicit source writes can
	// initialize them. No caller/source-stack/native permissions are changed.
	if !valid || root == SP || root == BP || width == 8 && arch == "386" {
		return 0, 0, false
	}
	if width != 1 && root != reg || width == 1 && !isGoYmbRegisterForArch(reg, arch) {
		return 0, 0, false
	}
	for index, candidate := range []Reg{AX, CX, DX, BX, SP, BP, SI, DI, "R8", "R9", "R10", "R11", "R12", "R13", "R14", "R15"} {
		if root == candidate && (arch != "386" || index < 8) {
			mask := x86CallerByteMask(reg, width)
			return index, mask, mask != 0
		}
	}
	return 0, 0, false
}

func x86ResultRead(arg Operand, width int64, state *x86ResultState, owned []bool, arch string) bool {
	switch arg.Kind {
	case OpImm:
		return arg.ImmRaw == "" && !arg.ImmIsFloat
	case OpReg:
		index, mask, valid := x86ResultGP(arg.Reg, width, arch)
		return valid && state.gp[index]&mask == mask
	case OpFP:
		if arg.FPOffset < 0 || width > int64(len(owned)) || arg.FPOffset > int64(len(owned))-width {
			return false
		}
		for offset := arg.FPOffset; offset < arg.FPOffset+width; offset++ {
			if !owned[offset] || !state.fp[offset] {
				return false
			}
		}
		return true
	case OpMem:
		offset, valid := x86ResultAUTOOffset(arg, width, len(state.auto))
		if !valid {
			return false
		}
		for index := offset; index < offset+int(width); index++ {
			if !state.auto[index] {
				return false
			}
		}
		return true
	}
	return false
}

func x86ResultStore(arg Operand, width int64, state *x86ResultState, owned []bool, resultOffset int64, arch string) bool {
	if arg.Kind == OpMem {
		offset, valid := x86ResultAUTOOffset(arg, width, len(state.auto))
		if !valid {
			return false
		}
		for index := offset; index < offset+int(width); index++ {
			state.auto[index] = true
		}
		return true
	}
	if arg.Kind == OpReg {
		index, mask, valid := x86ResultGP(arg.Reg, width, arch)
		if valid {
			state.gp[index] |= mask
			if width == 4 || width == 8 {
				state.gp[index] = 0xff
			}
		}
		return valid
	}
	if arg.Kind != OpFP || arg.FPOffset < resultOffset || width > int64(len(owned)) || arg.FPOffset > int64(len(owned))-width {
		return false
	}
	for offset := arg.FPOffset; offset < arg.FPOffset+width; offset++ {
		if !owned[offset] {
			return false
		}
		state.fp[offset] = true
	}
	return true
}

func x86ResultTransfer(ins Instr, state *x86ResultState, owned []bool, resultOffset int64, arch string) bool {
	if ins.Op == OpJMP {
		return true // The CFG checks its complete target independently.
	}
	if isAMD64ConditionalBranch(ins.Op) {
		return !isAMD64CounterZeroBranch(ins.Op) && state.flags
	}
	if ins.Op == "NOP" {
		return len(ins.Args) == 0
	}
	if form, recognized, err := x86ConditionalSetProperties(ins.Op, ins, arch); recognized {
		return err == nil && state.flags && x86ResultStore(form.destination, 1, state, owned, resultOffset, arch)
	}
	if width := x86RawReturnScalarMoveBytes(ins.Op); width != 0 {
		return !(width == 8 && arch == "386") && len(ins.Args) == 2 && !(ins.Args[0].Kind == OpFP && ins.Args[1].Kind == OpFP) &&
			x86ResultRead(ins.Args[0], width, state, owned, arch) &&
			x86ResultStore(ins.Args[1], width, state, owned, resultOffset, arch)
	}
	if form, recognized, err := amd64ScalarVectorMoveProperties(ins, arch); recognized {
		if err != nil || !form.spec.legacy || form.mask >= 0 || len(ins.Args) != 2 {
			return false
		}
		width := int64(form.spec.bits / 8)
		mask := uint16(1)<<uint(width) - 1
		xreg := func(arg Operand) (int, bool) {
			index, valid := amd64ParseXReg(arg.Reg)
			return index, arg.Kind == OpReg && valid && index < 16 && (arch != "386" || index < 8)
		}
		first, last := ins.Args[0], ins.Args[1]
		if source, valid := xreg(first); valid {
			if state.xmm[source]&mask != mask {
				return false
			}
		} else if !form.literal && !x86ResultRead(first, width, state, owned, arch) {
			return false
		}
		if destination, valid := xreg(last); valid {
			state.xmm[destination] |= mask
			if form.memory == 0 || form.legacyZero {
				state.xmm[destination] = 0xffff
			}
			return true
		}
		return x86ResultStore(last, width, state, owned, resultOffset, arch)
	}
	name := string(ins.Op)
	if len(name) < 2 {
		return false
	}
	base := name[:len(name)-1]
	width := map[byte]int64{'B': 1, 'W': 2, 'L': 4, 'Q': 8}[name[len(name)-1]]
	if width == 0 || width == 8 && arch == "386" || len(ins.Args) != 2 {
		return false
	}
	if base != "CMP" && base != "TEST" && base != "ADD" && base != "SUB" && base != "AND" && base != "OR" && base != "XOR" {
		return false
	}
	zero := base == "XOR" && ins.Args[0].Kind == OpReg && reflect.DeepEqual(ins.Args[0], ins.Args[1])
	if !zero && (!x86ResultRead(ins.Args[0], width, state, owned, arch) || !x86ResultRead(ins.Args[1], width, state, owned, arch)) {
		return false
	}
	state.flags = true
	if base == "CMP" || base == "TEST" {
		return true
	}
	return ins.Args[1].Kind == OpReg && x86ResultStore(ins.Args[1], width, state, owned, resultOffset, arch)
}
