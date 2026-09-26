package plan9asm

import (
	"strconv"
	"strings"

	"golang.org/x/arch/arm64/arm64asm"
)

func arm64RawPoolContains(offset, bytes, size int64) bool {
	return offset >= 0 && bytes > 0 && bytes <= size && offset <= size-bytes
}

func arm64RawPoolReplicateLoad(word uint32) (Instr, bool) {
	if ins, ok := decodeARM64RawSVEReplicateScalar(word); ok {
		return ins, true
	}
	return decodeARM64RawSVEReplicateBlock(word)
}

func arm64RawPoolLoadInBounds(ins arm64asm.Inst, word uint32, memory arm64asm.MemImmediate, offset, size int64) bool {
	// x/arch exposes the addressing mode but keeps the decoded immediate
	// private. Its AddrOffset printer is exactly [Base] or [Base,#decimal].
	// Use that checked representation rather than duplicating every encoder.
	text := memory.String()
	prefix := "[" + memory.Base.String()
	displacement := int64(0)
	if text != prefix+"]" {
		if !strings.HasPrefix(text, prefix+",#") || !strings.HasSuffix(text, "]") {
			return false
		}
		var err error
		displacement, err = strconv.ParseInt(strings.TrimSuffix(strings.TrimPrefix(text, prefix+",#"), "]"), 10, 32)
		if err != nil {
			return false
		}
	}
	return arm64RawPoolContains(offset+displacement, arm64RawPoolLoadBytes(ins, word), size)
}

func arm64RawPoolLoadBytes(ins arm64asm.Inst, word uint32) int64 {
	switch ins.Op {
	case arm64asm.LDRB, arm64asm.LDRSB, arm64asm.LDURB, arm64asm.LDURSB,
		arm64asm.LDTRB, arm64asm.LDTRSB, arm64asm.LDARB:
		return 1
	case arm64asm.LDRH, arm64asm.LDRSH, arm64asm.LDURH, arm64asm.LDURSH,
		arm64asm.LDTRH, arm64asm.LDTRSH, arm64asm.LDARH:
		return 2
	case arm64asm.LDRSW, arm64asm.LDURSW, arm64asm.LDTRSW:
		return 4
	case arm64asm.LDPSW:
		return 8
	case arm64asm.LDR, arm64asm.LDUR, arm64asm.LDTR, arm64asm.LDAR,
		arm64asm.LDP, arm64asm.LDNP:
		reg, ok := ins.Args[0].(arm64asm.Reg)
		if !ok {
			return 0
		}
		bytes := int64(0)
		switch {
		case reg >= arm64asm.B0 && reg <= arm64asm.B31:
			bytes = 1
		case reg >= arm64asm.H0 && reg <= arm64asm.H31:
			bytes = 2
		case reg >= arm64asm.W0 && reg <= arm64asm.WZR || reg >= arm64asm.S0 && reg <= arm64asm.S31:
			bytes = 4
		case reg >= arm64asm.X0 && reg <= arm64asm.XZR || reg >= arm64asm.D0 && reg <= arm64asm.D31:
			bytes = 8
		case reg >= arm64asm.Q0 && reg <= arm64asm.Q31:
			bytes = 16
		}
		if ins.Op == arm64asm.LDP || ins.Op == arm64asm.LDNP {
			bytes *= 2
		}
		return bytes
	case arm64asm.LD1R, arm64asm.LD2R, arm64asm.LD3R, arm64asm.LD4R:
		if form, ok := decodeARM64RawLDnR(word); ok && !form.post {
			return int64(form.count * form.arrangement.elementBits / 8)
		}
	case arm64asm.LD1, arm64asm.LD2, arm64asm.LD3, arm64asm.LD4:
		if form, ok := decodeARM64RawStructureLane(word); ok && form.load && !form.post {
			return int64(form.count * form.elementBits / 8)
		}
		// x/arch already validated this multiple-structure instruction.
		// Restrict to no writeback, then use opcode's register count and Q.
		if word&0xbfff0000 == 0x0c400000 {
			count := map[uint32]int64{0: 4, 2: 4, 4: 3, 6: 3, 7: 1, 8: 2, 10: 2}[word>>12&15]
			return count * (8 << (word >> 30 & 1))
		}
	}
	return 0
}
