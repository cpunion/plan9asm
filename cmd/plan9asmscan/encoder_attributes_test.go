package main

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
)

func TestX86EncoderCatalogRetainsEncodingRulesBeyondOperandClasses(t *testing.T) {
	const source = `package x86
var shared = []ytab{
	{zcase: Zvex_rm_v_r, zoffset: 2, args: argList{Yxm, Yxr, Yxr}},
	{zcase: Zevex_rm_v_r, zoffset: 0, args: argList{Yzm, Yzr, Yzr}},
	{zcase: Zevex_rm_v_k_r, zoffset: 3, args: argList{Yzm, Yzr, Yknot0, Yzr}},
}
var avxOptab = [...]Optab{
	{as: AVPMULLW, ytab: shared, prefix: Pavx, op: opBytes{
		avxEscape | vex128 | vex66 | vex0F | vexW0, 0xD5,
		avxEscape | evex512 | evex66 | evex0F | evexW0,
		evexN64 | evexZeroingEnabled, 0xD5,
	}},
}
`
	mutations := []struct {
		name, before, after string
	}{
		{"width", "evex512", "evex256"},
		{"broadcast", "evexN64 | evexZeroingEnabled", "evexN64 | evexBcstN4 | evexZeroingEnabled"},
		{"zeroing", "evexN64 | evexZeroingEnabled", "evexN64"},
		{"rounding", "evexN64 | evexZeroingEnabled", "evexN64 | evexRoundingEnabled"},
		{"sae", "evexN64 | evexZeroingEnabled", "evexN64 | evexSaeEnabled"},
		{"tuple", "evexN64", "evexN32"},
		{"w", "evexW0", "evexW1"},
		{"opcode", "0xD5", "0xE5"},
		{"prefix", "prefix: Pavx", "prefix: Pq4"},
		{"zcase", "zcase: Zevex_rm_v_r", "zcase: Zevex_r_v_rm"},
		{"offset", "zoffset: 2", "zoffset: 3"},
	}
	read := func(t *testing.T, text string) []encoderFormReport {
		t.Helper()
		root := t.TempDir()
		path := filepath.Join(root, "src", "cmd", "internal", "obj", "x86", "avx_optabs.go")
		writeEncoderFixture(t, path, text)
		forms, err := loadEncoderForms(root, "amd64")
		if err != nil {
			t.Fatal(err)
		}
		if len(forms) != 3 {
			t.Fatalf("operand-class denominator changed: %d", len(forms))
		}
		return forms
	}
	before := read(t, source)
	beforeJSON, err := json.Marshal(before)
	if err != nil {
		t.Fatal(err)
	}
	for _, mutation := range mutations {
		t.Run(mutation.name, func(t *testing.T) {
			if !strings.Contains(source, mutation.before) {
				t.Fatal("mutation must change an original encoding rule")
			}
			after := read(t, strings.Replace(source, mutation.before, mutation.after, 1))
			for i := range before {
				if before[i].Form != after[i].Form {
					t.Fatal("this regression must retain identical operand classes")
				}
			}
			afterJSON, err := json.Marshal(after)
			if err != nil {
				t.Fatal(err)
			}
			if bytes.Equal(beforeJSON, afterJSON) {
				t.Fatal("encoder catalog discarded an encoding rule despite identical operand classes")
			}
		})
	}
}
