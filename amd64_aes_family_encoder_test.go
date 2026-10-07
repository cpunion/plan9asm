package plan9asm

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"strings"
	"testing"

	"github.com/xgo-dev/plan9asm/internal/goasmrules"
)

func TestX86AESFamilyCompleteIndependentEncoderRows(t *testing.T) {
	root := filepath.Join(runtime.GOROOT(), "src/cmd/internal/obj/x86")
	data, err := os.ReadFile(filepath.Join(root, "asm6.go"))
	if err != nil {
		t.Fatal(err)
	}
	for _, table := range []struct{ name, body string }{{"yaes", "{Zlitm_r, 2, argList{Yxm, Yxr}},"}, {"yxshuf", "{Zibm_r, 2, argList{Yu8, Yxm, Yxr}},"}} {
		start := strings.Index(string(data), "var "+table.name+" = []ytab{")
		if start < 0 {
			t.Fatal(table.name)
		}
		actual := strings.SplitN(strings.SplitN(string(data)[start:], "\n}", 2)[0], "[]ytab{", 2)[1]
		a, err := x86SignedPackedFloatGoOpcodeTokens(actual)
		if err != nil {
			t.Fatal(err)
		}
		w, err := x86SignedPackedFloatGoOpcodeTokens(table.body)
		if err != nil || !reflect.DeepEqual(a, w) {
			t.Fatal("complete legacy zcase/zoffset/classes changed", table.name, actual)
		}
	}
	legacyRows := regexp.MustCompile(`\{A(AES[A-Z]+), ([^,]+), ([^,]+), opBytes\{([^}]+)\}\}`).FindAllStringSubmatch(string(data), -1)
	if len(legacyRows) != 6 || len(x86AESSpecs) != 12 {
		t.Fatal("whole family inventory changed", len(legacyRows), len(x86AESSpecs))
	}
	for i, name := range x86AESOracleNames[:6] {
		table, escape, opcode := "yaes", "0x38", []string{"0xdc", "0xdd", "0xde", "0xdf", "0xdb", "0xdf"}[i]
		if i == 5 {
			table, escape = "yxshuf", "0x3a"
		}
		want := fmt.Sprintf("{A%s, %s, Pq, opBytes{%s, %s, 0}}", name, table, escape, opcode)
		if !strings.Contains(string(data), want) {
			t.Fatal("legacy complete prefix/opcode attributes changed", want)
		}
	}
	rules, err := goasmrules.LoadX86AVXRules(runtime.GOROOT())
	if err != nil {
		t.Fatal(err)
	}
	counts := map[string]int{}
	for _, rule := range rules {
		if !strings.HasPrefix(rule.Opcode, "VAES") {
			continue
		}
		index := counts[rule.Opcode]
		counts[rule.Opcode]++
		isIMC, isKey := rule.Opcode == "VAESIMC", rule.Opcode == "VAESKEYGENASSIST"
		opcode := map[string]string{"VAESENC": "0xDC", "VAESENCLAST": "0xDD", "VAESDEC": "0xDE", "VAESDECLAST": "0xDF", "VAESIMC": "0xDB", "VAESKEYGENASSIST": "0xDF"}[rule.Opcode]
		if opcode == "" {
			t.Fatal("unaccounted actual Go AES member", rule.Opcode)
		}
		table, classes, zcase := "_yvaesdec", []string{"Yxm,Yxr,Yxr", "Yym,Yyr,Yyr", "YxmEvex,YxrEvex,YxrEvex", "YymEvex,YyrEvex,YyrEvex", "Yzm,Yzr,Yzr"}, "Zvex_rm_v_r"
		zoffset, offset := []int{2, 2, 3, 3, 3}, []int{0, 2, 4, 7, 10}
		var opbytes []string
		if isKey {
			table, classes, zcase = "_yvaeskeygenassist", []string{"Yu8,Yxm,Yxr", "Yi8,Yxm,Yxr"}, "Zvex_i_rm_r"
			zoffset, offset = []int{0, 2}, []int{0, 0}
			opbytes = []string{"avxEscape | vex128 | vex66 | vex0F3A | vexW0", opcode}
		} else if isIMC {
			table, classes = "_yvaesimc", []string{"Yxm,Yxr"}
			zoffset, offset = []int{2}, []int{0}
			opbytes = []string{"avxEscape | vex128 | vex66 | vex0F38 | vexW0", opcode}
		} else {
			for _, width := range []int{128, 256} {
				opbytes = append(opbytes, fmt.Sprintf("avxEscape | vex%d | vex66 | vex0F38 | vexW0", width), opcode)
			}
			for _, width := range []int{128, 256, 512} {
				opbytes = append(opbytes, fmt.Sprintf("avxEscape | evex%d | evex66 | evex0F38 | evexW0", width), fmt.Sprintf("evexN%d", width/8), opcode)
			}
			if index >= 2 {
				zcase = "Zevex_rm_v_r"
			}
		}
		if index >= len(classes) || rule.RowIndex != index || rule.Ytab != table || rule.Prefix != "Pavx" || rule.ZCase != zcase || rule.ZOffset != zoffset[index] || rule.EncodingOffset != offset[index] || strings.Join(rule.Classes, ",") != classes[index] || !reflect.DeepEqual(rule.OpBytes, opbytes) {
			t.Fatal("independent entire encoder row/attributes changed", rule)
		}
	}
	for i, name := range x86AESOracleNames[6:] {
		want := 5
		if i == 4 {
			want = 1
		}
		if i == 5 {
			want = 2
		}
		if counts[name] != want {
			t.Fatal(name, counts, want)
		}
	}
	aliases, err := os.ReadFile(filepath.Join(runtime.GOROOT(), "src/cmd/asm/internal/arch/arch.go"))
	if err != nil || regexp.MustCompile(`instructions\["V?AES`).Match(aliases) {
		t.Fatal("actual Go AES alias inventory changed", err)
	}
}

