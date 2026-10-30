package plan9asm

import (
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/xgo-dev/plan9asm/internal/goasmrules"
)

// Original literal bytes and selected row indices are independent of the
// source formatter. AX/AX is a genuine opcode-90 NOP, including its L spelling.
func x86ExchangeOriginalCases(t *testing.T, catalog goasmrules.X86LegacyCatalog) []x86OfficialPackedCase {
	t.Helper()
	type sample struct {
		row             int
		operands, bytes string
	}
	fixtures := map[string][]sample{
		"XCHGB": {{0, "CL,BL", "86cb"}, {0, "CL,0(SI)", "860e"}, {1, "0(SI),BL", "861e"}},
		"XADDB": {{0, "CL,BL", "0fc0cb"}, {0, "CL,0(SI)", "0fc00e"}},
	}
	for _, width := range []struct{ suffix, prefix string }{{"W", "66"}, {"L", ""}, {"Q", "48"}} {
		p := width.prefix
		fixtures["XCHG"+width.suffix] = []sample{
			{0, "AX,BX", p + "93"}, {0, "AX,AX", p + "90"}, {1, "BX,AX", p + "93"},
			{2, "CX,BX", p + "87cb"}, {2, "CX,0(SI)", p + "870e"}, {3, "0(SI),BX", p + "871e"},
		}
		fixtures["XADD"+width.suffix] = []sample{{0, "CX,BX", p + "0fc1cb"}, {0, "CX,0(SI)", p + "0fc10e"}}
	}
	var cases []x86OfficialPackedCase
	seen := map[string]bool{}
	for _, entry := range catalog.Optab {
		samples, found := fixtures[entry.Opcode]
		if !found {
			continue
		}
		if seen[entry.Opcode] {
			t.Fatal("duplicate exchange source table", entry.Opcode)
		}
		seen[entry.Opcode] = true
		for _, fixture := range samples {
			code, err := hex.DecodeString(fixture.bytes)
			if err != nil || fixture.row >= len(entry.Rows) {
				t.Fatal("independent exchange fixture changed", err)
			}
			cases = append(cases, x86OfficialPackedCase{
				row: fmt.Sprintf("optab/%d/%s/%s/%d", entry.TableIndex, entry.Opcode, entry.Ytab, fixture.row),
				witness: goasmrules.X86PackedBinaryWitness{Instruction: entry.Opcode + " " + fixture.operands,
					Destination: strings.Split(fixture.operands, ",")[1], Encoding: code},
			})
		}
	}
	if len(seen) != 8 || len(cases) != 29 {
		t.Fatal("independent full exchange byte inventory changed", len(seen), len(cases))
	}
	return cases
}

func TestX86ExchangeOriginalCompleteRowsAndModes(t *testing.T) {
	catalog, err := goasmrules.LoadX86LegacyRules(runtime.GOROOT())
	if err != nil {
		t.Fatal(err)
	}
	all := x86ExchangeOriginalCases(t, catalog)
	for _, arch := range []string{"amd64", "386"} {
		t.Run(arch, func(t *testing.T) {
			var source strings.Builder
			var cases []x86OfficialPackedCase
			for _, original := range all {
				if arch == "386" && strings.HasSuffix(strings.Fields(original.witness.Instruction)[0], "Q") {
					continue
				}
				fixture := original
				fixture.symbol = fmt.Sprint("ExchangeOriginal", len(cases))
				fmt.Fprintf(&source, "TEXT ·%s(SB),4,$0-0\n", fixture.symbol)
				fixture.line = strings.Count(source.String(), "\n") + 1
				fmt.Fprintln(&source, fixture.witness.Instruction)
				source.WriteString("RET\n")
				cases = append(cases, fixture)
			}
			want := 29
			if arch == "386" {
				want = 21
			}
			if len(cases) != want || want > 32 {
				t.Fatal("complete exchange mode/byte union changed", len(cases), want)
			}
			listing, err := x86ScalarLiteralGoListing(t, arch, source.String())
			if err != nil {
				t.Fatal(err)
			}
			if err := x86OfficialPackedVerifyListing(listing, cases); err != nil {
				t.Fatal("independent original exchange byte/row ownership", err)
			}
		})
	}
	for _, stem := range []string{"XADD", "XCHG"} {
		x86ScalarLiteralExpectedGoListing(t, "386", "TEXT ·NoQ(SB),4,$0-0\n"+stem+"Q CX,BX\nRET\n", false)
	}
}

func x86ExchangeObjectDisassembly(t *testing.T, llc, objdump, triple, ir, level string) string {
	t.Helper()
	dir := t.TempDir()
	input, object := filepath.Join(dir, "exchange.ll"), filepath.Join(dir, "exchange.o")
	if err := os.WriteFile(input, []byte(ir), 0600); err != nil {
		t.Fatal(err)
	}
	x86FloatCompareCaptureCommand(t, exec.Command(llc, level, "-mtriple="+triple, "-filetype=obj", input, "-o", object))
	bytes, err := os.ReadFile(object)
	if err != nil || len(bytes) == 0 {
		t.Fatal("exchange target object missing", err)
	}
	listing := x86FloatCompareCaptureCommand(t, exec.Command(objdump, "-d", object))
	t.Logf("EXCHANGE_TARGET_OBJECT triple=%s level=%s bytes=%d\n%s", triple, level, len(bytes), listing)
	return string(listing)
}

