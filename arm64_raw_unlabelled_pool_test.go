package plan9asm

import (
	"fmt"
	"runtime"
	"strings"
	"testing"
)

func arm64RawPoolControlFlowIR(t *testing.T, triple string) string {
	words := assembleARM64LLVMWords(t, []string{
		"adr x9, #52", "cbz x3, #32", "ldr w1, [x9]",
		"add x4, x3, x3, lsl #2", "sub x3, x3, #1", "cbnz x3, #-12",
		"mov x9, xzr", "str w1, [x0]", "ret",
		"ldr w1, [x9, #4]", "mov x9, xzr", "str w1, [x0]", "ret",
	}, "")
	words = append(words, 0x11223344, 0xaabbccdd)
	var source strings.Builder
	source.WriteString("TEXT branchpool(SB),$0-16\nMOVD out+0(FP),R0\nMOVD count+8(FP),R3\n")
	for _, word := range words {
		fmt.Fprintf(&source, "WORD $%#08x\n", word)
	}
	source.WriteString("RET\n")
	file, err := Parse(ArchARM64, source.String())
	if err != nil {
		t.Fatal(err)
	}
	ir, err := Translate(file, Options{Goarch: "arm64", TargetTriple: triple,
		Sigs: map[string]FuncSig{"branchpool": {Name: "branchpool", Args: []LLVMType{Ptr, I64}, Ret: Void,
			Frame: FrameLayout{Params: []FrameSlot{
				{Offset: 0, Type: Ptr, Index: 0, Field: -1},
				{Offset: 8, Type: I64, Index: 1, Field: -1},
			}}}}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(ir, "private constant [8 x i8]") {
		t.Fatal("missing proven data pool")
	}
	return ir
}

const arm64RawPoolControlFlowMain = `
#include <stdint.h>
extern void branchpool(uint32_t *, uint64_t);
int main(void) {
  for (unsigned count = 0; count < 5; count++) {
    uint32_t got = 0;
    branchpool(&got, count);
    if (got != (count ? 0x11223344U : 0xaabbccddU)) return 1;
  }
  return 0;
}
`

func TestARM64RawPoolControlFlowLLVM(t *testing.T) {
	llc, clang, ok := findLlcAndClang(t)
	if !ok {
		t.Fatal("LLVM 22 llc/clang not found")
	}
	for _, triple := range []string{"aarch64-apple-darwin", "aarch64-unknown-linux-gnu", "aarch64-pc-windows-msvc"} {
		t.Run(triple, func(t *testing.T) {
			ir := arm64RawPoolControlFlowIR(t, triple)
			compileLLVMToObject(t, llc, triple, "branchpool.ll", "branchpool.o", ir)
			native := runtime.GOOS == "darwin" && triple == "aarch64-apple-darwin" ||
				runtime.GOOS == "linux" && triple == "aarch64-unknown-linux-gnu"
			if runtime.GOARCH == "arm64" && native {
				compileAndRunRuntimeTestForTarget(t, llc, clang, "branchpool", triple, ir, arm64RawPoolControlFlowMain, nil)
			}
		})
	}
}

func TestARM64RawPoolAddressAcrossControlFlow(t *testing.T) {
	lines := []string{
		"adr x9, #48", "cbz x3, #28",
		"ldr w1, [x9]", "add x4, x4, x5, lsl #2", "sub x3, x3, #1", "cbnz x3, #-12",
		"mov x9, xzr", "ret",
		"ldr w1, [x9, #4]", "ldr w2, [x0, x4, lsl #2]", "mov x9, xzr", "ret",
	}
	for _, test := range []struct {
		name, old, replacement string
		want                   bool
	}{
		{name: "loop-and-diamond", want: true},
		{"escape-one-edge", "mov x9, xzr", "str x9, [x0]", false},
		{"shifted-address", "add x4, x4, x5, lsl #2", "add x4, x4, x9, lsl #2", false},
		{"indexed-address", "ldr w2, [x0, x4, lsl #2]", "ldr w2, [x0, x9, lsl #2]", false},
		{"call-with-live-address", "add x4, x4, x5, lsl #2", "blr x5", false},
	} {
		t.Run(test.name, func(t *testing.T) {
			asm := strings.Join(lines, "\n")
			if test.old != "" {
				asm = strings.Replace(asm, test.old, test.replacement, 1)
			}
			words := assembleARM64LLVMWords(t, strings.Split(asm, "\n"), "")
			var instructions []Instr
			for _, word := range words {
				instructions = append(instructions, Instr{Op: OpWORD, Args: []Operand{{Kind: OpImm, Imm: int64(word)}}})
			}
			if got := arm64RawAddressOnlyLoaded(instructions, 0, len(instructions)); got != test.want {
				t.Fatalf("load-only proof = %v, want %v", got, test.want)
			}
		})
	}
}

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

func arm64RawVoidPoolIR(t *testing.T, triple string) string {
	t.Helper()
	source := strings.ReplaceAll(arm64RawUnlabelledPoolSource, "rawpool", "voidpool")
	source = strings.Replace(source, "0xaa1f03e9", "0xd503201f", 1)
	source = strings.Replace(source, "0x9100800a", "0xd503201f", 1)
	file, err := Parse(ArchARM64, source)
	if err != nil {
		t.Fatal(err)
	}
	sig := FuncSig{Name: "voidpool", Args: []LLVMType{Ptr}, Ret: Void,
		Frame: FrameLayout{Params: []FrameSlot{{Offset: 0, Type: Ptr, Index: 0, Field: -1}}}}
	ir, err := Translate(file, Options{Goarch: "arm64", TargetTriple: triple, Sigs: map[string]FuncSig{"voidpool": sig}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(ir, "private constant [12 x i8]") {
		t.Fatal("missing proven data pool in a void function")
	}
	return ir
}

const arm64RawVoidPoolMain = `
#include <stdint.h>
extern void voidpool(uint32_t *);
int main(void) {
  uint32_t result[2] = {0};
  voidpool(result);
  return result[0] != 0xaabbccdd || result[1] != 0x11223344;
}
`

func TestARM64RawPoolExplicitVoidSignature(t *testing.T) {
	llc, clang, ok := findLlcAndClang(t)
	if !ok {
		t.Fatal("LLVM 22 llc/clang not found")
	}
	for _, triple := range []string{"aarch64-apple-darwin", "aarch64-unknown-linux-gnu", "aarch64-pc-windows-msvc"} {
		t.Run(triple, func(t *testing.T) {
			ir := arm64RawVoidPoolIR(t, triple)
			compileLLVMToObject(t, llc, triple, "voidpool.ll", "voidpool.o", ir)
			native := runtime.GOOS == "darwin" && triple == "aarch64-apple-darwin" ||
				runtime.GOOS == "linux" && triple == "aarch64-unknown-linux-gnu"
			if runtime.GOARCH == "arm64" && native {
				compileAndRunRuntimeTestForTarget(t, llc, clang, "voidpool", triple, ir, arm64RawVoidPoolMain, nil)
			}
		})
	}
}

func TestARM64RawPoolVoidReturnDoesNotHideEscapes(t *testing.T) {
	for _, test := range []struct {
		name   string
		ret    LLVMType
		custom bool
		frame  bool
		old    string
		word   string
	}{
		{name: "unknown-signature", ret: Void},
		{name: "register-result", ret: I64, frame: true},
		{name: "custom-register-ABI", ret: Void, custom: true, frame: true},
		{name: "stored-address", ret: Void, frame: true, old: "0xd503201f", word: "0xf9000009"},
		{name: "indirect-call", ret: Void, frame: true, old: "0xd503201f", word: "0xd63f00a0"},
		{name: "return-through-address", ret: Void, frame: true, old: "0xd65f03c0", word: "0xd65f0120"},
	} {
		t.Run(test.name, func(t *testing.T) {
			source := strings.Replace(arm64RawUnlabelledPoolSource, "0xaa1f03e9", "0xd503201f", 1)
			if test.old != "" {
				source = strings.Replace(source, test.old, test.word, 1)
			}
			file, err := Parse(ArchARM64, source)
			if err != nil {
				t.Fatal(err)
			}
			sig := FuncSig{Name: "rawpool", Args: []LLVMType{Ptr}, Ret: test.ret}
			if test.frame {
				sig.Frame.Params = []FrameSlot{{Offset: 0, Type: Ptr, Index: 0, Field: -1}}
			}
			if test.custom {
				sig.ArgRegs = []Reg{"R0"}
			}
			if _, err := Translate(file, Options{Goarch: "arm64", TargetTriple: "aarch64-unknown-linux-gnu", Sigs: map[string]FuncSig{"rawpool": sig}}); err == nil {
				t.Fatal("accepted an unproven or escaping pool address")
			}
		})
	}
	words := assembleARM64LLVMWords(t, []string{"adr x19, #12", "ldr x1, [x19]", "ret"}, "")
	var instructions []Instr
	for _, word := range words {
		instructions = append(instructions, Instr{Op: OpWORD, Args: []Operand{{Kind: OpImm, Imm: int64(word)}}})
	}
	if arm64RawAddressOnlyLoadedWithExit(instructions, 0, len(instructions), (1<<18)-1) {
		t.Fatal("callee-preserved register must not become a terminal scratch kill")
	}
}

func TestARM64RawPoolLoadKillsAndPostIndexEffects(t *testing.T) {
	for _, test := range []struct {
		instruction string
		want        bool
	}{
		{"ldr x9, [x0]", true}, {"ldr w9, [x0]", true},
		{"ldrb w9, [x0]", true}, {"ldrh w9, [x0]", true},
		{"ldrsb x9, [x0]", true}, {"ldrsb w9, [x0]", true},
		{"ldrsh x9, [x0]", true}, {"ldrsh w9, [x0]", true}, {"ldrsw x9, [x0]", true},
		{"ldur x9, [x0]", true}, {"ldur w9, [x0]", true},
		{"ldurb w9, [x0]", true}, {"ldurh w9, [x0]", true},
		{"ldursb x9, [x0]", true}, {"ldursb w9, [x0]", true},
		{"ldursh x9, [x0]", true}, {"ldursh w9, [x0]", true}, {"ldursw x9, [x0]", true},
		{"csel x9, x0, x1, eq", true}, {"csinc x9, x0, x1, eq", true},
		{"csinv x9, x0, x1, eq", true}, {"csneg x9, x0, x1, eq", true},
		{"ldr x9, [x9]", true},
		{"ld1 {v0.16b}, [x0], x1\nmov x9, xzr", true},
		{"ld1 {v0.16b}, [x0], x9\nmov x9, xzr", false},
		{"ld1 {v0.16b}, [x9], x0\nmov x9, xzr", false},
		{"csel x9, x0, x9, eq", false}, {"csinc x9, x9, x1, eq", false},
		{"str x9, [x0]\nmov x9, xzr", false},
	} {
		t.Run(test.instruction, func(t *testing.T) {
			lines := []string{"adr x9, #64", "ldr w1, [x9]"}
			lines = append(lines, strings.Split(test.instruction, "\n")...)
			lines = append(lines, "ret")
			var instructions []Instr
			for _, word := range assembleARM64LLVMWords(t, lines, "") {
				instructions = append(instructions, Instr{Op: OpWORD, Args: []Operand{{Kind: OpImm, Imm: int64(word)}}})
			}
			if got := arm64RawAddressOnlyLoaded(instructions, 0, len(instructions)); got != test.want {
				t.Fatalf("load-only proof=%v, want %v", got, test.want)
			}
		})
	}
}
