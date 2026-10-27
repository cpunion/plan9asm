package plan9asm

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

// These analysis-only fixtures deliberately exclude source capture, subprocesses
// and LLVM construction. Their complete proof must succeed before timing starts.
func fixedLocalReplayFixture(t testing.TB, body string) (Func, FuncSig) {
	t.Helper()
	source := "TEXT ·Value(SB),4,$64-16\nMOVQ a+0(FP),AX\n" +
		"MOVQ AX,0(SP)\nMOVQ AX,8(SP)\nMOVOU 0(SP),X0\nMOVOU X0,X1\n" +
		body + "\nMOVQ X0,AX\nMOVQ AX,ret+8(FP)\nRET\n"
	file, err := Parse(ArchAMD64, source)
	if err != nil {
		t.Fatal(err)
	}
	sig := FuncSig{Name: "Value", Args: []LLVMType{I64}, Ret: I64,
		Frame: FrameLayout{
			Params:  []FrameSlot{{Offset: 0, Type: I64, Index: 0, Field: -1}},
			Results: []FrameSlot{{Offset: 8, Type: I64, Index: 0, Field: -1}},
		}}
	return file.Funcs[0], sig
}

func fixedLocalReplayBody(shape string, length int) string {
	const instruction = "PADDQ X1,X0\n"
	switch shape {
	case "linear":
		return strings.Repeat(instruction, length)
	case "diamond":
		return "CMPQ AX,AX\nJEQ left\n" + strings.Repeat(instruction, length/2) +
			"JMP joined\nleft:\n" + strings.Repeat(instruction, length-length/2) + "joined:\n"
	case "loop":
		return "MOVQ $3,CX\nagain:\n" + strings.Repeat(instruction, length) +
			"SUBQ $1,CX\nJNE again\n"
	default:
		panic("unknown fixed-local replay fixture shape")
	}
}

func TestGoAMD64FixedLocalReplayInstructionBudget(t *testing.T) {
	for _, length := range []int{8, 128, 1024, 5000} {
		t.Run(fmt.Sprint(length), func(t *testing.T) {
			fn, sig := fixedLocalReplayFixture(t, fixedLocalReplayBody("linear", length))
			effects, err := fixedLocalEffectPlan(fn)
			if err != nil {
				t.Fatal(err)
			}
			_, instructions, err := fixedLocalEdges(fn)
			if err != nil {
				t.Fatal(err)
			}
			count := len(instructions)
			for _, deficit := range []string{"none", "step", "byte"} {
				t.Run(deficit, func(t *testing.T) {
					current := fn
					current.Instrs = append([]Instr(nil), fn.Instrs...)
					proof := &x86FixedLocalProof{frame: fn.FrameSize}
					budget := x86FixedLocalBudget{steps: count, bytes: int64(count) * fn.FrameSize}
					if deficit == "step" {
						budget.steps--
					}
					if deficit == "byte" {
						budget.bytes--
					}
					var layouts x86FixedLocalLayoutGraph
					err := replayX86FixedLocal(&current, sig, proof, 0, &budget, effects, &layouts)
					if deficit == "none" {
						if err != nil || budget.steps != 0 || budget.bytes != 0 {
							t.Fatalf("complete transfer: budget=%+v, error=%v", budget, err)
						}
						accesses := 0
						for _, ins := range current.Instrs {
							accesses += len(ins.x86FixedLocalMemory)
						}
						if accesses != 3 {
							t.Fatalf("got %d private accesses, want both stores and complete vector load", accesses)
						}
						return
					}
					if !errors.Is(err, ErrProbeNeedsContext) {
						t.Fatalf("exhausted %s budget returned %v", deficit, err)
					}
					for _, ins := range current.Instrs {
						if len(ins.x86FixedLocalMemory) != 0 {
							t.Fatal("incomplete replay published a provisional access")
						}
					}
				})
			}
		})
	}
}

