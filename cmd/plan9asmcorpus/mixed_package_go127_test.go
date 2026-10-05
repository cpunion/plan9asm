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

func TestOrdinaryMixedPackagesProducerKeepsEligibleObjects(t *testing.T) {
	compilerEnv := os.Environ()
	candidate := discoveryCandidate{
		Module: "example.invalid/mixed-packages", Version: "v1.0.0",
		AsmFiles: []string{"bad/probe_amd64.s", "good/probe_amd64.s"},
	}
	sources := map[string]string{
		"go.mod":             "module example.invalid/mixed-packages\n\ngo 1.20\n",
		"bad/probe.go":       "package bad\nfunc Probe()\n",
		"bad/probe_test.go":  "package bad_x\n",
		"bad/probe_amd64.s":  "#include \"textflag.h\"\nTEXT ·Probe(SB),NOSPLIT,$0-0\nRET\n",
		"good/probe.go":      "package good\nfunc Probe() uint64\n",
		"good/probe_test.go": "package good_test\n",
		"good/probe_amd64.s": "#include \"textflag.h\"\nTEXT ·Probe(SB),NOSPLIT,$0-8\nMOVQ $7,ret+0(FP)\nRET\n",
	}
	download := fixtureProfileModuleDownload(t, candidate, sources)
	configurePreCPPFixtureProxy(t, candidate, download, sources["go.mod"])
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	translator := filepath.Join(t.TempDir(), "plan9asmll")
	command := exec.CommandContext(ctx, "go", "build", "-p=1", "-o", translator, ".")
	command.Dir, command.Env = filepath.Join(root, "cmd", "plan9asmll"), compilerEnv
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("build actual translator: %v\n%s", err, output)
	}
	llc, err := exec.LookPath("llc")
	if err != nil {
		t.Fatal("required LLVM 22 llc: ", err)
	}
	targets := []string{"darwin/amd64", "linux/amd64", "windows/amd64"}
	matrix, _, configs, err := runDiscoveryCandidate(discoveryCorpusConfig{
		RepoRoot: root, Translator: translator, LLC: llc,
		Targets: targets, CandidateTimeout: 2 * time.Minute,
	}, candidate, filepath.Join(t.TempDir(), "candidate"))
	if err != nil {
		t.Fatalf("original Go package conflict lacks a bound source rejection: %v", err)
	}
	if matrix.Success != len(targets) || matrix.Failed != 0 || len(configs) != len(targets) ||
		len(matrix.SourceNotApplicableItems) != len(targets) {
		t.Fatalf("source refusal erased valid sibling objects or manufactured success: success=%d failed=%d configs=%d refusals=%d",
			matrix.Success, matrix.Failed, len(configs), len(matrix.SourceNotApplicableItems))
	}
	for _, item := range matrix.SourceNotApplicableItems {
		if item.Kind != "go_mixed_package_clauses_rejected_package_target" ||
			!equalDiscoveryStrings(item.AsmFiles, []string{"bad/probe_amd64.s"}) ||
			!strings.Contains(item.Reason, "found packages bad (probe.go) and bad_x (probe_test.go)") {
			t.Fatalf("unbound or foreign package refusal: %+v", item)
		}
	}
	for _, config := range configs {
		if !equalDiscoveryStrings(config.AsmFiles, []string{"good/probe_amd64.s"}) {
			t.Fatal("mixed-package refusal crossed into the valid external-test package")
		}
	}
	inventory := newDiscoveryFeatureInventory()
	refs, err := registerDiscoveryFeatureProfiles(inventory, matrix.FeatureProfiles)
	if err != nil {
		t.Fatal(err)
	}
	result := discoveryCorpusResult{
		Module: candidate.Module, Version: candidate.Version, Status: discoveryStatusPassed,
		DiscoveredAsmFiles: candidate.AsmFiles, ApplicableAsmFiles: discoveryConfigurationAsmFiles(configs),
		BuildConfigurations: configs, Translations: matrix.Success,
		OrdinarySelectionPlan: matrix.OrdinarySelectionPlan, FeatureProfiles: refs,
		FeatureConsumption:       matrix.FeatureConsumption,
		SourceNotApplicableItems: matrix.SourceNotApplicableItems, featureInventory: inventory,
	}
	if err := validateOrdinaryProfileResult(result, targets, matrix.OrdinarySelectionPlan.GoVersion); err != nil {
		t.Fatalf("original source conflict and actual LLVM object proof do not replay together: %v", err)
	}
	encoded, err := json.Marshal(matrix.OrdinarySelectionPlan)
	if err != nil || !strings.Contains(string(encoded), "mixed_package_rejections") {
		t.Fatal("portable report omitted the independently replayable package-clause proof: ", err)
	}
}
