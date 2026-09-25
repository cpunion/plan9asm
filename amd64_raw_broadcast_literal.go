package plan9asm

import (
	"encoding/binary"
	"fmt"
	"strings"
)

type x86RawLiteralRange struct {
	first int
	last  int
}

// decodeX86RawPackedBroadcastLiteral resolves a RIP-relative packed scalar
// broadcast only when its source bytes live in the same raw directive group.
// The caller must still prove that those bytes are not reachable instructions.
// A displacement into another TEXT, a missing constant pool, and segment- or
// address-size-relative loads stay on the rejecting decoder path.
func decodeX86RawPackedBroadcastLiteral(code []byte, offset, mode int) (Instr, int, x86RawLiteralRange, bool, error) {
	if mode != 64 || offset >= len(code) {
		return Instr{}, 0, x86RawLiteralRange{}, false, nil
	}

	// A segment or address-size prefix changes the effective source.
	switch code[offset] {
	case 0x64, 0x65, 0x67:
		return Instr{}, 0, x86RawLiteralRange{}, false, nil
	}

	modRMIndex := -1
	laneBytes := 0
	if len(code) >= offset+5 && code[offset] == 0xc4 && code[offset+1]&0x1f == 2 {
		modRMIndex = offset + 4
		switch code[offset+3] {
		case 0x78:
			laneBytes = 1
		case 0x79:
			laneBytes = 2
		case 0x58:
			laneBytes = 4
		case 0x59:
			laneBytes = 8
		}
	} else if len(code) >= offset+6 && code[offset] == 0x62 && code[offset+1]&0x0f == 2 {
		modRMIndex = offset + 5
		switch code[offset+4] {
		case 0x78:
			laneBytes = 1
		case 0x79:
			laneBytes = 2
		case 0x58:
			laneBytes = 4
		case 0x59:
			laneBytes = 8
		}
	}
	if laneBytes == 0 || modRMIndex < 0 || code[modRMIndex]&0xc7 != 0x05 {
		return Instr{}, 0, x86RawLiteralRange{}, false, nil
	}
	patched, value, literal, err := x86RawRIPLiteral(code, offset, modRMIndex, laneBytes)
	if err != nil {
		return Instr{}, 0, x86RawLiteralRange{}, true, err
	}
	length := len(patched)
	instruction, consumed, ok, err := decodedX86VEXPackedBroadcastInstruction(patched, mode)
	if err != nil || !ok || consumed != length {
		if err == nil {
			err = fmt.Errorf("RIP-relative packed broadcast did not match its VEX/EVEX grammar")
		}
		return Instr{}, 0, x86RawLiteralRange{}, true, err
	}

	setX86RawRIPLiteral(&instruction, value)
	return instruction, length, literal, true, nil
}

// decodeX86RawVMOVIntegerLiteral covers all memory-to-X VMOVD/VMOVQ VEX and
// EVEX encodings, including VMOVQ's alternate F3 7E encoding. The ordinary
// decoder validates the full grammar after the fixed-length ModRM rewrite.
func decodeX86RawVMOVIntegerLiteral(code []byte, offset, mode int) (Instr, int, x86RawLiteralRange, bool, error) {
	if mode != 64 || offset >= len(code) {
		return Instr{}, 0, x86RawLiteralRange{}, false, nil
	}

	switch code[offset] {
	case 0x64, 0x65, 0x67:
		return Instr{}, 0, x86RawLiteralRange{}, false, nil
	}
	modRMIndex := -1
	width := 0
	if len(code) >= offset+4 && code[offset] == 0xc5 {
		modRMIndex = offset + 3
		width = x86RawVMOVLoadWidth(code[offset+2], code[offset+1]&3, false, false)
	} else if len(code) >= offset+5 && code[offset] == 0xc4 && code[offset+1]&0x1f == 1 {
		modRMIndex = offset + 4
		width = x86RawVMOVLoadWidth(code[offset+3], code[offset+2]&3, code[offset+2]&0x80 != 0, false)
	} else if len(code) >= offset+6 && code[offset] == 0x62 && code[offset+1]&0x0f == 1 {
		modRMIndex = offset + 5
		width = x86RawVMOVLoadWidth(code[offset+4], code[offset+2]&3, code[offset+2]&0x80 != 0, true)
	}
	if width == 0 || modRMIndex < 0 || code[modRMIndex]&0xc7 != 0x05 {
		return Instr{}, 0, x86RawLiteralRange{}, false, nil
	}

	patched, value, literal, err := x86RawRIPLiteral(code, offset, modRMIndex, width)
	if err != nil {
		return Instr{}, 0, x86RawLiteralRange{}, true, err
	}
	instruction, consumed, ok, err := decodedX86VMOVQInstruction(patched, mode)
	if err != nil || !ok || consumed != len(patched) ||
		len(instruction.Args) != 2 || instruction.Args[0].Kind != OpMem {
		if err == nil {
			err = fmt.Errorf("RIP-relative VMOVD/VMOVQ load did not match its VEX/EVEX grammar")
		}
		return Instr{}, 0, x86RawLiteralRange{}, true, err
	}
	setX86RawRIPLiteral(&instruction, value)
	return instruction, len(patched), literal, true, nil
}

