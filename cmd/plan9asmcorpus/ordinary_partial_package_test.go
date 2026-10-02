package main

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/xgo-dev/plan9asm/internal/gotoolprofile"
)

func TestOrdinaryProducerPartialAsmDeclKeepsCompletePackageSelection(t *testing.T) {
	if runtimeGoMinorForProfileTest(t) < 27 {
		t.Skip("actual external-corpus producer requires Go 1.27; portable protocol guards run on Go 1.20")
	}
	compilerEnvironment := os.Environ()
	candidate := discoveryCandidate{
		Module: "example.invalid/partial-package", Version: "v1.0.0",
		AsmFiles: []string{"bad_amd64.s", "good_amd64.s"},
	}
	sources := map[string]string{
		"go.mod":       "module example.invalid/partial-package\n\ngo 1.20\n",
		"probe.go":     "package partial\nfunc Bad(x uint64) uint64\nfunc Good() uint64\n",
		"bad_amd64.s":  "TEXT ·Bad(SB),4,$0-8\nMOVQ x+0(FP),AX\nMOVQ AX,ret+8(FP)\nRET\n",
		"good_amd64.s": "TEXT ·Good(SB),4,$0-8\nMOVQ $7,ret+0(FP)\nRET\n",
	}
	download := fixtureProfileModuleDownload(t, candidate, sources)
	proxy := t.TempDir()
	versions := filepath.Join(proxy, candidate.Module, "@v")
	if err := os.MkdirAll(versions, 0755); err != nil {
		t.Fatal(err)
	}
	archive, err := os.ReadFile(download.Zip)
	if err != nil {
		t.Fatal(err)
	}
	for name, data := range map[string][]byte{
		"v1.0.0.mod":  []byte(sources["go.mod"]),
		"v1.0.0.info": []byte(`{"Version":"v1.0.0","Time":"2026-10-02T00:00:00Z"}`),
		"v1.0.0.zip":  archive,
	} {
		if err := os.WriteFile(filepath.Join(versions, name), data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("GOPROXY", "file://"+proxy)
	t.Setenv("GOSUMDB", "off")
	t.Setenv("GOMODCACHE", filepath.Join(t.TempDir(), "modules"))
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "plan9asmll")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	command := exec.CommandContext(ctx, "go", "build", "-p=1", "-o", binary, ".")
	command.Dir, command.Env = filepath.Join(root, "cmd", "plan9asmll"), compilerEnvironment
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("build our translator: %v\n%s", err, output)
	}
	llc, err := exec.LookPath("llc")
	if err != nil {
		t.Fatal(err)
	}
	matrix, _, configs, err := runDiscoveryCandidate(discoveryCorpusConfig{
		RepoRoot: root, Translator: binary, LLC: llc, Targets: []string{"linux/amd64"},
		CandidateTimeout: time.Minute,
	}, candidate, filepath.Join(t.TempDir(), "candidate"))
	if err != nil {
		t.Fatal(err)
	}
	if matrix.Success != 1 || len(matrix.SourceNotApplicableItems) != 1 || len(configs) != 1 ||
		!equalDiscoveryStrings(configs[0].AsmFiles, []string{"good_amd64.s"}) || len(matrix.FeatureConsumption) != 1 {
		t.Fatalf("partial asmdecl exclusion changed the translated denominator: %#v", matrix)
	}
	proof := matrix.FeatureConsumption[0]
	if len(proof.Packages) != 1 || !equalDiscoveryStrings(proof.Packages[0].SFiles, candidate.AsmFiles) {
		t.Fatal("actual Go-selected assembly was filtered to conceal the rejected sibling")
	}
	inventory := &discoveryFeatureInventory{}
	refs, err := registerDiscoveryFeatureProfiles(inventory, matrix.FeatureProfiles)
	if err != nil {
		t.Fatal(err)
	}
	result := discoveryCorpusResult{
		Module: candidate.Module, Version: candidate.Version, Status: discoveryStatusPassed,
		DiscoveredAsmFiles: candidate.AsmFiles, ApplicableAsmFiles: discoveryConfigurationAsmFiles(configs),
		BuildConfigurations: configs, Translations: matrix.Success,
		SourceNotApplicableItems: matrix.SourceNotApplicableItems,
		OrdinarySelectionPlan:    matrix.OrdinarySelectionPlan, FeatureProfiles: refs,
		FeatureConsumption: matrix.FeatureConsumption,
	}
	encoded, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	replay := func(data []byte) discoveryCorpusResult {
		t.Helper()
		var copied discoveryCorpusResult
		if err := json.Unmarshal(data, &copied); err != nil {
			t.Fatal(err)
		}
		copied.featureInventory = inventory
		return copied
	}
	validate := func(result discoveryCorpusResult) error {
		return validateOrdinaryProfileResult(result, []string{"linux/amd64"}, matrix.OrdinarySelectionPlan.GoVersion)
	}
	if err := validate(replay(encoded)); err != nil {
		t.Fatalf("actual partial-package proof JSON replay: %v", err)
	}
	for name, change := range map[string]func(*discoveryCorpusResult){
		"missing complete registration": func(r *discoveryCorpusResult) { r.FeatureConsumption[0].RegisteredAsmFiles = nil },
		"unregistered nonempty sibling": func(r *discoveryCorpusResult) { r.FeatureConsumption[0].RegisteredAsmFiles = []string{"good_amd64.s"} },
		"registration alone":            func(r *discoveryCorpusResult) { r.FeatureConsumption = nil },
		"no source diagnostic":          func(r *discoveryCorpusResult) { r.SourceNotApplicableItems = nil },
		"wrong diagnostic tags":         func(r *discoveryCorpusResult) { r.SourceNotApplicableItems[0].BuildTags = []string{"alternate"} },
		"wrong diagnostic profile":      func(r *discoveryCorpusResult) { r.SourceNotApplicableItems[0].ProfileID = "wrong" },
		"wrong package role":            func(r *discoveryCorpusResult) { r.FeatureConsumption[0].Packages[0].SourceRole = "owned_local_replace" },
		"wrong selected hash": func(r *discoveryCorpusResult) {
			r.FeatureConsumption[0].Packages[0].SourceSHA256["bad_amd64.s"] = discoveryFeatureBytesSHA256([]byte("changed"))
		},
		"concealed selected sibling": func(r *discoveryCorpusResult) {
			r.FeatureConsumption[0].Packages[0].SFiles = []string{"good_amd64.s"}
			delete(r.FeatureConsumption[0].Packages[0].SourceSHA256, "bad_amd64.s")
		},
	} {
		t.Run(name, func(t *testing.T) {
			changed := replay(encoded)
			change(&changed)
			if err := validate(changed); err == nil {
				t.Fatal("incomplete/changed package registration or scope discharge accepted")
			}
		})
	}
	// Removing an independently selected nonempty file from the explicit
	// translation input still fails without complete package registration.
	profile := matrix.FeatureProfiles[0]
	input := ordinaryProfileConsumerInput(matrix.OrdinarySelectionPlan, profile, candidate.Module, "", []string{"good_amd64.s"})
	input.AsmFiles = nil
	if err := gotoolprofile.ValidateSelection(input, proof, nil, true); err == nil {
		t.Fatal("source registration was mistaken for an executed translation")
	}
}

