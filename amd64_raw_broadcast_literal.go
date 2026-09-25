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
	patched, data, literal, err := x86RawRIPBytes(code, offset, modRMIndex, width)
	if err != nil {
		return nil, 0, x86RawLiteralRange{}, err
	}
	if width > 8 {
		return nil, 0, x86RawLiteralRange{}, fmt.Errorf("%d-byte RIP literal does not fit an immediate", width)
	}
	var value uint64
	for index, b := range data {
		value |= uint64(b) << (8 * index)
	}
	return patched, int64(value), literal, nil
}

func x86RawRIPBytes(code []byte, offset, modRMIndex, width int) ([]byte, []byte, x86RawLiteralRange, error) {
	return x86RawRIPBytesWithSuffix(code, offset, modRMIndex, width, 0)
}

func x86RawRIPBytesWithSuffix(code []byte, offset, modRMIndex, width, trailingBytes int) ([]byte, []byte, x86RawLiteralRange, error) {
	if trailingBytes < 0 || len(code) < modRMIndex+5+trailingBytes {
		return nil, nil, x86RawLiteralRange{}, fmt.Errorf("truncated RIP-relative literal")
	}
	length := modRMIndex + 5 + trailingBytes - offset
	displacement := int32(binary.LittleEndian.Uint32(code[modRMIndex+1 : modRMIndex+5]))
	first := offset + length + int(displacement)
	last := first + width
	if width <= 0 || first < 0 || last > len(code) || last < first {
		return nil, nil, x86RawLiteralRange{}, fmt.Errorf("RIP-relative literal [%d,%d) is outside directive group", first, last)
	}

	// A base-register disp32 has the same byte length as RIP+disp32, allowing
	// the existing VEX/EVEX grammar to validate every other encoding field.
	patched := append([]byte(nil), code[offset:offset+length]...)
	patched[modRMIndex-offset] = patched[modRMIndex-offset]&0x38 | 0x80
	data := append([]byte(nil), code[first:last]...)
	return patched, data, x86RawLiteralRange{first: first, last: last}, nil
}

// decodeX86RawPackedMoveRIPData handles source-local X/Y/Z memory loads for
// the complete packed float and VEX integer move decoder families.
func decodeX86RawPackedMoveRIPData(code []byte, offset, mode int) (Instr, int, x86RawLiteralRange, bool, error) {
	if mode != 64 || offset >= len(code) {
		return Instr{}, 0, x86RawLiteralRange{}, false, nil
	}
	p, matched := decodeX86RawVectorEncoding(code[offset:])
	packedFloat := (p.opcode == 0x10 || p.opcode == 0x28) && p.pp <= 1
	packedInteger := p.opcode == 0x6f && !p.evex && (p.pp == 1 || p.pp == 2)
	if !matched || p.mapNumber != 1 || (!packedFloat && !packedInteger) ||
		p.segment != "" || p.addressOverride {
		return Instr{}, 0, x86RawLiteralRange{}, false, nil
	}
	modRMIndex := offset + p.modRM
	if len(code) <= modRMIndex || code[modRMIndex]&0xc7 != 0x05 {
		return Instr{}, 0, x86RawLiteralRange{}, false, nil
	}
	width := 16 << p.vectorLength
	return x86RawRIPDataThroughDecoder(
		code, offset, mode, modRMIndex, width,
		decodedX86PackedMoveInstruction, "packed move",
	)
}

// decodeX86RawEVEXPackedIntegerMoveRIPData covers the six EVEX VMOVDQA/DQU
// load opcodes across X/Y/Z widths and optional K masks.
func decodeX86RawEVEXPackedIntegerMoveRIPData(code []byte, offset, mode int) (Instr, int, x86RawLiteralRange, bool, error) {
	if mode != 64 || offset >= len(code) {
		return Instr{}, 0, x86RawLiteralRange{}, false, nil
	}
	p, matched := decodeX86RawVectorEncoding(code[offset:])
	if !matched || !p.evex || p.mapNumber != 1 || p.opcode != 0x6f ||
		(p.pp != 1 && p.pp != 2 && p.pp != 3) ||
		p.segment != "" || p.addressOverride {
		return Instr{}, 0, x86RawLiteralRange{}, false, nil
	}
	modRMIndex := offset + p.modRM
	if len(code) <= modRMIndex || code[modRMIndex]&0xc7 != 0x05 {
		return Instr{}, 0, x86RawLiteralRange{}, false, nil
	}
	width := 16 << p.vectorLength
	return x86RawRIPDataThroughDecoder(
		code, offset, mode, modRMIndex, width,
		decodedX86EVEXPackedIntegerMoveInstruction, "EVEX packed integer move",
	)
}

