package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExtractSupportedOpsFindsIsolatedCompleteFloatMinMaxFamily(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("../..", "amd64_float_min_max_forms.go"))
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "amd64_float_min_max_forms.go"), data, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "translate.go"), []byte("package plan9asm\n"), 0600); err != nil {
		t.Fatal(err)
	}
	for _, arch := range []string{"amd64", "386", "arm", "arm64", "wasm"} {
		supported, err := extractSupportedOps(dir, arch)
		if err != nil {
			t.Fatal(err)
		}
		for _, op := range []string{
			"MINPS", "MINPD", "MINSS", "MINSD", "MAXPS", "MAXPD", "MAXSS", "MAXSD",
			"VMINPS", "VMINPD", "VMINSS", "VMINSD", "VMAXPS", "VMAXPD", "VMAXSS", "VMAXSD",
		} {
			_, found := supported[op]
			if found != (arch == "amd64" || arch == "386") {
				t.Fatal("isolated complete MIN/MAX table lost its architecture boundary", arch, op, found)
			}
		}
	}
}
