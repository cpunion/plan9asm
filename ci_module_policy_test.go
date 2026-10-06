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
	for _, item := range []struct{ job, flags string }{
		{"test", "-timeout=45m"},
		{"race", "-timeout=60m -- -race"},
		{"coverage", "-timeout=45m -- -coverprofile=coverage-root.out"},
	} {
		job, found := ciWorkflowJob(string(data), item.job)
		command := "go run ./cmd/plan9asmtest -module -out=_out/module-" + item.job + " -shards=32 -workers=2 " + item.flags
		if !found || !strings.Contains(job, command) {
			t.Errorf("%s must keep the complete actual module inventory/flags using bounded partition ownership", item.job)
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
	if !strings.Contains(test, "!cancelled() && runner.os != 'Windows' && matrix.nested_modules") {
		t.Error("both nested CLI roles must still execute after a main-module failure")
	}
}
