package plan9asm

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func arm64PaddingRuntimeInputs() []uint64 {
	values := []uint64{0, 1, 0x7f, 0x80, 0xff, 0x7fff, 0x8000, 0xffff, 0x7fffffff, 0x80000000, 0xffffffff, ^uint64(0)}
	x := uint64(0x6492b630f84d)
	for i := 0; i < 64; i++ {
		x ^= x << 13
		x ^= x >> 7
		x ^= x << 17
		values = append(values, x)
	}
	return values
}

// Only the native Go caller shim writes padding: the translated function and
// its original native-width load are identical. This makes unspecified bytes
// vary independently of the real narrow argument instead of assuming zero.
func arm64PaddingNativeCaller(test arm64PaddingCase) string {
	move := map[int]string{8: "MOVB", 16: "MOVH", 32: "MOVW"}[test.parameter]
	return fmt.Sprintf(`
TEXT ·pad_%s(SB),4,$80-24
    MOVD noise+8(FP),R0
    MOVD R0,8(RSP)
    %s rhs+0(FP),R0
    %s R0,8(RSP)
    MOVD $0xcafef00d01234567,R12
    MOVD R12,24(RSP)
    MOVD R12,32(RSP)
    MOVD g,R12
    MOVD R12,40(RSP)
    WORD $0xaa1203ec // mov x12,x18: Go's OS-reserved register is not named.
    MOVD R12,48(RSP)
    MOVD 0(RSP),R12
    MOVD R12,56(RSP)
    MOVD RSP,R12
    MOVD R12,64(RSP)
    CALL ·%s(SB)
    MOVD $0xcafef00d01234567,R12
    MOVD 24(RSP),R13
    CMP R12,R13
    BNE bad_%s
    MOVD 32(RSP),R13
    CMP R12,R13
    BNE bad_%s
    MOVD 40(RSP),R12
    CMP g,R12
    BNE bad_%s
    MOVD 48(RSP),R13
    WORD $0xaa1203ec
    CMP R12,R13
    BNE bad_%s
    MOVD 0(RSP),R12
    MOVD 56(RSP),R13
    CMP R12,R13
    BNE bad_%s
    MOVD RSP,R12
    MOVD 64(RSP),R13
    CMP R12,R13
    BNE bad_%s
    MOVD 16(RSP),R0
    MOVD R0,res+16(FP)
    RET
bad_%s:
    MOVD $0xdeadbeeff00dbabe,R0
    MOVD R0,res+16(FP)
    RET
`, test.name, move, move, test.name, test.name, test.name, test.name, test.name, test.name, test.name, test.name)
}

