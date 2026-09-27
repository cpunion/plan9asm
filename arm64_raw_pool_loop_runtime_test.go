package plan9asm

import (
	"fmt"
	"runtime"
	"strings"
	"testing"
)

func arm64RawPoolLoopIR(t *testing.T, triple string) string {
	t.Helper()
	lines := []string{
		"adr x9, #0", "mov x4, xzr", "cmp x1, #7", "b.hi #24", "cbz x1, #20",
		"ldr x3, [x9, x1, lsl #3]", "add x4, x4, x3", "subs x1, x1, #1", "b.ne #-12",
		"str x4, [x0]",

		"add x10, x9, #32", "mov x2, #-8",
		"ldp q0, q1, [x10], #-64", "add x2, x2, #8", "cbnz x2, #-8",
		"stp q0, q1, [x0, #16]",

		"mov x10, x9", "mov x2, #16",
		"ldp q2, q3, [x10, #16]!", "sub x2, x2, #16", "cbnz x2, #-8",
		"stp q2, q3, [x0, #48]",

		"add x10, x9, #24", "ldr x2, [x10], #8", "ldr x3, [x10, #8]!",
		"stp x2, x3, [x0, #80]", "mov x9, xzr", "mov x10, xzr", "ret",
	}
	lines[0] = fmt.Sprintf("adr x9, #%d", len(lines)*4)
	var source strings.Builder
	source.WriteString("TEXT pool_loop(SB),$0-16\nMOVD out+0(FP),R0\nMOVD count+8(FP),R1\n")
	for _, word := range assembleARM64LLVMWords(t, lines, "") {
		fmt.Fprintf(&source, "WORD $%#08x\n", word)
	}
	for i := 0; i < 16; i++ {
		fmt.Fprintf(&source, "WORD $%#08x\n", uint32(0x17b4a140+i))
	}
	source.WriteString("RET\n")
	requireARM64GoAssemblerResult(t, source.String(), true)
	file, err := Parse(ArchARM64, source.String())
	if err != nil {
		t.Fatal(err)
	}
	ir, err := Translate(file, Options{Goarch: "arm64", TargetTriple: triple,
		Sigs: map[string]FuncSig{"pool_loop": {
			Name: "pool_loop", Args: []LLVMType{Ptr, I64}, Ret: Void,
			Frame: FrameLayout{Params: []FrameSlot{
				{Offset: 0, Type: Ptr, Index: 0, Field: -1},
				{Offset: 8, Type: I64, Index: 1, Field: -1},
			}},
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	return ir
}

func TestARM64RawPoolLoopLLVM(t *testing.T) {
	llc := findLLVM22Tool("llc")
	if llc == "" {
		t.Fatal("LLVM 22 llc not found")
	}
	for _, triple := range []string{"aarch64-apple-darwin", "aarch64-unknown-linux-gnu", "aarch64-pc-windows-msvc"} {
		t.Run(triple, func(t *testing.T) {
			ir := arm64RawPoolLoopIR(t, triple)
			compileLLVMToObject(t, llc, triple, "pool_loop.ll", "pool_loop.o", ir)
			if runtime.GOARCH == "arm64" && runtime.GOOS == "darwin" && triple == "aarch64-apple-darwin" {
				clang := findLLVM22Tool("clang")
				if clang == "" {
					t.Fatal("LLVM 22 clang not found")
				}
				compileAndRunRuntimeTestForTarget(t, llc, clang, "pool_loop", triple, ir, arm64RawPoolLoopMain, nil)
			}
		})
	}
}

const arm64RawPoolLoopMain = `
#include <stdint.h>
#include <string.h>
extern void pool_loop(uint64_t *, uint64_t);
int main(void) {
  uint32_t words[16];
  for (unsigned i = 0; i < 16; i++) words[i] = 0x17b4a140 + i;
  uint64_t values[8];
  memcpy(values, words, sizeof(values));
  const uint64_t counts[] = {0, 1, 2, 3, 4, 5, 6, 7, 8, 16, 19, 20, 255,
                             1ULL << 32, 1ULL << 63, UINT64_MAX};
  for (unsigned i = 0; i < sizeof(counts) / sizeof(counts[0]); i++) {
    uint64_t result[14] = {0}, expected[12] = {0};
    result[0] = 0x12345678;
    result[13] = 0x87654321;
    if (counts[i] <= 7) {
      for (uint64_t n = 1; n <= counts[i]; n++) expected[0] += values[n];
    }
    memcpy(expected + 2, values + 4, 32);
    memcpy(expected + 6, values + 2, 32);
    expected[10] = values[3];
    expected[11] = values[5];
    pool_loop(result + 1, counts[i]);
    if (result[0] != 0x12345678 || result[13] != 0x87654321 ||
        memcmp(result + 1, expected, sizeof(expected)) != 0) return 1;
  }
  return 0;
}
`
