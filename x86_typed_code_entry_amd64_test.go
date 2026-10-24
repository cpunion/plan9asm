package plan9asm

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// This original Go program calls the physical ABI0 address. Independently,
// LLVM/C calls the translated typed entry. Equal executed answers do not mean
// those binary calling conventions are interchangeable.
func TestX86TypedCodeEntryExecutedGoAndLLVM(t *testing.T) {
	dir := t.TempDir()
	goOracle := "package entry\nimport \"testing\"\n" + x86TypedCodeCallbackDeclarations + `
func TestOriginal(t *testing.T) {
    values := []int64{-44, -2, 0, 7, 7, 31}
    var tls uint64 = 0x1122334455667788
    for repeat := 0; repeat < 20; repeat++ {
        for i := range values {
            for j := range values {
                want := int32(0)
                if values[i] < values[j] {
                    want = -1
                }
                if values[i] > values[j] {
                    want = 1
                }
                got := Forward(&tls, uintptr(values[i]), uintptr(values[j]), PC())
                if got != want {
                    t.Fatalf("compare %d/%d: got %d, want %d", i, j, got, want)
                }
            }
        }
    }
    if tls != 0x1122334455667788 {
        t.Fatal("TLS canary changed")
    }
}
`
	for name, source := range map[string]string{
		"go.mod":        "module entry\n\ngo 1.20\n",
		"entry_amd64.s": x86TypedCodeCallbackAssembly,
		"entry_test.go": goOracle,
	} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(source), 0600); err != nil {
			t.Fatal(err)
		}
	}
	ctx, cancel := runtimeOracleStageContext(t)
	command := exec.CommandContext(ctx, "go", "test", "-count=1", ".")
	command.Dir = dir
	command.Env = append(os.Environ(), "GOWORK=off", "GOTOOLCHAIN=local", "CGO_ENABLED=0")
	output, err := runtimeOracleCombinedOutput(ctx, command)
	cancel()
	if err != nil {
		t.Fatalf("original Go ABI0 callback: %v\n%s", err, output)
	}

	llc, clang, found := findLlcAndClang(t)
	if !found {
		t.Fatal("LLVM22 llc/clang required for typed code-entry execution")
	}
	ctx, cancel = runtimeOracleStageContext(t)
	compiler := runtimeOracleNativeCCommand(t, ctx, clang).Args
	cancel()
	pkg := mustGoPackage(t, "entry", "package entry\n"+x86TypedCodeCallbackDeclarations)
	triple := testTargetTriple(runtime.GOOS, "amd64")
	result, err := TranslateGoModule(pkg, []byte(x86TypedCodeCallbackAssembly), GoModuleOptions{
		GOARCH: "amd64", TargetTriple: triple,
		ResolveSym: func(symbol string) string { return strings.TrimPrefix(symbol, "·") },
	})
	if err != nil {
		t.Fatal(err)
	}
	ir := result.Module.String()
	result.Module.Dispose()
	const driver = `#include <stdint.h>
extern uintptr_t PC(void);
extern int32_t Target(uint64_t*, uintptr_t, uintptr_t);
extern int32_t Forward(uint64_t*, uintptr_t, uintptr_t, uintptr_t);
int main(void) {
    int64_t values[] = {-44, -2, 0, 7, 7, 31};
    uint64_t tls = UINT64_C(0x1122334455667788);
    uintptr_t pc = PC();
    if (pc != (uintptr_t)&Target) return 1; /* actual code, never a funcval box */
    for (int repeat = 0; repeat < 20; repeat++) {
        for (unsigned i = 0; i < 6; i++) for (unsigned j = 0; j < 6; j++) {
            int32_t want = (values[i] > values[j]) - (values[i] < values[j]);
            uintptr_t left = (uintptr_t)values[i], right = (uintptr_t)values[j];
            if (Forward(&tls, left, right, pc) != want) return 2;
            if (((int32_t (*)(uint64_t*, uintptr_t, uintptr_t))pc)(&tls, left, right) != want) return 3;
        }
    }
    return tls != UINT64_C(0x1122334455667788);
}
`
	for _, level := range []int{0, 2} {
		compileAndRunRuntimeTestWithCompiler(t, llc, compiler, "typed-code-entry", triple, ir, driver, nil, level)
	}
}
