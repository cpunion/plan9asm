package plan9asm

import (
	"fmt"
	"runtime"
	"strings"
	"testing"
)

func arm64RawPoolMaskedIR(t *testing.T, triple string) string {
	t.Helper()
	lines := []string{"adr x9,#0", "mov x2,#15", "mov x3,#-16", "mov x6,#16", "mov x7,#19"}
	forms := []string{"and x4,x1,x2", "and x4,x2,x1", "bic x4,x1,x3", "and x4,x1,#15"}
	for mode, limits := range [][2]string{{"#16", "#19"}, {"x6", "x7"}} {
		for index, form := range forms {
			lines = append(lines, "cmp x1,"+limits[0], "b.lo #32", "cmp x1,"+limits[1], "b.hi #24", form,
				"sub x5,x1,x4", "sub x5,x5,#16", "ldr x5,[x9,x5]", fmt.Sprintf("str x5,[x0,#%d]", (mode*len(forms)+index)*8))
		}
	}
	lines = append(lines, "mov x9,xzr", "ret")
	lines[0] = fmt.Sprintf("adr x9,#%d", len(lines)*4)
	var source strings.Builder
	source.WriteString("TEXT pool_masked(SB),$0-16\nMOVD out+0(FP),R0\nMOVD input+8(FP),R1\n")
	for _, word := range assembleARM64LLVMWords(t, lines, "") {
		fmt.Fprintf(&source, "WORD $%#08x\n", word)
	}
	source.WriteString("WORD $0x17b4a140\nWORD $0x17b4a141\nRET\n")
	requireARM64GoAssemblerResult(t, source.String(), true)
	file, err := Parse(ArchARM64, source.String())
	if err != nil {
		t.Fatal(err)
	}
	ir, err := Translate(file, Options{Goarch: "arm64", TargetTriple: triple,
		Sigs: map[string]FuncSig{"pool_masked": {
			Name: "pool_masked", Args: []LLVMType{Ptr, I64}, Ret: Void,
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

func TestARM64RawPoolMaskedLLVM(t *testing.T) {
	llc := findLLVM22Tool("llc")
	if llc == "" {
		t.Fatal("LLVM 22 llc not found")
	}
	for _, triple := range []string{"aarch64-apple-darwin", "aarch64-unknown-linux-gnu", "aarch64-pc-windows-msvc"} {
		t.Run(triple, func(t *testing.T) {
			ir := arm64RawPoolMaskedIR(t, triple)
			compileLLVMToObject(t, llc, triple, "pool_masked.ll", "pool_masked.o", ir)
			if runtime.GOOS == "darwin" && runtime.GOARCH == "arm64" && triple == "aarch64-apple-darwin" {
				clang := findLLVM22Tool("clang")
				if clang == "" {
					t.Fatal("LLVM 22 clang not found")
				}
				compileAndRunRuntimeTestForTarget(t, llc, clang, "pool_masked", triple, ir, arm64RawPoolMaskedMain, nil)
			}
		})
	}
}

const arm64RawPoolMaskedMain = `
#include <stdint.h>
extern void pool_masked(uint64_t *, uint64_t);
int main(void) {
  const uint64_t inputs[] = {0, 1, 15, 16, 17, 18, 19, 20, 31, 255, UINT64_C(1)<<63, UINT64_MAX};
  for (unsigned test = 0; test < sizeof(inputs)/sizeof(inputs[0]); test++) {
    uint64_t out[10] = {0x1234, 0, 0, 0, 0, 0, 0, 0, 0, 0x5678};
    uint64_t n = inputs[test];
    pool_masked(out+1, n);
    if (out[0] != 0x1234 || out[9] != 0x5678) return 1;
    uint64_t expected = n >= 16 && n <= 19 ? UINT64_C(0x17b4a14117b4a140) : 0;
    for (unsigned form = 0; form < 8; form++) {
      if (out[form+1] != expected) return 2;
    }
  }
  return 0;
}
`
