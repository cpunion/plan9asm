package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestDiscoveryCapturedCommandOutputIsBounded(t *testing.T) {
	var output boundedDiscoveryCommandOutput

	first := bytes.Repeat([]byte{'a'}, discoveryCommandOutputLimit-1)
	if count, err := output.Write(first); count != len(first) || err != nil {
		t.Fatalf("first write = %d, %v", count, err)
	}
	if output.truncated || output.Len() != len(first) {
		t.Fatalf("first write: length=%d truncated=%v", output.Len(), output.truncated)
	}

	last := []byte{'b', 'c', 'd'}
	if count, err := output.Write(last); count != len(last) || err != nil {
		t.Fatalf("overflow write = %d, %v", count, err)
	}
	if !output.truncated || output.Len() != discoveryCommandOutputLimit {
		t.Fatalf("overflow: length=%d truncated=%v", output.Len(), output.truncated)
	}
	if !bytes.Equal(output.Bytes()[discoveryCommandOutputLimit-2:], []byte{'a', 'b'}) {
		t.Fatal("bounded output did not preserve the exact limit")
	}

	more := bytes.Repeat([]byte{'z'}, discoveryCommandOutputLimit)
	if count, err := output.Write(more); count != len(more) || err != nil {
		t.Fatalf("discarded write = %d, %v", count, err)
	}
	if output.Len() != discoveryCommandOutputLimit {
		t.Fatalf("discarded write grew capture to %d bytes", output.Len())
	}
}

func TestDiscoveryAsmDeclOutputOverflowIsNotIgnored(t *testing.T) {
	dir := buildDiscoveryOverflowGoTool(t, false)
	err := runDiscoveryAsmDecl(context.Background(), dir, os.Environ(), "linux/amd64", nil, []string{"example.com/pkg"})
	if err == nil || !strings.Contains(err.Error(), "captured output exceeds") {
		t.Fatalf("go vet output overflow was ignored: %v", err)
	}
	if !errors.Is(err, errDiscoveryCommandOutputExceeded) {
		t.Fatalf("go vet output overflow lost its sentinel: %v", err)
	}
}

func TestDiscoveryAsmDeclRetryOutputOverflowIsNotIgnored(t *testing.T) {
	dir := buildDiscoveryOverflowGoTool(t, true)
	err := runDiscoveryAsmDecl(context.Background(), dir, os.Environ(), "linux/amd64", nil, []string{"example.com/pkg"})
	if err == nil || !errors.Is(err, errDiscoveryCommandOutputExceeded) {
		t.Fatalf("go list retry output overflow was ignored: %v", err)
	}
}

func TestDiscoveryAsmDeclRetainsEarlyABIErrorInLargeDiagnostic(t *testing.T) {
	for _, test := range []struct {
		name       string
		diagnostic string
	}{
		{
			name:       "ABI mismatch",
			diagnostic: "source_amd64.s:10:1: [amd64] source: wrong argument size 16; expected $...-8",
		},
		{
			name:       "infrastructure failure",
			diagnostic: "go: write build cache: no space left on device",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			dir := t.TempDir()
			writeTestFile(t, filepath.Join(dir, "go.mod"), "module example.com/pkg\n\ngo 1.20\n")
			toolDir := t.TempDir()
			source := filepath.Join(toolDir, "go.go")
			program := `package main

import (
	"bytes"
	"os"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "vet" {
		_, _ = os.Stderr.WriteString(` + fmt.Sprintf("%q", test.diagnostic+"\n") + `)
		_, _ = os.Stderr.Write(bytes.Repeat([]byte{'x'}, 70<<10))
		os.Exit(1)
	}
}
`
			writeTestFile(t, source, program)
			name := "go"
			if runtime.GOOS == "windows" {
				name += ".exe"
			}
			command := exec.Command("go", "build", "-o", filepath.Join(toolDir, name), source)
			if output, err := command.CombinedOutput(); err != nil {
				t.Fatalf("build fake Go tool: %v\n%s", err, output)
			}
			t.Setenv("PATH", toolDir+string(os.PathListSeparator)+os.Getenv("PATH"))

			err := runDiscoveryAsmDecl(context.Background(), dir, os.Environ(), "linux/amd64", nil, []string{"example.com/pkg"})
			if err == nil {
				t.Fatal("early diagnostic disappeared after display truncation")
			}
			if !strings.Contains(discoveryCommandDiagnostic(err), test.diagnostic) {
				t.Fatalf("full command diagnostic lost %q", test.diagnostic)
			}
			if strings.Contains(err.Error(), test.diagnostic) {
				t.Fatal("display error unexpectedly retained the large diagnostic prefix")
			}
			if test.name == "ABI mismatch" {
				rejected := discoveryAsmDeclRejectedFiles(
					[]string{"source_amd64.s", "other_amd64.s"}, discoveryCommandDiagnostic(err),
				)
				if len(rejected) != 1 || rejected[0] != "source_amd64.s" {
					t.Fatalf("ABI rejection affected wrong files: %v", rejected)
				}
				if evidence := discoveryAsmDeclRejectionEvidence(err, rejected); !strings.Contains(evidence, test.diagnostic) {
					t.Fatalf("persisted ABI evidence lost the early mismatch: %q", evidence)
				}
			} else if !isDiscoveryInfrastructureFailure(err) {
				t.Fatal("early resource failure was classified as source incompatibility")
			}
		})
	}
}

func buildDiscoveryOverflowGoTool(t *testing.T, vetFails bool) string {
	t.Helper()
	dir := t.TempDir()
	source := filepath.Join(dir, "overflow.go")
	program := `package main

import (
	"bytes"
	"os"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "vet" && ` + fmt.Sprint(vetFails) + ` {
		_, _ = os.Stderr.WriteString("package tests do not compile\n")
		os.Exit(1)
	}
	_, _ = os.Stdout.Write(bytes.Repeat([]byte{'x'}, (8 << 20) + 1))
}
`
	if err := os.WriteFile(source, []byte(program), 0644); err != nil {
		t.Fatal(err)
	}
	name := "go"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	cmd := exec.Command("go", "build", "-o", filepath.Join(dir, name), source)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build fake go tool: %v\n%s", err, output)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	return dir
}

func TestDiscoveryCapturedCommandOutputOverflowFails(t *testing.T) {
	const childEnv = "PLAN9ASM_TEST_DISCOVERY_OUTPUT_CHILD"
	if os.Getenv(childEnv) == "1" {
		chunk := bytes.Repeat([]byte{'x'}, 4096)
		for remaining := discoveryCommandOutputLimit + 1; remaining > 0; {
			count := remaining
			if count > len(chunk) {
				count = len(chunk)
			}
			if _, err := os.Stdout.Write(chunk[:count]); err != nil {
				os.Exit(2)
			}
			remaining -= count
		}
		os.Exit(0)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	env := replaceEnv(os.Environ(), map[string]string{childEnv: "1"})
	output, err := runCapturedCommandOutput(ctx, "", env, os.Args[0], "-test.run=^TestDiscoveryCapturedCommandOutputOverflowFails$")
	if err == nil || !strings.Contains(err.Error(), "captured output exceeds") {
		t.Fatalf("overflowed successful command must fail explicitly: output=%d error=%v", len(output), err)
	}
}
