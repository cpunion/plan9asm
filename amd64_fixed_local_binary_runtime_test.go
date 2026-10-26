//go:build unix

package plan9asm

import (
	"runtime"
	"strings"
	"testing"
)

func TestCrossLinuxRuntimeMatrixGoAMD64FixedLocalTypedBinary(t *testing.T) {
	execution := x86FloatCompareRuntimeTools(t, x86FloatCompareSSE)
	t.Setenv("GOCACHE", t.TempDir())
	assembly, declarations, original, driver := fixedLocalBinaryRuntimeProgram(t, execution.triple)
	want := strings.Split(strings.TrimSpace(string(x86ScalarFloatIntegerOriginalOutput(t, execution, assembly, original))), "\n")
	if len(want) != 28*10*3 {
		t.Fatalf("original Go produced %d rows, want 28 kernels x 10 values x 3 observations", len(want))
	}
	pkg, request, _, options := fixedLocalTransportFixture(t,
		assembly, declarations, runtime.GOOS+"/amd64", execution.triple)
	fixedLocalConversionPrepareToolCache(t, request)
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
					output := execution.compileAndRunLLVM(t, "private-binary", input, driver, level)
					got := strings.Split(strings.TrimSpace(string(output)), "\n")
					if len(got) != len(want) {
						t.Fatalf("LLVM row count = %d, original = %d", len(got), len(want))
					}
					for index := range want {
						if got[index] != want[index] {
							t.Fatalf("private packed value/canary row %d: original %s; LLVM %s", index, want[index], got[index])
						}
					}
				})
			}
		})
	if runtime.GOOS == "darwin" && runtime.GOARCH == "arm64" {
		t.Log("original Go and LLVM amd64 execution via Rosetta, not native x86 hardware")
	}
}
