package plan9asm

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// Local code addresses stay opaque LLVM blockaddresses. A reaching-definition proof,
// not an address comparison, distinguishes local control flow from native calls.
// Memory provenance is retained only for exactly identified owned frame cells.
type arm64ControlValue map[string]bool

type arm64ControlState struct {
	regs    map[Reg]arm64ControlValue
	memory  map[string]arm64ControlValue
	escaped arm64ControlValue
}

type arm64LocalControlPlan struct {
	before      map[[2]int]arm64ControlValue
	stackBefore map[[2]int]arm64ControlValue
	linkBefore  map[[2]int]arm64ControlValue
	reachable   []bool
	outer       string
	autoFrame   bool
}

type arm64GoFrame struct {
	present      bool
	restoresLink bool
}

// Go obj7 marks only source ABL instructions nonleaf, before expanding raw
// AWORD instructions. A decoded WORD BL must not invent a Go LR/FP prologue.
func arm64SourceGoFrame(fn Func) arm64GoFrame {
	noFrame, leaf := fn.FrameSize == -8, true
	for _, ins := range fn.Instrs {
		op := arm64ControlOp(ins)
		if op == "BL" || op == "CALL" {
			leaf = false
		}
		if op != OpTEXT {
			continue
		}
		_, rest := splitOpcode(ins.Raw)
		parts := strings.Split(rest, ",")
		if len(parts) != 3 {
			continue
		}
		if flags, ok := parseImmExpr(strings.TrimSpace(parts[1])); ok {
			noFrame = noFrame || flags&512 != 0 // objabi.NOFRAME
		} else {
			for _, flag := range strings.Split(parts[1], "|") {
				noFrame = noFrame || strings.TrimSpace(flag) == "NOFRAME"
			}
		}
	}
	present := !noFrame && (!leaf || fn.FrameSize > 0)
	return arm64GoFrame{present: present, restoresLink: present && !leaf}
}

func arm64ControlExternal() arm64ControlValue { return arm64ControlValue{"": true} }

func arm64ControlRead(values map[Reg]arm64ControlValue, reg Reg) arm64ControlValue {
	if value := values[reg]; value != nil {
		return value
	}
	return arm64ControlExternal()
}

func arm64ControlUnion(left, right arm64ControlValue) arm64ControlValue {
	value := make(arm64ControlValue, len(left)+len(right))
	for token := range left {
		value[token] = true
	}
	for token := range right {
		value[token] = true
	}
	// Disagreeing affine frame addresses become unknown, so moving-SP loops
	// cannot create an unbounded offset lattice.
	frame := ""
	for token := range value {
		if strings.HasPrefix(token, "sp:") || strings.HasPrefix(token, "fp:") {
			if frame != "" && frame != token {
				for candidate := range value {
					if strings.HasPrefix(candidate, "sp:") || strings.HasPrefix(candidate, "fp:") {
						delete(value, candidate)
					}
				}
				value[""] = true
				break
			}
			frame = token
		}
	}
	return value
}

func arm64ControlEqual(left, right arm64ControlValue) bool {
	if len(left) != len(right) {
		return false
	}
	for token := range left {
		if !right[token] {
			return false
		}
	}
	return true
}

func (state *arm64ControlState) clone() *arm64ControlState {
	next := &arm64ControlState{regs: make(map[Reg]arm64ControlValue), memory: make(map[string]arm64ControlValue), escaped: state.escaped}
	for reg, value := range state.regs {
		next.regs[reg] = value
	}
	for key, value := range state.memory {
		next.memory[key] = value
	}
	return next
}

func (state *arm64ControlState) merge(other *arm64ControlState) bool {
	changed := false
	escaped := arm64ControlUnion(state.escaped, other.escaped)
	if !arm64ControlEqual(state.escaped, escaped) {
		state.escaped, changed = escaped, true
	}
	for reg, value := range other.regs {
		merged := arm64ControlUnion(arm64ControlRead(state.regs, reg), value)
		if !arm64ControlEqual(state.regs[reg], merged) {
			state.regs[reg], changed = merged, true
		}
	}
	for reg, value := range state.regs {
		if other.regs[reg] == nil {
			merged := arm64ControlUnion(value, arm64ControlExternal())
			if !arm64ControlEqual(value, merged) {
				state.regs[reg], changed = merged, true
			}
		}
	}
	for key, value := range other.memory {
		old := state.memory[key]
		if old == nil {
			old = arm64ControlExternal()
		}
		merged := arm64ControlUnion(old, value)
		if !arm64ControlEqual(state.memory[key], merged) {
			state.memory[key], changed = merged, true
		}
	}
	for key, value := range state.memory {
		if other.memory[key] == nil {
			merged := arm64ControlUnion(value, arm64ControlExternal())
			if !arm64ControlEqual(value, merged) {
				state.memory[key], changed = merged, true
			}
		}
	}
	return changed
}

