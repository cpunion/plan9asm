package plan9asm

import (
	"reflect"
	"strconv"
	"strings"
)

// Raw leaf/C bodies can own an explicit stack allocation without a Go
// prologue. This is a separate whole-source proof: it does not invent a Go
// frame, turn arbitrary native calls into helpers, or weaken caller LR checks.
// Every local raw BL is interpreted in its bounded call context; every root
// return restores the entry SP, LR and BP. Incoming pointers cannot alias the
// fresh LLVM stack object unless an address escapes this checked source.
type arm64RawZeroState struct {
	addresses *arm64RawPrivateFrameState
	control   *arm64ControlState
}

func (s *arm64RawZeroState) clone() *arm64RawZeroState {
	return &arm64RawZeroState{addresses: s.addresses.clone(), control: s.control.clone()}
}

func (s *arm64RawZeroState) merge(other *arm64RawZeroState) (bool, bool) {
	if s.addresses.sp != other.addresses.sp {
		return false, false
	}
	changed := s.addresses.regs.merge(other.addresses.regs)
	spills := make(map[int64]arm64PrivateFrameAddress)
	for offset, value := range s.addresses.spills {
		if incoming, present := other.addresses.spills[offset]; !present || incoming != value {
			value = arm64PrivateFrameAddress{}
		}
		spills[offset] = value
	}
	for offset := range other.addresses.spills {
		if _, present := spills[offset]; !present {
			spills[offset] = arm64PrivateFrameAddress{}
		}
	}
	if !reflect.DeepEqual(spills, s.addresses.spills) {
		s.addresses.spills, changed = spills, true
	}
	return s.control.merge(other.control) || changed, true
}

type arm64RawZeroCall struct {
	block int
	sp    int64
}

type arm64RawZeroPoint struct {
	block   int
	context string
}

