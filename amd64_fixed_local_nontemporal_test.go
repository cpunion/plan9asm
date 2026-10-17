package plan9asm

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestGoAMD64FixedLocalNonTemporalCompleteFamilyProof(t *testing.T) {
	var original strings.Builder
	count := 0
	for _, row := range x86NonTemporalNativeOracle {
		banks := []string{"X"}
		if row.vector {
			banks = []string{"X", "Y", "Z"}
		}
		for _, bank := range banks {
			t.Run(row.name+"/"+bank, func(t *testing.T) {
				operation := row.name + " " + bank + "0,128(BP)\nSFENCE"
				initialize := "MOVOU 64(BP),X0\n"
				if row.vector {
					initialize = "VMOVUPS 64(BP)," + bank + "0\n"
				}
				if row.load {
					initialize = ""
					operation = row.name + " 64(BP)," + bank + "0\nMOVOU X0,128(BP)"
				}
				operation = initialize + operation
				pkg, _, file, options := fixedLocalPackedFixture(t, operation)
				original.WriteString(strings.Replace(fixedLocalPackedAssembly(operation), "·Value", fmt.Sprint("·Value", count), 1))
				count++
				if !GoAMD64FixedLocalFrameBindingNeeded(file, "amd64") {
					t.Error("native non-temporal access omitted its actual fixed-frame binding")
				}
				sig := options.Sigs[pkg.Path+".Value"]
				if _, err := proveX86FixedLocal(file.Funcs[0], sig); err != nil {
					t.Fatalf("complete initialized non-temporal frame: %v", err)
				}
				memory, misaligned := "128(BP)", "129(BP)"
				if row.load {
					memory, misaligned = "64(BP)", "65(BP)"
				}
				for _, displacement := range []string{misaligned, "320(BP)"} {
					changedOperation := strings.Replace(operation, memory, displacement, 1)
					changed := fixedLocalPackedAssembly(changedOperation)
					parsed, err := Parse(ArchAMD64, changed)
					if err != nil {
						t.Fatal(err)
					}
					if _, err := proveX86FixedLocal(parsed.Funcs[0], sig); !errors.Is(err, ErrProbeNeedsContext) {
						t.Fatal("misaligned or out-of-frame bytes acquired native backing", displacement, err)
					}
				}
			})
		}
	}
	if count != 17 {
		t.Fatal("independent fixed-frame non-temporal family is incomplete", count)
	}
	requireX86GoAssemblerResult(t, "amd64", original.String(), true)
}