func arm64ControlBase(reg Reg) Reg {
	if reg == Reg("RSP") {
		return SP
	}
	return reg
}

func arm64ControlOffset(value arm64ControlValue, off int64) arm64ControlValue {
	if len(value) != 1 {
		return arm64ControlExternal()
	}
	for token := range value {
		if strings.HasPrefix(token, "fpa:") {
			if off == 0 {
				return value
			}
			// FP-address allocas are per scalar slot, unlike the contiguous
			// SP backing frame. Adding bytes cannot prove the adjacent slot.
			return arm64ControlExternal()
		}
		if !strings.HasPrefix(token, "sp:") && !strings.HasPrefix(token, "fp:") {
			return arm64ControlExternal()
		}
		base, err := strconv.ParseInt(token[3:], 10, 64)
		if err != nil || (off > 0 && base > int64(1<<63-1)-off) || (off < 0 && base < int64(-1<<63)-off) {
			return arm64ControlExternal()
		}
		return arm64ControlValue{token[:3] + strconv.FormatInt(base+off, 10): true}
	}
	return arm64ControlExternal()
}

func (state *arm64ControlState) memoryKey(op Operand, post bool) string {
	if op.Kind == OpFP || op.Kind == OpFPAddr {
		return "fp:" + strconv.FormatInt(op.FPOffset, 10)
	}
	if op.Kind != OpMem || op.Mem.Index != "" || op.Mem.OffRaw != "" {
		return ""
	}
	off := op.Mem.Off
	if post {
		off = 0
	}
	value := arm64ControlOffset(arm64ControlRead(state.regs, arm64ControlBase(op.Mem.Base)), off)
	if len(value) == 1 {
		for token := range value {
			if strings.HasPrefix(token, "fpa:") {
				return "fp:" + token[4:]
			}
			return token
		}
	}
	return ""
}

func (state *arm64ControlState) source(op Operand, post bool) arm64ControlValue {
	if op.Kind == OpReg {
		return arm64ControlRead(state.regs, arm64ControlBase(op.Reg))
	}
	if op.Kind == OpFPAddr {
		return arm64ControlValue{"fpa:" + strconv.FormatInt(op.FPOffset, 10): true}
	}
	if key := state.memoryKey(op, post); key != "" && state.memory[key] != nil {
		return arm64ControlUnion(state.memory[key], arm64ControlLabels(state.escaped))
	}
	if op.Kind == OpMem || op.Kind == OpFP || op.Kind == OpSym {
		return arm64ControlUnion(arm64ControlExternal(), state.escaped)
	}
	return arm64ControlExternal()
}

func (state *arm64ControlState) write(op Operand, value arm64ControlValue, post bool) {
	if op.Kind == OpReg && op.Reg != ZR {
		state.regs[arm64ControlBase(op.Reg)] = value
	} else if key := state.memoryKey(op, post); key != "" {
		// Any nonidentical cell may overlap (including through unknown aliases).
		// Preserve a tainted local address rather than silently reclassify a
		// partially overwritten pointer as an external native address.
		state.invalidateOverlappingControlCells(key, 8)
		state.memory[key] = value
	} else if op.Kind == OpMem || op.Kind == OpFP || op.Kind == OpSym {
		state.escaped = arm64ControlUnion(state.escaped, arm64ControlAddressTaint(value))
		for key, old := range state.memory {
			state.memory[key] = arm64ControlUnion(old, arm64ControlExternal())
		}
	} else if op.Kind == OpRegList {
		for _, reg := range op.RegList {
			state.write(Operand{Kind: OpReg, Reg: reg}, value, false)
		}
	}
}

func arm64ControlLabels(value arm64ControlValue) arm64ControlValue {
	labels := make(arm64ControlValue)
	for token := range value {
		if strings.HasPrefix(token, "label:") {
			labels[token] = true
		}
	}
	return labels
}