func arm64RawZeroFrameUnexposed(fn Func) bool {
	if fn.FrameSize != 0 || arm64SourceGoFrame(fn).present || !arm64SourceFrameFlagsKnown(fn) || !arm64HasRawFrameCode(fn) {
		return false
	}
	text := 0
	for _, ins := range fn.Instrs {
		if ins.Op == OpTEXT {
			text++
		}
	}
	if text != 1 {
		return false
	}
	// The exact raw PC boundary checker supplies only same-source labels.
	// Run it on the original too: later normalization cannot erase an escape.
	normalized, blobs, err := prepareARM64RawPCRelative(fn)
	if err != nil || len(blobs) != 0 {
		return false
	}
	blocks := arm64SplitBlocks(normalized)
	indices := make(map[string]int, len(blocks))
	for index, block := range blocks {
		indices[block.name] = index
	}
	before := make(map[arm64RawZeroPoint]*arm64RawZeroState)
	contexts := map[string][]arm64RawZeroCall{"": nil}
	visited := make([]bool, len(blocks))
	queue := make([]arm64RawZeroPoint, 0, len(blocks))
	queued := make(map[arm64RawZeroPoint]bool)
	propagate := func(point arm64RawZeroPoint, state *arm64RawZeroState) bool {
		if point.block < 0 || point.block >= len(blocks) || len(before) >= 16384 {
			return false
		}
		changed := false
		if old := before[point]; old == nil {
			before[point], changed = state.clone(), true
		} else {
			var valid bool
			changed, valid = old.merge(state)
			if !valid {
				return false
			}
		}
		if changed && !queued[point] {
			queue, queued[point] = append(queue, point), true
		}
		return true
	}
	operations := 0
	for seed := range blocks {
		if visited[seed] {
			continue
		}
		entry := &arm64RawZeroState{
			addresses: &arm64RawPrivateFrameState{
				regs:   arm64PrivateFrameAddresses{"R29": {}},
				spills: make(map[int64]arm64PrivateFrameAddress),
			},
			control: &arm64ControlState{
				regs: map[Reg]arm64ControlValue{
					SP: {"sp:0": true}, "R29": {"entry-bp:": true}, "R30": {"label:__entry_lr": true},
				},
				memory: make(map[string]arm64ControlValue), unexposedSP: true,
			},
		}
		if !propagate(arm64RawZeroPoint{block: seed}, entry) {
			return false
		}
		for next := 0; next < len(queue); next++ {
			point := queue[next]
			queued[point] = false
			visited[point.block] = true
			state := before[point].clone()
			calls := contexts[point.context]
			terminated := false
			for _, original := range blocks[point.block].instrs {
				operations++
				if operations > 1<<24 {
					return false
				}
				ins := arm64ControlDecode(original)
				op := arm64ControlOp(ins)
				if arm64RawZeroReturn(original) {
					link := arm64ControlRead(state.control.regs, "R30")
					if len(calls) == 0 {
						bp := arm64ControlRead(state.control.regs, "R29")
						if state.addresses.sp != 0 || len(link) != 1 || !link["label:__entry_lr"] || len(bp) != 1 || !bp["entry-bp:"] {
							return false
						}
						delete(state.addresses.regs, "R29")
						if len(state.addresses.regs) != 0 {
							return false
						}
					} else {
						caller := calls[len(calls)-1]
						label := "label:" + blocks[caller.block].name
						if state.addresses.sp != caller.sp || len(link) != 1 || !link[label] {
							return false
						}
						parent := point.context[:strings.LastIndex(point.context, "/")]
						if !propagate(arm64RawZeroPoint{block: caller.block, context: parent}, state) {
							return false
						}
					}
					terminated = true
					break
				}
				if op == "BL" || op == "CALL" || op == "BLR" {
					if !original.arm64RawLocalCall || len(ins.Args) != 1 || len(calls) >= 16 || point.block+1 >= len(blocks) {
						return false
					}
					name, local := arm64BranchTarget(ins.Args[0])
					target, exists := indices[name]
					if !local || !exists || !arm64IsLocalBranchLink(ins) {
						return false
					}
					context := point.context + "/" + strconv.Itoa(point.block+1)
					contexts[context] = append(append([]arm64RawZeroCall(nil), calls...), arm64RawZeroCall{block: point.block + 1, sp: state.addresses.sp})
					state.control.regs["R30"] = arm64ControlValue{"label:" + blocks[point.block+1].name: true}
					if !propagate(arm64RawZeroPoint{block: target, context: context}, state) {
						return false
					}
					terminated = true
					break
				}
				if op == "B" || op == "JMP" || arm64IsConditionalBranch(op) {
					if len(ins.Args) == 0 {
						return false
					}
					name, local := arm64BranchTarget(ins.Args[len(ins.Args)-1])
					target, exists := indices[name]
					if !local || !exists || !propagate(arm64RawZeroPoint{block: target, context: point.context}, state) {
						return false
					}
					if arm64IsConditionalBranch(op) && !propagate(arm64RawZeroPoint{block: point.block + 1, context: point.context}, state) {
						return false
					}
					terminated = true
					break
				}
				if op == "ADR" || op == "ADRP" || op == OpRET || op == arm64RawDataOp || !state.transfer(original, ins) {
					return false
				}
			}
			if !terminated && !propagate(arm64RawZeroPoint{block: point.block + 1, context: point.context}, state) {
				return false
			}
		}
		queue = queue[:0]
	}
	return true
}

func arm64RawZeroReturn(original Instr) bool {
	if original.Op == OpWORD {
		decoded, err := decodeARM64RawWordInstruction(original)
		return err == nil && arm64RawPrivateSourceRET(decoded)
	}
	return arm64RawPrivateSourceRET(original)
}

func (s *arm64RawZeroState) transfer(original, ins Instr) bool {
	if ins.Op == "NOP" && len(ins.Args) == 0 && (original.Op != OpWORD ||
		len(original.Args) == 1 && original.Args[0].Kind == OpImm && original.Args[0].ImmRaw == "" && original.Args[0].Imm == 0xd503201f) {
		return true // Exact architectural NOP, not the wider system/hint class.
	}
	if form, known := arm64RawPrivateMemory(ins); known && form.memory.Kind == OpMem {
		mem := form.memory.Mem
		if address, private := s.addresses.address(mem.Base); private {
			offset := mem.Off
			if form.post {
				offset = 0
			}
			address, valid := arm64PrivateFrameOffset(address, offset)
			low := s.addresses.sp
			if form.pre && arm64StackReg(mem.Base) {
				low = address.offset
			}
			width := form.elementBytes * int64(len(form.registers))
			if !valid || !address.exact || mem.Index != "" || mem.OffRaw != "" || address.offset < low || address.offset > -width {
				return false // No implicit Go cell or unallocated red zone exists.
			}
		}
	}
	if !s.addresses.transfer(original, ins, -8) {
		return false
	}
	s.control.transfer(ins, arm64ControlOp(ins), strings.HasSuffix(string(ins.Op), ".P"), nil)
	return true
}
