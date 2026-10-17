package plan9asm

import (
	"errors"
	"testing"
)

// Original zero-argument format fixtures have no initialized entry registers
// or bound native pointers. Retain every Go object, public Context boundary
// and emitter-only LLVM format; initialized counterparts own public execution.
func x86ScalarCountUnboundFormatObject(t *testing.T, source, name, arch, triple string) string {
	t.Helper()
	x86LEAActualTargetGoObject(t, arch, triple, source)
	for _, api := range []string{"text", "module", "Go"} {
		ir, err := x86ScalarFPBitViewAPI(t, source, "func "+name+"()", arch, triple, api)
		if ir != "" || !errors.Is(err, ErrProbeNeedsContext) {
			t.Fatal("original format fixture invented a native input", api, err)
		}
	}
	file, err := Parse(ArchAMD64, source)
	if err != nil {
		t.Fatal(err)
	}
	return x86InstructionFormatObjectUnit(t, file,
		map[string]FuncSig{name: {Name: name, Ret: Void}}, arch, triple, 0,
		func(c *amd64Ctx, instruction Instr) (bool, bool, error) {
			return c.lowerScalarBitCount(instruction.Op, instruction)
		})
}
