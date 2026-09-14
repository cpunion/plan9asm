package main

import (
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestEscapeProxyPath(t *testing.T) {
	got, err := escapeProxyPath("github.com/Azure/azure-sdk-for-go")
	if err != nil {
		t.Fatal(err)
	}
	if want := "github.com/!azure/azure-sdk-for-go"; got != want {
		t.Fatalf("escapeProxyPath() = %q, want %q", got, want)
	}
}

func TestValidateConfigRejectsNonPositiveHTTPTimeout(t *testing.T) {
	err := validateConfig(config{
		indexURL:    defaultIndexURL,
		proxyURL:    defaultProxyURL,
		workers:     1,
		maxZipSize:  1,
		httpTimeout: 0,
	})
	if err == nil {
		t.Fatal("validateConfig() accepted a non-positive HTTP timeout")
	}
}

func TestSeenReportsRecordOnlyCompletedExactVersions(t *testing.T) {
	reportPath := filepath.Join(t.TempDir(), "previous.json")
	data, err := json.Marshal(discoveryReport{
		SchemaVersion: 1,
		Scanned: []moduleVersion{
			{Path: "example.com/a", Version: "v1.0.0"},
			{Path: "example.com/noasm", Version: "v2.0.0"},
		},
		Matched: []candidate{{Module: "example.com/old-match", Version: "v3.0.0"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(reportPath, data, 0o600); err != nil {
		t.Fatal(err)
	}
	seen, err := loadSeenReports([]string{reportPath})
	if err != nil {
		t.Fatal(err)
	}
	wantSeen := []moduleVersion{
		{Path: "example.com/a", Version: "v1.0.0"},
		{Path: "example.com/noasm", Version: "v2.0.0"},
		{Path: "example.com/old-match", Version: "v3.0.0"},
	}
	for _, item := range wantSeen {
		if _, ok := seen[scanKey(item)]; !ok {
			t.Errorf("completed version is not recorded: %s@%s", item.Path, item.Version)
		}
	}
	for _, item := range []moduleVersion{
		{Path: "example.com/a", Version: "v1.1.0"},
		{Path: "example.com/new", Version: "v1.0.0"},
	} {
		if _, ok := seen[scanKey(item)]; ok {
			t.Errorf("unscanned exact version is recorded: %s@%s", item.Path, item.Version)
		}
	}
}

func TestSeenReportsReadCompressedLedger(t *testing.T) {
	reportPath := filepath.Join(t.TempDir(), "scan-ledger.json.gz")
	file, err := os.Create(reportPath)
	if err != nil {
		t.Fatal(err)
	}
	zw := gzip.NewWriter(file)
	if err := json.NewEncoder(zw).Encode(discoveryReport{
		SchemaVersion: discoverySchema,
		Scanned: []moduleVersion{
			{Path: "example.com/with-asm", Version: "v1.0.0"},
			{Path: "example.com/without-asm", Version: "v2.0.0"},
		},
	}); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}

	seen, err := loadSeenReports([]string{reportPath})
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range []moduleVersion{
		{Path: "example.com/with-asm", Version: "v1.0.0"},
		{Path: "example.com/without-asm", Version: "v2.0.0"},
	} {
		if _, ok := seen[scanKey(item)]; !ok {
			t.Errorf("compressed ledger omitted %s@%s", item.Path, item.Version)
		}
	}
}

func TestLoadSeenReportsPreservesMatchAcrossReportOrder(t *testing.T) {
	dir := t.TempDir()
	item := candidate{
		Module: "example.com/asm", Version: "v1.0.0",
		Architectures: []string{"amd64"}, AsmFiles: []string{"asm_amd64.s"},
	}
	paths := []string{filepath.Join(dir, "matched.json"), filepath.Join(dir, "scanned.json")}
	for i, report := range []discoveryReport{
		{SchemaVersion: discoverySchema, Matched: []candidate{item}},
		{SchemaVersion: discoverySchema, Scanned: []moduleVersion{{Path: item.Module, Version: item.Version}}},
	} {
		data, err := json.Marshal(report)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(paths[i], data, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	seen, err := loadSeenReports(paths)
	if err != nil {
		t.Fatal(err)
	}
	if got := seen[scanKey(moduleVersion{Path: item.Module, Version: item.Version})].Match; !reflect.DeepEqual(got, item) {
		t.Fatalf("later scanned-only report erased assembly metadata: %#v", got)
	}
}

func TestShardedReportRoundTripIsDeterministic(t *testing.T) {
	report := discoveryReport{
		SchemaVersion: discoverySchema,
		GeneratedAt:   time.Date(2026, 9, 13, 12, 34, 56, 0, time.UTC),
		Since:         "2019-04-10T00:00:00Z",
		NextSince:     "2019-04-11T00:00:00Z",
		IndexEntries:  3,
		UniqueModules: 3,
		Scanned: []moduleVersion{
			{Path: "example.com/noasm", Version: "v2.0.0"},
			{Path: "example.com/asm", Version: "v1.0.0"},
		},
		Matched: []candidate{{
			Module:        "example.com/asm",
			Version:       "v1.0.0",
			Architectures: []string{"arm64", "amd64"},
			AsmFiles:      []string{"z_arm64.s", "a_amd64.s"},
		}},
		Failures: []scanFailure{{
			Module: "example.com/broken", Version: "v3.0.0", Error: "missing ZIP",
		}},
	}

	dir1 := filepath.Join(t.TempDir(), "run-one")
	if err := writeShardedReport(dir1, report); err != nil {
		t.Fatal(err)
	}
	reordered := report
	reordered.Scanned = append([]moduleVersion(nil), report.Scanned...)
	reordered.Scanned[0], reordered.Scanned[1] = reordered.Scanned[1], reordered.Scanned[0]
	dir2 := filepath.Join(t.TempDir(), "run-two")
	if err := writeShardedReport(dir2, reordered); err != nil {
		t.Fatal(err)
	}

	files1 := readDirectoryFiles(t, dir1)
	files2 := readDirectoryFiles(t, dir2)
	if !reflect.DeepEqual(files1, files2) {
		t.Fatalf("sharded output depends on input order\nfirst: %#v\nsecond: %#v", files1, files2)
	}
	if _, ok := files1["manifest.json"]; !ok {
		t.Fatal("sharded output has no manifest.json")
	}
	for name, data := range files1 {
		if strings.HasSuffix(name, ".gz") {
			t.Fatalf("sharded output contains opaque gzip file %q", name)
		}
		if strings.HasPrefix(name, "records/") && !bytes.HasSuffix(data, []byte("\n")) {
			t.Errorf("JSONL shard %q has no trailing newline", name)
		}
	}

	loaded, err := readSeenReport(dir1)
	if err != nil {
		t.Fatal(err)
	}
	sortDiscoveryReport(&report)
	if !reflect.DeepEqual(loaded, report) {
		t.Fatalf("round trip = %#v, want %#v", loaded, report)
	}

	hash := sha256.Sum256([]byte("example.com/asm"))
	wantShard := fmt.Sprintf("records/%02x.jsonl", hash[0])
	if _, ok := files1[wantShard]; !ok {
		t.Fatalf("module was not written to stable shard %q; files: %#v", wantShard, files1)
	}
}

func TestShardedLedgerUpdatePreservesVersionsAndUsesSemverOrder(t *testing.T) {
	ledger := filepath.Join(t.TempDir(), "ledger")
	initial := discoveryReport{
		SchemaVersion: discoverySchema,
		GeneratedAt:   time.Date(2026, 9, 13, 12, 0, 0, 0, time.UTC),
		Since:         "2026-09-01T00:00:00Z",
		NextSince:     "2026-09-02T00:00:00Z",
		IndexEntries:  10,
		Scanned: []moduleVersion{
			{Path: "example.com/a", Version: "v1.9.0"},
			{Path: "example.com/a", Version: "v1.10.0-rc.1"},
		},
		Matched: []candidate{{
			Module: "example.com/a", Version: "v1.9.0",
			Architectures: []string{"amd64"}, AsmFiles: []string{"old_amd64.s"},
		}},
		Failures: []scanFailure{{Module: "example.com/a", Version: "v1.10.0", Error: "temporary proxy failure"}},
	}
	if err := writeShardedReport(ledger, initial); err != nil {
		t.Fatal(err)
	}
	update := discoveryReport{
		SchemaVersion: discoverySchema,
		GeneratedAt:   time.Date(2026, 9, 14, 12, 0, 0, 0, time.UTC),
		Since:         "2026-09-02T00:00:00Z",
		NextSince:     "2026-09-03T00:00:00Z",
		IndexEntries:  11,
		Scanned:       []moduleVersion{{Path: "example.com/a", Version: "v1.10.0"}},
		Matched: []candidate{{
			Module: "example.com/a", Version: "v1.10.0",
			Architectures: []string{"arm64"}, AsmFiles: []string{"new_arm64.s"},
		}},
	}
	if err := writeShardedReport(ledger, update); err != nil {
		t.Fatal(err)
	}

	report, err := readSeenReport(ledger)
	if err != nil {
		t.Fatal(err)
	}
	wantScanned := []moduleVersion{
		{Path: "example.com/a", Version: "v1.9.0"},
		{Path: "example.com/a", Version: "v1.10.0-rc.1"},
		{Path: "example.com/a", Version: "v1.10.0"},
	}
	if !reflect.DeepEqual(report.Scanned, wantScanned) {
		t.Fatalf("scanned versions = %#v, want semver order %#v", report.Scanned, wantScanned)
	}
	if len(report.Matched) != 2 || len(report.Failures) != 0 {
		t.Fatalf("updated ledger lost matches or retained a completed failure: %#v", report)
	}
	if report.IndexEntries != 21 || report.UniqueModules != 1 {
		t.Fatalf("updated ledger counts = index %d, modules %d; want 21, 1", report.IndexEntries, report.UniqueModules)
	}
	files := readDirectoryFiles(t, ledger)
	for name := range files {
		if strings.HasPrefix(name, "runs/") {
			t.Fatalf("updated ledger created a per-run path %q", name)
		}
	}
	hash := sha256.Sum256([]byte("example.com/a"))
	shardName := fmt.Sprintf("records/%02x.jsonl", hash[0])
	shard := string(files[shardName])
	if strings.Index(shard, `"version":"v1.9.0"`) > strings.Index(shard, `"version":"v1.10.0-rc.1"`) ||
		strings.Index(shard, `"version":"v1.10.0-rc.1"`) > strings.Index(shard, `"version":"v1.10.0"`) {
		t.Fatalf("record shard is not in semver order:\n%s", shard)
	}
}

func TestSortDiscoveryReportUsesGoSemver(t *testing.T) {
	report := discoveryReport{Scanned: []moduleVersion{
		{Path: "example.com/a", Version: "v1.10.1-0.20260914000000-0123456789ab"},
		{Path: "example.com/a", Version: "v1.10.0+incompatible"},
		{Path: "example.com/a", Version: "v1.9.0"},
		{Path: "example.com/a", Version: "v1.10.0"},
		{Path: "example.com/a", Version: "v1.10.0-rc.1"},
	}}
	sortDiscoveryReport(&report)
	want := []string{
		"v1.9.0",
		"v1.10.0-rc.1",
		"v1.10.0",
		"v1.10.0+incompatible",
		"v1.10.1-0.20260914000000-0123456789ab",
	}
	for i, item := range report.Scanned {
		if item.Version != want[i] {
			t.Fatalf("semver order[%d] = %q, want %q; all: %#v", i, item.Version, want[i], report.Scanned)
		}
	}
}

func TestLoadSeenReportsAcceptsSingleLedgerDirectory(t *testing.T) {
	ledger := filepath.Join(t.TempDir(), "ledger")
	report := discoveryReport{
		SchemaVersion: discoverySchema,
		Scanned: []moduleVersion{
			{Path: "example.com/a", Version: "v1.0.0"},
			{Path: "example.com/b", Version: "v2.0.0"},
		},
	}
	if err := writeShardedReport(ledger, report); err != nil {
		t.Fatal(err)
	}
	seen, err := loadSeenReports([]string{ledger})
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range []moduleVersion{
		{Path: "example.com/a", Version: "v1.0.0"},
		{Path: "example.com/b", Version: "v2.0.0"},
	} {
		if _, ok := seen[scanKey(item)]; !ok {
			t.Errorf("ledger directory omitted %s@%s", item.Path, item.Version)
		}
	}
	loaded, err := readSeenReport(ledger)
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded.Scanned) != 2 || len(loaded.Failures) != 0 || loaded.UniqueModules != 2 {
		t.Fatalf("ledger report was not normalized: %#v", loaded)
	}
}

func TestShardedReportRejectsInvalidOrUnsafeInput(t *testing.T) {
	t.Run("existing output without manifest", func(t *testing.T) {
		dir := t.TempDir()
		if err := writeShardedReport(dir, discoveryReport{SchemaVersion: discoverySchema}); err == nil {
			t.Fatal("writeShardedReport() overwrote a directory that is not a ledger")
		}
	})

	t.Run("missing manifest", func(t *testing.T) {
		if _, err := readSeenReport(t.TempDir()); err == nil {
			t.Fatal("readSeenReport() accepted a directory without a manifest")
		}
	})

	tests := map[string]struct {
		manifest shardedManifest
		record   string
	}{
		"unsupported format": {
			manifest: shardedManifest{SchemaVersion: discoverySchema, Format: "unknown"},
		},
		"invalid JSONL": {
			manifest: shardedManifest{SchemaVersion: discoverySchema, Format: shardedFormat, Scanned: 1},
			record:   "not json\n",
		},
		"incomplete module": {
			manifest: shardedManifest{SchemaVersion: discoverySchema, Format: shardedFormat, Scanned: 1},
			record:   `{"kind":"scanned","module":"example.com/lib"}` + "\n",
		},
		"unknown kind": {
			manifest: shardedManifest{SchemaVersion: discoverySchema, Format: shardedFormat, Scanned: 1},
			record:   `{"kind":"mystery","module":"example.com/lib","version":"v1.0.0"}` + "\n",
		},
		"wrong count": {
			manifest: shardedManifest{SchemaVersion: discoverySchema, Format: shardedFormat, Scanned: 2},
			record:   `{"kind":"scanned","module":"example.com/lib","version":"v1.0.0"}` + "\n",
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			manifest, err := json.Marshal(test.manifest)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(dir, "manifest.json"), manifest, 0o600); err != nil {
				t.Fatal(err)
			}
			if test.record != "" {
				if err := os.Mkdir(filepath.Join(dir, "records"), 0o700); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(dir, "records", "00.jsonl"), []byte(test.record), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			if _, err := readSeenReport(dir); err == nil {
				t.Fatal("readSeenReport() accepted an invalid sharded report")
			}
		})
	}
}

func readDirectoryFiles(t *testing.T, root string) map[string][]byte {
	t.Helper()
	files := make(map[string][]byte)
	err := filepath.WalkDir(root, func(name string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		data, err := os.ReadFile(name)
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(root, name)
		if err != nil {
			return err
		}
		files[filepath.ToSlash(relative)] = data
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return files
}

func TestSeenReportsRejectInvalidInput(t *testing.T) {
	tests := map[string][]byte{
		"invalid-json":        []byte("not json"),
		"invalid-gzip-header": {0x1f, 0x8b},
	}
	var compressed bytes.Buffer
	zw := gzip.NewWriter(&compressed)
	_, _ = zw.Write([]byte(`{"schema_version":2}`))
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	corrupt := append([]byte(nil), compressed.Bytes()...)
	corrupt[len(corrupt)-1] ^= 0xff
	tests["invalid-gzip-checksum"] = corrupt

	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			reportPath := filepath.Join(t.TempDir(), name)
			if err := os.WriteFile(reportPath, data, 0o600); err != nil {
				t.Fatal(err)
			}
			if _, err := loadSeenReports([]string{reportPath}); err == nil {
				t.Fatal("loadSeenReports() accepted invalid input")
			}
		})
	}

	if _, err := loadSeenReports([]string{filepath.Join(t.TempDir(), "missing.json")}); err == nil {
		t.Fatal("loadSeenReports() accepted a missing report")
	}
}

func TestCommittedScanLedger(t *testing.T) {
	reportPath := filepath.Join("..", "..", "testdata", "discovery", "ledger")
	err := filepath.WalkDir(filepath.Dir(reportPath), func(name string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".gz") {
			t.Errorf("compressed discovery result must not be committed: %s", name)
		}
		if relative, relErr := filepath.Rel(reportPath, name); relErr == nil {
			relative = filepath.ToSlash(relative)
			if relative == "." || relative == ".." || strings.HasPrefix(relative, "../") {
				return nil
			}
			parts := strings.Split(relative, "/")
			if parts[0] != "manifest.json" && parts[0] != "records" {
				t.Errorf("discovery ledger contains a per-run or unexpected path: %s", relative)
			}
			if len(parts) > 2 {
				t.Errorf("discovery ledger must have exactly one records directory: %s", relative)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	report, err := readSeenReport(reportPath)
	if err != nil {
		t.Fatal(err)
	}
	if report.SchemaVersion != discoverySchema {
		t.Fatalf("ledger schema = %d, want %d", report.SchemaVersion, discoverySchema)
	}
	if report.IndexEntries == 0 || len(report.Scanned) == 0 {
		t.Fatalf("ledger is empty: %d index entries, %d scanned versions", report.IndexEntries, len(report.Scanned))
	}
	if _, err := time.Parse(time.RFC3339Nano, report.NextSince); err != nil {
		t.Fatalf("invalid ledger cursor %q: %v", report.NextSince, err)
	}

	seen, err := loadSeenReports([]string{reportPath})
	if err != nil {
		t.Fatal(err)
	}
	if len(seen) != len(report.Scanned) {
		t.Fatalf("ledger has %d unique completed versions, want %d", len(seen), len(report.Scanned))
	}
	for _, item := range report.Matched {
		key := scanKey(moduleVersion{Path: item.Module, Version: item.Version})
		if _, ok := seen[key]; !ok {
			t.Errorf("assembly match is not recorded as scanned: %s@%s", item.Module, item.Version)
		}
	}
	for _, item := range report.Failures {
		key := scanKey(moduleVersion{Path: item.Module, Version: item.Version})
		if _, ok := seen[key]; ok {
			t.Errorf("failed version must remain eligible for retry: %s@%s", item.Module, item.Version)
		}
	}

	issueMatches := map[string]candidate{}
	for _, item := range report.Matched {
		issueMatches[scanKey(moduleVersion{Path: item.Module, Version: item.Version})] = item
	}
	wantIssueFiles := map[string][]string{
		scanKey(moduleVersion{Path: "github.com/coder/websocket", Version: "v1.8.15"}): {
			"mask_amd64.s",
			"mask_arm64.s",
		},
		scanKey(moduleVersion{Path: "github.com/klauspost/compress", Version: "v1.20.0"}): {
			"huff0/decompress_amd64.s",
			"huff0/decompress_arm64.s",
			"internal/cpuinfo/cpuinfo_amd64.s",
			"s2/decode_amd64.s",
			"s2/decode_arm64.s",
			"s2/encodeblock_amd64.s",
			"s2/encodeblock_arm64.s",
			"zstd/fse_decoder_amd64.s",
			"zstd/fse_decoder_arm64.s",
			"zstd/internal/xxhash/xxhash_amd64.s",
			"zstd/internal/xxhash/xxhash_arm64.s",
			"zstd/matchlen_amd64.s",
			"zstd/seqdec_amd64.s",
			"zstd/seqdec_arm64.s",
		},
		scanKey(moduleVersion{Path: "github.com/tmthrgd/go-hex", Version: "v0.0.0-20190904060850-447a3041c3bc"}): {
			"hex_decode_amd64.s",
			"hex_encode_amd64.s",
		},
	}
	for exactKey, wantFiles := range wantIssueFiles {
		match, ok := issueMatches[exactKey]
		if !ok {
			t.Errorf("issue library version was not independently discovered: %s", strings.Replace(exactKey, "\x00", "@", 1))
			continue
		}
		if !reflect.DeepEqual(match.AsmFiles, wantFiles) {
			t.Errorf("discovered assembly for %s = %#v, want %#v", strings.Replace(exactKey, "\x00", "@", 1), match.AsmFiles, wantFiles)
		}
	}
}

func TestInspectModulesRecordsLatestVersionWhenInspectionFails(t *testing.T) {
	archive := testModuleZip(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/@latest"):
			_, _ = io.WriteString(w, `{"Version":"v1.1.0"}`)
		case strings.HasSuffix(r.URL.Path, "/@v/v1.0.0.zip"):
			serveRangeData(t, w, r, archive)
		case strings.HasSuffix(r.URL.Path, "/@v/v1.1.0.zip"):
			http.Error(w, "missing", http.StatusNotFound)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	matched, failures, scanned, skipped := inspectModules(context.Background(), server.Client(), server.URL, []moduleVersion{
		{Path: "example.com/lib", Version: "v1.0.0"},
	}, nil, 1, int64(len(archive))+1)
	if len(matched) != 0 || len(failures) != 1 {
		t.Fatalf("inspectModules() matched %d and failed %d modules, want 0 and 1", len(matched), len(failures))
	}
	if len(scanned) != 0 {
		t.Fatalf("scanned versions = %#v, want none", scanned)
	}
	if failures[0].Module != "example.com/lib" || failures[0].Version != "v1.1.0" {
		t.Fatalf("failure version = %#v, want exact latest version", failures[0])
	}
	if skipped != 0 {
		t.Fatalf("skipped versions = %d, want 0", skipped)
	}
}

func TestInspectModulesRecordsLatestResolutionFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/@latest") {
			_, _ = io.WriteString(w, `{}`)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	matched, failures, scanned, skipped := inspectModules(context.Background(), server.Client(), server.URL, []moduleVersion{
		{Path: "example.com/lib", Version: "v1.0.0"},
	}, nil, 1, 1<<20)
	if len(matched) != 0 || len(scanned) != 0 || skipped != 0 || len(failures) != 1 {
		t.Fatalf("unexpected result: matched=%#v failures=%#v scanned=%#v skipped=%d", matched, failures, scanned, skipped)
	}
	if failures[0].Version != "@latest" || failures[0].Error != "decode @latest: missing Version" {
		t.Fatalf("resolution failure = %#v", failures[0])
	}
}

func TestInspectModulesSkipsPreviouslyCompletedLatestVersion(t *testing.T) {
	archive := testModuleZip(t)
	var latestZipRequests atomic.Int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/@latest"):
			_, _ = io.WriteString(w, `{"Version":"v1.1.0"}`)
		case strings.HasSuffix(r.URL.Path, "/@v/v1.0.0.zip"):
			serveRangeData(t, w, r, archive)
		case strings.HasSuffix(r.URL.Path, "/@v/v1.1.0.zip"):
			latestZipRequests.Add(1)
			serveRangeData(t, w, r, archive)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	latest := moduleVersion{Path: "example.com/lib", Version: "v1.1.0"}
	priorMatch := candidate{
		Module: latest.Path, Version: latest.Version,
		Architectures: []string{"amd64"}, AsmFiles: []string{"asm_amd64.s"},
	}
	matched, failures, scanned, skipped := inspectModules(context.Background(), server.Client(), server.URL, []moduleVersion{
		{Path: "example.com/lib", Version: "v1.0.0"},
	}, map[string]seenResult{scanKey(latest): {Match: priorMatch}}, 1, int64(len(archive))+1)
	if !reflect.DeepEqual(matched, []candidate{priorMatch}) || len(failures) != 0 {
		t.Fatalf("inspectModules() matched %#v and failed %d modules, want reused match", matched, len(failures))
	}
	wantScanned := []moduleVersion{latest}
	if !reflect.DeepEqual(scanned, wantScanned) {
		t.Fatalf("scanned versions = %#v, want reused latest checkpoint %#v", scanned, wantScanned)
	}
	if got := latestZipRequests.Load(); got != 0 {
		t.Fatalf("previously completed latest version fetched %d times, want 0", got)
	}
	if skipped != 1 {
		t.Fatalf("skipped versions = %d, want 1", skipped)
	}
}

func TestInspectModulesChecksLatestWhenIndexedVersionHasNoAssembly(t *testing.T) {
	var noAsm bytes.Buffer
	zw := zip.NewWriter(&noAsm)
	file, err := zw.Create("example.com/lib@v1.0.0/lib.go")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := file.Write([]byte("package lib\n")); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	latestArchive := testModuleZip(t)
	var indexedZipRequests atomic.Int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/@latest"):
			_, _ = io.WriteString(w, `{"Version":"v1.1.0"}`)
		case strings.HasSuffix(r.URL.Path, "/@v/v1.0.0.zip"):
			indexedZipRequests.Add(1)
			serveRangeData(t, w, r, noAsm.Bytes())
		case strings.HasSuffix(r.URL.Path, "/@v/v1.1.0.zip"):
			serveRangeData(t, w, r, latestArchive)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	matched, failures, scanned, skipped := inspectModules(context.Background(), server.Client(), server.URL, []moduleVersion{
		{Path: "example.com/lib", Version: "v1.0.0"},
	}, nil, 1, int64(len(latestArchive))+1)
	if len(failures) != 0 {
		t.Fatalf("inspectModules() failures = %#v", failures)
	}
	wantScanned := []moduleVersion{{Path: "example.com/lib", Version: "v1.1.0"}}
	if !reflect.DeepEqual(scanned, wantScanned) {
		t.Fatalf("scanned versions = %#v, want %#v", scanned, wantScanned)
	}
	if len(matched) != 1 || matched[0].Module != "example.com/lib" || matched[0].Version != "v1.1.0" {
		t.Fatalf("matched versions = %#v, want latest module version", matched)
	}
	if got := indexedZipRequests.Load(); got != 0 {
		t.Fatalf("indexed archive fetched %d times, want 0; index should only discover module paths", got)
	}
	if skipped != 0 {
		t.Fatalf("skipped versions = %d, want 0", skipped)
	}
}

func TestDiscoverDoesNotTreatAnIndexedVersionAsTheLatestCheckpoint(t *testing.T) {
	archive := testModuleZip(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/index":
			_ = json.NewEncoder(w).Encode(indexEntry{
				Path:      "example.com/lib",
				Version:   "v1.0.0",
				Timestamp: time.Date(2019, 4, 10, 0, 0, 0, 0, time.UTC),
			})
		case strings.HasSuffix(r.URL.Path, "/@latest"):
			_, _ = io.WriteString(w, `{"Version":"v1.1.0"}`)
		case strings.HasSuffix(r.URL.Path, "/@v/v1.1.0.zip"):
			serveRangeData(t, w, r, archive)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	report, err := discover(context.Background(), server.Client(), config{
		indexURL:    server.URL + "/index",
		proxyURL:    server.URL,
		limit:       1,
		workers:     1,
		maxZipSize:  int64(len(archive)) + 1,
		httpTimeout: time.Second,
		seen: map[string]seenResult{
			scanKey(moduleVersion{Path: "example.com/lib", Version: "v1.0.0"}): {},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	want := []moduleVersion{{Path: "example.com/lib", Version: "v1.1.0"}}
	if !reflect.DeepEqual(report.Scanned, want) {
		t.Fatalf("scanned versions = %#v, want latest %#v", report.Scanned, want)
	}
}

func TestInspectModuleZipFindsTargetAssembly(t *testing.T) {
	var archive bytes.Buffer
	zw := zip.NewWriter(&archive)
	for _, name := range []string{
		"example.com/lib@v1.2.3/hash/hash.go",
		"example.com/lib@v1.2.3/hash/hash_amd64.s",
		"example.com/lib@v1.2.3/hash/hash_arm64.s",
		"example.com/lib@v1.2.3/hash/empty_arm.s",
		"example.com/lib@v1.2.3/hash/comments_386.s",
		"example.com/lib@v1.2.3/cpu/cpu.go",
		"example.com/lib@v1.2.3/cpu/asm.s",
		"example.com/lib@v1.2.3/testdata/rejected_arm.s",
		"example.com/lib@v1.2.3/hash/not-go-assembly.S",
		"example.com/lib@v1.2.3/README.md",
	} {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(name, "empty_arm.s") {
			continue
		}
		if strings.Contains(name, "comments_386.s") {
			if _, err := w.Write([]byte("//go:build 386\n\n/* license only */\n")); err != nil {
				t.Fatal(err)
			}
			continue
		}
		if _, err := w.Write([]byte("TEXT ·f(SB),0,$0-0\n")); err != nil {
			t.Fatal(err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}

	files, arches, err := inspectModuleZip(archive.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	wantFiles := []string{"cpu/asm.s", "hash/hash_amd64.s", "hash/hash_arm64.s"}
	if !reflect.DeepEqual(files, wantFiles) {
		t.Fatalf("assembly files = %#v, want %#v", files, wantFiles)
	}
	wantArches := []string{"amd64", "arm64", "unknown"}
	if !reflect.DeepEqual(arches, wantArches) {
		t.Fatalf("architectures = %#v, want %#v", arches, wantArches)
	}
}

func TestInferAssemblyArchitecture(t *testing.T) {
	tests := map[string]string{
		"foo_386.s":           "386",
		"foo_amd64.s":         "amd64",
		"foo_arm.s":           "arm",
		"foo_arm64.s":         "arm64",
		"foo_loong64.s":       "loong64",
		"foo_mips64x.s":       "unknown",
		"foo_ppc64le.s":       "ppc64le",
		"foo_riscv64.s":       "riscv64",
		"foo_s390x.s":         "s390x",
		"foo_wasm.s":          "wasm",
		"asm_darwin_x86_gc.s": "unknown",
	}
	for name, want := range tests {
		if got := inferAssemblyArchitecture(name); got != want {
			t.Errorf("inferAssemblyArchitecture(%q) = %q, want %q", name, got, want)
		}
	}
}

func TestSortDiscoveryReportRebuildsArchitectureHintsFromAllAssemblyPaths(t *testing.T) {
	report := discoveryReport{Matched: []candidate{{
		Module:        "example.com/future",
		Version:       "v1.0.0",
		Architectures: []string{"amd64"},
		AsmFiles: []string{
			"asm_amd64.s",
			"asm_riscv64.s",
			"portable.s",
		},
	}}}
	sortDiscoveryReport(&report)
	want := []string{"amd64", "riscv64", "unknown"}
	if got := report.Matched[0].Architectures; !reflect.DeepEqual(got, want) {
		t.Fatalf("architecture hints = %#v, want %#v", got, want)
	}
}

func TestHTTPReaderAtFetchesCentralDirectoryOutsideCachedTail(t *testing.T) {
	archive := testModuleZip(t)
	server := newRangeServer(t, archive, nil)
	defer server.Close()

	reader, size, err := newHTTPReaderAtWithTail(context.Background(), server.Client(), server.URL, int64(len(archive)), 22)
	if err != nil {
		t.Fatal(err)
	}
	files, arches, err := inspectModuleZipReader(reader, size)
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"pkg/asm_amd64.s"}; !reflect.DeepEqual(files, want) {
		t.Fatalf("assembly files = %#v, want %#v", files, want)
	}
	if want := []string{"amd64"}; !reflect.DeepEqual(arches, want) {
		t.Fatalf("architectures = %#v, want %#v", arches, want)
	}
}

func TestHTTPReaderAtDoesNotPrefetchLargeModuleTail(t *testing.T) {
	var archive bytes.Buffer
	zw := zip.NewWriter(&archive)
	large, err := zw.CreateHeader(&zip.FileHeader{
		Name:   "example.com/lib@v1.0.0/pkg/blob.bin",
		Method: zip.Store,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := io.CopyN(large, bytes.NewReader(make([]byte, 9<<20)), 9<<20); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{
		"example.com/lib@v1.0.0/pkg/pkg.go",
		"example.com/lib@v1.0.0/pkg/asm_amd64.s",
	} {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write([]byte("TEXT ·f(SB),0,$0-0\n")); err != nil {
			t.Fatal(err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}

	var fetched atomic.Int64
	var malformedRange atomic.Bool
	server := newRangeServer(t, archive.Bytes(), func(_ http.ResponseWriter, r *http.Request) bool {
		if r.Method != http.MethodGet {
			return false
		}
		if r.Header.Get("Range") == "" {
			fetched.Add(int64(archive.Len()))
			return false
		}
		var start, end int64
		if fields, err := fmt.Sscanf(strings.TrimPrefix(r.Header.Get("Range"), "bytes="), "%d-%d", &start, &end); err != nil || fields != 2 {
			malformedRange.Store(true)
			return false
		}
		fetched.Add(end - start + 1)
		return false
	})
	defer server.Close()

	reader, size, err := newHTTPReaderAt(context.Background(), server.Client(), server.URL, int64(archive.Len()))
	if err != nil {
		t.Fatal(err)
	}
	files, _, err := inspectModuleZipReader(reader, size)
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"pkg/asm_amd64.s"}; !reflect.DeepEqual(files, want) {
		t.Fatalf("assembly files = %#v, want %#v", files, want)
	}
	if malformedRange.Load() {
		t.Fatal("HTTP reader sent a malformed Range header")
	}
	if got, limit := fetched.Load(), int64(2<<20); got > limit {
		t.Fatalf("fetched %d bytes, want at most %d", got, limit)
	}
}

func TestHTTPReaderAtRetriesTransientResponses(t *testing.T) {
	archive := testModuleZip(t)
	var heads atomic.Int32
	server := newRangeServer(t, archive, func(w http.ResponseWriter, r *http.Request) bool {
		if r.Method == http.MethodHead && heads.Add(1) == 1 {
			http.Error(w, "try again", http.StatusServiceUnavailable)
			return true
		}
		return false
	})
	defer server.Close()

	reader, size, err := newHTTPReaderAtWithTail(context.Background(), server.Client(), server.URL, int64(len(archive)), 22)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := inspectModuleZipReader(reader, size); err != nil {
		t.Fatal(err)
	}
	if got := heads.Load(); got != 2 {
		t.Fatalf("HEAD requests = %d, want 2", got)
	}
}

func TestHTTPReaderAtReusesReadAheadForAssemblyEntries(t *testing.T) {
	var archive bytes.Buffer
	zw := zip.NewWriter(&archive)
	for _, name := range []string{
		"example.com/lib@v1.0.0/pkg/pkg.go",
		"example.com/lib@v1.0.0/pkg/first_amd64.s",
		"example.com/lib@v1.0.0/pkg/second_amd64.s",
	} {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write([]byte("TEXT ·f(SB),0,$0-0\n")); err != nil {
			t.Fatal(err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}

	var ranges atomic.Int32
	server := newRangeServer(t, archive.Bytes(), func(_ http.ResponseWriter, r *http.Request) bool {
		if r.Method == http.MethodGet && r.Header.Get("Range") != "" {
			ranges.Add(1)
		}
		return false
	})
	defer server.Close()
	reader, size, err := newHTTPReaderAtWithTail(context.Background(), server.Client(), server.URL, int64(archive.Len()), 22)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := inspectModuleZipReader(reader, size); err != nil {
		t.Fatal(err)
	}
	if got := ranges.Load(); got > 2 {
		t.Fatalf("range requests = %d, want at most 2 (tail plus one read-ahead block)", got)
	}
}

func TestReadIndexDoesNotSkipSharedBoundaryTimestamp(t *testing.T) {
	base := time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC)
	entries := make([]indexEntry, 2001)
	for i := range entries {
		timestamp := base.Add(time.Duration(i) * time.Nanosecond)
		if i == 2000 {
			timestamp = entries[1999].Timestamp
		}
		entries[i] = indexEntry{Path: fmt.Sprintf("example.com/mod%d", i), Version: "v1.0.0", Timestamp: timestamp}
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
		if err != nil {
			t.Fatal(err)
		}
		var since time.Time
		if raw := r.URL.Query().Get("since"); raw != "" {
			since, err = time.Parse(time.RFC3339Nano, raw)
			if err != nil {
				t.Fatal(err)
			}
		}
		enc := json.NewEncoder(w)
		n := 0
		for _, entry := range entries {
			if entry.Timestamp.Before(since) {
				continue
			}
			if n == limit {
				break
			}
			if err := enc.Encode(entry); err != nil {
				t.Fatal(err)
			}
			n++
		}
	}))
	defer server.Close()

	got, nextSince, err := readIndex(context.Background(), server.Client(), server.URL, "", len(entries))
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != len(entries) {
		t.Fatalf("readIndex() returned %d entries, want %d", len(got), len(entries))
	}
	if want := entries[len(entries)-1].Timestamp.Format(time.RFC3339Nano); nextSince != want {
		t.Fatalf("next since = %q, want %q", nextSince, want)
	}
}

func testModuleZip(t *testing.T) []byte {
	t.Helper()
	var archive bytes.Buffer
	zw := zip.NewWriter(&archive)
	for _, name := range []string{
		"example.com/lib@v1.0.0/pkg/pkg.go",
		"example.com/lib@v1.0.0/pkg/asm_amd64.s",
	} {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write([]byte("TEXT ·f(SB),0,$0-0\n")); err != nil {
			t.Fatal(err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return archive.Bytes()
}

func newRangeServer(t *testing.T, data []byte, intercept func(http.ResponseWriter, *http.Request) bool) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if intercept != nil && intercept(w, r) {
			return
		}
		serveRangeData(t, w, r, data)
	}))
}

func serveRangeData(t *testing.T, w http.ResponseWriter, r *http.Request, data []byte) {
	t.Helper()
	if r.Method == http.MethodHead {
		w.Header().Set("Content-Length", strconv.Itoa(len(data)))
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	rangeHeader := r.Header.Get("Range")
	if rangeHeader == "" {
		w.Header().Set("Content-Length", strconv.Itoa(len(data)))
		_, _ = w.Write(data)
		return
	}
	var start, end int
	if _, err := fmt.Sscanf(strings.TrimPrefix(rangeHeader, "bytes="), "%d-%d", &start, &end); err != nil || start < 0 || end < start || end >= len(data) {
		http.Error(w, "bad range", http.StatusRequestedRangeNotSatisfiable)
		return
	}
	w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, len(data)))
	w.Header().Set("Content-Length", strconv.Itoa(end-start+1))
	w.WriteHeader(http.StatusPartialContent)
	_, _ = w.Write(data[start : end+1])
}
