//go:build go1.27 && darwin && (amd64 || arm64)

package plan9asm

import "testing"

func TestGoX86NumericalSBReadMutableHostDiagnostic(t *testing.T) {
	execution := x86FloatCompareExecutionTools(t)
	t.Log("retargeted host numerical/volatile/canary/fault diagnostic, not Linux native Go ABI execution")
	numericalRuntimeExecute(t, "amd64", execution, false)
}
