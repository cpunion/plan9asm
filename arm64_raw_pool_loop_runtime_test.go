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
		"adr x9, #0", "mov x4, xzr", "cmp x1, #7", "b.hi #0", "cbz x1, #0",
		"ldr x3, [x9, x1, lsl #3]", "add x4, x4, x3", "subs x1, x1, #1",
	}
	for i := 0; i < 12; i++ {
		lines = append(lines, "ext v0.16b, v1.16b, v2.16b, #8", "uxtl v3.4s, v4.4h",
			"cmhi v5.8b, v6.8b, v7.8b", "mul x2, xzr, xzr")
	}
	lines = append(lines, fmt.Sprintf("b.ne #%d", (5-len(lines))*4))
	lines[3] = fmt.Sprintf("b.hi #%d", (len(lines)-3)*4)
	lines[4] = fmt.Sprintf("cbz x1, #%d", (len(lines)-4)*4)
	lines = append(lines,
		"str x4, [x0]",

		"add x10, x9, #32", "mov x2, #-8",
		"ldp q0, q1, [x10], #-64", "add x2, x2, #8", "cbnz x2, #-8",
		"stp q0, q1, [x0, #16]",

		"mov x10, x9", "mov x2, #16",
		"ldp q2, q3, [x10, #16]!", "sub x2, x2, #16", "cbnz x2, #-8",
		"stp q2, q3, [x0, #48]",

		"add x10, x9, #24", "ldr x2, [x10], #8", "ldr x3, [x10, #8]!",
		"stp x2, x3, [x0, #80]",

		"mov x10, x9", "cbz x1, #8", "adr x10, #0",
		"sub x12, x10, x2", "add x12, x12, x2", "ldr x3, [x12]", "str x3, [x0, #96]",
	)
	output := 104
	for _, op := range []string{"orr", "eor"} {
		for _, source := range []struct {
			operand string
			value   int
		}{
			{"x4", 16}, {"x4, lsl #4", 1}, {"x4, lsr #1", 32},
			{"x4, asr #1", 32}, {"x4, ror #63", 8}, {"#16", 0},
		} {
			lines = append(lines, "and x2, x1, #1", fmt.Sprintf("mov x4, #%d", source.value),
				fmt.Sprintf("%s x3, x2, %s", op, source.operand),
				"ldrb w3, [x9, x3]", fmt.Sprintf("str x3, [x0, #%d]", output))
			output += 8
		}
	}
	for _, branch := range []struct{ up, down string }{
		{"lt", "gt"}, {"lo", "hi"}, {"le", "ge"}, {"ls", "hs"},
	} {
		for _, ascending := range []bool{true, false} {
			for _, reversed := range []bool{false, true} {
				initial, difference, update := 8, "sub x5, x3, x2", "add x2, x2, #1"
				condition, compare := branch.up, "cmp x2, x3"
				if !ascending {
					initial, difference, update = 24, "sub x5, x2, x3", "sub x2, x2, #1"
					condition = branch.down
				}
				if reversed {
					compare = "cmp x3, x2"
					condition = branch.down
					if !ascending {
						condition = branch.up
					}
				}
				lines = append(lines, "and x2, x1, #7", fmt.Sprintf("add x2, x2, #%d", initial),
					"mov x3, #16", "mov x4, xzr")
				head := len(lines)
				lines = append(lines, difference, "ldrb w5, [x9, x5]", "add x4, x4, x5",
					update, compare, "orr w6, w6, w7")
				lines = append(lines, fmt.Sprintf("b.%s #%d", condition, (head-len(lines))*4),
					fmt.Sprintf("str x4, [x0, #%d]", output))
				output += 8
			}
		}
	}
	lines = append(lines, "mov x12, xzr", "mov x9, xzr", "mov x10, xzr", "ret")
	lines[0] = fmt.Sprintf("adr x9, #%d", len(lines)*4)
	for at, line := range lines {
		if line == "adr x10, #0" {
			lines[at] = fmt.Sprintf("adr x10, #%d", (len(lines)-at)*4+8)
		}
	}
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
    uint64_t result[43] = {0}, expected[41] = {0};
    result[0] = 0x12345678;
    result[42] = 0x87654321;
    if (counts[i] <= 7) {
      for (uint64_t n = 1; n <= counts[i]; n++) expected[0] += values[n];
    }
    memcpy(expected + 2, values + 4, 32);
    memcpy(expected + 6, values + 2, 32);
    expected[10] = values[3];
    expected[11] = values[5];
    expected[12] = values[counts[i] > 7 ? 1 : 0];
    const uint8_t *bytes = (const uint8_t *)words;
    uint64_t remaining = counts[i] > 7 ? counts[i] : 0;
    for (unsigned j = 13; j < 25; j++) expected[j] = bytes[16 + (remaining & 1)];
    unsigned output = 25;
    for (unsigned kind = 0; kind < 4; kind++) {
      for (unsigned direction = 0; direction < 2; direction++) {
        for (unsigned reversed = 0; reversed < 2; reversed++) {
          unsigned distance = direction ? 8 + (remaining & 7) : 8 - (remaining & 7);
          for (unsigned n = kind < 2 ? 1 : 0; n <= distance; n++) expected[output] += bytes[n];
          output++;
        }
      }
    }
    pool_loop(result + 1, counts[i]);
    if (result[0] != 0x12345678 || result[42] != 0x87654321 ||
        memcmp(result + 1, expected, sizeof(expected)) != 0) return 1;
  }
  return 0;
}
`
