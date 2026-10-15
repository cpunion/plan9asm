package plan9asm

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

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

func TestX86ContinuationModeAwareBPRegisterRoles(t *testing.T) {
	for _, target := range x86RawReturnTargets {
		for _, row := range []struct{ name, original string }{
			{"numerical", "BX"},
			{"closed-code-table-address", "SI"},
			{"existing-output-pointer", "DI"},
		} {
			for _, api := range []string{"text", "module", "owned-module"} {
				t.Run(target.triple+"/"+row.name+"/"+api, func(t *testing.T) {
					source, options := x86ContinuationResultFixture(t, target.arch, false)
					source = strings.ReplaceAll(source, row.original, "BP")
					options.TargetTriple = target.triple
					ir, err := x86ContinuationResultAPI(t, source, options, api)
					if target.arch == "386" {
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
