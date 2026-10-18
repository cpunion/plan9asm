package plan9asm

import (
	"fmt"
	"reflect"
	"testing"
)

// Intel's ENDBR32/64 operations update CET tracking, not numerical registers,
// FLAGS or memory. Keep the native instruction; these checks grant no CET entry.
// https://cdrdv2-public.intel.com/874239/252046-082-sdm-change-document.pdf
func TestX86EndBranchNumericalState(t *testing.T) {
	for _, arch := range []string{"386", "amd64"} {
		forms := []Instr{{Op: "ENDBR64"}}
		for _, last := range []byte{0xfa, 0xfb} {
			forms = append(forms, x86RawNOPSourceStateInstruction(t, arch, []byte{0xf3, 0x0f, 0x1e, last}))
		}
		for index, ins := range forms {
			t.Run(fmt.Sprintf("%s/form%d", arch, index), func(t *testing.T) {
				for _, consumer := range x86ResultGPSIMDConsumers {
					for _, initialized := range []bool{false, true} {
						state := x86ResultState{fp: []bool{false, true}, valueStack: &x86ResultValueStack{}}
						if initialized {
							state.gp[0], state.xmm[0], state.vectorUpper[0] = 0x53, 0x5511, 0xaabbccdd00000000
							state.pointerGP[3], state.masks[1], state.flags = true, true, x86ByteFlagsAll
							state.direction, state.valueStack.id = x86ResultDirectionBackward, true
						}
						before := state.clone()
						if !consumer.step(ins, &state, nil, 0, arch) || !reflect.DeepEqual(state, before) {
							t.Errorf("ENDBR consumed or invented numerical state: %s initialized=%t", consumer.name, initialized)
						}
					}
				}
				for _, direction := range []x86ResultDirection{0, 1, 2, 3} {
					if got := x86ResultDirectionAfter(ins, direction); got != direction {
						t.Error("ENDBR changed existing DF knowledge", direction, got)
					}
				}
				predicate := x86ResultPredicate{kind: x86ResultComparisonBytes, register: 1, width: 16}
				if x86ResultPredicateAfter(ins, predicate, arch) != predicate {
					t.Error("ENDBR revoked an unchanged numerical predicate")
				}
				addresses := [16]bool{3: true, 7: true}
				if x86ResultCallerAddressesAfter(ins, addresses, nil, arch) != addresses {
					t.Error("ENDBR changed existing address identities")
				}
				for _, count := range []x86ResultStringCount{x86StringCountUnknown, x86StringCountZero, x86StringCountNonzero} {
					if x86ResultStringCountAfter(ins, count, count, arch) != count {
						t.Error("ENDBR changed existing CX knowledge")
					}
				}
				state := x86ResultState{valueStack: &x86ResultValueStack{bytes: make([]uint8, 8)}}
				if !x86ResultValueStackContinuation(ins, &state, arch) {
					t.Error("ENDBR revoked unrelated private storage")
				}
			})
		}
	}
}

func TestX86EndBranchEffectsRequireExactSource(t *testing.T) {
	for _, arch := range []string{"386", "amd64"} {
		for _, last := range []byte{0xfa, 0xfb} {
			original := x86RawNOPSourceStateInstruction(t, arch, []byte{0xf3, 0x0f, 0x1e, last})
			for _, mutation := range []string{
				"opcode", "operands", "origin opcode", "origin operands", "missing origin",
				"missing encoding", "address", "unbound address", "segment", "raw return",
			} {
				t.Run(fmt.Sprintf("%s/%x/%s", arch, last, mutation), func(t *testing.T) {
					ins := original
					origin := *original.x86SourceRaw
					ins.x86SourceRaw = &origin
					switch mutation {
					case "opcode":
						ins.Op = "ENDBR64.Z"
					case "operands":
						ins.Args = []Operand{{Kind: OpReg, Reg: AX}}
					case "origin opcode":
						origin.op = "SWAPGS"
					case "origin operands":
						origin.args = []Operand{{Kind: OpReg, Reg: AX}}
					case "missing origin":
						ins.x86SourceRaw = nil
					case "missing encoding":
						ins.x86Encoded = false
					case "address":
						ins.x86AddressBits = 32
					case "unbound address":
						ins.x86UnboundAddressOverride = true
					case "segment":
						ins.x86SegmentPrefix = 0x64
					case "raw return":
						ins.x86RawNearReturn = new(x86RawNearReturnForm)
					}
					state := x86ResultState{flags: x86ByteFlagsAll}
					before := state.clone()
					if x86ResultTransfer(ins, &state, nil, 0, arch) || !reflect.DeepEqual(state, before) {
						t.Fatal("changed physical source gained ENDBR numerical effects")
					}
				})
			}
		}
	}
}

func TestX86EndBranchOriginalRuntimeObjects(t *testing.T) {
	format := x86X87FunctionLiteral(t, "amd64_raw_endbranch_test.go", "TestX86RawEndBranchRuntimePreservesState", "source")
	llc := findLLVM22Tool("llc")
	if llc == "" {
		t.Fatal("LLVM 22 llc required")
	}
	for _, target := range x86RawReturnTargets {
		if target.arch != "amd64" {
			continue // The unchanged runtime program uses a 64-bit pointer declaration.
		}
		for _, last := range []byte{0xfa, 0xfb} {
			t.Run(fmt.Sprintf("%s/%x", target.triple, last), func(t *testing.T) {
				source := fmt.Sprintf(format, last)
				x86LEAActualTargetGoObject(t, target.arch, target.triple, "#include \"textflag.h\"\n"+source)
				for _, api := range []string{"text", "module", "Go"} {
					ir, err := x86ScalarFPBitViewAPI(t, source, "func endbranchstate(out *byte)", target.arch, target.triple, api)
					if err != nil {
						t.Fatal("unchanged original GP/FLAGS ENDBR program", api, err)
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
