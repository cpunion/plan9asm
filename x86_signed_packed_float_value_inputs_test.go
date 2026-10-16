package plan9asm

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

type x86SignedPackedValueRow struct {
	op, bank, input, mask string
}

// Independent yxcvm2, _yvcvtdq2ps and _yvcvtdq2pd axes. MMX's physical
// x87-state ownership remains a separate negative boundary, not a vector row.
func x86SignedPackedValueRows() []x86SignedPackedValueRow {
	var rows []x86SignedPackedValueRow
	for _, op := range []string{"CVTPL2PS", "CVTPL2PD", "VCVTDQ2PS", "VCVTDQ2PD"} {
		banks := []string{"X"}
		modes := []string{"plain"}
		if strings.HasPrefix(op, "V") {
			banks = []string{"X", "Y", "Z"}
			modes = []string{"plain", "merge", "zero"}
		}
		for _, bank := range banks {
			inputBank := bank
			if op == "VCVTDQ2PD" {
				inputBank = "X"
				if bank == "Z" {
					inputBank = "Y"
				}
			}
			for _, mode := range modes {
				name, mask := op, ""
				if mode != "plain" {
					mask = "K1,"
				}
				if mode == "zero" {
					name += ".Z"
				}
				rows = append(rows, x86SignedPackedValueRow{name, bank, inputBank + "0", mask},
					x86SignedPackedValueRow{name, bank, "3(AX)(DX*2)", mask})
				if strings.HasPrefix(op, "V") {
					rows = append(rows, x86SignedPackedValueRow{strings.Replace(name, ".Z", "", 1) + ".BCST" + strings.TrimPrefix(name, op), bank, "3(AX)(DX*2)", mask})
				}
			}
		}
	}
	for _, rounding := range []string{"RN_SAE", "RD_SAE", "RU_SAE", "RZ_SAE"} {
		for _, mode := range []string{"plain", "merge", "zero"} {
			row := x86SignedPackedValueRow{"VCVTDQ2PS." + rounding, "Z", "Z0", ""}
			if mode != "plain" {
				row.mask = "K1,"
			}
			if mode == "zero" {
				row.op += ".Z"
			}
			rows = append(rows, row)
		}
	}
	return rows
}

func x86SignedPackedValueFixture(arch string, row x86SignedPackedValueRow, result bool) (string, string) {
	word, move := 8, "MOVQ"
	if arch == "386" {
		word, move = 4, "MOVL"
	}
	declaration, finish, extent := "func Value(p *byte,mask uintptr)", "", 2*word
	if result {
		declaration += " uint32"
		extent += 4
		finish = fmt.Sprintf("PEXTRD $0,X1,AX\nMOVL AX,ret+%d(FP)\n", 2*word)
	}
	return fmt.Sprintf("TEXT ·Value(SB),4,$0-%d\n%s p+0(FP),AX\n%s mask+%d(FP),CX\nKMOVD CX,K1\n"+
		"MOVL $1,DX\nVPXORD Z0,Z0,Z0\nVPXORD Z1,Z1,Z1\nSTC\n%s %s,%s%s1\n%sRET\n",
		extent, move, move, word, row.op, row.input, row.mask, row.bank, finish), declaration
}

func TestX86SignedPackedFloatInitializedValuesFiveTargetsThreeAPIs(t *testing.T) {
	llc := findLLVM22Tool("llc")
	if llc == "" {
		t.Fatal("LLVM22 llc required")
	}
	rows := x86SignedPackedValueRows()
	if len(rows) != 70 {
		t.Fatal("complete signed packed conversion denominator changed", len(rows))
	}
	for _, target := range x86RawReturnTargets {
		t.Run(target.triple, func(t *testing.T) {
			for start := 0; start < len(rows); start += 32 {
				end := start + 32
				if end > len(rows) {
					end = len(rows)
				}
				t.Run(fmt.Sprintf("%d-%d", start, end), func(t *testing.T) {
					var source, declarations strings.Builder
					var names []string
					for _, row := range rows[start:end] {
						for _, result := range []bool{false, true} {
							body, declaration := x86SignedPackedValueFixture(target.arch, row, result)
							name := fmt.Sprintf("SignedPackedValue%d", len(names))
							names = append(names, name)
							source.WriteString(strings.ReplaceAll(body, "Value", name))
							declarations.WriteString(strings.ReplaceAll(declaration, "Value", name) + "\n")
						}
					}
					x86LEAActualTargetGoObject(t, target.arch, target.triple, source.String())
					for _, api := range []string{"text", "module", "Go"} {
						t.Run(api, func(t *testing.T) {
							ir, err := x86ScalarFPBitViewAPI(t, source.String(), declarations.String(), target.arch, target.triple, api)
							if err != nil || ir == "" {
								t.Fatal("complete initialized signed packed conversion source rejected", err)
							}
							x86CompactExpandRequireAllDefinitions(t, ir, names)
							for _, text := range []string{ir, x86DwordProductIROptimize(t, ir)} {
								for _, level := range []string{"-O0", "-O2"} {
									x86PackedStringCompileObject(t, llc, target.triple, text, level)
								}
							}
						})
					}
				})
			}
		})
	}
}

