package plan9asm

import (
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

func TestX86MOVNTICompleteIndependentGoEncoderRows(t *testing.T) {
	data, err := os.ReadFile(filepath.Join(runtime.GOROOT(), "src/cmd/internal/obj/x86/asm6.go"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	table := regexp.MustCompile(`(?s)var yrl_ml = \[\]ytab\{(.*?)\n\}`).FindStringSubmatch(text)
	if len(table) != 2 || strings.Join(strings.Fields(table[1]), "") !=
		"{Zr_m,1,argList{Yrl,Yml}}," {
		t.Fatal("complete Go MOVNTI grammar, zcase or encoding offset changed")
	}
	rows := regexp.MustCompile(`\{A(MOVNTI[\w]*), yrl_ml, (P\w+), opBytes\{([^}]+)\}\}`).FindAllStringSubmatch(text, -1)
	if len(rows) != 2 {
		t.Fatal("complete Go scalar non-temporal family changed", rows)
	}
	for _, row := range rows {
		prefix, encoding := "Pm", "0xc3"
		if row[1] == "MOVNTIQ" {
			prefix, encoding = "Pw", "0x0f,0xc3"
		} else if row[1] != "MOVNTIL" {
			t.Fatal("unmodeled Go scalar non-temporal member", row[1])
		}
		if row[2] != prefix || strings.Join(strings.Fields(row[3]), "") != encoding {
			t.Fatal("complete Go MOVNTI physical encoding changed", row)
		}
	}
}

func TestX86MOVNTIRejectsCompleteGrammarBeforeIR(t *testing.T) {
	for _, arch := range []string{"amd64", "386"} {
		for _, name := range []string{"MOVNTIL", "MOVNTIQ"} {
			if arch == "386" && name == "MOVNTIQ" {
				continue // The original Go-mode rejection matrix retains this axis.
			}
			original := x86YMMReadSource(t, name+" AX,(BX)")
			for _, control := range []struct {
				name   string
				change func(*Instr)
			}{
				{"dispatch", func(i *Instr) { i.Op = "MOVL" }},
				{"suffix", func(i *Instr) { i.Op += ".Z" }},
				{"arity", func(i *Instr) { i.Args = i.Args[:1] }},
				{"immediate-source", func(i *Instr) { i.Args[0] = Operand{Kind: OpImm, Imm: 1} }},
				{"vector-source", func(i *Instr) { i.Args[0] = Operand{Kind: OpReg, Reg: "X0"} }},
				{"vector-destination", func(i *Instr) { i.Args[1] = Operand{Kind: OpReg, Reg: "X0"} }},
				{"address-immediate", func(i *Instr) { i.Args[1] = Operand{Kind: OpSym, Sym: "$16(BX)"} }},
			} {
				t.Run(arch+"/"+name+"/"+control.name, func(t *testing.T) {
					ins := original
					ins.Args = append([]Operand(nil), original.Args...)
					control.change(&ins)
					var ir strings.Builder
					ctx := &amd64Ctx{b: &ir, goarch: arch}
					known, terminated, err := ctx.lowerMOVNTI(Op(name), ins)
					if !known || terminated || err == nil || ir.Len() != 0 || ctx.tmp != 0 {
						t.Fatal("incomplete MOVNTI grammar emitted before rejection", known, terminated, err, ir.String())
					}
				})
			}
		}
	}
}

func TestX86MOVNTIRegisterDestinationFaultsBeforeInputs(t *testing.T) {
	for _, arch := range []string{"amd64", "386"} {
		for _, name := range []string{"MOVNTIL", "MOVNTIQ"} {
			if arch == "386" && name == "MOVNTIQ" {
				continue
			}
			t.Run(arch+"/"+name, func(t *testing.T) {
				// Go's Yml compatibility closure admits mod=3, but MOVNTI has
				// no ISA register destination. No source byte or memory input
				// may be read before this unconditional architectural #UD.
				ins := x86YMMReadSource(t, name+" AX,BX")
				var ir strings.Builder
				ctx := &amd64Ctx{b: &ir, goarch: arch}
				known, terminated, err := ctx.lowerMOVNTI(ins.Op, ins)
				if !known || !terminated || err != nil || ctx.tmp != 0 ||
					ir.String() != "  call void asm sideeffect \"ud2\", \"~{memory}\"()\n  unreachable\n" {
					t.Fatal("register MOVNTI did not fault before inputs", known, terminated, err, ir.String())
				}
			})
		}
	}
}

func TestX86MOVNTILogicalStorageDoesNotSupplyNativeBacking(t *testing.T) {
	for _, target := range x86RawReturnTargets {
		for _, source := range []string{
			"TEXT ·Value(SB),4,$0-4\nMOVL $1,AX\nMOVNTIL AX,ret+0(FP)\nRET\n",
			"TEXT ·Value(SB),4,$16-4\nMOVL $1,AX\nMOVNTIL AX,tmp-4(SP)\nMOVL $1,ret+0(FP)\nRET\n",
		} {
			t.Run(target.triple+"/"+strings.Split(source, "\n")[0], func(t *testing.T) {
				x86LEAActualTargetGoObject(t, target.arch, target.triple, source)
				for _, api := range []string{"text", "module", "Go"} {
					ir, err := x86ScalarFPBitViewAPI(t, source, "func Value() uint32", target.arch, target.triple, api)
					if ir != "" || !errors.Is(err, ErrProbeNeedsContext) {
						t.Fatal("logical numerical bytes impersonated MOVNTI native storage", api, err)
					}
				}
			})
		}
	}
}
