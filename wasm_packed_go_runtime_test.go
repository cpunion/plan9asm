package plan9asm

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// This oracle uses the selected Go compiler and that same toolchain's runtime
// helper. Missing helper files are hard failures, not version-based skips.
func TestWASMPackedFunctionAddressActualGoRuntime(t *testing.T) {
	source, _, goSource := wasmPackedRuntimeFixture(t)
	dir := t.TempDir()
	for name, data := range map[string]string{"go.mod": "module packed_pc_oracle\ngo 1.20\n", "main.go": goSource, "oracle_wasm.s": source} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(data), 0600); err != nil {
			t.Fatal(err)
		}
	}
	cmd := exec.Command("go", "build", "-o", "oracle.wasm", ".")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GOWORK=off", "GOOS=js", "GOARCH=wasm", "GOTOOLCHAIN=local")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("actual Go wasm oracle build: %v\n%s", err, out)
	}
	wasmPackedNode(t)
	helper := ""
	for _, directory := range []string{"lib", "misc"} {
		candidate := filepath.Join(runtime.GOROOT(), directory, "wasm", "go_js_wasm_exec")
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() && info.Mode()&0111 != 0 {
			helper = candidate
			break
		}
	}
	if helper == "" {
		t.Fatal("selected Go toolchain lacks its required lib/misc wasm runtime helper")
	}
	cmd = exec.Command(helper, filepath.Join(dir, "oracle.wasm"))
	// Go wasm has no native OS-thread creation; the build concurrency setting
	// must not request two runtime Ps in this single-thread host.
	cmd.Env = append(os.Environ(), "GOMAXPROCS=1")
	result, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("actual Go wasm runtime: %v\n%s", err, result)
	}
	t.Log(strings.TrimSpace(string(result)))
}
