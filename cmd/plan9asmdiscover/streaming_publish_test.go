package main

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
	"time"
)

func TestDiscoveryPublicationDoesNotRetainOldLedgerPayloads(t *testing.T) {
	const records = 8192
	ledger := writeSummaryPayloadFixture(t, records)
	runtime.GC()
	var before runtime.MemStats
	runtime.ReadMemStats(&before)

	stop := make(chan struct{})
	peak := make(chan uint64, 1)
	go func() {
		maximum := before.HeapAlloc
		ticker := time.NewTicker(time.Millisecond)
		defer ticker.Stop()
		for {
			var stats runtime.MemStats
			runtime.ReadMemStats(&stats)
			if stats.HeapAlloc > maximum {
				maximum = stats.HeapAlloc
			}
			select {
			case <-stop:
				peak <- maximum
				return
			case <-ticker.C:
			}
		}
	}()
	err := writeShardedReport(ledger, discoveryReport{
		SchemaVersion: discoverySchema,
		Scanned:       []moduleVersion{{Path: "example.com/new", Version: "v1.0.0"}},
	})
	close(stop)
	maximum := <-peak
	if err != nil {
		t.Fatal(err)
	}
	var growth uint64
	if maximum > before.HeapAlloc {
		growth = maximum - before.HeapAlloc
	}
	t.Logf("sampled HeapAlloc growth=%d bytes while publishing into %d records/%d payload bytes", growth, records, records*8192)
	if growth > 24<<20 {
		t.Errorf("publication retains old-ledger payloads: sampled HeapAlloc growth %d exceeds %d", growth, 24<<20)
	}
	summary, err := readShardedSummary(ledger)
	if err != nil || summary.ScannedRecords != 1 || summary.FailureRecords != records || summary.UniqueModules != records+1 {
		t.Fatalf("published summary=%+v: %v", summary, err)
	}
}

func TestDiscoveryStreamingPublicationMatchesReferenceMerge(t *testing.T) {
	base := discoveryReport{
		SchemaVersion: discoverySchema,
		GeneratedAt:   time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC),
		IndexRanges:   []indexRange{{Since: "2025-10-01T00:00:00Z", Before: "2026-10-01T00:00:00Z", IndexEntries: 20}},
		Skipped:       2,
		Scanned: []moduleVersion{
			{Path: "example.com/a", Version: "v0.9.0"},
			{Path: "example.com/a", Version: "v1.9.0"},
			{Path: "example.com/a/v2", Version: "v2.0.0"},
		},
		Matched: []candidate{{Module: "example.com/a", Version: "v1.9.0", AsmFiles: []string{"a_arm64.s"}}},
		Failures: []scanFailure{
			{Module: "example.com/a", Version: "@latest", Error: "resolution unavailable"},
			{Module: "example.com/b", Version: "v1.0.0", Error: "old failure"},
			{Module: "example.com/c", Version: "@latest", Error: "still retryable"},
		},
	}
	updates := []discoveryReport{
		{
			SchemaVersion: discoverySchema,
			GeneratedAt:   base.GeneratedAt.Add(time.Hour),
			IndexRanges:   []indexRange{{Since: "2026-10-01T00:00:00Z", Before: "2026-10-02T00:00:00Z", IndexEntries: 10}},
			Skipped:       1,
			Scanned:       []moduleVersion{{Path: "example.com/a", Version: "v1.10.0"}},
			Matched:       []candidate{{Module: "example.com/a", Version: "v1.10.0", AsmFiles: []string{"a_amd64.s"}}},
			Failures: []scanFailure{
				{Module: "example.com/b", Version: "v1.0.0", Error: "first payload"},
				{Module: "example.com/b", Version: "v1.0.0", Error: "last payload wins"},
			},
		},
		{
			SchemaVersion: discoverySchema,
			Scanned:       []moduleVersion{{Path: "example.com/b", Version: "v1.1.0"}},
			Matched:       []candidate{{Module: "example.com/a", Version: "v1.10.0", AsmFiles: []string{"b_arm.s"}}},
		},
		{},
	}
	for _, update := range updates {
		ledger := filepath.Join(t.TempDir(), "ledger")
		if err := writeShardedReport(ledger, base); err != nil {
			t.Fatal(err)
		}
		want, err := mergeDiscoveryReports(base, update)
		if err != nil {
			t.Fatal(err)
		}
		if err := writeShardedReport(ledger, update); err != nil {
			t.Fatal(err)
		}
		got, err := readShardedReport(ledger)
		if err != nil || !reflect.DeepEqual(got, want) {
			t.Fatalf("streamed publication=%+v, want %+v: %v", got, want, err)
		}
	}
}

func TestDiscoveryStreamingPublicationRejectsCorruptUnchangedTail(t *testing.T) {
	ledger := writeSummaryPayloadFixture(t, 256)
	manifestPath := filepath.Join(ledger, "manifest.json")
	before, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	shards, err := filepath.Glob(filepath.Join(ledger, "records", "*.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	last := shards[len(shards)-1]
	file, err := os.OpenFile(last, os.O_WRONLY|os.O_APPEND, 0)
	if err != nil {
		t.Fatal(err)
	}
	_, writeErr := file.WriteString("{corrupt unselected tail}\n")
	closeErr := file.Close()
	if writeErr != nil || closeErr != nil {
		t.Fatalf("write=%v close=%v", writeErr, closeErr)
	}
	if err := writeShardedReport(ledger, discoveryReport{}); err == nil {
		t.Fatal("publication accepted corrupt unchanged tail")
	}
	after, err := os.ReadFile(manifestPath)
	if err != nil || string(before) != string(after) {
		t.Fatalf("failed publication changed authoritative manifest: %v", err)
	}
	data, err := os.ReadFile(last)
	if err != nil || !bytes.HasSuffix(data, []byte("{corrupt unselected tail}\n")) {
		t.Fatalf("failed publication modified the unchanged shard: %v", err)
	}
	entries, err := os.ReadDir(filepath.Dir(ledger))
	if err != nil || len(entries) != 1 || entries[0].Name() != filepath.Base(ledger) {
		t.Fatalf("failed publication leaked a lock/staging directory: entries=%v err=%v", entries, err)
	}
}
