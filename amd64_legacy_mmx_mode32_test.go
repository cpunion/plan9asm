package plan9asm

import (
	"fmt"
	"testing"
)

func TestX86YMMMode32LegacyShiftActualGoRows(t *testing.T) {
	var names []string
	for _, rule := range x86YMMOriginalRules(t) {
		if rule.table == "yps" {
			names = append(names, rule.name)
		}
	}
	if len(names) != 8 {
		t.Fatal("complete original yps class changed", names)
	}

	for _, arch := range []string{"amd64", "386"} {
		for _, name := range names {
			for _, bank := range []string{"M", "X"} {
				for _, source := range []string{"$3", bank + "0", "8(AX)"} {
					row := fmt.Sprintf("%s %s,%s1", name, source, bank)
					t.Run(arch+"/"+row, func(t *testing.T) {
						text := "TEXT probe(SB),4,$0-0\n" + row + "\nRET\n"
						want := arch == "amd64" || bank == "X"
						goAssemblerOracleOutput(t, arch, text, "", want)

						form, known, err := x86LegacyMMXReadProperties(x86YMMReadSource(t, row), arch)
						if !known || (err == nil) != want {
							t.Fatal("named Go mode admission differs from pure selector", row, form, known, err)
						}
						if !want {
							file, err := Parse(ArchAMD64, text)
							if err != nil {
								t.Fatal(err)
							}
							if _, err := Translate(file, Options{
								Goarch: "386", TargetTriple: "i386-unknown-linux-gnu",
								Sigs: map[string]FuncSig{"probe": {Name: "probe", Ret: Void}},
							}); err == nil {
								t.Fatal("Go-rejected named mode32 MMX row produced LLVM IR", row)
							}
						}
					})
				}
			}
		}
	}
}
