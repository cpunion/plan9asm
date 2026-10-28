package plan9asm

// Original source rows, selected Go bytes and initialized API objects are
// distinct observations. These owners do not establish runtime semantics.
import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"

	"github.com/xgo-dev/plan9asm/internal/goasmrules"
)

var legacyGPRMWDraftNames = regexp.MustCompile(`^(ADC|ADD|AND|OR|SBB|SUB|XOR|INC|DEC|NEG|NOT|BTC|BTR|BTS)[BWLQ]$`)

func TestX86OfficialScalarRMWCompleteSourceCarrier(t *testing.T) {
	source, err := os.ReadFile(filepath.Join(runtime.GOROOT(), "src/cmd/internal/obj/x86/asm6.go"))
	if err != nil {
		t.Fatal(err)
	}
	if err := x86LockScalarEncoderCarrier(source); err != nil {
		t.Fatal(err)
	}
	catalog, err := goasmrules.LoadX86LegacyRules(runtime.GOROOT())
	if err != nil {
		t.Fatal(err)
	}
	names, rows := 0, 0
	for _, entry := range catalog.Optab {
		if legacyGPRMWDraftNames.MatchString(entry.Opcode) {
			names++
			rows += len(entry.Rows)
		}
	}
	if names != 53 || rows != 169 {
		t.Fatal("complete source family changed", names, rows)
	}
}

func TestX86OfficialScalarRMWFirstMatchActualGoBytes(t *testing.T) {
	// These are source/encoding anchors, deliberately not claimed as observable
	// translation or numerical execution. Keep the AL/AX original spelling.
	cases := []struct {
		instruction string
		encoding    []byte
	}{
		{"ADDB $128,AL", []byte{0x04, 0x80}},
		{"ADDB $128,AX", []byte{0x80, 0xc0, 0x80}},
		{"ADDL $1,AX", []byte{0x83, 0xc0, 0x01}},
		{"ADDL $128,AX", []byte{0x05, 0x80, 0, 0, 0}},
		{"ADDL $128,BX", []byte{0x81, 0xc3, 0x80, 0, 0, 0}},
		{"ADDL CX,BX", []byte{0x01, 0xcb}},
		{"ADDL 0(SI),BX", []byte{0x03, 0x1e}},
	}
	for _, arch := range []string{"amd64", "386"} {
		for _, row := range cases {
			t.Run(arch+"/"+row.instruction, func(t *testing.T) {
				source := "TEXT f(SB),4,$0-0\n" + row.instruction + "\nRET\n"
				actual := x86PackedShiftActualGoBytes(t, arch, source)
				expected := append(append([]byte(nil), row.encoding...), 0xc3)
				if !bytes.Equal(actual, expected) {
					t.Fatalf("original selected row bytes=%x want=%x", actual, expected)
				}
			})
		}
		for _, row := range []struct {
			instruction  string
			short, modrm byte
		}{
			{"INCL AX", 0x40, 0xc0}, {"DECL AX", 0x48, 0xc8},
		} {
			actual := x86PackedShiftActualGoBytes(t, arch, "TEXT f(SB),4,$0-0\n"+row.instruction+"\nRET\n")
			expected := []byte{row.short, 0xc3}
			if arch == "amd64" {
				expected = []byte{0xff, row.modrm, 0xc3}
			}
			if !bytes.Equal(actual, expected) {
				t.Fatalf("Px1 mode selection %s/%s: %x want=%x", arch, row.instruction, actual, expected)
			}
		}
	}
}

