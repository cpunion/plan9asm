package plan9asm

import (
	"fmt"
	"strings"
	"testing"
)

func x86RawCompactInputObjects(t *testing.T, bank string, expand bool) {
	t.Helper()
	members := []string{"VPCOMPRESSB", "VPCOMPRESSW", "VPCOMPRESSD", "VPCOMPRESSQ", "VCOMPRESSPS", "VCOMPRESSPD"}
	if expand {
		members = []string{"VPEXPANDB", "VPEXPANDW", "VPEXPANDD", "VPEXPANDQ", "VEXPANDPS", "VEXPANDPD"}
	}
	for _, target := range x86RawReturnTargets {
		t.Run(target.triple, func(t *testing.T) {
			word, suffix := goWordSize(target.arch), "Q"
			if word == 4 {
				suffix = "L"
			}
			var source, declarations strings.Builder
			count := 0
			for _, member := range members {
				for _, memory := range []bool{false, true} {
					for _, mode := range []string{"plain", "merge", "zero"} {
						if memory && !expand && mode == "zero" {
							continue // Reserved EVEX.z memory stores have their own #UD oracle.
						}
						op, mask := member, ""
						if mode != "plain" {
							mask = "K1,"
						}
						if mode == "zero" {
							op += ".Z"
						}
						first, last := bank+"2", bank+"3"
						if memory {
							if expand {
								first = "1(SI)"
							} else {
								last = "1(DI)"
							}
						}
						row := op + " " + first + "," + mask + last
						// The low-bank bytes have the same physical meaning in both
						// modes; Go's named frontend is not the raw decoder oracle.
						literal := x86RawNumericalInstruction(t, "amd64", row)
						name := fmt.Sprint("Compact", count)
						fmt.Fprintf(&source, "TEXT ·%s(SB),4,$0-%d\n"+
							"MOV%s out+0(FP),DI\nMOV%s input+%d(FP),SI\nMOV%s old+%d(FP),BX\n"+
							"VMOVUPS (SI),Z2\nVMOVUPS (BX),Z3\nKXNORQ K1,K1,K1\n"+
							"%s\nVMOVUPS Z3,65(DI)\nRET\n", name, 3*word,
							suffix, suffix, word, suffix, 2*word, literal)
						fmt.Fprintf(&declarations, "func %s(out *[144]byte,input,old *[64]byte)\n", name)
						count++
					}
				}
			}
			want := 30
			if expand {
				want = 36
			}
			if count != want {
				t.Fatal("independent full raw compact/expand inventory changed", count, want)
			}
			x86RawNumericalFamilyObjects(t, source.String(), declarations.String(), target.arch, target.triple)
		})
	}
}

func TestX86RawCompressCompleteXInputObjects(t *testing.T) {
	x86RawCompactInputObjects(t, "X", false)
}

func TestX86RawCompressCompleteYInputObjects(t *testing.T) {
	x86RawCompactInputObjects(t, "Y", false)
}

func TestX86RawCompressCompleteZInputObjects(t *testing.T) {
	x86RawCompactInputObjects(t, "Z", false)
}

func TestX86RawExpandCompleteXInputObjects(t *testing.T) {
	x86RawCompactInputObjects(t, "X", true)
}

func TestX86RawExpandCompleteYInputObjects(t *testing.T) {
	x86RawCompactInputObjects(t, "Y", true)
}

func TestX86RawExpandCompleteZInputObjects(t *testing.T) {
	x86RawCompactInputObjects(t, "Z", true)
}

func TestX86RawCompactPhysicalMode32Masks(t *testing.T) {
	count := 0
	for _, member := range []string{
		"VPCOMPRESSB", "VPCOMPRESSW", "VPCOMPRESSD", "VPCOMPRESSQ", "VCOMPRESSPS", "VCOMPRESSPD",
		"VPEXPANDB", "VPEXPANDW", "VPEXPANDD", "VPEXPANDQ", "VEXPANDPS", "VEXPANDPD",
	} {
		for _, bank := range []string{"X", "Y", "Z"} {
			for mask := 1; mask <= 7; mask++ {
				row := fmt.Sprintf("%s %s2,K%d,%s3", member, bank, mask, bank)
				literal := x86RawNumericalInstruction(t, "amd64", row)
				file, err := Parse(ArchAMD64, "TEXT Value(SB),4,$0-0\n"+literal+"RET\n")
				if err != nil {
					t.Fatal(err)
				}
				file, err = normalizeX86RawFile(file, "386")
				if err != nil {
					t.Fatal("valid physical low-bank encoding inherited a named frontend limit", row, err)
				}
				ins := file.Funcs[0].Instrs[1]
				form, known, err := x86PackedCompactProperties(ins.Op, ins, "386")
				if !known || err != nil || form.mask != 1 || ins.Args[1].Reg != Reg(fmt.Sprint("K", mask)) {
					t.Fatal("raw mask/operand identity changed", row, form, err)
				}
				count++
			}
		}
	}
	if count != 252 {
		t.Fatal("complete physical compact mask inventory changed", count)
	}
}