// decodeX86RawVBROADCASTI128RIPData resolves the complete VEX.256 m128-to-Y
// form through the existing VBROADCASTI128 grammar.
func decodeX86RawVBROADCASTI128RIPData(code []byte, offset, mode int) (Instr, int, x86RawLiteralRange, bool, error) {
	if mode != 64 || offset >= len(code) {
		return Instr{}, 0, x86RawLiteralRange{}, false, nil
	}
	p, matched := decodeX86RawVectorEncoding(code[offset:])
	if !matched || p.evex || p.mapNumber != 2 || p.pp != 1 ||
		p.vectorLength != 1 || p.opcode != 0x5a ||
		p.segment != "" || p.addressOverride {
		return Instr{}, 0, x86RawLiteralRange{}, false, nil
	}
	modRMIndex := offset + p.modRM
	if len(code) <= modRMIndex || code[modRMIndex]&0xc7 != 0x05 {
		return Instr{}, 0, x86RawLiteralRange{}, false, nil
	}
	return x86RawRIPDataThroughDecoder(
		code, offset, mode, modRMIndex, 16,
		decodedX86VBROADCASTI128Instruction, "VBROADCASTI128",
	)
}

// decodeX86RawPackedLogicalRIPData covers VEX VPAND/ANDN/OR/XOR and their
// EVEX D/Q counterparts. EVEX.b reads one scalar lane; other forms read the
// full X/Y/Z-width vector from the same unreachable source-local pool.
func decodeX86RawPackedLogicalRIPData(code []byte, offset, mode int) (Instr, int, x86RawLiteralRange, bool, error) {
	if mode != 64 || offset >= len(code) {
		return Instr{}, 0, x86RawLiteralRange{}, false, nil
	}
	p, matched := decodeX86RawVectorEncoding(code[offset:])
	if !matched || p.mapNumber != 1 || p.pp != 1 ||
		p.segment != "" || p.addressOverride {
		return Instr{}, 0, x86RawLiteralRange{}, false, nil
	}
	switch p.opcode {
	case 0xdb, 0xdf, 0xeb, 0xef:
	default:
		return Instr{}, 0, x86RawLiteralRange{}, false, nil
	}
	modRMIndex := offset + p.modRM
	if len(code) <= modRMIndex || code[modRMIndex]&0xc7 != 0x05 {
		return Instr{}, 0, x86RawLiteralRange{}, false, nil
	}
	width := 16 << p.vectorLength
	decoder := x86RawInstructionDecoder(decodedX86VEXPackedIntegerLogicalInstruction)
	family := "VEX packed integer logical"
	if p.evex {
		decoder = decodedX86EVEXPackedLogicalInstruction
		family = "EVEX packed logical"
		if p.broadcast {
			width = 4
			if p.w {
				width = 8
			}
		}
	}
	return x86RawRIPDataThroughDecoder(
		code, offset, mode, modRMIndex, width, decoder, family,
	)
}

// decodeX86RawScalarMoveRIPData covers memory-to-X VMOVSS/VMOVSD VEX and
// EVEX encodings. The typed scalar decoder enforces reserved fields and masks.
func decodeX86RawScalarMoveRIPData(code []byte, offset, mode int) (Instr, int, x86RawLiteralRange, bool, error) {
	if mode != 64 || offset >= len(code) {
		return Instr{}, 0, x86RawLiteralRange{}, false, nil
	}
	p, matched := decodeX86RawVectorEncoding(code[offset:])
	if !matched || p.mapNumber != 1 || p.opcode != 0x10 ||
		(p.pp != 2 && p.pp != 3) || p.segment != "" || p.addressOverride {
		return Instr{}, 0, x86RawLiteralRange{}, false, nil
	}
	modRMIndex := offset + p.modRM
	if len(code) <= modRMIndex || code[modRMIndex]&0xc7 != 0x05 {
		return Instr{}, 0, x86RawLiteralRange{}, false, nil
	}
	width := 4
	if p.pp == 3 {
		width = 8
	}
	return x86RawRIPDataThroughDecoder(
		code, offset, mode, modRMIndex, width,
		decodedX86ScalarMoveInstruction, "scalar move",
	)
}

