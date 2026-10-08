package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"
)

// The earlier large interest fixture contains unrelated scanned records only.
// This one has distinct matched keys and complete, safe assembly-path payloads.
func writeSupersededMatchedPayloadLedger(t *testing.T, directory bool, records int) string {
	t.Helper()
	ledger := writeSupersededInterestLedger(t, directory, 0, "")
	filePath := ledger
	if directory {
		filePath = filepath.Join(ledger, "records", "00.jsonl")
	}
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	writer := bufio.NewWriter(file)
	prefix := strings.Repeat("nested_source_directory/", 32)
	paths := make([]string, 4)
	for i := range paths {
		paths[i] = fmt.Sprintf("%soperation%d_amd64.s", prefix, i)
	}
	record := discoveryRecord{Kind: "matched", Version: "v1.0.0",
		Architectures: []string{"amd64", "arm64", "s390x"}, AsmFiles: paths}
	for i := 0; i < records; i++ {
		record.Module = fmt.Sprintf("example.com/unrelated-matched-%06d", i)
		data, err := json.Marshal(record)
		if err == nil {
			_, err = writer.Write(append(data, '\n'))
		}
		if err != nil {
			file.Close()
			t.Fatal(err)
		}
	}
	if err := writer.Flush(); err != nil {
		file.Close()
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	return ledger
}

func sampleSupersededMatchedHeap(run func() error) (uint64, error) {
	runtime.GC()
	var before runtime.MemStats
	runtime.ReadMemStats(&before)
	ready, stop := make(chan struct{}), make(chan struct{})
	observed := make(chan uint64, 1)
	go func() {
		ticker := time.NewTicker(time.Millisecond)
		defer ticker.Stop()
		maximum := before.HeapAlloc
		close(ready)
		for {
			var stats runtime.MemStats
			runtime.ReadMemStats(&stats)
			if stats.HeapAlloc > maximum {
				maximum = stats.HeapAlloc
			}
			select {
			case <-stop:
				observed <- maximum
				return
			case <-ticker.C:
			}
		}
	}()
	<-ready
	err := run()
	close(stop)
	maximum := <-observed // The sampler is closed before leaving the operation.
	if maximum < before.HeapAlloc {
		return 0, err
	}
	return maximum - before.HeapAlloc, err
}

func TestSupersededMatchedPayloadLookupIsBounded(t *testing.T) {
	const unrelated, budget = 18000, 24 << 20
	for _, directory := range []bool{false, true} {
		t.Run(fmt.Sprintf("directory=%t", directory), func(t *testing.T) {
			ledger := writeSupersededMatchedPayloadLedger(t, directory, unrelated)
			for _, selected := range []bool{true, false} {
				t.Run(fmt.Sprintf("selected=%t", selected), func(t *testing.T) {
					var expected []discoverySupersededSkip
					if selected {
						expected = []discoverySupersededSkip{supersededInterestSkip("example.com/current", "v1.3.0")}
					}
					root := writeSupersededInterestManifest(t, expected)
					growth, err := sampleSupersededMatchedHeap(func() error {
						skips, err := loadSupersededSkips(root, ledger)
						if err == nil && len(skips) != len(expected) {
							return fmt.Errorf("retained %d exceptions, want %d", len(skips), len(expected))
						}
						return err
					})
					if err != nil {
						t.Fatal(err)
					}
					t.Logf("unrelated matched=%d, assembly paths/record=4, sampled HeapAlloc growth=%d, budget=%d", unrelated, growth, budget)
					if growth > budget {
						t.Errorf("exception lookup retains unrelated matched payload: sampled growth %d exceeds %d", growth, budget)
					}
				})
			}
		})
	}
}

func TestSupersededMatchedMembershipPreservesExactOldAndReplacementKeys(t *testing.T) {
	const extra = `{"kind":"matched","module":"example.com/old","version":"v0.9.0","asm_files":["old.s"]}` + "\n" +
		`{"kind":"scanned","module":"example.com/old","version":"v1.1.0"}` + "\n" +
		`{"kind":"failure","module":"example.com/failed","version":"v1.5.0"}` + "\n"
	for _, directory := range []bool{false, true} {
		t.Run(fmt.Sprintf("directory=%t", directory), func(t *testing.T) {
			ledger := writeSupersededInterestLedger(t, directory, 50, extra)
			for _, test := range []struct {
				name, oldModule, oldVersion, replacement, version string
				accepted                                          bool
			}{
				{"scanned replacement", "example.com/old", "v1.2.0", "example.com/current", "v1.3.0", true},
				{"matched replacement", "example.com/old", "v1.2.0", "example.com/assembly", "v1.4.0", true},
				{"independent v0 line", "example.com/old", "v0.9.0", "example.com/current", "v1.3.0", true},
				{"missing old module", "example.com/unknown", "v1.2.0", "example.com/current", "v1.3.0", false},
				{"wrong old exact version", "example.com/old", "v1.2.1", "example.com/current", "v1.3.0", false},
				{"old scanned is not matched", "example.com/old", "v1.1.0", "example.com/current", "v1.3.0", false},
				{"old failure is not matched", "example.com/failed", "v1.5.0", "example.com/assembly", "v1.6.0", false},
				{"missing replacement", "example.com/old", "v1.2.0", "example.com/unknown", "v1.3.0", false},
				{"wrong replacement exact version", "example.com/old", "v1.2.0", "example.com/current", "v1.3.1", false},
				{"failed replacement is not scanned", "example.com/old", "v1.2.0", "example.com/failed", "v1.5.0", false},
			} {
				t.Run(test.name, func(t *testing.T) {
					skip := supersededInterestSkip(test.replacement, test.version)
					skip.Module, skip.Version = test.oldModule, test.oldVersion
					root := writeSupersededInterestManifest(t, []discoverySupersededSkip{skip})
					skips, err := loadSupersededSkips(root, ledger)
					if test.accepted {
						if err != nil || !reflect.DeepEqual(skips, map[string]discoverySupersededSkip{test.oldModule + "@" + test.oldVersion: skip}) {
							t.Fatalf("exact membership/evidence changed: %v, %v", skips, err)
						}
					} else if err == nil || skips != nil {
						t.Fatalf("unproved membership retained a certificate: %v, %v", skips, err)
					}
				})
			}
		})
	}
}

func TestSupersededMatchedMembershipValidatesEveryUnselectedTail(t *testing.T) {
	for _, directory := range []bool{false, true} {
		for _, tail := range []string{
			"{truncated unselected JSON",
			`{"kind":"matched","module":"example.com/unselected","version":"v1.0.0","asm_files":["../unsafe.s"]}`,
			`{"kind":"matched","module":"example.com/unselected","version":"v1.0.0","asm_files":[]}`,
			`{"kind":"matched","module":"bad module","version":"v1.0.0","asm_files":["safe.s"]}`,
			`{"kind":"matched","module":"example.com/unselected","version":"bad version","asm_files":["safe.s"]}`,
		} {
			t.Run(fmt.Sprintf("directory=%t/tail=%q", directory, tail), func(t *testing.T) {
				ledger := writeSupersededInterestLedger(t, directory, 10, tail)
				for _, selected := range []bool{true, false} {
					var skips []discoverySupersededSkip
					if selected {
						skips = []discoverySupersededSkip{supersededInterestSkip("example.com/current", "v1.3.0")}
					}
					root := writeSupersededInterestManifest(t, skips)
					if result, err := loadSupersededSkips(root, ledger); err == nil || result != nil {
						t.Fatalf("selected=%t hid invalid unselected tail: %v, %v", selected, result, err)
					}
					interests := map[string]bool{}
					if selected {
						interests["example.com/current@v1.3.0"] = true
					}
					if result, err := scannedDiscoveryVersionKeysFor(ledger, interests); err == nil || result != nil {
						t.Fatalf("replacement reader hid invalid matched tail: %v, %v", result, err)
					}
				}
			})
		}
	}
}

func TestSupersededMatchedMembershipDoesNotChangeCorpusCandidateDenominator(t *testing.T) {
	for _, directory := range []bool{false, true} {
		t.Run(fmt.Sprintf("directory=%t", directory), func(t *testing.T) {
			ledger := writeSupersededMatchedPayloadLedger(t, directory, 64)
			before, err := loadDiscoveryCandidates(ledger)
			if err != nil || len(before) != 66 {
				t.Fatalf("full business inventory: %d candidates, %v", len(before), err)
			}
			root := writeSupersededInterestManifest(t, []discoverySupersededSkip{supersededInterestSkip("example.com/current", "v1.3.0")})
			if skips, err := loadSupersededSkips(root, ledger); err != nil || len(skips) != 1 {
				t.Fatalf("bounded exception inventory: %v, %v", skips, err)
			}
			after, err := loadDiscoveryCandidates(ledger)
			if err != nil || !reflect.DeepEqual(before, after) {
				t.Fatalf("exception query changed full candidate payloads: %v", err)
			}
		})
	}
}
