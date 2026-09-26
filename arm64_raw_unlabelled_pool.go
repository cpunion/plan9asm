package plan9asm

import (
	"encoding/binary"
	"strings"

	"golang.org/x/arch/arm64/arm64asm"
)

// identifyARM64UnlabelledPool handles a trailing, contiguous raw routine and
// its appended constants. Walk actual control-flow edges before examining
// address references: data bytes can themselves look like branches or ADRs.
// Unknown source widths, indirect control flow, escaping addresses and mixed
// code/data regions remain on the ordinary fail-closed path.
func identifyARM64UnlabelledPool(fn Func, points []arm64RawLayoutPoint, known map[string]bool, returnClobbers uint32) (map[int]bool, []arm64RawDataBlob, map[int]string) {
	end := len(fn.Instrs)
	if end > 0 && fn.Instrs[end-1].Op == OpRET {
		end--
	}
	start := end
	for start > 0 && arm64RawLiteralWord(fn.Instrs[start-1]) {
		start--
	}
	if start == end {
		return nil, nil, nil
	}
	for _, ins := range fn.Instrs[:start] {
		if ins.Op == OpWORD {
			return nil, nil, nil
		}
		for _, arg := range ins.Args {
			if arg.Kind == OpMem && arg.Mem.Base == PC || arg.Kind == OpSym && strings.Contains(arg.Sym, fn.Sym+"+") {
				return nil, nil, nil
			}
		}
	}

	visited := make(map[int]bool)
	queue := []int{start}
	last := start
	addresses := make(map[int]int)
	for len(queue) > 0 {
		at := queue[len(queue)-1]
		queue = queue[:len(queue)-1]
		if visited[at] {
			continue
		}
		if at < start || at >= end {
			return nil, nil, nil
		}
		visited[at] = true
		if at > last {
			last = at
		}
		word := uint32(fn.Instrs[at].Args[0].Imm)
		// BR and BLR cannot prove their successor without value-flow analysis.
		if word&0xfffffc1f == 0xd61f0000 || word&0xfffffc1f == 0xd63f0000 {
			return nil, nil, nil
		}
		if word&0xfffffc1f == 0xd65f0000 { // RET Xn.
			continue
		}
		displacement, op, relative, err := arm64RawPCRelativeDisplacement(fn.Instrs[at])
		if err != nil {
			return nil, nil, nil
		}
		if relative {
			if displacement%4 != 0 {
				return nil, nil, nil
			}
			target := at + int(displacement/4)
			switch op {
			case "ADR":
				addresses[at] = target
			case "B", "BL", "CBZ", "CBNZ", "TBZ", "TBNZ":
				queue = append(queue, target)
			default:
				return nil, nil, nil
			}
		}
		if word&0xfc000000 != 0x14000000 { // Only unconditional B ends fallthrough.
			queue = append(queue, at+1)
		}
	}
	pool := last + 1
	if pool >= end {
		return nil, nil, nil
	}

	// An ADR can also construct a function pointer. Only classify its target
	// as data when the produced address is used exclusively by loads and killed
	// on every outgoing path (possibly by an explicit return contract).
	for at, target := range addresses {
		if target >= pool && target < end && !arm64RawAddressOnlyLoadedWithExit(fn.Instrs, at, pool, returnClobbers) {
			return nil, nil, nil
		}
	}
	insertions := make(map[int]string)
	label := func(at int) string {
		if name, ok := insertions[at]; ok {
			return name
		}
		name := arm64UniqueRawTargetLabel(known, points[at])
		known[name] = true
		insertions[at] = name
		return name
	}
	blob := arm64RawDataBlob{label: label(pool), aliases: make(map[string]int64)}
	for _, target := range addresses {
		if target >= pool && target < end {
			blob.aliases[label(target)] = int64(target-pool) * 4
		}
	}
	words := make(map[int]bool)
	for at := pool; at < end; at++ {
		words[at] = true
		var bytes [4]byte
		binary.LittleEndian.PutUint32(bytes[:], uint32(fn.Instrs[at].Args[0].Imm))
		blob.bytes = append(blob.bytes, bytes[:]...)
	}
	return words, []arm64RawDataBlob{blob}, insertions
}

func arm64RawAddressOnlyLoaded(instructions []Instr, at, end int) bool {
	return arm64RawAddressOnlyLoadedWithExit(instructions, at, end, 0)
}

