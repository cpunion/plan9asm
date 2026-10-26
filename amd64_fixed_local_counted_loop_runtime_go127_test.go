//go:build unix && go1.27

package plan9asm

import "testing"

func TestCrossLinuxRuntimeMatrixGoAMD64FixedLocalCountedLoop(t *testing.T) {
	assembly, declarations, expected := fixedLocalCountedLoopProgram(t)
	if len(expected) != 38 {
		t.Fatal("counted-loop runtime inventory changed", len(expected))
	}
	fixedLocalConstantResultsRuntime(t, assembly, declarations, expected)
}
