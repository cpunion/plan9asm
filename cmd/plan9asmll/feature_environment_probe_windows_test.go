//go:build windows

package main

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/xgo-dev/plan9asm/gotoolprofile"
)

// This test belongs only to the dated fork diagnostic branch.
func TestWindowsFeatureEnvironmentNormalizationProbe(t *testing.T) {
	input, directory := featureConsumerFixture(t)
	t.Chdir(directory)
	binary, err := exec.LookPath("go")
	if err != nil {
		t.Fatal(err)
	}
	binary, err = filepath.Abs(binary)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	raw := featureEnvironment(os.Environ(), input.Observed.Environment)
	output, stderr, err := gotoolprofile.RunBounded(ctx, directory, raw, binary,
		"env", "-json", "GOROOT", "GOCACHE", "GOMODCACHE", "GOPATH", "GOENV")
	if err != nil || len(stderr) != 0 {
		t.Fatalf("original environment: %v; %s", err, stderr)
	}
	var actual map[string]string
	if err := json.Unmarshal(output, &actual); err != nil {
		t.Fatal(err)
	}
	normalized, err := featurePackageEnvironment("linux", "amd64", raw)
	if err != nil {
		t.Fatal(err)
	}
	restore := func(key string) []string {
		var environment []string
		for _, item := range normalized {
			name, _, _ := strings.Cut(item, "=")
			if !strings.EqualFold(name, key) {
				environment = append(environment, item)
			}
		}
		return append(environment, key+"="+actual[key])
	}
	for _, row := range []struct {
		name string
		env  []string
	}{
		{"original", raw},
		{"normalized", normalized},
		{"original-GOROOT", restore("GOROOT")},
		{"original-GOCACHE", restore("GOCACHE")},
	} {
		t.Run(row.name, func(t *testing.T) {
			observed, err := gotoolprofile.Capture(ctx, binary, t.TempDir(), row.env,
				"linux/amd64", nil, gotoolprofile.RunBounded)
			if err != nil {
				t.Fatal(err)
			}
			before, _ := json.Marshal(input.Observed)
			after, _ := json.Marshal(observed)
			t.Logf("original profile: %s", before)
			t.Logf("reobserved profile: %s", after)
			if !reflect.DeepEqual(input.Observed, observed) {
				t.Error("unchanged actual SDK acquired a different feature profile")
			}
		})
	}
}
