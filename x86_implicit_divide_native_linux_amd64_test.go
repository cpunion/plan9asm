package plan9asm

import (
	"os/exec"
	"testing"
)

func TestCrossLinuxRuntimeMatrixX86ImplicitDivideOriginalGoAndNativeAMD64(t *testing.T) {
	x87IA32RequireNativeHost(t)
	x86RunImplicitDivideNative(t, x86FloatCompareExecutionTools(t), "amd64")
	x87IA32RequireNativeHost(t)
}

func TestCrossLinuxRuntimeMatrixX86ImplicitDivideOriginalGoAndNative386(t *testing.T) {
	x87IA32RequireNativeHost(t)
	llc := findLLVM22Tool("llc")
	compiler, err := exec.LookPath("i686-linux-gnu-gcc")
	if llc == "" || err != nil {
		t.Fatalf("required native mode32 division counterpart tools: LLVM22=%q compiler=%v", llc, err)
	}
	execution := x86FloatCompareExecution{llc: llc, triple: "i386-unknown-linux-gnu",
		compiler: []string{compiler, "-static", "-no-pie"}}
	x86RunImplicitDivideNative(t, execution, "386")
	x87IA32RequireNativeHost(t)
}