func fixedLocalReplayControlCases() []struct {
	name  string
	body  string
	valid bool
} {
	return []struct {
		name  string
		body  string
		valid bool
	}{
		{"linear", fixedLocalReplayBody("linear", 8), true},
		{"diamond", fixedLocalReplayBody("diamond", 8), true},
		{"loop", fixedLocalReplayBody("loop", 8), true},
		{"duplicate successor", "CMPQ AX,AX\nJEQ next\nnext:\nPADDQ X1,X0", true},
		{"interior target", "CMPQ AX,AX\nJEQ middle\nPADDQ X1,X0\nmiddle:\nPADDQ X1,X0", true},
		{"two returns", "CMPQ AX,AX\nJEQ other\nMOVQ AX,ret+8(FP)\nRET\nother:\nPADDQ X1,X0", true},
		{"unreachable read", "JMP done\nMOVQ 24(SP),AX\ndone:\nPADDQ X1,X0", true},
		{"equal private origins", "MOVQ SP,BX\nCMPQ AX,AX\nJEQ left\nADDQ $8,BX\nJMP joined\nleft:\nADDQ $8,BX\njoined:\nNOP", true},
		{"different private origins", "MOVQ SP,BX\nCMPQ AX,AX\nJEQ left\nADDQ $8,BX\nJMP joined\nleft:\nADDQ $16,BX\njoined:\nNOP", false},
		{"different address cells", "MOVQ SP,BX\nMOVQ BX,24(SP)\nCMPQ AX,AX\nJEQ joined\nMOVQ $0,24(SP)\njoined:\nNOP", false},
		{"missing predecessor bytes", "CMPQ AX,AX\nJEQ joined\nMOVQ AX,24(SP)\njoined:\nMOVQ 24(SP),AX", false},
		{"missing first iteration", "again:\nPADDQ X2,X0\nMOVOU X1,X2\nCMPQ AX,AX\nJNE again", false},
		{"lost backedge flags", "CMPQ AX,AX\nagain:\nSETNE CL\nIMULQ $3,AX\nJCS again", false},
		{"drifting private origin", "MOVQ SP,BX\nagain:\nADDQ $8,BX\nCMPQ AX,AX\nJNE again", false},
		{"one return missing result", "CMPQ AX,AX\nJEQ other\nRET\nother:\nPADDQ X1,X0", false},
		{"no reachable return", "again:\nJMP again", false},
	}
}

func TestGoAMD64FixedLocalReplayControlFlow(t *testing.T) {
	for _, row := range fixedLocalReplayControlCases() {
		t.Run(row.name, func(t *testing.T) {
			fn, sig := fixedLocalReplayFixture(t, row.body)
			_, err := proveX86FixedLocal(fn, sig)
			if (err == nil) != row.valid {
				t.Fatalf("valid=%v: %v", row.valid, err)
			}
			if err != nil && !errors.Is(err, ErrProbeNeedsContext) {
				t.Fatalf("unexpected rejection: %v", err)
			}
		})
	}
}

func TestGoAMD64FixedLocalReplayLinearBoundaries(t *testing.T) {
	for _, row := range []struct {
		name   string
		body   string
		linear []bool
	}{
		{"straight", "NOP\nNOP\nRET", []bool{true, true, true, false}},
		{"label", "NOP\nunused:\nNOP\nRET", []bool{true, true, true, false}},
		{"jump", "JMP next\nnext:\nNOP\nRET", []bool{true, false, true, false}},
		{"duplicate edge", "JEQ next\nnext:\nNOP\nRET", []bool{true, false, true, false}},
		{"interior target", "JEQ next\nNOP\nnext:\nNOP\nRET", []bool{true, false, false, true, false}},
		{"backedge", "again:\nNOP\nNOP\nJMP again", []bool{false, true, true, false}},
		{"unreachable predecessor", "JMP done\nNOP\ndone:\nRET", []bool{true, false, false, false}},
		{"missing return", "NOP", []bool{true, false}},
	} {
		t.Run(row.name, func(t *testing.T) {
			file, err := Parse(ArchAMD64, "TEXT ·Value(SB),4,$64-0\n"+row.body+"\n")
			if err != nil {
				t.Fatal(err)
			}
			fn := file.Funcs[0]
			edges, instructions, err := fixedLocalEdges(fn)
			if err != nil {
				t.Fatal(err)
			}
			if got := fixedLocalLinearEdges(fn, edges, instructions); !reflect.DeepEqual(got, row.linear) {
				t.Fatalf("linear edges = %v, want %v", got, row.linear)
			}
		})
	}
}

