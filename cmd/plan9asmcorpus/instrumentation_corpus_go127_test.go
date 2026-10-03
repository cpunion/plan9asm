//go:build go1.27

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

	"github.com/xgo-dev/plan9asm/internal/gotoolprofile"
)

func TestInstrumentationCorpusOriginalModeAssemblyReachesStrictObjects(t *testing.T) {
	compilerEnvironment := os.Environ()
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	translator := filepath.Join(t.TempDir(), "plan9asmll")
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
	defer cancel()
	build := exec.CommandContext(ctx, "go", "build", "-p=1", "-o", translator, ".")
	build.Dir = filepath.Join(root, "cmd", "plan9asmll")
	build.Env = compilerEnvironment
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build current translator: %v\n%s", err, output)
	}
	for _, test := range []struct {
		name, guard, cgo, target string
		objects                  bool
	}{
		{"default CGO1", "race", "1", "darwin/arm64", true},
		{"explicit CGO0", "race && !cgo", "0", "darwin/arm64", true},
		{"early target rejection", "race", "1", "windows/arm64", false},
		{"supported target missing CGO", "race && !cgo", "0", "linux/arm64", false},
	} {
		t.Run(test.name, func(t *testing.T) {
			testInstrumentationCorpusOriginalMode(t, root, translator, test.guard, test.cgo, test.target, test.objects)
		})
	}
}

