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
	cmd, err := wasmPackedGoRuntimeCommand(runtime.GOROOT(), wasmPackedNode(t), filepath.Join(dir, "oracle.wasm"))
	if err != nil {
		t.Fatal(err)
	}
	// Go wasm has no native OS-thread creation; the build concurrency setting
	// must not request two runtime Ps in this single-thread host.
	cmd.Env = append(os.Environ(), "GOMAXPROCS=1")
	result, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("actual Go wasm runtime: %v\n%s", err, result)
	}
	if strings.TrimSpace(string(result)) != "actual Go packed address + complete MOV widths + dynamic calls + native division PASS" {
		t.Fatalf("runtime helper exited without executing the complete Go oracle:\n%s", result)
	}
	t.Log(strings.TrimSpace(string(result)))
}

func wasmPackedGoRuntimeCommand(root, node, wasm string) (*exec.Cmd, error) {
	for _, directory := range []string{"lib", "misc"} {
		dir := filepath.Join(root, directory, "wasm")
		complete := true
		for _, name := range []string{"wasm_exec_node.js", "wasm_exec.js"} {
			info, err := os.Stat(filepath.Join(dir, name))
			if err != nil || !info.Mode().IsRegular() {
				complete = false
				break
			}
		}
		if complete {
			// These are the arguments of Go's go_js_wasm_exec wrapper,
			// without invoking its POSIX shell or requiring execute bits.
			return exec.Command(node, "--stack-size=8192", filepath.Join(dir, "wasm_exec_node.js"), wasm), nil
		}
	}
	return nil, fmt.Errorf("selected Go toolchain lacks its required lib/misc wasm runtime JavaScript files")
}
