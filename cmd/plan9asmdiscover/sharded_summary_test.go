package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestDiscoverySummaryDoesNotRetainRecordPayloads(t *testing.T) {
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
	summary, err := readShardedSummary(ledger)
	close(stop)
	maximum := <-peak
	if err != nil {
		t.Fatal(err)
	}
	if summary.FailureRecords != records || summary.ScannedRecords != 0 || summary.MatchedRecords != 0 {
		t.Fatalf("summary = %#v, want %d failures only", summary, records)
	}
	const budget = 24 << 20
	var growth uint64
	if maximum > before.HeapAlloc {
		growth = maximum - before.HeapAlloc
	}
	t.Logf("sampled HeapAlloc growth=%d bytes for %d records, %d payload bytes", growth, records, records*8192)
	if growth > budget {
		t.Fatalf("summary retains ledger-sized payloads: sampled HeapAlloc growth %d exceeds %d bytes", growth, budget)
	}
}

// Write one record at a time. The fixture itself does not keep the large input
// in Go memory, and shard order is independent of the implementation reader.
func writeSummaryPayloadFixture(t *testing.T, count int) string {
	t.Helper()
	ledger := t.TempDir()
	recordsDir := filepath.Join(ledger, "records")
	if err := os.Mkdir(recordsDir, 0o700); err != nil {
		t.Fatal(err)
	}
	files := make(map[byte]*os.File)
	writers := make(map[byte]*bufio.Writer)
	t.Cleanup(func() {
		for _, file := range files {
			_ = file.Close()
		}
	})
	payload := strings.Repeat("p", 8192)
	for i := 0; i < count; i++ {
		module := fmt.Sprintf("example.com/payload/%08d", i)
		shard := sha256.Sum256([]byte(module))[0]
		if files[shard] == nil {
			file, err := os.Create(filepath.Join(recordsDir, fmt.Sprintf("%02x.jsonl", shard)))
			if err != nil {
				t.Fatal(err)
			}
			files[shard] = file
			writers[shard] = bufio.NewWriter(file)
		}
		item := shardedRecord{Kind: "failure", Module: module, Version: "v1.0.0", Error: payload}
		if err := json.NewEncoder(writers[shard]).Encode(item); err != nil {
			t.Fatal(err)
		}
	}
	for shard, writer := range writers {
		if err := writer.Flush(); err != nil {
			t.Fatal(err)
		}
		if err := files[shard].Close(); err != nil {
			t.Fatal(err)
		}
		delete(files, shard)
	}
	manifest := shardedManifest{
		SchemaVersion: discoverySchema, Format: shardedFormat,
		GeneratedAt:   time.Date(2026, 10, 7, 0, 0, 0, 0, time.UTC),
		UniqueModules: count, Failures: count, ShardKey: "sha256(module)[0]",
	}
	writeSummaryFixtureJSON(t, filepath.Join(ledger, "manifest.json"), manifest)
	return ledger
}

func writeSummaryFixtureJSON(t *testing.T, name string, value any) {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(name, data, 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestDiscoverySummaryMatchesCompleteReader(t *testing.T) {
	report := discoveryReport{
		SchemaVersion: discoverySchema,
		IndexRanges: []indexRange{
			{Since: indexEpoch, Before: "2026-10-01T00:00:00Z", IndexEntries: 10},
			{Since: "2026-10-01T00:00:00Z", Before: "2026-10-07T00:00:00Z", IndexEntries: 20},
		},
		Scanned:  []moduleVersion{{Path: "example.com/a", Version: "v1.0.0"}, {Path: "example.com/a", Version: "v0.9.0"}},
		Matched:  []candidate{{Module: "example.com/a", Version: "v1.0.0", AsmFiles: []string{"a_amd64.s", "b_arm64.s"}}},
		Failures: []scanFailure{{Module: "example.com/b", Version: "@latest", Error: "unavailable"}},
	}
	ledger := filepath.Join(t.TempDir(), "ledger")
	if err := writeShardedReport(ledger, report); err != nil {
		t.Fatal(err)
	}
	complete, err := readShardedReport(ledger)
	if err != nil {
		t.Fatal(err)
	}
	got, err := readShardedSummary(ledger)
	if err != nil {
		t.Fatal(err)
	}
	if want := summarizeLedger(complete); !reflect.DeepEqual(got, want) {
		t.Fatalf("summary = %#v, want complete-reader view %#v", got, want)
	}
}

func TestDiscoveryStatusAndPublishedReadbackUseSummaryOnly(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), "main.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	var main *ast.FuncDecl
	for _, declaration := range file.Decls {
		if function, ok := declaration.(*ast.FuncDecl); ok && function.Name.Name == "main" {
			main = function
		}
	}
	if main == nil {
		t.Fatal("main function is absent")
	}
	statusSummary := false
	for _, statement := range main.Body.List {
		conditional, ok := statement.(*ast.IfStmt)
		if !ok {
			continue
		}
		star, ok := conditional.Cond.(*ast.StarExpr)
		if !ok {
			continue
		}
		name, ok := star.X.(*ast.Ident)
		if !ok || name.Name != "statusOnly" {
			continue
		}
		ast.Inspect(conditional.Body, func(node ast.Node) bool {
			if call, ok := node.(*ast.CallExpr); ok {
				if function, ok := call.Fun.(*ast.Ident); ok && function.Name == "readShardedSummary" {
					statusSummary = true
				}
			}
			return true
		})
		break
	}
	if !statusSummary {
		t.Error("actual -status branch does not independently use readShardedSummary before full existing-ledger loading")
	}
	var publishedFullRead bool
	ast.Inspect(main.Body, func(node ast.Node) bool {
		assignment, ok := node.(*ast.AssignStmt)
		if !ok || len(assignment.Lhs) == 0 {
			return true
		}
		name, ok := assignment.Lhs[0].(*ast.Ident)
		if ok && name.Name == "persisted" {
			for _, expression := range assignment.Rhs {
				if call, ok := expression.(*ast.CallExpr); ok {
					if function, ok := call.Fun.(*ast.Ident); ok && function.Name == "readShardedReport" {
						publishedFullRead = true
					}
				}
			}
		}
		return true
	})
	if publishedFullRead {
		t.Error("actual publication readback retains complete record payloads instead of a validated summary")
	}
}
