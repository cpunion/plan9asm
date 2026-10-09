package plan9asm

import (
	"fmt"
	"reflect"
	"strconv"
	"testing"
)

func TestSplitSemicolonsRetainsWholeGoQuotedTokens(t *testing.T) {
	for _, row := range []struct {
		source string
		want   []string
	}{
		{" ; RET; ;", []string{"", "RET", "", ""}},
		{`DATA value(SB)/4,$"a;b";RET`, []string{`DATA value(SB)/4,$"a;b"`, "RET"}},
		{`BYTE $';';RET`, []string{`BYTE $';'`, "RET"}},
		{`DATA value(SB)/8,$"a\";b";RET`, []string{`DATA value(SB)/8,$"a\";b"`, "RET"}},
		{`DATA value(SB)/8,$"a\\;b";RET`, []string{`DATA value(SB)/8,$"a\\;b"`, "RET"}},
		{`DATA value(SB)/8,$"unterminated;`, []string{`DATA value(SB)/8,$"unterminated;`}},
	} {
		t.Run(row.source, func(t *testing.T) {
			if got := splitSemicolons(row.source); !reflect.DeepEqual(got, row.want) {
				t.Fatalf("quoted token was split: got=%q want=%q", got, row.want)
			}
		})
	}
}

func TestParseSemicolonStringDataMatchesActualGoFiveArch(t *testing.T) {
	for _, target := range []struct {
		name string
		arch Arch
	}{
		{"386", ArchAMD64}, {"amd64", ArchAMD64}, {"arm", ArchARM},
		{"arm64", ArchARM64}, {"wasm", ArchWASM},
	} {
		t.Run(target.name, func(t *testing.T) {
			for _, payload := range []string{";", "a;b", "a\";b", "a\\;b", "//;/*;", "x;\x00;y"} {
				t.Run(strconv.Quote(payload), func(t *testing.T) {
					source := fmt.Sprintf("DATA text<>+0(SB)/8,$%s;GLOBL text<>(SB),8,$8\nTEXT Probe(SB),4,$0-0\nRET\n", strconv.Quote(payload))
					goAssemblerOracleOutput(t, target.name, source, "", true)
					file, err := Parse(target.arch, source)
					if err != nil || len(file.Data) != 1 || len(file.Globl) != 1 || len(file.Funcs) != 1 {
						t.Fatalf("actual Go-accepted quoted DATA lost statement boundaries: %v", err)
					}
					if string(file.Data[0].Payload) != payload || file.Data[0].Width != 8 {
						t.Fatalf("quoted DATA bytes changed: %+v", file.Data[0])
					}
				})
			}
		})
	}
}
