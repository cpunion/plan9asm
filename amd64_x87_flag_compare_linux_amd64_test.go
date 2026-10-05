package plan9asm

import "testing"

// A required cross-runtime counterpart, not a host-skipped 386 object test.
func TestCrossLinuxRuntimeMatrixX87DirectFlagCompare386(t *testing.T) {
	llc := findLLVM22Tool("llc")
	if llc == "" {
		t.Fatal("LLVM 22 llc not found")
	}
	runner := []string{"qemu-i386", "-L", "/usr/i686-linux-gnu"}
	t.Run("OriginalGo", func(t *testing.T) {
		x87DirectFlagCompareOriginalGo(t, "386", runner)
	})
	x87DirectFlagCompareRuntime(t, "386", "i386-unknown-linux-gnu", func(t *testing.T, ir, main string) {
		compileAndRunRuntimeTestWithCompiler(t, llc,
			[]string{"i686-linux-gnu-gcc", "-no-pie"}, "x87-direct-flags386",
			"i386-unknown-linux-gnu", ir, main, runner)
	})
}
