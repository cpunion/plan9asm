package plan9asm

import "testing"

func TestGoAssemblerOracleX86BoundPackageQualification(t *testing.T) {
	for _, alias := range []string{"MOVD", "MOVDQ2Q"} {
		for _, suffix := range []string{"", "+3", "-1"} {
			source := x86DiagnosticFixture(t, alias+" $·value"+suffix+"(SB),AX")
			listing := "MOVQ $example.com/forms.value" + suffix + "(SB),AX"
			output := x86DiagnosticListing(source, listing)
			if !goAssemblerOracleX86PackageDiagnostic(source, "386", output, "example.com/forms") {
				t.Fatal("exact source/command package qualification lost")
			}
			for _, changed := range []string{
				"MOVQ $other.example/forms.value" + suffix + "(SB),AX",
				"MOVQ $example.com/forms.other" + suffix + "(SB),AX",
				"MOVQ $example.com/forms.value+9(SB),AX",
				"MOVQ $example.com/forms.value" + suffix + "(SB),BX",
				"MOVQ $example.com/forms.value" + suffix + "(SB)",
				"MOVL $example.com/forms.value" + suffix + "(SB),AX",
			} {
				if goAssemblerOracleX86PackageDiagnostic(source, "386", x86DiagnosticListing(source, changed), "example.com/forms") {
					t.Fatal("changed complete listing acquired package role", changed)
				}
			}
			if goAssemblerOracleX86PackageDiagnostic(source, "386", output, "") ||
				goAssemblerOracleX86PackageDiagnostic(source, "arm64", output, "example.com/forms") {
				t.Fatal("missing command package or foreign architecture borrowed qualification")
			}
		}
	}
}
