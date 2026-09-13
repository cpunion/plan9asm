package main

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
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

func TestInspectModuleZipFindsTargetAssembly(t *testing.T) {
	var archive bytes.Buffer
	zw := zip.NewWriter(&archive)
	for _, name := range []string{
		"example.com/lib@v1.2.3/hash/hash.go",
		"example.com/lib@v1.2.3/hash/hash_amd64.s",
		"example.com/lib@v1.2.3/hash/hash_arm64.s",
		"example.com/lib@v1.2.3/hash/empty_arm.s",
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
		"foo_wasm.s":          "wasm",
		"asm_darwin_x86_gc.s": "unknown",
	}
	for name, want := range tests {
		if got := inferAssemblyArchitecture(name); got != want {
			t.Errorf("inferAssemblyArchitecture(%q) = %q, want %q", name, got, want)
		}
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
	}))
}
