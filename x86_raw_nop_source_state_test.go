package plan9asm

import (
	"fmt"
	"testing"
)

func TestX86RawNOPCompleteSourceStateRetention(t *testing.T) {
	for _, mode := range []int{32, 64} {
		arch := "amd64"
		if mode == 32 {
			arch = "386"
		}
		for _, code := range x86RawNOPWitnesses(mode) {
			t.Run(fmt.Sprintf("mode%d/%x", mode, code), func(t *testing.T) {
				instructions, err := decodeX86RawDirectiveGroup(code, mode, 0, "original NOP", nil)
				if err != nil || len(instructions) != 1 {
					t.Fatal("complete physical NOP handoff", instructions, err)
				}
				ins := instructions[0]
				for _, direction := range []x86ResultDirection{0, 1, 2, 3} {
					if got := x86ResultDirectionAfter(ins, direction); got != direction {
						t.Fatal("inert raw NOP lost or invented DF", direction, got)
					}
				}
				for _, count := range []x86ResultStringCount{x86StringCountUnknown, x86StringCountZero, x86StringCountNonzero} {
					if got := x86ResultStringCountAfter(ins, count, count, arch); got != count {
						t.Fatal("inert raw NOP lost or invented the unchanged CX fact", count, got)
					}
				}
				for _, predicate := range []x86ResultPredicate{{}, {kind: x86ResultNonzeroComparisonBytes, register: 3, width: 32}} {
					if got := x86ResultPredicateAfter(ins, predicate, arch); got != predicate {
						t.Fatal("inert raw NOP changed an existing relation", predicate, got)
					}
				}
				for _, before := range [][16]bool{{}, {3: true, 7: true}} {
					if got := x86ResultCallerAddressesAfter(ins, before, nil, arch); got != before {
						t.Fatal("inert raw NOP lost or invented an existing GP identity", before, got)
					}
				}
			})
		}
	}
}

func TestX86RawNOPPrivateStackAndIDRetention(t *testing.T) {
	for _, arch := range []string{"386", "amd64"} {
		width := "Q"
		if arch == "386" {
			width = "L"
		}
		for _, live := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/live=%t", arch, live), func(t *testing.T) {
				body := x86RawDisplacementDirectives([]byte{0x0f, 0x1f, 0xc0})
				if live {
					body = "PUSH" + width + " $7\n" + body + "POP" + width + " AX\n"
				}
				file, err := Parse(ArchAMD64, "TEXT value(SB),4,$0-0\n"+body+"RET\n")
				if err != nil {
					t.Fatal(err)
				}
				fn, err := decodeX86RawDirectives(file.Funcs[0], arch)
				if err != nil {
					t.Fatal(err)
				}
				for _, initialized := range []bool{false, true} {
					state := x86ResultState{valueStack: &x86ResultValueStack{id: initialized}}
					if initialized {
						state.direction, state.flags = x86ResultDirectionEither, x86ByteFlagsAll
					}
					err = checkX86ResultSourceFlow(fn, state, nil, 0, arch,
						func(_ int, ins Instr, state *x86ResultState, owned []bool, offset int64, arch string) bool {
							return x86ResultTransfer(ins, state, owned, offset, arch)
						}, func(after x86ResultState, _ int) error {
							if after.valueStack.id != initialized || after.direction != state.direction || after.flags != state.flags {
								return fmt.Errorf("inert padding changed private status: before=%+v after=%+v", state, after)
							}
							return nil
						})
					if err != nil {
						t.Fatal("sealed padding changed independent private numerical transport", initialized, err)
					}
				}
			})
		}
	}
}

func TestX86RawNOPOriginalStatusSourceObjects(t *testing.T) {
	llc := findLLVM22Tool("llc")
	if llc == "" {
		t.Fatal("LLVM 22 llc required")
	}
	for _, target := range x86OrderingHintTargets {
		bits := 64
		if target.arch == "386" {
			bits = 32
		}
		declarations, source, _, _ := x86RawNOPRuntimeFixture(bits)
		t.Run(target.triple, func(t *testing.T) {
			goAssemblerOracleOutput(t, target.arch, source, "", true)
			for _, api := range []string{"text", "module", "Go"} {
				ir, err := x86OrderingHintTranslateAPI(t, api, source, declarations, target.arch, target.triple)
				if err != nil {
					t.Fatal("unchanged raw-NOP GP/FLAGS/MXCSR/canary source", api, err)
				}
				for _, text := range []string{ir, x86MulDivRoundOptimize(t, ir)} {
					for _, level := range []string{"-O0", "-O2"} {
						x86PackedStringCompileObject(t, llc, target.triple, text, level)
					}
				}
			}
		})
	}
	t.Log("unchanged original state fixture: Go and LLVM objects, not native x86 execution")
}
