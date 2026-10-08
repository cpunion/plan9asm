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
	"runtime"
	"testing"
	"time"

	"github.com/xgo-dev/plan9asm/internal/ownedprocess"
)

func TestDiscoveryCollectionTransferActualCLIHeap(t *testing.T) {
	const childKey = "PLAN9ASM_COLLECTION_TRANSFER_ARGS"
	if encoded := os.Getenv(childKey); encoded != "" {
		var args []string
		if err := json.Unmarshal([]byte(encoded), &args); err != nil {
			t.Fatal(err)
		}
		flag.CommandLine = flag.NewFlagSet("plan9asmdiscover", flag.ExitOnError)
		os.Args = append([]string{"plan9asmdiscover"}, args...)
		runtime.GC()
		var before runtime.MemStats
		runtime.ReadMemStats(&before)
		stop, peak := make(chan struct{}), make(chan uint64, 1)
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
		main()
		close(stop)
		maximum := <-peak
		var growth uint64
		if maximum > before.HeapAlloc {
			growth = maximum - before.HeapAlloc
		}
		t.Logf("actual collection CLI sampled HeapAlloc growth=%d bytes; source=2x4096x8192, destination=4096x8192 payload bytes", growth)
		if growth > 24<<20 {
			t.Errorf("collection CLI retains complete ledger payloads: growth=%d exceeds %d", growth, 24<<20)
		}
		return
	}

	root := t.TempDir()
	source, destination := filepath.Join(root, "collection"), filepath.Join(root, "destination")
	if err := os.Mkdir(source, 0o700); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"a-first", "z-last"} {
		if err := os.Rename(writeSummaryPayloadFixture(t, 4096), filepath.Join(source, name)); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Rename(writeSummaryPayloadFixture(t, 4096), destination); err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal([]string{
		"-convert-report", source, "-out-dir", destination,
		"-index-url", "http://127.0.0.1:1", "-proxy-url", "http://127.0.0.1:1",
	})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	command := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestDiscoveryCollectionTransferActualCLIHeap$", "-test.v")
	command.Env = append(os.Environ(), childKey+"="+string(encoded))
	command.WaitDelay = time.Second
	var output bytes.Buffer
	command.Stdout, command.Stderr = &output, &output
	err = ownedprocess.Run(command)
	t.Logf("actual CLI child: %s", output.String())
	if err != nil || ctx.Err() != nil || errors.Is(err, ownedprocess.ErrCleanup) {
		t.Fatalf("bounded collection conversion failed: %v context=%v", err, ctx.Err())
	}
	summary, err := readShardedSummary(destination)
	if err != nil || summary.FailureRecords != 4096 || summary.UniqueModules != 4096 {
		t.Fatalf("actual collection conversion summary=%+v: %v", summary, err)
	}
}
