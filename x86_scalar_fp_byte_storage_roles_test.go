package plan9asm

import (
	"errors"
	"fmt"
	"testing"
)

func TestX86OrdinaryIntegerFPStoreKeepsIndependentStorageConsumers(t *testing.T) {
	for _, target := range x86RawReturnTargets {
		word, lea := 8, "LEAQ"
		if target.arch == "386" {
			word, lea = 4, "LEAL"
		}
		for _, row := range []struct {
			name, declaration, body string
			extent                  int
			accepted                bool
		}{
			{"direct-defined", "func Value() uint32", "MOVL $9,AX\nMOVL AX,ret+0(FP)", 4, true},
			{"direct-unknown", "func Value() uint32", "MOVL AX,ret+0(FP)", 4, false},
			{"existing-virtual-SP", "func Value() uint32", "MOVL $9,-4(SP)\nMOVL -4(SP),AX\nMOVL AX,ret+0(FP)", 4, true},
			{"existing-FP-address", "func Value(a,b uint32) uint32", lea + " a+0(FP),SI\nMOVL $41,4(SI)\nMOVL b+4(FP),AX\nMOVL AX,ret+8(FP)", 12, true},
			{"existing-FP-input-shadow", "func Value(x uint32) uint32", fmt.Sprintf("MOVL $41,x+0(FP)\nMOVL x+0(FP),AX\nMOVL AX,ret+%d(FP)", word), word + 4, true},
			{"partial-cannot-borrow-SP", "func Value() uint32", "MOVL $9,-4(SP)\nMOVB $1,ret+0(FP)", 4, false},
			{"partial-cannot-borrow-FP-address", "func Value(x uint32) uint32", fmt.Sprintf("%s x+0(FP),SI\nMOVB $1,ret+%d(FP)", lea, word), word + 4, false},
			{"interior-cannot-borrow-SP", "func Value() (uint32,uint8)", "MOVL $9,-4(SP)\nMOVL $0x04030201,ret+0(FP)\nMOVB ret+1(FP),AX\nMOVB AX,ret+4(FP)", 5, false},
			{"cross-field-cannot-borrow-FP-address", "func Value(x uint32) (uint16,uint16)", fmt.Sprintf("%s x+0(FP),SI\nMOVL $0x04030201,ret+%d(FP)", lea, word), word + 4, false},
		} {
			for _, api := range []string{"text", "module", "Go"} {
				t.Run(target.triple+"/"+row.name+"/"+api, func(t *testing.T) {
					source := fmt.Sprintf("TEXT ·Value(SB),4,$0-%d\n%s\nRET\n", row.extent, row.body)
					ir, err := x86ScalarFPBitViewAPI(t, source, row.declaration, target.arch, target.triple, api)
					if row.accepted {
						if err != nil || ir == "" {
							t.Fatal("existing logical storage consumer was replaced by another role", err)
						}
						if row.name == "direct-defined" {
							x86ScalarFPTupleSpanConstantResult(t, ir, []uint64{9})
						}
					} else if !errors.Is(err, ErrProbeNeedsContext) || ir != "" {
						t.Fatalf("new byte view borrowed an independent storage role: output=%t, error=%v", ir != "", err)
					}
				})
			}
		}
	}
}
