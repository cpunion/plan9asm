package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"
)

// The test executable can act as a narrow Git relay. It still runs the real
// Git command; the only injected operation changes a first-party input at the
// second source-status read, after the stored ledger comparison has finished.
func TestMain(m *testing.M) {
	if os.Getenv("PLAN9ASM_FROZEN_GIT_RELAY") == "1" &&
		strings.TrimSuffix(filepath.Base(os.Args[0]), ".exe") == "git" {
		if err := mutateFrozenAggregateFixture(os.Args[1:]); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(97)
		}
		cmd := exec.Command(os.Getenv("PLAN9ASM_FROZEN_REAL_GIT"), os.Args[1:]...)
		cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
		if err := cmd.Run(); err != nil {
			if exit, ok := err.(*exec.ExitError); ok {
				os.Exit(exit.ExitCode())
			}
			fmt.Fprintln(os.Stderr, err)
			os.Exit(98)
		}
		os.Exit(0)
	}
	os.Exit(m.Run())
}

func mutateFrozenAggregateFixture(args []string) error {
	if len(args) < 3 || args[2] != "status" {
		return nil
	}
	counter := os.Getenv("PLAN9ASM_FROZEN_GIT_COUNTER")
	count := 0
	data, err := os.ReadFile(counter)
	if err == nil {
		count, err = strconv.Atoi(string(data))
	} else if os.IsNotExist(err) {
		err = nil
	}
	if err != nil {
		return err
	}
	count++
	if err := os.WriteFile(counter, []byte(strconv.Itoa(count)), 0600); err != nil {
		return err
	}
	if count != 2 {
		return nil
	}
	file, err := os.OpenFile(os.Getenv("PLAN9ASM_FROZEN_MUTATE"), os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		return err
	}
	_, writeErr := file.WriteString("\n")
	closeErr := file.Close()
	if writeErr != nil {
		return writeErr
	}
	return closeErr
}

type frozenAggregateFixture struct {
	root, ledger, reports, stored, manifest string
}

func newFrozenAggregateFixture(t *testing.T) frozenAggregateFixture {
	t.Helper()
	root, source := verifiedProgressSourceFixture(t)
	ledger, reports, _ := writeDiscoveryReportFixture(t)
	targets := []string{"linux/amd64", "linux/arm64"}
	for shard := 0; shard < 2; shard++ {
		filename := filepath.Join(reports, "shard-"+strconv.Itoa(shard)+".json")
		report, err := readDiscoveryCorpusReport(filename)
		if err != nil {
			t.Fatal(err)
		}
		report.Provenance.Source = source
		report.Provenance.TranslatorRevision = source.Revision
		if err := writeDiscoveryCorpusReport(filename, report); err != nil {
			t.Fatal(err)
		}
	}
	progress, err := collectDiscoveryProgress(ledger, reports, targets, source, 2, root)
	if err != nil {
		t.Fatal(err)
	}
	stored := filepath.Join(t.TempDir(), "assembly-ledger")
	if err := writeAssemblyLedger(stored, progress, source.SHA256); err != nil {
		t.Fatal(err)
	}
	manifest := corpusManifest{SchemaVersion: 1, Targets: targets, Libraries: []libraryManifest{{
		ID: "fixture", Origin: originEcosystemScan, Module: "example.com/a", Version: "v1.0.0",
		Inventory: map[string]expectedInventory{"linux/amd64": {AsmFiles: 1, Packages: []string{"example.com/a"}}},
	}}}
	data, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	manifestPath := filepath.Join(t.TempDir(), "manifest.json")
	writeTestFile(t, manifestPath, string(data))
	return frozenAggregateFixture{root, ledger, reports, stored, manifestPath}
}

func (fixture frozenAggregateFixture) args(output string) []string {
	return []string{
		"-manifest=" + fixture.manifest, "-repo-root=" + fixture.root,
		"-discovery-ledger=" + fixture.ledger, "-verify-discovery-reports=" + fixture.reports,
		"-compare-assembly-ledger=" + fixture.stored, "-discovery-progress-output=" + output,
	}
}

func TestDiscoveryVerifiedProgressFrozenInputs(t *testing.T) {
	for _, kind := range []string{"source", "scan ledger", "assembly ledger"} {
		t.Run(kind, func(t *testing.T) {
			fixture := newFrozenAggregateFixture(t)
			mutation := filepath.Join(fixture.root, "source.txt")
			if kind == "scan ledger" {
				mutation = filepath.Join(fixture.ledger, "records", "00.jsonl")
			} else if kind == "assembly ledger" {
				entries, err := os.ReadDir(filepath.Join(fixture.stored, "records"))
				if err != nil || len(entries) == 0 {
					t.Fatalf("stored records: %v", err)
				}
				mutation = filepath.Join(fixture.stored, "records", entries[0].Name())
			}
			git, err := exec.LookPath("git")
			if err != nil {
				t.Fatal(err)
			}
			bin := t.TempDir()
			name := "git"
			if runtime.GOOS == "windows" {
				name += ".exe"
			}
			copyFrozenGitRelay(t, filepath.Join(bin, name))
			counter := filepath.Join(t.TempDir(), "status-count")
			t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
			t.Setenv("PLAN9ASM_FROZEN_GIT_RELAY", "1")
			t.Setenv("PLAN9ASM_FROZEN_REAL_GIT", git)
			t.Setenv("PLAN9ASM_FROZEN_GIT_COUNTER", counter)
			t.Setenv("PLAN9ASM_FROZEN_MUTATE", mutation)
			output := filepath.Join(t.TempDir(), "assembly-progress.json")
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			stdout, stderr, err := runVerifiedProgressCLI(ctx, fixture.root, fixture.args(output)...)
			calls, readErr := os.ReadFile(counter)
			if readErr != nil || string(calls) != "2" {
				t.Fatalf("mutation did not reach final source read: %q %v; %s", calls, readErr, stderr)
			}
			if err == nil || !strings.Contains(string(stderr), "changed") || len(stdout) != 0 {
				t.Fatalf("changed %s accepted: %v stdout=%q stderr=%s", kind, err, stdout, stderr)
			}
			if _, err := os.Lstat(output); !os.IsNotExist(err) {
				t.Fatalf("invalidated input published progress: %v", err)
			}
			entries, err := os.ReadDir(filepath.Dir(output))
			if err != nil || len(entries) != 0 {
				t.Fatalf("owned preparation leaked: %v %v", entries, err)
			}
		})
	}
}

func copyFrozenGitRelay(t *testing.T, destination string) {
	t.Helper()
	source, err := os.Open(os.Args[0])
	if err != nil {
		t.Fatal(err)
	}
	output, err := os.OpenFile(destination, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0700)
	if err != nil {
		source.Close()
		t.Fatal(err)
	}
	_, copyErr := io.Copy(output, source)
	closeSource, closeOutput := source.Close(), output.Close()
	if copyErr != nil || closeSource != nil || closeOutput != nil {
		t.Fatalf("copy Git relay: %v %v %v", copyErr, closeSource, closeOutput)
	}
}
