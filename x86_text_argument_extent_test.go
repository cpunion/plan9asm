package plan9asm

import (
	"errors"
	"testing"
)

// cmd/asm preserves an omitted argument size as ArgsSizeUnknown, not zero.
// An ordinary typed declaration still bounds every actual FP input; omission
// grants neither an unbounded frame nor an explicit extent mismatch.
func TestX86OmittedTEXTArgsKeepTypedVoidExtent(t *testing.T) {
	llc := findLLVM22Tool("llc")
	if llc == "" {
		t.Fatal("LLVM 22 llc required")
	}
	for _, target := range x86RawReturnTargets {
		for _, row := range []struct {
			extent string
			valid  bool
		}{
			{"$0", true}, {"$0-4", true}, {"$0-8", true},
			{"$0-0", false}, {"$0-3", false}, {"$0-4097", false},
		} {
			t.Run(target.triple+"/"+row.extent, func(t *testing.T) {
				source := "TEXT ·Value(SB),4," + row.extent + "\nKMOVD value+0(FP),K1\nRET\n"
				x86LEAActualTargetGoObject(t, target.arch, target.triple, source)
				for _, api := range []string{"text", "module", "Go"} {
					ir, err := x86ScalarFPBitViewAPI(t, source, "func Value(value uint32)\n", target.arch, target.triple, api)
					if !row.valid {
						if ir != "" || !errors.Is(err, ErrProbeNeedsContext) {
							t.Fatal("explicit TEXT mismatch gained typed argument authority", api, err)
						}
						continue
					}
					if err != nil {
						t.Fatal("valid omitted or matching argument extent rejected", api, err)
					}
					compileLLVMToObject(t, llc, target.triple, "text-extent.ll", "text-extent.o", ir)
				}
			})
		}
	}
}
