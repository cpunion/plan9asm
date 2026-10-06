//go:build unix

package plan9asm

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"testing"
)

func x86UndefinedTrapRequireOriginalSignal(t *testing.T, output []byte, err error) {
	t.Helper()
	var exit *exec.ExitError
	if !errors.As(err, &exit) || exit.ExitCode() != 2 ||
		!strings.Contains(string(output), "SIGILL: illegal instruction") ||
		!strings.Contains(string(output), "main.fault") {
		t.Fatalf("unchanged Go object did not raise the actual undefined-instruction fault: %v\n%s", err, output)
	}
}

func TestCrossLinuxRuntimeMatrixX86UndefinedTrapOriginalAndLLVM(t *testing.T) {
	execution := x86FloatCompareExecutionTools(t)
	for _, raw := range []bool{false, true} {
		t.Run(fmt.Sprintf("raw=%t", raw), func(t *testing.T) {
			goos := runtime.GOOS
			if strings.Contains(execution.triple, "linux") {
				goos = "linux"
			}
			original := x86UndefinedTrapBuildOriginal(t, goos, "amd64", raw)
			arguments := execution.command(original)
			ctx, cancel := runtimeOracleStageContext(t)
			command := exec.CommandContext(ctx, arguments[0], arguments[1:]...)
			// The Go runtime reports SIGILL and exits 2. It must not escalate
			// the oracle into a core-producing crash because of inherited env.
			command.Env = append(os.Environ(), "GOTRACEBACK=system")
			output, err := runtimeOracleCombinedOutput(ctx, command)
			cancel()
			x86UndefinedTrapRequireOriginalSignal(t, output, err)
			ir := x86UndefinedTrapIR(t, x86UndefinedTrapFixture(raw), "amd64", execution.triple, "Go")
			for _, level := range []string{"-O0", "-O2"} {
				t.Run(level, func(t *testing.T) {
					optimized := ir
					if level == "-O2" {
						optimized = x86UndefinedTrapOptimizedIR(t, ir)
					}
					output := execution.compileAndRunLLVM(t, "undefined", optimized, x86UndefinedTrapIndependentC, level)
					if string(output) != "undefined-instruction SIGILL verified\n" {
						t.Fatalf("independent real-signal oracle: %q", output)
					}
				})
			}
		})
	}
}

// A child returning normally, an int3/SIGTRAP, and a generic abort all fail.
// The parent observes the kernel's signal, independently of translator IR.
const x86UndefinedTrapIndependentC = `
#include <signal.h>
#include <stdio.h>
#include <sys/resource.h>
#include <sys/wait.h>
#include <unistd.h>

extern void fault(void);

int main(void) {
    pid_t child = fork();
    if (child < 0) return 1;
    if (child == 0) {
        struct rlimit no_core = {0, 0};
        if (setrlimit(RLIMIT_CORE, &no_core)) _exit(120);
        fault();
        _exit(91);
    }
    int status = 0;
    if (waitpid(child, &status, 0) != child) return 2;
    if (!WIFSIGNALED(status) || WTERMSIG(status) != SIGILL) {
        fprintf(stderr, "expected SIGILL, got child status=%x\n", status);
        return 3;
    }
    puts("undefined-instruction SIGILL verified");
    return 0;
}
`
