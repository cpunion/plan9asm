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

func runRootSplitFixture(t *testing.T, binary, dir, out, inputs, profile string, owner, part, count int) {
	t.Helper()
	args := []string{"-module", fmt.Sprintf("-module-root-partitions=%d", count), "-dir", dir, "-out", out,
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

func rootSplitRetainFile(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.WriteFile(path, data, 0600); err != nil {
			t.Error(err)
		}
	})
	return data
}

func rootSplitRequireRejected(t *testing.T, expected moduleReport, inputs string) {
	t.Helper()
	if _, err := checkedModuleArtifacts(expected, inputs); err == nil {
		t.Fatal("incomplete/foreign evidence became a complete module pass")
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
	for part := 0; part < 16; part++ {
		out := filepath.Join(inputs, "owner-0000", fmt.Sprintf("part-%04d", part))
		runRootSplitFixture(t, binary, dir, out, "", profile, 0, part, 16)
	}
	for owner := 1; owner < moduleOwners; owner++ {
		out := filepath.Join(inputs, fmt.Sprintf("owner-%04d", owner))
		runRootSplitFixture(t, binary, dir, out, "", profile, owner, -1, 16)
	}
	if _, err := os.Lstat(profile); !os.IsNotExist(err) {
		t.Fatalf("an owner published partial uploadable module coverage: %v", err)
	}
	verified := filepath.Join(t.TempDir(), "aggregate")
	runRootSplitFixture(t, binary, dir, verified, inputs, profile, -1, -1, 16)
	var expected moduleReport
	if err := ciReadJSON(filepath.Join(verified, "module.json"), &expected); err != nil {
		t.Fatal(err)
	}
	if expected.Protocol != moduleRootPartsProtocol || expected.Status != "passed" || expected.RootPartitions != 16 ||
		expected.RootInventory == nil || len(expected.RootInventory.Names) != 36 || len(expected.Packages) != 5 ||
		expected.Packages[0].Shards != 32 {
		t.Fatalf("complete fresh denominator/package union lost: %+v", expected)
	}
	data, err := os.ReadFile(profile)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"modulefixture/value.go:", "modulefixture/gotoolprofile/value.go:", "modulefixture/internal/other/value.go:"} {
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
	packagePath := filepath.Join(partPath, "package-0000")
	for _, suffix := range []string{"module.json", filepath.Join("package-0000", "partition.json"),
		filepath.Join("package-0000", "shard-0000.stdout")} {
		t.Run("trailing-EOF-"+suffix, func(t *testing.T) {
			path := filepath.Join(partPath, suffix)
			data := rootSplitRetainFile(t, path)
			if err := os.WriteFile(path, append(data, []byte("\n{}\n")...), 0600); err != nil {
				t.Fatal(err)
			}
			rootSplitRequireRejected(t, expected, inputs)
		})
	}
	for _, name := range []string{"unowned-pass", "unowned-command", "unowned-terminal", "missing-child", "duplicate-index"} {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(packagePath, "partition.json")
			rootSplitRetainFile(t, path)
			var suite suiteReport
			if err := ciReadJSON(path, &suite); err != nil {
				t.Fatal(err)
			}
			switch name {
			case "unowned-pass":
				suite.Shards[1].Status = "passed"
			case "unowned-command":
				suite.Shards[1].Command = append([]string(nil), suite.Shards[0].Command...)
			case "unowned-terminal":
				suite.Shards[1].Outcomes = map[string]string{suite.Shards[1].Names[0]: "pass"}
			case "missing-child":
				removed := false
				for child := range suite.Shards[0].Outcomes {
					if strings.Contains(child, "/") {
						delete(suite.Shards[0].Outcomes, child)
						removed = true
						break
					}
				}
				if !removed {
					t.Fatal("actual first-party shard had no child terminal")
				}
			case "duplicate-index":
				suite.Shards[8].Index = 0
			}
			if err := ciWriteJSON(path, suite); err != nil {
				t.Fatal(err)
			}
			rootSplitRequireRejected(t, expected, inputs)
		})
	}
	t.Run("paired-list-and-plan-shrink", func(t *testing.T) {
		// Both stored denominators change together; the independent aggregate
		// list, not a producer's matching hash, must still reject the deletion.
		path := filepath.Join(packagePath, "partition.json")
		rootSplitRetainFile(t, path)
		var suite suiteReport
		if err := ciReadJSON(path, &suite); err != nil {
			t.Fatal(err)
		}
		removed := suite.Names[0]
		suite.Names = suite.Names[1:]
		suite.InventorySHA = ciNamesSHA(suite.Names)
		listingPath := filepath.Join(packagePath, "inventory.stdout")
		listing := rootSplitRetainFile(t, listingPath)
		listing = []byte(strings.Replace(string(listing), removed+"\n", "", 1))
		if err := os.WriteFile(listingPath, listing, 0600); err != nil {
			t.Fatal(err)
		}
		if err := ciWriteJSON(path, suite); err != nil {
			t.Fatal(err)
		}
		if _, err := checkedModuleArtifacts(expected, inputs); err == nil ||
			!strings.Contains(err.Error(), "fresh complete root denominator") {
			t.Fatalf("producer-paired deletion bypassed fresh denominator: %v", err)
		}
	})
	for _, suffix := range []string{"stdout", "stderr", "cover"} {
		t.Run("unowned-artifact-"+suffix, func(t *testing.T) {
			path := filepath.Join(packagePath, "shard-0001."+suffix)
			if err := os.WriteFile(path, []byte("foreign\n"), 0600); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if err := os.Remove(path); err != nil {
					t.Error(err)
				}
			})
			rootSplitRequireRejected(t, expected, inputs)
		})
	}
	t.Run("extra-part", func(t *testing.T) {
		path := filepath.Join(inputs, "owner-0000", "part-0016")
		if err := os.Mkdir(path, 0700); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			if err := os.Remove(path); err != nil {
				t.Error(err)
			}
		})
		rootSplitRequireRejected(t, expected, inputs)
	})
	if _, err := checkedModuleArtifacts(expected, inputs); err != nil {
		t.Fatalf("independent controls changed original evidence: %v", err)
	}
}

