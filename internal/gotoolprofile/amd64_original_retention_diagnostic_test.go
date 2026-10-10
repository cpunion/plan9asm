package gotoolprofile

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAMD64OriginalObservationDriverBudgetPreservesCause(t *testing.T) {
	request := amd64OriginalObservationFixture(t, "driver-budget-cause")
	// These complete, inert package members fit the original source budget.
	// The separately selected real Go executable takes the combined graph
	// over its existing limit. No source, tool bytes or identity are changed.
	const sourceBytes = 4 * (13 << 20)
	info, err := os.Stat(request.GoBinary)
	if err != nil {
		t.Fatal(err)
	}
	if sourceBytes+info.Size() <= 64<<20 {
		t.Fatal("fixture no longer reaches the shared graph byte bound", info.Size())
	}
	for index := 0; index < 4; index++ {
		file, err := os.Create(filepath.Join(request.Directory, fmt.Sprintf("padding-%d.dat", index)))
		if err != nil {
			t.Fatal(err)
		}
		err = file.Truncate(13 << 20)
		closeErr := file.Close()
		if err != nil || closeErr != nil {
			t.Fatal(err, closeErr)
		}
	}
	ctx, cancel := amd64OriginalObservationContext(t)
	defer cancel()
	stdout, stderr, err := RunBounded(ctx, request.Directory, request.Environment,
		request.GoBinary, "list", "-export", "-json", "-mod=readonly", request.PackagePath)
	if err != nil || len(stdout) == 0 || len(stderr) != 0 {
		t.Fatalf("actual Go must accept unchanged first-party source: %v\n%s", err, stderr)
	}
	observed, err := CaptureAMD64OriginalPackageObservation(ctx, request, RunBounded)
	if observed != nil {
		_ = observed.Close()
		t.Fatal("over-budget observation acquired a source token")
	}
	if err == nil || !strings.Contains(err.Error(), "retain actual Go driver") ||
		!strings.Contains(err.Error(), "byte bound: retained=") || !strings.Contains(err.Error(), "limit=67108864") ||
		strings.Contains(err.Error(), "driver changed") {
		t.Fatalf("retention limit must not impersonate a driver identity mutation: %v", err)
	}
}
