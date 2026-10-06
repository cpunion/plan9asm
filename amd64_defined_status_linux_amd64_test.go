//go:build linux && amd64

package plan9asm

import "testing"

// Unlike emulated-host observation agreement, these required native roots
// retain the full independent ISA expectations, including PTEST's cleared AF.
// Missing tools, instructions or actual native backing are failures.
func TestCrossLinuxRuntimeMatrixX86DefinedStatusOriginalGoNumerical(t *testing.T) {
	x87IA32RequireNativeHost(t)
	x86DefinedStatusOriginalGoNumerical(t)
}

func TestCrossLinuxRuntimeMatrixX86DefinedStatusIndependentOriginalCNumerical(t *testing.T) {
	x87IA32RequireNativeHost(t)
	x86DefinedStatusIndependentOriginalCNumerical(t)
}