func TestX86OfficialScalarRMWRegisterCarrySeedObjects(t *testing.T) {
	catalog, err := goasmrules.LoadX86LegacyRules(runtime.GOROOT())
	if err != nil {
		t.Fatal(err)
	}
	llc := findLLVM22Tool("llc")
	if llc == "" {
		t.Fatal("LLVM22 llc required")
	}
	for _, arch := range []string{"amd64", "386"} {
		word, move, triple, want := 8, "Q", "x86_64-unknown-linux-gnu", 106
		if arch == "386" {
			word, move, triple, want = 4, "L", "i386-unknown-linux-gnu", 78
		}
		count, batch := 0, 0
		var source, declarations strings.Builder
		flush := func() {
			if source.Len() == 0 {
				return
			}
			// Full sources contain initialized real inputs and two observable
			// results, not no-result/unused statements or a constant7 return.
			x86LEAActualTargetGoObject(t, arch, triple, source.String())
			for _, api := range []string{"text", "module", "Go"} {
				ir, err := x86ScalarFPBitViewAPI(t, source.String(), declarations.String(), arch, triple, api)
				if err != nil || ir == "" {
					t.Fatalf("observable GP producer %s/%s/batch%d: %v", arch, api, batch, err)
				}
				for _, level := range []string{"-O0", "-O2"} {
					x86PackedStringCompileObject(t, llc, triple, ir, level)
				}
			}
			source.Reset()
			declarations.Reset()
			batch++
		}
		for _, entry := range catalog.Optab {
			op := entry.Opcode
			if !legacyGPRMWDraftNames.MatchString(op) || arch == "386" && strings.HasSuffix(op, "Q") {
				continue
			}
			width, stem := op[len(op)-1], op[:len(op)-1]
			operand, destination := "CX", "AX"
			if width == 'B' {
				operand, destination = "CL", "AL"
			}
			if strings.HasPrefix(stem, "BT") {
				operand = "$1"
			}
			instruction := op + " " + operand + "," + destination
			if stem == "INC" || stem == "DEC" || stem == "NEG" || stem == "NOT" {
				instruction = op + " " + destination
			}
			for _, seed := range []string{"CLC", "STC"} {
				name := fmt.Sprintf("GPRMW%d", count)
				fmt.Fprintf(&declarations, "func %s(old,value uintptr) (data,flags uintptr)\n", name)
				fmt.Fprintf(&source, "TEXT ·%s(SB),4,$0-%d\nMOV%s old+0(FP),AX\nMOV%s value+%d(FP),CX\nCMPL CX,CX\n%s\n%s\n", name, word*4, move, move, word, seed, instruction)
				// MOV observes actual data without changing status. POP's
				// partial-bit identity must be consumed by the next AND only.
				fmt.Fprintf(&source, "MOV%s AX,data+%d(FP)\nPUSHF%s\nPOP%s AX\nAND%s $%d,AX\nMOV%s AX,flags+%d(FP)\nRET\n", move, word*2, move, move, move, x86ScalarRuntimeStatusMask(stem), move, word*3)
				count++
				if count%32 == 0 {
					flush()
				}
			}
		}
		flush()
		if count != want {
			t.Fatal("observable full mnemonic closure changed", arch, count, want)
		}
	}
}

// This adapter gives the GP row its real input/output roles; only the existing
// byte-listing observer consumes a packed-case shell, with no packed initializer
// or vector fields granting data, FLAGS or source-storage capabilities.
func x86OfficialScalarRMWAssembly(cases []x86OfficialPackedCase, witnesses []goasmrules.X86ScalarRMWWitness, arch string) (string, string) {
	word, move := 8, "Q"
	if arch == "386" {
		word, move = 4, "L"
	}
	var source, declarations strings.Builder
	for index, witness := range witnesses {
		fixture := &cases[index]
		fixture.symbol = fmt.Sprintf("GPRow%d", index)
		stem := strings.Fields(witness.Instruction)[0]
		stem = stem[:len(stem)-1]
		memoryResult := witness.Destination == "0(SI)"
		result, size := "(data,flags uintptr)", word*5
		if memoryResult {
			result, size = "(flags uintptr)", word*4
		}
		fmt.Fprintf(&declarations, "func %s(memory *[16]byte,old,value uintptr) %s\n", fixture.symbol, result)
		fmt.Fprintf(&source, "TEXT ·%s(SB),4,$0-%d\nMOV%s memory+0(FP),SI\nMOV%s old+%d(FP),AX\nMOV%s AX,BX\nMOV%s value+%d(FP),CX\nCMPL CX,CX\nCLC\n",
			fixture.symbol, size, move, move, word, move, move, word*2)
		fixture.line = strings.Count(source.String(), "\n") + 1
		fmt.Fprintln(&source, witness.Instruction)
		flagOffset := word * 3
		if !memoryResult {
			register := "BX"
			if witness.Destination == "AX" || witness.Destination == "AL" {
				register = "AX"
			}
			fmt.Fprintf(&source, "MOV%s %s,data+%d(FP)\n", move, register, word*3)
			flagOffset += word
		}
		fmt.Fprintf(&source, "PUSHF%s\nPOP%s AX\nAND%s $%d,AX\nMOV%s AX,flags+%d(FP)\nRET\n",
			move, move, move, x86ScalarRuntimeStatusMask(stem), move, flagOffset)
	}
	return source.String(), declarations.String()
}

