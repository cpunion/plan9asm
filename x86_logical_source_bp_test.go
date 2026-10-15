package plan9asm

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestX86OrdinaryMemoryModeAwareFrameRegisters(t *testing.T) {
	for _, arch := range []string{"386", "amd64"} {
		for _, base := range []Reg{AX, BP, SP} {
			for _, index := range []Reg{"", AX, BP, SP} {
				t.Run(arch+"/"+string(base)+"/"+string(index), func(t *testing.T) {
					memory := MemRef{Base: base, Index: index}
					if index != "" {
						memory.Scale = 1
					}
					want := base != SP && index != SP && (arch == "386" || base != BP && index != BP)
					if got := x86OrdinaryNativeMemoryOperand(Operand{Kind: OpMem, Mem: memory}, arch); got != want {
						t.Fatal("ordinary native address classification differs from the selected GP mode", got, want)
					}
				})
			}
		}
	}
}

func TestX86DirectTailModeAwareBPRegisterRoles(t *testing.T) {
	for _, target := range x86RawReturnTargets {
		word, width, typ := "Q", 8, "uint64"
		if target.arch == "386" {
			word, width, typ = "L", 4, "uint32"
		}
		for _, row := range []struct {
			name, setup, body string
			complete          bool
		}{
			{"numerical", "MOVWORD n+0(FP),BP", "ADDWORD $7,BP\nMOVWORD BP,ret+RESULT(FP)", true},
			{"helper-write", "NOP", "MOVWORD n+0(FP),BP\nMOVWORD BP,ret+RESULT(FP)", true},
			{"logical-result-address", "MOVWORD n+0(FP),SI\nLEAWORD ret+RESULT(FP),BP", "MOVWORD SI,(BP)", true},
			{"unknown-entry", "NOP", "MOVWORD BP,ret+RESULT(FP)", false},
			{"partial-word", "MOVW $7,BP", "MOVWORD BP,ret+RESULT(FP)", false},
			{"unknown-overwrite", "MOVWORD n+0(FP),BP", "MOVWORD CX,BP\nMOVWORD BP,ret+RESULT(FP)", false},
			{"unbound-memory", "MOVWORD n+0(FP),BP", "MOVWORD (BP),AX\nMOVWORD AX,ret+RESULT(FP)", false},
			{"physical-SP", "MOVWORD n+0(FP),SP", "MOVWORD SP,ret+RESULT(FP)", false},
			{"physical-SP-memory", "NOP", "MOVWORD (SP),AX\nMOVWORD AX,ret+RESULT(FP)", false},
		} {
			for _, api := range []string{"text", "module", "Go"} {
				t.Run(target.triple+"/"+row.name+"/"+api, func(t *testing.T) {
					source := fmt.Sprintf("TEXT ·Value(SB),NOSPLIT,$0-%d\n%s\nJMP body<>(SB)\nTEXT body<>(SB),NOSPLIT,$0-0\n%s\nRET\n",
						2*width, row.setup, row.body)
					source = strings.NewReplacer("WORD", word, "RESULT", fmt.Sprint(width)).Replace(source)
					ir, err := x86ScalarFPBitViewAPI(t, source, "func Value(n "+typ+") "+typ, target.arch, target.triple, api)
					wantPass := target.arch == "386" && row.complete
					if wantPass {
						if err != nil || ir == "" {
							t.Fatal("ordinary mode32 BP lost its source-defined logical role", err)
						}
					} else if !errors.Is(err, ErrProbeNeedsContext) || ir != "" {
						t.Fatal("BP classification granted an unbound register/frame/memory role", err)
					}
				})
			}
		}
	}
}

func x86LogicalSourceBPContinuation(t *testing.T, arch, role string) (string, Options) {
	t.Helper()
	source, options := x86ContinuationResultFixture(t, arch, true)
	if register := map[string]string{"numerical": "BX", "code-table": "SI", "output-pointer": "DI"}[role]; register != "" {
		return strings.ReplaceAll(source, register, "BP"), options
	}
	word := "Q"
	if arch == "386" {
		word = "L"
	}
	body := " MOV" + word + " (SI),BP\n JMP BP"
	switch role {
	case "code-target":
	case "partial-target":
		body = " MOVW (SI),BP\n JMP BP"
	case "unknown-target":
		body = " MOV" + word + " $0,BP\n JMP BP"
	case "overwritten-target":
		body = " MOV" + word + " (SI),BP\n MOVB $0,BP\n JMP BP"
	default:
		t.Fatal("unknown BP continuation fixture role", role)
	}
	return strings.Replace(source, " JMP (SI)", body, 1), options
}

func TestX86ContinuationModeAwareBPRegisterRoles(t *testing.T) {
	for _, target := range x86RawReturnTargets {
		for _, row := range []struct {
			name     string
			complete bool
		}{
			{"numerical", true},
			{"code-table", true},
			{"output-pointer", true},
			{"code-target", true},
			{"partial-target", false},
			{"unknown-target", false},
			{"overwritten-target", false},
		} {
			for _, api := range []string{"text", "module", "owned-module"} {
				t.Run(target.triple+"/"+row.name+"/"+api, func(t *testing.T) {
					source, options := x86LogicalSourceBPContinuation(t, target.arch, row.name)
					options.TargetTriple = target.triple
					ir, err := x86ContinuationResultAPI(t, source, options, api)
					if target.arch == "386" && row.complete {
						if err != nil || ir == "" {
							t.Fatal("mode32 BP lost its independently closed continuation role", err)
						}
					} else if !errors.Is(err, ErrProbeNeedsContext) || ir != "" {
						t.Fatal("logical continuation admitted the AMD64 physical BP frame", err)
					}
				})
			}
		}
	}
}

func TestX86Mode32BPLogicalSourceOriginalGoAndLLVMObjects(t *testing.T) {
	llc := findLLVM22Tool("llc")
	if llc == "" {
		t.Fatal("LLVM22 llc required for mode32 logical BP objects")
	}
	for _, target := range x86RawReturnTargets {
		if target.arch != "386" {
			continue // This owner explicitly compiles the two mode32 targets.
		}
		for _, role := range []string{"direct-tail", "numerical", "code-table", "output-pointer", "code-target"} {
			t.Run(target.triple+"/"+role, func(t *testing.T) {
				source := strings.ReplaceAll(x86DirectTail386ASM, "SI", "BP")
				var options Options
				if role != "direct-tail" {
					source, options = x86LogicalSourceBPContinuation(t, "386", role)
					options.TargetTriple = target.triple
				}
				x86LEAActualTargetGoObject(t, target.arch, target.triple, source)
				apis := []string{"text", "module", "owned-module"}
				if role == "direct-tail" {
					apis[2] = "Go"
				}
				for _, api := range apis {
					t.Run(api, func(t *testing.T) {
						var ir string
						var err error
						if role == "direct-tail" {
							ir, err = x86ScalarFPBitViewAPI(t, source, x86DirectTail386Go, target.arch, target.triple, api)
						} else {
							ir, err = x86ContinuationResultAPI(t, source, options, api)
						}
						if err != nil {
							t.Fatal(err)
						}
						for _, optimized := range []bool{false, true} {
							input := ir
							if optimized {
								input = x86UndefinedTrapOptimizedIR(t, input)
							}
							for _, level := range []string{"-O0", "-O2"} {
								t.Run(fmt.Sprintf("IR_O2=%t/backend%s", optimized, level), func(t *testing.T) {
									x86PackedStringCompileObject(t, llc, target.triple, input, level)
								})
							}
						}
					})
				}
			})
		}
	}
}
