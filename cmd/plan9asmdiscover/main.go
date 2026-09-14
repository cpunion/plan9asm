// Command plan9asmdiscover finds public Go modules that contain Plan 9
// assembly. It reads the official module index, resolves each distinct module
// to @latest, and inspects the cached module ZIP without executing module code.
package main

import (
	"archive/zip"
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/xgo-dev/plan9asm/internal/discoverymeta"
	"golang.org/x/mod/semver"
)

const (
	defaultIndexURL    = "https://index.golang.org/index"
	defaultProxyURL    = "https://proxy.golang.org/cached-only"
	discoverySchema    = 2
	shardedFormat      = "plan9asmdiscover-jsonl-v1"
	indexPageLimit     = 2000
	defaultScanLimit   = 2000
	defaultMaxZipSize  = 2 << 30
	defaultHTTPTimeout = 60 * time.Second
	// ZIP readers need the final 65,557 bytes to locate the end record when
	// the archive uses the maximum-length comment. Earlier directory data is
	// fetched lazily through ReaderAt.
	zipDirectoryTail = 128 << 10
	maxAsmProbeSize  = 8 << 20
	zipReadAhead     = 1 << 20
	httpAttempts     = 3
	httpRetryDelay   = 100 * time.Millisecond
)

type indexEntry struct {
	Path      string    `json:"Path"`
	Version   string    `json:"Version"`
	Timestamp time.Time `json:"Timestamp"`
}

type latestInfo struct {
	Version string `json:"Version"`
}

type moduleVersion struct {
	Path    string `json:"module"`
	Version string `json:"version"`
}

type candidate struct {
	Module        string   `json:"module"`
	Version       string   `json:"version"`
	Architectures []string `json:"architectures"`
	AsmFiles      []string `json:"asm_files"`
}

type scanFailure struct {
	Module  string `json:"module"`
	Version string `json:"version"`
	Error   string `json:"error"`
}

type seenResult struct {
	Match candidate
}

type discoveryReport struct {
	SchemaVersion int             `json:"schema_version"`
	GeneratedAt   time.Time       `json:"generated_at"`
	Since         string          `json:"since,omitempty"`
	NextSince     string          `json:"next_since,omitempty"`
	IndexEntries  int             `json:"index_entries"`
	UniqueModules int             `json:"unique_modules"`
	Skipped       int             `json:"skipped_previously_scanned,omitempty"`
	Scanned       []moduleVersion `json:"scanned"`
	Matched       []candidate     `json:"matched"`
	Failures      []scanFailure   `json:"failures,omitempty"`
}

type shardedManifest struct {
	SchemaVersion int       `json:"schema_version"`
	Format        string    `json:"format"`
	GeneratedAt   time.Time `json:"generated_at"`
	Since         string    `json:"since,omitempty"`
	NextSince     string    `json:"next_since,omitempty"`
	IndexEntries  int       `json:"index_entries"`
	UniqueModules int       `json:"unique_modules"`
	Skipped       int       `json:"skipped_previously_scanned,omitempty"`
	Scanned       int       `json:"scanned_records"`
	Matched       int       `json:"matched_records"`
	Failures      int       `json:"failure_records"`
	ShardKey      string    `json:"shard_key"`
}

type shardedRecord struct {
	Kind          string   `json:"kind"`
	Module        string   `json:"module"`
	Version       string   `json:"version"`
	Architectures []string `json:"architectures,omitempty"`
	AsmFiles      []string `json:"asm_files,omitempty"`
	Error         string   `json:"error,omitempty"`
}

type config struct {
	indexURL    string
	proxyURL    string
	since       string
	limit       int
	workers     int
	maxZipSize  int64
	httpTimeout time.Duration
	seen        map[string]seenResult
}

type pathListFlag []string

func (f *pathListFlag) String() string { return strings.Join(*f, ",") }

func (f *pathListFlag) Set(value string) error {
	if strings.TrimSpace(value) == "" {
		return errors.New("seen report path must not be empty")
	}
	*f = append(*f, value)
	return nil
}

