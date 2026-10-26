//go:build unix && go1.27

package plan9asm

import (
	"fmt"
	"runtime"
	"strings"
	"testing"
)

func TestCrossLinuxRuntimeMatrixGoAMD64FixedLocalReadonlyDataAddress(t *testing.T) {
	execution := x86FloatCompareRuntimeTools(t, x86FloatCompareSSE)
	goAMD64FramePrepareOwnedToolCache(t)
	assembly, declarations, count := fixedLocalDataAddressKernels()
	if count != 30 {
		t.Fatal("complete DATA address runtime inventory changed", count)
	}
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
		// The source DATA's third little-endian qword is loaded and added to
		// itself. This oracle does not reuse the translator's address model.
		if value != "424446484a4c4e50" {
			t.Fatalf("original Go kernel %d returned %s, want 424446484a4c4e50", index, value)
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