func TestGoAMD64FixedLocalReplayMatchesInstructionSchedule(t *testing.T) {
	for _, row := range fixedLocalReplayControlCases() {
		t.Run(row.name, func(t *testing.T) {
			fn, sig := fixedLocalReplayFixture(t, row.body)
			edges, instructions, err := fixedLocalEdges(fn)
			if err != nil {
				t.Fatal(err)
			}
			effects, err := fixedLocalEffectPlan(fn)
			if err != nil {
				t.Fatal(err)
			}
			proof := &x86FixedLocalProof{frame: fn.FrameSize}
			var results [2]Func
			for schedule := range results {
				current := fn
				current.Instrs = append([]Instr(nil), fn.Instrs...)
				var linear []bool
				if schedule == 1 {
					linear = fixedLocalLinearEdges(fn, edges, instructions)
				}
				budget := x86FixedLocalBudget{steps: 1 << 18, bytes: 64 << 20}
				var layouts x86FixedLocalLayoutGraph
				for residue := 0; residue < 8; residue++ {
					err = replayX86FixedLocalCFG(&current, sig, proof, residue, &budget, effects, edges, instructions, linear, nil, &layouts)
					if (err == nil) != row.valid {
						t.Fatalf("schedule=%d residue=%d valid=%v: %v", schedule, residue, row.valid, err)
					}
					if err != nil {
						if !errors.Is(err, ErrProbeNeedsContext) || !reflect.DeepEqual(current.Instrs, fn.Instrs) {
							t.Fatalf("schedule=%d invalid replay changed input or rejection class: %v", schedule, err)
						}
						break
					}
				}
				results[schedule] = current
			}
			if !reflect.DeepEqual(results[0], results[1]) {
				t.Fatal("linear replay changed a source instruction or its eight-residue access certificate")
			}
		})
	}
}

func TestGoAMD64FixedLocalReplayAllocations(t *testing.T) {
	fn, sig := fixedLocalReplayFixture(t, fixedLocalReplayBody("linear", 128))
	edges, instructions, err := fixedLocalEdges(fn)
	if err != nil {
		t.Fatal(err)
	}
	effects, err := fixedLocalEffectPlan(fn)
	if err != nil {
		t.Fatal(err)
	}
	measure := func(linear []bool) float64 {
		return testing.AllocsPerRun(3, func() {
			current := fn
			current.Instrs = append([]Instr(nil), fn.Instrs...)
			proof := &x86FixedLocalProof{frame: fn.FrameSize}
			budget := x86FixedLocalBudget{steps: 1 << 18, bytes: 64 << 20}
			var layouts x86FixedLocalLayoutGraph
			if err := replayX86FixedLocalCFG(&current, sig, proof, 0, &budget, effects, edges, instructions, linear, nil, &layouts); err != nil {
				t.Fatal(err)
			}
		})
	}
	reference := measure(nil)
	blocks := measure(fixedLocalLinearEdges(fn, edges, instructions))
	if blocks >= reference/2 {
		t.Fatalf("linear state copies still dominate: %.0f allocations versus %.0f for instruction scheduling", blocks, reference)
	}
}

func BenchmarkFixedLocalReplayProof(b *testing.B) {
	for _, shape := range []string{"linear", "diamond", "loop"} {
		for _, length := range []int{8, 128, 1024, 5000} {
			b.Run(fmt.Sprintf("%s/%d", shape, length), func(b *testing.B) {
				fn, sig := fixedLocalReplayFixture(b, fixedLocalReplayBody(shape, length))
				if _, err := proveX86FixedLocal(fn, sig); err != nil {
					b.Fatalf("fixture must complete before timing: %v", err)
				}
				b.ReportAllocs()
				b.ResetTimer()
				for n := 0; n < b.N; n++ {
					if _, err := proveX86FixedLocal(fn, sig); err != nil {
						b.Fatal(err)
					}
				}
			})
		}
	}
}
