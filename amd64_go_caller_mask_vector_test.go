package plan9asm

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/xgo-dev/plan9asm/internal/goasmrules"
)

// The denominator comes from every original Go encoder row, not the caller
// registry or the library which first exposed its missing effect registration.
func x86OfficialMaskVectorRows(t *testing.T) []string {
	t.Helper()
	rules, err := goasmrules.LoadX86AVXRules(runtime.GOROOT())
	if err != nil {
		t.Fatal(err)
	}
	family := map[string]bool{
		"VPMOVM2B": true, "VPMOVM2W": true, "VPMOVM2D": true, "VPMOVM2Q": true,
		"VPMOVB2M": true, "VPMOVW2M": true, "VPMOVD2M": true, "VPMOVQ2M": true,
	}
	var rows []string
	for _, rule := range rules {
		if !family[rule.Opcode] {
			continue
		}
		toVector := strings.HasPrefix(rule.Opcode, "VPMOVM2")
		vectorIndex, maskIndex := 0, 1
		if toVector {
			vectorIndex, maskIndex = 1, 0
		}
		if len(rule.Classes) != 2 || rule.Classes[maskIndex] != "Yk" ||
			rule.Prefix != "Pavx" || rule.ZCase != "Zevex_rm_v_r" ||
			rule.ZOffset != 3 || rule.EncodingOffset != rule.RowIndex*3 || len(rule.OpBytes) != 9 {
			t.Fatalf("unmodeled official mask/vector row: %+v", rule)
		}
		bank := map[string]string{"YxrEvex": "X", "YyrEvex": "Y", "Yzr": "Z"}[rule.Classes[vectorIndex]]
		bits := map[string]string{"X": "128", "Y": "256", "Z": "512"}[bank]
		w := "evexW0"
		if strings.HasSuffix(rule.Opcode, "W") || strings.HasSuffix(rule.Opcode, "Q") ||
			strings.Contains(rule.Opcode, "W2M") || strings.Contains(rule.Opcode, "Q2M") {
			w = "evexW1"
		}
		terms := strings.Split(rule.OpBytes[rule.EncodingOffset], " | ")
		if bank == "" || !reflect.DeepEqual(terms, []string{"avxEscape", "evex" + bits, "evexF3", "evex0F38", w}) ||
			rule.OpBytes[rule.EncodingOffset+1] != "0" {
			t.Fatalf("unmodeled encoding generation/attributes: %+v", rule)
		}
		operands := bank + "1,K1"
		if toVector {
			operands = "K1," + bank + "1"
		}
		rows = append(rows, rule.Opcode+" "+operands)
	}
	if len(rows) != 24 {
		t.Fatalf("complete mask/vector Go-row denominator changed: %d", len(rows))
	}
	return rows
}

func TestX86GoCallerMaskVectorCompleteOfficialEffectRegistration(t *testing.T) {
	for _, row := range x86OfficialMaskVectorRows(t) {
		t.Run(row, func(t *testing.T) {
			source := "TEXT f(SB),4,$0-0\n" + row + "\nRET\n"
			goAssemblerOracleOutput(t, "amd64", source, "", true)
			file, err := Parse(ArchAMD64, source)
			if err != nil {
				t.Fatal(err)
			}
			if !x86CallerVectorMoveKnown(file.Funcs[0].Instrs[1]) {
				t.Fatal("Go-accepted complete mask/vector row lacks its caller effect")
			}
		})
	}
}

func TestX86GoCallerMaskVectorActualTypedSourceAndObjects(t *testing.T) {
	for index, row := range x86OfficialMaskVectorRows(t) {
		t.Run(fmt.Sprintf("%02d/%s", index, row), func(t *testing.T) {
			body := "KMOVQ AX,K1\nVMOVUPS (DI),Z1\n" + row + "\n"
			if strings.HasPrefix(row, "VPMOVM2") {
				body += "VMOVSS X1,(DX)"
			} else {
				body += "KMOVQ K1,AX\nMOVL AX,(DX)"
			}
			assembly := strings.Replace(x86CallerFixtureAssembly, "MOVL AX,(DX)", body, 1)
			goAssemblerOracleOutput(t, "amd64", assembly, "", true)
			pkg, _ := x86CallerCoreFixture(t, x86CallerFixtureSource, assembly)
			for _, triple := range []string{"x86_64-apple-darwin", "x86_64-unknown-linux-gnu", "x86_64-pc-windows-msvc"} {
				if err := x86CallerCoreTranslate(t, pkg, assembly, triple); err != nil {
					t.Fatal("actual selected typed caller/LLVM object failed", err)
				}
			}
		})
	}
}
