package plan9asm

import (
	"fmt"
	"strings"
)

// Availability is distinct from the code-address/alias proof. Typed Go calls
// transfer only declared argument/result leaves, not a hidden native machine
// state. Record real lowering reads/writes instead of guessing opcode effects.
type arm64MachineEffect struct {
	register string
	width    int
	write    bool
	call     *ARM64GoRegisterABI
	opaque   bool
	source   string
}

type arm64MachineBlock struct {
	name       string
	effects    []arm64MachineEffect
	successors []string
}

type arm64MachineAvailability struct {
	blocks    []arm64MachineBlock
	current   int
	suspended int
	skipReads bool
	source    string
	entry     *ARM64GoRegisterABI
	used      bool
}

const arm64WholeScalableRegister = 1 << 20

func newARM64MachineAvailability(c *arm64Ctx) *arm64MachineAvailability {
	needed := c.goRegisterEntry
	for _, block := range c.blocks {
		for _, instruction := range block.instrs {
			ins := arm64ControlDecode(instruction)
			op := arm64ControlOp(ins)
			if (op != "CALL" && op != "BL") || len(ins.Args) != 1 || ins.Args[0].Kind != OpSym {
				continue
			}
			sym := strings.TrimSuffix(ins.Args[0].Sym, "(SB)")
			if strings.HasSuffix(sym, "<ABIInternal>") && c.sigs[c.resolve(sym)].ARM64GoRegisterABI != nil {
				needed = true
			}
		}
	}
	if !needed {
		return nil
	}
	flow := &arm64MachineAvailability{blocks: make([]arm64MachineBlock, len(c.blocks)+1)}
	for i, block := range c.blocks {
		flow.blocks[i].name = block.name
	}
	flow.blocks[len(c.blocks)].name = c.localControl.outer
	if c.goRegisterEntry {
		flow.entry, flow.used = c.sig.ARM64GoRegisterABI, true
	}
	return flow
}

func arm64MachineRegister(reg Reg) string {
	if index, ok := arm64ParseFReg(reg); ok {
		return fmt.Sprintf("V%d", index)
	}
	if index, ok := arm64ParseVReg(reg); ok {
		return fmt.Sprintf("V%d", index)
	}
	if index, ok := arm64ParseZReg(reg); ok {
		return fmt.Sprintf("V%d", index)
	}
	if reg == "RSP" {
		return string(SP)
	}
	return string(reg)
}

func (c *arm64Ctx) recordMachineRegister(reg Reg, width int, write bool) {
	flow := c.machineAvailability
	if flow == nil || flow.suspended != 0 || reg == ZR || (!write && flow.skipReads) {
		return
	}
	flow.blocks[flow.current].effects = append(flow.blocks[flow.current].effects, arm64MachineEffect{
		register: arm64MachineRegister(reg), width: width, write: write, source: flow.source,
	})
}

func (c *arm64Ctx) recordMachineCall(abi *ARM64GoRegisterABI) {
	if flow := c.machineAvailability; flow != nil {
		if abi == nil {
			flow.blocks[flow.current].effects = append(flow.blocks[flow.current].effects, arm64MachineEffect{opaque: true, source: flow.source})
		} else {
			flow.used = true
			flow.blocks[flow.current].effects = append(flow.blocks[flow.current].effects, arm64MachineEffect{call: abi, source: flow.source})
		}
	}
}

func (c *arm64Ctx) loadRegisterWidth(reg Reg, width int) (string, error) {
	c.recordMachineRegister(reg, width, false)
	if flow := c.machineAvailability; flow != nil {
		flow.suspended++
		defer func() { flow.suspended-- }()
	}
	return c.loadReg(reg)
}

func arm64MachineLeafWidth(slot ARM64GoRegisterValue) int {
	if slot.Type == "float" {
		return 32
	}
	return 64
}

// Only proven register XOR zero idioms need no input value. Their ordinary
// typed lowerer still validates the complete operand grammar. Never infer this
// for memory operands, or for unequal vector views/lanes.
func arm64MachineZeroInputs(ins Instr) bool {
	ins = arm64ControlDecode(ins)
	if ins.Op != "EOR" && ins.Op != "EORW" && ins.Op != "VEOR" {
		return false
	}
	if len(ins.Args) != 2 && len(ins.Args) != 3 {
		return false
	}
	return ins.Args[0].Kind == OpReg && ins.Args[1].Kind == OpReg && ins.Args[0].Reg == ins.Args[1].Reg
}

