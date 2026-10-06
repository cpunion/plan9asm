//go:build unix

package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestModuleScopeBindsActualWASMAndLLVMConsumers(t *testing.T) {
	moduleFixtureEnvironment(t)
	goBinary, err := exec.LookPath("go")
	if err != nil {
		t.Fatal(err)
	}
	dir := moduleFixture(t, false)
	stem := filepath.Join(t.TempDir(), "scope")
	cfg := config{Go: goBinary, Dir: dir, Module: true, Timeout: 2 * time.Minute}
	scope, err := captureCIScope(context.Background(), cfg, stem)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"node", "wasm-ld", "llvm-shared/0", "sdk-wasm/wasm_exec.js", "sdk-wasm/wasm_exec_node.js"} {
		if scope.Tools[name] == "" {
			t.Errorf("complete module scope omitted actual required consumer %s", name)
		}
	}
	var origins map[string]string
	if err := ciReadJSON(stem+"-origins.json", &origins); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"wasm_exec.js", "wasm_exec_node.js"} {
		origin := origins["sdk-wasm/"+name]
		if !strings.HasPrefix(origin, runtime.GOROOT()+string(filepath.Separator)) {
			t.Errorf("wasm runtime came from another SDK: %s", origin)
		}
	}
}

func TestModuleActualCLIRejectsMissingNodeConsumer(t *testing.T) {
	moduleFixtureEnvironment(t)
	binary := parallelFixtureCLI(t)
	dir := moduleFixture(t, false)
	goBinary, err := exec.LookPath("go")
	if err != nil {
		t.Fatal(err)
	}
	llvmConfig, err := exec.LookPath(os.Getenv("LLVM_CONFIG"))
	if err != nil {
		t.Fatal(err)
	}
	git, err := exec.LookPath("git")
	if err != nil {
		t.Fatal(err)
	}
	gitDirectory := t.TempDir()
	if err := os.Symlink(git, filepath.Join(gitDirectory, "git")); err != nil {
		t.Fatal(err)
	}
	// Keep the real SDK and LLVM tools, but no Node interpreter. This is an
	// actual Go module execution, not a script impersonating a compiler.
	t.Setenv("PATH", strings.Join([]string{filepath.Dir(goBinary), filepath.Dir(llvmConfig), gitDirectory}, string(os.PathListSeparator)))
	if _, err := exec.LookPath("node"); err == nil {
		t.Fatal("missing-Node fixture still has an interpreter")
	}
	stem := filepath.Join(t.TempDir(), "cli")
	out := filepath.Join(t.TempDir(), "module")
	_, err = captureCommand(context.Background(), 5*time.Minute, dir, binary,
		[]string{"-module", "-dir", dir, "-out", out, "-shards=2", "-workers=2", "-timeout=2m", "--", "-p=2"}, stem)
	if err == nil {
		t.Fatal("module completed without its required Node runtime")
	}
	data, readErr := os.ReadFile(stem + ".stderr")
	if readErr != nil || !strings.Contains(string(data), "node") {
		t.Fatalf("missing runtime did not retain its concrete cause: %v\n%s", readErr, data)
	}
}