// decodeX86RawMinMaxRIPData covers the complete VEX/EVEX packed integer
// min/max family. EVEX.b reads one D/Q lane; other forms read a full vector.
func decodeX86RawMinMaxRIPData(code []byte, offset, mode int) (Instr, int, x86RawLiteralRange, bool, error) {
	if mode != 64 || offset >= len(code) {
		return Instr{}, 0, x86RawLiteralRange{}, false, nil
	}
	p, matched := decodeX86RawVectorEncoding(code[offset:])
	if !matched || p.segment != "" || p.addressOverride {
		return Instr{}, 0, x86RawLiteralRange{}, false, nil
	}
	modRMIndex := offset + p.modRM
	if len(code) <= modRMIndex || code[modRMIndex]&0xc7 != 0x05 {
		return Instr{}, 0, x86RawLiteralRange{}, false, nil
	}
	for _, spec := range amd64PackedIntegerMinMaxSpecs {
		if p.mapNumber != spec.mapNumber || p.opcode != spec.opcode {
			continue
		}
		if !p.evex && spec.laneBits == 64 {
			continue
		}
		if p.evex && spec.laneBits >= 32 && p.w != (spec.laneBits == 64) {
			continue
		}
		width := 16 << p.vectorLength
		if p.evex && p.broadcast {
			width = spec.laneBits / 8
		}
		return x86RawRIPDataThroughDecoder(
			code, offset, mode, modRMIndex, width,
			decodedX86PackedIntegerMinMaxInstruction, "packed integer min/max",
		)
	}
	return Instr{}, 0, x86RawLiteralRange{}, false, nil
}

// decodeX86RawPackedCompareRIPData covers all VPCMPEQ/GT B/W/D/Q source
// memory forms. The ordinary decoder checks the EVEX mask-result grammar.
func decodeX86RawPackedCompareRIPData(code []byte, offset, mode int) (Instr, int, x86RawLiteralRange, bool, error) {
	if mode != 64 || offset >= len(code) {
		return Instr{}, 0, x86RawLiteralRange{}, false, nil
	}
	p, matched := decodeX86RawVectorEncoding(code[offset:])
	if !matched || p.segment != "" || p.addressOverride {
		return Instr{}, 0, x86RawLiteralRange{}, false, nil
	}
	modRMIndex := offset + p.modRM
	if len(code) <= modRMIndex || code[modRMIndex]&0xc7 != 0x05 {
		return Instr{}, 0, x86RawLiteralRange{}, false, nil
	}
	for _, spec := range amd64PackedIntegerCompareSpecs {
		if p.mapNumber != spec.mapNumber || p.opcode != spec.opcode {
			continue
		}
		width := 16 << p.vectorLength
		if p.evex && p.broadcast {
			width = spec.laneBits / 8
		}
		return x86RawRIPDataThroughDecoder(
			code, offset, mode, modRMIndex, width,
			decodedX86PackedIntegerCompareInstruction, "packed integer compare",
		)
	}
	return Instr{}, 0, x86RawLiteralRange{}, false, nil
}

// decodeX86RawVariableShiftRIPData covers the VEX and EVEX per-lane variable
// shift family, including scalar D/Q memory broadcast in EVEX encodings.
func decodeX86RawVariableShiftRIPData(code []byte, offset, mode int) (Instr, int, x86RawLiteralRange, bool, error) {
	if mode != 64 || offset >= len(code) {
		return Instr{}, 0, x86RawLiteralRange{}, false, nil
	}
	p, matched := decodeX86RawVectorEncoding(code[offset:])
	if !matched || p.mapNumber != 2 || p.pp != 1 ||
		p.segment != "" || p.addressOverride {
		return Instr{}, 0, x86RawLiteralRange{}, false, nil
	}
	modRMIndex := offset + p.modRM
	if len(code) <= modRMIndex || code[modRMIndex]&0xc7 != 0x05 {
		return Instr{}, 0, x86RawLiteralRange{}, false, nil
	}
	for _, spec := range amd64PerLaneVariableShiftSpecs {
		if p.opcode != int(spec.opcode) || p.w != (spec.laneBits != 32) ||
			!p.evex && !spec.vex {
			continue
		}
		width := 16 << p.vectorLength
		if p.evex && p.broadcast {
			width = spec.laneBits / 8
		}
		return x86RawRIPDataThroughDecoder(
			code, offset, mode, modRMIndex, width,
			decodedX86PerLaneVariableShiftInstruction, "per-lane variable shift",
		)
	}
	return Instr{}, 0, x86RawLiteralRange{}, false, nil
}

