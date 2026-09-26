package plan9asm

import (
	"runtime"
	"strings"
	"testing"
)

const arm64RawUnlabelledPoolSource = `TEXT rawpool(SB),$0-8
MOVD out+0(FP),R0
WORD $0x10000109 // ADR X9, +32
WORD $0x1000010a // ADR X10, +32
WORD $0xb9400521 // LDR W1, [X9, #4]
WORD $0xb85fc142 // LDUR W2, [X10, #-4]
WORD $0x29000801 // STP W1, W2, [X0]
WORD $0xaa1f03e9 // MOV X9, XZR
WORD $0x9100800a // ADD X10, X0, #32 overwrites the pool pointer.
WORD $0xd65f03c0 // RET
WORD $0x11223344
WORD $0xaabbccdd
WORD $0x17b4a14d // Data that resembles a branch must not be decoded.
RET
`

func TestARM64RawUnlabelledPoolAliases(t *testing.T) {
	requireARM64GoAssemblerResult(t, arm64RawUnlabelledPoolSource, true)
	file, err := Parse(ArchARM64, arm64RawUnlabelledPoolSource)
	if err != nil {
		t.Fatal(err)
	}
	llc, clang, ok := findLlcAndClang(t)
	if !ok {
		t.Fatal("LLVM 22 llc/clang not found")
	}
	for _, triple := range []string{"aarch64-apple-darwin", "aarch64-unknown-linux-gnu", "aarch64-pc-windows-msvc"} {
		t.Run(triple, func(t *testing.T) {
			ir, err := Translate(file, Options{Goarch: "arm64", TargetTriple: triple,
				Sigs: map[string]FuncSig{"rawpool": {Name: "rawpool", Args: []LLVMType{Ptr}, Ret: Void,
					Frame: FrameLayout{Params: []FrameSlot{{Offset: 0, Type: Ptr, Index: 0, Field: -1}}}}}})
			if err != nil {
				t.Fatal(err)
			}
			if strings.Count(ir, "private constant [12 x i8]") != 1 {
				t.Fatalf("pool aliases must share one contiguous global:\n%s", ir)
			}
			compileLLVMToObject(t, llc, triple, "rawpool.ll", "rawpool.o", ir)
			native := runtime.GOOS == "darwin" && triple == "aarch64-apple-darwin" ||
				runtime.GOOS == "linux" && triple == "aarch64-unknown-linux-gnu"
			if runtime.GOARCH == "arm64" && native {
				compileAndRunRuntimeTestForTarget(t, llc, clang, "rawpool", triple, ir, `
#include <stdint.h>
extern void rawpool(uint32_t *);
int main(void) {
  uint32_t result[2] = {0};
  rawpool(result);
  return result[0] != 0xaabbccdd || result[1] != 0x11223344;
}
`, nil)
			}
		})
	}
}

func TestARM64RawUnlabelledPoolRejectsCodeAndEscapedAddresses(t *testing.T) {
	for _, change := range [][2]string{
		{"0x10000109", "0x14000008"}, // Executable branch into the apparent pool.
		{"0xaa1f03e9", "0xf9000009"}, // Escape the address with STR X9, [X0].
		{"0xd65f03c0", "0xd503201f"}, // Fall through into the words.
	} {
		source := strings.Replace(arm64RawUnlabelledPoolSource, change[0], change[1], 1)
		file, err := Parse(ArchARM64, source)
		if err != nil {
			t.Fatal(err)
		}
		_, _, err = prepareARM64RawPCRelative(file.Funcs[0])
		if err == nil {
			t.Errorf("unsafe pool %v was accepted", change)
		}
	}
}
