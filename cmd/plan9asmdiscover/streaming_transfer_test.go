package main

import (
	"runtime"
	"testing"
	"time"
)

func TestDiscoveryTransferDoesNotRetainEitherLedgerPayloads(t *testing.T) {
	const records = 8192
	source, destination := writeSummaryPayloadFixture(t, records), writeSummaryPayloadFixture(t, records)
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
	err := mergeLedgerStreaming(destination, source)
	close(stop)
	maximum := <-peak
	if err != nil {
		t.Fatal(err)
	}
	var growth uint64
	if maximum > before.HeapAlloc {
		growth = maximum - before.HeapAlloc
	}
	t.Logf("sampled HeapAlloc growth=%d bytes importing %d source + %d destination records, each %d payload bytes",
		growth, records, records, records*8192)
	if growth > 24<<20 {
		t.Errorf("conversion retains ledger payloads: sampled HeapAlloc growth %d exceeds %d", growth, 24<<20)
	}
	summary, err := readShardedSummary(destination)
	if err != nil || summary.FailureRecords != records || summary.UniqueModules != records {
		t.Fatalf("transfer summary=%+v: %v", summary, err)
	}
}
