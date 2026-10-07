package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// The initial test-only adapter executes the existing full reader/retry path.
// Replace it only after its real bounded/fairness counterexamples run RED.
func discoveryRetryQueueUnderTest(root string, limit int) ([]moduleVersion, error) {
	report, err := readShardedReport(root)
	if err != nil {
		return nil, err
	}
	return modulesForFailureRetry(report), nil
}

func TestDiscoveryRetryQueueBoundedDefaultMajorAndRotation(t *testing.T) {
	root := filepath.Join(t.TempDir(), "ledger")
	base := discoveryReport{Failures: []scanFailure{
		{Module: "example.invalid/a", Version: "v0.9.0", Error: "owned failure"},
		{Module: "example.invalid/a", Version: "v1.9.0", Error: "owned older failure"},
		{Module: "example.invalid/a", Version: "v1.10.0", Error: "owned latest line failure"},
		{Module: "example.invalid/a", Version: "@latest", Error: "owned resolution failure"},
		{Module: "example.invalid/b", Version: "v1.0.0", Error: "owned failure"},
	}}
	if err := writeShardedReport(root, base); err != nil {
		t.Fatal(err)
	}
	all := []moduleVersion{{Path: "example.invalid/a", Version: "v0.9.0"}, {Path: "example.invalid/a", Version: "v1.10.0"},
		{Path: "example.invalid/a", Version: "@latest"}, {Path: "example.invalid/b", Version: "v1.0.0"}}
	for _, test := range []struct {
		name   string
		cursor *moduleVersion
		want   []moduleVersion
	}{
		{name: "first", want: all[:2]},
		{name: "same-major old cursor", cursor: &moduleVersion{Path: "example.invalid/a", Version: "v1.9.0"}, want: all[1:3]},
		{name: "last wraps", cursor: &all[3], want: all[:2]},
		{name: "removed boundary", cursor: &moduleVersion{Path: "example.invalid/az", Version: "v1.0.0"}, want: []moduleVersion{all[3], all[0]}},
	} {
		t.Run(test.name, func(t *testing.T) {
			discoveryRetrySetManifestCursor(t, root, test.cursor)
			got, err := discoveryRetryQueueUnderTest(root, 2)
			if err != nil || !reflect.DeepEqual(got, test.want) || cap(got) > 2 {
				t.Fatalf("bounded major/rotation queue=%+v cap=%d want=%+v error=%v", got, cap(got), test.want, err)
			}
		})
	}
	if got, err := discoveryRetryQueueUnderTest(root, -1); err == nil || got != nil {
		t.Fatalf("negative retry budget accepted: %+v %v", got, err)
	}
	large := filepath.Join(t.TempDir(), "ledger")
	var failures []scanFailure
	for i := 0; i < defaultScanLimit+17; i++ {
		failures = append(failures, scanFailure{Module: fmt.Sprintf("example.invalid/%05d", i), Version: "v1.0.0", Error: "owned failure"})
	}
	if err := writeShardedReport(large, discoveryReport{Failures: failures}); err != nil {
		t.Fatal(err)
	}
	if got, err := discoveryRetryQueueUnderTest(large, 0); err != nil || len(got) != defaultScanLimit || cap(got) > defaultScanLimit {
		t.Fatalf("zero/default budget escaped bound: len=%d cap=%d error=%v", len(got), cap(got), err)
	}
}

func TestDiscoveryRetryQueueDrainsUnselectedTailAndCounts(t *testing.T) {
	for _, damage := range []string{"tail", "counts"} {
		t.Run(damage, func(t *testing.T) {
			root := filepath.Join(t.TempDir(), "ledger")
			if err := writeShardedReport(root, discoveryReport{Failures: []scanFailure{
				{Module: "example.invalid/a", Version: "v1.0.0", Error: "owned failure"},
				{Module: "example.invalid/b", Version: "v1.0.0", Error: "owned failure"},
			}}); err != nil {
				t.Fatal(err)
			}
			if damage == "tail" {
				if err := os.WriteFile(filepath.Join(root, "records", "ff.jsonl"), []byte("owned invalid unselected tail\n"), 0o600); err != nil {
					t.Fatal(err)
				}
			} else {
				discoveryRetryEditManifest(t, root, func(manifest map[string]any) { manifest["failure_records"] = float64(3) })
			}
			if got, err := discoveryRetryQueueUnderTest(root, 1); err == nil || got != nil {
				t.Fatalf("bounded queue exposed provisional records before full validation: %+v %v", got, err)
			}
		})
	}
}

