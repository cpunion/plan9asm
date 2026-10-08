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

func TestModuleLLVMLoaderRepeatedBuildRetainsIndependentActualScopes(t *testing.T) {
	moduleFixtureEnvironment(t)
	goBinary, err := exec.LookPath("go")
	if err != nil {
		t.Fatal(err)
	}
	out := t.TempDir()
	cfg := config{Go: goBinary, Dir: moduleLLVMConsumerFixture(t), Module: true,
		Timeout: 2 * time.Minute, Flags: []string{"-p=2"}}
	before := filepath.Join(out, "before")
	first, err := captureCIScope(context.Background(), cfg, before)
	if err != nil {
		t.Fatal("fresh first source/tool/environment/library scope", err)
	}
	firstID, err := moduleDependencyCommand(context.Background(), cfg, filepath.Join(out, "first-buildid"), goBinary,
		"tool", "buildid", before+"-llvm-loader")
	if err != nil || len(strings.TrimSpace(string(firstID))) == 0 {
		t.Fatal("actual first Go loader build identity", err)
	}
	// The second scope must build and execute its own output; it cannot execute
	// this first binary or copy the first loader's observation as a new proof.
	if err := os.Remove(before + "-llvm-loader"); err != nil {
		t.Fatal(err)
	}
	after := filepath.Join(out, "after")
	last, err := captureCIScope(context.Background(), cfg, after)
	if err != nil {
		t.Fatal("fresh second source/tool/environment/library scope", err)
	}
	lastID, err := moduleDependencyCommand(context.Background(), cfg, filepath.Join(out, "last-buildid"), goBinary,
		"tool", "buildid", after+"-llvm-loader")
	if err != nil {
		t.Fatal("actual second Go loader build identity", err)
	}
	if string(firstID) != string(lastID) {
		t.Errorf("identical request-owned probe content still changes the actual Go build identity: first=%q second=%q", firstID, lastID)
	}
	if !reflect.DeepEqual(first, last) {
		t.Fatal("independent unchanged scopes differ")
	}
	if err := verifyModuleDependencyArtifacts(out, moduleReport{Scope: first}); err != nil {
		t.Fatal("full dependency artifact readback", err)
	}
	for _, stem := range []string{before, after} {
		var origins map[string]string
		if err := ciReadJSON(stem+"-origins.json", &origins); err != nil {
			t.Fatal(err)
		}
		if origins["llvm-loader-source"] != filepath.Join(out, ".llvm-loader-source", "main.go") {
			t.Errorf("actual Go build source not bound to this request: %q", origins["llvm-loader-source"])
		}
		var loaded moduleLoadedLibrary
		if err := ciReadJSON(stem+"-llvm-loaded-library.json", &loaded); err != nil {
			t.Fatal(err)
		}
		run, err := os.ReadFile(stem + "-llvm-loader-run.stdout")
		if err != nil || string(run) != loaded.Path+"\n" || loaded.SHA != first.Tools["llvm-loaded"] {
			t.Fatal("fresh actual loader result/byte identity missing", err)
		}
	}
	t.Log("Two fresh actual Go builds and two loader executions; original 2m child budgets and complete source/tool/environment/library captures remain independent.")
}
