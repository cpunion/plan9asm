package main

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestLoadX86EncoderForms(t *testing.T) {
	goroot := t.TempDir()
	dir := filepath.Join(goroot, "src", "cmd", "internal", "obj", "x86")
	writeEncoderFixture(t, filepath.Join(dir, "asm6.go"), `package x86
var ynone = []ytab{{Zlit, 1, argList{}}}
var yadd = []ytab{
	{Zibo_m, 2, argList{Yi8, Yml}},
	{Zm_r, 1, argList{Yml, Yrl}},
}
var optab = [...]Optab{
	{AADDQ, yadd, Pw, opBytes{}},
	{obj.ACALL, ynone, Px, opBytes{}},
}
var ymovtab = []movtab{
	{AMOVQ, Yrl, Ynone, Yml, movRegMem2op, [4]uint8{}},
}
`)
	writeEncoderFixture(t, filepath.Join(dir, "avx_optabs.go"), `package x86
var _yvadd = []ytab{{zcase: Zvex, zoffset: 2, args: argList{Yxm, Yxr, Yxr}}}
var avxOptab = [...]Optab{{as: AVADDPS, ytab: _yvadd, prefix: Pavx, op: opBytes{}}}
`)
	forms, err := loadEncoderForms(goroot, "amd64")
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]bool{
		"ADDQ Yi8,Yml":       false,
		"ADDQ Yml,Yrl":       false,
		"CALL":               false,
		"MOVQ Yrl,Yml":       false,
		"VADDPS Yxm,Yxr,Yxr": false,
	}
	for _, form := range forms {
		if _, ok := want[form.Form]; ok {
			want[form.Form] = true
		}
	}
	for form, found := range want {
		if !found {
			t.Errorf("missing encoder form %q in %#v", form, forms)
		}
	}
}

func TestLoadFixedAndGeneratedEncoderForms(t *testing.T) {
	goroot := t.TempDir()
	armDir := filepath.Join(goroot, "src", "cmd", "internal", "obj", "arm")
	writeEncoderFixture(t, filepath.Join(armDir, "asm5.go"), `package arm
var optab = []Optab{{AADD, C_REG, C_NONE, C_REG, 1, 4, 0, 0, 0, 0}}
func buildop() { switch r { case AADD: opset(ASUB, r0) } }
`)
	arm, err := loadEncoderForms(goroot, "arm")
	if err != nil {
		t.Fatal(err)
	}
	assertEncoderForm(t, arm, "ADD from=C_REG,to=C_REG", "")
	assertEncoderForm(t, arm, "SUB from=C_REG,to=C_REG", "ADD")

	arm64Dir := filepath.Join(goroot, "src", "cmd", "internal", "obj", "arm64")
	writeEncoderFixture(t, filepath.Join(arm64Dir, "asm7.go"), `package arm64
var optab = []Optab{{AADD, C_ZREG, C_NONE, C_NONE, C_ZREG, C_NONE, 1, 4, 0, 0, 0}}
func buildop() { switch as { case AADD: oprangeset(ASUB, t) } }
`)
	writeEncoderFixture(t, filepath.Join(arm64Dir, "inst_gen.go"), `package arm64
var insts = [][]instEncoder{{{goOp: AADDVL, fixedBits: 1, args: cimm__XnSP__XdSP}}}
`)
	arm64, err := loadEncoderForms(goroot, "arm64")
	if err != nil {
		t.Fatal(err)
	}
	assertEncoderForm(t, arm64, "ADD from=C_ZREG,to=C_ZREG", "")
	assertEncoderForm(t, arm64, "SUB from=C_ZREG,to=C_ZREG", "ADD")
	assertEncoderForm(t, arm64, "ADDVL args=cimm__XnSP__XdSP", "")
}

func TestLoadEncoderFormsFromCurrentGOROOT(t *testing.T) {
	minimum := map[string]int{"386": 1000, "amd64": 1000, "arm": 100, "arm64": 400}
	for goarch, want := range minimum {
		t.Run(goarch, func(t *testing.T) {
			forms, err := loadEncoderForms(runtime.GOROOT(), goarch)
			if err != nil {
				t.Fatal(err)
			}
			if len(forms) < want {
				t.Fatalf("encoder forms = %d, want at least %d", len(forms), want)
			}
		})
	}
}

func writeEncoderFixture(t *testing.T, path, src string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
}

func assertEncoderForm(t *testing.T, forms []encoderFormReport, wantForm, wantAlias string) {
	t.Helper()
	for _, form := range forms {
		if form.Form == wantForm {
			if form.AliasedFrom != wantAlias {
				t.Fatalf("%s alias = %q, want %q", wantForm, form.AliasedFrom, wantAlias)
			}
			return
		}
	}
	t.Fatalf("missing encoder form %q in %#v", wantForm, forms)
}
