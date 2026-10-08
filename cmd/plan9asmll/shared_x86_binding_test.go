package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExtractSupportedOpsSharedX86RequiresTypedBinding(t *testing.T) {
	cases := []struct {
		name, source string
		want         bool
	}{
		{"typed_parameter", `func f(code Op) { switch code { case "VBOUND": } }`, true},
		{"string_named_op", `func f(op string) { switch op { case "VBOUND": } }`, false},
		{"shadowed_op", `func f(op Op, condition string) { { op := condition; switch op { case "VBOUND": } } }`, false},
		{"var_alias", `func f(code Op) { var name = string(code); switch name { case "VBOUND": } }`, true},
		{"parallel_alias", `func f(code Op, condition string) { other, name := condition, string(code); _ = other; switch name { case "VBOUND": } }`, true},
		{"instruction_field", `func f(source Instr) { switch source.Op { case "VBOUND": } }`, true},
		{"unrelated_field", `func f(source struct{Op string}) { switch source.Op { case "VBOUND": } }`, false},
		{"typed_conversion", `func f(code Op) { switch Op(string(code)) { case "VBOUND": } }`, true},
		{"untyped_conversion", `func f(condition string) { switch Op(condition) { case "VBOUND": } }`, false},
		{"strings_import_alias", `import textstrings "strings"
func f(code Op) { name, _, _ := textstrings.Cut(textstrings.ToUpper(string(code)), "."); switch name { case "VBOUND": } }`, true},
		{"cut_suffix", `import "strings"
func f(op Op) { _, suffix, _ := strings.Cut(string(op), "."); switch suffix { case "VBOUND": } }`, false},
		{"shadowed_strings", `func f(op Op) { strings := struct{ToUpper func(string) string}{}; switch strings.ToUpper(string(op)) { case "VBOUND": } }`, false},
		{"operand_comparison", `func f(op string) { if op == "VBOUND" {} }`, false},
		{"typed_comparison", `func f(code Op) { if code == "VBOUND" {} }`, true},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			dir := t.TempDir()
			for name, source := range map[string]string{
				"translate.go":   "package sample\n",
				"x86_binding.go": "package sample\n" + test.source,
			} {
				if err := os.WriteFile(filepath.Join(dir, name), []byte(source), 0600); err != nil {
					t.Fatal(err)
				}
			}
			for _, arch := range []string{"386", "amd64"} {
				t.Run(arch, func(t *testing.T) {
					supported, err := extractSupportedOps(dir, arch)
					if err != nil {
						t.Fatal(err)
					}
					if _, found := supported["VBOUND"]; found != test.want {
						t.Errorf("opcode role advertised=%t, want %t", found, test.want)
					}
				})
			}
		})
	}
}
