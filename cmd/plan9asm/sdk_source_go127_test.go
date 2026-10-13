//go:build go1.27

package main

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/xgo-dev/plan9asm"
	"github.com/xgo-dev/plan9asm/gotoolprofile"
)

func setSimpleSDKProfile(t *testing.T) {
	t.Helper()
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
}

func TestTranslateOriginalSDKCRC32NumericalSource(t *testing.T) {
	setSimpleSDKProfile(t)
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

func TestSimpleCLIOriginalSDKCRC32PublicationAndRejection(t *testing.T) {
	setSimpleSDKProfile(t)
	input := filepath.Join(runtime.GOROOT(), "src", "hash", "crc32", "crc32_amd64.s")
	original, err := os.ReadFile(input)
	if err != nil {
		t.Fatal(err)
	}
	for _, mode := range []string{"file", "package"} {
		t.Run(mode, func(t *testing.T) {
			folder := t.TempDir()
			meta := filepath.Join(folder, "metadata.json")
			output := filepath.Join(folder, "crc32.ll")
			args := []string{"-goos=linux", "-goarch=amd64", "-meta=" + meta}
			if mode == "file" {
				args = append(args, "-i="+input, "-o="+output)
			} else {
				args = append(args, "-dir="+folder, "hash/crc32")
				output = filepath.Join(folder, "crc32_amd64.s.ll")
			}
			if err := runTranspile(args); err != nil {
				t.Fatalf("public %s entry point rejected unchanged SDK source: %v", mode, err)
			}
			ir, err := os.ReadFile(output)
			if err != nil || !strings.Contains(string(ir), "hash/crc32.r2r1<>") ||
				!strings.Contains(string(ir), "internal/cpu.X86") || strings.Contains(string(ir), "hash/crc32.r2r1$local") {
				t.Fatalf("public SDK result lost the original native namespace: %v", err)
			}
			raw, err := os.ReadFile(meta)
			if err != nil || !json.Valid(raw) || !strings.Contains(string(raw), `"package": "hash/crc32"`) ||
				!strings.Contains(string(raw), `"resolved": "hash/crc32.ieeeCLMUL"`) {
				t.Fatalf("public SDK metadata is absent/incomplete: %v\n%s", err, raw)
			}
			ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
			defer cancel()
			object := filepath.Join(folder, "crc32.o")
			stdout, stderr, err := gotoolprofile.RunBounded(ctx, folder, os.Environ(), sdkLLVM22LLC(t, ctx),
				"-filetype=obj", "-o", object, output)
			if err != nil || len(stderr) != 0 {
				t.Fatalf("public SDK LLVM22 object: %v\n%s%s", err, stdout, stderr)
			}
			if info, err := os.Stat(object); err != nil || info.Size() == 0 {
				t.Fatal("public SDK object is absent", err)
			}
		})
	}
	for _, row := range []struct{ key, value string }{{"GOAMD64", "v2"}, {"GOFLAGS", "-tags=unobserved_sdk_profile"}} {
		t.Run("reject-"+row.key, func(t *testing.T) {
			t.Setenv(row.key, row.value)
			folder := t.TempDir()
			output, meta := filepath.Join(folder, "rejected.ll"), filepath.Join(folder, "rejected.json")
			const retained = "preexisting output\n"
			for _, name := range []string{output, meta} {
				if err := os.WriteFile(name, []byte(retained), 0600); err != nil {
					t.Fatal(err)
				}
			}
			err := runTranspile([]string{"-goos=linux", "-goarch=amd64", "-i=" + input, "-o=" + output, "-meta=" + meta})
			if !errors.Is(err, plan9asm.ErrProbeNeedsContext) {
				t.Fatalf("unobserved profile was not rejected with its missing source contract: %v", err)
			}
			for _, name := range []string{output, meta} {
				actual, err := os.ReadFile(name)
				if err != nil || string(actual) != retained {
					t.Fatalf("failed source observation published/overwrote output: %v", err)
				}
			}
		})
	}
	after, err := os.ReadFile(input)
	if err != nil || sha256.Sum256(after) != sha256.Sum256(original) {
		t.Fatal("public CLI changed original selected SDK source", err)
	}
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