func testInstrumentationCorpusOriginalMode(t *testing.T, root, translator, guard, cgo, target string, wantObjects bool) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
	defer cancel()
	candidate := discoveryCandidate{
		Module: "example.invalid/instrumented", Version: "v1.0.0", AsmFiles: []string{"probe_arm64.s"},
	}
	sources := map[string]string{
		"go.mod":        "module example.invalid/instrumented\n\ngo 1.20\n",
		"probe.go":      "//go:build " + guard + "\n\npackage instrumented\nconst answer = 41\nfunc Probe(x uint64) uint64\n",
		"probe_arm64.s": "//go:build " + guard + "\n\n#include \"go_asm.h\"\nTEXT ·Probe(SB),$0-16\nMOVD x+0(FP), R0\nADD $const_answer, R0\nMOVD R0, ret+8(FP)\nRET\n",
	}
	download := fixtureProfileModuleDownload(t, candidate, sources)
	configurePreCPPFixtureProxy(t, candidate, download, sources["go.mod"])
	original := exec.CommandContext(ctx, "go", "list", "-mod=readonly", "-p=1", "-race", "-export", "-json", ".")
	original.Dir = download.Dir
	goos, goarch, _ := strings.Cut(target, "/")
	original.Env = replaceEnv(os.Environ(), map[string]string{"GOOS": goos, "GOARCH": goarch, "CGO_ENABLED": cgo})
	output, err := original.Output()
	if !wantObjects {
		exited, ok := err.(*exec.ExitError)
		wanted := "-race is not supported on windows/arm64"
		if target == "linux/arm64" {
			wanted = "go: -race requires cgo"
			if runtime.GOOS == goos && runtime.GOARCH == goarch {
				wanted += "; enable cgo by setting CGO_ENABLED=1"
			}
		}
		if !ok || exited.ExitCode() != 2 || strings.TrimSpace(string(exited.Stderr)) != wanted {
			var diagnostic []byte
			if ok {
				diagnostic = exited.Stderr
			}
			t.Fatalf("original stock Go did not produce the exact early target rule: %v\nstdout=%s\nstderr=%s", err, output, diagnostic)
		}
	} else if err != nil {
		if exited, ok := err.(*exec.ExitError); ok {
			t.Fatalf("original stock Go mode export: %v\n%s", err, exited.Stderr)
		}
		t.Fatal(err)
	}
	var selected struct {
		GoFiles []string
		SFiles  []string
		Export  string
	}
	if wantObjects {
		if err := json.Unmarshal(output, &selected); err != nil ||
			!equalDiscoveryStrings(selected.GoFiles, []string{"probe.go"}) ||
			!equalDiscoveryStrings(selected.SFiles, []string{"probe_arm64.s"}) || selected.Export == "" {
			t.Fatalf("original mode did not select/export the exact Go/ASM pair: %+v, %v", selected, err)
		}
		if info, err := os.Stat(selected.Export); err != nil || info.Size() == 0 {
			t.Fatalf("original mode export lacks its physical Go object: %v", err)
		}
	}
	llc, err := exec.LookPath("llc")
	if err != nil {
		t.Fatal(err)
	}
	workDir := filepath.Join(t.TempDir(), "candidate")
	matrix, _, configs, err := runDiscoveryCandidate(discoveryCorpusConfig{
		RepoRoot: root, Translator: translator, LLC: llc, Targets: []string{target},
		CandidateTimeout: 3 * time.Minute,
	}, candidate, workDir)
	if !wantObjects {
		if err == nil || matrix.InstrumentationTranslations != 0 || matrix.Success != 0 || matrix.OrdinarySelectionPlan == nil {
			t.Fatalf("early target rejection became completion or lost the complete source denominator: matrix=%+v, error=%v", matrix, err)
		}
		inventory := newDiscoveryFeatureInventory()
		refs, refErr := registerDiscoveryFeatureProfiles(inventory, matrix.FeatureProfiles)
		if refErr != nil {
			t.Fatal(refErr)
		}
		result := discoveryCorpusResult{
			Module: candidate.Module, Version: candidate.Version, Status: discoveryStatusFailed, Error: err.Error(),
			DiscoveredAsmFiles: candidate.AsmFiles, OrdinarySelectionPlan: matrix.OrdinarySelectionPlan, FeatureProfiles: refs,
		}
		if err := validateDiscoveryInstrumentationFailure(result, inventory); err != nil {
			t.Fatalf("independent early rejection readback: %v", err)
		}
		outcome := result.OrdinarySelectionPlan.InstrumentationPlan.Outcomes[0]
		if target == "linux/arm64" {
			if outcome.Status != "failed" || outcome.Early != nil {
				t.Fatal("supported target's missing CGO was relabeled as official target rejection")
			}
		} else if outcome.Status != "early_target_rejected" || outcome.Early == nil {
			t.Fatal("official target rejection lost its actual source/driver evidence")
		}
		testInstrumentationCorpusTransport(t, result, inventory, target)
		mutations := []func(*discoveryInstrumentationPlan){
			func(p *discoveryInstrumentationPlan) { p.Outcomes = nil },
			func(p *discoveryInstrumentationPlan) {
				p.Requirements, p.Outcomes = []discoveryInstrumentationRequirement{}, nil
			},
			func(p *discoveryInstrumentationPlan) { p.Outcomes[0].Stage = "complete" },
			func(p *discoveryInstrumentationPlan) { p.Outcomes[0].Status = "llvm_objects_only" },
		}
		if outcome.Early != nil {
			mutations = append(mutations,
				func(p *discoveryInstrumentationPlan) { p.Outcomes[0].Early = nil },
				func(p *discoveryInstrumentationPlan) { p.Outcomes[0].Early.CGOEnabled = "0" })
		}
		for _, mutate := range mutations {
			body, err := json.Marshal(result)
			if err != nil {
				t.Fatal(err)
			}
			var changed discoveryCorpusResult
			if err := json.Unmarshal(body, &changed); err != nil {
				t.Fatal(err)
			}
			mutate(changed.OrdinarySelectionPlan.InstrumentationPlan)
			if err := validateDiscoveryInstrumentationFailure(changed, inventory); err == nil {
				t.Fatal("early rejection lost its required scope/source/driver evidence")
			}
		}
		return
	} else if err != nil {
		t.Fatalf("stock-Go-accepted original mode pair was not consumed by the corpus: %v", err)
	}
	if matrix.InstrumentationTranslations != 1 || matrix.Success != 1 || len(configs) != 0 ||
		len(matrix.FeatureConsumption) != 0 || matrix.OrdinarySelectionPlan == nil ||
		matrix.OrdinarySelectionPlan.InstrumentationPlan == nil || len(matrix.OrdinarySelectionPlan.InstrumentationPlan.Outcomes) != 1 {
		t.Fatalf("mode whole-file consumer was lost or relabeled as ordinary evidence: %+v", matrix)
	}
	inventory := newDiscoveryFeatureInventory()
	refs, err := registerDiscoveryFeatureProfiles(inventory, matrix.FeatureProfiles)
	if err != nil {
		t.Fatal(err)
	}
	result := discoveryCorpusResult{
		Module: candidate.Module, Version: candidate.Version, Status: discoveryStatusPassed,
		DiscoveredAsmFiles: candidate.AsmFiles,
		ApplicableAsmFiles: discoveryInstrumentationExecutedFiles(matrix.OrdinarySelectionPlan),
		Translations:       matrix.Success, InstrumentationTranslations: matrix.InstrumentationTranslations,
		BuildConfigurations: configs, SourceNotApplicableItems: matrix.SourceNotApplicableItems,
		OrdinarySelectionPlan: matrix.OrdinarySelectionPlan, FeatureProfiles: refs,
		FeatureConsumption: matrix.FeatureConsumption,
	}
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	var replayed discoveryCorpusResult
	if err := json.Unmarshal(data, &replayed); err != nil {
		t.Fatal(err)
	}
	replayed.featureInventory = inventory
	if err := validateOrdinaryProfileResult(replayed, []string{target}, matrix.OrdinarySelectionPlan.GoVersion); err != nil {
		t.Fatalf("independent complete corpus mode report reader: %v", err)
	}
	testInstrumentationCorpusTransport(t, replayed, inventory, target)
	tests := []struct {
		name   string
		mutate func(*discoveryCorpusResult)
	}{
		{"plan deleted", func(r *discoveryCorpusResult) { r.OrdinarySelectionPlan.InstrumentationPlan = nil }},
		{"paired entire mode erased", func(r *discoveryCorpusResult) {
			r.OrdinarySelectionPlan.InstrumentationPlan = nil
			r.InstrumentationTranslations, r.Translations = 0, 0
			r.Status, r.NotApplicableReason, r.ApplicableAsmFiles = discoveryStatusNotApplicable, "baseline has no selected package", nil
		}},
		{"paired empty denominator and outcomes", func(r *discoveryCorpusResult) {
			r.OrdinarySelectionPlan.InstrumentationPlan = &discoveryInstrumentationPlan{Protocol: discoveryInstrumentationPlanProtocol, Requirements: []discoveryInstrumentationRequirement{}}
			r.InstrumentationTranslations, r.Translations, r.ApplicableAsmFiles = 0, 0, nil
		}},
		{"null denominator", func(r *discoveryCorpusResult) { r.OrdinarySelectionPlan.InstrumentationPlan.Requirements = nil }},
		{"ordinary protocol", func(r *discoveryCorpusResult) {
			r.OrdinarySelectionPlan.InstrumentationPlan.Protocol = ordinarySelectionProtocol
		}},
		{"changed mode", func(r *discoveryCorpusResult) {
			r.OrdinarySelectionPlan.InstrumentationPlan.Requirements[0].Mode = "asan"
		}},
		{"changed CGO", func(r *discoveryCorpusResult) {
			r.OrdinarySelectionPlan.InstrumentationPlan.Requirements[0].CGO = map[string]string{"0": "1", "1": "0"}[cgo]
		}},
		{"mode as custom tag", func(r *discoveryCorpusResult) {
			r.OrdinarySelectionPlan.InstrumentationPlan.Requirements[0].Tags = []string{"race"}
		}},
		{"source pair deleted", func(r *discoveryCorpusResult) {
			r.OrdinarySelectionPlan.InstrumentationPlan.Requirements[0].Pairs = nil
		}},
		{"outcomes deleted", func(r *discoveryCorpusResult) { r.OrdinarySelectionPlan.InstrumentationPlan.Outcomes = nil }},
		{"extra outcome", func(r *discoveryCorpusResult) {
			p := r.OrdinarySelectionPlan.InstrumentationPlan
			p.Outcomes = append(p.Outcomes, p.Outcomes[0])
		}},
		{"original source inventory deleted", func(r *discoveryCorpusResult) { r.OrdinarySelectionPlan.InstrumentationPlan.Sources = nil }},
		{"full original Go source changed", func(r *discoveryCorpusResult) {
			r.OrdinarySelectionPlan.InstrumentationPlan.Sources["."].Contents["probe.go"] += "\n// changed original bytes\n"
		}},
		{"actual observation deleted", func(r *discoveryCorpusResult) { r.OrdinarySelectionPlan.InstrumentationPlan.Observations = nil }},
		{"raw CPP registration deleted", func(r *discoveryCorpusResult) { r.OrdinarySelectionPlan.InstrumentationPlan.CPPInputs = nil }},
		{"independent query deleted", func(r *discoveryCorpusResult) { r.OrdinarySelectionPlan.InstrumentationPlan.Outcomes[0].Query = nil }},
		{"objects deleted", func(r *discoveryCorpusResult) { r.OrdinarySelectionPlan.InstrumentationPlan.Outcomes[0].Objects = nil }},
		{"early support relabeled completion", func(r *discoveryCorpusResult) {
			r.OrdinarySelectionPlan.InstrumentationPlan.Outcomes[0].Status = "accepted_early_rule"
		}},
		{"count reduced", func(r *discoveryCorpusResult) { r.InstrumentationTranslations = 0 }},
		{"ordinary count substituted", func(r *discoveryCorpusResult) { r.Translations++ }},
		{"missing physical file scope", func(r *discoveryCorpusResult) { r.ApplicableAsmFiles = nil }},
		{"same paired layout forged", func(r *discoveryCorpusResult) {
			o := &r.OrdinarySelectionPlan.InstrumentationPlan.Outcomes[0]
			for _, metadata := range []*gotoolprofile.InstrumentationMetadataProof{o.Query.Metadata, o.Objects.Metadata, o.Objects.Consumption.Metadata} {
				metadata.Declarations[0].Layout = []byte(strings.ReplaceAll(string(metadata.Declarations[0].Layout), "i64", "i32"))
			}
			for index := range o.Objects.Consumption.Files[0].Declarations {
				d := &o.Objects.Consumption.Files[0].Declarations[index]
				d.Layout = []byte(strings.ReplaceAll(string(d.Layout), "i64", "i32"))
			}
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var changed discoveryCorpusResult
			if err := json.Unmarshal(data, &changed); err != nil {
				t.Fatal(err)
			}
			changed.featureInventory = inventory
			test.mutate(&changed)
			if err := validateOrdinaryProfileResult(changed, []string{target}, matrix.OrdinarySelectionPlan.GoVersion); err == nil {
				t.Fatal("forged/erased mode corpus evidence passed independent replay")
			}
		})
	}
	if output := os.Getenv("PLAN9ASM_INSTRUMENTATION_CORPUS_ARTIFACTS"); output != "" {
		output = filepath.Join(output, "cgo-"+cgo)
		if err := os.MkdirAll(output, 0700); err != nil {
			t.Fatal(err)
		}
		for name, body := range map[string][]byte{"result.json": data} {
			if err := os.WriteFile(filepath.Join(output, name), body, 0600); err != nil {
				t.Fatal(err)
			}
		}
		body, err := json.MarshalIndent(inventory, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(output, "feature-inventory.json"), body, 0600); err != nil {
			t.Fatal(err)
		}
	}
}

