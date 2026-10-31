package plan9asm

import (
	"encoding/hex"
	"fmt"
	"runtime"
	"strings"
	"testing"

	"github.com/xgo-dev/plan9asm/internal/goasmrules"
)

// Independent literal fixtures retain physical directions and Go's first
// matching PEXTRW register row; the second row is witnessed by memory only.
func explicitSIMD27OriginalCases(t *testing.T, catalog goasmrules.X86LegacyCatalog) []x86OfficialPackedCase {
	t.Helper()
	type sample struct {
		name        string
		row         int
		args, codes []string
	}
	samples := []sample{
		{"EXTRACTPS", 0, []string{"$3,X0,AX", "$3,X0,(AX)"}, []string{"660f3a17c003", "660f3a170003"}},
		{"INSERTPS", 0, []string{"$7,X0,X1", "$7,(AX),X1"}, []string{"660f3a21c807", "660f3a210807"}},
		{"PEXTRW", 0, []string{"$7,X0,AX"}, []string{"660fc5c007"}},
		{"PEXTRW", 1, []string{"$7,X0,(AX)"}, []string{"660f3a150007"}},
		{"PEXTRB", 0, []string{"$7,X0,AX", "$7,X0,(AX)"}, []string{"660f3a14c007", "660f3a140007"}},
		{"PEXTRD", 0, []string{"$7,X0,AX", "$7,X0,(AX)"}, []string{"660f3a16c007", "660f3a160007"}},
		{"PEXTRQ", 0, []string{"$7,X0,AX", "$7,X0,(AX)"}, []string{"66480f3a16c007", "66480f3a160007"}},
		{"PINSRW", 0, []string{"$7,AX,X1", "$7,(AX),X1"}, []string{"660fc4c807", "660fc40807"}},
		{"PINSRB", 0, []string{"$7,AX,X1", "$7,(AX),X1"}, []string{"660f3a20c807", "660f3a200807"}},
		{"PINSRD", 0, []string{"$7,AX,X1", "$7,(AX),X1"}, []string{"660f3a22c807", "660f3a220807"}},
		{"PINSRQ", 0, []string{"$7,AX,X1", "$7,(AX),X1"}, []string{"66480f3a22c807", "66480f3a220807"}},
		{"CMPPD", 0, []string{"X0,X1,$7", "(AX),X1,$7"}, []string{"660fc2c807", "660fc20807"}},
		{"CMPPS", 0, []string{"X0,X1,$7", "(AX),X1,$7"}, []string{"0fc2c807", "0fc20807"}},
		{"CMPSD", 0, []string{"X0,X1,$7", "(AX),X1,$7"}, []string{"f20fc2c807", "f20fc20807"}},
		{"CMPSS", 0, []string{"X0,X1,$7", "(AX),X1,$7"}, []string{"f30fc2c807", "f30fc20807"}},
		{"BLENDVPS", 0, []string{"X0,X0,X1", "X0,(AX),X1"}, []string{"660f3814c8", "660f381408"}},
		{"BLENDVPD", 0, []string{"X0,X0,X1", "X0,(AX),X1"}, []string{"660f3815c8", "660f381508"}},
		{"PBLENDVB", 0, []string{"X0,X0,X1", "X0,(AX),X1"}, []string{"660f3810c8", "660f381008"}},
		{"SHA1RNDS4", 0, []string{"$3,X0,X1", "$3,(AX),X1"}, []string{"0f3accc803", "0f3acc0803"}},
		{"SHA256RNDS2", 0, []string{"X0,X0,X1", "X0,(AX),X1"}, []string{"0f38cbc8", "0f38cb08"}},
		{"MOVMSKPD", 0, []string{"X0,AX"}, []string{"660f50c0"}},
		{"MOVMSKPS", 0, []string{"X0,AX"}, []string{"0f50c0"}},
		{"PMOVMSKB", 0, []string{"X0,AX"}, []string{"660fd7c0"}},
		{"PMOVMSKB", 1, []string{"M0,AX"}, []string{"0fd7c0"}},
		{"MOVHLPS", 0, []string{"X0,X1"}, []string{"0f12c8"}},
		{"MOVLHPS", 0, []string{"X0,X1"}, []string{"0f16c8"}},
		{"PHMINPOSUW", 0, []string{"X0,X1", "(AX),X1"}, []string{"660f3841c8", "660f384108"}},
		{"PTEST", 0, []string{"X0,X1", "(AX),X1"}, []string{"660f3817c8", "660f381708"}},
	}
	entries := map[string]goasmrules.X86LegacyOpcode{}
	for _, entry := range catalog.Optab {
		for _, s := range samples {
			if entry.Opcode == s.name {
				entries[s.name] = entry
			}
		}
	}
	if len(entries) != 26 {
		t.Fatal("incomplete independent original family", len(entries))
	}
	var cases []x86OfficialPackedCase
	for _, sample := range samples {
		entry := entries[sample.name]
		if sample.row >= len(entry.Rows) {
			t.Fatal("missing original row", sample)
		}
		for i, args := range sample.args {
			code, err := hex.DecodeString(sample.codes[i])
			if err != nil {
				t.Fatal(err)
			}
			destination := args[strings.LastIndex(args, ",")+1:]
			if entry.Ytab == "yxcmpi" {
				destination = "X1"
			}
			cases = append(cases, x86OfficialPackedCase{row: fmt.Sprintf("optab/%d/%s/%s/%d", entry.TableIndex, entry.Opcode, entry.Ytab, sample.row), witness: goasmrules.X86PackedBinaryWitness{Instruction: sample.name + " " + args, Encoding: code, Destination: destination}})
		}
	}
	if len(cases) != 48 {
		t.Fatal("independent whole-family formats changed", len(cases))
	}
	return cases
}

