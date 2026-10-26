package plan9asm

// A finite disjunction of opaque addresses. Empty is unknown, not parameter
// zero. Capacity overflow never invents an unrestricted address: a CFG meet
// loses that fact, while an instruction requiring it is rejected. Private
// relocated SP addresses and NUM remain in their independent domains.
type x86FixedLocalOrigins struct {
	first x86FixedLocalOrigin
	tail  *[3]x86FixedLocalOrigin
	count uint8
}

// The common singleton stays below Go's 128-byte inline map-value threshold.
// Multi-origin tails are immutable; all semantic comparisons use their content.
// Conservatively charge shared tails again at each retained/working map clone.
const fixedLocalOriginPayloadBytes int64 = 8 + 16 + 16 + 8 + 8 + 8
const fixedLocalOriginTailBytes int64 = 3 * (8 + 16 + 16 + 8)

func fixedLocalOriginStorage(state x86FixedLocalState) int64 {
	return fixedLocalOriginMapStorage(state.opaque) + fixedLocalOriginMapStorage(state.opaqueCells)
}

func fixedLocalOriginMapStorage[K comparable](values map[K]x86FixedLocalOrigins) int64 {
	bytes := int64(len(values)) * fixedLocalOriginPayloadBytes
	for _, value := range values {
		if value.tail != nil {
			bytes += fixedLocalOriginTailBytes
		}
	}
	return bytes
}

func fixedLocalOrigin(origin x86FixedLocalOrigin) x86FixedLocalOrigins {
	return x86FixedLocalOrigins{first: origin, count: 1}
}

func (origins x86FixedLocalOrigins) valid() bool {
	return origins.count == 1 && origins.tail == nil || origins.count > 1 && origins.count <= 4 && origins.tail != nil
}

func (origins x86FixedLocalOrigins) at(index uint8) x86FixedLocalOrigin {
	if index == 0 {
		return origins.first
	}
	return origins.tail[index-1]
}

func (origins x86FixedLocalOrigins) equal(other x86FixedLocalOrigins) bool {
	return origins.count == other.count && origins.first == other.first &&
		(origins.tail == other.tail || origins.tail != nil && other.tail != nil && *origins.tail == *other.tail)
}

func fixedLocalOriginSet(entries [4]x86FixedLocalOrigin, count uint8) x86FixedLocalOrigins {
	if count == 0 {
		return x86FixedLocalOrigins{}
	}
	result := fixedLocalOrigin(entries[0])
	if count > 1 {
		result.count = count
		result.tail = &[3]x86FixedLocalOrigin{entries[1], entries[2], entries[3]}
	}
	return result
}

func fixedLocalOriginLess(first, last x86FixedLocalOrigin) bool {
	if first.global != last.global {
		return first.global < last.global
	}
	if first.literal.symbol != last.literal.symbol {
		return first.literal.symbol < last.literal.symbol
	}
	if first.parameter != last.parameter {
		return first.parameter < last.parameter
	}
	return first.literal.offset < last.literal.offset
}

func (origins x86FixedLocalOrigins) union(other x86FixedLocalOrigins) (x86FixedLocalOrigins, bool) {
	if !origins.valid() || !other.valid() {
		return x86FixedLocalOrigins{}, false
	}
	if origins.equal(other) {
		return origins, true
	}
	var entries [4]x86FixedLocalOrigin
	for index := uint8(0); index < origins.count; index++ {
		entries[index] = origins.at(index)
	}
	count := origins.count
	for index := uint8(0); index < other.count; index++ {
		incoming := other.at(index)
		at := uint8(0)
		for at < count && fixedLocalOriginLess(entries[at], incoming) {
			at++
		}
		if at < count && entries[at] == incoming {
			continue
		}
		if count == 4 {
			return x86FixedLocalOrigins{}, false
		}
		copy(entries[at+1:], entries[at:count])
		entries[at] = incoming
		count++
	}
	if count == origins.count {
		return origins, true
	}
	return fixedLocalOriginSet(entries, count), true
}

func (origins x86FixedLocalOrigins) hasLiteral() bool {
	if !origins.valid() {
		return false
	}
	for index := uint8(0); index < origins.count; index++ {
		if origins.at(index).literal.symbol != "" {
			return true
		}
	}
	return false
}

func (origins x86FixedLocalOrigins) literal() (x86StaticSBAddress, bool) {
	return origins.first.literal, origins.count == 1 && origins.first.literal.symbol != ""
}

func fixedLocalOriginOwners(origins x86FixedLocalOrigins, proof *x86FixedLocalProof) bool {
	if !origins.valid() {
		return false
	}
	for index := uint8(0); index < origins.count; index++ {
		origin := origins.at(index)
		if origin.literal.symbol != "" {
			if _, valid := fixedLocalShiftLiteral(origin, 0, proof); !valid {
				return false
			}
		}
	}
	return true
}

func cloneFixedLocalOrigins[K comparable](input map[K]x86FixedLocalOrigins) map[K]x86FixedLocalOrigins {
	result := make(map[K]x86FixedLocalOrigins, len(input))
	for key, value := range input {
		result[key] = value
	}
	return result
}

func mergeFixedLocalOrigins[K comparable](previous, current map[K]x86FixedLocalOrigins) (bool, int64, error) {
	changed := false
	var work int64
	for key, old := range previous {
		incoming, found := current[key]
		if !found {
			delete(previous, key) // Unknown on either path grants no address role.
			changed = true
			continue
		}
		if old.valid() && old.equal(incoming) {
			continue
		}
		merged, valid := old.union(incoming)
		if !valid {
			if !old.valid() || !incoming.valid() {
				return false, work, goAMD64FrameContext("fixed local CFG has an invalid opaque origin")
			}
			// Preserve the old meet's unknown state for dead/overwritten
			// values. The separate numerical bytes stay undefined, so no
			// later scalar read or dereference acquires permission.
			delete(previous, key)
			changed = true
			continue
		}
		if !merged.equal(old) {
			previous[key] = merged
			changed = true
			work += fixedLocalOriginTailBytes
		}
	}
	return changed, work, nil
}

// The predicate refers to the CMP's captured literal value, not the current
// contents of its former register. Only the compared register is refined.
// Native addresses may equal that literal; no disjointness is assumed.
func fixedLocalLiteralEdge(state x86FixedLocalState, op Op, taken bool) (x86FixedLocalState, bool) {
	comparison := state.comparison
	origins, found := state.opaque[comparison.register]
	condition := x86BranchCondition(op)
	if !found || !origins.valid() || condition != "EQ" && condition != "NE" {
		return state, true
	}
	equal := taken == (condition == "EQ")
	next := fixedLocalOrigin(x86FixedLocalOrigin{literal: comparison.literal})
	if !equal {
		var entries [4]x86FixedLocalOrigin
		count := uint8(0)
		for index := uint8(0); index < origins.count; index++ {
			origin := origins.at(index)
			if origin.literal != comparison.literal {
				entries[count] = origin
				count++
			}
		}
		if count == origins.count {
			return state, true
		}
		next = fixedLocalOriginSet(entries, count)
	}
	if next.count == 0 {
		return state, false
	}
	if !next.equal(origins) {
		state.opaque = cloneFixedLocalOrigins(state.opaque)
		state.opaque[comparison.register] = next
	}
	return state, true
}
