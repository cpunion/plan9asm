//go:build unix

package plan9asm

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

func runtimeOracleFaultC(t *testing.T, name string) string {
	t.Helper()
	file, err := parser.ParseFile(token.NewFileSet(), name, nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	var found []string
	ast.Inspect(file, func(node ast.Node) bool {
		literal, ok := node.(*ast.BasicLit)
		if !ok || literal.Kind != token.STRING {
			return true
		}
		value, err := strconv.Unquote(literal.Value)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(value, "fork()") {
			found = append(found, value)
		}
		return true
	})
	if len(found) != 1 {
		t.Fatalf("%s: expected one actual fault-child C source, found %d", name, len(found))
	}
	return found[0]
}

func TestRuntimeOracleFaultChildrenDisableCoreDumps(t *testing.T) {
	child := regexp.MustCompile(`if\s*\(\s*(?:child\s*==\s*0|!child)\s*\)\s*\{`)
	policy := regexp.MustCompile(`^\s*struct\s+rlimit\s+no_core\s*=\s*\{\s*0\s*,\s*0\s*\}\s*;\s*if\s*\(\s*setrlimit\s*\(\s*RLIMIT_CORE\s*,\s*&no_core\s*\)\s*\)\s*_exit\s*\(\s*120\s*\)\s*;`)
	for _, name := range []string{
		"amd64_signed_integer_scalar_float_runtime_test.go",
		"amd64_packed_shift_fault_test.go",
		"amd64_packed_shift_native_count_runtime_test.go",
		"amd64_insert_extract_numerical_test.go",
	} {
		t.Run(name, func(t *testing.T) {
			source := runtimeOracleFaultC(t, name)
			branch := child.FindStringIndex(source)
			if branch == nil || !policy.MatchString(source[branch[1]:]) {
				t.Fatal("actual fault child must first set both core limits to zero and check failure before invoking its oracle")
			}
		})
	}
}

// Execute the actual shift fault helper with a harmless callback. Inheriting
// a nonzero core limit fails without raising a signal or creating a core file.
func TestRuntimeOracleFaultChildCoreLimitInheritance(t *testing.T) {
	_, clang, ok := findLlcAndClang(t)
	if !ok {
		t.Fatal("LLVM 22 tools required")
	}
	source := runtimeOracleFaultC(t, "amd64_packed_shift_fault_test.go")
	source += `
#include <sys/resource.h>

static void inspect_limits(uint8_t *out, const uint8_t *data, const uint8_t *count) {
    struct rlimit limit;
    if (getrlimit(RLIMIT_CORE, &limit) || limit.rlim_cur || limit.rlim_max) _exit(91);
    memset(out, 0, 32);
}

int main(int argc, char **argv) {
    struct rlimit original, parent, after;
    if (getrlimit(RLIMIT_CORE, &original)) return 10;
    parent = original;
    if (argc > 1) {
        parent.rlim_cur = parent.rlim_max = 0;
    } else {
        parent.rlim_cur = parent.rlim_max < 4096 ? parent.rlim_max : 4096;
    }
    if (setrlimit(RLIMIT_CORE, &parent)) return 11;
    long page = sysconf(_SC_PAGESIZE);
    if (page <= 0) return 12;
    int result = check_fault(inspect_limits, 0, 0, (size_t)page);
    if (getrlimit(RLIMIT_CORE, &after) || after.rlim_cur != parent.rlim_cur ||
        after.rlim_max != parent.rlim_max) return 13;
    if (argc == 1 && setrlimit(RLIMIT_CORE, &original)) return 14;
    fprintf(stdout, "parent=%llu/%llu child-status=%d\n",
        (unsigned long long)parent.rlim_cur, (unsigned long long)parent.rlim_max, result);
    return result;
}
`
	directory := t.TempDir()
	cPath, executable := filepath.Join(directory, "core.c"), filepath.Join(directory, "core")
	if err := os.WriteFile(cPath, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := runtimeOracleStageContext(t)
	output, err := runtimeOracleCombinedOutput(ctx, exec.CommandContext(ctx, clang, "-O2", cPath, "-o", executable))
	cancel()
	if err != nil {
		t.Fatalf("compile actual fault helper: %v\n%s", err, output)
	}
	for _, mode := range []string{"inherited", "hard-zero"} {
		t.Run(mode, func(t *testing.T) {
			var arguments []string
			if mode == "hard-zero" {
				arguments = append(arguments, mode)
			}
			ctx, cancel := runtimeOracleStageContext(t)
			output, err := runtimeOracleCombinedOutput(ctx, exec.CommandContext(ctx, executable, arguments...))
			cancel()
			if err != nil {
				t.Fatalf("actual child resource observation: %v\n%s", err, output)
			}
			t.Log(strings.TrimSpace(string(output)))
		})
	}
}