func TestModuleRootPartitionActualSparseAndEmptyPackageRoles(t *testing.T) {
	moduleFixtureEnvironment(t)
	binary := parallelFixtureCLI(t)
	for _, role := range []string{"sparse", "benchmark", "no-test"} {
		t.Run(role, func(t *testing.T) {
			dir := moduleFixture(t, false)
			path := filepath.Join(dir, "value_test.go")
			if role == "no-test" {
				if err := os.Remove(path); err != nil {
					t.Fatal(err)
				}
			} else if role == "benchmark" {
				source := `package fixture
import ("flag"; "fmt"; "os"; "testing")
func TestMain(m *testing.M) {
    flag.Parse()
    if flag.Lookup("test.list").Value.String() == "" {
        fmt.Println("ROOT_UNFILTERED_EXECUTION")
    }
    os.Exit(m.Run())
}
func BenchmarkNotRun(b *testing.B) { b.Fatal("benchmark executed") }
`
				if err := os.WriteFile(path, []byte(source), 0600); err != nil {
					t.Fatal(err)
				}
			}
			inputs := t.TempDir()
			if err := os.Mkdir(filepath.Join(inputs, "owner-0000"), 0700); err != nil {
				t.Fatal(err)
			}
			executed, empty, passedRoots := 0, 0, 0
			for part := 0; part < 8; part++ {
				out := filepath.Join(inputs, "owner-0000", fmt.Sprintf("part-%04d", part))
				runRootSplitFixture(t, binary, dir, out, "", "", 0, part, 8)
				var suite suiteReport
				if err := ciReadJSON(filepath.Join(out, "package-0000", "partition.json"), &suite); err != nil {
					t.Fatal(err)
				}
				passedRoots += suite.RootPassed
				if suite.Status == "completed_empty" {
					empty++
				}
				for _, shard := range suite.Shards {
					if len(shard.Command) == 0 {
						continue
					}
					executed++
					if role != "sparse" && (part != 0 || !suite.FullPackage || len(suite.Shards) != 1 ||
						len(shard.Outcomes) != 0 || strings.Contains(strings.Join(shard.Command, " "), "-run=")) {
						t.Fatalf("empty inventory gained filtered/foreign/named execution: %+v", suite)
					}
					if role == "benchmark" {
						data, err := os.ReadFile(filepath.Join(out, "package-0000", "shard-0000.stdout"))
						if err != nil || strings.Count(string(data), "ROOT_UNFILTERED_EXECUTION") != 1 {
							t.Fatalf("part zero did not execute exactly one actual unfiltered TestMain: %v\n%s", err, data)
						}
					}
				}
				listing, err := os.ReadFile(filepath.Join(out, "package-0000", "inventory.stdout"))
				if err != nil || strings.Contains(string(listing), "ROOT_UNFILTERED_EXECUTION") {
					t.Fatalf("listing observation was mislabeled as execution: %v\n%s", err, listing)
				}
			}
			if role == "sparse" {
				if empty == 0 || executed > 4 || passedRoots != 4 {
					t.Fatalf("sparse assignment fabricated execution: commands=%d empty=%d roots=%d", executed, empty, passedRoots)
				}
			} else if executed != 1 || empty != 7 || passedRoots != 0 {
				t.Fatalf("empty package was executed repeatedly or invented test passes: commands=%d empty=%d roots=%d", executed, empty, passedRoots)
			}
			for owner := 1; owner < moduleOwners; owner++ {
				out := filepath.Join(inputs, fmt.Sprintf("owner-%04d", owner))
				runRootSplitFixture(t, binary, dir, out, "", "", owner, -1, 8)
			}
			out := filepath.Join(t.TempDir(), "aggregate")
			runRootSplitFixture(t, binary, dir, out, inputs, "", -1, -1, 8)
			var report moduleReport
			if err := ciReadJSON(filepath.Join(out, "module.json"), &report); err != nil {
				t.Fatal(err)
			}
			if report.Status != "passed" || report.RootInventory == nil ||
				report.RootInventory.FullPackage != (role != "sparse") ||
				report.RootInventory.NoTestFiles != (role == "no-test") {
				t.Fatalf("fresh full-package role differs: %+v", report)
			}
		})
	}
}
