package plan9asm

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestX86LogicalScalarReturnsCannotBorrowImplementationZeros(t *testing.T) {
	for _, target := range x86RawReturnTargets {
		for _, typ := range []LLVMType{I1, I8, I16, I32, I64} {
			for _, body := range []string{"", "MOVL $7,BX\n", "MOVB $7,AH\n"} {
				file, err := Parse(ArchAMD64, "TEXT ·Value(SB),4,$0-0\n"+body+"RET\n")
				if err != nil {
					t.Fatal(err)
				}
				for _, api := range []string{"Translate", "TranslateModule", "TranslateModuleInContext"} {
					t.Run(fmt.Sprintf("%s/%s/%q/%s", target.triple, typ, body, api), func(t *testing.T) {
						options := Options{Goarch: target.arch, TargetTriple: target.triple,
							Sigs:       map[string]FuncSig{"Value": {Name: "Value", Ret: typ}},
							ResolveSym: func(name string) string { return strings.TrimPrefix(name, "·") },
						}
						ir, err := x86PackedFloatStorageAPI(api, file, options)
						if ir != "" || !errors.Is(err, ErrProbeNeedsContext) {
							t.Fatal("logical AX return borrowed implementation zero bytes", err)
						}
					})
				}
			}
		}
	}
}

func TestX86LogicalScalarReturnTransportCompleteWidths(t *testing.T) {
	for _, target := range x86RawReturnTargets {
		for _, row := range []struct {
			typ LLVMType
			op  string
		}{{I1, "MOVB"}, {I8, "MOVB"}, {I16, "MOVW"}, {I32, "MOVL"}, {I64, "MOVQ"}} {
			if target.arch == "386" && row.typ == I64 {
				// The existing logical emitter returns its virtual AX slot;
				// MOVL initializes its zero-extended value. No native EDX:AX
				// or ordinary Go ABI0 declaration follows from this transport.
				row.op = "MOVL"
			}
			for _, api := range []string{"Translate", "TranslateModule", "TranslateModuleInContext"} {
				t.Run(fmt.Sprintf("%s/%s/%s", target.triple, row.typ, api), func(t *testing.T) {
					file, err := Parse(ArchAMD64, "TEXT ·Value(SB),4,$0-0\n"+row.op+" $1,AX\nRET\n")
					if err != nil {
						t.Fatal(err)
					}
					options := Options{Goarch: target.arch, TargetTriple: target.triple,
						Sigs:       map[string]FuncSig{"Value": {Name: "Value", Ret: row.typ}},
						ResolveSym: func(name string) string { return strings.TrimPrefix(name, "·") },
					}
					ir, err := x86PackedFloatStorageAPI(api, file, options)
					if err != nil || ir == "" {
						t.Fatal("complete logical AX return lost its existing transport", err)
					}
					if !errors.Is(CheckX86ABI0ResultInitialization(file.Funcs[0], options.Sigs["Value"], target.arch), ErrProbeNeedsContext) {
						t.Fatal("manual AX transport manufactured a Go ABI0 declaration")
					}
				})
			}
		}
	}
}
