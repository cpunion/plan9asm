package plan9asm

import (
	"fmt"
	"strings"
	"testing"
)

func x86LogicalIntegerReturnProgram(t *testing.T, arch, triple string) (*File, Options, string, string) {
	t.Helper()
	rows := []struct {
		typ             LLVMType
		op, value, want string
	}{
		{I1, "MOVB", "2", "0"},
		{I1, "MOVB", "3", "1"},
		{I8, "MOVB", "-17", "239"},
		{I16, "MOVW", "-1234", "64302"},
		{I32, "MOVL", "-1234567", "UINT32_C(4293732729)"},
		{I64, "MOVQ", "-81985529216486896", "UINT64_C(0xfedcba9876543210)"},
	}
	if arch == "386" {
		rows[5].op, rows[5].value = "MOVL", "0x89abcdef"
		rows[5].want = "UINT64_C(0x89abcdef)"
	}
	var source, driver, checks, bridges strings.Builder
	driver.WriteString("#include <stdint.h>\n#include <stdio.h>\n")
	options := Options{Goarch: arch, TargetTriple: triple, Sigs: make(map[string]FuncSig),
		ResolveSym: func(name string) string { return strings.TrimPrefix(name, "·") }}
	for index, row := range rows {
		name := fmt.Sprintf("logicalReturn%d", index)
		fmt.Fprintf(&source, "TEXT ·%s(SB),4,$0-0\n%s $%s,AX\nRET\n", name, row.op, row.value)
		options.Sigs[name] = FuncSig{Name: name, Ret: row.typ}
		bridge := fmt.Sprintf("observeLogicalReturn%d", index)
		fmt.Fprintf(&driver, "extern uint64_t %s(void);\n", bridge)
		fmt.Fprintf(&checks, "if (%s()!=%s) {fprintf(stderr,\"%s\\n\"); return 1;}\n", bridge, row.want, name)
		// FuncSig describes a logical LLVM integer, not C's signext/zeroext
		// narrow-return ABI. Observe that exact type first, then cross the C
		// boundary using an unambiguous uint64_t result on every host.
		fmt.Fprintf(&bridges, "\ndefine i64 @%s() {\n%%value = call %s @%s()\n", bridge, row.typ, name)
		if row.typ == I64 {
			bridges.WriteString("ret i64 %value\n}\n")
		} else {
			fmt.Fprintf(&bridges, "%%wide = zext %s %%value to i64\nret i64 %%wide\n}\n", row.typ)
		}
	}
	fmt.Fprintf(&driver, "int main(void) {for(unsigned row=0;row<257;++row) {%s} return 0;}\n", checks.String())
	file, err := Parse(ArchAMD64, source.String())
	if err != nil {
		t.Fatal(err)
	}
	return file, options, driver.String(), bridges.String()
}

func TestX86LogicalIntegerReturnCompleteTargetObjects(t *testing.T) {
	llc := findLLVM22Tool("llc")
	if llc == "" {
		t.Fatal("logical integer return objects require LLVM22 llc")
	}
	for _, target := range x86RawReturnTargets {
		file, options, _, _ := x86LogicalIntegerReturnProgram(t, target.arch, target.triple)
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
		file, options, driver, bridges := x86LogicalIntegerReturnProgram(t, arch, execution.triple)
		for _, api := range []string{"Translate", "TranslateModule", "TranslateModuleInContext"} {
			t.Run(arch+"/"+api, func(t *testing.T) {
				ir, err := x86PackedFloatStorageAPI(api, file, options)
				if err != nil || ir == "" {
					t.Fatal("complete portable logical integer return source", err)
				}
				ir += bridges
				for _, text := range []string{ir, x86MulDivRoundOptimize(t, ir)} {
					for _, level := range []string{"-O0", "-O2"} {
						execution.compileAndRunLLVM(t, "logical-integer-return", text, driver, level)
					}
				}
			})
		}
	}
}
