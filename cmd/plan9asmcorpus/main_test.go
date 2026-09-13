package main

import (
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestRepositoryManifestTracksReportedAssemblyFailures(t *testing.T) {
	manifest, err := loadManifest(filepath.Join("..", "..", "testdata", "corpus", "reported-libraries.json"))
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{
		"coder-websocket":    "https://github.com/xgo-dev/llgo/issues/2464",
		"klauspost-compress": "https://github.com/xgo-dev/llgo/issues/2552",
		"go-hex":             "https://github.com/xgo-dev/llgo/issues/2576",
	}
	got := make(map[string]string, len(manifest.Libraries))
	for _, library := range manifest.Libraries {
		got[library.ID] = strings.Join(library.Issues, ",")
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("reported library issues = %#v, want %#v", got, want)
	}
	wantTargets := []string{
		"darwin/amd64",
		"darwin/arm64",
		"linux/386",
		"linux/amd64",
		"linux/arm",
		"linux/arm64",
		"windows/386",
		"windows/amd64",
		"windows/arm64",
		"js/wasm",
		"wasip1/wasm",
	}
	if !reflect.DeepEqual(manifest.Targets, wantTargets) {
		t.Fatalf("reported library targets = %#v, want %#v", manifest.Targets, wantTargets)
	}
}

func TestValidateReportRejectsSilentlySkippedTarget(t *testing.T) {
	library := libraryManifest{
		ID: "example",
		Inventory: map[string]expectedInventory{
			"linux/amd64": {AsmFiles: 1, Packages: []string{"example.com/lib"}},
		},
	}
	report := matrixReport{
		Targets:      []targetReport{{Goos: "linux", Goarch: "amd64", TotalPkgs: 1, AsmPackages: []string{"example.com/lib"}, TotalAsm: 1, Success: 1}},
		TotalTargets: 1,
		TotalAsm:     1,
		Success:      1,
	}
	err := validateReport([]string{"linux/amd64", "linux/arm64"}, library, report)
	if err == nil || !strings.Contains(err.Error(), "target coverage changed") {
		t.Fatalf("validateReport() error = %v, want target coverage failure", err)
	}
}

func TestValidateReportRejectsInventoryDrift(t *testing.T) {
	library := libraryManifest{
		ID: "example",
		Inventory: map[string]expectedInventory{
			"linux/amd64": {AsmFiles: 1, Packages: []string{"example.com/lib"}},
		},
	}
	report := matrixReport{
		Targets:      []targetReport{{Goos: "linux", Goarch: "amd64", TotalPkgs: 1, AsmPackages: []string{"example.com/lib/newasm"}, TotalAsm: 1, Success: 1}},
		TotalTargets: 1,
		TotalAsm:     1,
		Success:      1,
	}
	err := validateReport([]string{"linux/amd64"}, library, report)
	if err == nil || !strings.Contains(err.Error(), "assembly inventory changed") {
		t.Fatalf("validateReport() error = %v, want inventory failure", err)
	}
}
