//go:build go1.27

package main

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestOrdinaryCPPFormalIncludeOperandsDefaultTargetsAndReaders(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Minute)
	defer cancel()
	repo, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	translator := filepath.Join(t.TempDir(), "plan9asmll")
	if stdout, stderr, err := runDiscoveryMachineCommand(ctx, filepath.Join(repo, "cmd/plan9asmll"),
		os.Environ(), "go", "build", "-p=2", "-o", translator, "."); err != nil {
		t.Fatalf("actual translator build: %v\n%s\n%s", err, stdout, stderr)
	} else if len(stderr) != 0 {
		t.Logf("actual translator build notices:\n%s", stderr)
	}
	llc, err := exec.LookPath("llc")
	if err != nil {
		t.Fatal(err)
	}
	candidate := discoveryCandidate{
		Module: "example.invalid/ordinary-formal-includes", Version: "v1.0.0", AsmFiles: []string{"probe.s"},
	}
	sources := map[string]string{
		"go.mod":   "module example.invalid/ordinary-formal-includes\n\ngo 1.20\n",
		"probe.go": "package fixture\nfunc Probe()\n",
		"cpu.s":    "// Original nonempty selected sibling, not a discovered file.\n",
		"probe.s": "#include \"macros.h\"\n" +
			"INCLUDE(\"formal.h\")\nAPPLY(\"nested.h\")\nALIAS(\"alias.h\")\n" +
			"INCLUDE(\"../../src/runtime/go_tls.h\")\n" +
			"TEXT ·Probe(SB),4,$0-0\nRET\n" +
			"DATA ·unused(SB)/8,$\"nofile.h\"\nGLOBL ·unused(SB),$8\n",
		"macros.h": "#define INCLUDE(file) \\\n#include file \\\n\n" +
			"#define APPLY(file) INCLUDE(file)\n#define ALIAS INCLUDE\n",
		"formal.h": "GLOBL ·formalMarker(SB),$8\n",
		"nested.h": "GLOBL ·nestedMarker(SB),$8\n",
		"alias.h":  "GLOBL ·aliasMarker(SB),$8\n",
	}
	download := fixtureProfileModuleDownload(t, candidate, sources)
	configurePreCPPFixtureProxy(t, candidate, download, sources["go.mod"])
	targets, err := parseDiscoveryTargets("darwin/amd64,darwin/arm64,linux/386,linux/amd64,linux/arm,linux/arm64,windows/386,windows/amd64,windows/arm64,js/wasm,wasip1/wasm")
	if err != nil || len(targets) != 11 {
		t.Fatalf("original full target denominator: %v %v", targets, err)
	}
	for _, target := range targets {
		goos, goarch, _ := strings.Cut(target, "/")
		env := replaceEnv(os.Environ(), map[string]string{"GOOS": goos, "GOARCH": goarch, "CGO_ENABLED": "0"})
		stdout, stderr, err := runDiscoveryMachineCommand(ctx, download.Dir, env,
			"go", "list", "-p=2", "-mod=readonly", "-export", "-json", ".")
		if err != nil || len(stderr) != 0 {
			t.Fatalf("actual original %s export: %v\n%s\n%s", target, err, stdout, stderr)
		}
		var selected struct {
			GoFiles, SFiles []string
			Export          string
		}
		if err := json.Unmarshal(stdout, &selected); err != nil ||
			!equalDiscoveryStrings(selected.GoFiles, []string{"probe.go"}) ||
			!equalDiscoveryStrings(selected.SFiles, []string{"cpu.s", "probe.s"}) || selected.Export == "" {
			t.Fatalf("actual original selection %s: %+v %v", target, selected, err)
		}
		digest, err := discoveryFeatureFileSHA256(selected.Export)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("FORMAL_ORDINARY_ACTUAL_GO %s export_sha=%s source_sha=%s", target,
			digest, discoveryFeatureBytesSHA256([]byte(sources["probe.s"])))
	}
	matrix, _, configs, err := runDiscoveryCandidate(discoveryCorpusConfig{
		RepoRoot: repo, Translator: translator, LLC: llc, Targets: targets, CandidateTimeout: 5 * time.Minute,
	}, candidate, filepath.Join(t.TempDir(), "candidate"))
	if err != nil {
		t.Fatalf("actual original formal includes failed ordinary producer: %v", err)
	}
	if matrix.Success < len(targets) || matrix.NotApplicable != 0 || len(matrix.FeatureConsumption) != len(configs) {
		t.Fatalf("ordinary scopes missing: objects=%d N/A=%d consumers=%d configs=%d",
			matrix.Success, matrix.NotApplicable, len(matrix.FeatureConsumption), len(configs))
	}
	inventory := newDiscoveryFeatureInventory()
	refs, err := registerDiscoveryFeatureProfiles(inventory, matrix.FeatureProfiles)
	if err != nil {
		t.Fatal(err)
	}
	result := discoveryCorpusResult{
		Module: candidate.Module, Version: candidate.Version, Status: discoveryStatusPassed,
		DiscoveredAsmFiles: candidate.AsmFiles, ApplicableAsmFiles: discoveryConfigurationAsmFiles(configs),
		BuildConfigurations: configs, Translations: matrix.Success, OrdinarySelectionPlan: matrix.OrdinarySelectionPlan,
		FeatureProfiles: refs, FeatureConsumption: matrix.FeatureConsumption, featureInventory: inventory,
	}
	if err := validateOrdinaryProfileResult(result, targets, matrix.OrdinarySelectionPlan.GoVersion); err != nil {
		t.Fatal(err)
	}
	checkSelectedSiblingReportReaders(t, candidate, result, targets, inventory)
	encoded, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	originalBytes := 0
	for _, source := range matrix.OrdinarySelectionPlan.CPPInputs.Sources {
		originalBytes += len(*source.Original)
	}
	t.Logf("FORMAL_ORDINARY_CLOSED objects=%d configs=%d source_bytes=%d result_bytes=%d",
		matrix.Success, len(configs), originalBytes, len(encoded))
}