func TestDiscoveryRetryMetadataOptionalStrictAndMerge(t *testing.T) {
	root := filepath.Join(t.TempDir(), "ledger")
	if err := writeShardedReport(root, discoveryReport{}); err != nil {
		t.Fatal(err)
	}
	cursor := moduleVersion{Path: "example.invalid/gone", Version: "@latest"}
	discoveryRetrySetManifestCursor(t, root, &cursor)
	scan, err := scanShardedReport(root, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got := discoveryRetryJSONCursor(t, scan.metadata); !reflect.DeepEqual(got, &cursor) {
		t.Fatalf("streaming metadata discarded optional retry marker: %+v", got)
	}
	for _, invalid := range []moduleVersion{{Path: "only-path"}, {Version: "v1.0.0"}, {}} {
		discoveryRetrySetManifestCursor(t, root, &invalid)
		if _, err := scanShardedReport(root, nil); err == nil {
			t.Fatalf("persisted half/empty retry marker accepted: %+v", invalid)
		}
	}
	discoveryRetrySetManifestCursor(t, root, nil)
	if _, err := scanShardedReport(root, nil); err != nil {
		t.Fatal("old schema2 manifest without retry marker is not compatible", err)
	}
	base := discoveryRetryReportWithCursor(t, cursor)
	for _, test := range []struct {
		name   string
		update discoveryReport
		want   *moduleVersion
	}{
		{name: "preserve", want: &cursor},
		{name: "clear", update: discoveryRetryReportWithCursor(t, moduleVersion{})},
		{name: "replace", update: discoveryRetryReportWithCursor(t, moduleVersion{Path: "example.invalid/new", Version: "v1.10.0"}),
			want: &moduleVersion{Path: "example.invalid/new", Version: "v1.10.0"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			merged, err := mergeDiscoveryReports(base, test.update)
			if err != nil || !reflect.DeepEqual(discoveryRetryJSONCursor(t, merged), test.want) {
				t.Fatalf("retry metadata merge lost preserve/clear/replace: %+v %v", merged, err)
			}
		})
	}
	for _, bad := range []moduleVersion{{Path: "only-path"}, {Version: "v1.0.0"}} {
		if _, err := mergeDiscoveryReports(base, discoveryRetryReportWithCursor(t, bad)); err == nil {
			t.Fatalf("merge accepted half retry marker %+v", bad)
		}
	}
}

func discoveryRetryEditManifest(t *testing.T, root string, edit func(map[string]any)) {
	t.Helper()
	file := filepath.Join(root, "manifest.json")
	data, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	var manifest map[string]any
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatal(err)
	}
	edit(manifest)
	data, err = json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file, data, 0o600); err != nil {
		t.Fatal(err)
	}
}

func discoveryRetrySetManifestCursor(t *testing.T, root string, cursor *moduleVersion) {
	t.Helper()
	discoveryRetryEditManifest(t, root, func(manifest map[string]any) {
		if cursor == nil {
			delete(manifest, "retry_after")
		} else {
			manifest["retry_after"] = cursor
		}
	})
}

func discoveryRetryJSONCursor(t *testing.T, value any) *moduleVersion {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	var result struct {
		Cursor *moduleVersion `json:"retry_after"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatal(err)
	}
	return result.Cursor
}

func discoveryRetryReportWithCursor(t *testing.T, cursor moduleVersion) discoveryReport {
	t.Helper()
	data, err := json.Marshal(map[string]any{"retry_after": cursor})
	if err != nil {
		t.Fatal(err)
	}
	var report discoveryReport
	if err := json.Unmarshal(data, &report); err != nil {
		t.Fatal(err)
	}
	return report
}