func main() {
	indexURL := flag.String("index-url", defaultIndexURL, "Go module index endpoint")
	proxyURL := flag.String("proxy-url", defaultProxyURL, "Go module proxy endpoint")
	since := flag.String("since", "", "oldest index timestamp (RFC3339); empty starts at the beginning")
	limit := flag.Int("limit", defaultScanLimit, "maximum index records to inspect; 0 scans until the feed ends")
	workers := flag.Int("workers", 8, "concurrent module ZIP inspections")
	maxZipSize := flag.Int64("max-zip-bytes", defaultMaxZipSize, "maximum remote module ZIP size considered")
	httpTimeout := flag.Duration("http-timeout", defaultHTTPTimeout, "timeout for each index or proxy request")
	out := flag.String("out", "", "write JSON to this file instead of stdout")
	outDir := flag.String("out-dir", "", "create or update a diff-friendly sharded JSONL ledger directory")
	convertReport := flag.String("convert-report", "", "convert an existing JSON, JSON.gz, or sharded report into -out-dir")
	var seenReports pathListFlag
	flag.Var(&seenReports, "seen-report", "prior discovery JSON, JSON.gz, or sharded report directory to skip (repeatable)")
	flag.Parse()
	if *out != "" && *outDir != "" {
		fatal(errors.New("-out and -out-dir are mutually exclusive"))
	}
	if *convertReport != "" {
		if *outDir == "" {
			fatal(errors.New("-convert-report requires -out-dir"))
		}
		report, err := readSeenReport(*convertReport)
		if err != nil {
			fatal(err)
		}
		if err := writeShardedReport(*outDir, report); err != nil {
			fatal(err)
		}
		return
	}
	if *outDir != "" {
		if info, err := os.Stat(*outDir); err == nil {
			if !info.IsDir() {
				fatal(fmt.Errorf("-out-dir %q is not a directory", *outDir))
			}
			// Resolving @latest still happens for every module. The existing
			// ledger only avoids downloading an exact version already recorded.
			seenReports = append(seenReports, *outDir)
		} else if !errors.Is(err, os.ErrNotExist) {
			fatal(fmt.Errorf("read -out-dir %q: %w", *outDir, err))
		}
	}
	seen, err := loadSeenReports(seenReports)
	if err != nil {
		fatal(err)
	}

	cfg := config{
		indexURL:    *indexURL,
		proxyURL:    strings.TrimRight(*proxyURL, "/"),
		since:       *since,
		limit:       *limit,
		workers:     *workers,
		maxZipSize:  *maxZipSize,
		httpTimeout: *httpTimeout,
		seen:        seen,
	}
	if err := validateConfig(cfg); err != nil {
		fatal(err)
	}

	ctx := context.Background()
	client := &http.Client{Timeout: cfg.httpTimeout}
	report, err := discover(ctx, client, cfg)
	if err != nil {
		fatal(err)
	}
	if *outDir != "" {
		if err := writeShardedReport(*outDir, report); err != nil {
			fatal(err)
		}
		return
	}
	var writer io.Writer = os.Stdout
	if *out != "" {
		file, err := os.Create(*out)
		if err != nil {
			fatal(err)
		}
		defer file.Close()
		writer = file
	}
	enc := json.NewEncoder(writer)
	enc.SetIndent("", "  ")
	if err := enc.Encode(report); err != nil {
		fatal(err)
	}
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

func validateConfig(cfg config) error {
	if _, err := url.ParseRequestURI(cfg.indexURL); err != nil {
		return fmt.Errorf("invalid index URL: %w", err)
	}
	if _, err := url.ParseRequestURI(cfg.proxyURL); err != nil {
		return fmt.Errorf("invalid proxy URL: %w", err)
	}
	if cfg.since != "" {
		if _, err := time.Parse(time.RFC3339Nano, cfg.since); err != nil {
			return fmt.Errorf("invalid -since timestamp: %w", err)
		}
	}
	if cfg.limit < 0 {
		return errors.New("-limit must be nonnegative")
	}
	if cfg.workers <= 0 {
		return errors.New("-workers must be positive")
	}
	if cfg.maxZipSize <= 0 {
		return errors.New("-max-zip-bytes must be positive")
	}
	if cfg.httpTimeout <= 0 {
		return errors.New("-http-timeout must be positive")
	}
	return nil
}

func discover(ctx context.Context, client *http.Client, cfg config) (discoveryReport, error) {
	entries, nextSince, err := readIndex(ctx, client, cfg.indexURL, cfg.since, cfg.limit)
	if err != nil {
		return discoveryReport{}, err
	}
	moduleSet := make(map[string]string, len(entries))
	for _, entry := range entries {
		moduleSet[entry.Path] = entry.Version
	}
	modules := make([]moduleVersion, 0, len(moduleSet))
	for modulePath, version := range moduleSet {
		modules = append(modules, moduleVersion{Path: modulePath, Version: version})
	}
	sort.Slice(modules, func(i, j int) bool { return modules[i].Path < modules[j].Path })
	uniqueModules := len(modules)

	matched, failures, scanned, skipped := inspectModules(ctx, client, cfg.proxyURL, modules, cfg.seen, cfg.workers, cfg.maxZipSize)
	return discoveryReport{
		SchemaVersion: discoverySchema,
		GeneratedAt:   time.Now().UTC(),
		Since:         cfg.since,
		NextSince:     nextSince,
		IndexEntries:  len(entries),
		UniqueModules: uniqueModules,
		Skipped:       skipped,
		Scanned:       scanned,
		Matched:       matched,
		Failures:      failures,
	}, nil
}

func loadSeenReports(paths []string) (map[string]seenResult, error) {
	seen := make(map[string]seenResult)
	for _, reportPath := range paths {
		report, err := readSeenReport(reportPath)
		if err != nil {
			return nil, err
		}
		for _, item := range report.Scanned {
			if item.Path == "" || item.Version == "" {
				return nil, fmt.Errorf("seen report %q contains an incomplete scanned version", reportPath)
			}
			key := scanKey(item)
			if _, ok := seen[key]; !ok {
				seen[key] = seenResult{}
			}
		}
		// Older reports did not contain the complete scanned set, but their
		// successful matches are still safe to reuse.
		for _, item := range report.Matched {
			if item.Module != "" && item.Version != "" {
				seen[scanKey(moduleVersion{Path: item.Module, Version: item.Version})] = seenResult{Match: item}
			}
		}
	}
	return seen, nil
}

func readSeenReport(reportPath string) (discoveryReport, error) {
	info, err := os.Stat(reportPath)
	if err != nil {
		return discoveryReport{}, fmt.Errorf("read seen report %q: %w", reportPath, err)
	}
	if info.IsDir() {
		return readShardedReports(reportPath)
	}
	data, err := os.ReadFile(reportPath)
	if err != nil {
		return discoveryReport{}, fmt.Errorf("read seen report %q: %w", reportPath, err)
	}
	if bytes.HasPrefix(data, []byte{0x1f, 0x8b}) {
		reader, err := gzip.NewReader(bytes.NewReader(data))
		if err != nil {
			return discoveryReport{}, fmt.Errorf("decompress seen report %q: %w", reportPath, err)
		}
		data, err = io.ReadAll(reader)
		_ = reader.Close()
		if err != nil {
			return discoveryReport{}, fmt.Errorf("decompress seen report %q: %w", reportPath, err)
		}
	}
	var report discoveryReport
	if err := json.Unmarshal(data, &report); err != nil {
		return discoveryReport{}, fmt.Errorf("decode seen report %q: %w", reportPath, err)
	}
	return report, nil
}

func writeShardedReport(reportPath string, report discoveryReport) error {
	exists := false
	if info, err := os.Stat(reportPath); err == nil {
		if !info.IsDir() {
			return fmt.Errorf("write sharded report %q: path already exists and is not a directory", reportPath)
		}
		existing, err := readShardedReport(reportPath)
		if err != nil {
			return fmt.Errorf("update sharded report %q: %w", reportPath, err)
		}
		report, err = mergeDiscoveryReports(existing, report)
		if err != nil {
			return fmt.Errorf("update sharded report %q: %w", reportPath, err)
		}
		exists = true
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("write sharded report %q: %w", reportPath, err)
	} else {
		var mergeErr error
		report, mergeErr = mergeDiscoveryReports(discoveryReport{}, report)
		if mergeErr != nil {
			return fmt.Errorf("write sharded report %q: %w", reportPath, mergeErr)
		}
	}

	parent := filepath.Dir(reportPath)
	if err := os.MkdirAll(parent, 0o755); err != nil {
		return fmt.Errorf("create sharded report parent %q: %w", parent, err)
	}
	temporary, err := os.MkdirTemp(parent, ".plan9asmdiscover-*")
	if err != nil {
		return fmt.Errorf("create temporary sharded report: %w", err)
	}
	defer os.RemoveAll(temporary)

	records := make(map[byte][]shardedRecord)
	add := func(record shardedRecord) {
		hash := sha256.Sum256([]byte(record.Module))
		records[hash[0]] = append(records[hash[0]], record)
	}
	for _, item := range report.Scanned {
		add(shardedRecord{Kind: "scanned", Module: item.Path, Version: item.Version})
	}
	for _, item := range report.Matched {
		add(shardedRecord{
			Kind:          "matched",
			Module:        item.Module,
			Version:       item.Version,
			Architectures: item.Architectures,
			AsmFiles:      item.AsmFiles,
		})
	}
	for _, item := range report.Failures {
		add(shardedRecord{Kind: "failure", Module: item.Module, Version: item.Version, Error: item.Error})
	}

	recordsDir := filepath.Join(temporary, "records")
	if err := os.Mkdir(recordsDir, 0o755); err != nil {
		return fmt.Errorf("create record shard directory: %w", err)
	}
	shards := make([]int, 0, len(records))
	for shard := range records {
		shards = append(shards, int(shard))
	}
	sort.Ints(shards)
	for _, shard := range shards {
		items := records[byte(shard)]
		sort.Slice(items, func(i, j int) bool {
			return compareShardedRecords(items[i], items[j]) < 0
		})
		name := filepath.Join(recordsDir, fmt.Sprintf("%02x.jsonl", shard))
		file, err := os.OpenFile(name, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
		if err != nil {
			return fmt.Errorf("create record shard %q: %w", name, err)
		}
		encoder := json.NewEncoder(file)
		for _, item := range items {
			if err := encoder.Encode(item); err != nil {
				file.Close()
				return fmt.Errorf("write record shard %q: %w", name, err)
			}
		}
		if err := file.Close(); err != nil {
			return fmt.Errorf("close record shard %q: %w", name, err)
		}
	}

	manifest := shardedManifest{
		SchemaVersion: report.SchemaVersion,
		Format:        shardedFormat,
		GeneratedAt:   report.GeneratedAt,
		Since:         report.Since,
		NextSince:     report.NextSince,
		IndexEntries:  report.IndexEntries,
		UniqueModules: report.UniqueModules,
		Skipped:       report.Skipped,
		Scanned:       len(report.Scanned),
		Matched:       len(report.Matched),
		Failures:      len(report.Failures),
		ShardKey:      "sha256(module)[0]",
	}
	manifestFile, err := os.OpenFile(filepath.Join(temporary, "manifest.json"), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("create sharded report manifest: %w", err)
	}
	encoder := json.NewEncoder(manifestFile)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(manifest); err != nil {
		manifestFile.Close()
		return fmt.Errorf("write sharded report manifest: %w", err)
	}
	if err := manifestFile.Close(); err != nil {
		return fmt.Errorf("close sharded report manifest: %w", err)
	}
	if !exists {
		if err := os.Rename(temporary, reportPath); err != nil {
			return fmt.Errorf("publish sharded report %q: %w", reportPath, err)
		}
		return nil
	}
	backup := temporary + "-previous"
	if err := os.Rename(reportPath, backup); err != nil {
		return fmt.Errorf("replace sharded report %q: preserve previous ledger: %w", reportPath, err)
	}
	if err := os.Rename(temporary, reportPath); err != nil {
		if restoreErr := os.Rename(backup, reportPath); restoreErr != nil {
			return fmt.Errorf("replace sharded report %q: %v (also failed to restore previous ledger: %v)", reportPath, err, restoreErr)
		}
		return fmt.Errorf("replace sharded report %q: %w", reportPath, err)
	}
	if err := os.RemoveAll(backup); err != nil {
		return fmt.Errorf("remove previous sharded report %q after replacement: %w", backup, err)
	}
	return nil
}

func mergeDiscoveryReports(base, update discoveryReport) (discoveryReport, error) {
	if base.SchemaVersion != 0 && update.SchemaVersion != 0 && base.SchemaVersion != update.SchemaVersion {
		return discoveryReport{}, fmt.Errorf("cannot merge schema versions %d and %d", base.SchemaVersion, update.SchemaVersion)
	}
	merged := discoveryReport{SchemaVersion: base.SchemaVersion}
	if merged.SchemaVersion == 0 {
		merged.SchemaVersion = update.SchemaVersion
	}
	if merged.SchemaVersion == 0 {
		merged.SchemaVersion = discoverySchema
	}
	if merged.SchemaVersion != discoverySchema {
		return discoveryReport{}, fmt.Errorf("unsupported discovery schema %d", merged.SchemaVersion)
	}
	merged.GeneratedAt = base.GeneratedAt
	if update.GeneratedAt.After(merged.GeneratedAt) {
		merged.GeneratedAt = update.GeneratedAt
	}
	merged.Since = earlierTimestamp(base.Since, update.Since)
	merged.NextSince = laterTimestamp(base.NextSince, update.NextSince)
	overlaps := discoveryRangesOverlap(base.Since, base.NextSince, update.Since, update.NextSince)
	merged.IndexEntries = mergeRunCount(base.IndexEntries, update.IndexEntries, overlaps)
	merged.Skipped = mergeRunCount(base.Skipped, update.Skipped, overlaps)

	scanned := make(map[string]moduleVersion, len(base.Scanned)+len(update.Scanned)+len(base.Matched)+len(update.Matched))
	failures := make(map[string]scanFailure, len(base.Failures)+len(update.Failures))
	type candidateSets struct {
		candidate candidate
		arches    map[string]bool
		asmFiles  map[string]bool
	}
	matched := make(map[string]*candidateSets, len(base.Matched)+len(update.Matched))
	addCandidate := func(item candidate) {
		key := scanKey(moduleVersion{Path: item.Module, Version: item.Version})
		set := matched[key]
		if set == nil {
			set = &candidateSets{
				candidate: candidate{Module: item.Module, Version: item.Version},
				arches:    make(map[string]bool),
				asmFiles:  make(map[string]bool),
			}
			matched[key] = set
		}
		for _, arch := range item.Architectures {
			set.arches[arch] = true
		}
		for _, asmFile := range item.AsmFiles {
			set.asmFiles[asmFile] = true
		}
		scanned[key] = moduleVersion{Path: item.Module, Version: item.Version}
	}
	for _, report := range []discoveryReport{base, update} {
		for _, item := range report.Scanned {
			scanned[scanKey(item)] = item
		}
		for _, item := range report.Matched {
			addCandidate(item)
		}
		for _, item := range report.Failures {
			failures[scanKey(moduleVersion{Path: item.Module, Version: item.Version})] = item
		}
	}
	modules := make(map[string]bool)
	for key, item := range scanned {
		merged.Scanned = append(merged.Scanned, item)
		modules[item.Path] = true
		delete(failures, key)
	}
	for _, set := range matched {
		for arch := range set.arches {
			set.candidate.Architectures = append(set.candidate.Architectures, arch)
		}
		for asmFile := range set.asmFiles {
			set.candidate.AsmFiles = append(set.candidate.AsmFiles, asmFile)
		}
		merged.Matched = append(merged.Matched, set.candidate)
		modules[set.candidate.Module] = true
	}
	for _, item := range failures {
		merged.Failures = append(merged.Failures, item)
		modules[item.Module] = true
	}
	merged.UniqueModules = len(modules)
	sortDiscoveryReport(&merged)
	return merged, nil
}

func discoveryRangesOverlap(aSince, aNext, bSince, bNext string) bool {
	if aSince == "" || aNext == "" || bSince == "" || bNext == "" {
		return false
	}
	return aSince < bNext && bSince < aNext
}

func mergeRunCount(base, update int, overlaps bool) int {
	if overlaps {
		if update > base {
			return update
		}
		return base
	}
	return base + update
}

func readShardedReports(reportPath string) (discoveryReport, error) {
	if _, err := os.Stat(filepath.Join(reportPath, "manifest.json")); err == nil {
		return readShardedReport(reportPath)
	} else if !errors.Is(err, os.ErrNotExist) {
		return discoveryReport{}, fmt.Errorf("read sharded report %q: %w", reportPath, err)
	}
	// Accept the former per-run tree only as an import source. Newly written
	// ledgers always have one manifest and one records directory at the root.
	var manifests []string
	err := filepath.WalkDir(reportPath, func(name string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() && entry.Name() == "manifest.json" {
			manifests = append(manifests, name)
		}
		return nil
	})
	if err != nil {
		return discoveryReport{}, fmt.Errorf("read sharded reports %q: %w", reportPath, err)
	}
	if len(manifests) == 0 {
		return discoveryReport{}, fmt.Errorf("read sharded reports %q: no manifest.json found", reportPath)
	}
	sort.Strings(manifests)
	var combined discoveryReport
	for _, manifestPath := range manifests {
		report, err := readShardedReport(filepath.Dir(manifestPath))
		if err != nil {
			return discoveryReport{}, err
		}
		combined, err = mergeDiscoveryReports(combined, report)
		if err != nil {
			return discoveryReport{}, fmt.Errorf("read sharded reports %q: %w", reportPath, err)
		}
	}
	return combined, nil
}

func earlierTimestamp(current, candidate string) string {
	if current == "" || candidate != "" && candidate < current {
		return candidate
	}
	return current
}

func laterTimestamp(current, candidate string) string {
	if candidate > current {
		return candidate
	}
	return current
}

func readShardedReport(reportPath string) (discoveryReport, error) {
	manifestPath := filepath.Join(reportPath, "manifest.json")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return discoveryReport{}, fmt.Errorf("read sharded report manifest %q: %w", manifestPath, err)
	}
	var manifest shardedManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return discoveryReport{}, fmt.Errorf("decode sharded report manifest %q: %w", manifestPath, err)
	}
	if manifest.Format != shardedFormat {
		return discoveryReport{}, fmt.Errorf("decode sharded report manifest %q: unsupported format %q", manifestPath, manifest.Format)
	}
	if manifest.SchemaVersion != discoverySchema {
		return discoveryReport{}, fmt.Errorf("decode sharded report manifest %q: unsupported schema %d", manifestPath, manifest.SchemaVersion)
	}
	if manifest.ShardKey != "sha256(module)[0]" {
		return discoveryReport{}, fmt.Errorf("decode sharded report manifest %q: unsupported shard key %q", manifestPath, manifest.ShardKey)
	}
	report := discoveryReport{
		SchemaVersion: manifest.SchemaVersion,
		GeneratedAt:   manifest.GeneratedAt,
		Since:         manifest.Since,
		NextSince:     manifest.NextSince,
		IndexEntries:  manifest.IndexEntries,
		UniqueModules: manifest.UniqueModules,
		Skipped:       manifest.Skipped,
	}
	recordsDir := filepath.Join(reportPath, "records")
	entries, err := os.ReadDir(recordsDir)
	if err != nil {
		return discoveryReport{}, fmt.Errorf("list sharded report records %q: %w", reportPath, err)
	}
	shards := make([]string, 0, len(entries))
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || len(name) != len("00.jsonl") || !strings.HasSuffix(name, ".jsonl") {
			return discoveryReport{}, fmt.Errorf("list sharded report records %q: unexpected entry %q", reportPath, name)
		}
		if _, err := strconv.ParseUint(name[:2], 16, 8); err != nil {
			return discoveryReport{}, fmt.Errorf("list sharded report records %q: invalid shard name %q", reportPath, name)
		}
		shards = append(shards, filepath.Join(recordsDir, name))
	}
	sort.Strings(shards)
	for _, shard := range shards {
		shardNumber, _ := strconv.ParseUint(filepath.Base(shard)[:2], 16, 8)
		file, err := os.Open(shard)
		if err != nil {
			return discoveryReport{}, fmt.Errorf("read record shard %q: %w", shard, err)
		}
		scanner := bufio.NewScanner(file)
		scanner.Buffer(make([]byte, 64<<10), 16<<20)
		line := 0
		var previous shardedRecord
		for scanner.Scan() {
			line++
			var item shardedRecord
			if err := json.Unmarshal(scanner.Bytes(), &item); err != nil {
				file.Close()
				return discoveryReport{}, fmt.Errorf("decode record shard %s:%d: %w", shard, line, err)
			}
			if item.Module == "" || item.Version == "" {
				file.Close()
				return discoveryReport{}, fmt.Errorf("decode record shard %s:%d: incomplete module version", shard, line)
			}
			hash := sha256.Sum256([]byte(item.Module))
			if hash[0] != byte(shardNumber) {
				file.Close()
				return discoveryReport{}, fmt.Errorf("decode record shard %s:%d: module %q belongs in %02x.jsonl", shard, line, item.Module, hash[0])
			}
			if line > 1 && compareShardedRecords(previous, item) >= 0 {
				file.Close()
				return discoveryReport{}, fmt.Errorf("decode record shard %s:%d: records are duplicated or not in module/semver/kind order", shard, line)
			}
			previous = item
			switch item.Kind {
			case "scanned":
				if !semver.IsValid(item.Version) {
					file.Close()
					return discoveryReport{}, fmt.Errorf("decode record shard %s:%d: invalid scanned module version %q", shard, line, item.Version)
				}
				report.Scanned = append(report.Scanned, moduleVersion{Path: item.Module, Version: item.Version})
			case "matched":
				if !semver.IsValid(item.Version) {
					file.Close()
					return discoveryReport{}, fmt.Errorf("decode record shard %s:%d: invalid matched module version %q", shard, line, item.Version)
				}
				report.Matched = append(report.Matched, candidate{
					Module: item.Module, Version: item.Version,
					Architectures: item.Architectures, AsmFiles: item.AsmFiles,
				})
			case "failure":
				report.Failures = append(report.Failures, scanFailure{Module: item.Module, Version: item.Version, Error: item.Error})
			default:
				file.Close()
				return discoveryReport{}, fmt.Errorf("decode record shard %s:%d: unknown record kind %q", shard, line, item.Kind)
			}
		}
		scanErr := scanner.Err()
		closeErr := file.Close()
		if scanErr != nil {
			return discoveryReport{}, fmt.Errorf("read record shard %q: %w", shard, scanErr)
		}
		if closeErr != nil {
			return discoveryReport{}, fmt.Errorf("close record shard %q: %w", shard, closeErr)
		}
	}
	if len(report.Scanned) != manifest.Scanned || len(report.Matched) != manifest.Matched || len(report.Failures) != manifest.Failures {
		return discoveryReport{}, fmt.Errorf(
			"read sharded report %q: record counts are scanned=%d/%d matched=%d/%d failures=%d/%d",
			reportPath, len(report.Scanned), manifest.Scanned, len(report.Matched), manifest.Matched, len(report.Failures), manifest.Failures,
		)
	}
	sortDiscoveryReport(&report)
	return report, nil
}

