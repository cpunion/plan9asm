package plan9asm

import (
	"encoding/binary"
	"math/bits"

	"golang.org/x/arch/arm64/arm64asm"
)

// AND is not affine. A guarded interval can nevertheless prove its exact
// result, for example length in [8,15] implies (length & 24) == 8. Only then
// substitute a constant into a larger affine expression.
func (flow *arm64RawPoolValues) affineMaskInterval(at int, word uint32) (arm64PoolInterval, bool) {
	if word&0xff800000 != 0x92000000 { // AND Xd, Xn, logical immediate.
		return arm64PoolInterval{}, false
	}
	var code [4]byte
	binary.LittleEndian.PutUint32(code[:], word)
	ins, err := arm64asm.Decode(code[:])
	if err != nil || ins.Op != arm64asm.AND {
		return arm64PoolInterval{}, false
	}
	source, ok := ins.Args[1].(arm64asm.Reg)
	if !ok || source < arm64asm.X0 || source > arm64asm.XZR {
		return arm64PoolInterval{}, false
	}
	var mask uint64
	switch operand := ins.Args[2].(type) {
	case arm64asm.Imm:
		mask = uint64(operand.Imm)
	case arm64asm.Imm64:
		mask = operand.Imm
	default:
		return arm64PoolInterval{}, false
	}
	input := flow.affineInterval(at, arm64PoolRegisterExpression(int(source-arm64asm.X0)))
	if upper := flow.upper(at, source); upper < input.high {
		input.high = upper
	}
	if input.low > input.high {
		return arm64PoolUnknownInterval, false
	}
	return arm64PoolMaskInterval(input, mask), true
}

func arm64PoolMaskInterval(input arm64PoolInterval, mask uint64) arm64PoolInterval {
	// Bits above the highest differing endpoint bit are constant throughout
	// an unsigned interval. Everything below is conservatively free to vary.
	varying := uint64(1)<<uint(bits.Len64(input.low^input.high)) - 1
	low := input.low &^ varying & mask
	high := low | varying&mask
	if high > input.high {
		high = input.high // AND never increases an unsigned operand.
	}
	return arm64PoolInterval{low, high}
}
