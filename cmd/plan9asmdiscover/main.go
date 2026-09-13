// Command plan9asmdiscover finds public Go modules that contain Plan 9
// assembly. It reads the official module index, resolves each distinct module
// to @latest, and inspects the cached module ZIP without executing module code.
package main

import (
	"archive/zip"
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	defaultIndexURL    = "https://index.golang.org/index"
	defaultProxyURL    = "https://proxy.golang.org/cached-only"
	indexPageLimit     = 2000
	defaultScanLimit   = 2000
	defaultMaxZipSize  = 2 << 30
	defaultHTTPTimeout = 60 * time.Second
	zipDirectoryTail   = 8 << 20
	httpAttempts       = 3
	httpRetryDelay     = 100 * time.Millisecond
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
	Path    string
	Version string
}

type candidate struct {
	Module        string   `json:"module"`
	Version       string   `json:"version"`
	Architectures []string `json:"architectures"`
	AsmFiles      []string `json:"asm_files"`
}

type scanFailure struct {
	Module string `json:"module"`
	Error  string `json:"error"`
}

type discoveryReport struct {
	SchemaVersion int           `json:"schema_version"`
	GeneratedAt   time.Time     `json:"generated_at"`
	Since         string        `json:"since,omitempty"`
	NextSince     string        `json:"next_since,omitempty"`
	IndexEntries  int           `json:"index_entries"`
	UniqueModules int           `json:"unique_modules"`
	Matched       []candidate   `json:"matched"`
	Failures      []scanFailure `json:"failures,omitempty"`
}

type config struct {
	indexURL    string
	proxyURL    string
	since       string
	limit       int
	workers     int
	maxZipSize  int64
	httpTimeout time.Duration
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
	flag.Parse()

	cfg := config{
		indexURL:    *indexURL,
		proxyURL:    strings.TrimRight(*proxyURL, "/"),
		since:       *since,
		limit:       *limit,
		workers:     *workers,
		maxZipSize:  *maxZipSize,
		httpTimeout: *httpTimeout,
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

	matched, failures := inspectModules(ctx, client, cfg.proxyURL, modules, cfg.workers, cfg.maxZipSize)
	return discoveryReport{
		SchemaVersion: 1,
		GeneratedAt:   time.Now().UTC(),
		Since:         cfg.since,
		NextSince:     nextSince,
		IndexEntries:  len(entries),
		UniqueModules: len(modules),
		Matched:       matched,
		Failures:      failures,
	}, nil
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

func inspectModules(ctx context.Context, client *http.Client, proxyURL string, modules []moduleVersion, workers int, maxZipSize int64) ([]candidate, []scanFailure) {
	tasks := make(chan moduleVersion)
	results := make(chan candidate)
	errorsOut := make(chan scanFailure)
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for module := range tasks {
				item, err := inspectIndexedModule(ctx, client, proxyURL, module, maxZipSize)
				if err != nil {
					errorsOut <- scanFailure{Module: module.Path, Error: err.Error()}
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
	}()

	matched := make([]candidate, 0)
	failures := make([]scanFailure, 0)
	for results != nil || errorsOut != nil {
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
		}
	}
	sort.Slice(matched, func(i, j int) bool { return matched[i].Module < matched[j].Module })
	sort.Slice(failures, func(i, j int) bool { return failures[i].Module < failures[j].Module })
	return matched, failures
}

func inspectIndexedModule(ctx context.Context, client *http.Client, proxyURL string, module moduleVersion, maxZipSize int64) (candidate, error) {
	indexed, err := inspectModuleVersion(ctx, client, proxyURL, module.Path, module.Version, maxZipSize)
	if err != nil {
		return candidate{}, fmt.Errorf("inspect indexed version %s: %w", module.Version, err)
	}
	if len(indexed.AsmFiles) == 0 {
		return indexed, nil
	}
	latestVersion, err := resolveLatest(ctx, client, proxyURL, module.Path)
	if err != nil {
		return candidate{}, err
	}
	if latestVersion == module.Version {
		return indexed, nil
	}
	return inspectModuleVersion(ctx, client, proxyURL, module.Path, latestVersion, maxZipSize)
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
	if err := json.Unmarshal(latestBody, &latest); err != nil || latest.Version == "" {
		return "", fmt.Errorf("decode @latest: %w", err)
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
	asmFiles := make([]string, 0)
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
				asmFiles = append(asmFiles, rel)
			}
		}
	}
	files := make([]string, 0, len(asmFiles))
	archSet := make(map[string]struct{})
	for _, name := range asmFiles {
		if _, ok := goDirs[path.Dir(name)]; !ok {
			continue
		}
		files = append(files, name)
		archSet[inferAssemblyArchitecture(name)] = struct{}{}
	}
	arches := make([]string, 0, len(archSet))
	for arch := range archSet {
		arches = append(arches, arch)
	}
	sort.Strings(files)
	sort.Strings(arches)
	return files, arches, nil
}

type remoteZipReaderAt struct {
	ctx        context.Context
	client     *http.Client
	endpoint   string
	cachedTail []byte
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
	if offset >= r.tailStart {
		copy(p, r.cachedTail[offset-r.tailStart:])
		return len(p), nil
	}

	headers := make(http.Header)
	headers.Set("Range", fmt.Sprintf("bytes=%d-%d", offset, offset+int64(len(p))-1))
	resp, err := doHTTPRequest(r.ctx, r.client, http.MethodGet, r.endpoint, headers)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusPartialContent {
		return 0, fmt.Errorf("ZIP range request returned %s", resp.Status)
	}
	n, err := io.ReadFull(resp.Body, p)
	if err != nil {
		return n, err
	}
	return n, nil
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
	base := strings.TrimSuffix(path.Base(name), ".s")
	for _, arch := range []string{"386", "amd64", "arm", "arm64", "wasm"} {
		if strings.HasSuffix(base, "_"+arch) {
			return arch
		}
	}
	return "unknown"
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
