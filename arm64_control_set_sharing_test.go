package plan9asm

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

// This is the frozen, pre-sharing set operation, not a second source proof.
// Tests compare complete token maps, including unknown and affine identities.
func arm64ControlUnionReference(left, right arm64ControlValue) arm64ControlValue {
	value := make(arm64ControlValue, len(left)+len(right))
	for token := range left {
		value[token] = true
	}
	for token := range right {
		value[token] = true
	}
	frame := ""
	for token := range value {
		if strings.HasPrefix(token, "sp:") || strings.HasPrefix(token, "fp:") {
			if frame != "" && frame != token {
				for candidate := range value {
					if strings.HasPrefix(candidate, "sp:") || strings.HasPrefix(candidate, "fp:") {
						delete(value, candidate)
					}
				}
				value[""] = true
				break
			}
			frame = token
		}
	}
	return value
}

func arm64ControlSharingState(targets int) *arm64ControlState {
	value := make(arm64ControlValue, targets+1)
	value[""] = true
	for index := 0; index < targets; index++ {
		value[fmt.Sprintf("label:continuation_%d", index)] = true
	}
	state := &arm64ControlState{
		regs:    make(map[Reg]arm64ControlValue),
		memory:  make(map[string]arm64ControlValue),
		escaped: value,
	}
	for index := 0; index < 31; index++ {
		state.regs[Reg(fmt.Sprintf("R%d", index))] = value
	}
	state.regs[SP] = arm64ControlValue{"sp:0": true}
	state.memory["sp:0"] = value
	return state
}

func TestARM64ControlStableMergeDoesNotCopyTargetSets(t *testing.T) {
	state := arm64ControlSharingState(256)
	other := state.clone()
	before := state.clone()
	allocations := testing.AllocsPerRun(10, func() {
		if state.merge(other) {
			t.Fatal("an equal predecessor changed the control proof")
		}
	})
	if !reflect.DeepEqual(state, before) {
		t.Fatal("stable merge changed complete state")
	}
	if allocations != 0 {
		t.Fatalf("stable merge copied unchanged target sets: %.0f allocations per merge", allocations)
	}
}

func arm64ControlSharingCFG(calls int) (Func, error) {
	var source strings.Builder
	source.WriteString("TEXT controlFanout(SB),516,$0-0\nMOVD R30,R19\n")
	for index := 0; index < calls; index++ {
		fmt.Fprintf(&source, "call_%d:\nCALL helper\n", index)
	}
	source.WriteString("done:\nMOVD R19,R30\nRET\nhelper:\nRET\n")
	file, err := Parse(ArchARM64, source.String())
	if err != nil {
		return Func{}, err
	}
	return file.Funcs[0], nil
}

func arm64ControlSharingContext(fn Func) *arm64Ctx {
	return &arm64Ctx{
		blocks:           arm64SplitBlocks(fn),
		normalizedSource: fn,
		sig:              FuncSig{Name: "controlFanout", Ret: Void},
		resolve:          func(name string) string { return name },
	}
}

func TestARM64ControlSharingFirstPartyCFGCompleteTargets(t *testing.T) {
	const calls = 32
	fn, err := arm64ControlSharingCFG(calls)
	if err != nil {
		t.Fatal(err)
	}
	ctx := arm64ControlSharingContext(fn)
	if err := ctx.prepareLocalControl(); err != nil {
		t.Fatal(err)
	}
	helper := len(ctx.blocks) - 1
	targets, err := ctx.localControlTargets(helper)
	if err != nil {
		t.Fatal(err)
	}
	if len(targets) != calls {
		t.Fatalf("helper lost a reaching continuation: got %d, want %d: %v", len(targets), calls, targets)
	}
	for index, reachable := range ctx.localControl.reachable {
		if !reachable {
			t.Fatalf("first-party block %d disappeared from the CFG", index)
		}
	}
}

func TestARM64ControlSharingFrozenReferenceReceipts(t *testing.T) {
	for _, calls := range []int{16, 32, 64} {
		t.Run(fmt.Sprintf("calls_%d", calls), func(t *testing.T) {
			fn, err := arm64ControlSharingCFG(calls)
			if err != nil {
				t.Fatal(err)
			}
			ctx := arm64ControlSharingContext(fn)
			if err := ctx.prepareLocalControl(); err != nil {
				t.Fatal(err)
			}
			// JSON does not support array-valued map keys. Preserve every
			// block/instruction entry in deterministic original CFG order.
			type point struct {
				Block, Instruction  int
				Before, Stack, Link arm64ControlValue
			}
			var points []point
			for block, body := range ctx.blocks {
				for instruction := range body.instrs {
					key := [2]int{block, instruction}
					points = append(points, point{block, instruction,
						ctx.localControl.before[key], ctx.localControl.stackBefore[key], ctx.localControl.linkBefore[key]})
				}
			}
			receipt := struct {
				Original  Func
				Reachable []bool
				Outer     string
				AutoFrame bool
				Points    []point
			}{fn, ctx.localControl.reachable, ctx.localControl.outer, ctx.localControl.autoFrame, points}
			encoded, err := json.Marshal(receipt)
			if err != nil {
				t.Fatal(err)
			}
			t.Logf("complete-control-receipt %s", encoded)
			ir, err := Translate(&File{Arch: ArchARM64, Funcs: []Func{fn}}, Options{
				Goarch: "arm64", TargetTriple: "aarch64-unknown-linux-gnu",
				Sigs: map[string]FuncSig{"controlFanout": {Name: "controlFanout", Ret: Void}},
			})
			if err != nil {
				t.Fatal(err)
			}
			t.Logf("complete-ir-receipt %x", sha256.Sum256([]byte(ir)))
		})
	}
}

func BenchmarkARM64ControlStableMerge(b *testing.B) {
	for _, targets := range []int{64, 256, 1024} {
		b.Run(fmt.Sprintf("targets_%d", targets), func(b *testing.B) {
			state := arm64ControlSharingState(targets)
			other := state.clone()
			b.ReportAllocs()
			b.ResetTimer()
			for index := 0; index < b.N; index++ {
				if state.merge(other) {
					b.Fatal("stable predecessor changed state")
				}
			}
		})
	}
}

func BenchmarkARM64ControlFirstPartyCFG(b *testing.B) {
	for _, calls := range []int{16, 32, 64} {
		b.Run(fmt.Sprintf("calls_%d", calls), func(b *testing.B) {
			fn, err := arm64ControlSharingCFG(calls)
			if err != nil {
				b.Fatal(err)
			}
			b.ReportAllocs()
			b.ResetTimer()
			for index := 0; index < b.N; index++ {
				if err := arm64ControlSharingContext(fn).prepareLocalControl(); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
