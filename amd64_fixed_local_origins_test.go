package plan9asm

import (
	"errors"
	"reflect"
	"testing"
	"unsafe"
)

func TestGoAMD64FixedLocalOriginJoinStorageBudget(t *testing.T) {
	t.Run("invalid second map retains prior allocation cost", func(t *testing.T) {
		full := fixedLocalOriginSet([4]x86FixedLocalOrigin{
			{parameter: 0}, {parameter: 8}, {parameter: 16}, {parameter: 24},
		}, 4)
		previous := x86FixedLocalState{
			pointers: make(map[Reg]int64), addresses: make(map[int64]int64),
			opaque:      map[Reg]x86FixedLocalOrigins{AX: fixedLocalOrigin(x86FixedLocalOrigin{parameter: 0})},
			opaqueCells: map[int64]x86FixedLocalOrigins{0: full},
		}
		current := cloneFixedLocalState(previous)
		current.opaque[AX] = fixedLocalOrigin(x86FixedLocalOrigin{parameter: 8})
		current.opaqueCells[0] = x86FixedLocalOrigins{}
		oldBytes := fixedLocalOriginStorage(previous)
		_, work, err := mergeFixedLocalState(&previous, current)
		if !errors.Is(err, ErrProbeNeedsContext) || work != 144 {
			t.Fatal("failed later union lost the completed allocation charge", work, err)
		}
		retained, budget := int64(0), x86FixedLocalBudget{bytes: 144}
		if err := budget.retainOriginGrowth(fixedLocalOriginStorage(previous)-oldBytes, work, &retained); err != nil || budget.bytes != 0 {
			t.Fatal("failed attempt became free for later replay", err)
		}
	})
	for _, row := range []struct {
		name           string
		retained, work int64
		fails          bool
	}{
		{"exact retained and work capacity", (1 << 20) - 144, 144, false},
		{"retained tail exceeds capacity", (1 << 20) - 143, 144, true},
		{"working tail exceeds capacity", 0, 143, true},
	} {
		t.Run(row.name, func(t *testing.T) {
			previous := x86FixedLocalState{
				pointers: make(map[Reg]int64), addresses: make(map[int64]int64),
				opaque: map[Reg]x86FixedLocalOrigins{AX: fixedLocalOrigin(x86FixedLocalOrigin{parameter: 0})},
			}
			current := cloneFixedLocalState(previous)
			current.opaque[AX] = fixedLocalOrigin(x86FixedLocalOrigin{parameter: 8})
			oldBytes := fixedLocalOriginStorage(previous)
			changed, work, err := mergeFixedLocalState(&previous, current)
			if err != nil || !changed || work != 144 {
				t.Fatal("two singleton predecessors must create one bounded union", changed, err)
			}
			growth := fixedLocalOriginStorage(previous) - oldBytes
			if growth != 144 {
				t.Fatal("join did not account for its complete immutable tail", growth)
			}
			retained, budget := row.retained, x86FixedLocalBudget{bytes: row.work}
			err = budget.retainOriginGrowth(growth, work, &retained)
			if errors.Is(err, ErrProbeNeedsContext) != row.fails || !row.fails && err != nil {
				t.Fatal("join exceeded its retained/work bound", err)
			}
			if row.fails && (retained != row.retained || budget.bytes != row.work) {
				t.Fatal("failed accounting partially consumed its budget")
			}
		})
	}
	for _, replaceTail := range []bool{false, true} {
		name, wantGrowth := "new tail plus removed singleton", int64(80)
		if replaceTail {
			name, wantGrowth = "two to three candidates", 0
		}
		t.Run(name, func(t *testing.T) {
			previous := x86FixedLocalState{
				pointers: make(map[Reg]int64), addresses: make(map[int64]int64),
				opaque: map[Reg]x86FixedLocalOrigins{AX: fixedLocalOrigin(x86FixedLocalOrigin{parameter: 0})},
			}
			if replaceTail {
				previous.opaque[AX], _ = previous.opaque[AX].union(fixedLocalOrigin(x86FixedLocalOrigin{parameter: 8}))
			} else {
				previous.opaque[BX] = fixedLocalOrigin(x86FixedLocalOrigin{parameter: 8})
			}
			current := cloneFixedLocalState(previous)
			current.opaque[AX] = fixedLocalOrigin(x86FixedLocalOrigin{parameter: 16})
			delete(current.opaque, BX)
			oldBytes := fixedLocalOriginStorage(previous)
			changed, work, err := mergeFixedLocalState(&previous, current)
			growth := fixedLocalOriginStorage(previous) - oldBytes
			if err != nil || !changed || work != 144 || growth != wantGrowth {
				t.Fatal("live storage and newly allocated tail work were conflated", changed, work, growth, err)
			}
			retained, budget := int64(0), x86FixedLocalBudget{bytes: 143}
			if err := budget.retainOriginGrowth(growth, work, &retained); !errors.Is(err, ErrProbeNeedsContext) || retained != 0 || budget.bytes != 143 {
				t.Fatal("tail replacement bypassed its independent work budget", err)
			}
		})
	}
}

