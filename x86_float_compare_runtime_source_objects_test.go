package plan9asm

import (
	"fmt"
	"strings"
	"testing"
)

// Replay the original complete LLVM runtime fixtures, not cropped instruction
// probes. Compilation on this host is not native SSE/VEX/EVEX execution.
func TestX86FloatingCompareOriginalRuntimeProgramsThreeAPIObjects(t *testing.T) {
	llc := findLLVM22Tool("llc")
	if llc == "" {
		t.Fatal("LLVM22 llc required")
	}
	for _, triple := range []string{"x86_64-unknown-linux-gnu", "x86_64-apple-darwin", "x86_64-pc-windows-msvc"} {
		for _, opcode := range x86FloatCompareOpcodes() {
			t.Run(triple+"/"+opcode, func(t *testing.T) {
				cases := x86FloatCompareOracleCases(opcode)
				if len(cases) == 0 {
					t.Fatal("original complete runtime fixture omitted its opcode", opcode)
				}
				source, _ := x86FloatCompareOracleAssembly(cases, false)
				var declarations strings.Builder
				var names []string
				for index := range cases {
					name := fmt.Sprint("compare", index)
					names = append(names, name)
					fmt.Fprintf(&declarations, "func %s(out,a,b *[80]byte,mask uint64)\n", name)
				}
				x86LEAActualTargetGoObject(t, "amd64", triple, source)
				for _, api := range []string{"text", "module", "Go"} {
					t.Run(api, func(t *testing.T) {
						ir, err := x86ScalarFPBitViewAPI(t, source, declarations.String(), "amd64", triple, api)
						if err != nil {
							t.Fatal("complete original floating comparison runtime source remains unclosed", err)
						}
						for _, text := range []string{ir, x86DwordProductIROptimize(t, ir)} {
							x86CompactExpandRequireAllDefinitions(t, text, names)
							for _, level := range []string{"-O0", "-O2"} {
								x86PackedStringCompileObject(t, llc, triple, text, level)
							}
						}
					})
				}
			})
		}
	}
}
