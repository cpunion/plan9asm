package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
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