func arm64ControlAddressTaint(value arm64ControlValue) arm64ControlValue {
	taint := make(arm64ControlValue)
	for token := range value {
		if token != "" {
			taint[token] = true
		}
	}
	return taint
}

func (state *arm64ControlState) invalidateOverlappingControlCells(key string, width int64) {
	if len(key) < 4 {
		return
	}
	off, err := strconv.ParseInt(key[3:], 10, 64)
	if err != nil {
		return
	}
	for other, value := range state.memory {
		if len(other) < 4 || other[:3] != key[:3] {
			continue
		}
		start, err := strconv.ParseInt(other[3:], 10, 64)
		if err != nil {
			continue
		}
		// Unsigned distances avoid signed overflow at either int64 boundary.
		overlaps := (start <= off && uint64(off)-uint64(start) < 8) ||
			(start > off && uint64(start)-uint64(off) < uint64(width))
		if overlaps {
			state.memory[other] = arm64ControlUnion(value, arm64ControlExternal())
		}
	}
}

func arm64ControlElement(op Operand, offset int64) Operand {
	if op.Kind == OpFP {
		op.FPOffset += offset
	} else if op.Kind == OpMem {
		op.Mem.Off += offset
	}
	return op
}

func arm64ControlOp(ins Instr) Op {
	op := strings.ToUpper(string(ins.Op))
	if dot := strings.IndexByte(op, '.'); dot >= 0 {
		op = op[:dot]
	}
	return Op(op)
}

func arm64ControlDecode(ins Instr) Instr {
	if ins.Op == OpWORD {
		if decoded, err := decodeARM64RawWordInstruction(ins); err == nil {
			return arm64HardwareReturnAsBranch(decoded)
		}
	}
	return ins
}

// A hardware RET word has no Go-generated frame epilogue. Its branch effect
// must not be confused with the source-level RET pseudo instruction.
func arm64HardwareReturnAsBranch(ins Instr) Instr {
	if ins.Op != OpRET {
		return ins
	}
	target, ok := arm64RegisterReturnTarget(ins, false)
	if !ok {
		return ins
	}
	if target.Kind == OpReg {
		target = Operand{Kind: OpMem, Mem: MemRef{Base: target.Reg}}
	}
	ins.Op, ins.Args = "B", []Operand{target}
	return ins
}

