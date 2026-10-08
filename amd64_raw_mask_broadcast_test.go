package plan9asm

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

// This physical directory is independent of the decoder and is also checked
// against actual cmd/asm output below. K is a data operand, not a writemask.
func x86MaskBroadcastPhysicalBytes(op string, width, source, destination int) []byte {
	p0 := byte(0xf2 ^ ((destination >> 3 & 1) << 7) ^ ((destination >> 4 & 1) << 4))
	p1, opcode := byte(0x7e), byte(0x3a)
	if op == "VPBROADCASTMB2Q" {
		p1, opcode = 0xfe, 0x2a
	}
	return []byte{0x62, p0, p1, byte(width<<5 | 8), opcode, byte(0xc0 | destination&7<<3 | source)}
}

func TestX86BroadcastRawMaskDataOriginalGoBytes(t *testing.T) {
	for _, arch := range []string{"amd64", "386"} {
		var source strings.Builder
		source.WriteString("TEXT f(SB),4,$0-0\n")
		var expected []byte
		for _, op := range []string{"VPBROADCASTMB2Q", "VPBROADCASTMW2D"} {
			for width, bank := range []string{"X", "Y", "Z"} {
				for _, input := range []int{0, 7} {
					destinations := []int{0, 7}
					if arch == "amd64" {
						destinations = append(destinations, 31)
					}
					for _, destination := range destinations {
						fmt.Fprintf(&source, "%s K%d,%s%d\n", op, input, bank, destination)
						expected = append(expected, x86MaskBroadcastPhysicalBytes(op, width, input, destination)...)
					}
				}
			}
		}
		source.WriteString("RET\n")
		expected = append(expected, 0xc3)
		actual := assembleX87ControlBytes(t, arch, source.String())
		if !bytes.Equal(actual, expected) {
			t.Fatalf("%s actual Go mask broadcast bytes %x, expected %x", arch, actual, expected)
		}
		for offset := 0; offset < len(actual)-1; offset += 6 {
			mode := 64
			if arch == "386" {
				mode = 32
			}
			ins, length, known, err := decodedX86VEXPackedBroadcastInstruction(actual[offset:offset+6], mode)
			if err != nil || !known || length != 6 {
				t.Fatalf("%s actual Go bytes %x: known%t length%d err%v", arch, actual[offset:offset+6], known, length, err)
			}
			form, recognized, err := x86BroadcastProperties(ins.Op, ins, arch)
			if err != nil || !recognized || form.kind != x86BroadcastMask {
				t.Fatalf("%s physical mask grammar %+v: known%t err%v", arch, form, recognized, err)
			}
			// Decoding an initialized K input is not a new numerical proof.
			for _, consumer := range x86ResultGPSIMDConsumers {
				state := x86ResultState{}
				for index := range state.masks {
					state.masks[index] = true
				}
				if consumer.step(ins, &state, nil, 0, arch) {
					t.Fatalf("%s admitted mask-data numerical authority", consumer.name)
				}
			}
		}
	}
}

func TestX86BroadcastRawMaskDataCompletePhysicalBanks(t *testing.T) {
	count := 0
	for _, op := range []string{"VPBROADCASTMB2Q", "VPBROADCASTMW2D"} {
		for width, bank := range []string{"X", "Y", "Z"} {
			for input := 0; input < 8; input++ {
				for destination := 0; destination < 32; destination++ {
					code := x86MaskBroadcastPhysicalBytes(op, width, input, destination)
					ins, length, known, err := decodedX86VEXPackedBroadcastInstruction(code, 64)
					want := fmt.Sprintf("%s K%d, %s%d", op, input, bank, destination)
					if err != nil || !known || length != 6 || ins.Raw != want {
						t.Fatalf("decode %x: %+v length%d known%t err%v; want %q", code, ins, length, known, err, want)
					}
					_, _, known32, err32 := decodedX86VEXPackedBroadcastInstruction(code, 32)
					if !known32 || (err32 != nil) != (destination >= 8) {
						t.Fatalf("mode32 destination%d: known%t err%v", destination, known32, err32)
					}
					count++
				}
			}
		}
	}
	if count != 1536 {
		t.Fatalf("physical directory covered %d encodings", count)
	}
}

func TestX86BroadcastRawMaskDataReservedAndUnprovedForms(t *testing.T) {
	for _, op := range []string{"VPBROADCASTMB2Q", "VPBROADCASTMW2D"} {
		valid := x86MaskBroadcastPhysicalBytes(op, 0, 0, 0)
		for name, mutate := range map[string]func([]byte){
			"wrong W":            func(code []byte) { code[2] ^= 0x80 },
			"missing fixed bit":  func(code []byte) { code[2] &^= 4 },
			"used vvvv":          func(code []byte) { code[2] &^= 8 },
			"used high vvvv":     func(code []byte) { code[3] &^= 8 },
			"source extension B": func(code []byte) { code[1] &^= 0x20 },
			"source extension X": func(code []byte) { code[1] &^= 0x40 },
			"writemask":          func(code []byte) { code[3] |= 1 },
			"zeroing":            func(code []byte) { code[3] |= 0x80 },
			"broadcast bit":      func(code []byte) { code[3] |= 0x10 },
			"reserved length":    func(code []byte) { code[3] |= 0x60 },
			"memory input":       func(code []byte) { code[5] &^= 0xc0 },
		} {
			code := append([]byte(nil), valid...)
			mutate(code)
			if _, _, known, err := decodedX86VEXPackedBroadcastInstruction(code, 64); !known || err == nil {
				t.Fatalf("%s %s %x: known%t err%v", op, name, code, known, err)
			}
		}
		for _, prefix := range []byte{0x64, 0x65, 0x67} {
			code := append([]byte{prefix}, valid...)
			if _, _, known, err := decodedX86VEXPackedBroadcastInstruction(code, 64); !known || err == nil {
				t.Fatalf("%s forbidden prefix%x: known%t err%v", op, prefix, known, err)
			}
		}
		if _, _, known, err := decodedX86VEXPackedBroadcastInstruction(valid[:5], 64); !known || err == nil {
			t.Fatalf("%s missing ModRM: known%t err%v", op, known, err)
		}
		for length := 0; length < 5; length++ {
			if _, _, known, err := decodedX86VEXPackedBroadcastInstruction(valid[:length], 64); known || err != nil {
				t.Fatalf("%s incomplete prefix length%d: known%t err%v", op, length, known, err)
			}
		}
		for name, mutate := range map[string]func([]byte){
			"other map":    func(code []byte) { code[1] ^= 1 },
			"other prefix": func(code []byte) { code[2] ^= 1 },
			"other opcode": func(code []byte) { code[4] ^= 1 },
		} {
			code := append([]byte(nil), valid...)
			mutate(code)
			if _, _, known, err := decodedX86VEXPackedBroadcastInstruction(code, 64); known || err != nil {
				t.Fatalf("%s %s outside grammar: known%t err%v", op, name, known, err)
			}
		}
	}
}
