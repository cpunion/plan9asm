package plan9asm

import (
	"fmt"
	"sort"
	"strings"
)

// This is a source continuation contract, not a guessed callable signature.
// Only a closed, straight-line, zero-frame register leaf can share its caller's
// typed GP/NZCV state. Purity here restricts operand classes; real lowering still
// validates every opcode/form and records actual register/flag effects. Native
// asm/callouts are rejected by the machine-availability proof.
func arm64PrivateRegisterLeaf(fn Func) error {
	if fn.FrameSize != 0 && fn.FrameSize != -8 || fn.ArgSize != 0 || arm64SourceGoFrame(fn).present {
		return arm64GoABIContext("private helper %q has an unproved native frame", fn.Sym)
	}
	if len(fn.Instrs) < 2 || fn.Instrs[0].Op != OpTEXT {
		return arm64GoABIContext("private helper %q has no complete source body", fn.Sym)
	}
	if len(arm64SplitBlocks(fn)) != 1 {
		return arm64GoABIContext("private helper %q is not a straight-line register leaf", fn.Sym)
	}
	last := fn.Instrs[len(fn.Instrs)-1]
	if last.Op != OpRET || len(last.Args) != 0 {
		return arm64GoABIContext("private helper %q needs a zero-frame source RET", fn.Sym)
	}
	for _, ins := range fn.Instrs[1 : len(fn.Instrs)-1] {
		gp := false
		for _, arg := range ins.Args {
			switch arg.Kind {
			case OpImm:
				if arg.ImmRaw != "" {
					return unresolvedSymbolicImmediateError(arg)
				}
			case OpReg, OpRegShift, OpRegExtend:
				if !arm64PrivateDataGP(arg.Reg) || arg.ShiftReg != "" && !arm64PrivateDataGP(arg.ShiftReg) {
					return arm64GoABIContext("private helper %q touches a non-data register at %q", fn.Sym, ins.Raw)
				}
				gp = true
			case OpIdent:
				// Conditional integer operations have a named NZCV condition,
				// never an address, system register or hidden source label.
				if !armCondCodes[arg.Ident] {
					return arm64GoABIContext("private helper %q has a non-register operand at %q", fn.Sym, ins.Raw)
				}
			default:
				return arm64GoABIContext("private helper %q has memory/control operands at %q", fn.Sym, ins.Raw)
			}
		}
		if !gp {
			return arm64GoABIContext("private helper %q has unproved non-GP effects at %q", fn.Sym, ins.Raw)
		}
	}
	return nil
}

func arm64PrivateDataGP(reg Reg) bool {
	if reg == ZR {
		return true
	}
	// R18 is a physical platform register even after the source spelling
	// R18_PLATFORM has been normalized. R27 is the Go assembler scratch
	// register; R28/R29/R30 have hidden
	// runtime/frame/link roles. No source-private data promise is made for them.
	return isARM64GeneralOrZeroReg(reg) && !arm64PrivateReservedGP(reg) && reg != "R28" && reg != "R29" && reg != "R30"
}

func arm64PrivateReservedGP(reg Reg) bool {
	if root, ok := parseReg(string(reg)); ok {
		reg = root
	}
	return reg == "R18" || reg == "R27"
}

func arm64PrivateOperandMentionsReserved(arg Operand) bool {
	if arm64PrivateReservedGP(arg.Reg) || arm64PrivateReservedGP(arg.ShiftReg) || arm64PrivateReservedGP(arg.Mem.Base) || arm64PrivateReservedGP(arg.Mem.Index) {
		return true
	}
	for _, reg := range arg.RegList {
		if arm64PrivateReservedGP(reg) {
			return true
		}
	}
	if arg.Kind == OpSym && strings.HasPrefix(arg.Sym, "$") {
		address, ok := parseMem(strings.TrimSpace(strings.TrimPrefix(arg.Sym, "$")))
		if ok && (arm64PrivateReservedGP(address.Base) || arm64PrivateReservedGP(address.Index)) {
			return true
		}
	}
	return false
}

func arm64PrivateSymbolRef(sym string) (string, int64, bool) {
	return parseSBRef(strings.TrimPrefix(sym, "$"))
}