func (c *arm64Ctx) prepareLocalControl() error {
	plan := &arm64LocalControlPlan{
		before:      make(map[[2]int]arm64ControlValue),
		stackBefore: make(map[[2]int]arm64ControlValue),
		linkBefore:  make(map[[2]int]arm64ControlValue),
	}
	plan.autoFrame = c.sourceGoFrame.present
	indices := make(map[string]int, len(c.blocks))
	for index, block := range c.blocks {
		indices[block.name] = index
	}
	// The caller link has a distinguished abstract return destination. Its
	// actual value comes from llvm.returnaddress, never a fake blockaddress.
	plan.outer = "__arm64_caller_return"
	for indices[plan.outer] != 0 || c.blocks[0].name == plan.outer {
		plan.outer += "_"
	}
	entry := &arm64ControlState{regs: map[Reg]arm64ControlValue{
		SP: {"sp:0": true}, Reg("R30"): {"label:" + plan.outer: true},
	}, memory: make(map[string]arm64ControlValue)}
	if c.sourceGoFrame.present {
		entry.memory["sp:0"] = entry.regs[Reg("R30")]
	}
	for _, reg := range c.sig.ArgRegs {
		entry.regs[reg] = arm64ControlExternal()
	}
	before := make([]*arm64ControlState, len(c.blocks))
	before[0] = entry
	queue, queued := []int{0}, make([]bool, len(c.blocks))
	queued[0] = true
	propagate := func(index int, state *arm64ControlState) {
		if index < 0 || index >= len(before) {
			return
		}
		changed := false
		if before[index] == nil {
			before[index], changed = state.clone(), true
		} else {
			changed = before[index].merge(state)
		}
		if changed && !queued[index] {
			queue, queued[index] = append(queue, index), true
		}
	}
	for next := 0; next < len(queue); next++ {
		bi := queue[next]
		queued[bi] = false
		state := before[bi].clone()
		terminated := false
		for ii, original := range c.blocks[bi].instrs {
			ins := arm64ControlDecode(original)
			op := arm64ControlOp(ins)
			post := strings.HasSuffix(string(ins.Op), ".P")
			if op == "B" || op == "JMP" || op == OpRET {
				plan.stackBefore[[2]int{bi, ii}] = arm64ControlRead(state.regs, SP)
				link := arm64ControlRead(state.regs, Reg("R30"))
				if op == OpRET && c.sourceGoFrame.restoresLink {
					link = state.source(Operand{Kind: OpMem, Mem: MemRef{Base: SP}}, false)
				}
				plan.linkBefore[[2]int{bi, ii}] = link
			}
			if op == "BL" || op == "BLR" || op == "CALL" {
				if len(ins.Args) != 1 || bi+1 >= len(c.blocks) {
					continue // The ordinary grammar diagnoses malformed calls.
				}
				if ins.Args[0].Kind == OpReg || (ins.Args[0].Kind == OpMem && ins.Args[0].Mem.Base != PC) {
					value := state.branchValue(ins.Args[0])
					plan.before[[2]int{bi, ii}] = value
					state.regs[Reg("R30")] = arm64ControlValue{"label:" + c.blocks[bi+1].name: true}
					for token := range value {
						if strings.HasPrefix(token, "label:") {
							if target, ok := indices[token[6:]]; ok {
								propagate(target, state)
							}
						} else {
							nativeState := state.clone()
							c.invalidateControlCallResults(nativeState, Operand{})
							propagate(bi+1, nativeState)
						}
					}
					terminated = true
					break
				}
				state.regs[Reg("R30")] = arm64ControlValue{"label:" + c.blocks[bi+1].name: true}
				if arm64IsLocalBranchLink(ins) {
					if name, ok := arm64BranchTarget(ins.Args[0]); ok {
						if target, ok := indices[name]; ok {
							propagate(target, state)
						}
					}
				} else {
					c.invalidateControlCallResults(state, ins.Args[0])
					propagate(bi+1, state)
				}
				terminated = true
				break
			}
			if op == "B" || op == "JMP" {
				if len(ins.Args) == 1 {
					if ins.Args[0].Kind == OpReg || ins.Args[0].Kind == OpMem {
						value := state.branchValue(ins.Args[0])
						plan.before[[2]int{bi, ii}] = value
						for token := range value {
							if strings.HasPrefix(token, "label:") {
								if target, ok := indices[token[6:]]; ok {
									propagate(target, state)
								}
							}
						}
					} else if name, ok := arm64BranchTarget(ins.Args[0]); ok {
						if target, ok := indices[name]; ok {
							propagate(target, state)
						}
					}
				}
				terminated = true
				break
			}
			if op == OpRET {
				if target, register := arm64RegisterReturnTarget(ins, plan.autoFrame); register {
					value := state.branchValue(target)
					plan.before[[2]int{bi, ii}] = value
					for token := range value {
						if strings.HasPrefix(token, "label:") {
							if target, ok := indices[token[6:]]; ok {
								propagate(target, state)
							}
						}
					}
				}
				terminated = true
				break
			}
			if arm64IsConditionalBranch(op) {
				if len(ins.Args) != 0 {
					if name, ok := arm64BranchTarget(ins.Args[len(ins.Args)-1]); ok {
						if target, ok := indices[name]; ok {
							propagate(target, state)
						}
					}
				}
				propagate(bi+1, state)
				terminated = true
				break
			}
			state.transfer(ins, op, post, c.rawDataGlobals)
		}
		if !terminated {
			propagate(bi+1, state)
		}
	}
	plan.reachable = make([]bool, len(before))
	for index, state := range before {
		plan.reachable[index] = state != nil
	}
	c.localControl = plan
	return nil
}

// A dead source terminator has neither an unknown native target nor a caller
// return. Retain its grammar validation, then terminate its LLVM block without
// fabricating execution. Other instructions in the block are still lowered.
func (c *arm64Ctx) lowerProvenUnreachableControl(bi int) bool {
	if c.localControl == nil || bi < 0 || bi >= len(c.localControl.reachable) || c.localControl.reachable[bi] {
		return false
	}
	c.b.WriteString("  unreachable\n")
	return true
}