func TestX86OfficialScalarRMWRowsSelectedGoBytesThreeAPIsObjects(t *testing.T) {
	catalog, err := goasmrules.LoadX86LegacyRules(runtime.GOROOT())
	if err != nil {
		t.Fatal(err)
	}
	llc := findLLVM22Tool("llc")
	if llc == "" {
		t.Fatal("LLVM22 llc required")
	}
	for _, mode := range []struct {
		arch, triple     string
		selected, shapes int
	}{
		{"amd64", "x86_64-unknown-linux-gnu", 167, 278},
		{"386", "i386-unknown-linux-gnu", 124, 202},
	} {
		t.Run(mode.arch, func(t *testing.T) {
			rows, err := goasmrules.X86ScalarRMWRows(catalog, mode.arch)
			if err != nil {
				t.Fatal(err)
			}
			var witnesses []goasmrules.X86ScalarRMWWitness
			var cases []x86OfficialPackedCase
			selected := 0
			for _, row := range rows {
				if row.ModeFiltered != "" {
					continue
				}
				selected++
				for _, witness := range row.Witnesses {
					cases = append(cases, x86OfficialPackedCase{row: row.ID, witness: goasmrules.X86PackedBinaryWitness{Instruction: witness.Instruction, Encoding: witness.Encoding}})
					witnesses = append(witnesses, witness)
				}
			}
			if len(rows) != 169 || selected != mode.selected || len(witnesses) != mode.shapes {
				t.Fatal("source/mode/selected denominator changed", len(rows), selected, len(witnesses))
			}
			for start := 0; start < len(cases); start += 32 {
				end := start + 32
				if end > len(cases) {
					end = len(cases)
				}
				t.Run(fmt.Sprintf("batch%02d", start/32), func(t *testing.T) {
					batch := cases[start:end]
					source, declarations := x86OfficialScalarRMWAssembly(batch, witnesses[start:end], mode.arch)
					listing, err := x86ScalarLiteralGoListing(t, mode.arch, source)
					if err != nil {
						t.Fatal(err)
					}
					if err := x86OfficialPackedVerifyListing(listing, batch); err != nil {
						t.Fatal("actual selected Go byte identity", err)
					}
					for _, api := range []string{"text", "module", "Go"} {
						t.Run(api, func(t *testing.T) {
							ir, err := x86ScalarFPBitViewAPI(t, source, declarations, mode.arch, mode.triple, api)
							if err != nil || ir == "" {
								t.Fatal("initialized data/defined FLAGS continuation", err)
							}
							for _, level := range []string{"-O0", "-O2"} {
								x86PackedStringCompileObject(t, llc, mode.triple, ir, level)
							}
						})
					}
				})
			}
		})
	}
}

