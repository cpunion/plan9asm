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
	copyFrom string
	width    int
	write    bool
	call     *ARM64GoRegisterABI
	opaque   bool
	store    bool
	source   string
}

type arm64MachineBlock struct {
	name       string
	effects    []arm64MachineEffect
	successors []string
}

type arm64MachineAvailability struct {
	blocks     []arm64MachineBlock
	current    int
	suspended  int
	skipReads  bool
	source     string
	entry      *ARM64GoRegisterABI
	entryError error
	used       bool
	nativeIR   map[int]arm64MachineNativeProof
}

type arm64MachineNativeProof struct {
	line  string
	store bool
}

// A validated lowerer can prove that one emitted native instruction has no
// additional virtual GP/NZCV/data-store effects. Its actual typed operands are
// still recorded by loadReg/storeReg. Bind the proof to the exact emitted IR
// line; other callouts, including another callout in the same source
// instruction, remain opaque. A compiler memory barrier is not a data store.
func (c *arm64Ctx) emitMachineNeutralNativeIR(format string, args ...any) {
	c.emitMachineNativeIR(false, format, args...)
}

// Store-native IR has a separate continuation memory proof. It is not a
// neutral instruction: callers of this helper must validate the address
// sources, and the matching control transfer must invalidate affected cells.
func (c *arm64Ctx) emitMachineStoreNativeIR(format string, args ...any) {
	c.emitMachineNativeIR(true, format, args...)
}

func (c *arm64Ctx) emitMachineNativeIR(store bool, format string, args ...any) {
	start := c.b.Len()
	fmt.Fprintf(c.b, format, args...)
	if flow := c.machineAvailability; flow != nil {
		if flow.nativeIR == nil {
			flow.nativeIR = make(map[int]arm64MachineNativeProof)
		}
		flow.nativeIR[start] = arm64MachineNativeProof{line: c.b.String()[start:], store: store}
	}
}

const arm64WholeScalableRegister = 1 << 20

func newARM64MachineAvailability(c *arm64Ctx) *arm64MachineAvailability {
	needed := c.goRegisterEntry || c.privateRegisterEntry
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
	} else if c.privateRegisterEntry {
		// Explicit empty presence: only SP and the real caller LR are
		// available at an ABI0 entry. FP loads define the data registers.
		flow.entry, flow.used = &ARM64GoRegisterABI{}, true
		if len(c.sig.ArgRegs) != 0 {
			// Retain custom ArgRegs' scalar len/type contract, rather than
			// applying standard Go recursive ABIInternal assignment to it.
			if len(c.sig.ArgRegs) != len(c.sig.Args) {
				flow.entryError = arm64GoABIContext("custom private root entry has incomplete ArgRegs")
			} else {
				for i, reg := range c.sig.ArgRegs {
					if arm64MachineScalarWidth(c.sig.Args[i]) == 0 {
						flow.entryError = arm64GoABIContext("custom private root entry has unknown register width")
					}
					flow.entry.Params = append(flow.entry.Params, ARM64GoRegisterValue{Register: reg, Type: c.sig.Args[i]})
				}
			}
		}
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
	if width <= 0 {
		return "", arm64GoABIContext("unrecognized typed register read width %d", width)
	}
	c.recordMachineRegister(reg, width, false)
	if flow := c.machineAvailability; flow != nil {
		flow.suspended++
		defer func() { flow.suspended-- }()
	}
	return c.loadReg(reg)
}

func arm64MachineLeafWidth(slot ARM64GoRegisterValue) int {
	return arm64MachineScalarWidth(slot.Type)
}

func arm64MachineScalarWidth(typ LLVMType) int {
	switch typ {
	case I1:
		return 1
	case I8:
		return 8
	case I16:
		return 16
	case I32, "float":
		return 32
	case I64, Ptr, "double":
		return 64
	default:
		return 0
	}
}

// A full-register move transports unspecified bits without making them
// available to a typed caller. Its reaching low-bit guarantee is unchanged.
// This hook is used only after the scalar move lowerer validates both operands.
func (c *arm64Ctx) lowerMachineRegisterCopy(src, dst Reg) error {
	if flow := c.machineAvailability; flow != nil {
		flow.blocks[flow.current].effects = append(flow.blocks[flow.current].effects, arm64MachineEffect{
			register: arm64MachineRegister(dst), copyFrom: arm64MachineRegister(src),
			width: 64, source: flow.source,
		})
		flow.suspended++
		defer func() { flow.suspended-- }()
	}
	value, err := c.loadReg(src)
	if err != nil {
		return err
	}
	return c.storeReg(dst, value)
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
	if flow.entryError != nil {
		return flow.entryError
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
			if effect.store {
				// The source continuation analysis, not GP availability, owns
				// memory/return-link provenance. This effect is explicit so a
				// validated store cannot silently acquire a neutral proof.
				continue
			}
			if effect.copyFrom != "" {
				width := state.values[effect.copyFrom]
				if effect.copyFrom == string(ZR) {
					width = 64
				}
				if width == 0 {
					return arm64GoABIContext("%s copy has no supplied input at %q", effect.copyFrom, effect.source)
				}
				if effect.register != string(ZR) {
					if width > effect.width {
						width = effect.width
					}
					state.values[effect.register] = width
				}
			} else if effect.write {
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
	offset := start
	for _, line := range strings.SplitAfter(c.b.String()[start:], "\n") {
		proof := flow.nativeIR[offset]
		proven := proof.line == line && line != ""
		offset += len(line)
		if proven {
			if proof.store {
				flow.blocks[flow.current].effects = append(flow.blocks[flow.current].effects,
					arm64MachineEffect{store: true, source: flow.source})
			}
			continue
		}
		if strings.Contains(line, " asm ") || (!symbolCall && strings.Contains(line, " call ") && !strings.Contains(line, "@llvm.")) {
			flow.blocks[flow.current].effects = append(flow.blocks[flow.current].effects, arm64MachineEffect{opaque: true, source: flow.source})
			return
		}
	}
}
