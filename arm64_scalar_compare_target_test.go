package plan9asm

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/xgo-dev/llvm"
)

func arm64CompareTargetAPI(t *testing.T, api, source, triple string) (bool, error) {
	t.Helper()
	file, err := Parse(ArchARM64, source)
	if err != nil {
		return false, err
	}
	options := Options{Goarch: "arm64", TargetTriple: triple,
		Sigs: map[string]FuncSig{"comparison": {Name: "comparison", Ret: Void}},
	}
	if api == "text" {
		ir, err := Translate(file, options)
		return ir != "", err
	}
	context := llvm.NewContext()
	defer context.Dispose()
	if api == "module" {
		module, err := TranslateModuleInContext(context, file, options)
		present := !module.IsNil()
		if present {
			module.Dispose()
		}
		return present, err
	}
	pkg := mustGoPackage(t, "comparison", "package comparison\nfunc comparison()\n")
	bound, err := translateGoModuleInContext(context, pkg,
		[]byte(strings.ReplaceAll(source, "TEXT comparison", "TEXT ·comparison")),
		GoModuleOptions{GOOS: "linux", GOARCH: "arm64", TargetTriple: triple})
	if bound != nil {
		bound.Module.Dispose()
	}
	return bound != nil, err
}

func TestARM64ScalarCompareRequiresNativeTarget(t *testing.T) {
	var instructions []string
	for _, row := range arm64CompareCases([]int{31}, false) {
		if !row.dead {
			instructions = append(instructions, row.instruction())
		}
	}
	for _, row := range arm64CompareCases([]int{31}, true) {
		if !row.dead {
			instructions = append(instructions, row.instruction())
			if row.bits == 32 {
				row.bits = 16
				instructions = append(instructions, row.instruction())
			}
		}
	}
	for _, api := range []string{"text", "module", "Go"} {
		for _, triple := range []string{"", "x86_64-unknown-linux-gnu", "wasm32-unknown-unknown", "armv7-unknown-linux-gnueabihf", "aarch64_be-unknown-linux-gnu", "aarch64bad-unknown-linux-gnu"} {
			for _, dead := range []bool{false, true} {
				for _, instruction := range instructions {
					t.Run(fmt.Sprintf("%s/%s/dead=%t/%s", api, triple, dead, instruction), func(t *testing.T) {
						prefix := ""
						if dead {
							prefix = "RET\n"
						}
						source := "TEXT comparison(SB),4,$0-0\n" + prefix + instruction + "\nRET\n"
						present, err := arm64CompareTargetAPI(t, api, source, triple)
						if present || !errors.Is(err, ErrProbeNeedsContext) {
							t.Fatalf("native compare target missing but output=%t err=%v", present, err)
						}
					})
				}
			}
		}
	}
}

func TestARM64ScalarCompareTargetKeepsMalformedSourcePriority(t *testing.T) {
	for _, api := range []string{"text", "module", "Go"} {
		for _, triple := range []string{"", "x86_64-unknown-linux-gnu", "aarch64-unknown-linux-gnu"} {
			for _, invalid := range []string{"FCMPS F0", "FCMPD R0,F1", "FCMPES $0,F1", "FCMPED.P F0,F1", "WORD $0x1e202001", "WORD $0x1ea02000"} {
				for _, body := range []string{invalid, "FCMPS F0,F1\n" + invalid, "FCMPS F0,F1\nRET\n" + invalid, invalid + "\nFCMPD F0,F1"} {
					source := "TEXT comparison(SB),4,$0-0\n" + body + "\nRET\n"
					present, err := arm64CompareTargetAPI(t, api, source, triple)
					if present || err == nil || errors.Is(err, ErrProbeNeedsContext) {
						t.Fatalf("%s %s malformed source changed to native context/output: %t %v\n%s", api, triple, present, err, source)
					}
				}
			}
		}
	}
}
