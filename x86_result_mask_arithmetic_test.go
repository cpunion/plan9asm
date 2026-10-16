package plan9asm

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

// Independent of the implementation maps: the complete Go KADD/KSHIFT and
// KUNPCK families have 15 names and two operand rows (_ykaddb/_ykshiftlb).
var x86MaskNumericalOracle = []struct {
	op    string
	bits  int
	shift bool
}{
	{"KADDB", 8, false}, {"KADDW", 16, false},
	{"KADDD", 32, false}, {"KADDQ", 64, false},
	{"KSHIFTLB", 8, true}, {"KSHIFTLW", 16, true},
	{"KSHIFTLD", 32, true}, {"KSHIFTLQ", 64, true},
	{"KSHIFTRB", 8, true}, {"KSHIFTRW", 16, true},
	{"KSHIFTRD", 32, true}, {"KSHIFTRQ", 64, true},
	{"KUNPCKBW", 8, false}, {"KUNPCKWD", 16, false}, {"KUNPCKDQ", 32, false},
}

func TestX86MaskArithmeticNumericalInputsAndState(t *testing.T) {
	for _, arch := range []string{"386", "amd64"} {
		for _, row := range x86MaskNumericalOracle {
			t.Run(arch+"/"+row.op, func(t *testing.T) {
				for destination := 0; destination < 8; destination++ {
					for seed := 0; seed < 256; seed++ {
						first := "K0"
						if row.shift {
							first = "$255"
						}
						ins := x86ScalarLaneInstruction(t, fmt.Sprintf("%s %s,K7,K%d", row.op, first, destination))
						state := x86ResultState{flags: x86ByteFlagsAll}
						state.gp[0], state.pointerGP[0] = 0x93, true
						for mask := range state.masks {
							state.masks[mask] = seed&(1<<mask) != 0
						}
						before := state.clone()
						want := state.masks[7] && (row.shift || state.masks[0])
						if accepted := x86ScalarFPResultTransfer(ins, &state, nil, 0, arch); accepted != want {
							t.Fatalf("missing or invented input: destination=%d seed=%x accepted=%v want=%v", destination, seed, accepted, want)
						}
						if want {
							before.masks[destination] = true
						}
						if !reflect.DeepEqual(state, before) {
							t.Fatal("mask result changed GP, FLAGS, memory or an unrelated mask")
						}
					}
				}
			})
		}
	}
}

func TestX86MaskArithmeticRejectsNonIntegerImmediateConsumers(t *testing.T) {
	for _, row := range x86MaskNumericalOracle {
		if !row.shift {
			continue
		}
		for _, value := range []string{"$0.0", "$1.0", "$UNRESOLVED", "$-1", "$256"} {
			t.Run(row.op+"/"+value, func(t *testing.T) {
				ins := x86ScalarLaneInstruction(t, row.op+" "+value+",K1,K2")
				for _, arch := range []string{"386", "amd64"} {
					control := x86ScalarLaneInstruction(t, row.op+" $1,K1,K2")
					consumer, _ := newAMD64CtxWithFuncForTest(t, Func{Instrs: []Instr{control}}, FuncSig{Ret: Void}, nil)
					consumer.goarch = arch
					if err := consumer.lowerMaskArithmetic(control.Op, control); err != nil {
						t.Fatal("valid immediate control did not reach the lowerer", err)
					}
					if err := consumer.lowerMaskArithmetic(ins.Op, ins); err == nil {
						t.Fatal("lowerer accepted a non-Yu8 immediate", arch, value)
					}
					state := x86ResultState{}
					state.masks[1] = true
					before := state.clone()
					if x86ScalarFPResultTransfer(ins, &state, nil, 0, arch) || !reflect.DeepEqual(before, state) {
						t.Fatal("invalid immediate acquired numerical authority", arch, value)
					}
				}
			})
		}
	}
}

func TestX86MaskArithmeticVoidCannotInventInputs(t *testing.T) {
	for _, row := range x86MaskNumericalOracle {
		t.Run(row.op, func(t *testing.T) {
			first := "K0"
			if row.shift {
				first = "$255"
			}
			source := "TEXT ·Value(SB),4,$0-0\n" + row.op + " " + first + ",K7,K3\nRET\n"
			goAssemblerOracleOutput(t, "amd64", source, "", true)
			for _, arch := range []string{"386", "amd64"} {
				triple := "x86_64-unknown-linux-gnu"
				if arch == "386" {
					triple = "i386-unknown-linux-gnu"
				}
				for _, api := range []string{"text", "module", "Go"} {
					ir, err := x86ScalarFPBitViewAPI(t, source, "func Value()", arch, triple, api)
					if ir != "" || !errors.Is(err, ErrProbeNeedsContext) {
						t.Fatal("unknown mask input borrowed implementation zeros", arch, api, err)
					}
				}
			}
		})
	}
}

func TestX86MaskArithmeticCompleteResultFiveTargetsThreeAPIs(t *testing.T) {
	llc := findLLVM22Tool("llc")
	if llc == "" {
		t.Fatal("LLVM22 llc required")
	}
	for _, target := range x86RawReturnTargets {
		t.Run(target.triple, func(t *testing.T) {
			word, move := 8, "MOVQ"
			if target.arch == "386" {
				word, move = 4, "MOVL"
			}
			var source, declarations strings.Builder
			var names []string
			for _, row := range x86MaskNumericalOracle {
				counts := []int{0}
				if row.shift {
					counts = []int{0, row.bits - 1, row.bits, 255}
				}
				for _, count := range counts {
					name := fmt.Sprintf("MaskResult%d", len(names))
					names = append(names, name)
					first := "K0"
					if row.shift {
						first = fmt.Sprintf("$%d", count)
					}
					fmt.Fprintf(&declarations, "func %s(a,b uintptr) uintptr\n", name)
					fmt.Fprintf(&source, "TEXT ·%s(SB),4,$0-%d\n%s a+0(FP),AX\nKMOVD AX,K0\n"+
						"%s b+%d(FP),AX\nKMOVD AX,K7\n%s %s,K7,K3\nKMOVD K3,BX\n%s BX,ret+%d(FP)\nRET\n",
						name, 3*word, move, move, word, row.op, first, move, 2*word)
				}
			}
			if len(names) != 39 {
				t.Fatal("complete operation/count denominator changed", len(names))
			}
			x86LEAActualTargetGoObject(t, target.arch, target.triple, source.String())
			for _, api := range []string{"text", "module", "Go"} {
				ir, err := x86ScalarFPBitViewAPI(t, source.String(), declarations.String(), target.arch, target.triple, api)
				if err != nil || ir == "" {
					t.Fatal("complete initialized mask result rejected", api, err)
				}
				x86CompactExpandRequireAllDefinitions(t, ir, names)
				for _, text := range []string{ir, x86DwordProductIROptimize(t, ir)} {
					for _, level := range []string{"-O0", "-O2"} {
						x86PackedStringCompileObject(t, llc, target.triple, text, level)
					}
				}
			}
		})
	}
}
