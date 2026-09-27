package main

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestDiscoveryMixedAssemblerAndInfrastructureDiagnostics(t *testing.T) {
	const source = "pkg/file.s:12: unexpected EOF\nasm: assembly of pkg/file.s failed"
	for _, diagnostic := range []struct {
		name      string
		message   string
		retryable bool
	}{
		{"proxy", "reading https://example.com/pkg: 503 Service Unavailable", true},
		{"network EOF", "Get https://example.com/pkg: unexpected EOF", true},
		{"assembly URL EOF", "Get https://example.com/file.s:12: unexpected EOF", true},
		{"disk", "write object.o: no space left on device", false},
		{"killed", "go build: signal: killed", false},
		{"toolchain", "compile: version does not match go tool version", false},
		{"checksum", "verifying example.com/pkg: checksum mismatch\nSECURITY ERROR", false},
	} {
		t.Run(diagnostic.name, func(t *testing.T) {
			for _, combined := range []string{
				source + "\n" + diagnostic.message,
				diagnostic.message + "\n" + source,
			} {
				if !isDiscoveryGoBuildInfrastructureFailure(combined) {
					t.Fatalf("mixed failure was classified as source N/A: %s", combined)
				}
				if got := isDiscoveryRetryableNetworkFailure(combined); got != diagnostic.retryable {
					t.Fatalf("retryable=%v, want %v: %s", got, diagnostic.retryable, combined)
				}
			}
		})
	}
}

func TestDiscoveryAssemblerEOFRequiresMatchingSourceDiagnostic(t *testing.T) {
	for _, filename := range []string{"pkg/file.s", "pkg with spaces/file.s", `C:\source\file.s`} {
		for _, newline := range []string{"\n", "\r\n"} {
			for _, position := range []string{"12", "12:3"} {
				diagnostic := filename + ":" + position + ": unexpected EOF" + newline +
					"asm: assembly of " + filename + " failed" + newline
				if isDiscoveryGoBuildInfrastructureFailure(diagnostic) || isDiscoveryRetryableNetworkFailure(diagnostic) {
					t.Fatalf("deterministic assembler EOF must remain source N/A: %q", diagnostic)
				}
			}
		}
	}
	for _, diagnostic := range []string{
		"pkg/file.s:12: unexpected EOF\nasm: assembly of other.s failed",
		"pkg/File.s:12: unexpected EOF\nasm: assembly of pkg/file.s failed",
		"pkg/file.s:invalid: unexpected EOF\nasm: assembly of pkg/file.s failed",
		"pkg/file.s:12:3:4: unexpected EOF\nasm: assembly of pkg/file.s failed",
		"pkg/file.s:12: unexpected EOF while fetching dependency\nasm: assembly of pkg/file.s failed",
		"pkg/file.s:12: unexpected EOF",
	} {
		if !isDiscoveryGoBuildInfrastructureFailure(diagnostic) {
			t.Fatalf("ambiguous EOF diagnostic was treated as source rejection: %q", diagnostic)
		}
	}
}

func TestDiscoveryFatalFailurePrecedesTransientNetworkDiagnostic(t *testing.T) {
	for _, fatal := range []string{
		"verifying example.com/pkg: checksum mismatch\nSECURITY ERROR",
		"write object.o: no space left on device",
		"go build: signal: killed",
		"compile: version does not match go tool version",
		"go build: captured output exceeds 8388608 bytes",
		"go build: context deadline exceeded",
		"go build: context canceled",
		"exec: go: executable file not found in PATH",
		"fork/exec go: exec format error",
		"fork/exec go: permission denied",
		"fork/exec compile: resource temporarily unavailable",
		"fatal error: runtime: out of memory",
	} {
		if !isDiscoveryGoBuildInfrastructureFailure(fatal) {
			t.Errorf("infrastructure error was classified as source N/A: %s", fatal)
		}
		diagnostic := fatal + "\nreading https://example.com/pkg: 503 Service Unavailable"
		if !isDiscoveryGoBuildInfrastructureFailure(diagnostic) {
			t.Fatalf("fatal error was classified as source N/A: %s", diagnostic)
		}
		if isDiscoveryRetryableNetworkFailure(diagnostic) {
			t.Errorf("fatal error was retried because another dependency had a network failure: %s", diagnostic)
		}
	}
}

func TestDiscoveryChecksumFailureCannotBecomeSourceNotApplicable(t *testing.T) {
	diagnostic := "verifying example.com/pkg: checksum mismatch\nSECURITY ERROR"
	if !isDiscoveryGoBuildInfrastructureFailure(diagnostic) {
		t.Fatal("checksum mismatch was classified as source incompatibility")
	}
	if isDiscoveryRetryableNetworkFailure(diagnostic) {
		t.Fatal("checksum mismatch must not be retried as a transient network error")
	}
}