// decodeX86RawVPSHUFBRIPData covers every VEX and EVEX vector-width load
// accepted by Go's VPSHUFB table. The instruction has no scalar broadcast.
func decodeX86RawVPSHUFBRIPData(code []byte, offset, mode int) (Instr, int, x86RawLiteralRange, bool, error) {
	if mode != 64 || offset >= len(code) {
		return Instr{}, 0, x86RawLiteralRange{}, false, nil
	}
	p, matched := decodeX86RawVectorEncoding(code[offset:])
	if !matched || p.mapNumber != 2 || p.pp != 1 || p.opcode != 0 ||
		p.w || p.broadcast || p.segment != "" || p.addressOverride {
		return Instr{}, 0, x86RawLiteralRange{}, false, nil
	}
	modRMIndex := offset + p.modRM
	if len(code) <= modRMIndex || code[modRMIndex]&0xc7 != 0x05 {
		return Instr{}, 0, x86RawLiteralRange{}, false, nil
	}
	width := 16 << p.vectorLength
	return x86RawRIPDataThroughDecoder(
		code, offset, mode, modRMIndex, width,
		decodedX86VPSHUFBInstruction, "VPSHUFB",
	)
}

// decodeX86RawPackedArithmeticRIPData covers the complete VEX/EVEX VPADD
// and VPSUB families, including their D/Q scalar memory broadcasts.
func decodeX86RawPackedArithmeticRIPData(code []byte, offset, mode int) (Instr, int, x86RawLiteralRange, bool, error) {
	if mode != 64 || offset >= len(code) {
		return Instr{}, 0, x86RawLiteralRange{}, false, nil
	}
	p, matched := decodeX86RawVectorEncoding(code[offset:])
	if !matched || p.mapNumber != 1 || p.pp != 1 ||
		p.segment != "" || p.addressOverride {
		return Instr{}, 0, x86RawLiteralRange{}, false, nil
	}
	properties, recognized := decodedX86PackedIntegerArithmeticOps[byte(p.opcode)]
	if !recognized {
		return Instr{}, 0, x86RawLiteralRange{}, false, nil
	}
	modRMIndex := offset + p.modRM
	if len(code) <= modRMIndex || code[modRMIndex]&0xc7 != 0x05 {
		return Instr{}, 0, x86RawLiteralRange{}, false, nil
	}
	width := 16 << p.vectorLength
	if p.evex && p.broadcast {
		width = properties.laneBytes
	}
	return x86RawRIPDataThroughDecoder(
		code, offset, mode, modRMIndex, width,
		decodedX86PackedIntegerArithmeticInstruction, "packed integer arithmetic",
	)
}

// decodeX86RawFMA3RIPData handles all packed and scalar FMA3 encodings.
// Scalar operations read one lane; packed EVEX.b memory forms broadcast it.
func decodeX86RawFMA3RIPData(code []byte, offset, mode int) (Instr, int, x86RawLiteralRange, bool, error) {
	if mode != 64 || offset >= len(code) {
		return Instr{}, 0, x86RawLiteralRange{}, false, nil
	}
	p, matched := decodeX86RawVectorEncoding(code[offset:])
	if !matched || p.mapNumber != 2 || p.pp != 1 ||
		p.segment != "" || p.addressOverride {
		return Instr{}, 0, x86RawLiteralRange{}, false, nil
	}
	spec, recognized := decodedX86VEXFMA3Ops[byte(p.opcode)]
	if !recognized {
		return Instr{}, 0, x86RawLiteralRange{}, false, nil
	}
	modRMIndex := offset + p.modRM
	if len(code) <= modRMIndex || code[modRMIndex]&0xc7 != 0x05 {
		return Instr{}, 0, x86RawLiteralRange{}, false, nil
	}
	width := 16 << p.vectorLength
	if spec.scalar || p.evex && p.broadcast {
		width = 4
		if p.w {
			width = 8
		}
	}
	return x86RawRIPDataThroughDecoder(
		code, offset, mode, modRMIndex, width,
		decodedX86VEXFMA3Instruction, "FMA3",
	)
}