func sortDiscoveryReport(report *discoveryReport) {
	sort.Slice(report.Scanned, func(i, j int) bool {
		if report.Scanned[i].Path != report.Scanned[j].Path {
			return report.Scanned[i].Path < report.Scanned[j].Path
		}
		return compareModuleVersions(report.Scanned[i].Version, report.Scanned[j].Version) < 0
	})
	for i := range report.Matched {
		sort.Strings(report.Matched[i].AsmFiles)
		report.Matched[i].Architectures = discoverymeta.ArchitectureHints(report.Matched[i].AsmFiles)
	}
	sort.Slice(report.Matched, func(i, j int) bool {
		if report.Matched[i].Module != report.Matched[j].Module {
			return report.Matched[i].Module < report.Matched[j].Module
		}
		return compareModuleVersions(report.Matched[i].Version, report.Matched[j].Version) < 0
	})
	sort.Slice(report.Failures, func(i, j int) bool {
		if report.Failures[i].Module != report.Failures[j].Module {
			return report.Failures[i].Module < report.Failures[j].Module
		}
		return compareModuleVersions(report.Failures[i].Version, report.Failures[j].Version) < 0
	})
}

func compareModuleVersions(a, b string) int {
	aValid, bValid := semver.IsValid(a), semver.IsValid(b)
	if aValid && bValid {
		if compared := semver.Compare(a, b); compared != 0 {
			return compared
		}
	} else if aValid != bValid {
		if aValid {
			return -1
		}
		return 1
	}
	return strings.Compare(a, b)
}