// The implicit Go frame is separate from explicit source-level SP arithmetic.
// Go's RET epilogue pops a constant implicit frame; it does not repair a manual
// SP delta. LLVM also restores its native frame on return, so synthesizing an
// LLVM return is valid only when every reaching path restored the virtual SP.
func (c *arm64Ctx) requireCallerSPRestored(bi int, ins Instr) error {
	if c.localControl == nil {
		return fmt.Errorf("%w: ARM64 caller return has no SP reaching-definition proof: %q", ErrProbeNeedsContext, ins.Raw)
	}
	value := c.localControl.stackBefore[[2]int{bi, c.currentInstruction}]
	if len(value) != 1 || !value["sp:0"] {
		return fmt.Errorf("%w: ARM64 caller return requires restored entry SP on every reaching path: %q", ErrProbeNeedsContext, ins.Raw)
	}
	return nil
}

func (c *arm64Ctx) requireCallerLinkRestored(bi int, ins Instr) error {
	if c.localControl == nil {
		return fmt.Errorf("%w: ARM64 caller return has no LR reaching-definition proof: %q", ErrProbeNeedsContext, ins.Raw)
	}
	value := c.localControl.linkBefore[[2]int{bi, c.currentInstruction}]
	if len(value) != 1 || !value["label:"+c.localControl.outer] {
		return fmt.Errorf("%w: ARM64 caller return requires the caller LR on every reaching path: %q", ErrProbeNeedsContext, ins.Raw)
	}
	return nil
}

func (state *arm64ControlState) branchValue(op Operand) arm64ControlValue {
	reg := op.Reg
	if op.Kind == OpMem {
		reg = op.Mem.Base
		// Branch register encoding 31 reads XZR, never the memory-address SP.
		if reg == SP || reg == Reg("RSP") {
			return arm64ControlExternal()
		}
	}
	return arm64ControlRead(state.regs, reg)
}

func arm64RegisterReturnTarget(ins Instr, hasGoFrame bool) (Operand, bool) {
	if len(ins.Args) == 0 {
		return Operand{Kind: OpReg, Reg: Reg("R30")}, !hasGoFrame
	}
	if len(ins.Args) == 1 && ins.Args[0].Kind == OpImm {
		return Operand{Kind: OpReg, Reg: Reg("R30")}, !hasGoFrame
	}
	if len(ins.Args) != 1 || (ins.Args[0].Kind != OpReg && ins.Args[0].Kind != OpMem) {
		return Operand{}, false
	}
	target := ins.Args[0]
	reg := target.Reg
	if target.Kind == OpMem {
		reg = target.Mem.Base
		target.Mem.Off, target.Mem.OffRaw, target.Mem.Index = 0, "", ""
	}
	if reg == SP || reg == Reg("RSP") {
		target = Operand{Kind: OpReg, Reg: ZR}
	}
	return target, !hasGoFrame || reg != Reg("R30")
}

func arm64IsConditionalBranch(op Op) bool {
	switch op {
	case "BEQ", "BNE", "BLO", "BHI", "BLT", "BGE", "BLE", "BGT", "BHS", "BLS", "BMI", "BPL", "BVS", "BVC", "BCC", "BCS", "CBZ", "CBNZ", "CBZW", "CBNZW", "TBZ", "TBNZ":
		return true
	}
	return false
}