func TestX86SignedPackedFloatVoidCannotInventRegisterInputs(t *testing.T) {
	for _, target := range x86RawReturnTargets {
		for _, row := range x86SignedPackedValueRows() {
			if row.mask != "" || strings.Contains(row.input, "(") || strings.Contains(row.op, ".") {
				continue // Plain independent Void rows must not route via KMOV or a result store.
			}
			t.Run(target.triple+"/"+row.op+"/"+row.bank, func(t *testing.T) {
				source := "TEXT ·Value(SB),4,$0-0\nVPXORD Z0,Z0,Z0\nVPXORD Z1,Z1,Z1\n" +
					row.op + " " + row.input + "," + row.bank + "1\nRET\n"
				for _, api := range []string{"text", "module", "Go"} {
					if ir, err := x86ScalarFPBitViewAPI(t, source, "func Value()", target.arch, target.triple, api); err != nil || ir == "" {
						t.Fatal("initialized independent signed conversion rejected", api, err)
					}
					unknown := strings.Replace(source, "VPXORD Z0,Z0,Z0\n", "", 1)
					if ir, err := x86ScalarFPBitViewAPI(t, unknown, "func Value()", target.arch, target.triple, api); !errors.Is(err, ErrProbeNeedsContext) || ir != "" {
						t.Error("Void conversion borrowed unknown vector bytes", api, err)
					}
				}
			})
		}
	}
}

func TestX86SignedPackedFloatNativeInputsStayIndependent(t *testing.T) {
	for index, row := range x86SignedPackedValueRows() {
		if !strings.Contains(row.input, "(") {
			continue
		}
		source, declaration := x86SignedPackedValueFixture("amd64", row, true)
		pkg := mustGoPackage(t, "value", "package value\n"+declaration)
		file := mustParseX86RawReturn(t, source)
		signatures, err := goSigsForAsmFile(pkg, file, func(name string) string {
			return strings.TrimPrefix(name, "·")
		}, "amd64", nil, nil, nil)
		if err != nil {
			t.Fatal(err)
		}
		if err := CheckX86ABI0ResultInitialization(file.Funcs[0], signatures["Value"], "amd64"); !errors.Is(err, ErrProbeNeedsContext) {
			t.Fatal("standalone signed conversion acquired native memory authority", err)
		}
		variants := map[string]string{
			"unknown-pointer": strings.Replace(source, "p+0(FP),AX", "p+0(FP),DI", 1),
			"unknown-index":   strings.Replace(source, "MOVL $1,DX\n", "", 1),
			"numeric-pointer": strings.Replace(source, "MOVL $1,DX", "MOVL $42,AX\nMOVL $1,DX", 1),
		}
		if row.mask != "" {
			variants["unknown-mask"] = strings.Replace(source, "KMOVD CX,K1\n", "", 1)
			if !strings.HasSuffix(row.op, ".Z") {
				variants["unknown-merge"] = strings.Replace(source, "VPXORD Z1,Z1,Z1\n", "", 1)
			}
		}
		for name, changed := range variants {
			t.Run(fmt.Sprintf("%d/%s", index, name), func(t *testing.T) {
				for _, api := range []string{"text", "module", "Go"} {
					if ir, err := x86ScalarFPBitViewAPI(t, changed, declaration, "amd64", "x86_64-unknown-linux-gnu", api); !errors.Is(err, ErrProbeNeedsContext) || ir != "" {
						t.Fatal("signed conversion borrowed native or vector inputs", api, err)
					}
				}
			})
		}
		if row.mask != "" && strings.HasSuffix(row.op, ".Z") {
			t.Run(fmt.Sprintf("%d/zeroing-needs-no-old-destination", index), func(t *testing.T) {
				zero := strings.Replace(source, "VPXORD Z1,Z1,Z1\n", "", 1)
				for _, api := range []string{"text", "module", "Go"} {
					if ir, err := x86ScalarFPBitViewAPI(t, zero, declaration, "amd64", "x86_64-unknown-linux-gnu", api); err != nil || ir == "" {
						t.Fatal("zeroing conversion incorrectly demanded an old destination", api, err)
					}
				}
			})
		}
	}
}
