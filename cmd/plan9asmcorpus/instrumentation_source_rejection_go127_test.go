//go:build go1.27 && (darwin || linux) && (amd64 || arm64)

package main

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// The old gVisor release selects both version-specific constant files on the
// current SDK. A mode-only source refusal must not remove a valid sibling's
// ordinary object, or manufacture an object for the rejected package.
func TestInstrumentationSourceRejectionPreservesEligibleSiblingObject(t *testing.T) {
	compilerEnvironment := os.Environ()
	target := runtime.GOOS + "/" + runtime.GOARCH
	storeResult := func(value string) string {
		if runtime.GOARCH == "arm64" {
			return "MOVD $" + value + ",R0\nMOVD R0,ret+0(FP)\n"
		}
		return "MOVQ $" + value + ",ret+0(FP)\n"
	}
	badAssembly := "bad/probe_" + runtime.GOARCH + ".s"
	goodAssembly := "good/probe_" + runtime.GOARCH + ".s"
	candidate := discoveryCandidate{
		Module: "example.invalid/mode-source-rejection", Version: "v1.0.0",
		AsmFiles: []string{badAssembly, goodAssembly},
	}
	sources := map[string]string{
		"go.mod":                 "module example.invalid/mode-source-rejection\n\ngo 1.20\n",
		"bad/probe.go":           "//go:build race\n\npackage bad\nfunc Probe() uint64\n",
		"bad/constants_go125.go": "//go:build race && go1.25\n\npackage bad\nconst answer = 41\n",
		"bad/constants_go126.go": "//go:build race && go1.26\n\npackage bad\nconst answer = 42\n",
		badAssembly: "//go:build race\n\n#include \"textflag.h\"\n" +
			"TEXT ·Probe(SB),NOSPLIT,$0-8\n" + storeResult("41") + "RET\n",
		"good/probe.go": "package good\nfunc Probe() uint64\n",
		goodAssembly: "#include \"textflag.h\"\nTEXT ·Probe(SB),NOSPLIT,$0-8\n" +
			storeResult("7") + "RET\n",
	}
	download := fixtureProfileModuleDownload(t, candidate, sources)
	configurePreCPPFixtureProxy(t, candidate, download, sources["go.mod"])
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
	defer cancel()

	// Observe the real compiler refusal before invoking the corpus. A tool,
	// network, signal or CGO infrastructure failure is not this source witness.
	original := exec.CommandContext(ctx, filepath.Join(runtime.GOROOT(), "bin", "go"),
		"list", "-p=1", "-mod=readonly", "-race", "-export", "-json", ".")
	original.Dir = filepath.Join(download.Dir, "bad")
	original.Env = replaceEnv(compilerEnvironment, map[string]string{
		"GOOS": runtime.GOOS, "GOARCH": runtime.GOARCH, "CGO_ENABLED": "1",
		"GOENV": "off", "GOWORK": "off", "GOTOOLCHAIN": "local",
		"GOFLAGS": "", "GOEXPERIMENT": "",
	})
	stdout, err := original.Output()
	exit, plainExit := err.(*exec.ExitError)
	if !plainExit || exit.ExitCode() != 1 {
		t.Fatalf("original compiler did not plainly reject the duplicate source: error=%v stdout=%s", err, stdout)
	}
	// go list emits its selected package before the export compiler fails.
	// Keep that real selection witness instead of requiring an empty stream.
	var selected struct {
		Dir        string
		ImportPath string
		Name       string
		GoFiles    []string
		SFiles     []string
		Module     struct{ Path string }
	}
	if err := json.Unmarshal(stdout, &selected); err != nil ||
		selected.Dir != original.Dir || selected.ImportPath != candidate.Module+"/bad" ||
		selected.Name != "bad" || selected.Module.Path != candidate.Module ||
		strings.Join(selected.GoFiles, ",") != "constants_go125.go,constants_go126.go,probe.go" ||
		strings.Join(selected.SFiles, ",") != filepath.Base(badAssembly) {
		t.Fatalf("original compiler did not select the complete duplicate fixture: %v\n%s", err, stdout)
	}
	wanted := "# " + candidate.Module + "/bad\n" +
		"./constants_go126.go:4:7: answer redeclared in this block\n" +
		"\t./constants_go125.go:4:7: other declaration of answer\n"
	if string(exit.Stderr) != wanted {
		t.Fatalf("original refusal is not the complete source-bound compiler diagnostic:\n%s", exit.Stderr)
	}
	t.Logf("actual %s %s duplicate-source compiler exit1: %s", runtime.Version(), target, exit.Stderr)
	good := exec.CommandContext(ctx, filepath.Join(runtime.GOROOT(), "bin", "go"),
		"list", "-p=1", "-mod=readonly", "-export", "-compiled", "-json", ".")
	good.Dir = filepath.Join(download.Dir, "good")
	good.Env = replaceEnv(original.Env, map[string]string{"CGO_ENABLED": "0"})
	if output, err := good.Output(); err != nil {
		var diagnostic []byte
		if exit, ok := err.(*exec.ExitError); ok {
			diagnostic = exit.Stderr
		}
		t.Fatalf("original compiler did not accept the unchanged eligible sibling: %v\n%s\n%s", err, output, diagnostic)
	}

	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	translator := filepath.Join(t.TempDir(), "plan9asmll")
	build := exec.CommandContext(ctx, "go", "build", "-p=1", "-o", translator, ".")
	build.Dir, build.Env = filepath.Join(root, "cmd", "plan9asmll"), compilerEnvironment
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build actual translator: %v\n%s", err, output)
	}
	llc, err := exec.LookPath("llc")
	if err != nil {
		t.Fatal("required LLVM 22 llc: ", err)
	}
	matrix, _, _, err := runDiscoveryCandidate(discoveryCorpusConfig{
		RepoRoot: root, Translator: translator, LLC: llc,
		Targets: []string{target}, CandidateTimeout: 2 * time.Minute,
	}, candidate, filepath.Join(t.TempDir(), "candidate"))
	if err != nil {
		t.Fatalf("actual instrumented source refusal lacks its own scoped proof: %v", err)
	}
	if matrix.Success != 1 || matrix.Failed != 0 || matrix.InstrumentationTranslations != 0 ||
		matrix.OrdinarySelectionPlan == nil || matrix.OrdinarySelectionPlan.InstrumentationPlan == nil {
		t.Fatalf("rejected source removed the eligible object or manufactured mode success: success=%d failed=%d mode_objects=%d source=%+v target=%+v",
			matrix.Success, matrix.Failed, matrix.InstrumentationTranslations, matrix.SourceNotApplicableItems, matrix.NotApplicableItems)
	}
	outcomes := matrix.OrdinarySelectionPlan.InstrumentationPlan.Outcomes
	if len(outcomes) != 1 || outcomes[0].Status != "original_source_rejected" || outcomes[0].PackagePlan == nil ||
		outcomes[0].PackagePlan.PackageDirectory != "bad" || outcomes[0].Objects != nil {
		t.Fatalf("mode refusal is not independently scoped to the actual rejected package: %+v", outcomes)
	}
	encoded, err := json.Marshal(outcomes[0])
	if err != nil || !strings.Contains(string(encoded), "original_source_rejection") {
		t.Fatal("portable mode result omitted its typed compiler/source proof: ", err)
	}
}