func compareShardedRecords(a, b shardedRecord) int {
	if compared := strings.Compare(a.Module, b.Module); compared != 0 {
		return compared
	}
	if compared := compareModuleVersions(a.Version, b.Version); compared != 0 {
		return compared
	}
	return strings.Compare(a.Kind, b.Kind)
}

func scanKey(module moduleVersion) string {
	return module.Path + "\x00" + module.Version
}

func readIndex(ctx context.Context, client *http.Client, endpoint, since string, limit int) ([]indexEntry, string, error) {
	var entries []indexEntry
	cursor := since
	seen := make(map[string]struct{})
	seenAt := make(map[string]int)
	for limit == 0 || len(entries) < limit {
		pageSize := indexPageLimit
		if limit > 0 {
			pageSize = limit - len(entries) + seenAt[cursor]
			if pageSize > indexPageLimit {
				pageSize = indexPageLimit
			}
		}
		parsed, err := url.Parse(endpoint)
		if err != nil {
			return nil, "", err
		}
		query := parsed.Query()
		query.Set("limit", fmt.Sprint(pageSize))
		if cursor != "" {
			query.Set("since", cursor)
		}
		parsed.RawQuery = query.Encode()

		resp, err := doHTTPRequest(ctx, client, http.MethodGet, parsed.String(), nil)
		if err != nil {
			return nil, "", fmt.Errorf("read module index: %w", err)
		}
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return nil, "", fmt.Errorf("read module index: %s", resp.Status)
		}
		scanner := bufio.NewScanner(resp.Body)
		scanner.Buffer(make([]byte, 64<<10), 1<<20)
		page := make([]indexEntry, 0, pageSize)
		for scanner.Scan() {
			var entry indexEntry
			if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
				resp.Body.Close()
				return nil, "", fmt.Errorf("decode module index: %w", err)
			}
			if entry.Path == "" || entry.Timestamp.IsZero() {
				resp.Body.Close()
				return nil, "", errors.New("module index entry is missing Path or Timestamp")
			}
			page = append(page, entry)
		}
		scanErr := scanner.Err()
		resp.Body.Close()
		if scanErr != nil {
			return nil, "", fmt.Errorf("read module index: %w", scanErr)
		}
		if len(page) == 0 {
			break
		}
		newEntries := 0
		for _, entry := range page {
			timestamp := entry.Timestamp.Format(time.RFC3339Nano)
			key := entry.Path + "\x00" + entry.Version + "\x00" + timestamp
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			seenAt[timestamp]++
			entries = append(entries, entry)
			newEntries++
			cursor = timestamp
			if limit > 0 && len(entries) == limit {
				break
			}
		}
		if limit > 0 && len(entries) == limit {
			break
		}
		if len(page) < pageSize {
			break
		}
		lastTimestamp := page[len(page)-1].Timestamp.Format(time.RFC3339Nano)
		if newEntries == 0 && lastTimestamp == cursor {
			cursor = page[len(page)-1].Timestamp.Add(time.Nanosecond).Format(time.RFC3339Nano)
		} else {
			cursor = lastTimestamp
		}
	}
	return entries, cursor, nil
}