func TestX86OfficialScalarRMWOriginalShadowsAndModeSelection(t *testing.T) {
	catalog, err := goasmrules.LoadX86LegacyRules(runtime.GOROOT())
	if err != nil {
		t.Fatal(err)
	}
	for _, arch := range []string{"amd64", "386"} {
		rows, err := goasmrules.X86ScalarRMWRows(catalog, arch)
		if err != nil {
			t.Fatal(err)
		}
		byName := map[string][]goasmrules.X86ScalarRMWRow{}
		for _, row := range rows {
			byName[row.Opcode] = append(byName[row.Opcode], row)
		}
		var cases []x86OfficialPackedCase
		var source strings.Builder
		add := func(from, selected goasmrules.X86ScalarRMWRow, witness goasmrules.X86ScalarRMWWitness) {
			name := fmt.Sprintf("Shadow%d", len(cases))
			line := strings.Count(source.String(), "\n") + 2
			fmt.Fprintf(&source, "TEXT ·%s(SB),4,$0-0\n%s\nRET\n", name, witness.Instruction)
			cases = append(cases, x86OfficialPackedCase{row: from.ID + " -> " + selected.ID, symbol: name, line: line,
				witness: goasmrules.X86PackedBinaryWitness{Instruction: witness.Instruction, Encoding: witness.Encoding}})
		}
		for _, entry := range catalog.Optab {
			group := byName[entry.Opcode]
			if len(group) == 0 || arch == "386" && entry.Prefix == "Pw" {
				continue
			}
			switch entry.Ytab {
			case "yxorb":
				add(group[1], group[0], group[0].Witnesses[0])
				add(group[3], group[2], group[2].Witnesses[0])
				// Class match sees AX before Pb converts it to AL. Therefore
				// this spelling selects generic /digit, not accumulator opcode.
				alias := group[1].Witnesses[0]
				alias.Instruction = strings.Replace(alias.Instruction, "BL", "AX", 1)
				alias.Encoding = append([]byte(nil), alias.Encoding...)
				alias.Encoding[1] &^= 7
				add(group[0], group[1], alias)
			case "yaddl":
				short := group[0].Witnesses[0]
				short.Instruction = strings.Replace(short.Instruction, "BX", "AX", 1)
				short.Encoding = append([]byte(nil), short.Encoding...)
				short.Encoding[len(short.Encoding)-2] &^= 7
				add(group[1], group[0], short)
				add(group[2], group[0], group[0].Witnesses[0])
				add(group[2], group[1], group[1].Witnesses[0])
				add(group[4], group[3], group[3].Witnesses[0])
			case "yincl":
				if arch == "386" {
					add(group[1], group[0], group[0].Witnesses[0])
				} else {
					add(group[0], group[1], group[1].Witnesses[0])
				}
			}
		}
		want := 107
		if arch == "386" {
			want = 79
		}
		if len(cases) != want {
			t.Fatal("ordered shadow/mode observations changed", arch, len(cases), want)
		}
		listing, err := x86ScalarLiteralGoListing(t, arch, source.String())
		if err != nil {
			t.Fatal(err)
		}
		if err := x86OfficialPackedVerifyListing(listing, cases); err != nil {
			t.Fatal("actual Go first-match/alias selection", arch, err)
		}
		for _, fixture := range cases {
			t.Logf("SCALAR_GO_SELECTION arch=%s mapping=%q source=%q encoding=%x", arch, fixture.row,
				fixture.witness.Instruction, fixture.witness.Encoding)
		}
	}
	// Preserve all45 Q source rows: each has valid amd64 syntax which the actual
	// 386 assembler must reject independently, rather than one combined failing
	// source hiding acceptance of another row. These are source rejection only.
	rows, err := goasmrules.X86ScalarRMWRows(catalog, "amd64")
	if err != nil {
		t.Fatal(err)
	}
	qrows := 0
	for _, row := range rows {
		if !strings.HasSuffix(row.Opcode, "Q") {
			continue
		}
		qrows++
		t.Run(row.ID, func(t *testing.T) {
			if len(row.Witnesses) == 0 {
				t.Fatal("valid amd64 source row lost its syntax")
			}
			listing, _ := x86ScalarLiteralExpectedGoListing(t, "386", "TEXT f(SB),4,$0-0\n"+row.Witnesses[0].Instruction+"\nRET\n", false)
			t.Logf("SCALAR_GO_MODE_REJECTION row=%q source=%q diagnostic=%q", row.ID, row.Witnesses[0].Instruction, listing)
		})
	}
	if qrows != 45 {
		t.Fatal("Q-width mode boundary changed", qrows)
	}
}

func TestX86OfficialScalarRMWObservableInputAndFlagBoundaries(t *testing.T) {
	for _, arch := range []string{"amd64", "386"} {
		move, word, triple := "Q", 8, "x86_64-unknown-linux-gnu"
		if arch == "386" {
			move, word, triple = "L", 4, "i386-unknown-linux-gnu"
		}
		body := fmt.Sprintf("TEXT ·Value(SB),4,$0-%d\nMOV%s old+0(FP),AX\nMOV%s value+%d(FP),CX\nCMPL CX,CX\nCLC\nANDB CL,AL\nMOV%s AX,data+%d(FP)\nPUSHF%s\nPOP%s AX\nAND%s $2245,AX\nMOV%s AX,flags+%d(FP)\nRET\n", word*4, move, move, word, move, word*2, move, move, move, move, word*3)
		for _, control := range []struct {
			name, from, to string
			pass           bool
		}{
			{"complete", "", "", true},
			{"undefined-AF", "$2245", "$2261", false},
			{"unmasked-flags", "AND" + move + " $2245,AX\n", "", false},
			{"expired-popped-bits", "AND" + move + " $2245,AX", "NOP\nAND" + move + " $2245,AX", false},
			{"missing-old", "MOV" + move + " old+0(FP),AX\n", "", false},
			{"missing-source", "MOV" + move + fmt.Sprintf(" value+%d(FP),CX\n", word), "", false},
			{"unknown-carry", "CMPL CX,CX\nCLC\nANDB CL,AL", "ADCB CL,AL", false},
		} {
			t.Run(arch+"/"+control.name, func(t *testing.T) {
				source := body
				if control.from != "" {
					if strings.Count(source, control.from) != 1 {
						t.Fatal("mutation not uniquely source-bound")
					}
					source = strings.Replace(source, control.from, control.to, 1)
				}
				for _, api := range []string{"text", "module", "Go"} {
					ir, err := x86ScalarFPBitViewAPI(t, source, "func Value(old,value uintptr) (data,flags uintptr)", arch, triple, api)
					if control.pass && (err != nil || ir == "") || !control.pass && (ir != "" || !errors.Is(err, ErrProbeNeedsContext)) {
						t.Fatal("real data/defined-status proof changed", api, control.pass, err)
					}
				}
			})
		}
	}
}
