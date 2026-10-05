package plan9asm

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

// This fixture establishes only original Go AUTO byte behavior. It does not
// bind a native frame, caller memory, address lifetime or a translation right.
func x86ResultAUTOFixture(t *testing.T, arch, declaration string, frame int, body string) (*File, FuncSig, string) {
	t.Helper()
	_, _, sig, source := x86ResultInitializationFixture(t, arch, declaration, body)
	source = strings.Replace(source, ",$0-", fmt.Sprintf(",$%d-", frame), 1)
	file, err := Parse(ArchAMD64, source)
	if err != nil {
		t.Fatal(err)
	}
	goAssemblerOracleOutput(t, arch, source, "", true)
	return file, sig, source
}

func TestX86ABI0ResultInitializationNamedAUTONumericByteDomain(t *testing.T) {
	for _, arch := range []string{"amd64", "386"} {
		rows := []struct{ typ, move, register string }{
			{"uint8", "MOVB", "AX"}, {"uint16", "MOVW", "AX"}, {"uint32", "MOVL", "AX"},
			{"float32", "MOVSS", "X0"}, {"float64", "MOVSD", "X0"},
		}
		if arch == "amd64" {
			rows = append(rows, struct{ typ, move, register string }{"uint64", "MOVQ", "R14"})
		}
		for _, row := range rows {
			t.Run(arch+"/"+row.move, func(t *testing.T) {
				_, _, sig, _ := x86ResultInitializationFixture(t, arch, "func Value(x "+row.typ+") "+row.typ, "RET")
				body := fmt.Sprintf("%s x+0(FP),%s\n%s %s,spill-16(SP)\n%s spill-16(SP),%s\n%s %s,ret0+%d(FP)\nRET",
					row.move, row.register, row.move, row.register, row.move, row.register, row.move, row.register, sig.Frame.Results[0].Offset)
				file, signature, source := x86ResultAUTOFixture(t, arch, "func Value(x "+row.typ+") "+row.typ, 32, body)
				before := fmt.Sprintf("%#v %#v", file, signature)
				if err := CheckX86ABI0ResultInitialization(file.Funcs[0], signature, arch); err != nil {
					t.Fatalf("actual initialized logical AUTO bytes remain unmodeled: %v\n%s", err, source)
				}
				if after := fmt.Sprintf("%#v %#v", file, signature); after != before {
					t.Fatal("observer changed actual source or typed layout")
				}
			})
		}
	}
	for name, body := range map[string]string{
		"overlapping names are physical bytes": "MOVL $7,first-16(SP)\nMOVL $9,second-12(SP)\nMOVQ third-16(SP),AX\nMOVQ AX,ret0+0(FP)\nRET",
		"all branch stores":                    "MOVL $1,AX\nCMPL AX,$0\nJE other\nMOVQ $7,a-16(SP)\nJMP end\nother:\nMOVQ $9,b-16(SP)\nend:\nMOVQ c-16(SP),AX\nMOVQ AX,ret0+0(FP)\nRET",
		"initialized scalar loop":              "MOVL $3,count-8(SP)\nMOVQ $7,value-16(SP)\nloop:\nMOVL count-8(SP),AX\nSUBL $1,AX\nMOVL AX,count-8(SP)\nCMPL AX,$0\nJNE loop\nMOVQ value-16(SP),AX\nMOVQ AX,ret0+0(FP)\nRET",
		"byte SET into AUTO":                   "MOVQ $0,value-16(SP)\nMOVL $7,AX\nCMPL AX,$0\nSETNE value-16(SP)\nMOVQ value-16(SP),AX\nMOVQ AX,ret0+0(FP)\nRET",
	} {
		t.Run(name, func(t *testing.T) {
			file, sig, source := x86ResultAUTOFixture(t, "amd64", "func Value() uint64", 32, body)
			if err := CheckX86ABI0ResultInitialization(file.Funcs[0], sig, "amd64"); err != nil {
				t.Fatalf("actual defined byte/CFG AUTO input remains unmodeled: %v\n%s", err, source)
			}
		})
	}
}

