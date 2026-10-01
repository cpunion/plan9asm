package plan9asm

import (
	"errors"
	"strings"
	"testing"
)

func TestARM64InstructionProbeDoesNotRequireDataToPreserveCallerState(t *testing.T) {
	for _, instruction := range []string{
		"LDAXP (R25), (R30, R11)",
		"LDXRW (RSP), R30",
		"NEGSW R23<<1, R30",
		"VLD3.P (R30)(R2), [V14.D2, V15.D2, V16.D2]",
		"MOVD $0x1708(RSP), RSP",
		"MOVD $-0x10000(RSP), RSP",
		"MOVD R2, RSP",
	} {
		t.Run(instruction, func(t *testing.T) {
			// This is a compile-only harness, never an executable success.
			// A real zero-address branch is independent of the tested data
			// instruction's writes to the caller link and stack pointer.
			source := "TEXT isolatedData(SB),516,$0-0\n" + instruction + "\nB (ZR)\n"
			requireARM64GoAssemblerResult(t, source, true)
			file, err := Parse(ArchARM64, source)
			if err != nil {
				t.Fatal(err)
			}
			var tested Instr
			for _, ins := range file.Funcs[0].Instrs {
				if ins.Op != OpTEXT {
					tested = ins
					break
				}
			}
			if err := ProbeInstruction(ArchARM64, "arm64", tested); err != nil {
				t.Fatalf("a data form must not inherit a synthetic RET's caller-state requirement: %v", err)
			}
			if err := ProbeInstructionSequence(ArchARM64, "arm64", []Instr{tested}); err != nil {
				t.Fatalf("a sequence probe must use the same state-independent exit: %v", err)
			}

			llc := findLLVM22Tool("llc")
			if llc == "" {
				t.Fatal("LLVM 22 llc not found")
			}
			for _, triple := range []string{
				"aarch64-apple-darwin", "aarch64-unknown-linux-gnu",
				"aarch64-unknown-linux-musl", "aarch64-pc-windows-msvc",
			} {
				ir, err := Translate(file, Options{Goarch: "arm64", TargetTriple: triple,
					Sigs: map[string]FuncSig{"isolatedData": {Name: "isolatedData", Ret: Void}},
				})
				if err != nil {
					t.Fatal(err)
				}
				if !strings.Contains(ir, `asm sideeffect "br $0", "r,~{memory}"(i64 0)`+"\n  unreachable") {
					t.Fatal("compile-only exit must remain a real faulting terminator, not a fake return")
				}
				compileLLVMToObject(t, llc, triple, "isolated-data.ll", "isolated-data.o", ir)
			}
		})
	}
}

func TestARM64InstructionProbeRetainsNativeControlContracts(t *testing.T) {
	for _, instruction := range []string{
		"BL (R2)", "CALL (R15)", "RET R9", "JMP (R29)",
		"WORD $0xd61f0120", "WORD $0xd65f0120",
	} {
		t.Run(instruction, func(t *testing.T) {
			source := "TEXT isolatedControl(SB),516,$0-0\n" + instruction + "\nB (ZR)\n"
			requireARM64GoAssemblerResult(t, source, true)
			file, err := Parse(ArchARM64, source)
			if err != nil {
				t.Fatal(err)
			}
			sequence := file.Funcs[0].Instrs[1:2]
			if err := ProbeInstructionSequence(ArchARM64, "arm64", sequence); !errors.Is(err, ErrProbeNeedsContext) {
				t.Fatalf("a faulting probe exit cannot establish an unknown native target's ABI: %v", err)
			}
		})
	}
}
