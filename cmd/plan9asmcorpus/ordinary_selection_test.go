package main

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"go/build"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestOrdinarySelectionRejectsReasonOnlyNotApplicable(t *testing.T) {
	for _, reason := range []string{
		"no discovered assembly file belongs to a buildable Go package on the supported target matrix",
		"every target-selected assembly source has evidence-backed current-Go source or ABI incompatibility",
	} {
		report := discoveryCorpusReport{
			SchemaVersion: discoveryReportSchema, Targets: []string{"linux/arm64"},
			Selected: 1, NotApplicable: 1,
			Results: []discoveryCorpusResult{{
				Module: "example.com/ordinary", Version: "v1.0.0", Status: discoveryStatusNotApplicable,
				DiscoveredAsmFiles: []string{"native_arm64.s"}, NotApplicableReason: reason,
			}},
		}
		if err := validateDiscoveryCorpusAccounting(report); err == nil {
			t.Fatalf("accepted an ordinary N/A with only a reason: %s", reason)
		}
	}
}

func TestOrdinarySelectionCompactHeaderPreservesGoMatchFile(t *testing.T) {
	for _, header := range []string{
		"// License/documentation payload\n// +build feature\n\npackage ordinary\n",
		"// License/documentation payload\n// +build feature\npackage ordinary\n",
		"// License/documentation payload\n//go:build feature && !ignore\n\npackage ordinary\n",
		"/* block with apparent directive\n//go:build ignore\n*/\n\npackage ordinary\n",
	} {
		compact := compactOrdinarySelectionHeader(header)
		for _, tags := range [][]string{nil, {"feature"}, {"ignore"}, {"feature", "ignore"}} {
			matches := func(text string) (bool, error) {
				ctx := build.Default
				ctx.GOOS, ctx.GOARCH, ctx.Compiler, ctx.CgoEnabled = "linux", "arm64", "gc", false
				ctx.BuildTags = tags
				ctx.OpenFile = func(string) (io.ReadCloser, error) { return io.NopCloser(strings.NewReader(text)), nil }
				return ctx.MatchFile(".", "decl.go")
			}
			original, originalErr := matches(header)
			actual, actualErr := matches(compact)
			if actual != original || (actualErr == nil) != (originalErr == nil) {
				t.Fatalf("canonical header changed MatchFile: original=%v/%v compact=%v/%v", original, originalErr, actual, actualErr)
			}
		}
	}
}

func TestOrdinarySelectionLegacyReportsFailEveryReader(t *testing.T) {
	for _, schema := range []int{8, discoveryReportSchema} {
		t.Run(fmt.Sprintf("schema_%d", schema), func(t *testing.T) {
			ledger, reports, source := writeDiscoveryReportFixture(t)
			filenames, err := discoveryCorpusReportFiles(reports)
			if err != nil {
				t.Fatal(err)
			}
			for _, filename := range filenames {
				report, err := readDiscoveryCorpusReport(filename)
				if err != nil {
					t.Fatal(err)
				}
				report.SchemaVersion = schema
				report.NotApplicable, report.Passed, report.Translations = report.Passed, 0, 0
				for i := range report.Results {
					report.Results[i].Status = discoveryStatusNotApplicable
					report.Results[i].Translations = 0
					report.Results[i].NotApplicableReason = "no buildable package"
				}
				if err := writeDiscoveryCorpusReport(filename, report); err != nil {
					t.Fatal(err)
				}
			}
			targets := []string{"linux/amd64", "linux/arm64"}
			if err := verifyDiscoveryCorpusReports(ledger, reports, targets, source); err == nil {
				t.Fatal("strict aggregate accepted a legacy or relabeled reason-only N/A")
			}
			if _, err := collectDiscoveryProgress(ledger, reports, targets, source, 2); err == nil {
				t.Fatal("progress accepted a legacy or relabeled reason-only N/A")
			}
		})
	}
}

