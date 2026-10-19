package plan9asm

import (
	"testing"

	"github.com/xgo-dev/plan9asm/internal/goasmoracle"
)

func TestGoAssemblerOracleX86AliasMemoryDisplacementListing(t *testing.T) {
	// Complete alias directory, literal displacement rendering and both SIB
	// shapes. These fabricated observations are not instruction acceptance.
	for alias, canonical := range goAssemblerOracleX86Aliases {
		t.Run(alias, func(t *testing.T) {
			for _, displacement := range []struct{ original, printed string }{
				{"0", ""}, {"+0", ""}, {"-0", ""}, {"0x0", ""},
				{"0x10", "16"}, {"-0x10", "-16"}, {"017", "15"},
				{"0xffffffffffffffff", "-1"}, {"0b101", "5"}, {"1_024", "1024"},
			} {
				for _, address := range []string{"(AX)", "(AX)(CX*2)"} {
					source := x86DiagnosticFixture(t, alias+" "+displacement.original+address+",X1")
					instruction := canonical + " " + displacement.printed + address + ",X1"
					for _, arch := range []string{"386", "amd64"} {
						output := x86DiagnosticListing(source, instruction)
						if !goAssemblerOracleX86SourceDiagnostic(source, arch, output) {
							t.Fatal("complete alias and displacement rendering lost its exact source", arch, displacement, address)
						}
						if goasmoracle.SourceDiagnosticLine(source, output, 2) {
							t.Fatal("target-free parser acquired an x86 alias/displacement role")
						}
						for _, changed := range []string{
							canonical + " 7" + address + ",X1",
							canonical + " " + displacement.printed + "(DX),X1",
							canonical + " " + displacement.printed + address + ",X2",
							canonical + " " + displacement.printed + address + ",X1,X2",
						} {
							if goAssemblerOracleX86SourceDiagnostic(source, arch, x86DiagnosticListing(source, changed)) {
								t.Fatal("changed address, value or cardinality borrowed the displacement role", changed)
							}
						}
					}
					if !goAssemblerOracleX86SourceDiagnostic(source, "386", x86Mode32Diagnostic(source, instruction, "(24 21)")) ||
						goAssemblerOracleX86SourceDiagnostic(source, "arm64", x86DiagnosticListing(source, instruction)) {
						t.Fatal("displacement rendering lost its explicit mode/target boundary")
					}
				}
			}
		})
	}
}
