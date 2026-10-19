package plan9asm

import "testing"

func TestX86BMI2VariableShiftOriginalSourceGoObjects(t *testing.T) {
	source := x86X87FunctionLiteral(t, "amd64_lower_bmi2_shift_test.go",
		"TestAMD64BMI2VariableShiftRuntimeSemantics", "source")
	requireX86GoAssemblerResult(t, "amd64", source, true)
}