func TestARM64FPPaddingOriginalGoRandomizedBytes(t *testing.T) {
	if runtime.GOARCH != "arm64" {
		t.Skip("original Go runtime requires an ARM64 host; objects are tested independently")
	}
	cases := arm64PaddingCases()
	source, declarations := arm64PaddingSource(cases)
	declarations = strings.Replace(declarations, "package padbits", "package main", 1)
	var driver strings.Builder
	driver.WriteString("package main\n\nimport \"fmt\"\n\n")
	for _, test := range cases {
		fmt.Fprintf(&driver, "func pad_%s(rhs int%d, noise uint64) (res uint64)\n", test.name, test.parameter)
		source += arm64PaddingNativeCaller(test)
	}
	driver.WriteString("\nfunc main() {\n")
	driver.WriteString("    inputs := []uint64{")
	for _, input := range arm64PaddingRuntimeInputs() {
		fmt.Fprintf(&driver, "%d,", input)
	}
	driver.WriteString("}\n    count := 0\n")
	count := 0
	for _, test := range cases {
		driver.WriteString("    {\n        want := []uint64{")
		for _, input := range arm64PaddingRuntimeInputs() {
			fmt.Fprintf(&driver, "%d,", arm64PaddingExpected(test, input))
			count += 4
		}
		fmt.Fprintf(&driver, `}
        for i, input := range inputs {
            // Padding varies independently from the real parameter bits.
            noises := []uint64{0, ^uint64(0), input * 0x9e3779b97f4a7c15, input ^ 0xd1096542b630f84d}
            for _, noise := range noises {
                got := pad_%s(int%d(input), noise)
                if got != want[i] {
                    panic(fmt.Sprintf("%s input=%%x noise=%%x got=%%x want=%%x", input, noise, got, want[i]))
                }
                count++
            }
        }
    }
`, test.name, test.parameter, test.name)
	}
	driver.WriteString("    fmt.Printf(\"padding %d\\n\", count)\n}\n")
	dir := t.TempDir()
	for name, data := range map[string]string{
		"go.mod": "module padbits\n\ngo 1.20\n", "main.go": driver.String(),
		"decl.go": declarations, "padding_arm64.s": source,
	} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(data), 0600); err != nil {
			t.Fatal(err)
		}
	}
	ctx, cancel := runtimeOracleStageContext(t)
	defer cancel()
	command := exec.CommandContext(ctx, "go", "run", ".")
	command.Dir = dir
	command.Env = append(os.Environ(), "CGO_ENABLED=0")
	out, err := runtimeOracleCombinedOutput(ctx, command)
	if err != nil || string(out) != fmt.Sprintf("padding %d\n", count) {
		t.Fatalf("original Go random padding, adjacent stack/g/OS18/LR/SP canaries: %v\n%s", err, out)
	}
	t.Logf("%d original Go calls with randomized ABI0 padding and control/stack canaries", count)
}

func TestARM64FPPaddingLLVMIndependentNumericalAndCanaries(t *testing.T) {
	if runtime.GOARCH != "arm64" {
		t.Skip("LLVM numerical execution requires an ARM64 host; cross objects remain mandatory")
	}
	llc, clang, ok := findLlcAndClang(t)
	if !ok {
		t.Fatal("LLVM22 llc/clang required")
	}
	cases := arm64PaddingCases()
	source, declarations := arm64PaddingSource(cases)
	pkg := mustGoPackage(t, "padbits", declarations)
	triple := testTargetTriple(runtime.GOOS, runtime.GOARCH)
	translated, err := TranslateGoModule(pkg, []byte(source), GoModuleOptions{GOOS: runtime.GOOS, GOARCH: "arm64", TargetTriple: triple, ResolveSym: testResolveSym(pkg.Path)})
	if err != nil {
		t.Fatal(err)
	}
	defer translated.Module.Dispose()
	var driver strings.Builder
	driver.WriteString("#include <stdint.h>\n#include <stdio.h>\n\n")
	for _, test := range cases {
		symbol := "padbits." + test.name
		if runtime.GOOS == "darwin" {
			symbol = "_" + symbol
		}
		fmt.Fprintf(&driver, "extern uint64_t %s(int%d_t) __asm__(%q);\n", test.name, test.parameter, symbol)
	}
	driver.WriteString("\nint main(void) {\n")
	for _, test := range cases {
		for _, input := range arm64PaddingRuntimeInputs() {
			fmt.Fprintf(&driver, `    {
        volatile uint64_t guards[2] = { UINT64_C(0xcafef00d01234567), UINT64_C(0x76543210d00fefac) };
        uint64_t result = %s((int%d_t) UINT64_C(%d));
        if (result != UINT64_C(%d) || guards[0] != UINT64_C(0xcafef00d01234567) ||
                guards[1] != UINT64_C(0x76543210d00fefac)) {
            fprintf(stderr, "%s input=%%llx result=%%llx\n", (unsigned long long) UINT64_C(%d), (unsigned long long) result);
            return 1;
        }
    }
`, test.name, test.parameter, input, arm64PaddingExpected(test, input), test.name, input)
		}
	}
	driver.WriteString("    return 0;\n}\n")
	compileAndRunRuntimeTestForTarget(t, llc, clang, "arm64_fp_padding", triple, translated.Module.String(), driver.String(), nil)
	t.Logf("%d LLVM/C independent numerical cases; no assertion that native padding was initialized", len(cases)*len(arm64PaddingRuntimeInputs()))
}
