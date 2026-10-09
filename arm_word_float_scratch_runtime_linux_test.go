package plan9asm

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestCrossLinuxRuntimeMatrixARMWordFloatScratch(t *testing.T) {
	llc := requireARMScalarRuntime(t)
	const source = `TEXT oracle(SB),4,$0-4
MOVF $7.0,F15
MOVD $1.5,F0
MOVDW F0,R1
MOVW F15,R0
MOVW R0,result+0(FP)
RET
`
	directory := t.TempDir()
	for name, data := range map[string]string{
		"go.mod": "module armmovconvert\n\ngo 1.20\n",
		"main.go": `package main
import "fmt"
func oracle() uint32
func main() {
    result := oracle()
    if result != 1 { panic(fmt.Sprintf("original F15 scratch = %x, want 1", result)) }
    fmt.Println("ORIGINAL_ARM_CONVERSION_SCRATCH 1")
}
`,
		"original_arm.s": strings.ReplaceAll(source, "TEXT oracle(", "TEXT ·oracle("),
	} {
		if err := os.WriteFile(filepath.Join(directory, name), []byte(data), 0600); err != nil {
			t.Fatal(err)
		}
	}
	executable := filepath.Join(directory, "original")
	ctx, cancel := runtimeOracleStageContext(t)
	command := exec.CommandContext(ctx, "go", "build", "-p=1", "-o", executable, ".")
	command.Dir = directory
	command.Env = append(os.Environ(), "GOOS=linux", "GOARCH=arm", "GOARM=7", "CGO_ENABLED=0", "GOTOOLCHAIN=local")
	output, err := runtimeOracleCombinedOutput(ctx, command)
	cancel()
	if err != nil {
		t.Fatalf("actual Go conversion build: %v\n%s", err, output)
	}
	args := runtimeOracleGuestArguments(t, []string{"qemu-arm", "-L", "/usr/arm-linux-gnueabihf"}, executable)
	ctx, cancel = runtimeOracleStageContext(t)
	output, err = runtimeOracleCombinedOutput(ctx, exec.CommandContext(ctx, args[0], args[1:]...))
	cancel()
	if err != nil || string(output) != "ORIGINAL_ARM_CONVERSION_SCRATCH 1\n" {
		t.Fatalf("actual unchanged Go F15 effect: %v\n%s", err, output)
	}
	t.Log(strings.TrimSpace(string(output)))

	file, err := Parse(ArchARM, source)
	if err != nil {
		t.Fatal(err)
	}
	ir, err := Translate(file, Options{
		Goarch: "arm", TargetTriple: "armv7-unknown-linux-gnueabihf",
		Sigs: map[string]FuncSig{"oracle": {
			Name: "oracle", Ret: I32,
			Frame: FrameLayout{Results: []FrameSlot{{Type: I32, Field: -1}}},
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	const driver = `#include <stdint.h>
#include <stdio.h>
extern uint32_t oracle(void);
int main(void) {
    uint32_t value = oracle();
    if (value != 1) {
        fprintf(stderr,"LLVM F15 scratch = %x, want 1\n",value);
        return 1;
    }
    return 0;
}
`
	for _, level := range []int{0, 2} {
		compileAndRunRuntimeTestWithCompiler(t, llc, []string{"arm-linux-gnueabihf-gcc", "-no-pie"},
			"arm_conversion_scratch", "armv7-unknown-linux-gnueabihf", ir, driver,
			[]string{"qemu-arm", "-L", "/usr/arm-linux-gnueabihf"}, level)
	}
}