func TestDiscoveryRetryMixedAssemblerNetworkEOF(t *testing.T) {
	const source = "pkg/file.s:12: unexpected EOF\nasm: assembly of pkg/file.s failed"
	attempts := 0
	err := retryDiscoveryGoNetwork(context.Background(), []time.Duration{0}, func() error {
		attempts++
		if attempts == 1 {
			return errors.New(source + "\nGet https://example.com/pkg: unexpected EOF")
		}
		return errors.New(source)
	})
	if attempts != 2 || err == nil || !strings.Contains(err.Error(), source) {
		t.Fatalf("attempts=%d error=%v, want retry followed by source rejection", attempts, err)
	}
}

func TestDiscoveryActualAssemblerEOFRemainsSourceRejection(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, filepath.Join(dir, "go.mod"), "module example.com/truncatedasm\n\ngo 1.20\n")
	writeTestFile(t, filepath.Join(dir, "decl.go"), "package truncatedasm\n\nfunc f()\n")
	// Deliberately omit the final newline to exercise the real Go diagnostic.
	writeTestFile(t, filepath.Join(dir, "decl_amd64.s"), "TEXT ·f(SB), $0-0\nRET")
	env := replaceEnv(os.Environ(), map[string]string{"GOFLAGS": "-mod=mod", "GOWORK": "off"})
	err := runDiscoveryGoBuild(context.Background(), dir, env, "linux/amd64", nil, "example.com/truncatedasm")
	if err == nil || !strings.Contains(err.Error(), "unexpected EOF") {
		t.Fatalf("expected an actual assembler EOF, got %v", err)
	}
	if isDiscoveryInfrastructureFailure(err) || isDiscoveryRetryableNetworkFailure(err.Error()) {
		t.Fatalf("actual Go source rejection was classified as infrastructure failure: %v", err)
	}
}

func TestDiscoveryProgressRejectsInfrastructureSourceNotApplicable(t *testing.T) {
	for _, status := range []string{discoveryStatusPassed, discoveryStatusNotApplicable} {
		t.Run(status, func(t *testing.T) {
			ledger, reports, source := writeDiscoveryReportFixture(t)
			files, err := discoveryCorpusReportFiles(reports)
			if err != nil {
				t.Fatal(err)
			}
			for _, filename := range files {
				report, err := readDiscoveryCorpusReport(filename)
				if err != nil {
					t.Fatal(err)
				}
				if len(report.Results) == 0 {
					continue
				}
				result := &report.Results[0]
				result.SourceNotApplicableItems = []discoverySourceNotApplicableItem{{
					AsmFiles: result.DiscoveredAsmFiles,
					Targets:  []string{"linux/arm64"},
					Kind:     discoverySourceNotApplicableGoBuild,
					Reason: "pkg/file.s:12: unexpected EOF\nasm: assembly of pkg/file.s failed\n" +
						"reading https://example.com/pkg: 503 Service Unavailable",
				}}
				if status == discoveryStatusNotApplicable {
					result.Status = status
					result.NotApplicableReason = "current Go package rejected"
					report.Passed--
					report.NotApplicable++
					report.Translations -= result.Translations
					result.Translations = 0
				}
				if err := writeDiscoveryCorpusReport(filename, report); err != nil {
					t.Fatal(err)
				}
				targets := []string{"linux/amd64", "linux/arm64"}
				if _, err := collectDiscoveryProgress(ledger, reports, targets, source, 2); err == nil ||
					!strings.Contains(err.Error(), "infrastructure failure") {
					t.Fatalf("progress must reject infrastructure N/A before ledger publication: %v", err)
				}
				if err := verifyDiscoveryCorpusReports(ledger, reports, targets, source); err == nil ||
					!strings.Contains(err.Error(), "infrastructure failure") {
					t.Fatalf("final gate accepted infrastructure N/A: %v", err)
				}
				return
			}
			t.Fatal("fixture contains no candidate")
		})
	}
}

func TestDiscoveryHTTPResponseEOFIsRetryable(t *testing.T) {
	for _, method := range []string{
		http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut,
		http.MethodDelete, http.MethodPatch, http.MethodOptions,
		http.MethodConnect, http.MethodTrace,
	} {
		for _, scheme := range []string{"http", "https"} {
			failure := &url.Error{Op: method, URL: scheme + "://example.com/sumdb/supported", Err: io.EOF}
			for _, diagnostic := range []string{
				failure.Error(),
				"verifying go.mod: initializing sumdb.Client: checking tree#1: " + failure.Error(),
				failure.Error() + "\r\n",
			} {
				if !isDiscoveryGoBuildInfrastructureFailure(diagnostic) || !isDiscoveryRetryableNetworkFailure(diagnostic) {
					t.Errorf("HTTP response EOF must be an infrastructure retry: %s", diagnostic)
				}
			}
			terminal := "checksum mismatch\n" + failure.Error()
			if isDiscoveryRetryableNetworkFailure(terminal) {
				t.Errorf("HTTP EOF must not override checksum failure: %s", terminal)
			}
		}
	}
	for _, diagnostic := range []string{
		"source.s:12: EOF",
		"EOF",
		`Get "file:///source.s": EOF`,
		`Get "https://example.com/module": EOF in source`,
	} {
		if isDiscoveryRetryableNetworkFailure(diagnostic) {
			t.Errorf("non-HTTP EOF was retried as a network failure: %s", diagnostic)
		}
	}
}
