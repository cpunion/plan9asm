package plan9asm

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

// Independent architectural EAX/ECX inputs and zero-extended dword outputs.
// GP domain order is AX, CX, DX, BX; these rows do not use the emitter spec.
var x86CPUQueryResultOracle = []struct {
	name            string
	inputs, outputs []int
	registers       []string
	bytes           []byte
}{
	{"CPUID", []int{0, 1}, []int{0, 3, 1, 2}, []string{"AX", "BX", "CX", "DX"}, []byte{0x0f, 0xa2}},
	{"XGETBV", []int{1}, []int{0, 2}, []string{"AX", "DX"}, []byte{0x0f, 0x01, 0xd0}},
}

func TestX86CPUQueryImplicitNUMEffects(t *testing.T) {
	for _, arch := range []string{"386", "amd64"} {
		for _, row := range x86CPUQueryResultOracle {
			t.Run(arch+"/"+row.name, func(t *testing.T) {
				ins := x86ScalarLaneInstruction(t, row.name)
				before := x86ResultState{flags: x86ByteFlagsAll, direction: x86ResultDirectionBackward}
				before.xmm[3], before.masks[2] = 0x8888, true
				for index := range before.pointerGP {
					before.pointerGP[index], before.callerAddressGP[index] = true, true
				}
				for _, index := range row.inputs {
					before.gp[index], before.pointerGP[index] = 0x0f, false
				}
				state, want := before.clone(), before.clone()
				for _, index := range row.outputs {
					want.gp[index], want.pointerGP[index] = 0xff, false
				}
				if !x86ResultTransfer(ins, &state, nil, 0, arch) || !reflect.DeepEqual(state, want) {
					t.Error("query must consume only low dword inputs, define its full outputs and preserve unrelated state")
				}
				effect, known, err := x86TypedInstructionByteEffect(ins.Op, ins, arch)
				if !known || err != nil || !effect.requiresResultProof {
					t.Error("query inputs must remain mandatory even without a result", err)
				}
				for _, input := range row.inputs {
					for missing := uint(0); missing < 4; missing++ {
						state := before.clone()
						state.gp[input] &^= 1 << missing
						unchanged := state.clone()
						if x86ResultTransfer(ins, &state, nil, 0, arch) || !reflect.DeepEqual(state, unchanged) {
							t.Error("query invented a missing input byte or partially wrote its outputs", input, missing)
						}
					}
				}
			})
		}
	}
}

func TestX86CPUQueryVoidRequiresActualInputsThreeAPIs(t *testing.T) {
	for _, target := range x86RawReturnTargets {
		for _, row := range x86CPUQueryResultOracle {
			for _, raw := range []bool{false, true} {
				body := row.name + "\n"
				if raw {
					body = x86RawDisplacementDirectives(row.bytes)
				}
				for _, initialization := range []struct {
					name, source string
					valid        bool
				}{
					{"missing-both", "", false},
					{"missing-cx", "MOVL $0,AX\n", false},
					{"partial-cx", "MOVL $0,AX\nMOVB $0,CL\n", false},
					{"missing-ax", "MOVL $0,CX\n", row.name == "XGETBV"},
					{"complete", "MOVL $0,AX\nMOVL $0,CX\n", true},
				} {
					name := fmt.Sprintf("%s/%s/raw=%t/%s", target.triple, row.name, raw, initialization.name)
					t.Run(name, func(t *testing.T) {
						source := "TEXT ·Query(SB),4,$0-0\n" + initialization.source + body + "RET\n"
						x86LEAActualTargetGoObject(t, target.arch, target.triple, source)
						for _, api := range []string{"text", "module", "Go"} {
							ir, err := x86ScalarFPBitViewAPI(t, source, "func Query()", target.arch, target.triple, api)
							if initialization.valid {
								if err != nil || !strings.Contains(ir, "asm sideeffect \""+strings.ToLower(row.name)+"\"") {
									t.Error("initialized query lost its complete native opcode", api, err)
								}
							} else if ir != "" || !errors.Is(err, ErrProbeNeedsContext) {
								t.Error("Void query borrowed uninitialized implementation-slot inputs", api, err)
							}
						}
					})
				}
			}
		}
	}
}

func TestX86CPUQueryInitializedResultsThreeAPIsFiveTargets(t *testing.T) {
	llc := findLLVM22Tool("llc")
	if llc == "" {
		t.Fatal("LLVM 22 llc required")
	}
	for _, target := range x86RawReturnTargets {
		for _, row := range x86CPUQueryResultOracle {
			for _, raw := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/%s/raw=%t", target.triple, row.name, raw), func(t *testing.T) {
					var source strings.Builder
					fmt.Fprintf(&source, "TEXT ·Query(SB),4,$0-%d\nMOVL leaf+0(FP),AX\nMOVL subleaf+4(FP),CX\n", 8+4*len(row.outputs))
					if raw {
						source.WriteString(x86RawDisplacementDirectives(row.bytes))
					} else {
						source.WriteString(row.name + "\n")
					}
					for index, register := range row.registers {
						fmt.Fprintf(&source, "MOVL %s,out%d+%d(FP)\n", register, index, 8+4*index)
					}
					source.WriteString("RET\n")
					returns := strings.TrimSuffix(strings.Repeat("uint32,", len(row.outputs)), ",")
					declaration := "func Query(leaf,subleaf uint32) (" + returns + ")"
					x86LEAActualTargetGoObject(t, target.arch, target.triple, source.String())
					for _, api := range []string{"text", "module", "Go"} {
						ir, err := x86ScalarFPBitViewAPI(t, source.String(), declaration, target.arch, target.triple, api)
						if err != nil || ir == "" {
							t.Error("query lost its initialized implicit outputs", api, err)
							continue
						}
						for _, text := range []string{ir, x86DwordProductIROptimize(t, ir)} {
							for _, level := range []string{"-O0", "-O2"} {
								x86PackedStringCompileObject(t, llc, target.triple, text, level)
							}
						}
					}
				})
			}
		}
	}
}
