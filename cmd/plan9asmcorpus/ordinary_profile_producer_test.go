package main

import (
	"archive/zip"
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

func TestOrdinaryProducerActuallyConsumesCPUProfiles(t *testing.T) {
	compilerEnvironment := os.Environ()
	if runtimeGoMinorForProfileTest(t) < 27 {
		// External-corpus execution requires Go 1.27; Go 1.20 still tests the
		// shared API and portable guards, not a counterfeit corpus producer.
		t.Skip("actual external-corpus producer requires Go 1.27; portable protocol units also run on Go 1.20")
	}
	candidate := discoveryCandidate{Module: "example.invalid/profile", Version: "v1.0.0", AsmFiles: []string{"probe_amd64.s"}}
	sources := map[string]string{
		"go.mod":        "module example.invalid/profile\n\ngo 1.20\n",
		"empty.s":       "", // Go selects this even though Discovery inventories only nonempty assembly.
		"probe.go":      "//go:build amd64.v3\n\npackage profile\nfunc probe() uint64\n",
		"probe_amd64.s": "//go:build amd64.v3\n\n#ifdef GOAMD64_v4\nTEXT ·probe(SB),$0-8\nMOVQ $4, ret+0(FP)\nRET\n#else\nTEXT ·probe(SB),$0-8\nMOVQ $3, ret+0(FP)\nRET\n#endif\n",
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
	for file, data := range map[string][]byte{
		"v1.0.0.mod":  []byte(sources["go.mod"]),
		"v1.0.0.info": []byte(`{"Version":"v1.0.0","Time":"2026-10-01T00:00:00Z"}`),
		"v1.0.0.zip":  archive,
	} {
		if err := os.WriteFile(filepath.Join(versions, file), data, 0600); err != nil {
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
	command.Dir = filepath.Join(root, "cmd", "plan9asmll")
	command.Env = compilerEnvironment
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("build our translator: %v\n%s", err, output)
	}
	llc, err := exec.LookPath("llc")
	if err != nil {
		t.Fatal(err)
	}
	workDir := filepath.Join(t.TempDir(), "candidate")
	matrix, _, configs, err := runDiscoveryCandidate(discoveryCorpusConfig{
		RepoRoot: root, Translator: binary, LLC: llc, Targets: []string{"linux/amd64"},
		CandidateTimeout: time.Minute,
	}, candidate, workDir)
	if err != nil {
		t.Fatal(err)
	}
	if matrix.Success != 2 || matrix.NotApplicable != 0 || len(matrix.FeatureConsumption) != 2 || len(configs) != 2 || len(matrix.FeatureProfiles) != 3 {
		t.Fatalf("actual Go export/vet/LLVM profile consumers lost their scopes: %#v configs=%#v", matrix, configs)
	}
	if matrix.FeatureConsumption[0].Outputs[0].IR == matrix.FeatureConsumption[1].Outputs[0].IR {
		t.Fatal("distinct actual v3/v4 CPP branches produced the same LLVM IR")
	}
	for _, proof := range matrix.FeatureConsumption {
		if len(proof.Packages) != 1 || !equalDiscoveryStrings(proof.Packages[0].SFiles, []string{"empty.s", "probe_amd64.s"}) {
			t.Fatal("actual ordinary Go export lost its selected zero-byte assembly sibling")
		}
		var selected discoveryFeatureProfile
		for _, profile := range matrix.FeatureProfiles {
			if profile.ID == proof.ProfileID {
				selected = profile
			}
		}
		input := ordinaryProfileConsumerInput(matrix.OrdinarySelectionPlan, selected, candidate.Module, "", candidate.AsmFiles)
		if err := gotoolprofile.ValidateSelection(input, proof, nil, true); err != nil {
			t.Fatalf("actual Go-selected zero-byte sibling rejected on replay: %v", err)
		}
		for _, body := range []string{"\n", "// placeholder\n", "TEXT ·uncovered(SB),$0-0\nRET\n"} {
			// This checks the independent scope guard, not authenticity: the ZIP
			// verifier separately rejects changing an original file's contents.
			changed := *proof
			changed.Packages = append([]gotoolprofile.PackageProof(nil), proof.Packages...)
			changed.Packages[0].SourceSHA256 = make(map[string]string)
			for file, digest := range proof.Packages[0].SourceSHA256 {
				changed.Packages[0].SourceSHA256[file] = digest
			}
			input.Sources["empty.s"] = discoveryFeatureBytesSHA256([]byte(body))
			changed.Packages[0].SourceSHA256["empty.s"] = input.Sources["empty.s"]
			if err := gotoolprofile.ValidateSelection(input, &changed, nil, true); err == nil {
				t.Fatalf("nonempty selected assembly outside the tested scope accepted: %q", body)
			}
		}
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
		OrdinarySelectionPlan: matrix.OrdinarySelectionPlan, FeatureProfiles: refs,
		FeatureConsumption: matrix.FeatureConsumption,
	}
	encoded, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	var replayed discoveryCorpusResult
	if err := json.Unmarshal(encoded, &replayed); err != nil {
		t.Fatal(err)
	}
	replayed.featureInventory = inventory
	if err := validateOrdinaryProfileResult(replayed, []string{"linux/amd64"}, matrix.OrdinarySelectionPlan.GoVersion); err != nil {
		t.Fatalf("actual production proof JSON replay: %v", err)
	}
}

func runtimeGoMinorForProfileTest(t *testing.T) int {
	t.Helper()
	minor, err := discoveryGoMinor(runtime.Version())
	if err != nil {
		t.Fatal(err)
	}
	return minor
}

func TestOrdinaryProducerZeroByteAssemblySiblingFiveArchitectures(t *testing.T) {
	compilerEnvironment := os.Environ()
	if runtimeGoMinorForProfileTest(t) < 27 {
		t.Skip("actual external-corpus producer requires Go 1.27; portable protocol units also run on Go 1.20")
	}
	candidate := discoveryCandidate{
		Module: "example.invalid/empty-sibling", Version: "v1.0.0", AsmFiles: []string{"probe.s"},
	}
	sources := map[string]string{
		"go.mod":   "module example.invalid/empty-sibling\n\ngo 1.20\n",
		"probe.go": "package sibling\nfunc Probe()\n",
		"probe.s":  "TEXT ·Probe(SB),$0-0\nRET\n",
		"empty.s":  "",
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
		"v1.0.0.info": []byte(`{"Version":"v1.0.0","Time":"2026-10-01T00:00:00Z"}`),
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
	command := exec.Command("go", "build", "-p=1", "-o", binary, ".")
	command.Dir = filepath.Join(root, "cmd", "plan9asmll")
	command.Env = compilerEnvironment
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("build our translator: %v\n%s", err, output)
	}
	llc, err := exec.LookPath("llc")
	if err != nil {
		t.Fatal(err)
	}
	targets := []string{"linux/386", "linux/amd64", "linux/arm", "linux/arm64", "js/wasm"}
	matrix, _, configs, err := runDiscoveryCandidate(discoveryCorpusConfig{
		RepoRoot: root, Translator: binary, LLC: llc, Targets: targets, CandidateTimeout: 2 * time.Minute,
	}, candidate, filepath.Join(t.TempDir(), "candidate"))
	if err != nil {
		t.Fatal(err)
	}
	if matrix.Success != len(targets) || matrix.NotApplicable != 0 || len(configs) != len(targets) || len(matrix.FeatureConsumption) != len(targets) {
		t.Fatalf("zero-byte sibling changed the actual five-architecture scope: successes=%d N/A=%d configurations=%d consumers=%d",
			matrix.Success, matrix.NotApplicable, len(configs), len(matrix.FeatureConsumption))
	}
	for _, proof := range matrix.FeatureConsumption {
		if len(proof.Packages) != 1 || !equalDiscoveryStrings(proof.Packages[0].SFiles, []string{"empty.s", "probe.s"}) {
			t.Fatal("actual Go-selected package did not consume its zero-byte sibling")
		}
		if len(proof.CPP) != 1 || len(proof.Outputs) != 1 || proof.CPP[0].File != "probe.s" || proof.Outputs[0].File != "probe.s" {
			t.Fatal("zero-byte sibling manufactured a translation or LLVM output count")
		}
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
		OrdinarySelectionPlan: matrix.OrdinarySelectionPlan, FeatureProfiles: refs,
		FeatureConsumption: matrix.FeatureConsumption,
	}
	encoded, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	var replayed discoveryCorpusResult
	if err := json.Unmarshal(encoded, &replayed); err != nil {
		t.Fatal(err)
	}
	replayed.featureInventory = inventory
	if err := validateOrdinaryProfileResult(replayed, targets, matrix.OrdinarySelectionPlan.GoVersion); err != nil {
		t.Fatalf("actual five-architecture object proof JSON replay: %v", err)
	}
	if err := os.WriteFile(filepath.Join(download.Dir, "empty.s"), []byte("\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := verifyDiscoveryOrdinaryCPP(matrix.OrdinarySelectionPlan, download.Dir, root, candidate); err == nil {
		t.Fatal("changed zero-byte original source escaped exact ZIP/source stability checks")
	}
}

func TestOrdinaryProducerProfilesKeepFeatureOnlyAssemblyAndCPPBranches(t *testing.T) {
	candidate := discoveryCandidate{Module: "example.invalid/profile", Version: "v1.0.0", AsmFiles: []string{"probe_amd64.s"}}
	sources := map[string]string{
		"go.mod":        "module example.invalid/profile\n\ngo 1.20\n",
		"probe.go":      "//go:build amd64.v3\n\npackage profile\nfunc probe()\n",
		"probe_amd64.s": "//go:build amd64.v3\n\n#ifdef GOAMD64_v4\nTEXT ·probe(SB),$0-0\nNOP\nRET\n#else\nTEXT ·probe(SB),$0-0\nRET\n#endif\n",
	}
	download := fixtureProfileModuleDownload(t, candidate, sources)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	plan, profiles, configs, rejected, root, err := captureDiscoveryOrdinaryProfiles(ctx, candidate, download, []string{"linux/amd64"}, t.TempDir(), os.Environ())
	if err != nil {
		t.Fatal(err)
	}
	if root == "" || plan.CPPInputs == nil || len(plan.CPPInputs.Units) != 1 || len(rejected) != 0 {
		t.Fatalf("feature-only source/CPP union was lost: root=%q plan=%#v rejected=%#v", root, plan, rejected)
	}
	levels := make(map[string]bool)
	for _, profile := range profiles {
		levels[profile.Observed.Environment["GOAMD64"]] = true
	}
	if len(profiles) != 3 || !levels["v1"] || !levels["v3"] || !levels["v4"] {
		t.Fatalf("source + CPP-required profiles = %#v", levels)
	}
	_, scopes, err := replayOrdinarySelectionForProfiles(plan, candidate.AsmFiles, profiles)
	if err != nil {
		t.Fatal(err)
	}
	keys, err := ordinaryProfileConfigurationKeys(configs, scopes)
	if err != nil || len(keys) != 2 {
		t.Fatalf("production configurations collapsed v3/v4 scopes: %#v %v", configs, err)
	}
	if err := verifyDiscoveryOrdinaryCPP(plan, download.Dir, root, candidate); err != nil {
		t.Fatal(err)
	}
	// The producer still compares original bytes, not self-reported headers.
	writeTestFile(t, filepath.Join(download.Dir, "probe.go"), strings.ReplaceAll(sources["probe.go"], "amd64.v3", "amd64.v1"))
	if err := verifyDiscoveryOrdinaryCPP(plan, download.Dir, root, candidate); err == nil {
		t.Fatal("changed source header retained the original selection proof")
	}
}

func fixtureProfileModuleDownload(t *testing.T, candidate discoveryCandidate, sources map[string]string) moduleDownloadInfo {
	t.Helper()
	dir := t.TempDir()
	archivePath := filepath.Join(t.TempDir(), "source.zip")
	archive, err := os.Create(archivePath)
	if err != nil {
		t.Fatal(err)
	}
	writer := zip.NewWriter(archive)
	for file, data := range sources {
		if err := os.MkdirAll(filepath.Dir(filepath.Join(dir, file)), 0755); err != nil {
			t.Fatal(err)
		}
		writeTestFile(t, filepath.Join(dir, file), data)
		entry, err := writer.Create(candidate.Module + "@" + candidate.Version + "/" + file)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := entry.Write([]byte(data)); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	if err := archive.Close(); err != nil {
		t.Fatal(err)
	}
	return moduleDownloadInfo{Path: candidate.Module, Version: candidate.Version, Dir: dir, GoMod: filepath.Join(dir, "go.mod"), Zip: archivePath}
}