// Only the translating caller has a FuncSig. Standalone layout probes remain
// conservative. A parameter-frame function with an explicit void result has
// no register results; restrict terminal kills to R0..R17, avoiding platform,
// callee-preserved, frame, Go-context and link registers. Custom ArgRegs and
// missing frame contracts deliberately do not enable this rule.
func arm64RawPoolReturnClobbers(fn Func, sig FuncSig) uint32 {
	if sig.Ret != Void || len(sig.Frame.Results) != 0 || len(sig.ArgRegs) != 0 || len(sig.Frame.Params) == 0 || fn.ArgSize <= 0 {
		return 0
	}
	return (1 << 18) - 1
}

func arm64RawAddressOnlyLoadedWithExit(instructions []Instr, at, end int, returnClobbers uint32) bool {
	register := arm64asm.X0 + arm64asm.Reg(uint32(instructions[at].Args[0].Imm)&31)
	wordRegister := register - arm64asm.X0 + arm64asm.W0
	isAddress := func(r arm64asm.Reg) bool { return r == register || r == wordRegister }
	loaded := false
	visited := make(map[int]bool)
	queue := []int{at + 1}
	for len(queue) > 0 {
		i := queue[len(queue)-1]
		queue = queue[:len(queue)-1]
		if i < 0 || i >= end {
			return false
		}
		if visited[i] {
			continue
		}
		visited[i] = true
		if !arm64RawLiteralWord(instructions[i]) {
			return false
		}
		word := uint32(instructions[i].Args[0].Imm)
		var code [4]byte
		binary.LittleEndian.PutUint32(code[:], word)
		decoded, err := arm64asm.Decode(code[:])
		if err != nil {
			return false
		}
		// A canonical return can kill a scratch address only when the caller
		// supplied an explicit return contract. Calls and other indirect exits
		// still expose live registers; RET through the address is never data.
		if word == 0xd65f03c0 && returnClobbers&(1<<uint(register-arm64asm.X0)) != 0 {
			continue
		}
		switch decoded.Op {
		case arm64asm.BL, arm64asm.BLR, arm64asm.BR, arm64asm.RET, arm64asm.ERET, arm64asm.DRPS:
			return false
		}
		// Restrict kill recognition to these unambiguous destination-first
		// operations. Other encodings can read/write registers implicitly.
		writesFirst := false
		switch decoded.Op.String() {
		case "ADR", "MOV", "MOVZ", "MOVN", "ADD", "SUB", "AND", "ORR", "EOR":
			writesFirst = true
		case "LDR", "LDRB", "LDRH", "LDRSB", "LDRSH", "LDRSW",
			"LDUR", "LDURB", "LDURH", "LDURSB", "LDURSH", "LDURSW",
			"CSEL", "CSINC", "CSINV", "CSNEG":
			writesFirst = true
		}
		kills := false
		for n, arg := range decoded.Args {
			if arg == nil {
				break
			}
			switch arg := arg.(type) {
			case arm64asm.PCRel:
				switch decoded.Op {
				case arm64asm.ADR:
				case arm64asm.B, arm64asm.CBZ, arm64asm.CBNZ, arm64asm.TBZ, arm64asm.TBNZ:
					if int64(arg)%4 != 0 {
						return false
					}
					queue = append(queue, i+int(arg)/4)
				default:
					return false
				}
			case arm64asm.Reg:
				if isAddress(arg) {
					if n == 0 && writesFirst {
						kills = true
					} else {
						return false
					}
				}
			case arm64asm.RegSP:
				if isAddress(arm64asm.Reg(arg)) {
					if n == 0 && writesFirst {
						kills = true
					} else {
						return false
					}
				}
			case arm64asm.MemImmediate:
				if arg.Mode == arm64asm.AddrPostReg {
					// x/arch's arg_Xns_mem_post_Xm keeps Rm private, but its
					// decoder defines it as bits 20:16. An unrelated post-index
					// operand is harmless; using the address as Rm is not.
					index := arm64asm.X0 + arm64asm.Reg(word>>16&31)
					if isAddress(index) {
						return false
					}
				}
				if arm64asm.Reg(arg.Base) == register {
					if !strings.HasPrefix(decoded.Op.String(), "LD") || arg.Mode != arm64asm.AddrOffset {
						return false
					}
					loaded = true
				}
			case arm64asm.MemExtend:
				if isAddress(arm64asm.Reg(arg.Base)) || isAddress(arg.Index) {
					return false
				}
			case arm64asm.RegExtshiftAmount:
				// x/arch keeps this typed operand's register field private, but
				// prints the register before the comma separating its modifier.
				name := strings.SplitN(arg.String(), ",", 2)[0]
				if name == register.String() || name == wordRegister.String() {
					return false
				}
			}
		}
		if kills {
			continue
		}
		if word&0xfc000000 != 0x14000000 {
			queue = append(queue, i+1)
		}
	}
	return loaded
}
