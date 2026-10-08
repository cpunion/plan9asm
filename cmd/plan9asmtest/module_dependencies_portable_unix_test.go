//go:build unix

package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestModuleActualDependencyArtifactsSurviveProducerOriginRemoval(t *testing.T) {
	moduleFixtureEnvironment(t)
	goBinary, err := exec.LookPath("go")
	if err != nil {
		t.Fatal(err)
	}
	configBinary, err := exec.LookPath(os.Getenv("LLVM_CONFIG"))
	if err != nil {
		t.Fatal(err)
	}
	// llvm-config reports its physical relocatable prefix. Keep the fixture's
	// own installation in that same namespace, including Darwin's /var alias.
	producerRoot, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	producer := filepath.Join(producerRoot, "producer")
	for _, directory := range []string{"bin", "lib"} {
		if err := os.MkdirAll(filepath.Join(producer, directory), 0700); err != nil {
			t.Fatal(err)
		}
	}
	// Copy only the real llvm-config bytes. Its actual relocatable-prefix
	// query determines the library origins; no script impersonates a tool.
	localConfig := filepath.Join(producer, "bin", "llvm-config")
	if err := copyModuleFixtureTool(configBinary, localConfig); err != nil {
		t.Fatal(err)
	}
	cfg := config{Go: goBinary, Module: true, Timeout: 2 * time.Minute}
	cfg.Dir = moduleLLVMConsumerFixture(t)
	out := t.TempDir()
	data, err := moduleDependencyCommand(context.Background(), cfg, filepath.Join(out, "original-libraries"), configBinary, "--libfiles", "--link-shared")
	if err != nil {
		t.Fatal(err)
	}
	libraries, err := moduleSharedLibraries(data)
	if err != nil {
		t.Fatal(err)
	}
	for _, library := range libraries {
		if err := copyModuleFixtureTool(library, filepath.Join(producer, "lib", filepath.Base(library))); err != nil {
			t.Fatal(err)
		}
	}
	bindir, err := moduleDependencyCommand(context.Background(), cfg, filepath.Join(out, "original-bindir"), configBinary, "--bindir")
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"llc", "clang", "llvm-as", "opt", "llvm-dis", "llvm-objdump", "llvm-nm"} {
		if err := os.Symlink(filepath.Join(strings.TrimSpace(string(bindir)), name), filepath.Join(producer, "bin", name)); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("LLVM_CONFIG", localConfig)
	before, err := captureCIScope(context.Background(), cfg, filepath.Join(out, "before"))
	if err != nil {
		t.Fatal(err)
	}
	after, err := captureCIScope(context.Background(), cfg, filepath.Join(out, "after"))
	if err != nil || !reflect.DeepEqual(before, after) {
		t.Fatal("unchanged actual relocated producer scope", err)
	}
	report := moduleReport{Scope: before}
	if err := verifyModuleDependencyArtifacts(out, report); err != nil {
		t.Fatal("actual producer control", err)
	}
	var origins map[string]string
	if err := ciReadJSON(filepath.Join(out, "before-origins.json"), &origins); err != nil {
		t.Fatal(err)
	}
	t.Logf("actual relocated llvm-config library origin=%q requested producer=%q",
		origins["llvm-shared/0"], producer)
	if !strings.HasPrefix(origins["llvm-shared/0"], producer+string(filepath.Separator)) {
		t.Fatal("real llvm-config did not select the owned producer installation")
	}
	if err := os.Rename(producer, producer+"-closed"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(origins["llvm-shared/0"]); !os.IsNotExist(err) {
		t.Fatal("producer origin is unexpectedly still available", err)
	}
	relocated := filepath.Join(t.TempDir(), "downloaded-owner")
	if err := os.Rename(out, relocated); err != nil {
		t.Fatal(err)
	}
	if err := verifyModuleDependencyArtifacts(relocated, report); err != nil {
		t.Fatal("portable reader required a retired producer physical path", err)
	}
	path := filepath.Join(relocated, "before-llvm-loaded-library.json")
	var loaded moduleLoadedLibrary
	if err := ciReadJSON(path, &loaded); err != nil {
		t.Fatal(err)
	}
	for _, mutation := range []string{"missing-byte-identity", "changed-byte-identity", "changed-configured-index", "changed-loaded-path"} {
		t.Run(mutation, func(t *testing.T) {
			changed := loaded
			switch mutation {
			case "missing-byte-identity":
				changed.SHA = ""
			case "changed-byte-identity":
				changed.SHA = strings.Repeat("0", 64)
			case "changed-configured-index":
				changed.Index = len(libraries)
			case "changed-loaded-path":
				changed.Path = goBinary
			}
			if err := ciWriteJSON(path, changed); err != nil {
				t.Fatal(err)
			}
			if err := verifyModuleDependencyArtifacts(relocated, report); err == nil {
				t.Fatal("changed/missing loaded library identity accepted")
			}
			if err := ciWriteJSON(path, loaded); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func moduleLLVMConsumerFixture(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	for name, source := range map[string]string{
		"go.mod":      "module example.org/llvmconsumer\n\ngo 1.20\nrequire github.com/xgo-dev/llvm v0.9.9\n",
		"consumer.go": "package consumer\nimport \"github.com/xgo-dev/llvm\"\nfunc Context() llvm.Context { return llvm.NewContext() }\n",
	} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(source), 0600); err != nil {
			t.Fatal(err)
		}
	}
	sum, err := os.ReadFile(filepath.Join("..", "..", "go.sum"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "go.sum"), sum, 0600); err != nil {
		t.Fatal(err)
	}
	return dir
}
