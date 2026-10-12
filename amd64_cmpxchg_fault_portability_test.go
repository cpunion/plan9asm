package plan9asm

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestX86CMPXCHGFaultWindowsHeaderBoundary(t *testing.T) {
	clang := findLLVM22Tool("clang")
	if clang == "" {
		t.Fatal("LLVM 22 clang required")
	}
	directory := t.TempDir()
	if err := os.Mkdir(filepath.Join(directory, "sys"), 0700); err != nil {
		t.Fatal(err)
	}
	// This observes actual preprocessing of the complete driver, not a
	// Windows SDK, object, link or runtime claim. Missing POSIX dependencies
	// are poisoned so the host cannot accidentally supply them.
	for _, name := range []string{"stdint.h", "stdio.h", "string.h", "stdlib.h", "signal.h", "windows.h"} {
		if err := os.WriteFile(filepath.Join(directory, name), nil, 0600); err != nil {
			t.Fatal(err)
		}
	}
	for _, name := range []string{"unistd.h", "sys/mman.h", "sys/resource.h", "sys/wait.h"} {
		if err := os.WriteFile(filepath.Join(directory, filepath.FromSlash(name)), []byte("#error unexpected_POSIX_fault_dependency\n"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	input := filepath.Join(directory, "fault.c")
	if err := os.WriteFile(input, []byte(x86CMPXCHGFaultReference), 0600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := runtimeOracleStageContext(t)
	defer cancel()
	command := exec.CommandContext(ctx, clang, "-target", "x86_64-w64-windows-gnu", "-nostdinc",
		"-I", directory, "-E", "-P", input)
	output, err := runtimeOracleCombinedOutput(ctx, command)
	if err != nil {
		t.Fatalf("Windows fault driver still consumes POSIX headers: %v\n%s", err, output)
	}
	for _, required := range []string{"VirtualAlloc", "VirtualProtect", "VirtualFree", "kernel.original(p)", "kernel.translated(p)"} {
		if !strings.Contains(string(output), required) {
			t.Fatal("Windows fault driver lost its real guard or operation", required)
		}
	}
}
