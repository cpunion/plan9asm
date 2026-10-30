package plan9asm

import (
	"encoding/hex"
	"fmt"
	"runtime"
	"strings"
	"testing"

	"github.com/xgo-dev/plan9asm/internal/goasmrules"
)

type x86OrdinaryMoveOriginalCase struct {
	x86OfficialPackedCase
	original string
	mode64   bool
	rewrite  string
}

// Literal original-Go encodings are independent of the production formatter.
// In particular the MMX/XMM carriers are not GP numeric permissions, and an
// address rewrite is explicitly observed as LEA rather than a selected MOV row.
func x86OrdinaryMoveOriginalCases(t *testing.T, catalog goasmrules.X86LegacyCatalog) []x86OrdinaryMoveOriginalCase {
	t.Helper()
	type sample struct {
		row      int
		operands string
		code     string
	}
	fixtures := map[string][]sample{
		"MOVB": {{0, "CL,BL", "88cb"}, {0, "CL,0(SI)", "880e"}, {1, "0(SI),BL", "8a1e"}, {2, "$128,BL", "b380"}, {3, "$128,0(SI)", "c60680"}},
		"MOVW": {{0, "CX,BX", "6689cb"}, {0, "CX,0(SI)", "66890e"}, {1, "0(SI),BX", "668b1e"}, {2, "$128,BX", "66bb8000"}, {3, "$128,0(SI)", "66c7068000"}, {4, "$slot+0(SP),BX", "668d1c24"}},
		"MOVL": {{0, "CX,BX", "89cb"}, {0, "CX,0(SI)", "890e"}, {1, "0(SI),BX", "8b1e"}, {2, "$128,BX", "bb80000000"}, {3, "$128,0(SI)", "c70680000000"},
			{4, "CX,M1", "0f6ec9"}, {4, "0(SI),M1", "0f6e0e"}, {5, "M1,CX", "0f7ec9"}, {5, "M1,0(SI)", "0f7e0e"},
			{6, "CX,X1", "660f6ec9"}, {6, "0(SI),X1", "660f6e0e"}, {7, "X1,CX", "660f7ec9"}, {7, "X1,0(SI)", "660f7e0e"}, {8, "$slot+0(SP),BX", "8d1c24"}},
		"MOVQ": {{0, "0(SI),M1", "0f6f0e"}, {1, "M1,0(SI)", "0f7f0e"}, {2, "X0,M1", "f20fd6c8"},
			{3, "X0,X1", "f30f7ec8"}, {3, "0(SI),X1", "f30f7e0e"}, {4, "X1,0(SI)", "660fd60e"},
			{5, "CX,BX", "4889cb"}, {5, "CX,0(SI)", "48890e"}, {6, "0(SI),BX", "488b1e"},
			{7, "$0x1122334455667788,BX", "48bb8877665544332211"}, {8, "$128,0(SI)", "48c70680000000"},
			{9, "CX,M1", "480f6ec9"}, {10, "M1,CX", "480f7ec9"}, {11, "CX,X1", "66480f6ec9"}, {12, "X1,CX", "66480f7ec9"}, {13, "$slot+0(SP),BX", "488d1c24"}},
	}
	var result []x86OrdinaryMoveOriginalCase
	seen := map[string]bool{}
	for _, entry := range catalog.Optab {
		samples, found := fixtures[entry.Opcode]
		if !found {
			continue
		}
		if seen[entry.Opcode] {
			t.Fatal("duplicate original MOV table", entry.Opcode)
		}
		seen[entry.Opcode] = true
		for _, fixture := range samples {
			index := fixture.row
			if entry.Opcode == "MOVQ" && len(entry.Rows) == 15 && index >= 7 {
				index++
			}
			code, err := hex.DecodeString(fixture.code)
			if err != nil || index >= len(entry.Rows) {
				t.Fatal("independent MOV fixture changed", err)
			}
			original := entry.Opcode + " " + fixture.operands
			selected, rewrite := original, ""
			if entry.Rows[index].ZCase == "Zaut_r" && entry.Opcode != "MOVW" {
				selected = "LEA" + entry.Opcode[3:] + " " + fixture.operands
				rewrite = "original Go progedit rewrites TYPE_ADDR before ymov matching"
			}
			result = append(result, x86OrdinaryMoveOriginalCase{
				x86OfficialPackedCase: x86OfficialPackedCase{row: fmt.Sprintf("optab/%d/%s/%s/%d", entry.TableIndex, entry.Opcode, entry.Ytab, index),
					witness: goasmrules.X86PackedBinaryWitness{Instruction: selected, Destination: strings.Split(fixture.operands, ",")[1], Encoding: code}},
				original: original, mode64: entry.Prefix == "Pw8" && entry.Rows[index].UnmatchedOffset >= 8, rewrite: rewrite,
			})
		}
		if entry.Opcode == "MOVQ" && len(entry.Rows) == 15 {
			result = append(result, x86OrdinaryMoveOriginalCase{x86OfficialPackedCase: x86OfficialPackedCase{
				row: fmt.Sprintf("optab/%d/MOVQ/ymovq/7", entry.TableIndex), witness: goasmrules.X86PackedBinaryWitness{
					Instruction: "MOVQ $-1,BX", Destination: "BX", Encoding: []byte{0x48, 0xc7, 0xc3, 0xff, 0xff, 0xff, 0xff}}},
				original: "MOVQ $-1,BX", mode64: true})
		}
	}
	if len(seen) != 4 {
		t.Fatal("missing complete ordinary MOV source table", seen)
	}
	return result
}

