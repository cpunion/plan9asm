package plan9asm

import (
	"os"
	"strings"
	"testing"
)

func TestCIModuleSuitesKeepExhaustiveUnixAndWindowsRoles(t *testing.T) {
	data, err := os.ReadFile(".github/workflows/go-ci.yml")
	if err != nil {
		t.Fatal(err)
	}
	script, err := os.ReadFile("scripts/test-module-ci.sh")
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range []struct{ job, flags string }{
		{"test", "test) timeout=45m"},
		{"race", "race) timeout=60m; flags=(-race)"},
		{"coverage", "coverage) timeout=45m; flags=(-coverprofile=coverage-root.out)"},
	} {
		job, found := ciWorkflowJob(string(data), item.job)
		command := "bash scripts/test-module-ci.sh " + item.job + " execute ${{ matrix.module-owner }}"
		if !found || !strings.Contains(job, command) || !strings.Contains(string(script), item.flags) {
			t.Errorf("%s must keep the complete actual module inventory/flags using bounded partition ownership", item.job)
		}
		for _, required := range []string{"module-owner: [0, 1, 2]", "fail-fast: false", "actions/upload-artifact@v6"} {
			if !strings.Contains(job, required) {
				t.Errorf("%s missing exhaustive scheduling/artifact requirement %q", item.job, required)
			}
		}
	}
	aggregate, found := ciWorkflowJob(string(data), "module-completion")
	if !found || !strings.Contains(aggregate, "needs: [test, race, coverage]") || !strings.Contains(aggregate, "!cancelled()") ||
		!strings.Contains(aggregate, "owner-0000") || !strings.Contains(aggregate, "owner-0001") || !strings.Contains(aggregate, "owner-0002") {
		t.Error("missing mandatory complete cross-owner readback")
	}
	for _, required := range []string{"-module -out=", "-shards=32 -workers=2", "-verify-module=_out/module-$flavor-owners", "go run ./cmd/plan9asmtest"} {
		if !strings.Contains(string(script), required) {
			t.Errorf("module driver changed complete inventory/scheduling: %s", required)
		}
	}
	test, _ := ciWorkflowJob(string(data), "test")
	for _, command := range []string{
		"go test -timeout=30m ./... -coverprofile=coverage-windows-root.out",
		"(cd cmd/plan9asm && go test -timeout=30m ./... -coverprofile=../../coverage-windows-plan9asm.out)",
		"(cd cmd/plan9asmll && go test -timeout=30m ./... -coverprofile=../../coverage-windows-plan9asmll.out)",
	} {
		if !strings.Contains(test, command) {
			t.Errorf("existing full Windows role changed: %s", command)
		}
	}
	if !strings.Contains(test, "!cancelled() && runner.os != 'Windows' && matrix.nested_modules && matrix.module-owner == 2") {
		t.Error("both nested CLI roles must still execute after a main-module failure")
	}
	coverage, _ := ciWorkflowJob(string(data), "coverage")
	if strings.Contains(coverage, "codecov/codecov-action") || !strings.Contains(aggregate, "files: ./coverage-root.out") {
		t.Error("coverage upload must use complete module union, never one owner's partial file")
	}
}