func (state *arm64ControlState) transfer(ins Instr, op Op, post bool, data map[string]string) {
	if len(ins.Args) == 0 {
		return
	}
	if op == OpWORD {
		writes, known := uint32(0), false
		stores := true
		if len(ins.Args) == 1 && ins.Args[0].Kind == OpImm && ins.Args[0].ImmRaw == "" {
			word := uint32(ins.Args[0].Imm)
			if effects, ok := arm64RawContinuationEffectsForWord(word); ok {
				writes, stores, known = effects.gpWrites, effects.stores, true
			} else if effects, ok := arm64RawSVEEffects(word); ok {
				writes, stores, known = effects.gpWrites, effects.stores, true
			} else {
				writes, known = arm64RawPoolGPWrites(word)
			}
		}
		for reg, value := range state.regs {
			index, gp := arm64StackIndex(reg)
			if gp && (!known || writes&(1<<uint(index)) != 0) {
				state.regs[reg] = arm64ControlUnion(value, arm64ControlExternal())
			}
		}
		if stores {
			for key, value := range state.memory {
				state.memory[key] = arm64ControlUnion(value, arm64ControlExternal())
			}
		}
		return
	}
	if form, handled, err := parseARM64AtomicPairForm(op, ins); handled && err == nil {
		state.transferAtomicPair(form)
		return
	}
	if op == "ADR" && len(ins.Args) == 2 {
		if target, ok := arm64BranchTarget(ins.Args[0]); ok && data[target] == "" {
			state.write(ins.Args[1], arm64ControlValue{"label:" + target: true}, false)
			return
		}
	}
	if form, handled, err := parseARM64RegisterAddressForm(op, ins); handled && err == nil {
		if form.usesScratch {
			state.regs[Reg("R27")] = arm64ControlExternal()
		}
		value := state.source(Operand{Kind: OpReg, Reg: form.base}, false)
		if form.offset != 0 {
			value = arm64ControlUnion(arm64ControlOffset(value, form.offset), arm64ControlLabels(value))
		}
		state.write(ins.Args[1], value, false)
		return
	}
	if op == "MOVD" && len(ins.Args) == 2 {
		value := state.source(ins.Args[0], post)
		state.write(ins.Args[1], value, post)
	} else if (op == "LDP" || op == "STP") && len(ins.Args) == 2 {
		memory, pair := ins.Args[0], ins.Args[1]
		if op == "STP" {
			pair, memory = ins.Args[0], ins.Args[1]
		}
		if pair.Kind != OpRegList || len(pair.RegList) != 2 {
			return // The typed operand grammar rejects the malformed pair.
		}
		if post && memory.Kind == OpMem {
			memory.Mem.Off = 0
		}
		var values [2]arm64ControlValue
		for index, reg := range pair.RegList {
			source := Operand{Kind: OpReg, Reg: reg}
			if op == "LDP" {
				source = arm64ControlElement(memory, int64(index)*8)
			}
			values[index] = state.source(source, false)
		}
		for index, reg := range pair.RegList {
			destination := arm64ControlElement(memory, int64(index)*8)
			if op == "LDP" {
				destination = Operand{Kind: OpReg, Reg: reg}
			}
			state.write(destination, values[index], false)
		}
	} else if (op == "ADD" || op == "SUB") && len(ins.Args) >= 2 && ins.Args[0].Kind == OpImm && ins.Args[0].ImmRaw == "" {
		src, dst := ins.Args[1], ins.Args[len(ins.Args)-1]
		off := ins.Args[0].Imm
		if op == "SUB" {
			if off == int64(-1<<63) {
				state.write(dst, arm64ControlExternal(), false)
				return
			}
			off = -off
		}
		value := state.source(src, false)
		if off != 0 {
			value = arm64ControlUnion(arm64ControlOffset(value, off), arm64ControlLabels(value))
		}
		state.write(dst, value, false)
	} else {
		switch op {
		case "CMP", "CMPW", "CMN", "CMNW", "TST", "TSTW", "CCMP", "CCMPW", "CCMN", "CCMNW", "PCDATA", "FUNCDATA":
			return
		}
		// Unmodeled transforms of code addresses are tainted, not native
		// addresses. This deliberately fails closed even if a bit trick could
		// happen to reconstitute the original blockaddress at runtime.
		value := arm64ControlExternal()
		for _, arg := range ins.Args {
			value = arm64ControlUnion(value, arm64ControlLabels(state.source(arg, post)))
			for _, reg := range arg.RegList {
				value = arm64ControlUnion(value, arm64ControlLabels(arm64ControlRead(state.regs, reg)))
			}
		}
		destination := ins.Args[len(ins.Args)-1]
		if key := state.memoryKey(destination, post); key != "" {
			state.invalidateOverlappingControlCells(key, 64)
			value = arm64ControlUnion(value, arm64ControlLabels(state.memory[key]))
		}
		state.write(destination, value, post)
	}
	if post || strings.HasSuffix(string(ins.Op), ".W") {
		for _, operand := range ins.Args {
			if operand.Kind == OpMem {
				base := arm64ControlBase(operand.Mem.Base)
				state.regs[base] = arm64ControlOffset(arm64ControlRead(state.regs, base), operand.Mem.Off)
			}
		}
	}
}

