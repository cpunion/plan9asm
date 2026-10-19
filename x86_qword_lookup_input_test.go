package plan9asm

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

func x86QwordLookupCompleteInputObjects(t *testing.T, bank string) {
	t.Helper()
	for _, target := range x86RawReturnTargets {
		t.Run(target.triple, func(t *testing.T) {
			word, suffix := goWordSize(target.arch), "Q"
			if word == 4 {
				suffix = "L"
			}
			var source, declarations strings.Builder
			count := 0
			for _, name := range []string{"VPMULTISHIFTQB", "VPSHUFBITQMB"} {
				modes := []string{"plain"}
				if target.arch == "amd64" {
					modes = append(modes, "merge")
					if name == "VPMULTISHIFTQB" {
						modes = append(modes, "zero")
					}
				}
				for _, mode := range modes {
					inputs := []string{bank + "0", "(SI)"}
					if name == "VPMULTISHIFTQB" {
						inputs = append(inputs, "broadcast")
					}
					for _, input := range inputs {
						op, mask, operand := name, "", input
						if input == "broadcast" {
							op, operand = op+".BCST", "(SI)"
						}
						if mode != "plain" {
							mask = "K1,"
						}
						if mode == "zero" {
							op += ".Z"
						}
						destination, store := bank+"2", "VMOVUPS Z2,(DI)\n"
						if name == "VPSHUFBITQMB" {
							destination, store = "K3", "KMOVQ K3,(DI)\n"
						}
						function := fmt.Sprint("Lookup", count)
						fmt.Fprintf(&source, "TEXT ·%s(SB),4,$0-%d\nMOV%s out+0(FP),DI\n"+
							"MOV%s a+%d(FP),SI\nMOV%s b+%d(FP),BX\nMOV%s old+%d(FP),CX\n"+
							"VMOVUPS (SI),Z0\nVMOVUPS (BX),Z1\nVMOVUPS (CX),Z2\n"+
							"MOVL $-1431655766,AX\nKMOVD AX,K1\nPUSH%s $0\nPOPF%s\n%s %s,%s1,%s%s\n"+
							"%sPUSHF%s\nPOP%s AX\nMOV%s AX,64(DI)\nRET\n",
							function, 4*word, suffix, suffix, word, suffix, 2*word, suffix, 3*word,
							suffix, suffix, op, operand, bank, mask, destination, store, suffix, suffix, suffix)
						fmt.Fprintf(&declarations, "func %s(out *[80]byte,a,b,old *[64]byte)\n", function)
						count++
					}
				}
			}
			want := 13
			if target.arch == "386" {
				want = 5
			}
			if count != want {
				t.Fatal("complete named qword lookup source/mask inventory changed", count, want)
			}
			x86RawNumericalFamilyObjects(t, source.String(), declarations.String(), target.arch, target.triple)
		})
	}
}

func TestX86QwordLookupCompleteXInputObjects(t *testing.T) {
	x86QwordLookupCompleteInputObjects(t, "X")
}
func TestX86QwordLookupCompleteYInputObjects(t *testing.T) {
	x86QwordLookupCompleteInputObjects(t, "Y")
}
func TestX86QwordLookupCompleteZInputObjects(t *testing.T) {
	x86QwordLookupCompleteInputObjects(t, "Z")
}

func TestX86QwordLookupDeadNativeInputsCannotBorrowVoidOrBool(t *testing.T) {
	for _, target := range x86RawReturnTargets {
		for _, instruction := range []string{"VPMULTISHIFTQB (SI),X1,X2", "VPSHUFBITQMB (SI),X1,K3"} {
			for _, declaration := range []string{"func Value()", "func Value(unused bool)"} {
				for _, api := range []string{"text", "module", "Go"} {
					source := "TEXT ·Value(SB),4,$0\n" + instruction + "\nRET\n"
					ir, err := x86ScalarFPBitViewAPI(t, source, declaration, target.arch, target.triple, api)
					if ir != "" || !errors.Is(err, ErrProbeNeedsContext) {
						t.Fatal("dead qword lookup silently discarded its actual source inputs", target, instruction, api, err)
					}
				}
			}
		}
	}
}
