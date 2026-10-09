//go:build unix

package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func runRootSplitFixture(t *testing.T, binary, dir, out, inputs, profile string, owner, part int) {
	t.Helper()
	args := []string{"-module", "-module-root-partitions=8", "-dir", dir, "-out", out,
		"-shards=32", "-workers=2", "-timeout=2m"}
	if inputs != "" {
		args = append(args, "-verify-module", inputs)
	} else {
		args = append(args, fmt.Sprintf("-module-owner=%d", owner), fmt.Sprintf("-module-root-partition=%d", part))
	}
	args = append(args, "--", "-p=1")
	if profile != "" {
		args = append(args, "-covermode=atomic", "-coverprofile="+profile)
	}
	stem := filepath.Join(t.TempDir(), "cli")
	if _, err := captureCommand(context.Background(), 3*time.Minute, ".", binary, args, stem); err != nil {
		t.Fatal(errorsWithActualOutput(err, stem))
	}
}

func TestModuleRootPartitionActualCompleteUnionAndCoverage(t *testing.T) {
	moduleFixtureEnvironment(t)
	binary := parallelFixtureCLI(t)
	dir := moduleRootPartitionFixture(t)
	inputs := t.TempDir()
	if err := os.Mkdir(filepath.Join(inputs, "owner-0000"), 0700); err != nil {
		t.Fatal(err)
	}
	profile := filepath.Join(t.TempDir(), "whole.cover")
	for part := 0; part < 8; part++ {
		out := filepath.Join(inputs, "owner-0000", fmt.Sprintf("part-%04d", part))
		runRootSplitFixture(t, binary, dir, out, "", profile, 0, part)
	}
	for owner := 1; owner < moduleOwners; owner++ {
		out := filepath.Join(inputs, fmt.Sprintf("owner-%04d", owner))
		runRootSplitFixture(t, binary, dir, out, "", profile, owner, -1)
	}
	if _, err := os.Lstat(profile); !os.IsNotExist(err) {
		t.Fatalf("an owner published partial uploadable module coverage: %v", err)
	}
	verified := filepath.Join(t.TempDir(), "aggregate")
	runRootSplitFixture(t, binary, dir, verified, inputs, profile, -1, -1)
	var expected moduleReport
	if err := ciReadJSON(filepath.Join(verified, "module.json"), &expected); err != nil {
		t.Fatal(err)
	}
	if expected.Protocol != moduleRootPartsProtocol || expected.Status != "passed" ||
		expected.RootInventory == nil || len(expected.RootInventory.Names) != 36 || len(expected.Packages) != 5 ||
		expected.Packages[0].Shards != 32 {
		t.Fatalf("complete fresh denominator/package union lost: %+v", expected)
	}
	data, err := os.ReadFile(profile)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"modulefixture/value.go:", "modulefixture/cmd/plan9asmcorpus/value.go:", "modulefixture/internal/other/value.go:"} {
		if !strings.Contains(string(data), name) {
			t.Fatalf("complete module coverage lost %s: %s", name, data)
		}
	}
	if _, err := os.Lstat(filepath.Join(verified, "root-inventory.stdout")); err != nil {
		t.Fatalf("aggregate did not retain its own real complete root list: %v", err)
	}
	results, err := checkedModuleArtifacts(expected, inputs)
	if err != nil || len(results) != 5 || len(results[0].rootCoverage) != 32 {
		t.Fatalf("independent full raw-shard replay: %+v %v", results, err)
	}

	partPath := filepath.Join(inputs, "owner-0000", "part-0000")
	modulePath := filepath.Join(partPath, "module.json")
	var original moduleReport
	if err := ciReadJSON(modulePath, &original); err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		name   string
		mutate func(*moduleReport)
	}{
		{"part", func(r *moduleReport) { value := 1; r.RootPartition = &value }},
		{"count", func(r *moduleReport) { r.RootPartitions = 7 }},
		{"owner", func(r *moduleReport) { r.Owner = 1 }},
		{"legacy", func(r *moduleReport) { r.Protocol = moduleProtocol }},
		{"failed", func(r *moduleReport) { r.Status = "failed" }},
		{"budget", func(r *moduleReport) { r.Timeout = "3m0s" }},
		{"workers", func(r *moduleReport) { r.Workers = 3 }},
		{"source", func(r *moduleReport) { r.Scope.SourceSHA = strings.Repeat("0", 64) }},
		{"sdk", func(r *moduleReport) { r.Scope.SDKSource = strings.Repeat("0", 64) }},
		{"tools", func(r *moduleReport) { r.Scope.Tools["go"] = strings.Repeat("0", 64) }},
		{"environment", func(r *moduleReport) { r.Scope.Environment["GOGCCFLAGS"] += " -Dforeign" }},
	} {
		t.Run(test.name, func(t *testing.T) {
			var changed moduleReport
			if err := ciReadJSON(modulePath, &changed); err != nil {
				t.Fatal(err)
			}
			test.mutate(&changed)
			if err := ciWriteJSON(modulePath, changed); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if err := ciWriteJSON(modulePath, original); err != nil {
					t.Error(err)
				}
			})
			if _, err := checkedModuleArtifacts(expected, inputs); err == nil {
				t.Fatal("incomplete/foreign root fragment became a full module pass")
			}
		})
	}
	for _, suffix := range []string{"partition.json", "shard-0000.stdout", "shard-0000.cover"} {
		t.Run("missing-"+suffix, func(t *testing.T) {
			path := filepath.Join(partPath, "package-0000", suffix)
			backup := filepath.Join(t.TempDir(), "retained")
			if err := os.Rename(path, backup); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if err := os.Rename(backup, path); err != nil {
					t.Error(err)
				}
			})
			if _, err := checkedModuleArtifacts(expected, inputs); err == nil {
				t.Fatal("missing original report/raw terminal/profile became pass")
			}
		})
	}
	t.Run("missing-whole-part", func(t *testing.T) {
		backup := filepath.Join(t.TempDir(), "retained")
		if err := os.Rename(partPath, backup); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			if err := os.Rename(backup, partPath); err != nil {
				t.Error(err)
			}
		})
		if _, err := checkedModuleArtifacts(expected, inputs); err == nil {
			t.Fatal("missing whole root part became pass")
		}
	})
	if _, err := checkedModuleArtifacts(expected, inputs); err != nil {
		t.Fatalf("independent controls changed original evidence: %v", err)
	}
}
