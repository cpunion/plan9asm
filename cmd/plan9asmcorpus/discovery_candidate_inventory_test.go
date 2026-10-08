package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func writeDiscoveryForeignPayloadLedger(t *testing.T, directory bool, foreign int) (string, discoveryCandidate) {
	t.Helper()
	const shardIndex, shardCount = 36, 64
	selected := discoveryCandidate{Module: "example.com/exact/0000", Version: "v1.0.0",
		Architectures: []string{"amd64"}, AsmFiles: []string{"selected_amd64.s"}}
	for index := 0; discoveryCandidateShard(selected, shardCount) != shardIndex; index++ {
		selected.Module = fmt.Sprintf("example.com/selected/%06d", index)
	}
	ledger := filepath.Join(t.TempDir(), "ledger.jsonl")
	filePath := ledger
	if directory {
		if err := os.MkdirAll(filepath.Join(ledger, "records"), 0700); err != nil {
			t.Fatal(err)
		}
		filePath = filepath.Join(ledger, "records", "00.jsonl")
	}
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	writer := bufio.NewWriter(file)
	encoder := json.NewEncoder(writer)
	write := func(candidate discoveryCandidate) {
		t.Helper()
		err := encoder.Encode(discoveryRecord{Kind: "matched", Module: candidate.Module,
			Version: candidate.Version, Architectures: candidate.Architectures, AsmFiles: candidate.AsmFiles})
		if err != nil {
			file.Close()
			t.Fatal(err)
		}
	}
	write(selected)
	prefix := strings.Repeat("nested_source_directory/", 32)
	paths := make([]string, 4)
	for index := range paths {
		paths[index] = fmt.Sprintf("%soperation%d_amd64.s", prefix, index)
	}
	for index, written := 0, 0; written < foreign; index++ {
		candidate := discoveryCandidate{Module: fmt.Sprintf("example.com/foreign/%06d", index),
			Version: "v1.0.0", Architectures: []string{"amd64", "arm64", "s390x"}, AsmFiles: paths}
		if discoveryCandidateShard(candidate, shardCount) == shardIndex {
			continue
		}
		write(candidate)
		written++
	}
	if err := writer.Flush(); err != nil {
		file.Close()
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	return ledger, selected
}

func TestDiscoveryCorpusRetainsOnlySelectedShardPayload(t *testing.T) {
	const foreign, budget = 18000, 32 << 20
	for _, directory := range []bool{false, true} {
		t.Run(fmt.Sprintf("directory=%t", directory), func(t *testing.T) {
			ledger, selected := writeDiscoveryForeignPayloadLedger(t, directory, foreign)
			provenance := fixtureDiscoveryProvenance(t, ledger)
			cause := errors.New("first-party memory orchestration fixture")
			cfg := discoveryCorpusConfig{
				LedgerPath: ledger, RepoRoot: t.TempDir(), ShardIndex: 36, ShardCount: 64,
				Targets: []string{"linux/amd64"}, CandidateTimeout: time.Minute,
				ReportPath: filepath.Join(t.TempDir(), "shard.json"),
				captureProvenance: func(discoveryCorpusConfig) (discoveryCorpusProvenance, error) {
					return provenance, nil
				},
			}
			calls := 0
			cfg.runCandidate = func(_ discoveryCorpusConfig, candidate discoveryCandidate, _ string) (matrixReport, []string, []discoveryBuildConfiguration, error) {
				calls++
				if !reflect.DeepEqual(candidate, selected) {
					t.Errorf("selected payload changed: got %+v, want %+v", candidate, selected)
				}
				// An orchestration fixture is not translation/pass evidence.
				return matrixReport{}, nil, nil, cause
			}
			growth, err := sampleSupersededMatchedHeap(func() error {
				err := runDiscoveryCorpus(cfg)
				if err == nil || !strings.Contains(err.Error(), "failed=1") {
					return fmt.Errorf("intentional first-party failure was suppressed: %v", err)
				}
				return nil
			})
			if err != nil {
				t.Fatal(err)
			}
			report, err := readDiscoveryCorpusReport(cfg.ReportPath)
			if err != nil {
				t.Fatal(err)
			}
			if calls != 1 || report.CandidateTotal != foreign+1 || report.EligibleCandidates != foreign+1 ||
				report.Partial || report.Selected != 1 || report.Failed != 1 || report.Passed != 0 ||
				report.NotApplicable != 0 || report.Results[0].Error != cause.Error() {
				t.Fatalf("global denominator/selected failure changed: calls=%d report=%+v", calls, report)
			}
			t.Logf("foreign matched=%d, paths/record=4, selected=1, sampled HeapAlloc growth=%d, budget=%d", foreign, growth, budget)
			if growth > budget {
				t.Errorf("corpus retains foreign shard payloads: sampled growth %d exceeds %d", growth, budget)
			}
		})
	}
}
