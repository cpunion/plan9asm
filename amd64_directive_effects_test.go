package plan9asm

import (
	"fmt"
	"testing"
)

func TestX86DirectiveResultInitialization(t *testing.T) {
	for _, arch := range []string{"amd64", "386"} {
		for _, directive := range []string{
			"PCALIGN $16",
			"PCALIGNMAX $16,$7",
			"PCDATA $0,$-1",
			"FUNCDATA $0,metadata(SB)",
		} {
			t.Run(arch+"/"+directive, func(t *testing.T) {
				body := fmt.Sprintf("MOVL $1,AX\n%s\nMOVL AX,ret0+0(FP)\nRET", directive)
				_, file, sig, _ := x86ResultInitializationFixture(t, arch, "func Value() uint32", body)
				for name, check := range map[string]func(Func, FuncSig, string) error{
					"public": CheckX86ABI0ResultInitialization,
					"private": checkX86ScalarFPResultInitialization,
				} {
					if err := check(file.Funcs[0], sig, arch); err != nil {
						t.Errorf("%s: valid directive changed numerical availability: %v", name, err)
					}
				}
			})
		}
	}
}
