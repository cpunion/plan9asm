package plan9asm

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

// Native opcode outcomes, not CPUID advertisements, establish availability and
// fault behavior. This is an instruction-level C observer tied to original Go
// opcode bytes, not execution of the stock Go ABI0 caller or a native CI claim.
func TestX86CMPXCHGOriginalOpcodeOutcomeEquivalence(t *testing.T) {
	execution := x86FloatCompareExecutionTools(t)
	x86CMPXCHGRunFaultCollector(t, execution, false)
}

func x86CMPXCHGRunFaultCollector(t *testing.T, execution x86FloatCompareExecution, native bool) {
	t.Helper()
	var assembly, declarations, prototypes, originals, cases strings.Builder
	declarations.WriteString("package ordering\n")
	for index, row := range x86CMPXCHGOriginalRows {
		for _, locked := range []bool{false, true} {
			name := fmt.Sprintf("cas%d_%t", index, locked)
			body := x86CMPXCHGOriginalBody(index, "(SI)")
			if locked {
				body = "LOCK\n\t" + body
			}
			// Use the source register initialized by the reference: desired low
			// is in CX for scalar forms, and BX/CX for pair forms.
			source := strings.Replace(x86NamedLockOracleSource("amd64", body), "·locked", "·"+name, 1)
			actual := x86OrderingHintGoBytes(t, "amd64", source)
			opcode := []byte{0x0f, 0xb1, 0x0e}
			switch row.bits {
			case 8:
				if !row.pair {
					opcode = []byte{0x0f, 0xb0, 0x0e}
				}
			case 16:
				opcode = append([]byte{0x66}, opcode...)
			case 64:
				if row.pair {
					opcode = []byte{0x0f, 0xc7, 0x0e}
				} else {
					opcode = append([]byte{0x48}, opcode...)
				}
			case 128:
				opcode = []byte{0x48, 0x0f, 0xc7, 0x0e}
			}
			if locked {
				opcode = append([]byte{0xf0}, opcode...)
			}
			if !bytes.HasSuffix(actual, append(append([]byte(nil), opcode...), 0xc3)) {
				t.Fatalf("complete original opcode/prefix binding changed: %s %x", name, actual)
			}
			t.Logf("actual Go %s terminal machine bytes=%x; object identity only", name, opcode)
			assembly.WriteString(source)
			fmt.Fprintf(&declarations, "func %s(p *[32]byte)\n", name)
			fmt.Fprintf(&prototypes, "extern void %s(unsigned char *);\n", name)
			fmt.Fprintf(&cases, "    {%s, original_%s, %d, %d},\n", name, name, row.bits/8, index)
			prefix := ""
			if locked {
				prefix = "lock; "
			}
			if row.pair {
				part, mnemonic := "uint32_t", "cmpxchg8b"
				if row.bits == 128 {
					part, mnemonic = "uint64_t", "cmpxchg16b"
				}
				fmt.Fprintf(&originals, `static void original_%s(unsigned char *p) {
    %s low = 2, high = 5;
    __asm__ volatile("%s%s (%%4)"
        : "+a"(low), "+d"(high)
        : "b"((%s)3), "c"((%s)4), "r"(p)
        : "cc", "memory");
}
`, name, part, prefix, mnemonic, part, part)
			} else {
				typ := fmt.Sprintf("uint%d_t", row.bits)
				width := map[int]string{8: "b", 16: "w", 32: "l", 64: "q"}[row.bits]
				fmt.Fprintf(&originals, `static void original_%s(unsigned char *p) {
    %s accumulator = 2;
    __asm__ volatile("%scmpxchg%s %%1, (%%2)"
        : "+a"(accumulator)
        : "c"((%s)4), "r"(p)
        : "cc", "memory");
}
`, name, typ, prefix, width, typ)
			}
		}
	}
	ir, err := x86OrderingHintTranslateAPI(t, "text", assembly.String(), declarations.String(), "amd64", execution.triple)
	if err != nil {
		t.Fatal(err)
	}
	requireNative := 0
	if native {
		requireNative = 1
	}
	reference := fmt.Sprintf(x86CMPXCHGFaultReference, prototypes.String(), originals.String(), cases.String(), requireNative)
	for _, level := range []string{"-O0", "-O2"} {
		t.Run(level, func(t *testing.T) {
			output := execution.compileAndRunLLVM(t, "cmpxchg-fault", ir, reference, level)
			if strings.Count(string(output), "CMPXCHG opcode outcomes 96\n") != 1 ||
				!strings.HasSuffix(string(output), "CMPXCHG outcome failures 0\n") ||
				native && !strings.Contains(string(output), "CMPXCHG ISA deviations 0\n") {
				t.Fatalf("complete native opcode/fault/access observer: %s", output)
			}
			t.Log(string(output))
		})
	}
}

