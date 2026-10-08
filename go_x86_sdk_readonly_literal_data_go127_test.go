//go:build go1.27

package plan9asm

import "testing"

// These are two different actual SDK packages, not library-name allowlists.
// Complete source/header/CPP/object observation must close before the new
// initializer factory is asked to establish its independent role.
func TestGoX86SDKReadonlyLiteralDataActualSources(t *testing.T) {
	for _, pkg := range []string{"hash/crc32", "internal/chacha8rand"} {
		t.Run(pkg, func(t *testing.T) {
			ctx, request := sdkASMDataRequest(t, "amd64", pkg)
			source, err := CaptureGoX86SDKASMDataSource(ctx, request)
			if err != nil {
				t.Fatal("actual complete SDK observation precondition", err)
			}
			t.Cleanup(func() { source.Close() })
			sdkASMDataRetain(t, source)
			owner, err := CaptureGoX86SDKReadonlyLiteralDataOwner(ctx, source)
			if err != nil || owner == nil {
				t.Fatal("actual SDK private readonly literal initializer is not certified", err)
			}
		})
	}
}
