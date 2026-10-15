package plan9asm

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

type x86Legacy386SplitInputCase struct {
	name, typ, nonzero string
	bytes, half        int
}

// This inventory is independent of the new byte-view selector. The old 386
// evalFPToI64 reader already bitcasts complete float inputs and obtains an
// I64/double high word at +4. It never treats an arbitrary offset as that role.
var x86Legacy386SplitInputCases = []x86Legacy386SplitInputCase{
	{"natural-I32", "uint32", "1", 4, 0},
	{"natural-F32", "float32", "0x3FF0000020000000", 4, 0},
	{"low-I64", "uint64", "4294967297", 8, 0},
	{"high-I64", "uint64", "4294967297", 8, 4},
	{"low-F64", "float64", "0x3FF0000100000001", 8, 0},
	{"high-F64", "float64", "0x3FF0000100000001", 8, 4},
}

func x86Legacy386SplitInputFloatSource(row x86Legacy386SplitInputCase) (string, string) {
	result := row.bytes + 8
	// Both branches initialize the same one-element x87 stack before its
	// complete floating result store. The tested input word affects control,
	// so an unresolved-offset zero cannot impersonate the legacy reader.
	body := fmt.Sprintf("MOVL x+%d(FP),AX\nANDL $1,AX\nTESTL AX,AX\nJEQ one\nFMOVD q+%d(FP),F0\nJMP done\none:\nFLD1\ndone:\nFMOVDP F0,ret+%d(FP)\nRET\n",
		row.half, row.bytes, result)
	source := fmt.Sprintf("TEXT ·Value(SB),4,$0-%d\n%s", result+8, body)
	return source, "func Value(x " + row.typ + ", q float64) float64"
}

func TestX86Legacy386SplitInputFloatResultThreeAPIs(t *testing.T) {
	for _, target := range x86RawReturnTargets {
		if target.arch != "386" {
			continue // This exact old source-half transport is mode32-only.
		}
		for _, row := range x86Legacy386SplitInputCases {
			source, declaration := x86Legacy386SplitInputFloatSource(row)
			for _, api := range []string{"text", "module", "Go"} {
				t.Run(target.triple+"/"+row.name+"/"+api, func(t *testing.T) {
					ir, err := x86ScalarFPBitViewAPI(t, source, declaration, target.arch, target.triple, api)
					if err != nil || ir == "" {
						t.Fatal("existing typed split-input/floating-result consumer was intercepted", err)
					}
					inputType := map[string]string{"uint32": "i32", "uint64": "i64", "float32": "float", "float64": "double"}[row.typ]
					zero := "0"
					if strings.HasPrefix(row.typ, "float") {
						zero = "0.000000e+00"
					}
					for _, input := range []struct {
						literal string
						bits    uint64
					}{{zero, 0x3ff0000000000000}, {row.nonzero, 0x4000000000000000}} {
						// This test-owned caller observes exact returned bits after
						// in-process optimization, not native x87 execution.
						probe := fmt.Sprintf("\ndefine i64 @Probe() {\n %%r = call double @Value(%s %s, double 2.000000e+00)\n %%b = bitcast double %%r to i64\n ret i64 %%b\n}\n", inputType, input.literal)
						x86ScalarFPNamedConstantResult(t, ir+probe, "Probe", []uint64{input.bits})
					}
				})
			}
		}
	}
}

func TestX86Legacy386SplitInputCannotBypassNewViewOrNUMProof(t *testing.T) {
	base, declaration := x86Legacy386SplitInputFloatSource(x86Legacy386SplitInputCases[4])
	for _, target := range x86RawReturnTargets {
		if target.arch != "386" {
			continue
		}
		for _, row := range []struct {
			name, source, declaration string
		}{
			{"interior-offset", strings.Replace(base, "MOVL x+0(FP)", "MOVL x+1(FP)", 1), declaration},
			{"byte-view", strings.Replace(base, "MOVL x+0(FP)", "MOVB x+0(FP)", 1), declaration},
			{"word-view", strings.Replace(base, "MOVL x+0(FP)", "MOVW x+0(FP)", 1), declaration},
			{"cross-fields", strings.Replace(base, "MOVL x+0(FP)", "MOVL x+6(FP)", 1), declaration},
			{"padding", "TEXT ·Value(SB),4,$0-20\nMOVL x+1(FP),AX\nFMOVD q+4(FP),F0\nFMOVDP F0,ret+12(FP)\nRET\n", "func Value(x byte,q float64) float64"},
			{"missing-result", "TEXT ·Value(SB),4,$0-24\nMOVL x+4(FP),AX\nRET\n", declaration},
			{"unknown-x87-result", "TEXT ·Value(SB),4,$0-24\nMOVL x+4(FP),AX\nFMOVDP F0,ret+16(FP)\nRET\n", declaration},
			{"unknown-GP-result", "TEXT ·Value(SB),4,$0-24\nMOVL x+4(FP),AX\nMOVL BX,ret+16(FP)\nMOVL $0,ret+20(FP)\nRET\n", declaration},
			{"partial-result", "TEXT ·Value(SB),4,$0-24\nMOVL x+4(FP),AX\nMOVL AX,ret+16(FP)\nRET\n", declaration},
			{"unknown-MUL-source", strings.Replace(base, "ANDL $1,AX", "MULL BX\nANDL $1,AX", 1), declaration},
			{"known-MUL-unclosed-x87", strings.Replace(base, "ANDL $1,AX", "MOVL $3,BX\nMULL BX\nANDL $1,AX", 1), declaration},
			{"native-store-under-MUL", strings.Replace(base, "ANDL $1,AX", "MOVL $3,BX\nMULL BX\nMOVL AX,0(CX)\nANDL $1,AX", 1), declaration},
		} {
			for _, api := range []string{"text", "module", "Go"} {
				t.Run(target.triple+"/"+row.name+"/"+api, func(t *testing.T) {
					ir, err := x86ScalarFPBitViewAPI(t, row.source, row.declaration, target.arch, target.triple, api)
					if !errors.Is(err, ErrProbeNeedsContext) || ir != "" {
						t.Fatalf("legacy input role supplied an independent missing proof: output=%t, error=%v", ir != "", err)
					}
				})
			}
		}
	}
}
