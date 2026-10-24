package plan9asm

import (
	"os/exec"
	"strings"
	"testing"
)

// Own one disposable compiler cache for the complete root. Tool construction
// precedes every source lease; the setup fixture creates no source observation.
func goAMD64FramePrepareOwnedToolCache(t *testing.T) {
	t.Helper()
	t.Setenv("GOCACHE", t.TempDir())
	_, request := goAMD64FrameSourceFixture(t)
	fixedLocalConversionPrepareToolCache(t, request)
}

// Tool construction is setup, not a source/object observation. In particular,
// this output cannot replace Capture's independent bounded tool query or seals.
func fixedLocalConversionPrepareToolCache(t *testing.T, request GoAMD64FrameCaptureRequest) {
	t.Helper()
	cache := ""
	for _, entry := range request.Environment {
		key, value, _ := strings.Cut(entry, "=")
		if key == "GOCACHE" {
			if cache != "" || value == "" {
				t.Fatal("conversion setup requires one explicit owned compiler cache")
			}
			cache = value
		}
	}
	if cache == "" || request.GoBinary == "" || request.Directory == "" {
		t.Fatal("conversion setup lost its actual tool, directory or compiler cache")
	}
	ctx, cancel := runtimeOracleStageContext(t)
	defer cancel()
	command := exec.CommandContext(ctx, request.GoBinary, "tool", "-n", "objdump")
	command.Dir, command.Env = request.Directory, request.Environment
	if output, err := runtimeOracleCombinedOutput(ctx, command); err != nil {
		t.Fatalf("prepare actual Go objdump before source capture: %v\n%s", err, output)
	}
}
