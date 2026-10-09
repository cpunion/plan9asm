//go:build go1.27

package main

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// A tag-selected Go body can be invalid on one architecture even though its
// assembly and complete generated-header properties are valid. Planning is
// only an optimization: this scope still needs the normal final Go check,
// and healthy target/tag siblings must remain eligible for their own objects.
func TestCPPHeaderPlanningDefersExactOriginalGoSourceRejection(t *testing.T) {
	compilerEnv := os.Environ()
	candidate := discoveryCandidate{
		Module: "example.invalid/header-rejection", Version: "v1.0.0",
		AsmFiles: []string{"probe_386.s", "probe_amd64.s"},
	}
	sources := map[string]string{
		"go.mod":        "module " + candidate.Module + "\ngo 1.20\n",
		"probe.go":      "package witness\nconst Answer=7\nfunc Probe()\n",
		"poison.go":     "//go:build durable && 386\n\npackage witness\nfunc unused() { _ = missingArchitectureValue }\n",
		"probe_386.s":   "#include \"go_asm.h\"\nTEXT ·Probe(SB),4,$0-0\nRET\n",
		"probe_amd64.s": "#include \"go_asm.h\"\nTEXT ·Probe(SB),4,$0-0\nRET\n",
	}
	seed := fixtureProfileModuleDownload(t, candidate, sources)
	installGeneratedFixtureProxy(t, candidate, seed, sources["go.mod"])
	cacheRoot := os.Getenv("GOMODCACHE")
	t.Cleanup(func() {
		if err := filepath.Walk(cacheRoot, func(file string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return os.Chmod(file, 0700)
			}
			return nil
		}); err != nil {
			t.Error(err)
		}
	})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	translator := filepath.Join(t.TempDir(), "plan9asmll")
	command := exec.CommandContext(ctx, "go", "build", "-p=1", "-o", translator, ".")
	command.Dir, command.Env = filepath.Join(root, "cmd/plan9asmll"), compilerEnv
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("build first-party metadata query consumer: %v\n%s", err, output)
	}
	goBinary, err := resolveDiscoveryExecutable("go")
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	data, _, err := runDiscoveryMachineCommand(ctx, dir, os.Environ(), goBinary, "mod", "download", "-json", candidate.exactKey())
	if err != nil {
		t.Fatal(err)
	}
	var download moduleDownloadInfo
	if err := json.Unmarshal(data, &download); err != nil || download.Dir == "" {
		t.Fatalf("actual first-party proxy download: %s; %v", data, err)
	}
	plan, err := captureOrdinarySelectionInputs(candidate, download.Dir, []string{"linux/386", "linux/amd64"})
	if err != nil {
		t.Fatal(err)
	}
	plan.ModuleSum = download.Sum
	plan.ZipSHA256, err = discoveryFeatureFileSHA256(download.Zip)
	if err != nil {
		t.Fatal(err)
	}
	if err := installDiscoveryHeaderPlanning(ctx, plan, candidate, download, dir, runtime.GOROOT(), goBinary,
		t.TempDir(), translator, os.Environ(), &discoveryFeatureObservationCache{}); err != nil {
		t.Fatal(err)
	}
	badRequest := discoveryFeatureProfileRequest{Target: "linux/386", Baseline: true}
	header, err := plan.headerPlanningCapture(badRequest, "probe_386.s", []string{"durable"})
	var unsupported *discoveryHeaderPlanningUnsupported
	if header != nil || !errors.As(err, &unsupported) {
		t.Fatalf("original target/tag Go rejection must defer only planning, not abort the whole candidate: %v", err)
	}
	if plan.HeaderPlanning != nil || len(plan.headerPlanningInventory.Observations) != 0 {
		t.Fatal("failed planning query manufactured a successful header/profile witness")
	}
	for _, scope := range []struct {
		target, assembly string
		tags             []string
		bad              bool
	}{
		{"linux/386", "probe_386.s", []string{"durable"}, true},
		{"linux/386", "probe_386.s", nil, false},
		{"linux/amd64", "probe_amd64.s", []string{"durable"}, false},
	} {
		t.Run(scope.target+"/"+strings.Join(scope.tags, ","), func(t *testing.T) {
			groups := []discoveryPackageGroup{{Pattern: candidate.Module, AsmFiles: []string{scope.assembly}}}
			checks := 0
			outcomes, err := runDiscoveryPackageChecks(groups, func(patterns []string) error {
				checks++
				return runDiscoveryGoBuild(ctx, download.Dir, os.Environ(), scope.target, scope.tags, patterns...)
			})
			if err != nil || checks != 1 || len(outcomes) != 1 {
				t.Fatalf("mandatory final source check was not independently executed: checks=%d outcomes=%v error=%v", checks, outcomes, err)
			}
			if scope.bad {
				diagnostic := discoveryCommandDiagnostic(outcomes[0])
				if outcomes[0] == nil || !strings.Contains(diagnostic, "undefined: missingArchitectureValue") {
					t.Fatalf("exact original Go rejection disappeared: %v", outcomes[0])
				}
				item := discoverySourceNotApplicableItem{
					Kind: discoverySourceNotApplicableGoBuild, AsmFiles: groups[0].AsmFiles,
					Targets: []string{scope.target}, Reason: diagnostic,
					Diagnostic: captureDiscoverySourceDiagnostic(diagnostic),
				}
				if err := validateDiscoverySourceRejectionDiagnostic(item, false); err != nil {
					t.Fatalf("final rejection has no independently structured diagnostic: %v", err)
				}
				t.Logf("FINAL_ORIGINAL_SOURCE_REJECTION %s tags=%v diagnostic=%s", scope.target, scope.tags, item.Diagnostic.SHA256)
				return
			}
			if outcomes[0] != nil {
				t.Fatalf("healthy target/tag sibling was excluded: %v", outcomes[0])
			}
			request := discoveryFeatureProfileRequest{Target: scope.target, Baseline: true}
			if header, err := plan.headerPlanningCapture(request, scope.assembly, scope.tags); err != nil || header == nil {
				t.Fatalf("eligible sibling still needs its own actual header producer: %v", err)
			}
		})
	}
	if plan.HeaderPlanning == nil || len(plan.HeaderPlanning.Queries) != 2 {
		t.Fatal("only the two independently compiled healthy scopes may publish planning headers")
	}
}