func inspectModules(ctx context.Context, client *http.Client, proxyURL string, modules []moduleVersion, seen map[string]seenResult, workers int, maxZipSize int64) ([]candidate, []scanFailure, []moduleVersion, int) {
	tasks := make(chan moduleVersion)
	results := make(chan candidate)
	errorsOut := make(chan scanFailure)
	completed := make(chan moduleVersion)
	reused := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for module := range tasks {
				item, latest, didScan, wasReused, err := inspectLatestModule(ctx, client, proxyURL, module.Path, seen, maxZipSize)
				if didScan || wasReused {
					completed <- latest
				}
				if wasReused {
					reused <- struct{}{}
				}
				if err != nil {
					errorsOut <- scanFailure{Module: latest.Path, Version: latest.Version, Error: err.Error()}
					continue
				}
				if len(item.AsmFiles) > 0 {
					results <- item
				}
			}
		}()
	}
	go func() {
		for _, module := range modules {
			tasks <- module
		}
		close(tasks)
		wg.Wait()
		close(results)
		close(errorsOut)
		close(completed)
		close(reused)
	}()

	matched := make([]candidate, 0)
	failures := make([]scanFailure, 0)
	scannedSet := make(map[string]moduleVersion)
	skipped := 0
	for results != nil || errorsOut != nil || completed != nil || reused != nil {
		select {
		case item, ok := <-results:
			if !ok {
				results = nil
				continue
			}
			matched = append(matched, item)
		case failure, ok := <-errorsOut:
			if !ok {
				errorsOut = nil
				continue
			}
			failures = append(failures, failure)
		case item, ok := <-completed:
			if !ok {
				completed = nil
				continue
			}
			scannedSet[scanKey(item)] = item
		case _, ok := <-reused:
			if !ok {
				reused = nil
				continue
			}
			skipped++
		}
	}
	scanned := make([]moduleVersion, 0, len(scannedSet))
	for _, item := range scannedSet {
		scanned = append(scanned, item)
	}
	sort.Slice(matched, func(i, j int) bool { return matched[i].Module < matched[j].Module })
	sort.Slice(failures, func(i, j int) bool { return failures[i].Module < failures[j].Module })
	sort.Slice(scanned, func(i, j int) bool {
		if scanned[i].Path != scanned[j].Path {
			return scanned[i].Path < scanned[j].Path
		}
		return scanned[i].Version < scanned[j].Version
	})
	return matched, failures, scanned, skipped
}