// decodeX86RawBinaryFloatRIPData covers VADD/MUL/SUB/MIN/DIV/MAX packed and
// scalar forms plus scalar VSQRT, with EVEX packed memory broadcast.
func decodeX86RawBinaryFloatRIPData(code []byte, offset, mode int) (Instr, int, x86RawLiteralRange, bool, error) {
	if mode != 64 || offset >= len(code) {
		return Instr{}, 0, x86RawLiteralRange{}, false, nil
	}
	p, matched := decodeX86RawVectorEncoding(code[offset:])
	if !matched || p.mapNumber != 1 || p.segment != "" || p.addressOverride {
		return Instr{}, 0, x86RawLiteralRange{}, false, nil
	}
	switch p.opcode {
	case 0x58, 0x59, 0x5c, 0x5d, 0x5e, 0x5f:
	case 0x51:
		if p.pp < 2 {
			return Instr{}, 0, x86RawLiteralRange{}, false, nil
		}
	default:
		return Instr{}, 0, x86RawLiteralRange{}, false, nil
	}
	modRMIndex := offset + p.modRM
	if len(code) <= modRMIndex || code[modRMIndex]&0xc7 != 0x05 {
		return Instr{}, 0, x86RawLiteralRange{}, false, nil
	}
	width := 16 << p.vectorLength
	if p.pp >= 2 || p.evex && p.broadcast {
		width = 4
		if p.pp == 1 || p.pp == 3 {
			width = 8
		}
	}
	return x86RawRIPDataThroughDecoder(
		code, offset, mode, modRMIndex, width,
		decodedX86VEXBinaryFloatInstruction, "binary floating",
	)
}

// decodeX86RawScalarFlagCompareRIPData covers VCOMIS{S,D} and
// VUCOMIS{S,D} VEX/EVEX memory sources. EVEX.b remains register-only SAE.
func decodeX86RawScalarFlagCompareRIPData(code []byte, offset, mode int) (Instr, int, x86RawLiteralRange, bool, error) {
	if mode != 64 || offset >= len(code) {
		return Instr{}, 0, x86RawLiteralRange{}, false, nil
	}
	p, matched := decodeX86RawVectorEncoding(code[offset:])
	if !matched || p.mapNumber != 1 || (p.opcode != 0x2e && p.opcode != 0x2f) ||
		(p.pp != 0 && p.pp != 1) || p.segment != "" || p.addressOverride {
		return Instr{}, 0, x86RawLiteralRange{}, false, nil
	}
	modRMIndex := offset + p.modRM
	if len(code) <= modRMIndex || code[modRMIndex]&0xc7 != 0x05 {
		return Instr{}, 0, x86RawLiteralRange{}, false, nil
	}
	width := 4
	if p.pp == 1 {
		width = 8
	}
	return x86RawRIPDataThroughDecoder(
		code, offset, mode, modRMIndex, width,
		decodedX86ScalarFlagCompareInstruction, "scalar flag compare",
	)
}

// decodeX86RawVectorFloatCompareRIPData accounts for VCMP's trailing imm8
// when resolving the RIP displacement and its second Plan 9 operand.
func decodeX86RawVectorFloatCompareRIPData(code []byte, offset, mode int) (Instr, int, x86RawLiteralRange, bool, error) {
	if mode != 64 || offset >= len(code) {
		return Instr{}, 0, x86RawLiteralRange{}, false, nil
	}
	p, matched := decodeX86RawVectorEncoding(code[offset:])
	if !matched || p.mapNumber != 1 || p.opcode != 0xc2 ||
		p.segment != "" || p.addressOverride {
		return Instr{}, 0, x86RawLiteralRange{}, false, nil
	}
	modRMIndex := offset + p.modRM
	if len(code) <= modRMIndex || code[modRMIndex]&0xc7 != 0x05 {
		return Instr{}, 0, x86RawLiteralRange{}, false, nil
	}
	width := 16 << p.vectorLength
	if p.pp >= 2 || p.evex && p.broadcast {
		width = 4
		if p.pp == 1 || p.pp == 3 {
			width = 8
		}
	}
	return x86RawRIPDataThroughDecoderOperand(
		code, offset, mode, modRMIndex, width, 1, 1,
		decodedX86VectorFloatCompareInstruction, "vector floating compare",
	)
}

