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

func TestProfileSelectionActualCPUDoesNotBecomeCustomTag(t *testing.T) {
	goBinary, err := exec.LookPath("go")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	moduleDir := t.TempDir()
	writeTestFile(t, moduleDir+"/vector_amd64.s", "//go:build amd64.v3 && !amd64.v4\n\nTEXT ·Vector(SB),$0-0\nRET\n")
	writeTestFile(t, moduleDir+"/vector_amd64.go", "//go:build amd64.v3 && !amd64.v4\n\npackage vector\n\nfunc Vector()\n")
	candidate := discoveryCandidate{Module: "example.invalid/vector", Version: "v1.0.0", AsmFiles: []string{"vector_amd64.s"}}
	plan, err := captureOrdinarySelectionInputs(candidate, moduleDir, []string{"linux/amd64"})
	if err != nil {
		t.Fatal(err)
	}
	profiles, err := captureDiscoveryFeatureProfiles(ctx, goBinary, t.TempDir(), targetFeatureTestEnv(), plan, candidate.AsmFiles)
	if err != nil {
		t.Fatal(err)
	}
	decisions, scopes, err := replayOrdinarySelectionForProfiles(plan, candidate.AsmFiles, profiles)
	if err != nil {
		t.Fatal(err)
	}
	if len(scopes) != 1 || len(decisions) != 2 {
		t.Fatalf("CPU-only source scope was lost or faked with custom tags: scopes=%+v decisions=%+v", scopes, decisions)
	}
	for _, profile := range profiles {
		for scope := range scopes {
			if strings.Contains(scope.Tags, "amd64.") || scope.ProfileID == profile.ID && profile.Request.Baseline {
				t.Fatalf("baseline scalar context claimed feature-only assembly: %+v", scope)
			}
		}
	}
	if _, _, err := replayOrdinarySelectionForProfiles(plan, candidate.AsmFiles, profiles[:1]); err == nil {
		t.Fatal("accepted baseline-only source replay with missing required CPU profile")
	}
}

func TestProfileSelectionRootDirectoryAndCustomTagsRemainEligible(t *testing.T) {
	goBinary, err := exec.LookPath("go")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	moduleDir := t.TempDir()
	sources := map[string]string{
		"normal.s":            "//go:build accelerate\n\nTEXT ·Normal(SB),$0-0\nRET\n",
		"normal.go":           "package normal\n",
		"_ignored.s":          "TEXT ·Ignored(SB),$0-0\nRET\n",
		"testdata/example.s":  "TEXT ·Fixture(SB),$0-0\nRET\n",
		"testdata/example.go": "package fixture\n",
		"nested/go.mod":       "module example.invalid/nested\n\ngo 1.20\n",
		"nested/example.s":    "TEXT ·Nested(SB),$0-0\nRET\n",
		"nested/example.go":   "package nested\n",
	}
	for file, source := range sources {
		name := filepath.Join(moduleDir, filepath.FromSlash(file))
		if err := os.MkdirAll(filepath.Dir(name), 0700); err != nil {
			t.Fatal(err)
		}
		writeTestFile(t, name, source)
	}
	candidate := discoveryCandidate{Module: "example.invalid/source", Version: "v1.0.0", AsmFiles: []string{"normal.s", "_ignored.s", "testdata/example.s", "nested/example.s"}}
	plan, err := captureOrdinarySelectionInputs(candidate, moduleDir, []string{"linux/amd64"})
	if err != nil {
		t.Fatal(err)
	}
	profiles, err := captureDiscoveryFeatureProfiles(ctx, goBinary, t.TempDir(), targetFeatureTestEnv(), plan, candidate.AsmFiles)
	if err != nil {
		t.Fatal(err)
	}
	decisions, scopes, err := replayOrdinarySelectionForProfiles(plan, candidate.AsmFiles, profiles)
	if err != nil {
		t.Fatal(err)
	}
	if len(scopes) != 1 {
		t.Fatalf("root directory or ordinary custom tag coverage was lost: %+v", scopes)
	}
	for scope := range scopes {
		if scope.File != "normal.s" || scope.Tags != "accelerate" || scope.ProfileID != profiles[0].ID {
			t.Fatalf("wrong actual source/custom-tag scope: %+v", scope)
		}
	}
	seen := make(map[string]bool)
	for _, decision := range decisions {
		seen[decision.Kind] = true
	}
	for _, kind := range []string{nativeLayoutSelected, ordinarySelectionIgnoredFilename, nativeLayoutIgnoredDirectory, nativeLayoutNestedModule} {
		if !seen[kind] {
			t.Errorf("lost explicit filename/directory/nested policy evidence: %s", kind)
		}
	}
}

