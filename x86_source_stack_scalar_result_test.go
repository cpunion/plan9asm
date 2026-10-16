package plan9asm

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

func x86SourceStackRuntimeDeclarations(arch string) string {
	var declaration strings.Builder
	for _, row := range x86SourceStackRuntimeCases(arch) {
		fmt.Fprintf(&declaration, "func %s(x uint%d) uint%d\n", row.name, row.width, row.width)
	}
	declaration.WriteString("func PhysicalBool(x bool) bool\n")
	declaration.WriteString("func PhysicalBoolIntermediate(x bool) bool\n")
	declaration.WriteString("func PhysicalBoolNamedFP(x bool) bool\n")
	return declaration.String()
}

// Share every original CI kernel, including its physical loads, frame-pointer
// transport, private PUSH/POP, branch join and saved FLAGS. This object owner
// cannot disappear merely because native x86 execution is inapplicable here.
func TestX86PhysicalSourceStackScalarRuntimeSourceThreeAPIsFiveTargets(t *testing.T) {
	llc := findLLVM22Tool("llc")
	if llc == "" {
		t.Fatal("LLVM22 llc required")
	}
	for _, target := range x86RawReturnTargets {
		t.Run(target.triple, func(t *testing.T) {
			source, _ := x86SourceStackRuntimeSource(target.arch)
			declaration := x86SourceStackRuntimeDeclarations(target.arch)
			x86LEAActualTargetGoObject(t, target.arch, target.triple, source)
			for _, api := range []string{"text", "module", "Go"} {
				t.Run(api, func(t *testing.T) {
					ir, err := x86ScalarFPBitViewAPI(t, source, declaration, target.arch, target.triple, api)
					if err != nil {
						t.Fatal("complete physical source stack lost its independent numerical proof", err)
					}
					for _, text := range []string{ir, x86DwordProductIROptimize(t, ir)} {
						for _, level := range []string{"-O0", "-O2"} {
							x86PackedStringCompileObject(t, llc, target.triple, text, level)
						}
					}
				})
			}
		})
	}
}

func TestX86PhysicalSourceStackScalarDoesNotInventDataOrStorage(t *testing.T) {
	for _, target := range x86RawReturnTargets {
		word, suffix, typ, resultType := int64(8), "Q", "uint64", I64
		if target.arch == "386" {
			word, suffix, typ, resultType = 4, "L", "uint32", I32
		}
		read := "MOV" + suffix + " @WORD(SP),AX\n"
		write := "MOV" + suffix + " AX,@RESULT(SP)\nBYTE $0xc3"
		for _, row := range []struct{ name, body string }{
			{"unknown-result", "MOV" + suffix + " BX,@RESULT(SP)\nBYTE $0xc3"},
			{"partial-result", read + "MOVB AX,@RESULT(SP)\nBYTE $0xc3"},
			{"unknown-wide-arithmetic", "MOVB @WORD(SP),AX\nADD" + suffix + " $7,AX\n" + write},
			{"unknown-saved-data", "PUSH" + suffix + " BX\nPOP" + suffix + " AX\n" + write},
			{"unknown-flags", read + "CMP" + suffix + " BX,$0\nJE done\ndone:\n" + write},
			{"unknown-result-on-one-path", read + "CMP" + suffix + " AX,$0\nJE done\nMOV" + suffix + " AX,@RESULT(SP)\ndone:\nBYTE $0xc3"},
			{"input-write", read + "MOV" + suffix + " AX,@WORD(SP)\n" + write},
			{"return-pc-write", read + "MOV" + suffix + " AX,0(SP)\n" + write},
			{"numerical-stack-address", "MOV" + suffix + " SP,AX\n" + write},
			{"unknown-native-store", read + "MOV" + suffix + " AX,(BX)\n" + write},
		} {
			source, _, _ := x86SourceStackTypedFixture(target.arch, typ, resultType, word, "MOV"+suffix, row.body)
			declaration := "func Value(x " + typ + ") " + typ
			for _, api := range []string{"text", "module", "Go"} {
				t.Run(target.triple+"/"+row.name+"/"+api, func(t *testing.T) {
					ir, err := x86ScalarFPBitViewAPI(t, source, declaration, target.arch, target.triple, api)
					if ir != "" || !errors.Is(err, ErrProbeNeedsContext) {
						t.Fatalf("physical storage created unrelated numerical/address facts: %v", err)
					}
				})
			}
		}
	}
}
