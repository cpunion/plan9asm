package plan9asm

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestCrossLinuxRuntimeMatrixARM64NativeScratchClosedCall(t *testing.T) {
	llc, triple, compiler, runner := arm64FPPairRuntimeTools(t)
	goos := runtime.GOOS
	if len(runner) != 0 {
		goos = "linux"
	}
	for _, operation := range []string{"add", "W27-full-overwrite"} {
		t.Run(operation, func(t *testing.T) {
			assembly := arm64ScratchSelectedSource
			goExpected, cExpected := "n+3", "n+3"
			if operation == "W27-full-overwrite" {
				assembly = strings.Replace(assembly, "ADD $3,R1", "MOVD $-1,R27\nMOVWU R1,R27\nMOVD R27,R1", 1)
				goExpected, cExpected = "uint64(uint32(n))", "(uint64_t)(uint32_t)n"
			}
			_, _, file, options, source := arm64ScratchActualBoundSourceWithAssembly(t, goos, triple, assembly)
			// The actual selected ordinary package and unchanged ASM execute via
			// their compiler-produced ABI0 wrappers. This is not a handwritten
			// native entry or a replacement for a downloaded library function.
			directory := filepath.Join(source.request.Directory, "caller")
			if err := os.Mkdir(directory, 0700); err != nil {
				t.Fatal(err)
			}
			main := fmt.Sprintf(`package main
import "example.invalid/scratchfixture"
func main() {
    canary := [3]uint64{0xabc,0xbad,0xdef}
    samples := 0
    for i := uint64(0); i < 1024; i++ {
        n := i*0x9e3779b97f4a7c15 ^ (i<<32)
        canary[1] = scratchfixture.Scratch(n)
        if canary[1] != %s || canary[0] != 0xabc || canary[2] != 0xdef {
            panic("original whole-call numeric/upper-alias/canary mismatch")
        }
        samples++
    }
    if samples != 1024 { panic("original sample denominator changed") }
}
`, goExpected)
			if err := os.WriteFile(filepath.Join(directory, "main.go"), []byte(main), 0600); err != nil {
				t.Fatal(err)
			}
			args := []string{"run", "-p=2"}
			if len(runner) != 0 {
				args = append(args, "-exec=qemu-aarch64")
			}
			args = append(args, "./caller")
			ctx, cancel := runtimeOracleStageContext(t)
			command := exec.CommandContext(ctx, source.request.GoBinary, args...)
			command.Dir, command.Env = source.request.Directory, source.request.Environment
			output, err := runtimeOracleCombinedOutput(ctx, command)
			cancel()
			if err != nil {
				t.Fatalf("unchanged actual selected Go whole-call oracle: %v\n%s", err, output)
			}
			t.Logf("actual original Go %s: 1024 mathematical/full-X27/canary observations; retained output:\n%s", operation, output)
			ir, err := Translate(file, options)
			if err != nil {
				t.Fatal(err)
			}
			mainC := fmt.Sprintf(`#include <stdint.h>
extern uint64_t Scratch(uint64_t);
int main(void) {
    uint64_t canary[3] = {0xabc,0xbad,0xdef};
    unsigned samples = 0;
    for (uint64_t i = 0; i < 1024; i++) {
        uint64_t n = i*UINT64_C(0x9e3779b97f4a7c15) ^ (i<<32);
        canary[1] = Scratch(n);
        if (canary[1] != %s || canary[0] != 0xabc || canary[2] != 0xdef) return 1;
        samples++;
    }
    return samples == 1024 ? 0 : 2;
}
`, cExpected)
			for _, level := range []string{"-O0", "-O2"} {
				t.Run(level, func(t *testing.T) {
					arm64ScratchCompileRuntimeLevel(t, llc, triple, compiler, runner, level, ir, mainC)
				})
			}
		})
	}
}

func arm64ScratchCompileRuntimeLevel(t *testing.T, llc, triple string, compiler, runner []string, level, ir, main string) {
	t.Helper()
	if level != "-O0" && level != "-O2" || len(compiler) == 0 {
		t.Fatal("missing actual compiler or unknown optimization level")
	}
	directory := t.TempDir()
	input, object := filepath.Join(directory, "scratch.ll"), filepath.Join(directory, "scratch.o")
	caller, executable := filepath.Join(directory, "caller.c"), filepath.Join(directory, "caller")
	for name, data := range map[string]string{input: ir, caller: main} {
		if err := os.WriteFile(name, []byte(data), 0600); err != nil {
			t.Fatal(err)
		}
	}
	ctx, cancel := runtimeOracleStageContext(t)
	command := exec.CommandContext(ctx, llc, level, "-mtriple="+triple, "-filetype=obj", input, "-o", object)
	output, err := runtimeOracleCombinedOutput(ctx, command)
	cancel()
	if err != nil {
		t.Fatalf("LLVM22 %s object: %v\n%s", level, err, output)
	}
	args := append([]string(nil), compiler[1:]...)
	args = append(args, object, caller, level, "-o", executable)
	ctx, cancel = runtimeOracleStageContext(t)
	command = exec.CommandContext(ctx, compiler[0], args...)
	output, err = runtimeOracleCombinedOutput(ctx, command)
	cancel()
	if err != nil {
		t.Fatalf("actual %s compiler/link: %v\n%s", level, err, output)
	}
	args = runtimeOracleGuestArguments(t, runner, executable)
	ctx, cancel = runtimeOracleStageContext(t)
	command = exec.CommandContext(ctx, args[0], args[1:]...)
	output, err = runtimeOracleCombinedOutput(ctx, command)
	cancel()
	if err != nil {
		t.Fatalf("actual %s whole-call numerical/canary run: %v\n%s", level, err, output)
	}
	t.Logf("LLVM22 %s: 1024 mathematical/full-X27/canary observations", level)
}
