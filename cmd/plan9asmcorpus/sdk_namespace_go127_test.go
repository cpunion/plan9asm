//go:build go1.27

package main

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// A module can own a standard import prefix. Stock Go rejects this exact
// duplicate before compiler selection, with a positionless loader error.
// The source refusal is not an assembly object or instruction pass.
func TestOrdinarySDKNamespaceCollisionProducerSourceNA(t *testing.T) {
	candidate := discoveryCandidate{
		Module: "example.invalid/sdk-namespace", Version: "v1.0.0",
		AsmFiles: []string{"compile/internal/ssa/probe_amd64.s"},
	}
	sources := map[string]string{
		"go.mod":                             "module cmd\n\ngo 1.20\n",
		"compile/internal/ssa/probe.go":      "package ssa\nfunc Probe()\n",
		"compile/internal/ssa/probe_amd64.s": "#include \"textflag.h\"\nTEXT ·Probe(SB),NOSPLIT,$0-0\nRET\n",
	}
	download := fixtureProfileModuleDownload(t, candidate, sources)
	configurePreCPPFixtureProxy(t, candidate, download, sources["go.mod"])
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	env := replaceEnv(os.Environ(), map[string]string{
		"GOWORK": "off", "GOTOOLCHAIN": "local", "CGO_ENABLED": "0",
		"GOOS": "linux", "GOARCH": "amd64", "GOFLAGS": "", "GOEXPERIMENT": "",
	})
	stdout, stderr, failure := runDiscoveryMachineCommand(ctx, download.Dir, env,
		"go", "list", "-export", "-mod=mod", "cmd/compile/internal/ssa")
	var exit *exec.ExitError
	if !errors.As(failure, &exit) || exit.ExitCode() != 1 || len(stdout) != 0 ||
		!strings.HasPrefix(string(stderr), "ambiguous import: found package cmd/compile/internal/ssa in multiple modules:\n") ||
		!strings.Contains(string(stderr), filepath.Join(download.Dir, "compile", "internal", "ssa")) ||
		!strings.Contains(string(stderr), filepath.Join(runtime.GOROOT(), "src", "cmd", "compile", "internal", "ssa")) {
		t.Fatalf("original first-party module did not expose the stock SDK collision: error=%v stdout=%s stderr=%s", failure, stdout, stderr)
	}
	t.Logf("actual stock Go namespace collision (not a concrete file:line compiler diagnostic):\n%s", stderr)
	translator := filepath.Join(t.TempDir(), "plan9asmll")
	command := exec.CommandContext(ctx, "go", "build", "-p=1", "-o", translator, ".")
	command.Dir, command.Env = filepath.Join(root, "cmd", "plan9asmll"), os.Environ()
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("build actual translator: %v\n%s", err, output)
	}
	llc, err := exec.LookPath("llc")
	if err != nil {
		t.Fatal("required LLVM 22 llc: ", err)
	}
	matrix, _, configs, err := runDiscoveryCandidate(discoveryCorpusConfig{
		RepoRoot: root, Translator: translator, LLC: llc,
		Targets: []string{"linux/amd64"}, CandidateTimeout: 2 * time.Minute,
	}, candidate, filepath.Join(t.TempDir(), "candidate"))
	if err != nil {
		t.Fatalf("exact original module/SDK namespace collision lacks a typed source rejection: %v", err)
	}
	if matrix.Success != 0 || matrix.Failed != 0 || len(configs) != 0 || len(matrix.SourceNotApplicableItems) != 1 {
		t.Fatalf("namespace refusal manufactured objects or lost its source scope: success=%d failed=%d configs=%d items=%+v",
			matrix.Success, matrix.Failed, len(configs), matrix.SourceNotApplicableItems)
	}
}
