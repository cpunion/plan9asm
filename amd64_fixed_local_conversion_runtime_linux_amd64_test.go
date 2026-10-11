package plan9asm

import "testing"

func TestCrossLinuxRuntimeMatrixGoAMD64FixedLocalConversionNative(t *testing.T) {
	fixedLocalConversionRuntime(t, x86PackedDoubleDwordNativeTools(t), true)
}