func testInstrumentationCorpusTransport(t *testing.T, result discoveryCorpusResult, inventory *discoveryFeatureInventory, target string) {
	t.Helper()
	ledger := filepath.Join(t.TempDir(), "scan")
	if err := os.MkdirAll(filepath.Join(ledger, "records"), 0700); err != nil {
		t.Fatal(err)
	}
	record := discoveryRecord{Kind: "matched", Module: result.Module, Version: result.Version,
		Architectures: []string{"arm64"}, AsmFiles: result.DiscoveredAsmFiles}
	body, err := json.Marshal(record)
	if err != nil {
		t.Fatal(err)
	}
	writeTestFile(t, filepath.Join(ledger, "records", "source.jsonl"), string(body)+"\n")
	provenance := fixtureDiscoveryProvenance(t, ledger)
	provenance.GoVersion = result.OrdinarySelectionPlan.GoVersion
	for _, outcome := range result.OrdinarySelectionPlan.InstrumentationPlan.Outcomes {
		if outcome.Objects != nil {
			provenance.LLCBinarySHA256 = outcome.Objects.LLVM.SHA256
			match := discoveryLLVMVersionPattern.FindStringSubmatch(outcome.Objects.LLVM.Version)
			if len(match) != 2 {
				t.Fatal("actual mode object lacked its LLVM version")
			}
			provenance.LLVMVersion = match[1]
		}
	}
	report := discoveryCorpusReport{
		SchemaVersion: discoveryReportSchema, Provenance: provenance, Targets: []string{target},
		ShardCount: 1, CandidateTotal: 1, EligibleCandidates: 1, Selected: 1,
		Translations: result.Translations, ModeTranslations: result.InstrumentationTranslations,
		FeatureInventory: inventory, Results: []discoveryCorpusResult{result},
	}
	if result.Status == discoveryStatusPassed {
		report.Passed = 1
	} else {
		report.Failed = 1
	}
	if err := validateDiscoveryCorpusAccounting(report); err != nil {
		t.Fatalf("actual mode report accounting: %v", err)
	}
	reports := t.TempDir()
	if err := writeDiscoveryCorpusReport(filepath.Join(reports, "shard-0.json"), report); err != nil {
		t.Fatal(err)
	}
	progress, err := collectDiscoveryProgress(ledger, reports, []string{target}, provenance.Source, 1)
	if err != nil {
		t.Fatalf("complete mode progress readback: %v", err)
	}
	if progress.ModeTranslations != result.InstrumentationTranslations ||
		progress.Translations != result.Translations || progress.Verified != (result.Status == discoveryStatusPassed) {
		t.Fatalf("mode evidence was relabeled/lost during collection: %+v", progress)
	}
	output := filepath.Join(t.TempDir(), "assembly-ledger")
	if err := writeAssemblyLedger(output, progress, strings.Repeat("a", 64)); err != nil {
		t.Fatal(err)
	}
	replayed, err := readAssemblyLedger(output, progress.LedgerSHA256, strings.Repeat("a", 64))
	if err != nil {
		t.Fatalf("complete mode ledger independent readback: %v", err)
	}
	if replayed.ModeTranslations != progress.ModeTranslations || replayed.Verified != progress.Verified ||
		replayed.Candidates[0].OrdinarySelectionPlan.InstrumentationPlan == nil {
		t.Fatal("ledger silently downgraded complete mode scopes")
	}
	serialized, err := json.Marshal(replayed)
	if err != nil {
		t.Fatal(err)
	}
	for index, mutate := range []func(*discoveryProgress){
		func(p *discoveryProgress) { p.ModeTranslations++ },
		func(p *discoveryProgress) { p.Candidates[0].InstrumentationTranslations++ },
		func(p *discoveryProgress) { p.Candidates[0].OrdinarySelectionPlan.InstrumentationPlan = nil },
		func(p *discoveryProgress) { p.Candidates[0].OrdinarySelectionPlan.InstrumentationPlan.Outcomes = nil },
		func(p *discoveryProgress) {
			p.Candidates[0].OrdinarySelectionPlan.InstrumentationPlan.Requirements = []discoveryInstrumentationRequirement{}
			p.Candidates[0].OrdinarySelectionPlan.InstrumentationPlan.Outcomes = nil
			p.ModeTranslations, p.Candidates[0].InstrumentationTranslations = 0, 0
		},
	} {
		var changed discoveryProgress
		if err := json.Unmarshal(serialized, &changed); err != nil {
			t.Fatal(err)
		}
		mutate(&changed)
		if err := validateAssemblyLedgerProgress(changed); err == nil {
			t.Fatalf("ledger mutation %d erased or changed the independent mode denominator", index)
		}
	}
	changed := report
	changed.ModeTranslations++
	if err := validateDiscoveryCorpusAccounting(changed); err == nil {
		t.Fatal("report mode count disagreed with actual per-case objects")
	}
	if result.InstrumentationTranslations != 0 {
		changed = report
		changed.Provenance.LLCBinarySHA256 = strings.Repeat("f", 64)
		if err := validateDiscoveryCorpusAccounting(changed); err == nil {
			t.Fatal("report substituted the actual LLVM tool")
		}
	}
	t.Logf("%s CGO scopes: independent report/progress/ledger readback; status=%s objects=%d", target, result.Status, result.InstrumentationTranslations)
}
