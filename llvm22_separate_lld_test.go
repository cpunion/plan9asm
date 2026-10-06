package plan9asm

import (
	"testing"
)

func TestLLVM22RequiredSeparateWasmLinker(t *testing.T) {
	if path := findLLVM22Tool("wasm-ld"); path == "" {
		t.Fatal("required LLVM 22 wasm linker was not found; LLD may be packaged separately from LLVM_CONFIG")
	}
}

func TestLLVM22VersionMatchingRejectsOtherMajors(t *testing.T) {
	for _, test := range []struct {
		version string
		valid   bool
	}{
		{"Homebrew LLD 22.1.8\n", true},
		{"Debian LLVM version 22.1.8\n", true},
		{"clang version 22.1.8\n", true},
		{"LLD 21.1.8\n", false},
		{"LLD 23.0.0\n", false},
		{"LLD 220.1.0\n", false},
		{"LLVM version 23.0.0\n", false},
		{"clang version 21.1.0\n", false},
	} {
		t.Run(test.version, func(t *testing.T) {
			if llvm22ToolVersionRE.MatchString(test.version) != test.valid {
				t.Fatalf("major version matching changed: %q", test.version)
			}
		})
	}
}