func TestX86AESFamilyActualGoRawEncoderForms(t *testing.T) {
	for _, arch := range []string{"amd64", "386"} {
		for _, name := range x86AESOracleNames {
			widths := []int{16}
			if strings.HasPrefix(name, "V") && !strings.Contains(name, "IMC") && !strings.Contains(name, "KEYGEN") {
				widths = append(widths, 32, 64)
			}
			for _, width := range widths {
				bank := map[int]string{16: "X", 32: "Y", 64: "Z"}[width]
				for _, source := range []string{bank + "0", fmt.Sprintf("%d(AX)", width), fmt.Sprintf("-%d(AX)(CX*2)", width), "(FS)", "(GS)"} {
					row, _ := x86AESOracleRow(name, source, bank)
					t.Run(arch+"/"+row, func(t *testing.T) {
						code := x86OrderingHintGoBytes(t, arch, "TEXT f(SB),4,$0-0\n"+row+"\nRET\n")
						file, err := Parse(ArchAMD64, "TEXT f(SB),4,$0-0\n"+x86DwordProductDirectives(code[:len(code)-1])+"\nRET\n")
						if err != nil {
							t.Fatal(err)
						}
						decoded, err := decodeX86RawDirectives(file.Funcs[0], arch)
						if err != nil {
							t.Fatal("actual Go AES physical row missing", row, fmt.Sprintf("%x", code), err)
						}
						ins := decoded.Instrs[1]
						form, known, err := x86AESProperties(ins.Op, ins, arch)
						if !known || err != nil || form.byteWidth != width {
							t.Fatal("actual Go raw AES identity", ins, err)
						}
						named := x86YMMReadSource(t, row)
						for i := range named.Args {
							if named.Args[i].Kind == OpMem && (named.Args[i].Mem.Base == FS || named.Args[i].Mem.Base == GS) {
								named.Args[i].Mem.Segment, named.Args[i].Mem.Base = named.Args[i].Mem.Base, ""
							}
							if named.Args[i].Kind == OpMem && named.Args[i].Mem.Index == "" {
								named.Args[i].Mem.Scale = ins.Args[i].Mem.Scale // no-index scale is not an encoded input
							}
							if named.Args[i].Kind == OpMem && strings.Contains(name, "KEYGEN") {
								named.Args[i].Mem.Segment = "" // actual From3 prefix omission, independently bound above
							}
						}
						if ins.Op != named.Op || !reflect.DeepEqual(ins.Args, named.Args) {
							t.Fatalf("physical original operand/tuple width changed %s code=%x actual=%#v named=%#v", row, code, ins.Args, named.Args)
						}
					})
				}
			}
		}
	}
	// Independent fixed byte witness: prefix, map, opcode and immediate. The
	// source encoder is checked independently of our decoder/typed grammar.
	for name, want := range map[string][]byte{"AESENC X0,X1": {0x66, 0x0f, 0x38, 0xdc, 0xc8, 0xc3}, "AESKEYGENASSIST $27,X0,X2": {0x66, 0x0f, 0x3a, 0xdf, 0xd0, 27, 0xc3}, "VAESENC X0,X1,X2": {0xc4, 0xe2, 0x71, 0xdc, 0xd0, 0xc3}, "VAESKEYGENASSIST $-128,X0,X2": {0xc4, 0xe3, 0x79, 0xdf, 0xd0, 128, 0xc3}} {
		got := x86OrderingHintGoBytes(t, "amd64", "TEXT f(SB),4,$0-0\n"+name+"\nRET\n")
		if !bytes.Equal(got, want) {
			t.Fatal("independent fixed AES prefix/opcode bytes", name, got, want)
		}
	}
}