func TestOrdinaryPartialPackageRegistrationPortableGuards(t *testing.T) {
	candidate := discoveryCandidate{
		Module: "example.com/partial", Version: "v1.0.0",
		AsmFiles: []string{"bad_amd64.s", "good_amd64.s"},
	}
	plan := fixtureOrdinarySelectionForCandidate(t, candidate, []string{"linux/amd64"}, map[string]string{
		"go.mod":       "module example.com/partial\n",
		"probe.go":     "package partial\nfunc Bad(x uint64) uint64\nfunc Good() uint64\n",
		"bad_amd64.s":  "TEXT ·Bad(SB),4,$0-8\nRET\n",
		"good_amd64.s": "TEXT ·Good(SB),4,$0-8\nMOVQ $7,ret+0(FP)\nRET\n",
	})
	result := discoveryCorpusResult{
		Module: candidate.Module, Version: candidate.Version, Status: discoveryStatusPassed,
		DiscoveredAsmFiles: candidate.AsmFiles, ApplicableAsmFiles: []string{"good_amd64.s"},
		BuildConfigurations: []discoveryBuildConfiguration{{Targets: plan.Targets, AsmFiles: []string{"good_amd64.s"}}},
		Translations:        1, OrdinarySelectionPlan: plan,
		SourceNotApplicableItems: []discoverySourceNotApplicableItem{{
			AsmFiles: []string{"bad_amd64.s"}, Targets: plan.Targets,
			Kind:   discoverySourceNotApplicableAsmDecl,
			Reason: "bad_amd64.s:1: [amd64] Bad: wrong argument size 8; expected $...-16",
		}},
	}
	inventory := &discoveryFeatureInventory{}
	fixtureProfileEvidence(t, &result, inventory)
	validate := func(r discoveryCorpusResult) error {
		return validateOrdinaryProfileResult(r, plan.Targets, plan.GoVersion)
	}
	if err := validate(result); err != nil {
		t.Fatal(err)
	}
	checkPartialPackageReportReaders(t, result, inventory)
	encoded, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	for name, change := range map[string]func(*discoveryCorpusResult){
		"missing registration": func(r *discoveryCorpusResult) { r.FeatureConsumption[0].RegisteredAsmFiles = nil },
		"duplicate registration": func(r *discoveryCorpusResult) {
			r.FeatureConsumption[0].RegisteredAsmFiles = []string{"bad_amd64.s", "bad_amd64.s", "good_amd64.s"}
		},
		"unselected registration": func(r *discoveryCorpusResult) {
			r.FeatureConsumption[0].RegisteredAsmFiles = []string{"bad_amd64.s", "good_amd64.s", "ignored_arm64.s"}
		},
		"missing diagnostic":            func(r *discoveryCorpusResult) { r.SourceNotApplicableItems = nil },
		"registration without output":   func(r *discoveryCorpusResult) { r.FeatureConsumption[0].Outputs = nil },
		"registration without consumer": func(r *discoveryCorpusResult) { r.FeatureConsumption = nil },
		"wrong consumption tags":        func(r *discoveryCorpusResult) { r.FeatureConsumption[0].CustomTags = []string{"alternate"} },
		"wrong target":                  func(r *discoveryCorpusResult) { r.BuildConfigurations[0].Targets = []string{"darwin/amd64"} },
	} {
		t.Run(name, func(t *testing.T) {
			var changed discoveryCorpusResult
			if err := json.Unmarshal(encoded, &changed); err != nil {
				t.Fatal(err)
			}
			changed.featureInventory = inventory
			change(&changed)
			if err := validate(changed); err == nil {
				t.Fatal("partial package proof was weakened by relabeling its registration")
			}
		})
	}
	profiles, err := resolveDiscoveryFeatureProfiles(inventory, result.FeatureProfiles)
	if err != nil {
		t.Fatal(err)
	}
	input := ordinaryProfileConsumerInput(plan, profiles[0], candidate.Module, "", []string{"good_amd64.s"})
	for name, change := range map[string]func(*gotoolprofile.ConsumerInput){
		"missing registration":  func(in *gotoolprofile.ConsumerInput) { in.RegisteredAsmFiles = nil },
		"missing sibling":       func(in *gotoolprofile.ConsumerInput) { in.RegisteredAsmFiles = []string{"good_amd64.s"} },
		"unsorted registration": func(in *gotoolprofile.ConsumerInput) { in.RegisteredAsmFiles = []string{"good_amd64.s", "bad_amd64.s"} },
		"duplicate registration": func(in *gotoolprofile.ConsumerInput) {
			in.RegisteredAsmFiles = []string{"bad_amd64.s", "bad_amd64.s", "good_amd64.s"}
		},
		"foreign translation": func(in *gotoolprofile.ConsumerInput) { in.AsmFiles = []string{"foreign_amd64.s"} },
	} {
		t.Run("actual selection/"+name, func(t *testing.T) {
			copied := *input
			change(&copied)
			if err := gotoolprofile.ValidatePackageSelection(&copied, result.FeatureConsumption[0].Packages, nil); err == nil {
				t.Fatal("actual nonempty package selection escaped the complete registration guard")
			}
		})
	}
}

