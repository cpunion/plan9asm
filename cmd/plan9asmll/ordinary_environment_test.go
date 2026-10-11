package main

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"golang.org/x/tools/go/packages"
)

func TestOrdinaryPackageEnvironmentRetainsExplicitSelection(t *testing.T) {
	input := append(os.Environ(), "GOENV=off", "GOWORK=off", "GOTOOLCHAIN=local",
		"CGO_ENABLED=1", "GOFLAGS=-tags=original", "GOAMD64=v3", "GOPACKAGESDRIVER=custom")
	before := append([]string(nil), input...)
	env, err := ordinaryPackageEnvironment("linux", "amd64", input)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(before, input) {
		t.Fatal("package-load setup mutated its caller's environment")
	}
	values := make(map[string]string)
	for _, item := range env {
		key, value, _ := strings.Cut(item, "=")
		if _, duplicate := values[key]; duplicate {
			t.Fatalf("duplicate effective environment key %s", key)
		}
		values[key] = value
	}
	for key, expected := range map[string]string{
		"GOOS": "linux", "GOARCH": "amd64", "CGO_ENABLED": "1",
		"GOFLAGS": "-tags=original", "GOAMD64": "v3", "GOPACKAGESDRIVER": "custom",
	} {
		if values[key] != expected {
			t.Errorf("explicit %s changed: got %q, want %q", key, values[key], expected)
		}
	}
	for _, key := range []string{"GOCACHE", "GOMODCACHE", "GOPATH", "TMPDIR", "GOTMPDIR"} {
		if !filepath.IsAbs(values[key]) {
			t.Errorf("missing absolute effective %s", key)
		}
	}
}

func TestCallerSelectionRetainsActualLoadEnvironment(t *testing.T) {
	pkg := &packages.Package{Name: "example", PkgPath: "example.invalid/p"}
	env := []string{"GOOS=linux", "GOARCH=amd64", "GOFLAGS=", "GOPACKAGESDRIVER=off"}
	want := append([]string(nil), env...)
	t.Setenv("GOFLAGS", "-overlay=unrelated.json")
	selected := captureX86CallerSelectionsInEnvironment([]*packages.Package{pkg}, "linux/amd64", env)[pkg]
	env[0] = "GOOS=windows"
	if !reflect.DeepEqual(selected.env, want) {
		t.Fatal("selection reconstructed ambient defaults or borrowed a mutable environment")
	}
}
