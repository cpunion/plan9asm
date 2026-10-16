package plan9asm

import (
	"fmt"
	"runtime"
	"strings"
	"testing"
)

// The original Go AUTO address is above the post-prologue SP. Private PUSH
// storage is below it. Keep those distinct even though parsed AUTO offsets
// are negative. No physical SP/address or native frame permission is supplied.
func x86LogicalAUTOPrivateStackSources(arch string) (string, string, string) {
	word, typ, ctype, bytes := "Q", "uint64", "uint64_t", 8
	if arch == "386" {
		word, typ, ctype, bytes = "L", "uint32", "uint32_t", 4
	}
	var source, declarations, driver strings.Builder
	fmt.Fprintf(&driver, "#include <stdint.h>\n#include <stdio.h>\ntypedef %s word;\n", ctype)
	var checks strings.Builder
	for index, push := range []string{"PUSHW $7", "PUSH" + word + " $7", "PUSHFW", "PUSHF" + word} {
		pop := "POP" + word + " BX"
		if index == 0 || index == 2 {
			pop = "POPW BX"
		}
		name := fmt.Sprintf("AutoStack%d", index)
		fmt.Fprintf(&source, `TEXT ·%s(SB),4,$32-%d
PUSH%s $2
POPF%s
MOV%s value+0(FP),AX
MOV%s AX,saved-%d(SP)
%s
%s
MOV%s saved-%d(SP),AX
MOV%s AX,ret+%d(FP)
RET
`, name, 2*bytes, word, word, word, word, bytes, push, pop, word, bytes, word, bytes)
		fmt.Fprintf(&declarations, "func %s(value %s) %s\n", name, typ, typ)
		fmt.Fprintf(&driver, "extern word %s(word);\n", name)
		fmt.Fprintf(&checks, `{
    word got=%s(value);
    if(got!=value) {
        fprintf(stderr,"AUTO/private-stack row=%d value=%%llx got=%%llx\n",
            (unsigned long long)value,(unsigned long long)got);
        return 1;
    }
}
`, name, index)
	}
	driver.WriteString(`int main(void) {
    const word values[]={0,1,7,255,256,65535,65536,UINT32_MAX,(word)UINT64_MAX,
        (word)UINT64_C(0x1122334455667788),(word)UINT64_C(0x8000000000000000)};
    for(unsigned i=0;i<sizeof(values)/sizeof(values[0]);i++) {
        word value=values[i];
`)
	driver.WriteString(checks.String())
	driver.WriteString("    }\n    return 0;\n}\n")
	return source.String(), declarations.String(), driver.String()
}

func TestX86LogicalAUTOPrivateStackFiveTargetObjects(t *testing.T) {
	llc := findLLVM22Tool("llc")
	if llc == "" {
		t.Fatal("LLVM22 llc required")
	}
	for _, target := range x86RawReturnTargets {
		t.Run(target.triple, func(t *testing.T) {
			source, declarations, _ := x86LogicalAUTOPrivateStackSources(target.arch)
			x86LEAActualTargetGoObject(t, target.arch, target.triple, source)
			for _, api := range []string{"text", "module", "Go"} {
				ir, err := x86ScalarFPBitViewAPI(t, source, declarations, target.arch, target.triple, api)
				if err != nil {
					t.Fatal(api, err)
				}
				for _, text := range []string{ir, x86DwordProductIROptimize(t, ir)} {
					for _, level := range []string{"-O0", "-O2"} {
						x86PackedStringCompileObject(t, llc, target.triple, text, level)
					}
				}
			}
		})
	}
}

func TestX86LogicalAUTOPrivateStackPortableValues(t *testing.T) {
	llc, clang, found := findLlcAndClang(t)
	if !found {
		t.Fatal("LLVM22 llc and clang required")
	}
	triple := testTargetTriple(runtime.GOOS, runtime.GOARCH)
	for _, arch := range []string{"386", "amd64"} {
		source, declarations, driver := x86LogicalAUTOPrivateStackSources(arch)
		for _, api := range []string{"text", "module", "Go"} {
			t.Run(arch+"/"+api, func(t *testing.T) {
				ir, err := x86ScalarFPBitViewAPI(t, source, declarations, arch, triple, api)
				if err != nil {
					t.Fatal(err)
				}
				for _, text := range []string{ir, x86DwordProductIROptimize(t, ir)} {
					for _, level := range []int{0, 2} {
						compileAndRunRuntimeTestWithCompiler(t, llc, []string{clang}, "auto-private-stack-values", triple, text, driver, nil, level)
					}
				}
			})
		}
	}
}
