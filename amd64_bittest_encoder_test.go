package plan9asm

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

func TestX86BitTestCompleteGoEncoderRows(t *testing.T) {
	contents, err := os.ReadFile(filepath.Join(runtime.GOROOT(), "src/cmd/internal/obj/x86/asm6.go"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(contents)
	start := strings.Index(text, "var ybtl = []ytab{")
	if start < 0 {
		t.Fatal("actual Go omitted the bit-string encoder table")
	}
	actual := strings.SplitN(strings.SplitN(text[start:], "\n}", 2)[0], "[]ytab{", 2)[1]
	got, err := x86SignedPackedFloatGoOpcodeTokens(actual)
	want, wantErr := x86SignedPackedFloatGoOpcodeTokens(`{Zibo_m,2,argList{Yi8,Yml}}, {Zr_m,1,argList{Yrl,Yml}},`)
	if err != nil || wantErr != nil || !reflect.DeepEqual(got, want) {
		t.Fatal("complete ordered ybtl rows/classes/zcase/offset changed")
	}
	rows := regexp.MustCompile(`(?m)^\s*\{A(BT(?:C|R|S)?[WLQ]),\s*(\w+),\s*(\w+),\s*opBytes\{([^}]+)\}\},`).FindAllStringSubmatch(text, -1)
	if len(rows) != 12 {
		t.Fatalf("complete Go family changed: %d rows", len(rows))
	}
	seen := make(map[string]bool)
	for _, row := range rows {
		op, suffix := row[1], row[1][len(row[1])-1]
		stem := op[:len(op)-1]
		group := map[string]int{"BT": 4, "BTC": 7, "BTR": 6, "BTS": 5}[stem]
		registerOpcode := map[string]int{"BT": 0xa3, "BTC": 0xbb, "BTR": 0xb3, "BTS": 0xab}[stem]
		prefix := map[byte]string{'W': "Pq", 'L': "Pm", 'Q': "Pw"}[suffix]
		encoding := fmt.Sprintf("0xba,0%d,0x%x", group, registerOpcode)
		if suffix == 'Q' {
			encoding = fmt.Sprintf("0x0f,0xba,0%d,0x0f,0x%x", group, registerOpcode)
		}
		got, err := x86SignedPackedFloatGoOpcodeTokens(row[4])
		want, wantErr := x86SignedPackedFloatGoOpcodeTokens(encoding)
		if seen[op] || row[2] != "ybtl" || row[3] != prefix || err != nil || wantErr != nil || !reflect.DeepEqual(got, want) {
			t.Fatal("complete Go family registration/prefix/opcode changed", row)
		}
		seen[op] = true
	}
	for _, op := range x86BitTestOracleNames("amd64") {
		if !seen[op] {
			t.Fatal("independent witness absent from actual Go optab", op)
		}
	}
}

func TestX86BitTestAllImmediateBytesActualGo(t *testing.T) {
	for _, arch := range []string{"386", "amd64"} {
		var source strings.Builder
		var expected []byte
		source.WriteString("TEXT f(SB),4,$0-0\n")
		count := 0
		for _, op := range x86BitTestOracleNames(arch) {
			stem, suffix := op[:len(op)-1], op[len(op)-1]
			group := byte(map[string]int{"BT": 4, "BTC": 7, "BTR": 6, "BTS": 5}[stem])
			for immediate := -128; immediate < 128; immediate++ {
				for _, destination := range []string{"AX", "7(DX)"} {
					fmt.Fprintf(&source, "%s $%d,%s\n", op, immediate, destination)
					if suffix == 'W' {
						expected = append(expected, 0x66)
					} else if suffix == 'Q' {
						expected = append(expected, 0x48)
					}
					expected = append(expected, 0x0f, 0xba)
					if destination == "AX" {
						expected = append(expected, 0xc0|group<<3)
					} else {
						// Go emits exactly the source displacement, not a new
						// displacement inferred from high immediate bits.
						expected = append(expected, 0x42|group<<3, 7)
					}
					expected = append(expected, byte(immediate))
					ins := x86PureGPUnknownInstruction(t, fmt.Sprintf("%s $%d,%s", op, immediate, destination))
					form, known, err := x86BitTestProperties(ins.Op, ins, arch)
					if !known || err != nil || form.index.Imm != int64(immediate) {
						t.Fatal("typed immediate differs from independent original Go byte", ins, err)
					}
					count++
				}
			}
		}
		source.WriteString("RET\n")
		expected = append(expected, 0xc3)
		actual := x86PackedShiftActualGoBytes(t, arch, source.String())
		if !bytes.Equal(actual, expected) {
			t.Fatalf("%s complete original immediate bytes differ: got %d bytes, want %d", arch, len(actual), len(expected))
		}
		t.Logf("%s: %d actual Go bit-string encodings, every imm8, register and memory", arch, count)
	}
}

func TestX86BitTestCompleteFormsFiveTargetsThreeAPIs(t *testing.T) {
	llc := findLLVM22Tool("llc")
	if llc == "" {
		t.Fatal("LLVM22 llc required")
	}
	for _, target := range x86RawReturnTargets {
		t.Run(target.triple, func(t *testing.T) {
			var source, declarations strings.Builder
			count := 0
			for _, op := range x86BitTestOracleNames(target.arch) {
				rows := []string{"$-128,AX", "$127,CX", "CX,AX", "AX,AX", "SP,BP", "CX,SP",
					"$63,7(DX)", "CX,7(DX)", "CX,-3(DX)(SI*4)", "$1,data-3(SB)", "CX,data+3(SB)",
					"$1,8(FS)", "CX,8(GS)", "$1,4096", "CX,4096"}
				if target.arch == "amd64" {
					rows = append(rows, "R8,R15", "R9,-7(R12)(R11*8)")
				} else {
					rows = append(rows, "$4294967295,AX", "CX,data(SB)(SP)", "$1,data(SB)(BX*4)")
				}
				registers := []string{"AX", "CX", "DX", "BX", "SP", "BP", "SI", "DI"}
				if target.arch == "amd64" {
					registers = append(registers, "R8", "R9", "R10", "R11", "R12", "R13", "R14", "R15")
				}
				for _, register := range registers {
					rows = append(rows, register+",AX", "$1,"+register, "AX,"+register)
				}
				for _, row := range rows {
					body := op + " " + row
					ins := x86PureGPUnknownInstruction(t, body)
					if _, known, err := x86BitTestProperties(ins.Op, ins, target.arch); !known || err != nil {
						t.Fatalf("typed shape %s: %v", body, err)
					}
					fmt.Fprintf(&source, "TEXT ·form%d(SB),4,$0-0\n%s\nRET\n", count, body)
					fmt.Fprintf(&declarations, "func form%d()\n", count)
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
					compileLLVMToObject(t, llc, target.triple, "bittest.ll", "bittest.o", ir)
				})
			}
			t.Logf("%d actual Go witnesses, three LLVM API object consumers", count)
		})
	}
}

func TestX86BitTestMalformedAndModeForms(t *testing.T) {
	for _, arch := range []string{"386", "amd64"} {
		for _, op := range x86BitTestOracleNames(arch) {
			for _, row := range []string{"$128,AX", "$-129,AX", "$1.0,AX", "AL,AX", "X0,AX", "AX,X0", "AX,$1", "AX", "(AX),BX", "$1,(AX)(SP*2)"} {
				body := op + " " + row
				goAssemblerOracleOutput(t, arch, "TEXT f(SB),4,$0-0\n"+body+"\nRET\n", "", false)
				ins := x86PureGPUnknownInstruction(t, body)
				if _, known, err := x86BitTestProperties(ins.Op, ins, arch); !known || err == nil {
					t.Errorf("malformed original form admitted: %s %s", arch, body)
				}
			}
			for _, mutate := range []func(*Instr){
				func(i *Instr) { i.Op += ".Z" },
				func(i *Instr) { i.Args[0].ImmRaw = "unknown" },
				func(i *Instr) { i.Args[1].Mem.OffRaw = "unknown" },
				func(i *Instr) { i.Args[1].Mem.Index = CX; i.Args[1].Mem.Scale = 3 },
			} {
				ins := x86PureGPUnknownInstruction(t, op+" $1,(DX)")
				mutate(&ins)
				if _, known, err := x86BitTestProperties(ins.Op, ins, arch); !known || err == nil {
					t.Errorf("typed malformed shape admitted: %s %+v", arch, ins)
				}
			}
		}
	}
	for _, body := range []string{"BTQ AX,BX", "BTCQ $1,(AX)", "BTRQ AX,BX", "BTSQ AX,BX", "BTL R8,AX", "BTSL AX,R15"} {
		goAssemblerOracleOutput(t, "386", "TEXT f(SB),4,$0-0\n"+body+"\nRET\n", "", false)
		ins := x86PureGPUnknownInstruction(t, body)
		if _, known, err := x86BitTestProperties(ins.Op, ins, "386"); !known || err == nil {
			t.Error("386 accepted unavailable native encoding", body)
		}
	}
}
