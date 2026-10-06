package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCPPHeaderPlanningWorkspaceRejectsLateReplacementAndBytes(t *testing.T) {
	for _, mutation := range []string{"bytes", "file", "directory", "alias"} {
		t.Run(mutation, func(t *testing.T) {
			parent := t.TempDir()
			directory := filepath.Join(parent, "query")
			if err := os.Mkdir(directory, 0700); err != nil {
				t.Fatal(err)
			}
			module := "module example.invalid/query\n"
			name := filepath.Join(directory, "go.mod")
			writeTestFile(t, name, module)
			guard, err := captureDiscoveryHeaderQueryWorkspace(directory, module)
			if err != nil || guard.verify() != nil {
				t.Fatalf("unmodified physical workspace control: %v", err)
			}
			switch mutation {
			case "bytes":
				writeTestFile(t, name, module+"// changed\n")
			case "file":
				if err := os.Rename(name, name+".original"); err != nil {
					t.Fatal(err)
				}
				writeTestFile(t, name, module)
			case "directory", "alias":
				if err := os.Rename(directory, directory+".original"); err != nil {
					t.Fatal(err)
				}
				if mutation == "alias" {
					if err := os.Symlink(directory+".original", directory); err != nil {
						t.Fatal(err)
					}
				} else {
					if err := os.Mkdir(directory, 0700); err != nil {
						t.Fatal(err)
					}
					writeTestFile(t, name, module)
				}
			}
			if guard.verify() == nil {
				t.Fatal("late replacement acquired a completed actual query receipt")
			}
		})
	}
}

func TestCPPHeaderPlanningSourceBoundsAndExactRequestKey(t *testing.T) {
	valid := &discoveryCPPHeaderPlanning{Sources: map[string]string{"header.go": "package header\n"}}
	if err := validateDiscoveryPlanningSources(valid); err != nil {
		t.Fatal(err)
	}
	for _, sources := range []map[string]string{
		nil, {"../header.go": "package header\n"}, {"header_test.go": "package header\n"},
		{"header.go": strings.Repeat("x", 2<<20+1)}, {"header.go": "package header\x00"},
	} {
		if validateDiscoveryPlanningSources(&discoveryCPPHeaderPlanning{Sources: sources}) == nil {
			t.Fatal("unbounded or nonordinary source collection acquired planning authority")
		}
	}
	oversized := make(map[string]string)
	for _, name := range []string{"a", "b", "c", "d", "e", "f", "g", "h", "i"} {
		oversized[name+".go"] = strings.Repeat("x", 2<<20)
	}
	if validateDiscoveryPlanningSources(&discoveryCPPHeaderPlanning{Sources: oversized}) == nil {
		t.Fatal("deduplicated original bytes exceeded the complete role bound")
	}
	request := discoveryFeatureProfileRequest{Target: "linux/amd64", Baseline: true}
	key := discoveryHeaderPlanningKey(request, "header.s", nil)
	for _, different := range []string{
		discoveryHeaderPlanningKey(request, "other.s", nil),
		discoveryHeaderPlanningKey(request, "header.s", []string{"private"}),
		discoveryHeaderPlanningKey(discoveryFeatureProfileRequest{Target: "linux/amd64", Overrides: map[string]string{"GOAMD64": "v2"}}, "header.s", nil),
		discoveryHeaderPlanningKey(discoveryFeatureProfileRequest{Target: "linux/amd64", Baseline: true, Expected: map[string]bool{"goexperiment.newinliner": true}}, "header.s", nil),
	} {
		if different == key {
			t.Fatal("planning key discarded a required target/source/tag/request field")
		}
	}
}

func TestCPPHeaderPlanningLanguageRetainsCompleteOriginalModuleBytes(t *testing.T) {
	for _, test := range []struct{ directive, expected string }{
		{"", "go1.16"}, {"go 1.20\n", "go1.20"}, {"go 1.27.1\n", "go1.27"},
	} {
		contents := "module example.invalid/header\n" + test.directive
		plan := &discoveryOrdinarySelectionPlan{Sources: []ordinarySelectionSource{{
			File: "go.mod", Header: contents, SHA256: discoveryFeatureBytesSHA256([]byte(contents)),
		}}}
		language, err := discoveryHeaderPlanningLanguage(plan)
		if err != nil || language != test.expected {
			t.Fatalf("exact original module language: got=%s want=%s error=%v", language, test.expected, err)
		}
		plan.Sources[0].Header += "// changed\n"
		if _, err := discoveryHeaderPlanningLanguage(plan); err == nil {
			t.Fatal("partial/changed module text supplied a planning language")
		}
	}
}
