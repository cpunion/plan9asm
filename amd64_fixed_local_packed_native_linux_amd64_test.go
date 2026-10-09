package plan9asm

import "testing"

func TestCrossLinuxRuntimeMatrixGoAMD64FixedLocalPackedNativeWide(t *testing.T) {
	execution := x86PackedDoubleDwordNativeTools(t)
	fixedLocalPackedRuntime(t, execution, true, true)
}