// decodeX86RawPackedMultiplyRIPData covers the five packed word/D/Q multiply
// opcode rows, including EVEX D/Q memory broadcast and masked widths.
func decodeX86RawPackedMultiplyRIPData(code []byte, offset, mode int) (Instr, int, x86RawLiteralRange, bool, error) {
	if mode != 64 || offset >= len(code) {
		return Instr{}, 0, x86RawLiteralRange{}, false, nil
	}
	p, matched := decodeX86RawVectorEncoding(code[offset:])
	if !matched || p.pp != 1 || p.segment != "" || p.addressOverride {
		return Instr{}, 0, x86RawLiteralRange{}, false, nil
	}
	form, recognized := x86PackedMultiplyFormFor(byte(p.mapNumber), byte(p.opcode))
	if !recognized || p.evex && p.broadcast && form.broadcastByte == 0 {
		return Instr{}, 0, x86RawLiteralRange{}, false, nil
	}
	modRMIndex := offset + p.modRM
	if len(code) <= modRMIndex || code[modRMIndex]&0xc7 != 0x05 {
		return Instr{}, 0, x86RawLiteralRange{}, false, nil
	}
	width := 16 << p.vectorLength
	if p.evex && p.broadcast {
		width = form.broadcastByte
		if p.w {
			width = 8
		}
	}
	return x86RawRIPDataThroughDecoder(
		code, offset, mode, modRMIndex, width,
		decodedX86PackedMultiplyInstruction, "packed multiply",
	)
}

// decodeX86RawVEXPackedMADDRIPData covers both VPMADDWD and VPMADDUBSW
// VEX.128/256 opcode rows. Their memory operand reads the full vector.
func decodeX86RawVEXPackedMADDRIPData(code []byte, offset, mode int) (Instr, int, x86RawLiteralRange, bool, error) {
	if mode != 64 || offset >= len(code) {
		return Instr{}, 0, x86RawLiteralRange{}, false, nil
	}
	p, matched := decodeX86RawVectorEncoding(code[offset:])
	if !matched || p.evex || p.pp != 1 || p.segment != "" || p.addressOverride {
		return Instr{}, 0, x86RawLiteralRange{}, false, nil
	}
	if !(p.mapNumber == 1 && p.opcode == 0xf5 ||
		p.mapNumber == 2 && p.opcode == 0x04) {
		return Instr{}, 0, x86RawLiteralRange{}, false, nil
	}
	modRMIndex := offset + p.modRM
	if len(code) <= modRMIndex || code[modRMIndex]&0xc7 != 0x05 {
		return Instr{}, 0, x86RawLiteralRange{}, false, nil
	}
	return x86RawRIPDataThroughDecoder(
		code, offset, mode, modRMIndex, 16<<p.vectorLength,
		decodedX86VEXPackedMADDInstruction, "packed multiply-add",
	)
}

type x86RawInstructionDecoder func([]byte, int) (Instr, int, bool, error)

func x86RawRIPDataThroughDecoder(
	code []byte, offset, mode, modRMIndex, width int,
	decode x86RawInstructionDecoder, family string,
) (Instr, int, x86RawLiteralRange, bool, error) {
	return x86RawRIPDataThroughDecoderOperand(
		code, offset, mode, modRMIndex, width, 0, 0, decode, family,
	)
}

func x86RawRIPDataThroughDecoderOperand(
	code []byte, offset, mode, modRMIndex, width, trailingBytes, sourceIndex int,
	decode x86RawInstructionDecoder, family string,
) (Instr, int, x86RawLiteralRange, bool, error) {
	patched, data, literal, err := x86RawRIPBytesWithSuffix(code, offset, modRMIndex, width, trailingBytes)
	if err != nil {
		return Instr{}, 0, x86RawLiteralRange{}, true, err
	}
	instruction, consumed, ok, err := decode(patched, mode)
	if err != nil || !ok || consumed != len(patched) ||
		len(instruction.Args) <= sourceIndex || instruction.Args[sourceIndex].Kind != OpMem {
		if err == nil {
			err = fmt.Errorf("RIP-relative %s did not match its grammar", family)
		}
		return Instr{}, 0, x86RawLiteralRange{}, true, err
	}
	setX86RawRIPDataOperand(&instruction, sourceIndex, data)
	return instruction, len(patched), literal, true, nil
}

func setX86RawRIPDataOperand(instruction *Instr, sourceIndex int, data []byte) {
	instruction.Args[sourceIndex] = Operand{Kind: OpSym, Sym: "·__plan9asm_raw_literal_pending(SB)"}
	instruction.x86Encoded = true
	instruction.x86RIPLiteral = true
	instruction.x86RIPLiteralData = data
	rawArgs := make([]string, len(instruction.Args))
	for index, operand := range instruction.Args {
		rawArgs[index] = operand.String()
	}
	instruction.Raw = fmt.Sprintf("%s %s", instruction.Op, strings.Join(rawArgs, ", "))
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
