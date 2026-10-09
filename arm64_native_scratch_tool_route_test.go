package plan9asm

import (
	"crypto/sha256"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xgo-dev/plan9asm/internal/goasmcmd"
)

func TestARM64NativeScratchActualAliasedBuiltinToolCache(t *testing.T) {
	pkg, request := arm64ScratchActualSourceFixture(t, "linux")
	ctx, cancel := runtimeOracleStageContext(t)
	defer cancel()
	query, err := goasmcmd.Observe(ctx, request.Directory, os.Environ(),
		request.GoBinary, "env", "GOCACHE")
	if err != nil {
		t.Fatal(err)
	}
	cache, err := filepath.EvalSymlinks(strings.TrimSpace(string(query.Stdout)))
	if err != nil || !filepath.IsAbs(cache) {
		t.Fatalf("actual stock Go cache identity: %q %v", cache, err)
	}
	alias := filepath.Join(t.TempDir(), "actual-go-cache")
	if err := os.Symlink(cache, alias); err != nil {
		t.Fatal(err)
	}
	request.Environment = append(request.Environment, "GOCACHE="+alias)
	observations := make(map[string]goasmcmd.Observation)
	for _, tool := range []string{"asm", "compile", "objdump", "pack"} {
		observed, err := goasmcmd.Observe(ctx, request.Directory, request.Environment,
			request.GoBinary, "tool", "-n", tool)
		if err != nil {
			t.Fatalf("actual stock %s route: %v", tool, err)
		}
		observations[tool] = observed
		encoded, err := json.Marshal(observed)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("SCRATCH_TOOL_CHILD %s", encoded)
	}
	source, err := CaptureGoARM64ScratchSource(ctx, pkg, request)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := source.Close(); err != nil {
			t.Error(err)
		}
	})
	for tool, observed := range observations {
		original := strings.TrimSpace(string(observed.Stdout))
		physical, err := filepath.EvalSymlinks(original)
		if err != nil {
			t.Fatal(err)
		}
		data, err := goAMD64FrameRead(physical)
		if err != nil {
			t.Fatal(err)
		}
		if digest, present := source.inputs[physical]; !present || digest != sha256.Sum256(data) {
			t.Fatalf("actual %s binary bytes were not retained at their physical route", tool)
		}
	}
	if err := source.Verify(ctx); err != nil {
		t.Fatal(err)
	}
}
