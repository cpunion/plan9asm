package plan9asm

import "golang.org/x/arch/x86/x86asm"

type x86NumericalNativeBoundary struct {
	PC     int64
	Length int
}

// This is a byte-boundary observation, not a raw/source/literal or memory
// capability. In particular, vector instructions are never put into scalar:
// numerical-read binding still needs a real x86asm scalar decode, its exact
// original source ordinal, relocation, layout, owner and live source guard.
// Failure discards the complete observation, including otherwise valid prefixes
// of the TEXT. Unreachable bytes after RET or a branch are still checked.
func x86NumericalNativeBoundaries(code []byte, mode int) ([]x86NumericalNativeBoundary, map[int64]x86asm.Inst, error) {
	if mode != 32 && mode != 64 || len(code) == 0 {
		return nil, nil, numericalSBContext("empty TEXT or unknown native x86 mode")
	}
	var boundaries []x86NumericalNativeBoundary
	scalar := make(map[int64]x86asm.Inst)
	for pc := int64(0); pc < int64(len(code)); {
		remaining := code[pc:]
		length := 0
		if x86NumericalVectorPrefix(remaining, mode) {
			var err error
			length, err = x86NumericalVectorBoundary(remaining, mode)
			if err != nil {
				return nil, nil, numericalSBContext("whole native vector boundary at %d: %v", pc, err)
			}
		} else {
			native, err := x86asm.Decode(remaining, mode)
			if err != nil || native.Op == 0 || native.Op == x86asm.CALL || native.Op == x86asm.LCALL {
				return nil, nil, numericalSBContext("unknown/calling whole native instruction boundary at %d", pc)
			}
			for _, prefix := range native.Prefix {
				if prefix.IsVEX() || prefix&x86asm.PrefixInvalid != 0 {
					return nil, nil, numericalSBContext("unbound vector or invalid native prefix at %d", pc)
				}
			}
			length = native.Len
			scalar[pc] = native
		}
		if length <= 0 || length > 15 || length > len(remaining) {
			return nil, nil, numericalSBContext("invalid complete native instruction length at %d", pc)
		}
		boundaries = append(boundaries, x86NumericalNativeBoundary{PC: pc, Length: length})
		pc += int64(length)
	}
	return boundaries, scalar, nil
}

// Detect even incomplete/unsupported vector prefixes before generic decoding:
// x/arch can report an unrelated legacy instruction and a wrong positive Len.
// In mode32, C4/C5/62 with a memory ModRM remain legacy LES/LDS/BOUND.
func x86NumericalVectorPrefix(code []byte, mode int) bool {
	for i, value := range code {
		switch {
		case value == 0xc4 || value == 0xc5 || value == 0x62:
			return mode == 64 || i+1 == len(code) || code[i+1]&0xc0 == 0xc0
		case value == 0x26 || value == 0x2e || value == 0x36 || value == 0x3e ||
			value == 0x64 || value == 0x65 || value == 0x66 || value == 0x67 ||
			value == 0xf0 || value == 0xf2 || value == 0xf3 || mode == 64 && value&0xf0 == 0x40:
			if i == 14 {
				return false
			}
		default:
			return false
		}
	}
	return false
}

func x86NumericalVectorBoundary(code []byte, mode int) (int, error) {
	memory := x86RawMemoryDecoder{physicalPC: true}
	decoders := [...]func([]byte, int, x86RawMemoryDecoder) (Instr, int, bool, error){
		decodedX86NumericalConversionBoundary,
		decodedX86VEXFMA3InstructionWithMemory,
		decodedX86VEXBinaryFloatInstructionWithMemory,
		decodedX86AESInstructionWithMemory,
		decodedX86MaskMoveInstructionWithMemory,
		decodedX86PackedMoveInstructionWithMemory,
		decodedX86EVEXPackedIntegerMoveInstructionWithMemory,
		decodedX86VEXNonTemporalMoveInstructionWithMemory,
		decodedX86VEXPackedFloatLogicalInstructionWithMemory,
		decodedX86VEXPackedIntegerLogicalInstructionWithMemory,
		decodedX86EVEXPackedLogicalInstructionWithMemory,
		decodedX86RawEVEXFloatLogicalInstructionWithMemory,
		decodedX86VINSERT128InstructionWithMemory,
		decodedX86EVEXPackedInsertInstructionWithMemory,
		decodedX86VEXTRACT128InstructionWithMemory,
		decodedX86VEXPackedBroadcastInstructionWithMemory,
		decodedX86PackedIntegerCompareInstructionWithMemory,
		decodedX86VMOVQInstructionWithMemory,
		decodedX86VEXVectorTestInstructionWithMemory,
	}
	if ins, length, known, err := decodedX86VEXMoveMaskInstruction(code, mode); known || err != nil {
		if err == nil && mode == 32 {
			err = x86Mode32VectorRegisterContext(ins, "386")
		}
		return length, err
	}
	for _, decode := range decoders {
		if ins, length, known, err := decode(code, mode, memory); known || err != nil {
			if err == nil && mode == 32 {
				err = x86Mode32VectorRegisterContext(ins, "386")
			}
			return length, err
		}
	}
	if p, known := decodeX86RawVectorEncoding(code); known && p.mapNumber == 3 && p.pp == 1 && p.opcode == 0x44 {
		_, length, _, err := decodedX86RawVectorCLMULInstructionWithMemory(code, mode, p, memory)
		return length, err
	}
	// The complete physical EVEX grammar is shared with the source decoder.
	// Its separate source-syntax validation deliberately cannot interpret RIP.
	if _, length, known, err := decodedX86TernaryLogicInstructionWithMemory(code, mode, memory); known || err != nil {
		return length, err
	}
	if _, length, known, err := decodedX86VectorZeroInstruction(code, mode); known || err != nil {
		return length, err
	}
	return 0, numericalSBContext("unregistered or truncated whole native VEX/EVEX encoding")
}

func decodedX86NumericalConversionBoundary(code []byte, mode int, memory x86RawMemoryDecoder) (Instr, int, bool, error) {
	if ins, length, known, err := decodedX86PackedNumericConversionInstructionWithMemory(code, mode, memory); known || err != nil {
		return ins, length, known, err
	}
	p, known := decodeX86RawVectorEncoding(code)
	if known && p.mapNumber == 1 {
		switch {
		case p.opcode == 0x51 && p.pp < 2:
			return decodedX86SameWidthConversionInstructionWithMemory(code, mode, memory)
		case p.opcode == 0x5b && p.pp == 0:
			return decodedX86SameWidthConversionInstructionWithMemory(code, mode, memory)
		case p.opcode == 0xe6 && p.pp == 2:
			return decodedX86PackedDoubleDwordInstructionWithMemory(code, mode, memory)
		}
	}
	return Instr{}, 0, false, nil
}
