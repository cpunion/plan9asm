package plan9asm

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

// Independent complete Go yscond name inventory, not the implementation map.
var x86NativeSetOracle = strings.Fields("SETCC SETCS SETEQ SETGE SETGT SETHI SETLE SETLS SETLT SETMI SETNE SETOC SETOS SETPC SETPL SETPS")

func x86NativeSetFixture(arch string) (string, string, []string) {
	word, move := 8, "MOVQ"
	if arch == "386" {
		word, move = 4, "MOVL"
	}
	var source, declarations strings.Builder
	var names []string
	for _, op := range x86NativeSetOracle {
		name := "Native" + op
		names = append(names, name)
		fmt.Fprintf(&declarations, "func %s(p *[8]byte,x,y uint32)\n", name)
		fmt.Fprintf(&source, "TEXT ·%s(SB),4,$0-%d\n%s p+0(FP),AX\nMOVL x+%d(FP),BX\n"+
			"MOVL y+%d(FP),CX\nCMPL BX,CX\n%s 3(AX)\nRET\n", name, word+8, move, word, word+4, op)
	}
	return source.String(), declarations.String(), names
}

func TestX86NativeConditionalSetCompleteFiveTargetsThreeAPIs(t *testing.T) {
	llc := findLLVM22Tool("llc")
	if llc == "" {
		t.Fatal("LLVM22 llc required")
	}
	for _, target := range x86RawReturnTargets {
		t.Run(target.triple, func(t *testing.T) {
			source, declarations, names := x86NativeSetFixture(target.arch)
			if len(names) != 16 {
				t.Fatal("complete yscond denominator changed")
			}
			x86LEAActualTargetGoObject(t, target.arch, target.triple, source)
			for _, api := range []string{"text", "module", "Go"} {
				ir, err := x86ScalarFPBitViewAPI(t, source, declarations, target.arch, target.triple, api)
				if err != nil || ir == "" {
					t.Fatal("initialized native SETcc rejected", api, err)
				}
				x86CompactExpandRequireAllDefinitions(t, ir, names)
				for _, text := range []string{ir, x86DwordProductIROptimize(t, ir)} {
					if strings.Count(text, "store volatile i8 ") != len(names) {
						t.Fatal("SETcc lost an observable exact-byte write", api)
					}
					for _, level := range []string{"-O0", "-O2"} {
						x86PackedStringCompileObject(t, llc, target.triple, text, level)
					}
				}
			}
		})
	}
}

func TestX86NativeConditionalSetCannotInventFlagsOrAddress(t *testing.T) {
	for _, arch := range []string{"386", "amd64"} {
		source, declarations, _ := x86NativeSetFixture(arch)
		triple := "x86_64-unknown-linux-gnu"
		if arch == "386" {
			triple = "i386-unknown-linux-gnu"
		}
		for name, changed := range map[string]string{
			"unknown-flags": strings.ReplaceAll(source, "CMPL BX,CX\n", ""),
			"unknown-base":  strings.ReplaceAll(source, "p+0(FP),AX", "p+0(FP),DI"),
			"numeric-base":  strings.ReplaceAll(source, "CMPL BX,CX", "MOVL $42,AX\nCMPL BX,CX"),
		} {
			t.Run(arch+"/"+name, func(t *testing.T) {
				for _, api := range []string{"text", "module", "Go"} {
					ir, err := x86ScalarFPBitViewAPI(t, changed, declarations, arch, triple, api)
					if ir != "" || !errors.Is(err, ErrProbeNeedsContext) {
						t.Fatal("native SETcc invented a source flag or pointer", api, err)
					}
				}
			})
		}
	}
}
