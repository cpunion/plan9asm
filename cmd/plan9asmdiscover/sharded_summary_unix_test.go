//go:build unix

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/xgo-dev/plan9asm/internal/ownedprocess"
)

func TestDiscoveryStatusActualCLI(t *testing.T) {
	const childKey = "PLAN9ASM_STATUS_TEST_LEDGER"
	if ledger := os.Getenv(childKey); ledger != "" {
		flag.CommandLine = flag.NewFlagSet("plan9asmdiscover", flag.ExitOnError)
		os.Args = []string{"plan9asmdiscover", "-status", "-out-dir", ledger,
			"-index-url", "http://127.0.0.1:1", "-proxy-url", "http://127.0.0.1:1"}
		main()
		os.Exit(0)
	}
	ledger := filepath.Join(t.TempDir(), "ledger")
	if err := writeShardedReport(ledger, discoveryReport{
		SchemaVersion: discoverySchema,
		IndexRanges:   []indexRange{{Since: indexEpoch, Before: "2026-10-07T00:00:00Z", IndexEntries: 1}},
		Scanned:       []moduleVersion{{Path: "example.com/a", Version: "v1.0.0"}},
	}); err != nil {
		t.Fatal(err)
	}
	complete, err := readShardedReport(ledger)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestDiscoveryStatusActualCLI$")
	command.Env = append(os.Environ(), childKey+"="+ledger)
	command.WaitDelay = time.Second
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	if err := ownedprocess.Run(command); err != nil {
		t.Fatalf("actual -status failed: %v\n%s", err, stderr.String())
	}
	if ctx.Err() != nil || stderr.Len() != 0 {
		t.Fatalf("actual -status has context/stderr failure: %v, %s", ctx.Err(), stderr.String())
	}
	var got ledgerSummary
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("decode actual -status: %v\n%s", err, stdout.String())
	}
	if want := summarizeLedger(complete); !reflect.DeepEqual(got, want) {
		t.Fatalf("actual CLI summary = %#v, want %#v", got, want)
	}

	// The last shard is outside the fixture's original data shard. Status must
	// consume it too, not stop after matching manifest counts in an earlier shard.
	if err := os.WriteFile(filepath.Join(ledger, "records", "ff.jsonl"), []byte("invalid tail\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	command = exec.CommandContext(ctx, os.Args[0], "-test.run=^TestDiscoveryStatusActualCLI$")
	command.Env = append(os.Environ(), childKey+"="+ledger)
	command.WaitDelay = time.Second
	stdout.Reset()
	stderr.Reset()
	command.Stdout, command.Stderr = &stdout, &stderr
	err = ownedprocess.Run(command)
	var exit *exec.ExitError
	if ctx.Err() != nil || errors.Is(err, ownedprocess.ErrCleanup) || !errors.As(err, &exit) || exit.ExitCode() != 1 {
		t.Fatalf("actual invalid-tail status is not a plain failure: %v, context=%v, stderr=%s", err, ctx.Err(), stderr.String())
	}
	if stdout.Len() != 0 || !strings.Contains(stderr.String(), filepath.Join(ledger, "records", "ff.jsonl")+":1") ||
		!strings.Contains(stderr.String(), "decode record shard") {
		t.Fatalf("actual invalid tail lacks bound diagnostic/zero summary: stdout=%s stderr=%s", stdout.String(), stderr.String())
	}
}
