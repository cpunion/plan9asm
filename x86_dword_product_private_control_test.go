package plan9asm

import (
	"fmt"
	"strings"
	"testing"
)

func TestX86DwordProductRawPrivateFlagsFiveTargetObjects(t *testing.T) {
	for _, target := range x86RawReturnTargets {
		t.Run(target.triple, func(t *testing.T) {
			word, suffix, move := goWordSize(target.arch), "Q", "MOVQ"
			if word == 4 {
				suffix, move = "L", "MOVL"
			}
			var source, declarations strings.Builder
			count := 0
			for _, row := range x86FixedBinaryValueRows(target.arch,
				[]string{"VPMULDQ", "VPMULUDQ", "VPMULDQ.BCST", "VPMULUDQ.BCST"}) {
				body, declaration := x86FixedBinaryValueFixture(target.arch, row.op, row.bank, row.mode, false)
				instruction := ""
				for _, line := range strings.Split(body, "\n") {
					if strings.HasPrefix(line, "VPMULDQ") || strings.HasPrefix(line, "VPMULUDQ") {
						instruction = line
					}
				}
				if instruction == "" {
					t.Fatal("complete initialized product source lost its operation")
				}
				raw := x86RawNumericalInstruction(t, target.arch, instruction)
				body = strings.Replace(body, fmt.Sprintf("$0-%d", word*2), fmt.Sprintf("$0-%d", word*3), 1)
				body = strings.Replace(body, instruction, fmt.Sprintf("PUSH%s mask+%d(FP)\nPOPF%s\n%s"+
					"PUSHF%s\nPOP%s CX\n%s CX,ret+%d(FP)\n", suffix, word, suffix, raw,
					suffix, suffix, move, word*2), 1)
				name := fmt.Sprint("ProductFlags", count)
				source.WriteString(strings.ReplaceAll(body, "Value", name))
				declarations.WriteString(strings.ReplaceAll(declaration, "Value", name) + " uintptr\n")
				count++
			}
			want := 36
			if word == 4 {
				want = 12
			}
			if count != want {
				t.Fatal("complete raw product full/broadcast/mask private-control inventory changed", count)
			}
			x86RawNumericalFamilyObjects(t, source.String(), declarations.String(), target.arch, target.triple)
		})
	}
}
