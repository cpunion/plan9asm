package plan9asm

import (
	"encoding/binary"
	"math"

	"golang.org/x/arch/arm64/arm64asm"
)

func (flow *arm64RawPoolValues) counterStrideBound(at int, expression arm64PoolAffine, initial arm64PoolInterval, delta uint64) (arm64PoolInterval, bool) {
	signed := int64(delta)
	if signed == 0 || signed < -(1<<30) || signed > 1<<30 {
		return arm64PoolInterval{}, false
	}
	step := uint64(signed)
	if signed < 0 {
		step = -step
	}
	if step&(step-1) != 0 {
		return arm64PoolInterval{}, false
	}
	bound := arm64PoolInterval{step, initial.high}
	if signed < 0 {
		if initial.low < step || initial.high > math.MaxInt64 {
			return arm64PoolInterval{}, false
		}
	} else {
		if initial.low <= math.MaxInt64 || initial.high > -step {
			return arm64PoolInterval{}, false
		}
		bound = arm64PoolInterval{initial.low, -step}
	}
	return bound, flow.multipleOfPowerOfTwo(at, expression, step)
}

// Retain only the residue through arithmetic definitions. In particular,
// -(n & 24) is divisible by eight even though its interval includes values
// between the two actual possibilities -16 and -8. No branch predicate is
// needed; every predecessor must establish residue zero independently.
func (flow *arm64RawPoolValues) multipleOfPowerOfTwo(at int, expression arm64PoolAffine, step uint64) bool {
	mask := step - 1
	normalize := func(value arm64PoolAffine) arm64PoolAffine {
		value.constant &= mask
		for register, coefficient := range value.coefficient {
			value.coefficient[register] = int64(uint64(coefficient) & mask)
		}
		return value
	}
	queue := []arm64PoolAffineQuery{{at, normalize(expression)}}
	visited := make(map[arm64PoolAffineQuery]bool)
	found := false
	for work := 0; len(queue) > 0; work++ {
		if work >= 4096 || flow.affineWork >= 16384 {
			return false
		}
		flow.affineWork++
		state := queue[len(queue)-1]
		queue = queue[:len(queue)-1]
		if visited[state] {
			continue
		}
		visited[state] = true
		if state.expression.relocations != 0 {
			return false
		}
		if state.expression.isConstant() {
			if state.expression.constant != 0 {
				return false
			}
			found = true
			continue
		}
		if state.at < 0 || state.at >= len(flow.before) || len(flow.before[state.at]) == 0 {
			return false
		}
		for _, previous := range flow.before[state.at] {
			if previous < 0 || flow.opaque[previous] && !flow.opaquePreserves(previous, state.expression) {
				return false
			}
			word := flow.words[previous]
			writes, known := arm64RawPoolGPWrites(word)
			if !known {
				return false
			}
			next := arm64PoolAffineQuery{previous, state.expression}
			if affected := writes & next.expression.registerMask(); affected != 0 {
				if savedAt, saved, ok := flow.rewindStackLoad(previous, next.expression); ok {
					queue = append(queue, arm64PoolAffineQuery{savedAt, normalize(saved)})
					continue
				}
				destination, value, valid := flow.affineDefinition(word)
				if !valid {
					destination, value, valid = flow.maskedResidueDefinition(previous, word, mask)
				}
				if !valid || affected != 1<<uint(destination) || !next.expression.substitute(destination, value) {
					return false
				}
				next.expression = normalize(next.expression)
			}
			queue = append(queue, next)
		}
	}
	return found
}

func (flow *arm64RawPoolValues) maskedResidueDefinition(at int, word uint32, residueMask uint64) (int, arm64PoolAffine, bool) {
	var code [4]byte
	binary.LittleEndian.PutUint32(code[:], word)
	ins, err := arm64asm.Decode(code[:])
	if err != nil || word>>31 == 0 || word&31 == 31 ||
		(ins.Op != arm64asm.AND && ins.Op != arm64asm.ANDS && ins.Op != arm64asm.BIC && ins.Op != arm64asm.BICS) {
		return 0, arm64PoolAffine{}, false
	}
	var mask uint64
	switch {
	case word&0x1f800000 == 0x12000000:
		switch immediate := ins.Args[2].(type) {
		case arm64asm.Imm:
			mask = uint64(immediate.Imm)
		case arm64asm.Imm64:
			mask = immediate.Imm
		default:
			return 0, arm64PoolAffine{}, false
		}
	case word&0x1f000000 == 0x0a000000:
		// An independent mask query must not reset the residue walk's budget
		// or reuse a pool-relative offset as the mask's numeric bit pattern.
		numeric := *flow.numericValues()
		numeric.clearValueCaches()
		value := numeric.invariantInterval(at, arm64PoolRegisterExpression(int(word>>16&31)))
		flow.affineWork += numeric.affineWork
		if value.low != value.high {
			return 0, arm64PoolAffine{}, false
		}
		operand := arm64PoolLogicalOperand{must: value.low, may: value.low}.shifted(word>>22&3, word>>10&63)
		mask = operand.must
	default:
		return 0, arm64PoolAffine{}, false
	}
	if ins.Op == arm64asm.BIC || ins.Op == arm64asm.BICS {
		mask = ^mask
	}
	switch mask & residueMask {
	case 0:
		return int(word & 31), arm64PoolAffine{}, true
	case residueMask:
		return int(word & 31), arm64PoolRegisterExpression(int(word >> 5 & 31)), true
	default:
		return 0, arm64PoolAffine{}, false
	}
}