func TestX86LegacyExplicitSIMDOriginalGoBytesAndModes(t *testing.T) {
	catalog, err := goasmrules.LoadX86LegacyRules(runtime.GOROOT())
	if err != nil {
		t.Fatal(err)
	}
	originals := explicitSIMD27OriginalCases(t, catalog)
	for _, arch := range []string{"amd64", "386"} {
		t.Run(arch, func(t *testing.T) {
			var accepted []x86OfficialPackedCase
			for _, original := range originals {
				name := strings.Fields(original.witness.Instruction)[0]
				if arch == "386" && (name == "PEXTRQ" || name == "PINSRQ") {
					t.Run("Q-refusal/"+original.witness.Instruction, func(t *testing.T) {
						listing, err := x86ScalarLiteralExpectedGoListing(t, arch, "TEXT ·NoQ(SB),4,$0-0\n"+original.witness.Instruction+"\nRET\n", false)
						t.Logf("ORIGINAL_EXPLICIT_SIMD_Q_REFUSAL source=%q exit=%v diagnostic=%q", original.witness.Instruction, err, listing)
					})
					continue
				}
				accepted = append(accepted, original)
			}
			for start := 0; start < len(accepted); start += 32 {
				end := start + 32
				if end > len(accepted) {
					end = len(accepted)
				}
				var source strings.Builder
				batch := append([]x86OfficialPackedCase(nil), accepted[start:end]...)
				for i := range batch {
					batch[i].symbol = fmt.Sprintf("ExplicitSIMD%d", start+i)
					fmt.Fprintf(&source, "TEXT ·%s(SB),4,$0-0\n", batch[i].symbol)
					batch[i].line = strings.Count(source.String(), "\n") + 1
					fmt.Fprintln(&source, batch[i].witness.Instruction)
					source.WriteString("RET\n")
				}
				listing, err := x86ScalarLiteralGoListing(t, arch, source.String())
				if err != nil {
					t.Fatal(err)
				}
				if err := x86OfficialPackedVerifyListing(listing, batch); err != nil {
					t.Fatal("independent original SIMD bytes", err)
				}
			}
			want := 48
			if arch == "386" {
				want = 44
			}
			if len(accepted) != want {
				t.Fatal("mode inventory changed", len(accepted))
			}
		})
	}
}
