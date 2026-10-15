package plan9asm

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

type x86ScalarFPByteInputCase struct {
	x86ScalarFPTupleSpanCase
	probe string
}

func x86ScalarFPByteInputCases(arch string) []x86ScalarFPByteInputCase {
	word, bits, move := 8, 64, "MOVQ"
	if arch == "386" {
		word, bits, move = 4, 32, "MOVL"
	}
	rows := []x86ScalarFPByteInputCase{
		{x86ScalarFPTupleSpanCase{
			name: "readonly-numeric-inputs", declaration: "func Value(a,b uint8) uint16",
			source: fmt.Sprintf("TEXT ·Value(SB),4,$0-%d\nMOVW a+0(FP),AX\nMOVW AX,ret+%d(FP)\nRET\n", word+2, word),
			want:   []uint64{0x0403},
		}, "define i16 @Probe() {\n %r = call i16 @Value(i8 3,i8 4)\n ret i16 %r\n}\n"},
		{x86ScalarFPTupleSpanCase{
			name: "readonly-canonical-byte", declaration: "func Value(a bool,b uint8) (bool,uint8,uint16)",
			source: fmt.Sprintf("TEXT ·Value(SB),4,$0-%d\nMOVW a+0(FP),AX\nMOVW AX,ret+%d(FP)\nMOVW AX,ret+%d(FP)\nRET\n", word+4, word, word+2),
			want:   []uint64{1, 2, 0x0201},
		}, "define {i1,i8,i16} @Probe() {\n %r = call {i1,i8,i16} @Value(i1 true,i8 2)\n ret {i1,i8,i16} %r\n}\n"},
	}
	read, expected := "a+0(FP)", uint64(0x112233443f800000)
	if arch == "386" {
		read, expected = "a+1(FP)", 0x443f8000
	}
	rows = append(rows, x86ScalarFPByteInputCase{x86ScalarFPTupleSpanCase{
		name: "floating-input-bits", declaration: fmt.Sprintf("func Value(a float32,b uint32) uint%d", bits),
		source: fmt.Sprintf("TEXT ·Value(SB),4,$0-%d\n%s %s,AX\n%s AX,ret+8(FP)\nRET\n", 8+word, move, read, move),
		want:   []uint64{expected},
	}, fmt.Sprintf("define i%d @Probe() {\n %%r = call i%d @Value(float 1.000000e+00,i32 287454020)\n ret i%d %%r\n}\n", bits, bits, bits)})
	body := "MOVQ $0x3ff0000000000000,AX\nMOVQ AX,ret+0(FP)\nMOVB $0x11,ret+1(FP)\nMOVQ ret+0(FP),AX\nMOVQ AX,ret+8(FP)\n"
	if arch == "386" {
		body = "MOVL $0,ret+0(FP)\nMOVL $0x3ff00000,ret+4(FP)\nMOVB $0x11,ret+1(FP)\nMOVL ret+0(FP),AX\nMOVL AX,ret+8(FP)\nMOVL ret+4(FP),DX\nMOVL DX,ret+12(FP)\n"
	}
	rows = append(rows, x86ScalarFPByteInputCase{x86ScalarFPTupleSpanCase{
		name: "partial-floating-result-bits", declaration: "func Value() (float64,uint64)",
		source: "TEXT ·Value(SB),4,$0-16\n" + body + "RET\n",
		want:   []uint64{0x3ff0000000001100, 0x3ff0000000001100},
	}, "define {i64,i64} @Probe() {\n %r = call {double,i64} @Value()\n %f = extractvalue {double,i64} %r,0\n %b = bitcast double %f to i64\n %n = extractvalue {double,i64} %r,1\n %a = insertvalue {i64,i64} poison,i64 %b,0\n %v = insertvalue {i64,i64} %a,i64 %n,1\n ret {i64,i64} %v\n}\n"})
	return rows
}