func inspectLatestModule(ctx context.Context, client *http.Client, proxyURL, modulePath string, seen map[string]seenResult, maxZipSize int64) (candidate, moduleVersion, bool, bool, error) {
	latest := moduleVersion{Path: modulePath, Version: "@latest"}
	latestVersion, err := resolveLatest(ctx, client, proxyURL, modulePath)
	if err != nil {
		return candidate{}, latest, false, false, err
	}
	latest.Version = latestVersion
	if previous, ok := seen[scanKey(latest)]; ok {
		return previous.Match, latest, false, true, nil
	}
	item, err := inspectModuleVersion(ctx, client, proxyURL, modulePath, latestVersion, maxZipSize)
	if err != nil {
		return candidate{}, latest, false, false, err
	}
	return item, latest, true, false, nil
}

func resolveLatest(ctx context.Context, client *http.Client, proxyURL, modulePath string) (string, error) {
	escapedModule, err := escapeProxyPath(modulePath)
	if err != nil {
		return "", err
	}
	latestURL := proxyURL + "/" + escapedModule + "/@latest"
	latestBody, err := get(ctx, client, latestURL, 1<<20)
	if err != nil {
		return "", fmt.Errorf("resolve @latest: %w", err)
	}
	var latest latestInfo
	if err := json.Unmarshal(latestBody, &latest); err != nil {
		return "", fmt.Errorf("decode @latest: %w", err)
	}
	if latest.Version == "" {
		return "", errors.New("decode @latest: missing Version")
	}
	return latest.Version, nil
}