func TestOrdinarySelectionLedgerRejectsMissingProof(t *testing.T) {
	ledger, reports, source := writeDiscoveryReportFixture(t)
	progress, err := collectDiscoveryProgress(ledger, reports, []string{"linux/amd64", "linux/arm64"}, source, 2)
	if err != nil {
		t.Fatal(err)
	}
	candidate := &progress.Candidates[0]
	progress.Passed--
	progress.NotApplicable++
	progress.Translations -= candidate.Translations
	candidate.Status, candidate.Translations = discoveryStatusNotApplicable, 0
	candidate.NotApplicableReason = "no buildable package"
	if err := validateAssemblyLedgerProgress(progress); err == nil {
		t.Fatal("ledger accepted reason-only N/A without exact source-selection proof")
	}
}

func TestOrdinarySelectionZIPBindsSourceAndDirectoryInventory(t *testing.T) {
	sourceDir := t.TempDir()
	writeTestFile(t, filepath.Join(sourceDir, "decl.go"), "package ordinary\n")
	writeTestFile(t, filepath.Join(sourceDir, "native_arm64.s"), "TEXT ·F(SB),$0-0\nRET\n")
	archivePath := filepath.Join(t.TempDir(), "module.zip")
	archive, err := os.Create(archivePath)
	if err != nil {
		t.Fatal(err)
	}
	writer := zip.NewWriter(archive)
	for name, data := range map[string]string{
		"example.com/ordinary@v1.0.0/":               "",
		"example.com/ordinary@v1.0.0/empty/":         "",
		"example.com/ordinary@v1.0.0/decl.go":        "package ordinary\n",
		"example.com/ordinary@v1.0.0/native_arm64.s": "TEXT ·F(SB),$0-0\nRET\n",
	} {
		entry, err := writer.Create(name)
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
	candidate := discoveryCandidate{Module: "example.com/ordinary", Version: "v1.0.0", AsmFiles: []string{"native_arm64.s"}}
	plan, err := captureOrdinarySelectionPlan(candidate, sourceDir, []string{"linux/arm64"})
	if err != nil {
		t.Fatal(err)
	}
	if err := verifyOrdinarySelectionZIP(plan, archivePath, candidate.Module, candidate.Version, ""); err != nil {
		t.Fatalf("legal directory marker rejected: %v", err)
	}
	for _, file := range []string{"decl.go", "new.go", "native_arm64.s"} {
		t.Run(file, func(t *testing.T) {
			changed := *plan
			changed.Sources = append([]ordinarySelectionSource(nil), plan.Sources...)
			for i := range changed.Sources {
				if changed.Sources[i].File == file {
					changed.Sources[i].SHA256 = strings.Repeat("f", 64)
				}
			}
			if file == "new.go" {
				changed.Directories = append([]ordinarySelectionDirectory(nil), plan.Directories...)
				changed.Directories[0].Entries = append(append([]ordinarySelectionEntry(nil), plan.Directories[0].Entries...), ordinarySelectionEntry{Name: file, Kind: "file"})
			}
			if err := verifyOrdinarySelectionZIP(&changed, archivePath, candidate.Module, candidate.Version, ""); err == nil {
				t.Fatal("source/directory mutation accepted as original exact module ZIP")
			}
		})
	}
	changed := *plan
	changed.Sources = append([]ordinarySelectionSource(nil), plan.Sources...)
	changed.Sources[0].Header = "//go:build ignore\n\npackage ordinary\n"
	changed.Sources[0].HeaderSHA256 = fmt.Sprintf("%x", sha256.Sum256([]byte(changed.Sources[0].Header)))
	if err := verifyOrdinarySelectionZIP(&changed, archivePath, candidate.Module, candidate.Version, ""); err == nil {
		t.Fatal("forged selection header with unchanged original source SHA accepted")
	}
	if err := verifyOrdinarySelectionZIP(plan, archivePath, candidate.Module, candidate.Version, "h1:changed"); err == nil {
		t.Fatal("changed module h1 accepted")
	}
}

func TestOrdinarySelectionInputMutationAfterCaptureFails(t *testing.T) {
	for _, file := range []string{"decl.go", "native_arm64.s", "layout.h", "new.go"} {
		t.Run(file, func(t *testing.T) {
			dir := t.TempDir()
			for name, source := range map[string]string{
				"decl.go": "package ordinary\n", "native_arm64.s": "TEXT ·F(SB),$0-0\nRET\n", "layout.h": "#define FIELD 8\n",
			} {
				writeTestFile(t, filepath.Join(dir, name), source)
			}
			candidate := discoveryCandidate{Module: "example.com/ordinary", Version: "v1.0.0", AsmFiles: []string{"native_arm64.s"}}
			plan, err := captureOrdinarySelectionPlan(candidate, dir, []string{"linux/arm64"})
			if err != nil {
				t.Fatal(err)
			}
			if err := verifyOrdinarySelectionUnchanged(plan, dir, candidate); err != nil {
				t.Fatal(err)
			}
			writeTestFile(t, filepath.Join(dir, file), "// changed after capture\npackage ordinary\n")
			if err := verifyOrdinarySelectionUnchanged(plan, dir, candidate); err == nil {
				t.Fatal("changed selected source/header or added package file accepted")
			}
		})
	}
}

func TestOrdinarySelectionUnsuffixedRejectionRetainsCustomTags(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, filepath.Join(dir, "decl.go"), "package ordinary\n")
	writeTestFile(t, filepath.Join(dir, "native.s"), "//go:build feature\n\nTEXT ·F(SB),$0-0\nDELIBERATELY_INVALID_OPCODE\n")
	candidate := discoveryCandidate{Module: "example.com/ordinary", Version: "v1.0.0", AsmFiles: []string{"native.s"}}
	var evidence []discoverySourceNotApplicableItem
	configs, err := discoveryBuildConfigurationsWithEvidence(candidate, dir, []string{"linux/arm64"}, &evidence)
	if err != nil {
		t.Fatal(err)
	}
	if len(configs) != 0 || len(evidence) != 1 || !equalDiscoveryStrings(evidence[0].BuildTags, []string{"feature"}) {
		t.Fatalf("actual assembler rejection lost selected tags: configs=%+v evidence=%+v", configs, evidence)
	}
	plan, err := captureOrdinarySelectionPlan(candidate, dir, []string{"linux/arm64"})
	if err != nil {
		t.Fatal(err)
	}
	plan.ModuleSum, plan.ZipSHA256 = "h1:"+strings.Repeat("A", 43)+"=", strings.Repeat("a", 64)
	result := discoveryCorpusResult{Module: candidate.Module, Version: candidate.Version, Status: discoveryStatusNotApplicable,
		DiscoveredAsmFiles: candidate.AsmFiles, SourceNotApplicableItems: evidence, OrdinarySelectionPlan: plan}
	if err := validateOrdinarySelectionResult(result, plan.Targets, plan.GoVersion); err != nil {
		t.Fatal(err)
	}
}

