//go:build go1.27

package plan9asm

import "testing"

func TestCrossLinuxRuntimeMatrixGoAMD64FixedLocalTerminalReturns(t *testing.T) {
	assembly, declarations := fixedLocalTerminalReturnProgram()
	inputs := []uint64{0, 1, 2, ^uint64(0)}
	var expected [][]string
	for range inputs {
		expected = append(expected, []string{"0000000000000007", "0000000000000007"})
	}
	fixedLocalResultsRuntime(t, assembly, declarations, inputs, expected)
	t.Log("native terminal inventory: functions=2 inputs=4 observations_per_program=8 APIs=3 optimizations=2 execute=true")
}
