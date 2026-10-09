package plan9asm

import "testing"

// This required owner is separate from Rosetta SSE and object-only evidence.
// CPUID/OS-state and original-opcode preflight must execute successfully.
func TestCrossLinuxRuntimeMatrixX86PackedConversionNativeAVX512Status(t *testing.T) {
	x86PackedConversionRuntimeMatrix(t, true)
}

func TestCrossLinuxRuntimeMatrixX86PackedConversionNativeIndependentC(t *testing.T) {
	amd64PackedNumericConversionNativeRuntimeSemantics(t)
}

func TestCrossLinuxRuntimeMatrixX86PackedConversionNativeDeadFaults(t *testing.T) {
	x86PackedConversionDeadFaults(t, true, true)
}