func fixtureOrdinarySelection(t *testing.T, files []string, targets []string, sources map[string]string) *discoveryOrdinarySelectionPlan {
	t.Helper()
	dir := t.TempDir()
	for file, data := range sources {
		if err := os.MkdirAll(filepath.Dir(filepath.Join(dir, filepath.FromSlash(file))), 0755); err != nil {
			t.Fatal(err)
		}
		writeTestFile(t, filepath.Join(dir, filepath.FromSlash(file)), data)
	}
	plan, err := captureOrdinarySelectionPlan(discoveryCandidate{
		Module: "example.com/ordinary", Version: "v1.0.0", AsmFiles: files,
	}, dir, targets)
	if err != nil {
		t.Fatal(err)
	}
	plan.ModuleSum = "h1:" + strings.Repeat("A", 43) + "="
	plan.ZipSHA256 = strings.Repeat("a", 64)
	// Synthetic report fixtures describe the pinned external-corpus context,
	// independently of the Go version running the repository's unit matrix.
	plan.GoVersion = "go1.27.1"
	plan.ReleaseTags = nil
	for minor := 1; minor <= 27; minor++ {
		plan.ReleaseTags = append(plan.ReleaseTags, fmt.Sprintf("go1.%d", minor))
	}
	plan.Decisions, _, err = replayOrdinarySelection(plan, files)
	if err != nil {
		t.Fatal(err)
	}
	return plan
}

