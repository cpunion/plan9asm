//go:build go1.27

package plan9asm

import "testing"

// This test-only RED asks the existing producer to observe the complete real
// SDK package. Its r2r1 Context is the class boundary the separate observation
// producer must cross, without granting the existing producer read authority.
func TestGoX86SDKASMDataLiveOriginalCRC32Red(t *testing.T) {
	ctx, request, _ := numericalSDKCRC32Request(t)
	source, err := CaptureGoX86SDKNumericalReadSource(ctx, request)
	if source != nil {
		defer source.Close()
	}
	if err != nil {
		t.Fatal("actual original SDK DATA observation is not closed", err)
	}
}
