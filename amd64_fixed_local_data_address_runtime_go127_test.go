//go:build unix && go1.27

package plan9asm

import (
	"fmt"
	"runtime"
	"strings"
	"testing"
)

func TestCrossLinuxRuntimeMatrixGoAMD64FixedLocalReadonlyDataAddress(t *testing.T) {
	assembly, declarations, count := fixedLocalDataAddressKernels()
	if count != 30 {
		t.Fatal("complete DATA address runtime inventory changed", count)
	}
	expected := make([]string, count)
	for index := range expected {
		// The source DATA's third qword is loaded and added to itself.
		expected[index] = "424446484a4c4e50"
	}
	fixedLocalConstantResultsRuntime(t, assembly, declarations, expected)
}

func TestCrossLinuxRuntimeMatrixGoAMD64FixedLocalConstantIndex(t *testing.T) {
	cases := fixedLocalConstantIndexCases()
	if len(cases) != 30 {
		t.Fatal("complete constant index runtime inventory changed", len(cases))
	}
	var assembly, declarations strings.Builder
	assembly.WriteString(fixedLocalAddressData)
	var expected []string
	for index, row := range cases {
		name := fmt.Sprintf("Value%d", index)
		assembly.WriteString(fixedLocalConstantIndexFunction(name, row.body))
		fmt.Fprintf(&declarations, "func %s(a uint64) uint64\n", name)
		// Indexed and fixed-address reads each sum to 19+20+22+26, plus
		// the adjacent private canary 77. The fixed tail is reset each time.
		value := "00000000000000fb"
		if row.literal {
			value = "64686c7074787c80" // Sum of the four independent source DATA qwords.
		}
		expected = append(expected, value)
	}
	fixedLocalConstantResultsRuntime(t, assembly.String(), declarations.String(), expected)
}

func fixedLocalConstantResultsRuntime(t *testing.T, assembly, declarations string, expected []string) {
	t.Helper()
	execution := x86FloatCompareRuntimeTools(t, x86FloatCompareSSE)
	goAMD64FramePrepareOwnedToolCache(t)
	count := len(expected)
	prefix := ""
	if strings.Contains(execution.triple, "apple") {
		prefix = "_"
	}
	var goCalls, cDeclarations, cCalls strings.Builder
	for index := 0; index < count; index++ {
		fmt.Fprintf(&goCalls, "fmt.Printf(\"%%016x\\n\",Value%d(^uint64(0)))\n", index)
		label := runtimeCAsmLabel(prefix + fmt.Sprintf("example.invalid/framefixture.Value%d", index))
		fmt.Fprintf(&cDeclarations, "extern uint64_t value%d(uint64_t) __asm__(%s);\n", index, label)
		fmt.Fprintf(&cCalls, "printf(\"%%016llx\\n\",(unsigned long long)value%d(UINT64_MAX));\n", index)
	}
	original := "package main\nimport \"fmt\"\n" + declarations + "func main() {\n" + goCalls.String() + "}\n"
	want := strings.Split(strings.TrimSpace(string(x86ScalarFloatIntegerOriginalOutput(t, execution, assembly, original))), "\n")
	if len(want) != count {
		t.Fatal("original Go DATA address output inventory changed", len(want), count)
	}
	for index, value := range want {
		if value != expected[index] {
			t.Fatalf("original Go kernel %d returned %s, want %s", index, value, expected[index])
		}
	}
	driver := "#include <stdint.h>\n#include <stdio.h>\n" + runtimeCAsmLabelPreamble +
		cDeclarations.String() + "int main(void) {\n" + cCalls.String() + "return 0;\n}\n"
	pkg, request, _, options := fixedLocalTransportFixture(t, assembly, declarations, runtime.GOOS+"/amd64", execution.triple)
	fixedLocalFreshAPIs(t, pkg, request, options, []string{"Translate", "Module", "GoModule"},
		func(t *testing.T, ir string, err error) {
			if err != nil {
				t.Fatal(err)
			}
			for _, level := range []string{"-O0", "-O2"} {
				t.Run(level, func(t *testing.T) {
					input := ir
					if level == "-O2" {
						input = x86MulDivRoundOptimize(t, ir)
					}
					output := execution.compileAndRunLLVM(t, "readonly-data-address", input, driver, level)
					got := strings.Split(strings.TrimSpace(string(output)), "\n")
					if len(got) != len(want) {
						t.Fatal("LLVM DATA address output inventory changed", len(got), len(want))
					}
					for index := range want {
						if got[index] != want[index] {
							t.Fatalf("DATA address kernel %d: original %s, LLVM %s", index, want[index], got[index])
						}
					}
				})
			}
		})
	if runtime.GOOS == "darwin" && runtime.GOARCH == "arm64" {
		t.Log("original Go and LLVM amd64 execute through Rosetta, not native x86 hardware")
	}
}
