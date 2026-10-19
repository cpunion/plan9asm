package plan9asm

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestX86PrivatePOPNativeStoresFiveTargets(t *testing.T) {
	llc := findLLVM22Tool("llc")
	if llc == "" {
		t.Fatal("LLVM22 llc required")
	}
	for _, target := range x86RawReturnTargets {
		t.Run(target.triple, func(t *testing.T) {
			move, typ, word, wide := "MOVQ", "uint64", 8, "Q"
			if target.arch == "386" {
				move, typ, word, wide = "MOVL", "uint32", 4, "L"
			}
			for _, width := range []struct {
				suffix string
				bytes  int
			}{{"W", 2}, {wide, word}} {
				t.Run(width.suffix, func(t *testing.T) {
					declaration := fmt.Sprintf("func Value(out *byte,value %s) %s\n", typ, typ)
					source := fmt.Sprintf("TEXT ·Value(SB),4,$0-%d\n%s out+0(FP),DX\nMOVL $1,CX\nPUSH%s value+%d(FP)\nPOP%s -1(DX)(CX*2)\nPUSH%s value+%d(FP)\nPOP%s -1(DX)(CX*2)\n%s value+%d(FP),AX\n%s AX,ret+%d(FP)\nRET\n",
						3*word, move, width.suffix, word, width.suffix, width.suffix, word, width.suffix, move, word, move, 2*word)
					x86LEAActualTargetGoObject(t, target.arch, target.triple, source)
					for _, api := range []string{"text", "module", "Go"} {
						t.Run(api, func(t *testing.T) {
							ir, err := x86ScalarFPBitViewAPI(t, source, declaration, target.arch, target.triple, api)
							if err != nil || ir == "" {
								t.Fatal("actual bounded private POP lost its native write continuation", err)
							}
							for _, input := range []string{ir, x86DwordProductIROptimize(t, ir)} {
								if strings.Count(input, fmt.Sprintf("store volatile i%d ", width.bytes*8)) != 2 {
									t.Fatal("overwritten private POP stores lost exact observable width", width)
								}
								compileLLVMToObject(t, llc, target.triple, "private-pop-store.ll", "private-pop-store.o", input)
							}
						})
					}
					for name, changed := range map[string]string{
						"unknown stack bytes": strings.ReplaceAll(source, "PUSH"+width.suffix+" value+"+fmt.Sprint(word)+"(FP)", "PUSH"+width.suffix+" AX"),
						"unknown index":       strings.Replace(source, "MOVL $1,CX\n", "", 1),
						"unknown base":        strings.Replace(source, move+" out+0(FP),DX\n", "", 1),
						"unbalanced":          strings.Replace(source, "RET\n", "PUSH"+width.suffix+" $1\nRET\n", 1),
					} {
						t.Run(name, func(t *testing.T) {
							for _, api := range []string{"text", "module", "Go"} {
								ir, err := x86ScalarFPBitViewAPI(t, changed, declaration, target.arch, target.triple, api)
								if ir != "" || !errors.Is(err, ErrProbeNeedsContext) {
									t.Fatal("private POP invented stack/address/balance facts", api, err)
								}
							}
						})
					}
				})
			}
		})
	}
}