func checkPartialPackageReportReaders(t *testing.T, result discoveryCorpusResult, inventory *discoveryFeatureInventory) {
	t.Helper()
	ledger := filepath.Join(t.TempDir(), "ledger")
	if err := os.MkdirAll(filepath.Join(ledger, "records"), 0755); err != nil {
		t.Fatal(err)
	}
	record, err := json.Marshal(discoveryRecord{
		Kind: "matched", Module: result.Module, Version: result.Version,
		Architectures: []string{"amd64"}, AsmFiles: result.DiscoveredAsmFiles,
	})
	if err != nil {
		t.Fatal(err)
	}
	writeTestFile(t, filepath.Join(ledger, "records", "00.jsonl"), string(record)+"\n")
	provenance := fixtureDiscoveryProvenance(t, ledger)
	report := discoveryCorpusReport{
		SchemaVersion: discoveryReportSchema, Provenance: provenance,
		Targets: result.OrdinarySelectionPlan.Targets, ShardCount: 1,
		CandidateTotal: 1, EligibleCandidates: 1, Selected: 1, Passed: 1,
		Translations: result.Translations, Results: []discoveryCorpusResult{result}, FeatureInventory: inventory,
	}
	reports := t.TempDir()
	filename := filepath.Join(reports, "shard-0.json")
	if err := writeDiscoveryCorpusReport(filename, report); err != nil {
		t.Fatal(err)
	}
	if err := verifyDiscoveryCorpusReports(ledger, reports, report.Targets, provenance.Source); err != nil {
		t.Fatalf("complete partial-package aggregate: %v", err)
	}
	progress, err := collectDiscoveryProgress(ledger, reports, report.Targets, provenance.Source, 1)
	if err != nil {
		t.Fatal(err)
	}
	output := filepath.Join(t.TempDir(), "assembly-ledger")
	semanticSource := discoveryFeatureBytesSHA256([]byte("synthetic semantic source"))
	if err := writeAssemblyLedger(output, progress, semanticSource); err != nil {
		t.Fatal(err)
	}
	replayed, err := readAssemblyLedger(output, progress.LedgerSHA256, semanticSource)
	if err != nil || !replayed.Verified || len(replayed.Candidates) != 1 ||
		!equalDiscoveryStrings(replayed.Candidates[0].FeatureConsumption[0].RegisteredAsmFiles, result.DiscoveredAsmFiles) {
		t.Fatalf("compact ledger lost the separate complete package registration: %v", err)
	}
	encoded, err := json.Marshal(report)
	if err != nil {
		t.Fatal(err)
	}
	progressBytes, err := json.Marshal(progress)
	if err != nil {
		t.Fatal(err)
	}
	for _, missing := range []string{"registration", "source diagnostic", "object"} {
		t.Run("strict readers/"+missing, func(t *testing.T) {
			var changed discoveryCorpusReport
			if err := json.Unmarshal(encoded, &changed); err != nil {
				t.Fatal(err)
			}
			var compact discoveryProgress
			if err := json.Unmarshal(progressBytes, &compact); err != nil {
				t.Fatal(err)
			}
			switch missing {
			case "registration":
				changed.Results[0].FeatureConsumption[0].RegisteredAsmFiles = nil
				compact.Candidates[0].FeatureConsumption[0].RegisteredAsmFiles = nil
			case "source diagnostic":
				changed.Results[0].SourceNotApplicableItems = nil
				compact.Candidates[0].SourceNotApplicableItems = nil
			case "object":
				changed.Results[0].FeatureConsumption[0].Outputs = nil
				compact.Candidates[0].FeatureConsumption[0].Outputs = nil
			}
			if err := writeDiscoveryCorpusReport(filename, changed); err != nil {
				t.Fatal(err)
			}
			if err := verifyDiscoveryCorpusReports(ledger, reports, report.Targets, provenance.Source); err == nil {
				t.Fatal("aggregate promoted an incomplete partial-package proof")
			}
			if _, err := collectDiscoveryProgress(ledger, reports, report.Targets, provenance.Source, 1); err == nil {
				t.Fatal("progress promoted an incomplete partial-package proof")
			}
			if err := validateAssemblyLedgerProgress(compact); err == nil {
				t.Fatal("compact ledger promoted an incomplete partial-package proof")
			}
		})
	}
}
