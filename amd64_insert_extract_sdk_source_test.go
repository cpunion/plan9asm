package plan9asm

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestX86InsertExtractActualSDKTablesAreMandatory(t *testing.T) {
	const rootName = "TestX86InsertExtractEveryGoOpcodeAttribute"
	for _, name := range []string{"avx_optabs.go", "asm6.go"} {
		t.Run(name, func(t *testing.T) {
			// Execute the actual attribute root from this SDK's test binary.
			// A partial fixture GOROOT supplies no replacement encoder evidence.
			fixture := t.TempDir()
			directory := filepath.Join(fixture, "src", "cmd", "internal", "obj", "x86")
			if err := os.MkdirAll(directory, 0700); err != nil {
				t.Fatal(err)
			}
			if name == "asm6.go" {
				original, err := os.ReadFile(filepath.Join(runtime.GOROOT(), "src", "cmd", "internal", "obj", "x86", "avx_optabs.go"))
				if err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(directory, "avx_optabs.go"), original, 0600); err != nil {
					t.Fatal(err)
				}
			}
			missing := filepath.Join(directory, name)
			if _, err := os.ReadFile(missing); !os.IsNotExist(err) {
				t.Fatalf("missing-table fixture is not absent: %v", err)
			}

			stage, cancelStage := runtimeOracleStageContext(t)
			defer cancelStage()
			ctx, cancel := context.WithTimeout(stage, 30*time.Second)
			defer cancel()
			command := exec.CommandContext(ctx, os.Args[0], "-test.run=^"+rootName+"$", "-test.count=1", "-test.timeout=20s")
			for _, item := range os.Environ() {
				if !strings.HasPrefix(item, "GOROOT=") {
					command.Env = append(command.Env, item)
				}
			}
			command.Env = append(command.Env, "GOROOT="+fixture)
			output, err := runtimeOracleCombinedOutput(ctx, command)
			exit, plainExit := err.(*exec.ExitError)
			if ctx.Err() != nil || !plainExit || exit.ProcessState == nil ||
				!exit.ProcessState.Exited() || exit.ExitCode() != 1 {
				t.Fatalf("actual SDK-table root did not fail with plain exit 1: %v (context %v)\n%s", err, ctx.Err(), output)
			}

			// Bind the complete failure to this one root, its real source file,
			// and the exact missing table. Panics, cleanup errors, extra failures,
			// unrelated diagnostics and a PASS from version-return are not evidence.
			_, missingErr := os.ReadFile(missing)
			if !os.IsNotExist(missingErr) {
				t.Fatalf("missing-table fixture changed during the child: %v", missingErr)
			}
			pattern := fmt.Sprintf(`\A--- FAIL: %s \([0-9]+\.[0-9]+s\)\n    amd64_insert_extract_forms_test\.go:[0-9]+: %s\nFAIL\n\z`,
				rootName, regexp.QuoteMeta(missingErr.Error()))
			if !regexp.MustCompile(pattern).Match(output) {
				t.Fatalf("unbound actual SDK-table failure (want %s):\n%s", missing, output)
			}
			t.Logf("actual %s child: plain exit 1, exact %s diagnostic\n%s", runtime.Version(), name, output)
		})
	}
}
