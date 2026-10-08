//go:build go1.27

package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/xgo-dev/plan9asm/internal/gotoolprofile"
)

// The generated file-proxy module is an owned mechanism fixture, not an
// authenticated external-library pass. Its original cache source stays intact.
func TestAMD64DownloadedFixedLocalActualProfileCLI(t *testing.T) {
	_, assembly, config := amd64FixedLocalCLIFixture(t, "PADDL 16(SP),X0")
	ctx := config.Context
	input := &featureInput{
		Protocol: featureInputProtocol,
		Module:   "example.invalid/fixedlocal", Version: "v1.0.0",
		SourceRoot: filepath.Dir(assembly),
		Sources:    make(map[string]string), Headers: make(map[string]string),
		Directories: make(map[string][]string), AsmFiles: []string{"value_amd64.s"},
	}
	for _, name := range []string{"go.mod", "value.go", "value_amd64.s"} {
		data, err := os.ReadFile(filepath.Join(input.SourceRoot, name))
		if err != nil {
			t.Fatal(err)
		}
		input.Sources[name] = featureBytesSHA256(data)
		input.Directories["."] = append(input.Directories["."], name)
		if name != "go.mod" {
			input.Headers[name] = string(data)
		}
	}
	sort.Strings(input.Directories["."])
	modeInput := &gotoolprofile.InstrumentationConsumerInput{Original: input}
	invocation, env := instrumentationVersionedFixture(t, ctx, modeInput)
	input = modeInput.Original
	for _, value := range env {
		key, value, valid := strings.Cut(value, "=")
		if !valid {
			t.Fatal("malformed exact-download fixture environment")
		}
		t.Setenv(key, value)
	}
	binary, err := exec.LookPath("go")
	if err != nil {
		t.Fatal(err)
	}
	binary, err = filepath.Abs(binary)
	if err != nil {
		t.Fatal(err)
	}
	input.Observed, err = gotoolprofile.Capture(ctx, binary, t.TempDir(), env,
		"linux/amd64", nil, gotoolprofile.RunBounded)
	if err != nil {
		t.Fatal(err)
	}
	input.ID = gotoolprofile.ProfileID(input.Observed)
	stdout, stderr, err := gotoolprofile.RunBounded(ctx, invocation, env, binary,
		"list", "-export", "-compiled", "-json", "-mod=readonly", input.Module)
	if err != nil || len(stderr) != 0 {
		t.Fatalf("actual exact-cache original Go export failed: %v\n%s%s", err, stdout, stderr)
	}
	t.Logf("ACTUAL_EXACT_CACHE_ORIGINAL_GO_EXPORT\n%s", stdout)
	for name, expected := range input.Sources {
		actual, err := gotoolprofile.FileSHA256(filepath.Join(input.SourceRoot, name))
		if err != nil || actual != expected {
			t.Fatalf("original cache source changed before CLI: %s: %v", name, err)
		}
	}

	t.Chdir(invocation)
	config.FeaturePath = writeFeatureInput(t, input)
	output := t.TempDir()
	report, _, err := runOneTarget(targetSpec{Goos: "linux", Goarch: "amd64"},
		[]string{input.Module}, nil, input.AsmFiles, input.Module, output,
		false, 0, true, false, true, "", config)
	if err != nil || report.Success != 1 || report.Failed != 0 {
		t.Fatalf("actual original Go object compiled, but exact downloaded profile CLI has no fixed-local consumer: %v; %#v", err, report)
	}
	if err := gotoolprofile.ValidateSelection(input, report.FeatureSelection, nil, true); err != nil {
		t.Fatal(err)
	}
	if len(report.FeatureSelection.CPP) != 1 || len(report.FeatureSelection.Outputs) != 1 ||
		report.FeatureSelection.Outputs[0].IR == "" || report.FeatureSelection.Outputs[0].Object == "" {
		t.Fatal("exact downloaded CLI omitted its actual CPP/LLVM object consumption")
	}
	for name, expected := range input.Sources {
		actual, err := gotoolprofile.FileSHA256(filepath.Join(input.SourceRoot, name))
		if err != nil || actual != expected {
			t.Fatalf("CLI rewrote original cache source: %s: %v", name, err)
		}
	}
}
