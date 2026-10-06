package plan9asm

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"testing"

	"golang.org/x/arch/x86/x86asm"
)

// This census is independent of the implementation. Ordered Go tables,
// opcode prefixes and all opBytes remain available to each comparison.
type x86YMMOriginalRule struct {
	name, table, prefix, opBytes string
}

func x86YMMOriginalRules(t *testing.T) []x86YMMOriginalRule {
	t.Helper()
	source, err := os.ReadFile(filepath.Join(runtime.GOROOT(), "src/cmd/internal/obj/x86/asm6.go"))
	if err != nil {
		t.Fatal(err)
	}
	var rules []x86YMMOriginalRule
	pattern := regexp.MustCompile(`\{A([A-Z0-9]+), (ymm|ymmxmm0f38|ymshuf|yps), (\w+), opBytes\{([^}]+)\}\}`)
	for _, row := range pattern.FindAllStringSubmatch(string(source), -1) {
		rules = append(rules, x86YMMOriginalRule{row[1], row[2], row[3], row[4]})
	}
	if len(rules) != 43 {
		t.Fatalf("complete integer MMX-read rule census changed: %d", len(rules))
	}
	return rules
}

func x86YMMReadSource(t *testing.T, row string) Instr {
	t.Helper()
	file, err := Parse(ArchAMD64, "TEXT probe(SB),4,$0-0\n"+row+"\nRET\n")
	if err != nil {
		t.Fatal(err)
	}
	return file.Funcs[0].Instrs[1]
}

func TestX86YMMExistingDescriptorsOriginalBank(t *testing.T) {
	var names []string
	for _, rule := range x86YMMOriginalRules(t) {
		if strings.HasPrefix(rule.name, "PUNPCK") || rule.name == "PAND" || rule.name == "PANDN" ||
			rule.name == "POR" || rule.name == "PXOR" || rule.table == "yps" && !strings.HasPrefix(rule.name, "PSRA") {
			names = append(names, rule.name)
		}
	}
	sort.Strings(names)
	if len(names) != 16 {
		t.Fatal("logical/unpack/uniform logical-shift descriptor class changed", names)
	}
	for _, name := range names {
		for _, source := range []string{"AX", "CX", "DX", "BX", "SP", "BP", "SI", "DI", "R8", "R9", "R10", "R11", "R12", "R13", "R14", "R15"} {
			t.Run(name+"/"+source, func(t *testing.T) {
				ins := x86YMMReadSource(t, name+" "+source+",M1")
				before := fmt.Sprintf("%#v", ins)
				var known bool
				var err error
				switch {
				case strings.HasPrefix(name, "PUNPCK"):
					_, known, err = x86PackedUnpackProperties(ins, "amd64")
				case strings.HasPrefix(name, "PS"):
					_, known, err = x86PackedShiftProperties(ins.Op, ins, "amd64")
				default:
					_, known, err = x86PackedLogicalProperties(ins, "amd64")
				}
				if !known || err != nil {
					t.Errorf("actual Go MMX-read spelling is absent from pure descriptor: %t/%v", known, err)
				}
				if before != fmt.Sprintf("%#v", ins) {
					t.Fatal("pure descriptor mutated original source")
				}
			})
		}
	}
}

func TestX86YMMOriginalGoIntegerReadBanks(t *testing.T) {
	registers := []string{"AX", "CX", "DX", "BX", "SP", "BP", "SI", "DI", "R8", "R9", "R10", "R11", "R12", "R13", "R14", "R15", "M0", "M7"}
	var rows []string
	var expected []x86asm.Reg
	for _, rule := range x86YMMOriginalRules(t) {
		for index, source := range registers {
			prefix := ""
			if rule.table == "ymshuf" {
				prefix = "$27,"
			}
			rows = append(rows, rule.name+" "+prefix+source+",M1")
			bank := index & 7
			if source == "M0" {
				bank = 0
			} else if source == "M7" {
				bank = 7
			}
			expected = append(expected, x86asm.M0+x86asm.Reg(bank))
		}
	}
	// Alias has independent original Go evidence, not a prefix-name inference.
	rows = append(rows, "PADDD AX,M1")
	expected = append(expected, x86asm.M0)
	code := x86PackedShiftActualGoBytes(t, "amd64", "TEXT probe(SB),4,$0-0\n"+strings.Join(rows, "\n")+"\nRET\n")
	for index, row := range rows {
		physical, err := x86asm.Decode(code, 64)
		if err != nil {
			t.Fatal("original object cannot be decoded", row, err)
		}
		if physical.Args[0] != x86asm.M1 || physical.Args[1] != expected[index] {
			t.Fatalf("original Go selected different physical register bank: %s bytes=%x operands=%v expected M1/%s",
				row, code[:physical.Len], physical.Args, expected[index])
		}
		code = code[physical.Len:]
	}
	if len(code) == 0 || code[0] != 0xc3 || !bytes.Equal(bytes.Trim(code[1:], "\x00"), nil) {
		t.Fatalf("original complete native instruction sequence has an unexpected tail: %x", code)
	}
	t.Logf("actual Go integer MMX-read rows=%d, physical M input independent of GP spelling", len(rows))
}

func TestX86YMMOriginalGoOrderedGPControls(t *testing.T) {
	for _, row := range []string{"MOVQ AX,M1", "PINSRB $0,AX,X1", "PINSRD $0,AX,X1", "PINSRQ $0,AX,X1",
		"PEXTRB $0,X1,AX", "PEXTRD $0,X1,AX", "PEXTRQ $0,X1,AX"} {
		t.Run(row, func(t *testing.T) {
			code := x86PackedShiftActualGoBytes(t, "amd64", "TEXT probe(SB),4,$0-0\n"+row+"\nRET\n")
			physical, err := x86asm.Decode(code, 64)
			if err != nil {
				t.Fatal(err)
			}
			for _, arg := range physical.Args {
				if arg == x86asm.M0 {
					t.Fatal("actual transfer/insertion/extraction GP role was changed to MMX", physical)
				}
			}
			if reflect.DeepEqual(physical.Args, x86asm.Args{}) {
				t.Fatal("control has no actual decoded operands")
			}
		})
	}
}
