package plan9asm

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/xgo-dev/plan9asm/internal/goasmrules"
)

func x86Legacy127Targets(arch string) []string {
	if arch == "386" {
		return []string{"i386-unknown-linux-gnu", "i686-pc-windows-msvc"}
	}
	return []string{"x86_64-unknown-linux-gnu", "x86_64-apple-darwin", "x86_64-pc-windows-msvc"}
}

// Inputs are explicit pointer/scalar parameters. Seeding both X inputs is
// deliberate initialized context, not a fact inferred from source syntax.
// MMX has a separate real region, output observation and mandatory EMMS.
func x86LegacyMediaInitializedFunction(name, operation, arch string) (string, string) {
	word, move := goWordSize(arch), "Q"
	if word == 4 {
		move = "L"
	}
	source := fmt.Sprintf("TEXT ·%s(SB),4,$0-%d\nMOV%s out+0(FP),DX\nMOV%s a+%d(FP),AX\nMOV%s b+%d(FP),BX\nMOV%s mask+%d(FP),CX\n"+
		"MOVOU (AX),X0\nMOVOU (BX),X1\nCMPL CX,CX\n", name, 3*word+8, move, move, word, move, 2*word, move, 3*word)
	mmx := strings.Contains(operation, "M1")
	if mmx {
		source += "MOVQ (AX),M0\nMOVQ (BX),M1\n"
	}
	source += operation + "\n"
	if mmx {
		source += "MOVQ M1,32(DX)\nEMMS\n"
	}
	// Observe the whole X carrier and memory, including scalar/half merge
	// bytes. COMI observes its defined status rather than a fictitious X write.
	source += "MOVOU X1,0(DX)\nMOVOU (AX),X2\nMOVOU X2,16(DX)\n" +
		"SETCS CL\nMOVB CL,40(DX)\nSETEQ CL\nMOVB CL,41(DX)\nSETPS CL\nMOVB CL,42(DX)\nRET\n"
	return source, fmt.Sprintf("func %s(out *[48]byte,a,b *[16]byte,mask uint64)\n", name)
}

// This batches complete functions, not a compiler process per source row.
// Whole-source Go byte identity precedes each three-API O0/O2 object batch.
func x86Legacy127Objects(t *testing.T, arch string, cases []x86OfficialPackedCase,
	function func(string, string, string) (string, string)) {
	t.Helper()
	llc := findLLVM22Tool("llc")
	if llc == "" {
		t.Fatal("LLVM22 llc required")
	}
	for start := 0; start < len(cases); start += 32 {
		end := start + 32
		if end > len(cases) {
			end = len(cases)
		}
		t.Run(fmt.Sprintf("%s/batch%d", arch, start/32), func(t *testing.T) {
			batch := cases[start:end]
			var source, declarations strings.Builder
			for index := range batch {
				fixture := &batch[index]
				fixture.symbol = fmt.Sprint("Legacy127Row", index)
				body, declaration := function(fixture.symbol, fixture.witness.Instruction, arch)
				position := strings.Index(body, "\n"+fixture.witness.Instruction+"\n")
				if position < 0 || strings.Count(body, "\n"+fixture.witness.Instruction+"\n") != 1 {
					t.Fatal("initialized producer lost the exact original instruction", fixture.row)
				}
				fixture.line = strings.Count(source.String(), "\n") + strings.Count(body[:position+1], "\n") + 1
				source.WriteString(body)
				declarations.WriteString(declaration)
			}
			listing, err := x86ScalarLiteralGoListing(t, arch, source.String())
			if err != nil {
				t.Fatal(err)
			}
			if err := x86OfficialPackedVerifyListing(listing, batch); err != nil {
				t.Fatal("actual original source row/prefix/direction/byte identity", err)
			}
			for _, target := range x86Legacy127Targets(arch) {
				t.Run(target, func(t *testing.T) {
					for _, api := range []string{"text", "module", "Go"} {
						t.Run(api, func(t *testing.T) {
							ir, err := x86ScalarFPBitViewAPI(t, source.String(), declarations.String(), arch, target, api)
							if err != nil || ir == "" {
								t.Fatal("initialized whole-function translation", err)
							}
							for _, level := range []string{"-O0", "-O2"} {
								x86PackedStringCompileObject(t, llc, target, ir, level)
							}
						})
					}
				})
			}
		})
	}
}

