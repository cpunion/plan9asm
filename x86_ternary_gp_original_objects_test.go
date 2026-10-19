package plan9asm

import (
	"fmt"
	"testing"
)

func TestX86TernaryGPOriginalRuntimeObjects(t *testing.T) {
	parallel, _, _, _ := x86ParallelBitsOriginalRuntimeFixture(t)
	for _, row := range []struct{ name, source, declaration string }{
		{"ANDN", x86ANDNOriginalRuntimeSource, "func andnsemantics(out *[40]byte)"},
		{"parallel", parallel.String(), x86ParallelBitsOriginalDeclarations()},
	} {
		for _, triple := range []string{"x86_64-apple-darwin", "x86_64-unknown-linux-gnu", "x86_64-pc-windows-msvc"} {
			t.Run(row.name+"/"+triple, func(t *testing.T) {
				x86RawNumericalFamilyObjects(t, row.source, row.declaration, "amd64", triple)
			})
		}
	}
}

func x86ParallelBitsOriginalDeclarations() string {
	var declarations string
	for index := 0; index < 24; index++ {
		declarations += fmt.Sprintf("func parallelbits%d(value uint64,mask *uint64,out *[16]byte)\n", index)
	}
	return declarations
}
