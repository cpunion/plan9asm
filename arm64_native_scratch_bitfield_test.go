package plan9asm

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/arch/arm64/arm64asm"
)

func TestARM64NativeScratchBitfieldGoEncodingRows(t *testing.T) {
	// Independent source witnesses follow Go's obj7 cases 42, 43 and 45,
	// plus the immediate shifts encoded as bitfields by case 8. The expected
	// effects come from each source operation, not the translator decoder.
	type witness struct {
		name, instruction string
		merge             bool
	}
	var witnesses []witness
	for _, op := range []string{"SBFM", "BFM", "UBFM"} {
		for _, suffix := range []string{"", "W"} {
			witnesses = append(witnesses, witness{
				op + suffix, op + suffix + " $1,R1,$17,R4", op == "BFM",
			})
		}
	}
	for _, op := range []string{"BFI", "BFXIL", "SBFIZ", "SBFX", "UBFIZ", "UBFX"} {
		for _, suffix := range []string{"", "W"} {
			witnesses = append(witnesses, witness{
				op + suffix, op + suffix + " $3,R1,$7,R4", op == "BFI" || op == "BFXIL",
			})
		}
	}
	for _, op := range []string{
		"SXTB", "SXTBW", "SXTH", "SXTHW", "SXTW",
		"UXTB", "UXTBW", "UXTH", "UXTHW", "UXTW",
		"MOVB", "MOVH", "MOVW", "MOVBU", "MOVHU", "MOVWU",
	} {
		witnesses = append(witnesses, witness{op, op + " R1,R4", false})
	}
	for _, op := range []string{"ASR", "LSL", "LSR"} {
		for _, suffix := range []string{"", "W"} {
			witnesses = append(witnesses, witness{op + suffix, op + suffix + " $3,R1,R4", false})
		}
	}
	var assembly strings.Builder
	for _, witness := range witnesses {
		fmt.Fprintf(&assembly, "TEXT %s(SB),516,$0-0\n%s\nRET\n", witness.name, witness.instruction)
	}
	directory := t.TempDir()
	input, object := filepath.Join(directory, "bitfields_arm64.s"), filepath.Join(directory, "original.o")
	if err := os.WriteFile(input, []byte(assembly.String()), 0600); err != nil {
		t.Fatal(err)
	}
	listing, err := arm64MemoryAuditGoCommand(t, "linux", "tool", "asm", "-S", "-o", object, input)
	if err != nil {
		t.Fatalf("original Go bitfield family: %v\n%s", err, listing)
	}
	disassembly, err := arm64MemoryAuditGoCommand(t, "linux", "tool", "objdump", object)
	if err != nil {
		t.Fatalf("actual original Go bitfield object: %v\n%s", err, disassembly)
	}
	bodies, err := arm64ScratchObjectBodies(disassembly)
	if err != nil {
		t.Fatal(err)
	}
	for _, witness := range witnesses {
		t.Run(witness.name, func(t *testing.T) {
			body, present := bodies[witness.name]
			if !present || len(body.code) < 2 || body.code[1] != 0xd65f03c0 || len(body.relocs) != 0 {
				t.Fatal("original Go bitfield instruction/return identity missing")
			}
			var bytes [4]byte
			binary.LittleEndian.PutUint32(bytes[:], body.code[0])
			instruction, err := arm64asm.Decode(bytes[:])
			if err != nil {
				t.Fatal(err)
			}
			reads := uint64(1 << 1)
			if witness.merge {
				reads |= 1 << 4
			}
			want := arm64ScratchInputEffect{reads: reads, writes: 1 << 4}
			got, valid := arm64ScratchNativeInputEffect(instruction, false)
			if !valid || got != want {
				t.Fatalf("Go encoding %08x %v: native bitfield effect=%+v registered=%v, want %+v",
					body.code[0], instruction, got, valid, want)
			}
		})
	}
}