// Run before KeepFunc: an excluded sibling or DATA initializer must not hide
// address-taking or an alternate native use of a helper we fold into a root.
func coalesceARM64PrivateRegisterHelpers(file *File) (*File, error) {
	if file.Arch != ArchARM64 {
		return file, nil
	}
	// Do not replace existing explicit FP/tail-entry contracts. This bounded
	// continuation mechanism starts only at real source BL/CALL sites, then
	// audits every other use of each such helper across the complete file.
	called := map[string]bool{}
	for _, fn := range file.Funcs {
		for _, ins := range fn.Instrs {
			if (ins.Op == "CALL" || ins.Op == "BL") && len(ins.Args) == 1 && ins.Args[0].Kind == OpSym {
				base, _, ok := arm64PrivateSymbolRef(ins.Args[0].Sym)
				if ok {
					called[base] = true
				}
			}
		}
	}
	helpers := map[string]Func{}
	for _, fn := range file.Funcs {
		if !strings.HasSuffix(fn.Sym, "<>") || !called[fn.Sym] {
			continue
		}
		fp := false
		for _, ins := range fn.Instrs {
			for _, arg := range ins.Args {
				fp = fp || arg.Kind == OpFP || arg.Kind == OpFPAddr
			}
		}
		if fp || fn.ArgSize != 0 {
			continue // Ordinary FP helpers retain their existing typed ABI path.
		}
		if _, duplicate := helpers[fn.Sym]; duplicate {
			return nil, fmt.Errorf("duplicate ARM64 private TEXT %q", fn.Sym)
		}
		if err := arm64PrivateRegisterLeaf(fn); err != nil {
			return nil, err
		}
		helpers[fn.Sym] = fn
	}
	if len(helpers) == 0 {
		return file, nil
	}
	for _, data := range file.Data {
		for _, symbol := range []string{data.Sym, data.Addr} {
			base, _, _ := arm64PrivateSymbolRef(symbol)
			if _, escaped := helpers[base]; escaped {
				return nil, arm64GoABIContext("private helper %q escapes through DATA", base)
			}
		}
	}
	for _, global := range file.Globl {
		base, _, _ := arm64PrivateSymbolRef(global.Sym)
		if _, escaped := helpers[base]; escaped {
			return nil, arm64GoABIContext("private helper %q has a GLOBL/native address use", base)
		}
	}
	uses := map[string]int{}
	roots := map[string]map[string]bool{}
	for _, fn := range file.Funcs {
		for _, ins := range fn.Instrs {
			if ins.Op == OpTEXT {
				continue
			}
			// Folding removes the helper's standalone native TEXT identity.
			// A sibling excluded by KeepFunc must not hide an observation of
			// that original byte layout or a raw branch into the helper.
			if ins.Op == OpWORD || ins.Op == "DWORD" || ins.Op == "ADR" || ins.Op == "ADRP" || ins.Op == "PCALIGN" {
				return nil, arm64GoABIContext("private helper file requires native byte-layout effects at %q", ins.Raw)
			}
			for ai, arg := range ins.Args {
				memory, hasMemory := arg.Mem, arg.Kind == OpMem
				if arg.Kind == OpSym && strings.HasPrefix(arg.Sym, "$") {
					memory, hasMemory = parseMem(strings.TrimSpace(strings.TrimPrefix(arg.Sym, "$")))
				}
				if hasMemory && memory.Base == PC {
					return nil, arm64GoABIContext("private helper file has a native PC use at %q", ins.Raw)
				}
				if hasMemory && memory.Sym != "" {
					base, _, _ := arm64PrivateSymbolRef(memory.Sym)
					if _, escaped := helpers[base]; escaped {
						return nil, arm64GoABIContext("private helper %q has a native memory/address use at %q", base, ins.Raw)
					}
				}
				if arg.Kind != OpSym {
					continue
				}
				base, off, ok := arm64PrivateSymbolRef(arg.Sym)
				if _, helper := helpers[base]; !ok || !helper {
					continue
				}
				if (ins.Op != "CALL" && ins.Op != "BL") || ai != 0 || len(ins.Args) != 1 || off != 0 || strings.HasPrefix(arg.Sym, "$") {
					return nil, arm64GoABIContext("private helper %q has a non-call/address use at %q", base, ins.Raw)
				}
				if roots[fn.Sym] == nil {
					roots[fn.Sym] = map[string]bool{}
				}
				roots[fn.Sym][base] = true
				uses[base]++
			}
		}
	}
	for helper := range helpers {
		if uses[helper] == 0 {
			return nil, arm64GoABIContext("private helper %q has no closed source caller", helper)
		}
	}
	result := *file
	result.Funcs = make([]Func, 0, len(file.Funcs)-len(helpers))
	for _, fn := range file.Funcs {
		if _, helper := helpers[fn.Sym]; helper {
			continue
		}
		if len(roots[fn.Sym]) == 0 {
			result.Funcs = append(result.Funcs, fn)
			continue
		}
		copyFn := fn
		if len(fn.Instrs) == 0 {
			return nil, arm64GoABIContext("private helper root %q has no source terminator", fn.Sym)
		}
		last := fn.Instrs[len(fn.Instrs)-1].Op
		if last != OpRET && last != "B" && last != "JMP" {
			return nil, arm64GoABIContext("private helper root %q has no source terminator", fn.Sym)
		}
		copyFn.Instrs = append([]Instr(nil), fn.Instrs...)
		copyFn.arm64PrivateRegisterEntry = true
		// Retain the pre-coalescing source proof. Only audited register-only
		// helpers are appended below; their rewritten local calls must not
		// erase the original real-symbol call's private-frame provenance.
		copyFn.arm64PrivateUnexposedFrame = arm64CallFrameUnexposed(fn)
		names := make([]string, 0, len(roots[fn.Sym]))
		for name := range roots[fn.Sym] {
			names = append(names, name)
		}
		sort.Strings(names)
		labels := map[string]string{}
		occupied := map[string]bool{"entry": true}
		for _, ins := range fn.Instrs {
			if ins.Op == OpLABEL && len(ins.Args) == 1 {
				occupied[ins.Args[0].Sym] = true
			}
		}
		for i, name := range names {
			label := fmt.Sprintf("__arm64_private_%d", i)
			for occupied[label] {
				label += "_"
			}
			occupied[label], labels[name] = true, label
		}
		for i, ins := range copyFn.Instrs {
			if ins.Op == OpWORD || ins.Op == "ADR" || ins.Op == "ADRP" || ins.Op == "PCALIGN" {
				return nil, arm64GoABIContext("private helper root %q requires native byte-layout effects at %q", fn.Sym, ins.Raw)
			}
			for _, arg := range ins.Args {
				if arm64PrivateOperandMentionsReserved(arg) {
					return nil, arm64GoABIContext("private helper root %q observes unmodelled platform/scratch state at %q", fn.Sym, ins.Raw)
				}
			}
			if (ins.Op == "CALL" || ins.Op == "BL") && len(ins.Args) == 1 && ins.Args[0].Kind == OpSym {
				name, _, _ := arm64PrivateSymbolRef(ins.Args[0].Sym)
				if label := labels[name]; label != "" {
					copyFn.Instrs[i].Args = []Operand{{Kind: OpIdent, Ident: label}}
				}
			}
		}
		var err error
		copyFn, err = normalizeARM64NamedPCRelative(copyFn)
		if err != nil {
			return nil, err
		}
		for _, name := range names {
			copyFn.Instrs = append(copyFn.Instrs, Instr{Op: OpLABEL, Args: []Operand{{Kind: OpLabel, Sym: labels[name]}}})
			helper := helpers[name]
			copyFn.Instrs = append(copyFn.Instrs, helper.Instrs[1:len(helper.Instrs)-1]...)
			copyFn.Instrs = append(copyFn.Instrs, Instr{Op: "B", Args: []Operand{{Kind: OpMem, Mem: MemRef{Base: "R30"}}}, Raw: "B (R30)"})
		}
		result.Funcs = append(result.Funcs, copyFn)
	}
	return &result, nil
}
