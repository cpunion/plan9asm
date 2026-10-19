package plan9asm

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestX86CMOVNativeCompleteFamilyFiveTargetObjects(t *testing.T) {
	for _, target := range x86RawReturnTargets {
		t.Run(target.triple, func(t *testing.T) {
			word, suffix := goWordSize(target.arch), "Q"
			if word == 4 {
				suffix = "L"
			}
			var source, declarations strings.Builder
			for index, name := range x86CMOVOracleNames(target.arch) {
				function := fmt.Sprint("ConditionalRead", index)
				fmt.Fprintf(&source, "TEXT ·%s(SB),4,$0-%d\nMOV%s p+0(FP),SI\n"+
					"MOV%s old+%d(FP),AX\nPUSH%s flags+%d(FP)\nPOPF%s\n%s (SI),AX\n"+
					"MOV%s AX,ret+%d(FP)\nRET\n", function, 4*word, suffix,
					suffix, word, suffix, 2*word, suffix, name, suffix, 3*word)
				fmt.Fprintf(&declarations, "func %s(p *byte,old,flags uintptr) uintptr\n", function)
			}
			x86RawNumericalFamilyObjects(t, source.String(), declarations.String(), target.arch, target.triple)
		})
	}
}

func TestX86CMOVNativeCannotInventPotentialInputs(t *testing.T) {
	for _, target := range x86RawReturnTargets {
		word, suffix := goWordSize(target.arch), "Q"
		if word == 4 {
			suffix = "L"
		}
		for _, name := range x86CMOVOracleNames(target.arch) {
			for _, role := range []string{"unknown-pointer", "numerical-pointer", "unknown-old", "unknown-flags", "unknown-index"} {
				setup := fmt.Sprintf("MOV%s p+0(FP),SI\nMOV%s old+%d(FP),AX\n"+
					"PUSH%s flags+%d(FP)\nPOPF%s\n", suffix, suffix, word, suffix, 2*word, suffix)
				memory := "(SI)"
				switch role {
				case "unknown-pointer":
					setup = strings.Replace(setup, fmt.Sprintf("MOV%s p+0(FP),SI\n", suffix), "", 1)
				case "numerical-pointer":
					setup = strings.Replace(setup, "p+0(FP)", fmt.Sprintf("old+%d(FP)", word), 1)
				case "unknown-old":
					setup = strings.Replace(setup, fmt.Sprintf("MOV%s old+%d(FP),AX\n", suffix, word), "", 1)
				case "unknown-flags":
					setup = strings.Replace(setup, fmt.Sprintf("PUSH%s flags+%d(FP)\nPOPF%s\n", suffix, 2*word, suffix), "", 1)
				case "unknown-index":
					memory = "(SI)(BX*2)"
				}
				source := fmt.Sprintf("TEXT ·Value(SB),4,$0-%d\n%s%s %s,AX\nRET\n", 3*word, setup, name, memory)
				for _, api := range []string{"text", "module", "Go"} {
					ir, err := x86ScalarFPBitViewAPI(t, source, "func Value(p *byte,old,flags uintptr)", target.arch, target.triple, api)
					if ir != "" || !errors.Is(err, ErrProbeNeedsContext) {
						t.Fatal("native CMOV invented a potential input, including for Void", name, role, target.triple, api, err)
					}
				}
			}
		}
	}
}