func (c *arm64Ctx) invalidateControlCallResults(state *arm64ControlState, operand Operand) {
	name := c.resolve(strings.TrimSuffix(operand.Sym, "(SB)"))
	sig, known := c.sigs[name]
	// A native/Go callee has no virtual-GP preservation contract. In
	// particular, even a scalar result may return the call's link address.
	possibleCode := arm64ControlLabels(state.escaped)
	for _, value := range state.regs {
		possibleCode = arm64ControlUnion(possibleCode, arm64ControlLabels(value))
	}
	returnValue := arm64ControlUnion(arm64ControlExternal(), possibleCode)
	mayWriteFrame := !known || len(state.escaped) != 0
	internal := strings.HasSuffix(strings.TrimSuffix(operand.Sym, "(SB)"), "<ABIInternal>")
	if internal && sig.ARM64GoRegisterABI != nil {
		// The classic ABI0 Frame is not the location of an explicit
		// ABIInternal call's values. Inspect every typed scalar register leaf,
		// including pointers nested within aggregates. This only invalidates
		// possible aliases; it asserts no ownership or register preservation.
		if err := arm64ValidateGoRegisterABI(sig); err != nil {
			mayWriteFrame = true
		} else {
			for _, param := range sig.ARM64GoRegisterABI.Params {
				value := arm64ControlRead(state.regs, param.Register)
				state.escaped = arm64ControlUnion(state.escaped, arm64ControlAddressTaint(value))
				returnValue = arm64ControlUnion(returnValue, arm64ControlLabels(value))
				mayWriteFrame = mayWriteFrame || param.Type == Ptr
				for token := range value {
					if strings.HasPrefix(token, "sp:") || strings.HasPrefix(token, "fp:") || strings.HasPrefix(token, "fpa:") {
						mayWriteFrame = true
					}
				}
			}
		}
	}
	for index := range sig.Args {
		if internal && sig.ARM64GoRegisterABI != nil {
			break
		}
		var value arm64ControlValue
		if len(sig.Frame.Params) != 0 && len(sig.ArgRegs) == 0 {
			for _, param := range sig.Frame.Params {
				if param.Index == index {
					value = arm64ControlUnion(value, state.source(Operand{Kind: OpMem, Mem: MemRef{Base: SP, Off: param.Offset + 8}}, false))
				}
			}
		} else {
			reg := Reg(fmt.Sprintf("R%d", index))
			if index < len(sig.ArgRegs) {
				reg = sig.ArgRegs[index]
			}
			value = arm64ControlRead(state.regs, reg)
		}
		state.escaped = arm64ControlUnion(state.escaped, arm64ControlAddressTaint(value))
		returnValue = arm64ControlUnion(returnValue, arm64ControlLabels(value))
		if sig.Args[index] == Ptr {
			mayWriteFrame = true
		}
		for token := range value {
			if strings.HasPrefix(token, "sp:") || strings.HasPrefix(token, "fp:") || strings.HasPrefix(token, "fpa:") {
				mayWriteFrame = true
			}
		}
	}
	for index := 0; index < 31; index++ {
		reg := Reg(fmt.Sprintf("R%d", index))
		if reg != Reg("R30") {
			state.regs[reg] = arm64ControlUnion(arm64ControlRead(state.regs, reg), returnValue)
		}
	}
	if mayWriteFrame {
		for key, value := range state.memory {
			state.memory[key] = arm64ControlUnion(value, returnValue)
		}
	}
	if !known {
		return
	}
	if internal && sig.ARM64GoRegisterABI != nil {
		for _, result := range sig.ARM64GoRegisterABI.Results {
			if !isARM64ABIFloatingType(result.Type) {
				state.regs[result.Register] = returnValue
			}
		}
		return
	}
	if len(sig.Frame.Results) == 0 {
		if sig.Ret != Void {
			state.regs[Reg("R0")] = returnValue
		}
		return
	}
	for _, result := range sig.Frame.Results {
		key := state.memoryKey(Operand{Kind: OpMem, Mem: MemRef{Base: SP, Off: result.Offset + 8}}, false)
		if key != "" {
			state.memory[key] = returnValue
		}
	}
}

func (c *arm64Ctx) localControlTargets(bi int) ([]string, error) {
	if c.localControl == nil {
		return nil, nil
	}
	value := c.localControl.before[[2]int{bi, c.currentInstruction}]
	var targets []string
	external := false
	for token := range value {
		if strings.HasPrefix(token, "label:") {
			targets = append(targets, token[6:])
		} else {
			external = true
		}
	}
	if len(targets) != 0 && external {
		return nil, fmt.Errorf("%w: ARM64 register branch has mixed local-code and external/unknown reaching definitions", ErrProbeNeedsContext)
	}
	if len(targets) > 1 && value["label:"+c.localControl.outer] {
		return nil, fmt.Errorf("%w: ARM64 register branch mixes a native caller link with local blockaddresses", ErrProbeNeedsContext)
	}
	sort.Strings(targets)
	return targets, nil
}
