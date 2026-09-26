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
func identifyARM64UnlabelledPool(fn Func, points []arm64RawLayoutPoint, known map[string]bool) (map[int]bool, []arm64RawDataBlob, map[int]string) {
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
	// before the next control-flow boundary; do not guess from comments.
	for at, target := range addresses {
		if target >= pool && target < end && !arm64RawAddressOnlyLoaded(fn.Instrs, at, pool) {
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
	register := arm64asm.X0 + arm64asm.Reg(uint32(instructions[at].Args[0].Imm)&31)
	loaded := false
	for i := at + 1; i < end; i++ {
		if !arm64RawLiteralWord(instructions[i]) {
			return false
		}
		var code [4]byte
		binary.LittleEndian.PutUint32(code[:], uint32(instructions[i].Args[0].Imm))
		decoded, err := arm64asm.Decode(code[:])
		if err != nil {
			return false
		}
		// Restrict kill recognition to these unambiguous destination-first
		// operations. Other encodings can read/write registers implicitly.
		writesFirst := false
		switch decoded.Op.String() {
		case "ADR", "MOV", "MOVZ", "MOVN", "ADD", "SUB", "AND", "ORR", "EOR":
			writesFirst = true
		}
		kills := false
		for n, arg := range decoded.Args {
			if arg == nil {
				break
			}
			switch arg := arg.(type) {
			case arm64asm.PCRel:
				if decoded.Op != arm64asm.ADR {
					return false
				}
			case arm64asm.Reg:
				if arg == register || arg == register-arm64asm.X0+arm64asm.W0 {
					if n == 0 && writesFirst {
						kills = true
					} else {
						return false
					}
				}
			case arm64asm.RegSP:
				if arm64asm.Reg(arg) == register {
					if n == 0 && writesFirst {
						kills = true
					} else {
						return false
					}
				}
			case arm64asm.MemImmediate:
				if arm64asm.Reg(arg.Base) == register {
					if !strings.HasPrefix(decoded.Op.String(), "LD") || arg.Mode != arm64asm.AddrOffset {
						return false
					}
					loaded = true
				}
			case arm64asm.MemExtend, arm64asm.RegExtshiftAmount:
				// Register-indexed addressing and shifted uses need a broader
				// value-flow proof; keep them on the strict path for now.
				return false
			}
		}
		if kills {
			return loaded
		}
		if arm64RawIsNonFallthroughTerminator(instructions[i]) || decoded.Op == arm64asm.BLR {
			return false
		}
	}
	return false
}
