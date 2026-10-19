package plan9asm

import (
	"strings"
	"testing"
)

// Preserve complete original Go objects and their missing-input public Context
// before compiling a private format unit. This proves formats, never initialized
// numerical inputs, native entry, memory availability or executed semantics.
func x86UnboundVoidVectorMoveFormatObject(t *testing.T, source, arch, triple string) string {
	t.Helper()
	file, signatures := x86UnboundVoidFormatSources(t, source, arch, triple)
	normalized, err := normalizeX86RawFile(file, arch)
	if err != nil {
		t.Fatal("complete original move format lost its decoded bytes", err)
	}
	return x86InstructionFormatObjectUnit(t, normalized, signatures, arch, triple, 0,
		func(c *amd64Ctx, instruction Instr) (bool, bool, error) {
			if known, terminated, err := c.lowerPackedVectorMove(instruction.Op, instruction); known {
				return known, terminated, err
			}
			return c.lowerVectorScalarMove(instruction.Op, instruction)
		})
}

func x86UnboundVoidRawVectorMoveProbeObject(t *testing.T, file *File, arch, triple string) string {
	t.Helper()
	var source strings.Builder
	for _, function := range file.Funcs {
		text, valid := x86RawReplaySourceText(function)
		if !valid {
			t.Fatal("move format probe lost its complete source")
		}
		source.WriteString(text)
	}
	return x86UnboundVoidVectorMoveFormatObject(t, source.String(), arch, triple)
}