func TestGoAMD64FixedLocalOriginValueStorage(t *testing.T) {
	// Go stores map values above 128 bytes indirectly. The common singleton
	// must not require an additional allocation for every state-map entry.
	if size := unsafe.Sizeof(x86FixedLocalOrigins{}); size > 128 {
		t.Fatalf("opaque map value is %d bytes, exceeding the inline threshold", size)
	}
}

func TestGoAMD64FixedLocalOriginSetValueSemantics(t *testing.T) {
	set := func(offsets ...int64) x86FixedLocalOrigins {
		value := fixedLocalOrigin(x86FixedLocalOrigin{literal: x86StaticSBAddress{symbol: "choice<>", offset: offsets[0]}})
		for _, offset := range offsets[1:] {
			var valid bool
			value, valid = value.union(fixedLocalOrigin(x86FixedLocalOrigin{literal: x86StaticSBAddress{symbol: "choice<>", offset: offset}}))
			if !valid {
				t.Fatal("bounded set rejected")
			}
		}
		return value
	}
	t.Run("invalid equal values remain unknown", func(t *testing.T) {
		for _, value := range []x86FixedLocalOrigins{{}, {count: 2}, {count: 5, tail: &[3]x86FixedLocalOrigin{}}} {
			previous := map[Reg]x86FixedLocalOrigins{AX: value}
			if _, _, err := mergeFixedLocalOrigins(previous, map[Reg]x86FixedLocalOrigins{AX: value}); !errors.Is(err, ErrProbeNeedsContext) {
				t.Fatal("equal fast path granted an invalid origin", err)
			}
		}
	})
	t.Run("independent order and duplicate joins", func(t *testing.T) {
		for _, offsets := range [][]int64{{0, 8, 16}, {16, 0, 8}, {8, 16, 8, 0, 16}} {
			first, second := set(0, 8, 16), set(offsets...)
			previous := map[Reg]x86FixedLocalOrigins{AX: first}
			if changed, work, err := mergeFixedLocalOrigins(previous, map[Reg]x86FixedLocalOrigins{AX: second}); err != nil || changed || work != 0 {
				t.Fatal("equal address sets must converge independently of allocation/order", changed, err)
			}
			if !reflect.DeepEqual(first, second) || !reflect.DeepEqual(previous[AX], first) {
				t.Fatal("canonical value content changed")
			}
		}
	})
	t.Run("union does not change either input", func(t *testing.T) {
		first, second := set(0, 8), set(16, 24)
		merged, valid := first.union(second)
		if !valid || !reflect.DeepEqual(merged, set(0, 8, 16, 24)) ||
			!reflect.DeepEqual(first, set(0, 8)) || !reflect.DeepEqual(second, set(16, 24)) {
			t.Fatal("union mutated a shared immutable input")
		}
	})
	t.Run("affine write does not change prior state", func(t *testing.T) {
		fn, sig, _ := fixedLocalPointerChoiceFixture(t, "MOVQ (SI),AX")
		proof := &x86FixedLocalProof{frame: fn.FrameSize, binding: sig.x86FixedLocal}
		original := set(0, 8)
		state := x86FixedLocalState{
			opaque: map[Reg]x86FixedLocalOrigins{AX: original},
			known:  make(map[Reg]uint8), booleans: make(map[Reg]uint8),
		}
		ins := Instr{Op: "ADDQ", Args: []Operand{{Kind: OpImm, Imm: 8}, {Kind: OpReg, Reg: AX}}}
		if handled, err := fixedLocalInputAddressStep(&ins, &state, proof, 0); err != nil || !handled {
			t.Fatal("bounded affine write rejected", handled, err)
		}
		if !reflect.DeepEqual(original, set(0, 8)) || !reflect.DeepEqual(state.opaque[AX], set(8, 16)) {
			t.Fatal("affine write mutated the prior immutable set")
		}
	})
}