func TestX86ABI0ResultInitializationAUTOUnknownAndEscapeStayContext(t *testing.T) {
	for name, body := range map[string]string{
		"entry not initialized":       "MOVQ value-16(SP),AX\nMOVQ AX,ret0+0(FP)\nRET",
		"missing upper AUTO bytes":    "MOVL $1,value-16(SP)\nMOVQ value-16(SP),AX\nMOVQ AX,ret0+0(FP)\nRET",
		"missing GP byte stored":      "MOVB $1,AX\nMOVQ AX,value-16(SP)\nMOVQ value-16(SP),AX\nMOVQ AX,ret0+0(FP)\nRET",
		"join missing initialization": "MOVL $1,AX\nCMPL AX,$0\nJE end\nMOVQ $7,value-16(SP)\nend:\nMOVQ value-16(SP),AX\nMOVQ AX,ret0+0(FP)\nRET",
		"loop no entry definition":    "MOVL $1,AX\nJMP condition\nbody:\nMOVQ $1,value-16(SP)\ncondition:\nCMPL AX,$0\nJNE body\nMOVQ value-16(SP),AX\nMOVQ AX,ret0+0(FP)\nRET",
		"AUTO positive offset":        "MOVQ $1,value+0(SP)\nMOVQ value+0(SP),AX\nMOVQ AX,ret0+0(FP)\nRET",
		"past frame start":            "MOVQ $1,value-40(SP)\nMOVQ value-40(SP),AX\nMOVQ AX,ret0+0(FP)\nRET",
		"cross frame end":             "MOVQ $1,value-4(SP)\nMOVQ value-4(SP),AX\nMOVQ AX,ret0+0(FP)\nRET",
		"physical SP not AUTO":        "MOVQ $1,-16(SP)\nMOVQ -16(SP),AX\nMOVQ AX,ret0+0(FP)\nRET",
		"address materialization":     "MOVQ $1,value-16(SP)\nLEAQ value-16(SP),AX\nMOVQ AX,ret0+0(FP)\nRET",
		"SP copied":                   "MOVQ SP,AX\nMOVQ AX,ret0+0(FP)\nRET",
		"SP adjusted":                 "ADDQ $8,SP\nMOVQ $1,value-16(SP)\nMOVQ value-16(SP),AX\nMOVQ AX,ret0+0(FP)\nRET",
		"unknown native load":         "MOVQ $1,value-16(SP)\nMOVQ (R8),AX\nMOVQ AX,ret0+0(FP)\nRET",
		"calling frame":               "MOVQ $1,value-16(SP)\nCALL ·Other(SB)\nMOVQ value-16(SP),AX\nMOVQ AX,ret0+0(FP)\nRET",
		"raw instruction":             "MOVQ $1,value-16(SP)\nBYTE $0x90\nMOVQ value-16(SP),AX\nMOVQ AX,ret0+0(FP)\nRET",
		"dead unknown effect":         "MOVQ $1,value-16(SP)\nJMP end\nMOVQ (R8),AX\nend:\nMOVQ value-16(SP),AX\nMOVQ AX,ret0+0(FP)\nRET",
	} {
		t.Run(name, func(t *testing.T) {
			file, sig, _ := x86ResultAUTOFixture(t, "amd64", "func Value() uint64", 32, body)
			if err := CheckX86ABI0ResultInitialization(file.Funcs[0], sig, "amd64"); !errors.Is(err, ErrProbeNeedsContext) {
				t.Fatalf("unknown bytes/address/call/extent were washed into result initialization: %v", err)
			}
		})
	}
}

func TestX86ABI0ResultInitializationAUTOLiteralRoleCannotBeInvented(t *testing.T) {
	file, sig, _ := x86ResultAUTOFixture(t, "amd64", "func Value() uint64", 32,
		"MOVQ $1,value-16(SP)\nMOVQ value-16(SP),AX\nMOVQ AX,ret0+0(FP)\nRET")
	for name, mutate := range map[string]func(*Func){
		"different parsed offset": func(fn *Func) { fn.Instrs[1].Args[1].Mem.Off = -24 },
		"different original role": func(fn *Func) { fn.Instrs[1].Raw = "MOVQ $1,-16(SP)" },
		"empty original literal":  func(fn *Func) { fn.Instrs[1].Raw = "" },
		"different TEXT frame":    func(fn *Func) { fn.FrameSize = 64 },
		"no AUTO extent":          func(fn *Func) { fn.FrameSize = 0 },
	} {
		t.Run(name, func(t *testing.T) {
			fn := file.Funcs[0]
			fn.Instrs = append([]Instr(nil), fn.Instrs...)
			for index := range fn.Instrs {
				fn.Instrs[index].Args = append([]Operand(nil), fn.Instrs[index].Args...)
			}
			mutate(&fn)
			before := fmt.Sprintf("%#v %#v", fn, sig)
			if err := CheckX86ABI0ResultInitialization(fn, sig, "amd64"); !errors.Is(err, ErrProbeNeedsContext) {
				t.Fatalf("changed AUTO source identity obtained byte knowledge: %v", err)
			}
			if !reflect.DeepEqual(before, fmt.Sprintf("%#v %#v", fn, sig)) {
				t.Fatal("observer modified identity to manufacture AUTO source")
			}
		})
	}
}
