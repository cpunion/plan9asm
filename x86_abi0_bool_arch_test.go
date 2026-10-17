package plan9asm

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
	"testing"
)

func x86ABI0BoolArchFixture(arch string) (string, string) {
	word, lea, move := 8, "LEAQ", "MOVQ"
	if arch == "386" {
		word, lea, move = 4, "LEAL", "MOVL"
	}
	var source, declarations strings.Builder
	for index, row := range []struct {
		declaration, body string
		args              int
	}{
		{"func Value() bool", "MOVB $0,ret+0(FP)", 1},
		{"func Value() bool", "MOVB $1,ret+0(FP)", 1},
		{"func Value(x uint32) bool", fmt.Sprintf("MOVL x+0(FP),AX\nCMPL AX,$0\nSETEQ ret+%d(FP)", word), word + 1},
		{"func Value(x bool) bool", fmt.Sprintf("MOVB x+0(FP),AL\nMOVB AL,ret+%d(FP)", word), word + 1},
		{"func Value(x bool) bool", fmt.Sprintf("MOVB $2,x+0(FP)\nMOVB x+0(FP),AL\nCMPB AL,$2\nSETEQ ret+%d(FP)", word), word + 1},
		{"func Value() bool", fmt.Sprintf("%s ret+0(FP),AX\n%s AX,BX\nMOVB $1,(BX)", lea, move), 1},
		{"func Value(x bool) bool", fmt.Sprintf("%s x+0(FP),AX\nMOVB $2,(AX)\nMOVB (AX),CL\nCMPB CL,$2\nSETEQ ret+%d(FP)", lea, word), word + 1},
		{"func Value(p *uint32,old,new uint32) bool", fmt.Sprintf("%s p+0(FP),BX\nMOVL old+%d(FP),AX\nMOVL new+%d(FP),CX\nLOCK\nCMPXCHGL CX,(BX)\nSETEQ ret+%d(FP)", move, word, word+4, word+8), word + 9},
	} {
		name := fmt.Sprintf("BoolArch%d", index)
		fmt.Fprintf(&declarations, "%s\n", strings.Replace(row.declaration, "Value", name, 1))
		fmt.Fprintf(&source, "TEXT ·%s(SB),4,$0-%d\n%s\nRET\n", name, row.args, row.body)
	}
	return source.String(), declarations.String()
}

func TestX86ABI0CanonicalBoolArchitectureObjects(t *testing.T) {
	llc := findLLVM22Tool("llc")
	if llc == "" {
		t.Fatal("LLVM22 llc required")
	}
	for _, target := range x86RawReturnTargets {
		t.Run(target.triple, func(t *testing.T) {
			source, declarations := x86ABI0BoolArchFixture(target.arch)
			x86LEAActualTargetGoObject(t, target.arch, target.triple, source)
			for _, api := range []string{"text", "module", "Go"} {
				ir, err := x86ScalarFPBitViewAPI(t, source, declarations, target.arch, target.triple, api)
				if err != nil {
					t.Fatal(api, err)
				}
				for _, level := range []string{"-O0", "-O2"} {
					x86PackedStringCompileObject(t, llc, target.triple, ir, level)
				}
			}
		})
	}
}

func TestX86ABI0CanonicalBoolArchitectureNoncanonicalResults(t *testing.T) {
	for _, target := range x86RawReturnTargets {
		for _, body := range []string{
			"MOVB $2,ret+0(FP)\nRET", "MOVB $255,ret+0(FP)\nRET",
			"MOVB $2,AL\nMOVB AL,ret+0(FP)\nRET", "RET",
			"MOVL $1,AX\nCMPL AX,$0\nJNE missing\nMOVB $1,ret+0(FP)\nmissing:\nRET",
		} {
			for _, api := range []string{"text", "module", "Go"} {
				ir, err := x86ScalarFPBitViewAPI(t, "TEXT ·Value(SB),4,$0-1\n"+body+"\n",
					"func Value() bool", target.arch, target.triple, api)
				if ir != "" || !errors.Is(err, ErrProbeNeedsContext) {
					t.Errorf("noncanonical or missing bool became logical i1: %s/%s/%s: %v", target.triple, api, body, err)
				}
			}
		}
	}
}

// This portable oracle exercises the physical byte before canonical RET,
// independently of native LOCK/CMPXCHG and native Go ABI execution.
func TestX86AMD64CanonicalBoolTemporaryBytePortable(t *testing.T) {
	llc, clang, found := findLlcAndClang(t)
	if !found {
		t.Fatal("LLVM22 llc and clang required")
	}
	const source = "TEXT ·Value(SB),4,$0-9\nMOVB $2,x+0(FP)\nMOVB x+0(FP),AL\nCMPB AL,$2\nSETEQ ret+8(FP)\nRET\n"
	const driver = "#include <stdbool.h>\nextern bool Value(bool);\nint main(void){return Value(false)&&Value(true)?0:1;}\n"
	triple := testTargetTriple(runtime.GOOS, runtime.GOARCH)
	for _, api := range []string{"text", "module", "Go"} {
		t.Run(api, func(t *testing.T) {
			ir, err := x86ScalarFPBitViewAPI(t, source, "func Value(x bool) bool", "amd64", triple, api)
			if err != nil {
				t.Fatal(api, err)
			}
			for _, text := range []string{ir, x86DwordProductIROptimize(t, ir)} {
				for _, level := range []int{0, 2} {
					compileAndRunRuntimeTestWithCompiler(t, llc, []string{clang}, "bool-byte", triple, text, driver, nil, level)
				}
			}
		})
	}
}
