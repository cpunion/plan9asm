package plan9asm

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestCrossLinuxRuntimeMatrixX86TimestampOriginalAndLLVM(t *testing.T) {
	execution := x86FloatCompareExecutionTools(t)
	// First execute the unchanged Go objects. A missing/unsupported native ISA
	// fails this axis; it is not permission to accept a constant LLVM result.
	x86TimestampRunOriginalGo(t, execution)
	var declarations, assembly strings.Builder
	declarations.WriteString("package timestamp\n")
	for _, op := range []Op{"RDTSC", "RDTSCP"} {
		declaration, source := x86TimestampOriginalFixture(op)
		declarations.WriteString(declaration)
		assembly.WriteString(source)
	}
	goAssemblerOracleOutput(t, "amd64", assembly.String(), "", true)
	ir := x86NarrowOracleIR(t, assembly.String(), declarations.String(), execution.triple)
	for _, level := range []string{"-O0", "-O2"} {
		t.Run(level, func(t *testing.T) {
			output := execution.compileAndRunLLVM(t, "timestamp", ir, x86TimestampIndependentC, level)
			if string(output) != "timestamp observations 128\n" {
				t.Fatalf("independent timestamp/register/flags/canary oracle: %q", output)
			}
		})
	}
}

func x86TimestampRunOriginalGo(t *testing.T, execution x86FloatCompareExecution) {
	t.Helper()
	var declarations, assembly strings.Builder
	declarations.WriteString("package main\n")
	for _, op := range []Op{"RDTSC", "RDTSCP"} {
		declaration, source := x86TimestampOriginalFixture(op)
		declarations.WriteString(declaration)
		assembly.WriteString(source)
	}
	const driver = `package main

import (
    "fmt"
    "runtime"
)

func main() {
    runtime.LockOSThread()
    observations := 0
    for index, sample := range []func(*[8]uint64){rdtsc, rdtscp} {
        for iteration := 0; iteration < 64; iteration++ {
            out := [8]uint64{0, 0, 0, 0, 0, 0, 0xabcdef0123456789, 0x13579bdf2468ace0}
            sample(&out)
            stamp := out[0] | out[1] << 32
            if stamp == 0 || out[0] >> 32 != 0 || out[1] >> 32 != 0 ||
                    out[3] != 0x7e6d5c4b3a291807 || out[4] & 0x8c5 != 0x8c5 ||
                    out[6] != 0xabcdef0123456789 || out[7] != 0x13579bdf2468ace0 {
                panic(fmt.Sprintf("original timestamp %d: %x", index, out))
            }
            if index == 0 && out[2] != 0xa1b2c3d4e5f60718 || index == 1 && out[2] >> 32 != 0 {
                panic(fmt.Sprintf("original timestamp CX %d: %x", index, out))
            }
            observations++
        }
    }
    fmt.Printf("original timestamp observations %d\n", observations)
}
`
	dir := t.TempDir()
	for name, text := range map[string]string{
		"go.mod": "module timestamp\n\ngo 1.20\n", "main.go": driver,
		"decl.go": declarations.String(), "timestamp_amd64.s": assembly.String(),
	} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
	}
	executable := filepath.Join(dir, "original-go")
	command := exec.Command("go", "build", "-p=1", "-o", executable, ".")
	command.Dir = dir
	command.Env = append(os.Environ(), "CGO_ENABLED=0", "GOARCH=amd64", "GOTOOLCHAIN=local")
	x86FloatCompareCaptureCommand(t, command)
	if output := execution.run(t, executable); string(output) != "original timestamp observations 128\n" {
		t.Fatalf("actual original Go objects did not establish timestamp semantics: %q", output)
	}
	t.Log("128 original Go timestamp/register/flags/canary observations")
}

const x86TimestampIndependentC = `
#include <stdint.h>
#include <stdio.h>

extern void rdtsc(uint64_t *);
extern void rdtscp(uint64_t *);

static uint64_t read_counter(void) {
    unsigned low, high;
    __asm__ volatile("lfence; rdtsc; lfence" : "=a"(low), "=d"(high) :: "memory");
    return (uint64_t)low | (uint64_t)high << 32;
}

int main(void) {
    void (*samples[])(uint64_t *) = { rdtsc, rdtscp };
    unsigned observations = 0;
    for (unsigned index = 0; index < 2; index++) {
        for (unsigned iteration = 0; iteration < 64; iteration++) {
            unsigned before_csr, after_csr;
            uint64_t out[8] = { 0, 0, 0, 0, 0, 0,
                UINT64_C(0xabcdef0123456789), UINT64_C(0x13579bdf2468ace0) };
            __asm__ volatile("stmxcsr %0" : "=m"(before_csr));
            uint64_t before = read_counter();
            samples[index](out);
            uint64_t after = read_counter();
            __asm__ volatile("stmxcsr %0" : "=m"(after_csr));
            uint64_t got = out[0] | out[1] << 32;
            if (!got || got < before || got > after || out[0] >> 32 || out[1] >> 32 ||
                    out[3] != UINT64_C(0x7e6d5c4b3a291807) || (out[4] & 0x8c5) != 0x8c5 ||
                    out[5] != before_csr || after_csr != before_csr ||
                    out[6] != UINT64_C(0xabcdef0123456789) ||
                    out[7] != UINT64_C(0x13579bdf2468ace0)) {
                fprintf(stderr, "timestamp %u got=%llx window=%llx..%llx flags=%llx csr=%llx/%x/%x\n",
                    index, (unsigned long long)got, (unsigned long long)before, (unsigned long long)after,
                    (unsigned long long)out[4], (unsigned long long)out[5], before_csr, after_csr);
                return 1;
            }
            if ((!index && out[2] != UINT64_C(0xa1b2c3d4e5f60718)) || (index && out[2] >> 32)) {
                fprintf(stderr, "timestamp CX %u=%llx\n", index, (unsigned long long)out[2]);
                return 2;
            }
            observations++;
        }
    }
    printf("timestamp observations %u\n", observations);
    return 0;
}
`
