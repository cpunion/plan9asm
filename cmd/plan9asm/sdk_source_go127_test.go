//go:build go1.27

package main

import (
	"context"
	"crypto/sha256"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/xgo-dev/plan9asm/gotoolprofile"
)

func TestTranslateOriginalSDKCRC32NumericalSource(t *testing.T) {
	for key, value := range map[string]string{
		"GOENV": "off", "GOWORK": "off", "GOTOOLCHAIN": "local",
		"GOFLAGS": "", "GOEXPERIMENT": "", "GOPACKAGESDRIVER": "off",
		"GOAMD64": "v1", "GO386": "sse2",
	} {
		t.Setenv(key, value)
	}
	for _, key := range []string{"CGO_ENABLED", "GOPATH", "GOMODCACHE", "GOTMPDIR"} {
		t.Setenv(key, "")
		if err := os.Unsetenv(key); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("GOCACHE", t.TempDir())
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	input := filepath.Join(runtime.GOROOT(), "src", "hash", "crc32", "crc32_amd64.s")
	original, err := os.ReadFile(input)
	if err != nil || !strings.Contains(string(original), "CMPB internal∕cpu·X86+") {
		t.Fatal("actual SDK CRC32 counterexample is absent", err)
	}
	pkg, err := loadPackage("hash/crc32", "linux", "amd64")
	if err != nil {
		t.Fatal(err)
	}
	tr, ok, err := translateAsmForPackage(pkg.Package, input, "linux", "amd64", false, pkg.selected)
	if err != nil || !ok {
		t.Fatalf("unchanged original SDK translation: %v (TEXT=%t)", err, ok)
	}
	llc := sdkLLVM22LLC(t, ctx)
	folder := t.TempDir()
	ir, object := filepath.Join(folder, "crc32.ll"), filepath.Join(folder, "crc32.o")
	if err := os.WriteFile(ir, []byte(tr.LLVMIR), 0600); err != nil {
		t.Fatal(err)
	}
	stdout, stderr, err := gotoolprofile.RunBounded(ctx, folder, os.Environ(), llc, "-filetype=obj", "-o", object, ir)
	if err != nil || len(stderr) != 0 {
		t.Fatalf("actual SDK LLVM22 object: %v\n%s%s", err, stdout, stderr)
	}
	info, err := os.Stat(object)
	if err != nil || info.Size() == 0 {
		t.Fatal("LLVM22 object is absent", err)
	}
	after, err := os.ReadFile(input)
	if err != nil || sha256.Sum256(after) != sha256.Sum256(original) {
		t.Fatal("selected original SDK bytes changed", err)
	}
	t.Logf("original SDK SHA256=%x; LLVM22 object bytes=%d", sha256.Sum256(original), info.Size())
}

func sdkLLVM22LLC(t *testing.T, ctx context.Context) string {
	t.Helper()
	candidates := []string{"llc-22", "llc"}
	if config := os.Getenv("LLVM_CONFIG"); config != "" {
		bin, stderr, err := gotoolprofile.RunBounded(ctx, "", os.Environ(), config, "--bindir")
		if err != nil || len(stderr) != 0 {
			t.Fatalf("selected LLVM_CONFIG: %v\n%s%s", err, bin, stderr)
		}
		candidates = []string{filepath.Join(strings.TrimSpace(string(bin)), "llc")}
	}
	for _, candidate := range candidates {
		llc, err := exec.LookPath(candidate)
		if err != nil {
			continue
		}
		version, stderr, err := gotoolprofile.RunBounded(ctx, "", os.Environ(), llc, "--version")
		if err == nil && len(stderr) == 0 && strings.Contains(string(version), "LLVM version 22.") {
			return llc
		}
	}
	t.Fatal("required LLVM22 llc is absent")
	return ""
}
