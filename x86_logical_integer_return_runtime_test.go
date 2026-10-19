package plan9asm

import (
	"fmt"
	"strings"
	"testing"
)

func x86LogicalIntegerReturnProgram(t *testing.T, arch, triple string) (*File, Options, string) {
	t.Helper()
	rows := []struct {
		typ              LLVMType
		op, value, cType string
		want             string
	}{
		{I1, "MOVB", "2", "_Bool", "0"},
		{I1, "MOVB", "3", "_Bool", "1"},
		{I8, "MOVB", "-17", "uint8_t", "239"},
		{I16, "MOVW", "-1234", "uint16_t", "64302"},
		{I32, "MOVL", "-1234567", "uint32_t", "UINT32_C(4293732729)"},
		{I64, "MOVQ", "-81985529216486896", "uint64_t", "UINT64_C(0xfedcba9876543210)"},
	}
	if arch == "386" {
		rows[5].op, rows[5].value = "MOVL", "0x89abcdef"
		rows[5].want = "UINT64_C(0x89abcdef)"
	}
	var source, driver, checks strings.Builder
	driver.WriteString("#include <stdint.h>\n#include <stdio.h>\n")
	options := Options{Goarch: arch, TargetTriple: triple, Sigs: make(map[string]FuncSig),
		ResolveSym: func(name string) string { return strings.TrimPrefix(name, "·") }}
	for index, row := range rows {
		name := fmt.Sprintf("logicalReturn%d", index)
		fmt.Fprintf(&source, "TEXT ·%s(SB),4,$0-0\n%s $%s,AX\nRET\n", name, row.op, row.value)
		options.Sigs[name] = FuncSig{Name: name, Ret: row.typ}
		fmt.Fprintf(&driver, "extern %s %s(void);\n", row.cType, name)
		fmt.Fprintf(&checks, "if (%s()!=%s) {fprintf(stderr,\"%s\\n\"); return 1;}\n", name, row.want, name)
	}
	fmt.Fprintf(&driver, "int main(void) {for(unsigned row=0;row<257;++row) {%s} return 0;}\n", checks.String())
	file, err := Parse(ArchAMD64, source.String())
	if err != nil {
		t.Fatal(err)
	}
	return file, options, driver.String()
}

func TestX86LogicalIntegerReturnCompleteTargetObjects(t *testing.T) {
	llc := findLLVM22Tool("llc")
	if llc == "" {
		t.Fatal("logical integer return objects require LLVM22 llc")
	}
	for _, target := range x86RawReturnTargets {
		file, options, _ := x86LogicalIntegerReturnProgram(t, target.arch, target.triple)
		for _, api := range []string{"Translate", "TranslateModule", "TranslateModuleInContext"} {
			t.Run(target.triple+"/"+api, func(t *testing.T) {
				ir, err := x86PackedFloatStorageAPI(api, file, options)
				if err != nil || ir == "" {
					t.Fatal("complete logical integer return source", err)
				}
				for _, text := range []string{ir, x86MulDivRoundOptimize(t, ir)} {
					for _, level := range []string{"-O0", "-O2"} {
						x86PackedStringCompileObject(t, llc, target.triple, text, level)
					}
				}
			})
		}
	}
}

func TestX86LogicalIntegerReturnPortableExactValues(t *testing.T) {
	execution := x86ScalarCountHostExecution(t)
	for _, arch := range []string{"386", "amd64"} {
		file, options, driver := x86LogicalIntegerReturnProgram(t, arch, execution.triple)
		for _, api := range []string{"Translate", "TranslateModule", "TranslateModuleInContext"} {
			t.Run(arch+"/"+api, func(t *testing.T) {
				ir, err := x86PackedFloatStorageAPI(api, file, options)
				if err != nil || ir == "" {
					t.Fatal("complete portable logical integer return source", err)
				}
				for _, text := range []string{ir, x86MulDivRoundOptimize(t, ir)} {
					for _, level := range []string{"-O0", "-O2"} {
						execution.compileAndRunLLVM(t, "logical-integer-return", text, driver, level)
					}
				}
			})
		}
	}
}
