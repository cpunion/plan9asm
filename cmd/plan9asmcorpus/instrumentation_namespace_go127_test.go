//go:build go1.27

package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/xgo-dev/plan9asm/internal/gotoolprofile"
)

func TestInstrumentationCorpusOriginalVersionedAliasReachesStrictObjects(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	translator := filepath.Join(t.TempDir(), "plan9asmll")
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
	defer cancel()
	_, stderr, err := gotoolprofile.RunBounded(ctx, filepath.Join(root, "cmd", "plan9asmll"), os.Environ(),
		"go", "build", "-p=1", "-o", translator, ".")
	if err != nil {
		t.Fatalf("build current translator: %v\n%s", err, stderr)
	}
	testInstrumentationCorpusOriginalMode(t, root, translator, "race", "1", "darwin/arm64", true,
		"example.invalid/instrumented-mirror")
}