func x86OrdinaryMoveCheckOriginalGo(t *testing.T, arch string, cases []x86OrdinaryMoveOriginalCase) {
	t.Helper()
	for start := 0; start < len(cases); start += 32 {
		end := start + 32
		if end > len(cases) {
			end = len(cases)
		}
		t.Run(fmt.Sprintf("%s/batch%d", arch, start/32), func(t *testing.T) {
			var source strings.Builder
			var originals []x86OfficialPackedCase
			for index, fixture := range cases[start:end] {
				fixture.symbol = fmt.Sprintf("OrdinaryMove%d", start+index)
				fixture.line = strings.Count(source.String(), "\n") + 2
				fmt.Fprintf(&source, "TEXT ·%s(SB),4,$0-0\n%s\nRET\n", fixture.symbol, fixture.original)
				originals = append(originals, fixture.x86OfficialPackedCase)
			}
			listing, err := x86ScalarLiteralGoListing(t, arch, source.String())
			if err != nil {
				t.Fatal("original ordinary MOV source fixture failed", err)
			}
			if err := x86OfficialPackedVerifyListing(listing, originals); err != nil {
				t.Fatal("original MOV row/direction/first-match bytes differ", err)
			}
		})
	}
}

func TestX86OfficialOrdinaryMoveCompleteOriginalGo(t *testing.T) {
	catalog, err := goasmrules.LoadX86LegacyRules(runtime.GOROOT())
	if err != nil {
		t.Fatal(err)
	}
	all := x86OrdinaryMoveOriginalCases(t, catalog)
	for _, arch := range []string{"amd64", "386"} {
		var accepted []x86OrdinaryMoveOriginalCase
		seen := map[string]bool{}
		for _, fixture := range all {
			if arch == "386" && fixture.mode64 {
				if !seen[fixture.row] {
					t.Run("386/refusal/"+fixture.row, func(t *testing.T) {
						goAssemblerOracleOutput(t, arch, "TEXT f(SB),4,$0-0\n"+fixture.original+"\nRET\n", "", false)
					})
				}
				seen[fixture.row] = true
				continue
			}
			accepted = append(accepted, fixture)
		}
		x86OrdinaryMoveCheckOriginalGo(t, arch, accepted)
	}
}

func TestX86OfficialOrdinaryMoveImmediateFirstMatch(t *testing.T) {
	for _, arch := range []string{"amd64", "386"} {
		var cases []x86OrdinaryMoveOriginalCase
		for _, row := range []struct{ instruction, code string }{
			{"MOVB $-1,BL", "b3ff"}, {"MOVW $65535,BX", "66bbffff"}, {"MOVL $-1,BX", "bbffffffff"},
		} {
			code, _ := hex.DecodeString(row.code)
			cases = append(cases, x86OrdinaryMoveOriginalCase{x86OfficialPackedCase: x86OfficialPackedCase{row: "independent immediate selector", witness: goasmrules.X86PackedBinaryWitness{Instruction: row.instruction, Encoding: code}}, original: row.instruction})
		}
		if arch == "amd64" {
			for _, row := range []struct{ instruction, code string }{
				{"MOVQ $0,BX", "bb00000000"}, {"MOVQ $2147483648,BX", "bb00000080"}, {"MOVQ $4294967295,BX", "bbffffffff"},
				{"MOVQ $-1,BX", "48c7c3ffffffff"}, {"MOVQ $4294967296,BX", "48bb0000000001000000"},
			} {
				code, _ := hex.DecodeString(row.code)
				cases = append(cases, x86OrdinaryMoveOriginalCase{x86OfficialPackedCase: x86OfficialPackedCase{row: "independent MOVQ shortening", witness: goasmrules.X86PackedBinaryWitness{Instruction: row.instruction, Encoding: code}}, original: row.instruction})
			}
		}
		x86OrdinaryMoveCheckOriginalGo(t, arch, cases)
	}
}
