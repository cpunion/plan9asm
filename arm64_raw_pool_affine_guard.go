package plan9asm

import "math"

func (flow *arm64RawPoolValues) affineEdgeConstraint(edge arm64RawPoolEdge) (arm64PoolConstraint, bool) {
	var constraint arm64PoolConstraint
	word := flow.words[edge.from]
	if word&0xff000010 != 0x54000000 {
		return constraint, false
	}
	target := edge.from + int(int32(word<<8)>>13)
	if target == edge.from+1 || edge.to != target && edge.to != edge.from+1 {
		return constraint, false
	}
	predecessors := flow.before[edge.from]
	if edge.from == 0 || len(predecessors) != 1 || predecessors[0] != edge.from-1 {
		return constraint, false
	}
	compare := flow.words[edge.from-1]
	// CMP/CMN immediate, 64-bit only. SP and W comparisons do not establish
	// this expression domain's 64-bit GP constraints.
	if compare&0xbf80001f != 0xb100001f || compare>>5&31 == 31 {
		return constraint, false
	}
	constraint.expression = arm64PoolRegisterExpression(int(compare >> 5 & 31))
	constraint.interval = arm64PoolUnknownInterval
	immediate := uint64(compare >> 10 & 4095)
	if compare&(1<<22) != 0 {
		immediate <<= 12
	}
	condition := word & 15
	if edge.to != target {
		condition ^= 1
	}
	if compare&(1<<30) == 0 { // CMN: C is the carry out of unsigned addition.
		if immediate == 0 {
			return constraint, false
		}
		switch condition {
		case 2:
			constraint.interval.low = math.MaxUint64 - immediate + 1
		case 3:
			constraint.interval.high = math.MaxUint64 - immediate
		case 0:
			constraint.interval.low = -immediate
			constraint.interval.high = -immediate
		default:
			return constraint, false
		}
		return constraint, true
	}
	switch condition {
	case 0:
		constraint.interval = arm64PoolInterval{immediate, immediate}
	case 2:
		constraint.interval.low = immediate
	case 3:
		if immediate == 0 {
			return constraint, false
		}
		constraint.interval.high = immediate - 1
	case 8:
		constraint.interval.low = immediate + 1
	case 9:
		constraint.interval.high = immediate
	default:
		return constraint, false
	}
	return constraint, true
}
