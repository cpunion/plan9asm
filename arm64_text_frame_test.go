package plan9asm

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

func TestARM64TextFrameAllocation(t *testing.T) {
	for _, test := range []struct {
		name             string
		frame, low, high int64
		body             string
	}{
		{"zero", 0, 0, 0, ""},
		{"no-frame", -8, 0, 0, ""},
		{"small", 16, 0, 16, ""},
		{"snappy-table", 32904, 0, 32904, "ADD $128,RSP,R17\nMOVD R0,(R17)\n"},
		{"negative-sp", 32904, -512, 32904, "SUB $512,RSP\nMOVD R0,(RSP)\nADD $512,RSP\n"},
		{"positive-sp", 32904, 0, 32904 + 512, "ADD $512,RSP\nMOVD R0,(RSP)\nSUB $512,RSP\n"},
		{"explicit-beyond-frame", 16, 0, 4160, "MOVD R0,4096(RSP)\n"},
	} {
		t.Run(test.name, func(t *testing.T) {
			source := fmt.Sprintf("TEXT stack_bounds(SB),$%d-0\nMOVD RSP,R9\n%sRET\n", test.frame, test.body)
			requireARM64GoAssemblerResult(t, source, true)
			ir := arm64StackTestIR(t, source)
			match := regexp.MustCompile(`getelementptr inbounds \[(\d+) x i8\], ptr %local_stack, i32 0, i64 (\d+)`).FindStringSubmatch(ir)
			if len(match) != 3 {
				t.Fatal("missing local stack allocation")
			}
			size, _ := strconv.ParseInt(match[1], 10, 64)
			bias, _ := strconv.ParseInt(match[2], 10, 64)
			if bias+test.low < 0 || bias+test.high > size {
				t.Fatalf("allocation [%d,%d) does not contain declared frame/accesses [%d,%d)", -bias, size-bias, test.low, test.high)
			}
		})
	}
}

func TestARM64TextFrameRejectsInvalidSize(t *testing.T) {
	for _, frame := range []int64{-16, -1, arm64MaxLocalStackSpan + 8, 1<<63 - 1} {
		t.Run(strconv.FormatInt(frame, 10), func(t *testing.T) {
			file, err := Parse(ArchARM64, fmt.Sprintf("TEXT frame(SB),$%d-0\nMOVD RSP,R9\nRET\n", frame))
			if err != nil {
				t.Fatal(err)
			}
			_, err = Translate(file, Options{Goarch: "arm64", Sigs: map[string]FuncSig{
				"frame": {Name: "frame", Ret: Void},
			}})
			if err == nil {
				t.Fatalf("invalid or oversized TEXT frame %d accepted", frame)
			}
		})
	}
}

const arm64TextFrameSource = `TEXT frame_table(SB),$32904-16
MOVD value+0(FP),R0
MOVD R0,64(RSP)
ADD $128,RSP,R17
ADD $32768,R17,R6
MOVD R0,(R6)
MOVD $32768,R4
fill:
MOVD R0,(R17)
ADD $8,R17
SUB $8,R4
CBNZ R4,fill
ADD $128,RSP,R17
MOVD $32768,R4
BL frame_observe(SB)
MOVD $0,R5
check:
MOVD (R17),R1
EOR R0,R1
ORR R1,R5
ADD $8,R17
SUB $8,R4
CBNZ R4,check
MOVD 64(RSP),R1
EOR R0,R1
ORR R1,R5
MOVD (R6),R1
EOR R0,R1
ORR R1,R5
MOVD R5,ret+8(FP)
RET
`