func inspectModuleVersion(ctx context.Context, client *http.Client, proxyURL, modulePath, version string, maxZipSize int64) (candidate, error) {
	escapedModule, err := escapeProxyPath(modulePath)
	if err != nil {
		return candidate{}, err
	}
	escapedVersion, err := escapeProxyPath(version)
	if err != nil {
		return candidate{}, err
	}
	zipURL := proxyURL + "/" + escapedModule + "/@v/" + escapedVersion + ".zip"
	reader, zipSize, err := newHTTPReaderAt(ctx, client, zipURL, maxZipSize)
	if err != nil {
		return candidate{}, fmt.Errorf("open %s ZIP: %w", version, err)
	}
	files, arches, err := inspectModuleZipReader(reader, zipSize)
	if err != nil {
		return candidate{}, fmt.Errorf("inspect %s: %w", version, err)
	}
	return candidate{Module: modulePath, Version: version, Architectures: arches, AsmFiles: files}, nil
}

func get(ctx context.Context, client *http.Client, endpoint string, maxBytes int64) ([]byte, error) {
	resp, err := doHTTPRequest(ctx, client, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s", resp.Status)
	}
	if resp.ContentLength > maxBytes {
		return nil, fmt.Errorf("response is %d bytes (limit %d)", resp.ContentLength, maxBytes)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBytes+1))
	if err != nil {
		return nil, err
	}
	if int64(len(body)) > maxBytes {
		return nil, fmt.Errorf("response exceeds %d bytes", maxBytes)
	}
	return body, nil
}

func inspectModuleZip(data []byte) ([]string, []string, error) {
	return inspectModuleZipReader(bytes.NewReader(data), int64(len(data)))
}

func inspectModuleZipReader(reader io.ReaderAt, size int64) ([]string, []string, error) {
	zr, err := zip.NewReader(reader, size)
	if err != nil {
		return nil, nil, err
	}
	type asmEntry struct {
		name string
		file *zip.File
	}
	asmFiles := make([]asmEntry, 0)
	goDirs := make(map[string]struct{})
	for _, file := range zr.File {
		name := file.Name
		at := strings.IndexByte(name, '@')
		if at < 0 {
			continue
		}
		rootSlash := strings.IndexByte(name[at:], '/')
		if rootSlash < 0 {
			continue
		}
		rootSlash += at
		if rootSlash == len(name)-1 {
			continue
		}
		rel := name[rootSlash+1:]
		if containsPathElement(rel, "testdata") {
			continue
		}
		switch path.Ext(rel) {
		case ".go":
			goDirs[path.Dir(rel)] = struct{}{}
		case ".s":
			if file.UncompressedSize64 > 0 {
				asmFiles = append(asmFiles, asmEntry{name: rel, file: file})
			}
		}
	}
	files := make([]string, 0, len(asmFiles))
	archSet := make(map[string]struct{})
	for _, asm := range asmFiles {
		if _, ok := goDirs[path.Dir(asm.name)]; !ok {
			continue
		}
		hasContent, err := zipAssemblyFileHasContent(asm.file)
		if err != nil {
			return nil, nil, fmt.Errorf("read %s: %w", asm.name, err)
		}
		if !hasContent {
			continue
		}
		files = append(files, asm.name)
		archSet[inferAssemblyArchitecture(asm.name)] = struct{}{}
	}
	arches := make([]string, 0, len(archSet))
	for arch := range archSet {
		arches = append(arches, arch)
	}
	sort.Strings(files)
	sort.Strings(arches)
	return files, arches, nil
}

func zipAssemblyFileHasContent(file *zip.File) (bool, error) {
	if file.UncompressedSize64 > maxAsmProbeSize {
		// Keep unusually large files in the candidate set so the corpus stage can
		// validate them; discovery must not silently discard code it did not read.
		return true, nil
	}
	r, err := file.Open()
	if err != nil {
		return false, err
	}
	defer r.Close()
	data, err := io.ReadAll(io.LimitReader(r, maxAsmProbeSize+1))
	if err != nil {
		return false, err
	}
	if len(data) > maxAsmProbeSize {
		return true, nil
	}
	return assemblySourceHasContent(data), nil
}

func assemblySourceHasContent(data []byte) bool {
	for i := 0; i < len(data); {
		switch {
		case data[i] == ' ' || data[i] == '\t' || data[i] == '\r' || data[i] == '\n' || data[i] == '\f':
			i++
		case i+1 < len(data) && data[i] == '/' && data[i+1] == '/':
			i += 2
			for i < len(data) && data[i] != '\n' {
				i++
			}
		case i+1 < len(data) && data[i] == '/' && data[i+1] == '*':
			end := bytes.Index(data[i+2:], []byte("*/"))
			if end < 0 {
				return true
			}
			i += end + 4
		default:
			return true
		}
	}
	return false
}

type remoteZipReaderAt struct {
	ctx        context.Context
	client     *http.Client
	endpoint   string
	cachedTail []byte
	cacheMu    sync.Mutex
	cachedData map[int64][]byte
	tailStart  int64
	size       int64
}

