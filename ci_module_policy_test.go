package plan9asm

import (
	"fmt"
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
	for _, item := range []struct {
		job, flags string
		parts      int
	}{
		{"test", "test) timeout=45m", 8},
		{"race", "race) timeout=60m; flags=(-race); root_partitions=16", 16},
		{"coverage", "coverage) timeout=45m; flags=(-coverprofile=coverage-root.out)", 8},
	} {
		job, found := ciWorkflowJob(string(data), item.job)
		command := "bash scripts/test-module-ci.sh " + item.job + " execute ${{ matrix.module-owner }} ${{ matrix.root-partition }}"
		if !found || !strings.Contains(job, command) || !strings.Contains(string(script), item.flags) {
			t.Errorf("%s must keep the complete actual module inventory/flags using bounded partition ownership", item.job)
		}
		var roles []string
		for part := 0; part < item.parts; part++ {
			roles = append(roles, fmt.Sprintf("root-%d", part))
		}
		roles = append(roles, "profile", "other")
		for _, required := range []string{"module-role: [" + strings.Join(roles, ", ") + "]", "fail-fast: false", "actions/upload-artifact@v6"} {
			if !strings.Contains(job, required) {
				t.Errorf("%s missing exhaustive scheduling/artifact requirement %q", item.job, required)
			}
		}
		if strings.Contains(job, "module-owner: [") || strings.Contains(job, "max-parallel:") || strings.Contains(job, "continue-on-error:") {
			t.Errorf("%s repeats package owners or masks/limits scheduling", item.job)
		}
	}
	aggregate, found := ciWorkflowJob(string(data), "module-completion")
	if !found || !strings.Contains(aggregate, "needs: [test, race, coverage]") || !strings.Contains(aggregate, "!cancelled()") ||
		!strings.Contains(aggregate, "owner-0000") || !strings.Contains(aggregate, "owner-0001") || !strings.Contains(aggregate, "owner-0002") {
		t.Error("missing mandatory complete cross-owner readback")
	}
	for part := 0; part < 16; part++ {
		if !strings.Contains(aggregate, fmt.Sprintf("Download root part %d", part)) ||
			!strings.Contains(aggregate, fmt.Sprintf("-owners/owner-0000/part-%04d", part)) {
			t.Errorf("missing exact root part %d download", part)
		}
	}
	for _, required := range []string{"-module -out=", "-shards=32 -workers=2", "root_partitions=8", "-module-root-partitions=$root_partitions", "-verify-module=_out/module-$flavor-owners", "go run ./cmd/plan9asmtest"} {
		if !strings.Contains(string(script), required) {
			t.Errorf("module driver changed complete inventory/scheduling: %s", required)
		}
	}
	test, _ := ciWorkflowJob(string(data), "test")
	if !strings.Contains(test, "go test -count=1 -timeout=30m ./... -coverprofile=coverage-windows-root.out") {
		t.Error("Windows main-module suite must remain complete")
	}
	windowsCLI, found := ciWorkflowJob(string(data), "windows-cli")
	if !found || !strings.Contains(windowsCLI, "needs: ci_policy") ||
		!strings.Contains(windowsCLI, "command: [plan9asm, plan9asmll]") ||
		!strings.Contains(windowsCLI, "working-directory: cmd/${{ matrix.command }}") ||
		!strings.Contains(windowsCLI, "go test -count=1 -timeout=30m ./... -coverprofile=../../coverage-windows-${{ matrix.command }}.out") {
		t.Error("both complete Windows CLI suites need independent jobs, unaffected by root failures")
	}
	if !strings.Contains(windowsCLI, "working-directory: cmd/${{ matrix.command }}\n        shell: bash\n") ||
		!strings.Contains(windowsCLI, "test -s ../../coverage-windows-${{ matrix.command }}.out") {
		t.Error("Windows CLI coverage requires explicit Bash argument handling and a nonempty profile before upload")
	}
	if !strings.Contains(test, "| grep -Fx 'TestAMD64DownloadedFixedLocalLLVM22Numerical'") {
		t.Error("required numerical inventory must use the provisioned fixed-line filter, not optional ripgrep")
	}
	if !strings.Contains(test, "!cancelled() && runner.os != 'Windows' && matrix.nested_modules && matrix.module-owner == 2") {
		t.Error("both nested CLI roles must still execute after a main-module failure")
	}
	coverage, _ := ciWorkflowJob(string(data), "coverage")
	if strings.Contains(coverage, "codecov/codecov-action") || !strings.Contains(aggregate, "files: ./coverage-root.out") {
		t.Error("coverage upload must use complete module union, never one owner's partial file")
	}
}