func x86Legacy127SaveIdentity(t *testing.T, variable string, value interface{}) {
	t.Helper()
	if path := os.Getenv(variable); path != "" {
		data, err := json.MarshalIndent(value, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, append(data, '\n'), 0600); err != nil {
			t.Fatal(err)
		}
	}
}

func TestX86OfficialLegacyMediaRemainderRowsObjects(t *testing.T) {
	catalog, err := goasmrules.LoadX86LegacyRules(runtime.GOROOT())
	if err != nil {
		t.Fatal(err)
	}
	modes := map[string][]goasmrules.X86LegacyMediaRow{}
	for _, arch := range []string{"amd64", "386"} {
		rows, err := goasmrules.X86LegacyMediaRemainderRows(catalog, arch)
		if err != nil {
			t.Fatal(err)
		}
		modes[arch] = rows
		var cases []x86OfficialPackedCase
		for _, row := range rows {
			for _, witness := range row.Witnesses {
				cases = append(cases, x86OfficialPackedCase{row: row.ID, witness: witness})
			}
		}
		want := 160
		if arch == "386" {
			want = 70
		}
		if len(rows) != 96 || len(cases) != want {
			t.Fatal("complete source/mode witness union changed", arch, len(rows), len(cases))
		}
		x86Legacy127Objects(t, arch, cases, x86LegacyMediaInitializedFunction)
	}
	if !t.Failed() {
		x86Legacy127SaveIdentity(t, "PLAN9ASM_LEGACY_MEDIA_IDENTITY_OUTPUT", struct {
			Source goasmrules.X86LegacyCatalog
			Modes  map[string][]goasmrules.X86LegacyMediaRow
		}{catalog, modes})
	}
}

func TestX86LegacyMediaRemainderInitializedInputsRemainRequired(t *testing.T) {
	for _, item := range []struct{ operation, removed string }{
		{"PADDB M0,M1", "MOVQ (BX),M1\n"},
		{"PADDB M0,M1", "MOVQ (AX),M0\n"},
		{"PADDB M0,M1", "EMMS\n"},
		{"CVTSD2SS X0,X1", "MOVOU (BX),X1\n"},
		{"COMISD X0,X1", "MOVOU (AX),X0\n"},
	} {
		t.Run(item.operation+"/"+strings.TrimSpace(item.removed), func(t *testing.T) {
			source, declaration := x86LegacyMediaInitializedFunction("MediaMissing", item.operation, "amd64")
			if strings.Count(source, item.removed) != 1 {
				t.Fatal("negative did not remove exactly one initialized input/state boundary")
			}
			source = strings.Replace(source, item.removed, "", 1)
			for _, api := range []string{"text", "module", "Go"} {
				if ir, err := x86ScalarFPBitViewAPI(t, source, declaration, "amd64", "x86_64-unknown-linux-gnu", api); ir != "" || !errors.Is(err, ErrProbeNeedsContext) {
					t.Fatal("source transport supplied missing NUM/old bytes or MMX state", api, err)
				}
			}
		})
	}
}

func TestX86LegacyMediaRemainderVoidAndForeignNativeRemainRejected(t *testing.T) {
	for _, operation := range []string{"PADDB M0,M1", "COMISD X0,X1", "CVTSD2SS X0,X1"} {
		for _, api := range []string{"text", "module", "Go"} {
			source := "TEXT ·MediaVoid(SB),4,$0-0\n" + operation + "\nRET\n"
			if ir, err := x86ScalarFPBitViewAPI(t, source, "func MediaVoid()", "amd64", "x86_64-unknown-linux-gnu", api); ir != "" || !errors.Is(err, ErrProbeNeedsContext) {
				t.Fatal("dead Void result hid uninitialized media inputs", operation, api, err)
			}
		}
	}
	source, declaration := x86LegacyMediaInitializedFunction("MediaForeign", "PADDB M0,M1", "amd64")
	for _, api := range []string{"text", "module", "Go"} {
		if ir, err := x86ScalarFPBitViewAPI(t, source, declaration, "amd64", "aarch64-unknown-linux-gnu", api); err == nil || ir != "" {
			t.Fatal("source MMX transport supplied foreign native-target permission", api, err)
		}
	}
}
