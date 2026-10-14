package plan9asm

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

func x86PhysicalBoolObjectSource(arch string) (string, string, string) {
	word, suffix := int64(8), "Q"
	if arch == "386" {
		word, suffix = 4, "L"
	}
	var source, declarations strings.Builder
	for index, body := range []string{
		"MOVB @WORD(SP),AX\nMOVB AX,@RESULT(SP)\nBYTE $0xc3",
		"MOVB $0,@RESULT(SP)\nBYTE $0xc3",
		"MOVB $1,@RESULT(SP)\nBYTE $0xc3",
		"MOVB @WORD(SP),AH\nMOVB AH,BL\nMOVB BL,@RESULT(SP)\nBYTE $0xc3",
		"MOVB @WORD(SP),AX\nPUSHW AX\nMOVW $0,AX\nPOPW AX\nMOVB AX,@RESULT(SP)\nBYTE $0xc3",
		"MOVB @WORD(SP),AX\nPUSH" + suffix + " AX\nMOV" + suffix + " $0,AX\nPOP" + suffix + " AX\nMOVB AX,@RESULT(SP)\nBYTE $0xc3",
		"MOVB $3,@RESULT(SP)\nMOVB @RESULT(SP),AX\nANDB $1,AX\nMOVB AX,@RESULT(SP)\nBYTE $0xc3",
		"MOVB $2,@RESULT(SP)\nMOVB @RESULT(SP),AX\nSHRB $1,AX\nANDB $1,AX\nMOVB AX,@RESULT(SP)\nBYTE $0xc3",
	} {
		name := fmt.Sprintf("PhysicalBool%d", index)
		text, _, _ := x86SourceStackTypedFixture(arch, "bool", I1, word, "MOVB", body)
		source.WriteString(strings.ReplaceAll(text, "·Value", "·"+name))
		fmt.Fprintf(&declarations, "func %s(x bool) bool\n", name)
	}
	negative := strings.Replace(source.String(), fmt.Sprintf("MOVB $1,%d(SP)", 2*word), fmt.Sprintf("MOVB $3,%d(SP)", 2*word), 1)
	return source.String(), declarations.String(), negative
}

func TestX86PhysicalSourceStackBoolFiveTargetOriginalGoAndLLVMObjects(t *testing.T) {
	llc := findLLVM22Tool("llc")
	if llc == "" {
		t.Fatal("physical source-stack bool objects require LLVM22 llc")
	}
	for _, target := range x86RawReturnTargets {
		t.Run(target.triple, func(t *testing.T) {
			source, declarations, negative := x86PhysicalBoolObjectSource(target.arch)
			if strings.Count(source, "TEXT ·") != 8 || source == negative {
				t.Fatal("complete physical bool source inventory changed")
			}
			x86LEAActualTargetGoObject(t, target.arch, target.triple, source)
			x86LEAActualTargetGoObject(t, target.arch, target.triple, negative)
			for _, api := range []string{"text", "module", "Go"} {
				t.Run(api, func(t *testing.T) {
					ir, err := x86ScalarFPBitViewAPI(t, source, declarations, target.arch, target.triple, api)
					if err != nil || ir == "" {
						t.Fatal("complete original physical bool source rejected", err)
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
					if bad, err := x86ScalarFPBitViewAPI(t, negative, declarations, target.arch, target.triple, api); !errors.Is(err, ErrProbeNeedsContext) || bad != "" {
						t.Fatal("noncanonical complete original physical bool source published LLVM", err)
					}
				})
			}
		})
	}
}
