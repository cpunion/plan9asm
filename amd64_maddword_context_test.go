package plan9asm

import (
	"errors"
	"strings"
	"testing"
)

func TestX86WordMultiplyAddLegacyVirtualMemoryCannotProveAlignment(t *testing.T) {
	for _, operand := range []string{"source+0(FP)", "(SP)", "(BP)"} {
		source := "TEXT f(SB),4,$0-64\nPMADDWL " + operand + ",X1\nRET\n"
		requireX86GoAssemblerResult(t, "amd64", source, true)
		file, err := Parse(ArchAMD64, source)
		if err != nil {
			t.Fatal(err)
		}
		for _, triple := range []string{"x86_64-apple-darwin", "x86_64-unknown-linux-gnu", "x86_64-pc-windows-msvc", "i386-unknown-linux-gnu", "i686-pc-windows-msvc"} {
			arch := "amd64"
			if strings.HasPrefix(triple, "i") {
				arch = "386"
			}
			_, err := Translate(file, Options{Goarch: arch, TargetTriple: triple, Sigs: map[string]FuncSig{"f": x86WordMaddFPType(t)}})
			if !errors.Is(err, ErrProbeNeedsContext) {
				t.Fatalf("%s virtual%s obtained original m128 alignment/fault semantics: %v", triple, operand, err)
			}
		}
	}
}
