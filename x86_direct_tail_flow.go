package plan9asm

import (
	"fmt"
	"strings"
)

func (graph *x86DirectTailGraph) root(root int, byName map[string]int) (*x86DirectTailRoot, error) {
	members := map[int]bool{root: true}
	queue := []int{root}
	for len(queue) != 0 {
		at := queue[0]
		queue = queue[1:]
		fn := graph.original.Funcs[at]
		if graph.goarch == "386" && strings.HasSuffix(fn.Sym, "<ABIInternal>") {
			return nil, x86DirectTailContext("386 source has no AMD64 register-entry contract")
		}
		if fn.FrameSize != 0 || len(fn.X86RawText) != 0 || fn.x86SourceStack != nil || fn.x86FixedLocal != nil ||
			fn.x86RawReplay != nil || len(fn.x86ContinuationAddresses) != 0 || len(fn.x86IndirectLabels) != 0 ||
			graph.helpers[fn.Sym] && fn.ArgSize != 0 {
			return nil, x86DirectTailContext("native/frame/helper argument role is not closed: %s", fn.Sym)
		}
		for _, ins := range fn.Instrs {
			if isX86RawDirective(ins) || ins.x86Encoded || ins.x86SourceRaw != nil || ins.x86RawNearReturn != nil ||
				ins.x86AddressBits != 0 || ins.x86UnboundAddressOverride || ins.x86SegmentPrefix != 0 ||
				len(ins.x86SourceStackFP) != 0 || len(ins.x86CallerMemory) != 0 || len(ins.x86FixedLocalMemory) != 0 ||
				ins.x86NumericalRead != nil || ins.x86SDKLiteral != nil ||
				ins.Op == "CALL" {
				return nil, x86DirectTailContext("raw/call/native effect has no continuation role: %s", ins.Raw)
			}
			for _, arg := range ins.Args {
				if arg.Kind == OpReg && (arg.Reg == SP || arg.Reg == BP) ||
					arg.Kind == OpMem && (arg.Mem.Base == SP || arg.Mem.Base == BP || arg.Mem.Base == "PC" || arg.Mem.Index == SP || arg.Mem.Index == BP) ||
					arg.Kind == OpFPAddr {
					return nil, x86DirectTailContext("physical SP/BP/PC/frame address has no continuation role: %s", ins.Raw)
				}
			}
			if ins.Op != OpJMP || len(ins.Args) != 1 || ins.Args[0].Kind != OpSym {
				continue
			}
			name, offset, valid := parseSBRef(ins.Args[0].Sym)
			if !valid || offset != 0 || !graph.helpers[name] {
				return nil, x86DirectTailContext("indirect/external/missing/offset tail target: %s", ins.Raw)
			}
			index, found := byName[name]
			if !found {
				return nil, x86DirectTailContext("tail target is not an original source TEXT")
			}
			if !members[index] {
				members[index] = true
				queue = append(queue, index)
			}
		}
	}
	view := &x86DirectTailRoot{graph: graph, text: root}
	starts := make(map[int]int)
	labels := make(map[int]map[string]int)
	order := []int{root}
	for index := range graph.original.Funcs {
		if index != root && members[index] {
			order = append(order, index)
		}
	}
	for _, index := range order {
		starts[index] = len(view.points)
		labels[index] = make(map[string]int)
		for ordinal, ins := range graph.original.Funcs[index].Instrs {
			if ins.Op == OpTEXT {
				continue
			}
			if ins.Op == OpLABEL {
				if len(ins.Args) != 1 || ins.Args[0].Kind != OpLabel {
					return nil, x86DirectTailContext("malformed original local label")
				}
				label := ins.Args[0].Sym
				if _, repeated := labels[index][label]; repeated {
					return nil, x86DirectTailContext("duplicate original local label")
				}
				labels[index][label] = len(view.points)
				continue
			}
			view.points = append(view.points, x86DirectTailPoint{text: index, instruction: ordinal})
			view.instructions = append(view.instructions, ins)
			view.ordinals = append(view.ordinals, ordinal)
		}
		if starts[index] == len(view.points) || len(view.points) > 4096 {
			return nil, x86DirectTailContext("empty/oversized source continuation")
		}
	}
	view.edges = make([][]int, len(view.points))
	for pc, ins := range view.instructions {
		if ins.Op == OpRET {
			if len(ins.Args) != 0 {
				return nil, x86DirectTailContext("nonordinary source RET")
			}
			continue
		}
		next := []int{pc + 1}
		if ins.Op == OpJMP || isAMD64ConditionalBranch(ins.Op) {
			args := ins.Args
			if ins.Op != OpJMP && !isAMD64CounterZeroBranch(ins.Op) && len(args) == 2 && args[0].Kind == OpImm &&
				args[0].ImmRaw == "" && !args[0].ImmIsFloat && (args[0].Imm == 0 || args[0].Imm == 1) {
				args = args[1:]
			}
			if len(args) != 1 {
				return nil, x86DirectTailContext("unbound source branch operands")
			}
			arg := args[0]
			label := arg.Sym
			if arg.Kind == OpIdent {
				label = arg.Ident
			} else if arg.Kind == OpReg {
				label = string(arg.Reg)
			}
			target, found := labels[view.points[pc].text][label]
			if ins.Op == OpJMP && arg.Kind == OpSym {
				name, offset, valid := parseSBRef(arg.Sym)
				index, exists := byName[name]
				if !valid || !exists || !members[index] || !graph.helpers[name] || offset != 0 {
					return nil, x86DirectTailContext("unbound private tail node")
				}
				target, found = starts[index], true
			}
			if !found || target >= len(view.points) || view.points[target].text != view.points[pc].text &&
				!(ins.Op == OpJMP && arg.Kind == OpSym) {
				return nil, x86DirectTailContext("indirect/external/missing source branch")
			}
			if ins.Op == OpJMP {
				next = []int{target}
			} else {
				next = append(next, target)
			}
		}
		for index, target := range next {
			if target < 0 || target >= len(view.points) || index == 0 && ins.Op != OpJMP && view.points[target].text != view.points[pc].text {
				return nil, x86DirectTailContext("source falls outside its original TEXT continuation")
			}
		}
		view.edges[pc] = next
	}
	return view, nil
}

func (view *x86DirectTailRoot) blocks() []amd64Block {
	blocks := make([]amd64Block, len(view.points))
	for pc, point := range view.points {
		blocks[pc] = amd64Block{name: fmt.Sprintf("p9_tail_%d_%d", point.text, point.instruction),
			instrs: []Instr{view.instructions[pc]}, synthetic: true}
	}
	return blocks
}

func (view *x86DirectTailRoot) features() Func {
	fn := view.graph.original.Funcs[view.text]
	fn.Instrs = view.instructions // A cost/feature projection, never new source.
	return fn
}

func (file *File) x86DirectTailHelper(symbol string) bool {
	return file.x86DirectTails != nil && file.x86DirectTails.helpers[symbol]
}

func (file *File) x86DirectTailView(symbol string) *x86DirectTailRoot {
	if file.x86DirectTails == nil {
		return nil
	}
	return file.x86DirectTails.roots[symbol]
}

func x86DirectTailSigNeedsOwnBoolFrame(sig FuncSig, fn Func, goarch string) bool {
	if goarch == "amd64" && strings.HasSuffix(fn.Sym, "<ABIInternal>") {
		return false // The existing canonical register proof still applies.
	}
	for _, slot := range sig.Frame.Results {
		if slot.Type == I1 {
			return true
		}
	}
	return sig.Ret == I1
}