func TestX86IntegerFPByteViewsReadonlyInputsAndFloatingBits(t *testing.T) {
	for _, target := range x86RawReturnTargets {
		for _, row := range x86ScalarFPByteInputCases(target.arch) {
			for _, api := range []string{"text", "module", "Go"} {
				t.Run(target.triple+"/"+row.name+"/"+api, func(t *testing.T) {
					ir, err := x86ScalarFPBitViewAPI(t, row.source, row.declaration, target.arch, target.triple, api)
					if err != nil || ir == "" {
						t.Fatal("complete typed bit transport rejected", err)
					}
					// A test-owned ordinary LLVM caller supplies independent input
					// constants; it adds no source/native/entry permission to Value.
					x86ScalarFPNamedConstantResult(t, ir+row.probe, "Probe", row.want)
				})
			}
		}
	}
}

var x86ScalarFPSparseCompleteCase = x86ScalarFPTupleSpanCase{
	name: "complete-byte-writes", declaration: "func Value() (uint32,uint8)",
	source: "TEXT ·Value(SB),4,$0-5\nMOVW $0x0201,ret+0(FP)\nMOVW $0x0403,ret+2(FP)\nMOVB ret+1(FP),AX\nMOVB AX,ret+4(FP)\nRET\n",
	want:   []uint64{0x04030201, 2},
}

func TestX86IntegerFPByteViewsSparseMustInitialization(t *testing.T) {
	for _, target := range x86RawReturnTargets {
		for _, row := range []struct{ name, body string }{
			{"only-low-byte", "MOVB $1,ret+0(FP)"},
			{"only-high-word", "MOVW $1,ret+2(FP)"},
			{"sparse-GP-source", "MOVB $1,AX\nMOVL AX,ret+0(FP)"},
			{"undefined-interior-read", "MOVB ret+1(FP),AX\nMOVL AX,ret+0(FP)"},
		} {
			for _, api := range []string{"text", "module", "Go"} {
				t.Run(target.triple+"/"+row.name+"/"+api, func(t *testing.T) {
					ir, err := x86ScalarFPBitViewAPI(t, "TEXT ·Value(SB),4,$0-4\n"+row.body+"\nRET\n", "func Value() uint32", target.arch, target.triple, api)
					if !errors.Is(err, ErrProbeNeedsContext) || ir != "" {
						t.Fatalf("implementation zeroes initialized an unknown source byte: output=%t, error=%v", ir != "", err)
					}
				})
			}
		}
		for _, api := range []string{"text", "module", "Go"} {
			t.Run(target.triple+"/complete/"+api, func(t *testing.T) {
				row := x86ScalarFPSparseCompleteCase
				ir, err := x86ScalarFPBitViewAPI(t, row.source, row.declaration, target.arch, target.triple, api)
				if err != nil || ir == "" {
					t.Fatal("complete independent partial writes rejected", err)
				}
				x86ScalarFPTupleSpanConstantResult(t, ir, row.want)
			})
		}
	}
}

func TestX86IntegerFPByteViewsMOVQSourceAliases(t *testing.T) {
	for _, target := range x86RawReturnTargets {
		if target.arch != "amd64" {
			continue // Mode32's legal X/m64 and rejected GP controls are separate.
		}
		for _, read := range []bool{false, true} {
			for _, row := range x86ScalarFPTupleSpanCases(target.arch, read) {
				if !strings.Contains(row.source, "MOVQ") {
					continue
				}
				for _, alias := range []string{"MOVD", "MOVDQ2Q"} {
					for _, api := range []string{"text", "module", "Go"} {
						t.Run(target.triple+"/"+alias+"/"+row.name+"/"+api, func(t *testing.T) {
							source := strings.ReplaceAll(row.source, "MOVQ", alias)
							ir, err := x86ScalarFPBitViewAPI(t, source, row.declaration, target.arch, target.triple, api)
							if err != nil || ir == "" {
								t.Fatal("official source alias lost exact byte transport", err)
							}
							x86ScalarFPTupleSpanConstantResult(t, ir, row.want)
						})
					}
				}
			}
		}
	}
}