func arm64TextFrameRuntime(t *testing.T, triple string) (string, string) {
	t.Helper()
	requireARM64GoAssemblerResult(t, arm64TextFrameSource, true)
	file, err := Parse(ArchARM64, arm64TextFrameSource)
	if err != nil {
		t.Fatal(err)
	}
	ir, err := Translate(file, Options{Goarch: "arm64", TargetTriple: triple,
		Sigs: map[string]FuncSig{"frame_table": {
			Name: "frame_table", Args: []LLVMType{I64}, Ret: I64,
			Frame: FrameLayout{
				Params:  []FrameSlot{{Offset: 0, Type: I64, Index: 0, Field: -1}},
				Results: []FrameSlot{{Offset: 8, Type: I64, Index: 0, Field: -1}},
			},
		}, "frame_observe": {
			Name: "frame_observe", Args: []LLVMType{Ptr, I64, I64}, Ret: Void,
			ArgRegs: []Reg{"R17", "R4", "R0"},
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	const main = `
#include <stdint.h>
#include <stdio.h>
extern uint64_t frame_table(uint64_t);
void frame_observe(volatile uint64_t* data, uint64_t bytes, uint64_t value) {
  for (uint64_t i = 0; i < bytes / sizeof(*data); i++) {
    if (data[i] != value) {
      fprintf(stderr, "frame observation mismatch at %llu\n", (unsigned long long)i);
      __builtin_trap();
    }
  }
}
int main(void) {
  volatile uint64_t canary[2] = {UINT64_C(0x123456789abcdef0), UINT64_C(0xfedcba9876543210)};
  uint64_t value = UINT64_C(0x89abcdef01234567);
  for (unsigned i = 0; i < 128; i++) {
    uint64_t got = frame_table(value);
    if (got != 0) {
      fprintf(stderr, "frame table mismatch: %llx\n", (unsigned long long)got);
      return 1;
    }
    if (canary[0] != UINT64_C(0x123456789abcdef0) || canary[1] != UINT64_C(0xfedcba9876543210)) return 2;
    value = value * UINT64_C(6364136223846793005) + 1;
  }
  return 0;
}
`
	return ir, main
}

func TestARM64TextFrameLLVM(t *testing.T) {
	llc := findLLVM22Tool("llc")
	if llc == "" {
		t.Fatal("LLVM 22 llc not found")
	}
	for _, triple := range []string{
		"aarch64-apple-darwin", "aarch64-unknown-linux-gnu",
		"aarch64-unknown-freebsd", "aarch64-pc-windows-msvc",
	} {
		t.Run(triple, func(t *testing.T) {
			ir, main := arm64TextFrameRuntime(t, triple)
			compileLLVMToObject(t, llc, triple, "frame.ll", "frame.o", ir)
			native := runtime.GOARCH == "arm64" && ((runtime.GOOS == "darwin" && strings.Contains(triple, "apple")) ||
				(runtime.GOOS == "linux" && strings.Contains(triple, "linux")) ||
				(runtime.GOOS == "freebsd" && strings.Contains(triple, "freebsd")))
			if native {
				clang := findLLVM22Tool("clang")
				if clang == "" {
					t.Fatal("LLVM 22 clang not found")
				}
				compileAndRunRuntimeTestForTarget(t, llc, clang, "frame", triple, ir, main, nil)
			}
		})
	}
}

func TestARM64TextFrameNativeGo(t *testing.T) {
	cross := runtime.GOOS == "linux" && runtime.GOARCH == "amd64" && os.Getenv("PLAN9ASM_CROSS_EXEC") == "1"
	if runtime.GOARCH != "arm64" && !cross {
		t.Skip("native Go oracle runs on arm64 or required Linux/QEMU")
	}
	const main = `package main
func frame_table(value uint64) uint64
func main() {
  value := uint64(0x89abcdef01234567)
  for i := 0; i < 128; i++ {
    if got := frame_table(value); got != 0 { panic("declared frame table mismatch") }
    value = value * 6364136223846793005 + 1
  }
}
`
	dir := t.TempDir()
	for name, data := range map[string]string{
		"go.mod":        "module frameoracle\n\ngo 1.20\n",
		"main.go":       main,
		"frame_arm64.s": strings.Replace(strings.ReplaceAll(arm64TextFrameSource, "BL frame_observe(SB)\n", ""), "TEXT frame_table", "TEXT ·frame_table", 1),
	} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(data), 0600); err != nil {
			t.Fatal(err)
		}
	}
	args := []string{"run", "."}
	if cross {
		if _, err := exec.LookPath("qemu-aarch64"); err != nil {
			t.Fatal(err)
		}
		args = []string{"run", "-exec=qemu-aarch64", "."}
	}
	cmd := exec.Command("go", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GOWORK=off", "GOARCH=arm64", "CGO_ENABLED=0")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Go TEXT frame oracle: %v\n%s", err, out)
	}
}