func TestX86ExchangeMemoryPrefixOriginalAndLLVMObjects(t *testing.T) {
	llc, objdump := findLLVM22Tool("llc"), findLLVM22Tool("llvm-objdump")
	if llc == "" || objdump == "" {
		t.Fatal("LLVM22 llc and llvm-objdump are required")
	}
	for _, target := range x86RawReturnTargets {
		t.Run(target.triple, func(t *testing.T) {
			widths := "BWLQ"
			if target.arch == "386" {
				widths = "BWL"
			}
			var source, declarations strings.Builder
			for _, width := range widths {
				for _, locked := range []bool{false, true} {
					body := fmt.Sprintf("XADD%c CX,1(SI)", width)
					if width == 'B' {
						body = "XADDB CL,1(SI)"
					}
					if locked {
						body = "LOCK\n" + body
					}
					program, declaration := x86UnlockedScalarRMWFixture(target.arch, body, true)
					name := fmt.Sprintf("Prefix%c_%t", width, locked)
					source.WriteString(strings.Replace(program, "·Value", "·"+name, 1))
					declarations.WriteString(strings.Replace(declaration, "Value", name, 1) + "\n")
				}
			}
			// Original assembler output retains every explicit prefix, independently
			// of the translator's template and LLVM's machine encoder.
			listing, err := x86ScalarLiteralGoListing(t, target.arch, source.String())
			if err != nil {
				t.Fatal(err)
			}
			if strings.Count(listing, "\tLOCK") != len(widths) {
				t.Fatal("original Go source lost explicit LOCK adjacency", listing)
			}
			for _, api := range []string{"text", "module", "Go"} {
				t.Run(api, func(t *testing.T) {
					ir, err := x86ScalarFPBitViewAPI(t, source.String(), declarations.String(), target.arch, target.triple, api)
					if err != nil || ir == "" {
						t.Fatal("initialized complete bare/LOCK exchange source rejected", err)
					}
					for _, level := range []string{"-O0", "-O2"} {
						disassembly := x86ExchangeObjectDisassembly(t, llc, objdump, target.triple, ir, level)
						count, locked := 0, 0
						for _, line := range strings.Split(disassembly, "\n") {
							if !strings.Contains(line, "xadd") {
								continue
							}
							count++
							if strings.Contains(line, "lock") {
								locked++
							}
						}
						if count != 2*len(widths) || locked != len(widths) {
							t.Error("EXCHANGE_ORIGINAL_PREFIX_CHANGED", api, level, "xadd", count, "locked", locked, "want", 2*len(widths), len(widths))
						}
					}
				})
			}
		})
	}
}

func TestX86ExchangeHighByteOriginalPhysicalViews(t *testing.T) {
	for _, stem := range []struct{ name, opcode string }{{"XCHG", "86"}, {"XADD", "0fc0"}} {
		for _, sample := range []struct {
			operands, prefix, modrm string
			context                 bool
		}{
			{"AH,AL", "", "e0", false}, {"AL,AH", "", "c4", false}, {"AH,BL", "", "e3", false},
			{"AH,R8", "41", "e0", true}, {"R8,AH", "44", "c4", true}, {"AH,(R8)", "41", "20", true},
			{"AH,BPB", "40", "e5", true}, {"BPB,AH", "40", "ec", true}, {"AH,SIB", "40", "e6", true}, {"AH,DIB", "40", "e7", true},
			{"BL,BPB", "40", "dd", false}, {"BPB,BL", "40", "eb", false},
		} {
			operation := stem.name + "B " + sample.operands
			t.Run(operation, func(t *testing.T) {
				source := "TEXT ·ExchangeByte(SB),4,$0-0\n" + operation + "\nRET\n"
				code, err := hex.DecodeString(sample.prefix + stem.opcode + sample.modrm)
				if err != nil {
					t.Fatal(err)
				}
				fixture := x86OfficialPackedCase{symbol: "ExchangeByte", line: 2, row: operation,
					witness: goasmrules.X86PackedBinaryWitness{Instruction: operation, Encoding: code}}
				listing, err := x86ScalarLiteralGoListing(t, "amd64", source)
				if err != nil {
					t.Fatal(err)
				}
				if err := x86OfficialPackedVerifyListing(listing, []x86OfficialPackedCase{fixture}); err != nil {
					t.Fatal("actual physical byte-view oracle", err)
				}
				instruction := x86ScalarLaneInstruction(t, operation)
				spec := amd64ExchangeSpecs[instruction.Op]
				_, err = parseAMD64ExchangeForm("amd64", instruction.Op, spec, instruction)
				if sample.context && !errors.Is(err, ErrProbeNeedsContext) || !sample.context && err != nil {
					t.Error("EXCHANGE_HIGH_BYTE_PHYSICAL_VIEW_CHANGED", sample.context, err)
				}
			})
		}
	}
}