func newHTTPReaderAt(ctx context.Context, client *http.Client, endpoint string, maxSize int64) (io.ReaderAt, int64, error) {
	return newHTTPReaderAtWithTail(ctx, client, endpoint, maxSize, zipDirectoryTail)
}

func newHTTPReaderAtWithTail(ctx context.Context, client *http.Client, endpoint string, maxSize, tailSize int64) (io.ReaderAt, int64, error) {
	if tailSize <= 0 {
		return nil, 0, errors.New("ZIP tail size must be positive")
	}
	resp, err := doHTTPRequest(ctx, client, http.MethodHead, endpoint, nil)
	if err != nil {
		return nil, 0, err
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, 0, fmt.Errorf("%s", resp.Status)
	}
	if resp.ContentLength <= 0 {
		return nil, 0, errors.New("ZIP response has no content length")
	}
	if resp.ContentLength > maxSize {
		return nil, 0, fmt.Errorf("ZIP is %d bytes (limit %d)", resp.ContentLength, maxSize)
	}
	size := resp.ContentLength
	start := size - tailSize
	if start < 0 {
		start = 0
	}
	headers := make(http.Header)
	if start > 0 {
		headers.Set("Range", fmt.Sprintf("bytes=%d-%d", start, size-1))
	}
	resp, err = doHTTPRequest(ctx, client, http.MethodGet, endpoint, headers)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	wantStatus := http.StatusOK
	if start > 0 {
		wantStatus = http.StatusPartialContent
	}
	if resp.StatusCode != wantStatus {
		return nil, 0, fmt.Errorf("ZIP tail request returned %s", resp.Status)
	}
	wantBytes := size - start
	data, err := io.ReadAll(io.LimitReader(resp.Body, wantBytes+1))
	if err != nil {
		return nil, 0, err
	}
	if int64(len(data)) != wantBytes {
		return nil, 0, fmt.Errorf("ZIP tail returned %d bytes, want %d", len(data), wantBytes)
	}
	return &remoteZipReaderAt{
		ctx:        ctx,
		client:     client,
		endpoint:   endpoint,
		cachedTail: data,
		cachedData: make(map[int64][]byte),
		tailStart:  start,
		size:       size,
	}, size, nil
}

func (r *remoteZipReaderAt) ReadAt(p []byte, offset int64) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	if offset < 0 || offset >= r.size || int64(len(p)) > r.size-offset {
		return 0, io.EOF
	}
	written := 0
	for written < len(p) {
		current := offset + int64(written)
		if current >= r.tailStart {
			written += copy(p[written:], r.cachedTail[current-r.tailStart:])
			continue
		}
		blockStart := current - current%zipReadAhead
		r.cacheMu.Lock()
		block, ok := r.cachedData[blockStart]
		r.cacheMu.Unlock()
		if !ok {
			blockEnd := blockStart + zipReadAhead
			if blockEnd > r.tailStart {
				blockEnd = r.tailStart
			}
			var err error
			block, err = r.readRange(blockStart, blockEnd)
			if err != nil {
				return written, err
			}
			r.cacheMu.Lock()
			r.cachedData[blockStart] = block
			r.cacheMu.Unlock()
		}
		n := copy(p[written:], block[current-blockStart:])
		if n == 0 {
			return written, io.ErrUnexpectedEOF
		}
		written += n
	}
	return written, nil
}

func (r *remoteZipReaderAt) readRange(start, end int64) ([]byte, error) {
	headers := make(http.Header)
	headers.Set("Range", fmt.Sprintf("bytes=%d-%d", start, end-1))
	resp, err := doHTTPRequest(r.ctx, r.client, http.MethodGet, r.endpoint, headers)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusPartialContent {
		return nil, fmt.Errorf("ZIP range request returned %s", resp.Status)
	}
	want := end - start
	data, err := io.ReadAll(io.LimitReader(resp.Body, want+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) != want {
		return nil, fmt.Errorf("ZIP range returned %d bytes, want %d", len(data), want)
	}
	return data, nil
}

func doHTTPRequest(ctx context.Context, client *http.Client, method, endpoint string, headers http.Header) (*http.Response, error) {
	var lastErr error
	for attempt := 1; attempt <= httpAttempts; attempt++ {
		req, err := http.NewRequestWithContext(ctx, method, endpoint, nil)
		if err != nil {
			return nil, err
		}
		req.Header = headers.Clone()
		resp, err := client.Do(req)
		if err == nil && !isTransientHTTPStatus(resp.StatusCode) {
			return resp, nil
		}
		if err != nil {
			lastErr = err
		} else {
			lastErr = fmt.Errorf("%s", resp.Status)
			if attempt == httpAttempts {
				return resp, nil
			}
			_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 4<<10))
			resp.Body.Close()
		}
		if attempt == httpAttempts {
			break
		}
		timer := time.NewTimer(time.Duration(attempt) * httpRetryDelay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, ctx.Err()
		case <-timer.C:
		}
	}
	return nil, lastErr
}

func isTransientHTTPStatus(status int) bool {
	return status == http.StatusTooManyRequests || status >= 500
}

func containsPathElement(name, element string) bool {
	for _, part := range strings.Split(name, "/") {
		if part == element {
			return true
		}
	}
	return false
}

func inferAssemblyArchitecture(name string) string {
	return discoverymeta.InferArchitecture(name)
}

// escapeProxyPath implements the uppercase escaping used by the Go module
// proxy protocol. Index entries are already valid module paths.
func escapeProxyPath(value string) (string, error) {
	if value == "" {
		return "", errors.New("empty module path or version")
	}
	var out strings.Builder
	for _, r := range value {
		switch {
		case r == '!':
			out.WriteString("!!")
		case r >= 'A' && r <= 'Z':
			out.WriteByte('!')
			out.WriteRune(r + ('a' - 'A'))
		default:
			out.WriteRune(r)
		}
	}
	return out.String(), nil
}
