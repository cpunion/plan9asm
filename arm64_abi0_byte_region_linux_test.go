package plan9asm

import (
	"os"
	"os/exec"
	"runtime"
	"testing"
)

func TestCrossLinuxRuntimeMatrixARM64ABI0ByteRegionOriginal(t *testing.T) {
	var runner []string
	if runtime.GOARCH != "arm64" {
		if runtime.GOARCH != "amd64" || os.Getenv("PLAN9ASM_CROSS_EXEC") != "1" {
			t.Fatal("required ARM64 original-code execution needs native ARM64 or the Linux cross-runtime job")
		}
		qemu, err := exec.LookPath("qemu-aarch64")
		if err != nil {
			t.Fatal("required qemu-aarch64: ", err)
		}
		runner = []string{qemu}
		t.Log("Linux QEMU original ARM64 machine-code observation, not bare-metal native ARM64 proof")
	} else {
		t.Log("Linux native ARM64 original-code observation")
	}
	arm64ABI0ByteRegionOriginalRun(t, "linux", runner)
}