const x86CMPXCHGFaultReference = `
#include <stdint.h>
#include <stdio.h>
#include <string.h>
#include <signal.h>
#include <stdlib.h>
#include <unistd.h>
#include <sys/mman.h>
#include <sys/resource.h>
#include <sys/wait.h>

%s
%s
typedef struct {
    void (*translated)(unsigned char *);
    void (*original)(unsigned char *);
    unsigned width, family;
} Kernel;

static int observe(Kernel kernel, unsigned address, int original) {
    pid_t child = fork();
    if (child < 0) return -1;
    if (child == 0) {
        struct rlimit no_core = {0, 0};
        if (setrlimit(RLIMIT_CORE, &no_core)) _exit(120);

        _Alignas(16) unsigned char raw[64];
        memset(raw, 0xa7, sizeof(raw));
        unsigned char *p = raw + 16;
        if (address == 1) p++;
        if (address == 2) p = NULL;
        if (address == 3) {
            long page = sysconf(_SC_PAGESIZE);
            if (page <= 0) _exit(81);
            void *mapping = mmap(NULL, 2 * (size_t)page, PROT_READ | PROT_WRITE,
                MAP_PRIVATE | MAP_ANONYMOUS, -1, 0);
            if (mapping == MAP_FAILED) _exit(82);
            memset(mapping, 0xa7, (size_t)page);
            if (mprotect((unsigned char *)mapping + page, (size_t)page, PROT_NONE)) _exit(83);
            p = (unsigned char *)mapping + page - (kernel.width - 1);
        }
        unsigned char expected[16] = {0};
        uint64_t low = 2, high = 5, desired_low = 3, desired_high = 4;
        if (kernel.family == 4) {
            uint32_t pair[2] = {2, 5}, desired[2] = {3, 4};
            memcpy(expected, desired, 8);
            if (address < 2) memcpy(p, pair, 8);
        } else if (kernel.family == 5) {
            memcpy(expected, &desired_low, 8);
            memcpy(expected + 8, &desired_high, 8);
            if (address < 2) {
                memcpy(p, &low, 8);
                memcpy(p + 8, &high, 8);
            }
        } else {
            memcpy(expected, &desired_high, kernel.width);
            if (address < 2) memcpy(p, &low, kernel.width);
        }
        if (original) kernel.original(p);
        else kernel.translated(p);
        if (address >= 2) _exit(84);
        if (memcmp(p, expected, kernel.width)) _exit(85);
        size_t begin = (size_t)(p - raw);
        for (size_t i = 0; i < sizeof(raw); i++) {
            if ((i < begin || i >= begin + kernel.width) && raw[i] != 0xa7) _exit(86);
        }
        _exit(0);
    }
    int status;
    if (waitpid(child, &status, 0) != child) return -2;
    if (WIFSIGNALED(status)) return 128 + WTERMSIG(status);
    if (WIFEXITED(status)) return WEXITSTATUS(status);
    return -3;
}

int main(void) {
    Kernel kernels[] = {
%s
    };
    const unsigned require_native = %d;
    unsigned observations = 0;
    unsigned failures = 0, deviations = 0;
    for (unsigned index = 0; index < 12; index++) {
        for (unsigned address = 0; address < 4; address++) {
            int original = observe(kernels[index], address, 1);
            int translated = observe(kernels[index], address, 0);
            printf("kernel=%%u address=%%u original=%%d translated=%%d\n",
                index, address, original, translated);
            if (original != translated) failures++;
            /* Exit codes from setup, bounds, canary or wait failures never
               become an accepted outcome, even when both paths agree. */
            if (original != 0 && original != 128 + SIGSEGV &&
                    original != 128 + SIGBUS && original != 128 + SIGILL) failures++;
            if (translated != 0 && translated != 128 + SIGSEGV &&
                    translated != 128 + SIGBUS && translated != 128 + SIGILL) failures++;
            int fault = address >= 2 || (kernels[index].width == 16 && address == 1);
            if (fault) {
                if (original != 128 + SIGSEGV && original != 128 + SIGBUS) deviations++;
                /* The required Linux native axis observes #GP as SIGSEGV.
                   Rosetta's unprefixed normal return remains a deviation. */
                if (require_native && kernels[index].width == 16 && address == 1 &&
                        original != 128 + SIGSEGV) failures++;
            } else if (original != 0) {
                /* In particular an actual SIGILL is not a CPUID-based skip. */
                deviations++;
            }
            observations += 2;
        }
    }
    printf("CMPXCHG opcode outcomes %%u\n", observations);
    printf("CMPXCHG ISA deviations %%u\n", deviations);
    if (require_native && deviations) failures++;
    printf("CMPXCHG outcome failures %%u\n", failures);
    return failures ? 1 : 0;
}
`
