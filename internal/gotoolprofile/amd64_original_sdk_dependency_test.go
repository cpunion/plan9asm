package gotoolprofile

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// SDK dependencies are actual source/export inputs, not ordinary third-party
// ASM packages or owners of the requested function's private local frame.
// A routine import must not make their special macros impersonate an ordinary
// package role. The unchanged original Go compiler accepts this complete graph.
func TestAMD64OriginalObservationActualSDKDependencyRoles(t *testing.T) {
	for _, target := range []string{"linux/amd64", "darwin/amd64", "windows/amd64"} {
		t.Run(target, func(t *testing.T) {
			request := amd64OriginalObservationFixture(t, "bodyful")
			request.Target = target
			goos, _, _ := strings.Cut(target, "/")
			request.Environment = replaceEnv(request.Environment, map[string]string{"GOOS": goos})
			source := "package observation\nimport \"errors\"\n" +
				"func Value(a uint64) uint64\n" +
				"func Error() error { return errors.New(\"first-party witness\") }\n"
			if err := os.WriteFile(filepath.Join(request.Directory, "value.go"), []byte(source), 0600); err != nil {
				t.Fatal(err)
			}
			ctx, cancel := amd64OriginalObservationContext(t)
			defer cancel()
			output, diagnostic, err := RunBounded(ctx, request.Directory, request.Environment,
				request.GoBinary, "list", "-export", "-mod=readonly", ".")
			if err != nil {
				t.Fatalf("original Go rejects SDK-dependent first-party source: %v\n%s\n%s", err, output, diagnostic)
			}
			t.Logf("ORIGINAL_SDK_DEPENDENCY_EXPORT target=%s package=%s", target, strings.TrimSpace(string(output)))
			observed, err := CaptureAMD64OriginalPackageObservation(ctx, request, RunBounded)
			if err != nil {
				t.Fatalf("ordinary root must retain SDK dependencies under their own source/export role: %v", err)
			}
			defer observed.Close()
			if observed.state.packages["runtime"].Export == "" || observed.state.packages["errors"].Export == "" {
				t.Fatal("SDK graph vanished instead of retaining its actual compiler exports")
			}
		})
	}
}