func TestX86AESFamilyThreeAPIsFiveTargetNumericResults(t *testing.T) {
	llc := findLLVM22Tool("llc")
	if llc == "" {
		t.Fatal("LLVM22 required")
	}
	for _, target := range x86RawReturnTargets {
		t.Run(target.triple, func(t *testing.T) {
			var source, declarations strings.Builder
			count := 0
			for _, name := range x86AESOracleNames {
				for _, memory := range []bool{false, true} {
					if memory && !strings.HasPrefix(name, "V") {
						continue
					}
					first, params, offset := "X0", "", 0
					if memory {
						first, params, offset = "a+0(FP)", "a,b uint64", 16
					}
					row, destination := x86AESOracleRow(name, first, "X")
					fmt.Fprintf(&source, "TEXT ·F%d(SB),4,$0-%d\nMOVL $1,AX\nMOVL AX,X0\nMOVL AX,X1\n%s\nMOVQ X%d,ret+%d(FP)\nRET\n", count, offset+8, row, destination, offset)
					fmt.Fprintf(&declarations, "func F%d(%s) uint64\n", count, params)
					count++
				}
			}
			x86LEAActualTargetGoObject(t, target.arch, target.triple, source.String())
			for _, api := range []string{"text", "module", "Go"} {
				t.Run(api, func(t *testing.T) {
					ir, err := x86ScalarFPBitViewAPI(t, source.String(), declarations.String(), target.arch, target.triple, api)
					if err != nil {
						t.Fatal(err)
					}
					compileLLVMToObject(t, llc, target.triple, "aes-results-"+api+".ll", "aes-results-"+api+".o", ir)
				})
			}
		})
	}
}

func TestX86AESFamilyPermissionAndEncodingBoundaries(t *testing.T) {
	for _, name := range x86AESOracleNames {
		for _, source := range []string{"(SP)", "(BP)", "tmp(SP)", "a+0(FP)"} {
			if source == "a+0(FP)" && strings.HasPrefix(name, "V") {
				continue
			}
			row, _ := x86AESOracleRow(name, source, "X")
			goAssemblerOracleOutput(t, "amd64", "TEXT f(SB),4,$0-16\n"+row+"\nRET\n", "", true)
			ir, err := x86ScalarFPBitViewAPI(t, "TEXT ·F(SB),4,$0-16\n"+row+"\nRET\n", "func F(a,b uint64)\n", "amd64", "x86_64-unknown-linux-gnu", "text")
			if !errors.Is(err, ErrProbeNeedsContext) || ir != "" {
				t.Fatal("AES grammar granted unbound native stack/legacy FP storage", row, err)
			}
		}
	}
}
