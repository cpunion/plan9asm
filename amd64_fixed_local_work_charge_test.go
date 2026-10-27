package plan9asm

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestGoAMD64FixedLocalOrdinaryFullWorkBoundary(t *testing.T) {
	fn, sig := fixedLocalReplayFixture(t, "MOVQ $0,CX\n"+strings.Repeat("NOP\n", 55))
	fn.FrameSize = 64 << 10
	_, instructions, err := fixedLocalEdges(fn)
	if err != nil || len(instructions) != 64 {
		t.Fatal("independent 64-transfer ordinary fixture changed", len(instructions), err)
	}
	for _, deficit := range []int64{0, 1} {
		t.Run(fmt.Sprint(deficit), func(t *testing.T) {
			budget := x86FixedLocalBudget{steps: 1 << 18, bytes: (64 << 20) - deficit}
			proven, err := proveX86FixedLocalAttempt(fn, sig, &budget, false)
			if deficit == 0 {
				if err != nil || proven.x86FixedLocal == nil || budget.bytes != 0 || budget.steps != (1<<18)-64*16 {
					t.Fatalf("ordinary original full allowance: %+v, %v", budget, err)
				}
			} else if !errors.Is(err, ErrProbeNeedsContext) || !reflect.DeepEqual(proven, fn) {
				t.Fatal("one-byte-short ordinary replay must reject atomically", err)
			}
		})
	}
}

func TestGoAMD64FixedLocalPartitionVectorWork(t *testing.T) {
	body := strings.Replace(fixedLocalLayoutPingPong, "ret+16(FP)", "ret+8(FP)", 1)
	body = strings.Replace(body, "loop:\n", "loop:\n"+strings.Repeat("PADDQ X0,X0\n", 1024), 1)
	source := "TEXT ·Value(SB),4,$65536-16\nMOVQ a+0(FP),CX\nMOVQ $0,AX\n" +
		"MOVQ AX,0(SP)\nMOVQ AX,8(SP)\nMOVOU 0(SP),X0\n" + body
	goAssemblerOracleOutput(t, "amd64", source, "", true)
	file, err := Parse(ArchAMD64, source)
	if err != nil || len(file.Funcs) != 1 {
		t.Fatal(err)
	}
	fn := file.Funcs[0]
	_, sig := fixedLocalReplayFixture(t, "")
	budget := x86FixedLocalBudget{steps: 1 << 18, bytes: 64 << 20}
	proven, err := proveX86FixedLocalAttempt(fn, sig, &budget, true)
	if err != nil || proven.x86FixedLocal == nil {
		t.Fatal("initialized register work must not scan a whole frame per instruction", err)
	}
	used := (1 << 18) - budget.steps
	if used < 16*2*1024 || budget.bytes <= 0 {
		t.Fatal("both layouts and sixteen replays must retain instruction and byte accounting", used, budget)
	}
	budget = x86FixedLocalBudget{steps: used - 1, bytes: 64 << 20}
	proven, err = proveX86FixedLocalAttempt(fn, sig, &budget, true)
	if !errors.Is(err, ErrProbeNeedsContext) || budget.steps != 0 || !reflect.DeepEqual(proven, fn) {
		t.Fatal("register-only work still requires every step, with atomic failure", budget, err)
	}
}

func TestGoAMD64FixedLocalPartitionClonePrecharge(t *testing.T) {
	fn, sig := fixedLocalReplayFixture(t, "")
	proof := &x86FixedLocalProof{frame: 64}
	// Initial state: two one-word bitmaps (16 bytes), one SP map entry (64).
	// Reading and copying that payload costs 160 before the first instruction.
	budget := x86FixedLocalBudget{steps: 100, bytes: 159}
	var layouts x86FixedLocalLayoutGraph
	err := replayX86FixedLocal(&fn, sig, proof, 0, &budget, nil, &layouts)
	if !errors.Is(err, ErrProbeNeedsContext) || budget.steps != 100 || budget.bytes != 159 || len(layouts.nodes) != 0 {
		t.Fatal("state copy must be charged before executing or publishing", budget, err)
	}
	for _, ins := range fn.Instrs {
		if len(ins.x86FixedLocalMemory) != 0 {
			t.Fatal("failed clone debit published an access")
		}
	}
}

func TestGoAMD64FixedLocalPartitionCandidateWork(t *testing.T) {
	for _, count := range []int{3, 17} {
		for _, deficit := range []int64{0, 1} {
			t.Run(fmt.Sprintf("%d/%d", count, deficit), func(t *testing.T) {
				states := make(map[int]x86FixedLocalState)
				layouts := x86FixedLocalLayouts{at: map[int][]int{5: {}}, nodes: make([]x86FixedLocalLayoutNode, count)}
				for index := 0; index < count; index++ {
					states[index] = x86FixedLocalState{
						pointers: map[Reg]int64{SP: 0, "R14": int64(index * 8)}, known: map[Reg]uint8{AX: 255},
						bytes: x86FixedLocalByteSet{0}, booleanBytes: x86FixedLocalByteSet{0},
					}
					layouts.nodes[index].pc = 5
					layouts.at[5] = append(layouts.at[5], index)
				}
				// Each state has 16 bitmap bytes + three 64-byte map entries.
				// One comparison costs 64 lookup + 2*208 prior + 208 incoming.
				const perCandidate = 688
				budget := x86FixedLocalBudget{steps: 100, bytes: int64(count*count*perCandidate) - deficit}
				for query := 0; query < count; query++ {
					before := append([]x86FixedLocalLayoutNode(nil), layouts.nodes...)
					target, err := layouts.target(states, 0, 0, 5, states[count-1], &budget)
					if deficit != 0 && query == count-1 {
						if !errors.Is(err, ErrProbeNeedsContext) || !reflect.DeepEqual(before, layouts.nodes) || budget.bytes != perCandidate-1 {
							t.Fatal("last candidate must reject before changing an edge", budget, err)
						}
					} else if err != nil || target != count-1 {
						t.Fatal("complete candidate inventory must be inspected", target, err)
					}
				}
				if budget.steps != 100 || deficit == 0 && budget.bytes != 0 {
					t.Fatal("candidate comparisons require quadratic work, not instruction steps", budget)
				}
			})
		}
	}
}
