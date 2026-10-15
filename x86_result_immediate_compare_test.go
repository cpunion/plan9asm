package plan9asm

import (
	"fmt"
	"reflect"
	"testing"
)

// This inventory follows all eight Go _yvpcmpb members independently of the
// implementation's lowering table. Every member has X/Y/Z and an optional
// K1-K7 input mask; its result is K0-K7, never a vector merge destination.
var x86ImmediateCompareResultMembers = []string{
	"VPCMPB", "VPCMPW", "VPCMPD", "VPCMPQ",
	"VPCMPUB", "VPCMPUW", "VPCMPUD", "VPCMPUQ",
}

func TestX86ImmediatePackedCompareMaskAvailability(t *testing.T) {
	for _, op := range x86ImmediateCompareResultMembers {
		for _, vector := range []struct {
			bank  string
			bytes int
		}{{"X", 16}, {"Y", 32}, {"Z", 64}} {
			for _, masked := range []bool{false, true} {
				name := fmt.Sprintf("%s/%s/masked-%t", op, vector.bank, masked)
				t.Run(name, func(t *testing.T) {
					mask := ""
					if masked {
						mask = "K2,"
					}
					row := fmt.Sprintf("%s $4,%s0,%s1,%sK2", op, vector.bank, vector.bank, mask)
					ins := x86ScalarLaneInstruction(t, row)
					state := x86ResultState{flags: x86ByteFlagsAll, direction: x86ResultDirectionForward}
					state.writeVector(0, 64, ^uint64(0), false)
					state.writeVector(1, 64, ^uint64(0), false)
					state.masks[2] = masked
					before := state.clone()
					if !x86ResultTransfer(ins, &state, nil, 0, "amd64") {
						t.Fatal("complete Go immediate comparison has no numerical mask-result effect", row)
					}
					before.masks[2] = true
					if !reflect.DeepEqual(state, before) {
						t.Fatal("mask comparison changed vectors, FLAGS, DF or unrelated state")
					}
				})
			}
		}
	}
}

func TestX86ImmediatePackedCompareMissingBytesRemainUnknown(t *testing.T) {
	for _, op := range x86ImmediateCompareResultMembers {
		for _, vector := range []struct {
			bank  string
			bytes int
		}{{"X", 16}, {"Y", 32}, {"Z", 64}} {
			t.Run(op+"/"+vector.bank, func(t *testing.T) {
				for _, predicate := range []int{3, 4, 7} {
					row := fmt.Sprintf("%s $%d,%s0,%s1,K1,K2", op, predicate, vector.bank, vector.bank)
					ins := x86ScalarLaneInstruction(t, row)
					for _, maskKnown := range []bool{false, true} {
						for input := 0; input < 2; input++ {
							for missing := 0; missing < vector.bytes; missing++ {
								state := x86ResultState{flags: x86ByteFlagsAll, masks: [8]bool{1: maskKnown}}
								state.writeVector(0, 64, ^uint64(0), false)
								state.writeVector(1, 64, ^uint64(0), false)
								state.writeVector(input, 64, ^(uint64(1) << uint(missing)), false)
								before := state.clone()
								if x86ResultTransfer(ins, &state, nil, 0, "amd64") || !reflect.DeepEqual(state, before) {
									t.Fatalf("undefined comparison became numerical: %s mask=%t input=%d missing=%d", row, maskKnown, input, missing)
								}
							}
						}
					}
				}
			})
		}
	}
}

func TestX86ImmediatePackedCompareResultThreeAPIs(t *testing.T) {
	for _, target := range x86RawReturnTargets {
		if target.arch != "amd64" {
			continue // Go's mode32 frontend rejects these four/five-operand rows.
		}
		for _, op := range x86ImmediateCompareResultMembers {
			for _, bank := range []string{"X", "Y", "Z"} {
				body := fmt.Sprintf("MOVQ x+0(FP),AX\nMOVQ AX,X0\nVPBROADCASTQ X0,Z0\n"+
					"VPXORD Z1,Z1,Z1\n%s $4,%s1,%s0,K2\nKMOVQ K2,AX\nMOVQ AX,ret+8(FP)\nRET\n", op, bank, bank)
				source := "TEXT ·Value(SB),4,$0-16\n" + body
				for _, api := range []string{"text", "module", "Go"} {
					t.Run(target.triple+"/"+op+"/"+bank+"/"+api, func(t *testing.T) {
						ir, err := x86ScalarFPBitViewAPI(t, source, "func Value(x uint64) uint64", target.arch, target.triple, api)
						if err != nil || ir == "" {
							t.Fatal("initialized immediate compare cannot publish its mask result", err)
						}
						bits := 8
						switch op {
						case "VPCMPW", "VPCMPUW":
							bits = 16
						case "VPCMPD", "VPCMPUD":
							bits = 32
						case "VPCMPQ", "VPCMPUQ":
							bits = 64
						}
						lanes := map[string]int{"X": 128, "Y": 256, "Z": 512}[bank] / bits
						for _, input := range []struct {
							literal string
							want    uint64
						}{{"0", 0}, {"-1", ^uint64(0) >> uint(64-lanes)}} {
							probe := fmt.Sprintf("\ndefine i64 @Probe() {\n %%r = call i64 @Value(i64 %s)\n ret i64 %%r\n}\n", input.literal)
							x86ScalarFPNamedConstantResult(t, ir+probe, "Probe", []uint64{input.want})
						}
					})
				}
			}
		}
	}
}
