package plan9asm

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestARM64RawWordRejectsVCONBeforeTruncation(t *testing.T) {
	// All low words are real instructions. Reject the original Go constant
	// class before a decoder can turn an invalid source into one of them.
	for _, word := range []int64{0xd503201f, 0x1f0f7ce0, 0x14000000} {
		for _, value := range []int64{word + (1 << 32), word - (1 << 33)} {
			line := fmt.Sprintf("WORD $%d", value)
			source := "TEXT comparison(SB),516,$0-0\n" + line + "\nRET\n"
			requireARM64GoAssemblerResult(t, source, false)
			instruction := arm64ImmediateTestInstruction(t, line)
			if err := arm64RawWordPreflightError(instruction); err == nil || errors.Is(err, ErrProbeNeedsContext) {
				t.Errorf("Go-rejected C_VCON became a native instruction: %s: %v", line, err)
			}
			if _, err := decodeARM64RawWordInstructionTarget(instruction, "target"); err == nil || errors.Is(err, ErrProbeNeedsContext) {
				t.Errorf("raw decoder truncated Go-rejected source: %s: %v", line, err)
			}
			if word == 0xd503201f {
				context := &arm64Ctx{b: &strings.Builder{}}
				if err := context.lowerRawWord(instruction); err == nil || context.b.Len() != 0 {
					t.Errorf("raw lowering emitted a truncated native NOP: %s: %v", line, err)
				}
			}
			for _, target := range []string{"", "x86_64-unknown-linux-gnu", "aarch64-unknown-linux-gnu"} {
				for _, api := range []string{"text", "module", "Go", "public-module"} {
					for _, dead := range []bool{false, true} {
						t.Run(fmt.Sprintf("%x/%s/%s/dead=%t", value, target, api, dead), func(t *testing.T) {
							body := source
							if dead {
								body = "TEXT comparison(SB),516,$0-0\nRET\n" + line + "\n"
							}
							present, err := arm64FPControlTargetAPI(t, api, body, target)
							if present || err == nil || errors.Is(err, ErrProbeNeedsContext) {
								t.Fatalf("source constant class disappeared before translation: output=%t err=%v", present, err)
							}
						})
					}
				}
			}
		}
	}
}

func TestARM64RawWordWideLCONStillAccepted(t *testing.T) {
	// Go's C_LCON row includes wide MOV/bitmask classes, not only int32 or
	// uint32. These all emit low-word UDF encodings and require no CPU state.
	for _, value := range []int64{0, 0xffff, 1 << 32, 0x1234 << 32, 0x1234 << 48, 0x0000ffff0000ffff} {
		line := fmt.Sprintf("WORD $%d", value)
		source := "TEXT comparison(SB),516,$0-0\n" + line + "\nRET\n"
		requireARM64GoAssemblerResult(t, source, true)
		if err := arm64RawWordPreflightError(arm64ImmediateTestInstruction(t, line)); err != nil {
			t.Fatalf("Go-accepted wide LCON rejected: %s: %v", line, err)
		}
		for _, api := range []string{"text", "module", "Go", "public-module"} {
			present, err := arm64FPControlTargetAPI(t, api, source, "aarch64-unknown-linux-gnu")
			if err != nil || !present {
				t.Fatalf("%s lost Go-accepted wide source: %s: %v", api, line, err)
			}
		}
	}
}

func TestARM64RawWordConstantClassesMatchGo(t *testing.T) {
	accepted := []int64{0, 1, 0xfff, 0x123000, 0xfff000, 0xffffff, 0x11223344,
		-0x7fffffff, 0xffffffff, 0x00ff00ff00ff00ff, 0x5555555555555555, 0x7fffffffffffffff, -1 << 63}
	for _, shift := range []uint{0, 16, 32, 48} {
		value := int64(uint64(0x1234) << shift)
		accepted = append(accepted, value, ^value)
	}
	for _, group := range []struct {
		accepted bool
		values   []int64
	}{
		{true, accepted},
		{false, []int64{0x1122334455667788, -0x1122334455667788, 0x1d503201f, -0x12afcdfe1}},
	} {
		for _, value := range group.values {
			line := fmt.Sprintf("WORD $%d", value)
			requireARM64GoAssemblerResult(t, "TEXT original(SB),516,$0-0\n"+line+"\nRET\n", group.accepted)
			word, err := arm64RawWordValue(arm64ImmediateTestInstruction(t, line))
			if group.accepted {
				if err != nil || word != uint32(value) {
					t.Errorf("Go-accepted constant lost its exact low word: %s: %#x, %v", line, word, err)
				}
			} else if err == nil || errors.Is(err, ErrProbeNeedsContext) {
				t.Errorf("Go-rejected constant acquired a decoder word: %s: %#x, %v", line, word, err)
			}
		}
	}
}
