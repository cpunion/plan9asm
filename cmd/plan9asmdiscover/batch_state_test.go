package main

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestDiscoveryBatchStateRetainsOnlyActiveRecords(t *testing.T) {
	const records = 4096
	report := discoveryReport{SchemaVersion: discoverySchema}
	for i := 0; i < records; i++ {
		report.Scanned = append(report.Scanned, moduleVersion{
			Path: fmt.Sprintf("example.org/batch/%06d", i), Version: "v1.0.0",
		})
	}
	ledger := filepath.Join(t.TempDir(), "ledger")
	if err := writeShardedReport(ledger, report); err != nil {
		t.Fatal(err)
	}
	active := []moduleVersion{{Path: "example.org/batch/000042", Version: "v1.1.0"}}
	state, err := loadBatchState([]string{ledger}, active)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("retained seen=%d sourceSeen=%d checkpoints=%d for one active path, %d input records",
		len(state.seen), len(state.sourceSeen), len(state.checkpoints), records)
	if len(state.seen) != 1 || len(state.sourceSeen) != 0 || len(state.checkpoints) != 1 {
		t.Fatalf("retains ledger-sized maps: seen=%d sourceSeen=%d checkpoints=%d, want 1/0/1",
			len(state.seen), len(state.sourceSeen), len(state.checkpoints))
	}
	if _, ok := state.seen[scanKey(moduleVersion{Path: active[0].Path, Version: "v1.0.0"})]; !ok {
		t.Fatal("active exact-version checkpoint missing")
	}
}

func TestDiscoveryBatchStatePreservesCaseAndVersionDomains(t *testing.T) {
	pseudo := "v0.0.0-20251001000000-0123456789ab"
	items := []moduleVersion{
		{Path: "github.com/OWNER/Repo", Version: pseudo},
		{Path: "github.com/owner/repo/v2", Version: "v2.0.0"},
		{Path: "github.com/owner/repo/Sub", Version: pseudo},
		{Path: "github.com/owner/repo/sub", Version: pseudo},
		{Path: "example.org/Repo", Version: "v0.9.0"},
		{Path: "example.org/Repo", Version: "v1.0.0"},
		{Path: "example.org/repo", Version: "v1.0.0"},
	}
	ledger := filepath.Join(t.TempDir(), "ledger")
	if err := writeShardedReport(ledger, discoveryReport{Scanned: items}); err != nil {
		t.Fatal(err)
	}
	state, err := loadBatchState([]string{ledger}, []moduleVersion{
		{Path: "github.com/owner/repo", Version: pseudo},
		{Path: "github.com/OWNER/REPO/Sub", Version: pseudo},
		{Path: "example.org/Repo", Version: "v1.1.0"},
	})
	if err != nil {
		t.Fatal(err)
	}
	wantSeen := make(map[string]seenResult)
	wantSource := make(map[string]seenResult)
	wantCheckpoints := make(map[string]moduleCheckpoint)
	for _, item := range []moduleVersion{items[0], items[2], items[4], items[5]} {
		rememberSeenResult(wantSeen, scanKey(item), seenResult{})
		rememberSeenResult(wantSource, sourceScanKey(item), seenResult{})
		chooseModuleCheckpoint(wantCheckpoints, item.Path, item.Version)
	}
	if !reflect.DeepEqual(state, scanBatchState{wantSeen, wantSource, wantCheckpoints}) {
		t.Fatalf("batch crosses case/submodule/major boundaries: %+v", state)
	}
}

func TestDiscoveryBatchStateEmptyInterestStillValidatesAllShards(t *testing.T) {
	ledger := filepath.Join(t.TempDir(), "ledger")
	if err := writeShardedReport(ledger, discoveryReport{
		Scanned: []moduleVersion{{Path: "example.org/unselected", Version: "v1.0.0"}},
	}); err != nil {
		t.Fatal(err)
	}
	state, err := loadBatchState([]string{ledger}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(state.seen)+len(state.sourceSeen)+len(state.checkpoints) != 0 {
		t.Errorf("empty batch retains unrelated state: %+v", state)
	}
	entries, err := os.ReadDir(filepath.Join(ledger, "records"))
	if err != nil || len(entries) != 1 {
		t.Fatalf("fixture inventory: %v %v", entries, err)
	}
	shard := filepath.Join(ledger, "records", entries[0].Name())
	file, err := os.OpenFile(shard, os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := file.WriteString("not-json\n"); err != nil {
		_ = file.Close()
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	state, err = loadBatchState([]string{ledger}, nil)
	if err == nil || !strings.Contains(err.Error(), "decode record shard") ||
		state.seen != nil || state.sourceSeen != nil || state.checkpoints != nil {
		t.Fatalf("unselected invalid tail supplied state: %+v, %v", state, err)
	}
}