func TestOrdinarySelectionRootAndPreciseExclusions(t *testing.T) {
	for _, tc := range []struct {
		name, file, goFile, goSource, asmSource, want, boundary string
	}{
		{name: "root is eligible", file: "native_arm64.s", goFile: "decl.go", goSource: "package ordinary\n", asmSource: "TEXT ·F(SB),$0-0\nRET\n", want: nativeLayoutSelected},
		{name: "root suffix excludes target", file: "native_windows_arm64.s", goFile: "decl.go", goSource: "package ordinary\n", want: ordinarySelectionFilename},
		{name: "hidden filename is not ignored directory", file: "_native_arm64.s", goFile: "decl.go", goSource: "package ordinary\n", want: ordinarySelectionIgnoredFilename},
		{name: "cgo constraint", file: "native_arm64.s", goFile: "decl.go", goSource: "//go:build cgo\n\npackage ordinary\n", want: ordinarySelectionCgoDisabled},
		{name: "assembly ignore constraint", file: "native_arm64.s", goFile: "decl.go", goSource: "package ordinary\n", asmSource: "//go:build ignore\n\nTEXT ·F(SB),$0-0\nRET\n", want: nativeLayoutBuildConstraints},
		{name: "test-only source", file: "native_arm64.s", goFile: "decl_test.go", goSource: "package ordinary\n", want: discoverySourceNotApplicableNoGoPackage},
		{name: "invalid package clause", file: "native_arm64.s", goFile: "decl.go", goSource: "not Go source\n", want: discoverySourceNotApplicableNoGoPackage},
		{name: "ignored testdata directory", file: "testdata/x/native_arm64.s", goFile: "testdata/x/decl.go", goSource: "package ordinary\n", want: nativeLayoutIgnoredDirectory, boundary: "testdata"},
		{name: "nested module boundary", file: "nested/native_arm64.s", goFile: "nested/decl.go", goSource: "package ordinary\n", want: nativeLayoutNestedModule, boundary: "nested/go.mod"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			sources := map[string]string{tc.file: tc.asmSource, tc.goFile: tc.goSource}
			if tc.boundary == "nested/go.mod" {
				sources[tc.boundary] = "module example.com/nested\n"
			}
			plan := fixtureOrdinarySelection(t, []string{tc.file}, []string{"linux/arm64"}, sources)
			if len(plan.Decisions) != 1 || plan.Decisions[0].Kind != tc.want || plan.Decisions[0].Boundary != tc.boundary {
				t.Fatalf("decisions = %+v, want %s boundary %s", plan.Decisions, tc.want, tc.boundary)
			}
			result := discoveryCorpusResult{
				Module: plan.Module, Version: plan.Version, Status: discoveryStatusNotApplicable,
				DiscoveredAsmFiles: []string{tc.file}, OrdinarySelectionPlan: plan,
			}
			if tc.want == nativeLayoutSelected {
				result.SourceNotApplicableItems = []discoverySourceNotApplicableItem{{
					AsmFile: tc.file, Targets: plan.Targets, Kind: discoverySourceNotApplicableGoBuild,
					Reason: "decl.go:3: undefined: missingDeclaration",
				}}
			} else if tc.want == discoverySourceNotApplicableNoGoPackage {
				result.SourceNotApplicableItems = []discoverySourceNotApplicableItem{{
					AsmFile: tc.file, Targets: plan.Targets, Kind: discoverySourceNotApplicableNoGoPackage,
					Reason: plan.Decisions[0].Diagnostic,
				}}
			}
			if err := validateOrdinarySelectionResult(result, plan.Targets, plan.GoVersion); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestOrdinarySelectionRejectsMissingOrForgedScopes(t *testing.T) {
	files := []string{"native_arm64.s"}
	plan := fixtureOrdinarySelection(t, files, []string{"linux/arm64"}, map[string]string{
		"native_arm64.s": "TEXT ·F(SB),$0-0\nRET\n", "decl.go": "package ordinary\n",
	})
	result := discoveryCorpusResult{Module: plan.Module, Version: plan.Version, Status: discoveryStatusNotApplicable,
		DiscoveredAsmFiles: files, OrdinarySelectionPlan: plan,
		SourceNotApplicableItems: []discoverySourceNotApplicableItem{{AsmFile: files[0], Targets: plan.Targets,
			Kind: discoverySourceNotApplicableGoBuild, Reason: "decl.go:2: undefined: missingDeclaration"}},
	}
	for _, tc := range []struct {
		name string
		edit func(*discoveryCorpusResult)
	}{
		{"missing proof", func(r *discoveryCorpusResult) { r.OrdinarySelectionPlan = nil }},
		{"missing exclusion evidence", func(r *discoveryCorpusResult) { r.SourceNotApplicableItems = nil }},
		{"missing directory", func(r *discoveryCorpusResult) { r.OrdinarySelectionPlan.Directories = nil }},
		{"missing source", func(r *discoveryCorpusResult) { r.OrdinarySelectionPlan.Sources = r.OrdinarySelectionPlan.Sources[1:] }},
		{"missing decision", func(r *discoveryCorpusResult) { r.OrdinarySelectionPlan.Decisions = nil }},
		{"forged header", func(r *discoveryCorpusResult) {
			r.OrdinarySelectionPlan.Sources[0].Header = "//go:build ignore\n\npackage ordinary\n"
		}},
		{"virtual root ignored", func(r *discoveryCorpusResult) {
			r.OrdinarySelectionPlan.Decisions[0].Kind = nativeLayoutIgnoredDirectory
			r.OrdinarySelectionPlan.Decisions[0].Boundary = "."
		}},
		{"wrong target", func(r *discoveryCorpusResult) { r.SourceNotApplicableItems[0].Targets = []string{"windows/arm64"} }},
		{"wrong tags", func(r *discoveryCorpusResult) { r.SourceNotApplicableItems[0].BuildTags = []string{"other"} }},
		{"duplicate evidence", func(r *discoveryCorpusResult) {
			r.SourceNotApplicableItems = append(r.SourceNotApplicableItems, r.SourceNotApplicableItems[0])
		}},
		{"infrastructure", func(r *discoveryCorpusResult) {
			r.SourceNotApplicableItems[0].Reason = "checksum mismatch: SECURITY ERROR"
		}},
		{"unknown source reason kind", func(r *discoveryCorpusResult) {
			r.SourceNotApplicableItems[0].Kind = "unsupported_instruction"
		}},
		{"changed Go release constraints", func(r *discoveryCorpusResult) { r.OrdinarySelectionPlan.ReleaseTags = []string{"go1.1"} }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			data, _ := json.Marshal(result)
			var changed discoveryCorpusResult
			if err := json.Unmarshal(data, &changed); err != nil {
				t.Fatal(err)
			}
			tc.edit(&changed)
			if err := validateOrdinarySelectionResult(changed, plan.Targets, plan.GoVersion); err == nil {
				t.Fatal("accepted forged or missing source-selection evidence")
			}
		})
	}
}
