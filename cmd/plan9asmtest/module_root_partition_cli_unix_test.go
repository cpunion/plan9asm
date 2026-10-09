//go:build unix

package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Use the actual CLI and an actual first-party main module. Invalid split
// requests must fail at role validation, not because the flags are absent or
// a missing LLVM/Node/source input happens to stop execution later.
func TestModuleRootPartitionCLIRejectsInvalidRolesBeforeIO(t *testing.T) {
	moduleFixtureEnvironment(t)
	binary := parallelFixtureCLI(t)
	dir := moduleFixture(t, false)
	cases := []struct {
		name string
		args []string
	}{
		{"missing-part", []string{"-module", "-module-owner=0", "-module-root-partitions=8"}},
		{"out-of-range", []string{"-module", "-module-owner=0", "-module-root-partitions=8", "-module-root-partition=8"}},
		{"corpus-part", []string{"-module", "-module-owner=1", "-module-root-partitions=8", "-module-root-partition=0"}},
		{"other-part", []string{"-module", "-module-owner=2", "-module-root-partitions=8", "-module-root-partition=0"}},
		{"ordinary-part", []string{"-module-root-partitions=8", "-module-root-partition=0"}},
		{"local-part", []string{"-module", "-module-root-partitions=8", "-module-root-partition=0"}},
		{"aggregate-part", []string{"-module", "-verify-module=missing", "-module-root-partitions=8", "-module-root-partition=0"}},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			out := filepath.Join(t.TempDir(), "must-not-exist")
			stem := filepath.Join(t.TempDir(), "cli")
			args := append([]string(nil), test.args...)
			args = append(args, "-dir", dir, "-out", out, "-shards=32", "-workers=2", "-timeout=2m")
			_, err := captureCommand(context.Background(), 30*time.Second, ".", binary, args, stem)
			stderr, readErr := os.ReadFile(stem + ".stderr")
			if err == nil || readErr != nil || !strings.Contains(string(stderr), "root partition scheduling") {
				t.Fatalf("actual invalid role did not reach explicit split validation: %v, read=%v\n%s", err, readErr, stderr)
			}
			if strings.Contains(string(stderr), "flag provided but not defined") {
				t.Fatalf("missing CLI grammar, not the requested role guard:\n%s", stderr)
			}
			if _, err := os.Lstat(out); !os.IsNotExist(err) {
				t.Fatalf("invalid scheduling role performed output IO: %v", err)
			}
		})
	}
}
