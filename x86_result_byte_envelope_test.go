package plan9asm

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestX86ABI0ResultInitializationCompleteStateByteEnvelope(t *testing.T) {
	// 1944 FP bytes + 16 AUTO bytes + (16 GP + 64 X + 8 K + 1 flags)
	// is 2049 logical payload bytes. A linear CFG pays its source scan and
	// eight conservative copy/join units for each visited instruction. This
	// tests the logical work envelope, not Go's actual heap allocation size.
	for _, count := range []int{3639, 3640, 3684} {
		for _, target := range x86RawReturnTargets {
			t.Run(fmt.Sprintf("%d/%s", count, target.triple), func(t *testing.T) {
				body := "MOVL $1,AX\nMOVL AX,spill-4(SP)\n" +
					"MOVL AX,ret0+1936(FP)\nMOVL AX,ret0+1940(FP)\n" +
					strings.Repeat("NOP\n", count-5) + "RET"
				_, _, sig, source := x86ResultInitializationFixture(t, target.arch, "func Value(x [1936]byte) uint64", body)
				source = strings.Replace(source, ",$0-", ",$16-", 1)
				file, err := Parse(ArchAMD64, source)
				if err != nil {
					t.Fatal(err)
				}
				receipt := x86LEAActualTargetGoObject(t, target.arch, target.triple, source)
				t.Logf("actual complete-envelope source object: %+v", receipt)
				if sig.Frame.Results[0].Offset != 1936 || file.Funcs[0].ArgSize != 1944 || len(file.Funcs[0].Instrs) != count+1 {
					t.Fatal("actual typed source dimensions changed")
				}
				for _, query := range []struct {
					name string
					run  func(Func, FuncSig, string) error
				}{
					{"public", CheckX86ABI0ResultInitialization},
					{"private-scalar-views", checkX86ScalarFPResultInitialization},
				} {
					err := query.run(file.Funcs[0], sig, target.arch)
					if count == 3639 {
						if err != nil {
							t.Errorf("%s rejected a bounded complete payload: %v", query.name, err)
						}
					} else if !errors.Is(err, ErrProbeNeedsContext) || !strings.Contains(err.Error(), "byte budget exhausted") {
						t.Errorf("%s undercounted the complete GP/X/K/flag state: %v", query.name, err)
					}
				}
			})
		}
	}
}