func TestProfileSelectionCustomTagsIncludeSharedPackageAssembly(t *testing.T) {
	binary, err := exec.LookPath("go")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	root := t.TempDir()
	sources := map[string]string{
		"go.mod":                    "module example.invalid/tag-scopes\n\ngo 1.20\n",
		"pkg/decl.go":               "package fixture\nfunc Shared()\nfunc Variant()\n",
		"pkg/shared_arm64.s":        "TEXT ·Shared(SB),$0-0\nRET\n",
		"pkg/basic_arm64.s":         "//go:build arm64 && !lse2\n\nTEXT ·Variant(SB),$0-0\nRET\n",
		"pkg/lse_arm64.s":           "//go:build arm64 && lse2\n\nTEXT ·Variant(SB),$0-0\nRET\n",
		"other/decl.go":             "package other\nfunc Independent()\n",
		"other/independent_arm64.s": "TEXT ·Independent(SB),$0-0\nRET\n",
	}
	var asmFiles []string
	for file, source := range sources {
		name := filepath.Join(root, filepath.FromSlash(file))
		if err := os.MkdirAll(filepath.Dir(name), 0700); err != nil {
			t.Fatal(err)
		}
		writeTestFile(t, name, source)
		if strings.HasSuffix(file, ".s") {
			asmFiles = append(asmFiles, file)
		}
	}
	asmFiles = uniqueSortedDiscoveryStrings(asmFiles)
	candidate := discoveryCandidate{Module: "example.invalid/tag-scopes", Version: "v1.0.0", AsmFiles: asmFiles}
	plan, err := captureOrdinarySelectionInputs(candidate, root, []string{"linux/arm64"})
	if err != nil {
		t.Fatal(err)
	}
	profiles, err := captureDiscoveryFeatureProfiles(ctx, binary, t.TempDir(), targetFeatureTestEnv(), plan, asmFiles)
	if err != nil {
		t.Fatal(err)
	}
	decisions, scopes, err := replayOrdinarySelectionForProfiles(plan, asmFiles, profiles)
	if err != nil {
		t.Fatal(err)
	}
	if len(profiles) != 1 {
		t.Fatalf("ordinary custom tag invented CPU profiles: %d", len(profiles))
	}
	var actual []string
	for scope := range scopes {
		if scope.Target != "linux/arm64" || scope.ProfileID != profiles[0].ID {
			t.Fatalf("scope lost its actual target/profile: %+v", scope)
		}
		actual = append(actual, scope.File+"#"+scope.Tags)
	}
	want := []string{
		"other/independent_arm64.s#",
		"pkg/basic_arm64.s#",
		"pkg/lse_arm64.s#lse2",
		"pkg/shared_arm64.s#",
		"pkg/shared_arm64.s#lse2",
	}
	actual = uniqueSortedDiscoveryStrings(actual)
	if !reflect.DeepEqual(actual, want) {
		t.Fatalf("package/tag scope closure = %v, want %v", actual, want)
	}
	selected := make(map[discoveryProfileScope]bool)
	for _, decision := range decisions {
		if decision.Kind != nativeLayoutSelected {
			continue
		}
		for _, file := range decision.AsmFiles {
			for _, target := range decision.Targets {
				key := discoveryProfileScope{File: file, Target: target, ProfileID: decision.ProfileID, Tags: strings.Join(decision.BuildTags, "\x00")}
				if selected[key] {
					t.Fatalf("duplicate source selection decision: %+v", key)
				}
				selected[key] = true
			}
		}
	}
	if !reflect.DeepEqual(scopes, selected) {
		t.Fatalf("source decisions do not preserve expanded four-part scopes: %v != %v", selected, scopes)
	}
}