// decodeX86RawScalarFloatBroadcastLiteral resolves VBROADCASTSS/SD loads
// through the same source-local, unreachable constant-pool proof.
func decodeX86RawScalarFloatBroadcastLiteral(code []byte, offset, mode int) (Instr, int, x86RawLiteralRange, bool, error) {
	if mode != 64 || offset >= len(code) {
		return Instr{}, 0, x86RawLiteralRange{}, false, nil
	}
	p, matched := decodeX86RawVectorEncoding(code[offset:])
	if !matched || p.mapNumber != 2 || p.pp != 1 ||
		(p.opcode != 0x18 && p.opcode != 0x19) ||
		p.segment != "" || p.addressOverride {
		return Instr{}, 0, x86RawLiteralRange{}, false, nil
	}
	modRMIndex := offset + p.modRM
	if len(code) <= modRMIndex || code[modRMIndex]&0xc7 != 0x05 {
		return Instr{}, 0, x86RawLiteralRange{}, false, nil
	}
	width := 4
	if p.opcode == 0x19 {
		width = 8
	}
	patched, value, literal, err := x86RawRIPLiteral(code, offset, modRMIndex, width)
	if err != nil {
		return Instr{}, 0, x86RawLiteralRange{}, true, err
	}
	instruction, consumed, ok, err := decodedX86ScalarBroadcastInstruction(patched, mode)
	if err != nil || !ok || consumed != len(patched) ||
		len(instruction.Args) < 2 || instruction.Args[0].Kind != OpMem {
		if err == nil {
			err = fmt.Errorf("RIP-relative scalar float broadcast did not match its VEX/EVEX grammar")
		}
		return Instr{}, 0, x86RawLiteralRange{}, true, err
	}
	setX86RawRIPLiteral(&instruction, value)
	return instruction, len(patched), literal, true, nil
}

func x86RawVMOVLoadWidth(opcode, pp byte, width64, evex bool) int {
	if opcode == 0x6e && pp == 1 {
		if width64 {
			return 8
		}
		return 4
	}
	if opcode == 0x7e && pp == 2 && (width64 == evex) {
		return 8
	}
	return 0
}

func x86RawRIPLiteral(code []byte, offset, modRMIndex, width int) ([]byte, int64, x86RawLiteralRange, error) {
	if len(code) < modRMIndex+5 {
		return nil, 0, x86RawLiteralRange{}, fmt.Errorf("truncated RIP-relative literal")
	}
	length := modRMIndex + 5 - offset
	displacement := int32(binary.LittleEndian.Uint32(code[modRMIndex+1 : modRMIndex+5]))
	first := offset + length + int(displacement)
	last := first + width
	if first < 0 || last > len(code) || last < first {
		return nil, 0, x86RawLiteralRange{}, fmt.Errorf("RIP-relative literal [%d,%d) is outside directive group", first, last)
	}

	// A base-register disp32 has the same byte length as RIP+disp32, allowing
	// the existing VEX/EVEX grammar to validate every other encoding field.
	patched := append([]byte(nil), code[offset:offset+length]...)
	patched[modRMIndex-offset] = patched[modRMIndex-offset]&0x38 | 0x80
	var value uint64
	for index, b := range code[first:last] {
		value |= uint64(b) << (8 * index)
	}
	return patched, int64(value), x86RawLiteralRange{first: first, last: last}, nil
}

func setX86RawRIPLiteral(instruction *Instr, value int64) {
	instruction.Args[0] = Operand{Kind: OpImm, Imm: value}
	instruction.x86Encoded = true
	instruction.x86RIPLiteral = true
	rawArgs := make([]string, len(instruction.Args))
	for index, operand := range instruction.Args {
		rawArgs[index] = operand.String()
	}
	instruction.Raw = fmt.Sprintf("%s %s", instruction.Op, strings.Join(rawArgs, ", "))
}