type arm64MachineState struct {
	active bool
	values map[string]int
}

func arm64MachineBoundary(slots []ARM64GoRegisterValue) arm64MachineState {
	state := arm64MachineState{active: true, values: map[string]int{string(SP): 64, "R30": 64}}
	for _, slot := range slots {
		state.values[arm64MachineRegister(slot.Register)] = arm64MachineLeafWidth(slot)
	}
	return state
}

func (state arm64MachineState) clone() arm64MachineState {
	out := arm64MachineState{active: state.active, values: make(map[string]int, len(state.values))}
	for reg, width := range state.values {
		out.values[reg] = width
	}
	return out
}

// Join uses the minimum defined width on every incoming active path. A path
// without a typed entry/call restriction does not provide an extra promise.
func (state *arm64MachineState) merge(other arm64MachineState) bool {
	if !other.active {
		return false
	}
	if !state.active {
		*state = other.clone()
		return true
	}
	changed := false
	for reg, width := range state.values {
		if next := other.values[reg]; next < width {
			state.values[reg] = next
			changed = true
		}
	}
	return changed
}

func (flow *arm64MachineAvailability) validate() error {
	if flow == nil || !flow.used {
		return nil
	}
	indices := make(map[string]int, len(flow.blocks))
	for i, block := range flow.blocks {
		indices[block.name] = i
	}
	entry := arm64MachineState{values: make(map[string]int)}
	if flow.entry != nil {
		entry = arm64MachineBoundary(flow.entry.Params)
	}
	before := make([]*arm64MachineState, len(flow.blocks))
	before[0] = &entry
	queue, queued := []int{0}, make([]bool, len(flow.blocks))
	queued[0] = true
	for next := 0; next < len(queue); next++ {
		index := queue[next]
		queued[index] = false
		state := before[index].clone()
		block := flow.blocks[index]
		for _, effect := range block.effects {
			if effect.call != nil {
				state = arm64MachineBoundary(effect.call.Results)
				continue
			}
			if !state.active {
				continue
			}
			if effect.opaque {
				return arm64GoABIContext("native/raw effects need a complete machine-state contract at %q", effect.source)
			}
			if effect.write {
				state.values[effect.register] = effect.width
			} else if state.values[effect.register] < effect.width {
				return arm64GoABIContext("%s read (%d bits) is not supplied by entry, call results or a reaching write at %q", effect.register, effect.width, effect.source)
			}
		}
		for _, target := range block.successors {
			i, ok := indices[target]
			if !ok {
				return arm64GoABIContext("machine-state CFG has an unknown successor %q", target)
			}
			changed := before[i] == nil
			if changed {
				copy := state.clone()
				before[i] = &copy
			} else {
				changed = before[i].merge(state)
			}
			if changed && !queued[i] {
				queue, queued[i] = append(queue, i), true
			}
		}
	}
	return nil
}

func (c *arm64Ctx) recordMachineOpaqueIR(start int, original Instr) {
	flow := c.machineAvailability
	if flow == nil {
		return
	}
	ins := arm64ControlDecode(original)
	op := arm64ControlOp(ins)
	// Symbol calls are recorded at their typed boundary. Other external
	// callouts (e.g. an unmodeled syscall/native helper) cannot inherit the
	// source machine state just because their IR is a normal LLVM call.
	symbolCall := (op == "CALL" || op == "BL" || op == "B" || op == "JMP" || op == OpRET) &&
		len(ins.Args) == 1 && ins.Args[0].Kind == OpSym && strings.HasSuffix(ins.Args[0].Sym, "(SB)")
	for _, line := range strings.Split(c.b.String()[start:], "\n") {
		if strings.Contains(line, " asm ") || (!symbolCall && strings.Contains(line, " call ") && !strings.Contains(line, "@llvm.")) {
			flow.blocks[flow.current].effects = append(flow.blocks[flow.current].effects, arm64MachineEffect{opaque: true, source: flow.source})
			return
		}
	}
}
