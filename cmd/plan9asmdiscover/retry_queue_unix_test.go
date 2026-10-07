//go:build unix

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/xgo-dev/plan9asm/internal/ownedprocess"
)

func TestDiscoveryRetryActualCLI(t *testing.T) {
	const childKey = "PLAN9ASM_DISCOVERY_RETRY_ARGS"
	if encoded := os.Getenv(childKey); encoded != "" {
		var args []string
		if err := json.Unmarshal([]byte(encoded), &args); err != nil {
			t.Fatal(err)
		}
		flag.CommandLine = flag.NewFlagSet("plan9asmdiscover", flag.ExitOnError)
		os.Args = append([]string{"plan9asmdiscover"}, args...)
		main()
		os.Exit(0)
	}
	ledger := filepath.Join(t.TempDir(), "ledger")
	base := discoveryReport{
		IndexRanges: []indexRange{{Since: "2026-10-01T00:00:00Z", Before: "2026-10-02T00:00:00Z", IndexEntries: 9}},
		Failures: []scanFailure{
			{Module: "example.invalid/a", Version: "v1.0.0", Error: "owned persistent failure"},
			{Module: "example.invalid/b", Version: "v1.0.0", Error: "owned persistent failure"},
			{Module: "example.invalid/c", Version: "v1.0.0", Error: "owned persistent failure"},
		},
	}
	if err := writeShardedReport(ledger, base); err != nil {
		t.Fatal(err)
	}
	var mu sync.Mutex
	var requests []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		requests = append(requests, r.URL.Path)
		mu.Unlock()
		http.Error(w, "owned persistent latest failure", http.StatusNotFound)
	}))
	defer server.Close()
	for batch := 0; batch < 6; batch++ {
		encoded, err := json.Marshal([]string{"-scan-mode", "retry", "-limit", "1", "-workers", "1", "-out-dir", ledger,
			"-index-url", server.URL + "/must-not-read-index", "-proxy-url", server.URL})
		if err != nil {
			t.Fatal(err)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		command := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestDiscoveryRetryActualCLI$")
		command.Env = append(os.Environ(), childKey+"="+string(encoded))
		command.WaitDelay = time.Second
		var stdout, stderr bytes.Buffer
		command.Stdout, command.Stderr = &stdout, &stderr
		err = ownedprocess.Run(command)
		contextErr := ctx.Err()
		cancel()
		if err != nil || contextErr != nil {
			t.Fatalf("actual retry batch%d did not close successfully: %v context=%v stderr=%s", batch, err, contextErr, stderr.String())
		}
		var summary ledgerSummary
		if err := json.Unmarshal(stdout.Bytes(), &summary); err != nil {
			t.Fatalf("actual retry emitted invalid summary: %v\n%s", err, stdout.String())
		}
		mu.Lock()
		got := append([]string(nil), requests...)
		requests = nil
		mu.Unlock()
		want := []string{"/example.invalid/" + string(rune('a'+batch%3)) + "/@latest"}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("actual -limit1 retry batch%d exceeded or starved its bounded queue: requests=%v want=%v", batch, got, want)
		}
		if strings.Contains(stderr.String(), "must-not-read-index") || summary.IndexRanges != 1 || summary.IndexEntries != 9 ||
			summary.HistoryBefore != base.IndexRanges[0].Since || summary.IncrementalSince != base.IndexRanges[0].Before {
			t.Fatalf("retry changed index coverage: summary=%+v stderr=%s", summary, stderr.String())
		}
		persisted, err := scanShardedReport(ledger, nil)
		if err != nil || !reflect.DeepEqual(persisted.metadata.IndexRanges, base.IndexRanges) || persisted.metadata.IndexEntries != 9 {
			t.Fatalf("retry publication changed physical index ranges: %+v %v", persisted.metadata, err)
		}
	}
}
