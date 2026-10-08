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

func TestDiscoveryTransferActualCLI(t *testing.T) {
	const childKey = "PLAN9ASM_DISCOVERY_TRANSFER_ARGS"
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
	for _, name := range []string{"sharded", "legacy-json", "invalid-tail"} {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			source, destination := filepath.Join(root, "source"), filepath.Join(root, "destination")
			change := discoveryReport{
				Scanned: []moduleVersion{{Path: "example.invalid/convert", Version: "v1.1.0"}},
				Matched: []candidate{{Module: "example.invalid/convert", Version: "v1.1.0", AsmFiles: []string{"source_amd64.s"}}},
			}
			if err := writeShardedReport(source, change); err != nil {
				t.Fatal(err)
			}
			if err := writeShardedReport(destination, discoveryReport{
				Scanned: []moduleVersion{{Path: "example.invalid/convert", Version: "v1.0.0"}},
			}); err != nil {
				t.Fatal(err)
			}
			if name == "legacy-json" {
				source = filepath.Join(root, "legacy.json")
				writeSummaryFixtureJSON(t, source, change)
			} else if name == "invalid-tail" {
				if err := os.WriteFile(filepath.Join(source, "records", "ff.jsonl"), []byte("invalid tail\n"), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			before := readDirectoryFiles(t, destination)
			encoded, err := json.Marshal([]string{"-convert-report", source, "-out-dir", destination,
				"-index-url", "http://127.0.0.1:1", "-proxy-url", "http://127.0.0.1:1"})
			if err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			command := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestDiscoveryTransferActualCLI$")
			command.Env = append(os.Environ(), childKey+"="+string(encoded))
			command.WaitDelay = time.Second
			var stdout, stderr bytes.Buffer
			command.Stdout, command.Stderr = &stdout, &stderr
			err = ownedprocess.Run(command)
			if ctx.Err() != nil || errors.Is(err, ownedprocess.ErrCleanup) {
				t.Fatalf("conversion child did not close: %v context=%v", err, ctx.Err())
			}
			if name == "invalid-tail" {
				var exit *exec.ExitError
				if !errors.As(err, &exit) || exit.ExitCode() != 1 || stdout.Len() != 0 ||
					!strings.Contains(stderr.String(), "decode record shard "+filepath.Join(source, "records", "ff.jsonl")+":1") {
					t.Fatalf("invalid tail lacks plain bound failure: %v stdout=%s stderr=%s", err, stdout.String(), stderr.String())
				}
				if !reflect.DeepEqual(before, readDirectoryFiles(t, destination)) {
					t.Fatal("failed actual CLI changed authoritative destination")
				}
			} else {
				if err != nil || stdout.Len() != 0 || stderr.Len() != 0 {
					t.Fatalf("actual conversion failed: %v stdout=%s stderr=%s", err, stdout.String(), stderr.String())
				}
				got, err := readShardedReport(destination)
				if err != nil || len(got.Scanned) != 1 || got.Scanned[0].Version != "v1.1.0" || len(got.Matched) != 1 {
					t.Fatalf("actual conversion lost normalized source: %+v %v", got, err)
				}
			}
		})
	}
}
