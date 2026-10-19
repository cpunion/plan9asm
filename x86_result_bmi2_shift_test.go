package plan9asm

import (
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/xgo-dev/plan9asm/internal/goasmrules"
)

var x86BMI2ResultNames = []string{
	"SARXL", "SARXQ", "SHLXL", "SHLXQ", "SHRXL", "SHRXQ", "RORXL", "RORXQ",
}

func TestX86BMI2ResultCompleteIndependentEncoderAttributes(t *testing.T) {
	rules, err := goasmrules.LoadX86AVXRules(runtime.GOROOT())
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range x86BMI2ResultNames {
		prefix, table, zcase, opcode := "vexF3", "_ybextrl", "Zvex_v_rm_r", "0xF7"
		classes, advance := []string{"Yrl", "Yml", "Yrl"}, 2
		if strings.HasPrefix(name, "SHLX") {
			prefix = "vex66"
		} else if strings.HasPrefix(name, "SHRX") {
			prefix = "vexF2"
		} else if strings.HasPrefix(name, "RORX") {
			prefix, table, zcase, opcode = "vexF2", "_yrorxl", "Zvex_i_rm_r", "0xF0"
			classes, advance = []string{"Yu8", "Yml", "Yrl"}, 0
		}
		mapName, wide := "vex0F38", 0
		if opcode == "0xF0" {
			mapName = "vex0F3A"
		}
		if strings.HasSuffix(name, "Q") {
			wide = 1
		}
		first := goasmrules.X86AVXRule{
			Opcode: name, Ytab: table, Prefix: "Pavx", Classes: classes,
			ZCase: zcase, ZOffset: advance,
			OpBytes: []string{fmt.Sprintf("avxEscape | vex128 | %s | %s | vexW%d", prefix, mapName, wide), opcode},
		}
		expected := []goasmrules.X86AVXRule{first}
		if advance == 0 {
			second := first
			second.Classes = []string{"Yi8", "Yml", "Yrl"}
			second.RowIndex, second.ZOffset = 1, 2
			expected = append(expected, second)
		}
		var actual []goasmrules.X86AVXRule
		for _, rule := range rules {
			if rule.Opcode == name {
				actual = append(actual, rule)
			}
		}
		if !reflect.DeepEqual(actual, expected) {
			t.Fatal("complete independent BMI2 encoding rows changed", name, actual, expected)
		}
	}
}

func x86BMI2ResultInstruction(t *testing.T, arch, op, input string, raw bool) string {
	t.Helper()
	count := "CX"
	if strings.HasPrefix(op, "RORX") {
		count = "$7"
	}
	instruction := op + " " + count + "," + input + ",AX"
	if !raw {
		return instruction
	}
	code := x86OrderingHintGoBytes(t, arch, "TEXT original(SB),4,$0-0\n"+instruction+"\nRET\n")
	return x86DwordProductDirectives(code[:len(code)-1])
}

func TestX86BMI2ResultVoidCannotBorrowInputs(t *testing.T) {
	for _, target := range x86RawReturnTargets {
		for _, op := range x86BMI2ResultNames {
			for _, raw := range []bool{false, true} {
				for _, unusedBool := range []bool{false, true} {
					t.Run(fmt.Sprintf("%s/%s/raw=%t/bool=%t", target.triple, op, raw, unusedBool), func(t *testing.T) {
						extent, declaration := 0, "func Value()"
						if unusedBool {
							extent, declaration = 1, "func Value(unused bool)"
						}
						source := fmt.Sprintf("TEXT ·Value(SB),4,$0-%d\n%s\nRET\n", extent,
							x86BMI2ResultInstruction(t, target.arch, op, "BX", raw))
						for _, api := range []string{"text", "module", "Go"} {
							ir, err := x86ScalarFPBitViewAPI(t, source, declaration, target.arch, target.triple, api)
							if ir != "" || !errors.Is(err, ErrProbeNeedsContext) {
								t.Fatal("dead BMI2 input borrowed zeroed implementation registers", api, err)
							}
						}
					})
				}
			}
		}
	}
}

func TestX86BMI2ResultInitializedFiveTargetObjects(t *testing.T) {
	llc := findLLVM22Tool("llc")
	if llc == "" {
		t.Fatal("LLVM22 llc required")
	}
	for _, target := range x86RawReturnTargets {
		t.Run(target.triple, func(t *testing.T) {
			word, move := 8, "MOVQ"
			if target.arch == "386" {
				word, move = 4, "MOVL"
			}
			var source, declarations strings.Builder
			count := 0
			for _, op := range x86BMI2ResultNames {
				for _, memory := range []bool{false, true} {
					for _, raw := range []bool{false, true} {
						name := fmt.Sprintf("Value%d", count)
						input, argument := "BX", "uintptr"
						if memory {
							input, argument = "(BX)", "*[8]byte"
						}
						fmt.Fprintf(&declarations, "func %s(value %s,count uintptr) uintptr\n", name, argument)
						fmt.Fprintf(&source, "TEXT ·%s(SB),4,$0-%d\n%s value+0(FP),BX\n%s count+%d(FP),CX\n%s\n%s AX,ret+%d(FP)\nRET\n",
							name, 3*word, move, move, word, x86BMI2ResultInstruction(t, target.arch, op, input, raw), move, 2*word)
						count++
					}
				}
			}
			if count != 32 {
				t.Fatal("complete register/memory and named/raw denominator changed", count)
			}
			x86LEAActualTargetGoObject(t, target.arch, target.triple, source.String())
			for _, api := range []string{"text", "module", "Go"} {
				ir, err := x86ScalarFPBitViewAPI(t, source.String(), declarations.String(), target.arch, target.triple, api)
				if err != nil {
					t.Fatal("complete BMI2 source lost its numerical/native continuation", api, err)
				}
				for _, text := range []string{ir, x86DwordProductIROptimize(t, ir)} {
					for _, level := range []string{"-O0", "-O2"} {
						x86PackedStringCompileObject(t, llc, target.triple, text, level)
					}
				}
			}
		})
	}
}
