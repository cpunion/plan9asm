package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// These are first-party reader fixtures, not external-library pass receipts.
// The subprocess calls the real CLI, including its flag and source guards.
func TestDiscoveryVerifiedProgressRealCLI(t *testing.T) {
	for _, name := range []string{"complete", "missing shard", "failed candidate", "ledger mismatch", "trailing report data"} {
		t.Run(name, func(t *testing.T) {
			root, source := verifiedProgressSourceFixture(t)
			ledger, reports, _ := writeDiscoveryReportFixture(t)
			targets := []string{"linux/amd64", "linux/arm64"}
			for shard := 0; shard < 2; shard++ {
				filename := filepath.Join(reports, "shard-"+string(rune('0'+shard))+".json")
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
			filename := filepath.Join(reports, "shard-1.json")
			switch name {
			case "missing shard":
				if err := os.Remove(filename); err != nil {
					t.Fatal(err)
				}
			case "failed candidate", "ledger mismatch":
				report, err := readDiscoveryCorpusReport(filename)
				if err != nil || len(report.Results) == 0 {
					t.Fatalf("candidate fixture: %v", err)
				}
				if name == "failed candidate" {
					report.Results[0].Status = discoveryStatusFailed
					report.Results[0].Error = "first-party translation failure"
					report.Passed--
					report.Failed++
				} else {
					// Keep the independently valid report complete, but make its
					// outcome differ from the already stored valid snapshot.
					report.Results[0].Error = "unused diagnostic"
				}
				if err := writeDiscoveryCorpusReport(filename, report); err != nil {
					t.Fatal(err)
				}
			case "trailing report data":
				data, err := os.ReadFile(filename)
				if err != nil {
					t.Fatal(err)
				}
				writeTestFile(t, filename, string(data)+"{}\n")
			}
			manifest := corpusManifest{SchemaVersion: 1, Targets: targets, Libraries: []libraryManifest{{
				ID: "fixture", Origin: originEcosystemScan, Module: "example.com/a", Version: "v1.0.0",
				Inventory: map[string]expectedInventory{"linux/amd64": {AsmFiles: 1, Packages: []string{"example.com/a"}}},
			}}}
			manifestPath := filepath.Join(t.TempDir(), "manifest.json")
			data, err := json.Marshal(manifest)
			if err != nil {
				t.Fatal(err)
			}
			writeTestFile(t, manifestPath, string(data))
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer cancel()
			stdout, stderr, err := runVerifiedProgressCLI(ctx, root,
				"-manifest="+manifestPath, "-repo-root="+root, "-discovery-ledger="+ledger,
				"-verify-discovery-reports="+reports, "-compare-assembly-ledger="+stored,
				"-emit-discovery-progress")
			if name == "complete" {
				if err != nil || !strings.Contains(string(stderr), "verified discovery corpus reports") {
					t.Fatalf("complete gate: %v stderr=%s", err, stderr)
				}
			} else if err == nil {
				t.Fatalf("%s passed the strict gate", name)
			}
			if name == "trailing report data" {
				if len(stdout) != 0 || !strings.Contains(string(stderr), "more than one JSON value") {
					t.Fatalf("unaudited report emitted progress: %q %s", stdout, stderr)
				}
				return
			}
			var observed discoveryProgress
			decoder := json.NewDecoder(bytes.NewReader(stdout))
			decoder.DisallowUnknownFields()
			if err := decoder.Decode(&observed); err != nil {
				t.Fatalf("audited progress was lost: %v stdout=%q stderr=%s", err, stdout, stderr)
			}
			if err := ensureJSONEOF(decoder); err != nil {
				t.Fatal(err)
			}
			if observed.Source != source || observed.CandidateTotal != 2 {
				t.Fatalf("wrong audited inventory: %+v", observed)
			}
			if name == "missing shard" && (observed.Complete || observed.Verified || observed.Pending != 2) {
				t.Fatalf("missing work was promoted: %+v", observed)
			}
			if name == "failed candidate" && (observed.Verified || observed.Failed != 1) {
				t.Fatalf("failure was promoted: %+v", observed)
			}
		})
	}
}

func verifiedProgressSourceFixture(t *testing.T) (string, discoverySourceIdentity) {
	t.Helper()
	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "source.txt"), "first-party immutable reader fixture\n")
	for _, args := range [][]string{{"init", "-q"}, {"add", "."}, {"commit", "-qm", "fixture"}} {
		cmd := exec.Command("git", append([]string{"-C", root, "-c", "user.name=Test", "-c", "user.email=test@example.invalid",
			"-c", "commit.gpgsign=false", "-c", "core.hooksPath=" + filepath.Join(root, "no-hooks")}, args...)...)
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v %s", args, err, output)
		}
	}
	source, err := collectDiscoverySource(root)
	if err != nil || source.Dirty {
		t.Fatalf("source fixture: %+v %v", source, err)
	}
	return root, source
}

func runVerifiedProgressCLI(ctx context.Context, root string, args ...string) ([]byte, []byte, error) {
	env := replaceEnv(os.Environ(), map[string]string{"PLAN9ASM_VERIFIED_PROGRESS_CLI_HELPER": "1"})
	return runDiscoveryMachineCommand(ctx, root, env, os.Args[0],
		append([]string{"-test.run=^TestDiscoveryVerifiedProgressCLIHelper$", "--"}, args...)...)
}

func TestDiscoveryVerifiedProgressCLIHelper(t *testing.T) {
	if os.Getenv("PLAN9ASM_VERIFIED_PROGRESS_CLI_HELPER") != "1" {
		return
	}
	for index, arg := range os.Args {
		if arg == "--" {
			os.Args = append(os.Args[:1:1], os.Args[index+1:]...)
			flag.CommandLine = flag.NewFlagSet("plan9asmcorpus", flag.ExitOnError)
			main()
			os.Exit(0)
		}
	}
	t.Fatal("missing helper argument boundary")
}
