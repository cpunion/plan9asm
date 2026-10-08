package plan9asm

import (
	"encoding/binary"

	"github.com/xgo-dev/plan9asm/internal/gostaticstorage"
)

// This offline, private observation is not a live caller region or authority.
// The consumer must independently bind its complete source, allocation,
// continuation, initialized bytes and native effects before using a carrier.
type arm64ByteCarrierEntryObservation struct {
	allocation, dataStart, dataBytes int64
	callPC, continuationPC           int64
	parameterStores, resultLoads     []arm64ByteCarrierMemory
}

func arm64ByteCarrierTypedEntry(text gostaticstorage.TextRecord, declaration FuncSig) (arm64ByteCarrierEntryObservation, error) {
	var observed arm64ByteCarrierEntryObservation
	fail := func() (arm64ByteCarrierEntryObservation, error) {
		return arm64ByteCarrierEntryObservation{}, arm64ByteCarrierContext("complete actual numeric compiler entry/effects differ")
	}
	if text.Name != declaration.Name || text.ABI != 1 || text.Size != int64(len(text.Code)) || len(text.Code)%4 != 0 ||
		len(text.Code) < 36 || len(text.Relocations) != 1 || len(declaration.Args) != 1 || declaration.Args[0] != I64 ||
		len(declaration.Frame.Params) != 1 || len(declaration.Frame.Results) == 0 || len(declaration.Frame.Results) > 8 {
		return fail()
	}
	parameter := declaration.Frame.Params[0]
	if parameter.Offset != 0 || parameter.Type != I64 || parameter.Index != 0 || len(frameSlotFields(parameter)) != 0 {
		return fail()
	}
	fields, aggregate := parseLiteralStructFields(declaration.Ret)
	if !aggregate {
		fields = []LLVMType{declaration.Ret}
	}
	if len(fields) != len(declaration.Frame.Results) {
		return fail()
	}
	word := func(pc int64) uint32 { return binary.LittleEndian.Uint32(text.Code[pc:]) }
	// The entire observed wrapper is straight-line. Saved FP/LR are not part
	// of its data carrier. Neither a SUB alone nor a rounded signature size
	// can produce this observation.
	first := word(0)
	if first&^uint32(511<<12) != 0xf8000ffe || word(4) != 0xf81f83fd || word(8) != 0xd10023fd || word(12) != 0xf90007e0 {
		return fail()
	}
	allocation := int64(first >> 12 & 511)
	if allocation&256 != 0 {
		allocation -= 512
	}
	allocation = -allocation
	if allocation <= 8 || allocation > 240 || allocation%16 != 0 {
		return fail()
	}
	call := text.Relocations[0]
	if call.Offset != 16 || call.Size != 4 || call.Type != 9 || call.Add != 0 || call.Target.Name != text.Name ||
		!call.Target.ABIKnown || call.Target.ABI != 0 || word(16) != 0x94000000 {
		return fail()
	}
	observed.allocation, observed.dataStart, observed.dataBytes = allocation, 8, allocation-8
	observed.callPC, observed.continuationPC = 16, 20
	observed.parameterStores = []arm64ByteCarrierMemory{{PC: 12, Offset: 8, Bytes: 8, Instruction: word(12)}}
	pc := int64(20)
	previousEnd := int64(8)
	for index := 0; index < len(fields); {
		field := declaration.Frame.Results[index]
		if field.Index != index || field.Type != fields[index] || len(frameSlotFields(field)) != 0 ||
			field.Type != I8 && field.Type != I16 && field.Type != I32 && field.Type != I64 ||
			field.Offset < previousEnd || field.Offset+8+frameTypeSize(field.Type, 8) > allocation || pc+16 > int64(len(text.Code)) {
			return fail()
		}
		instruction := word(pc)
		width := frameTypeSize(field.Type, 8)
		if instruction&0xffc003e0 == 0xa94003e0 {
			// Integer LDP is one complete instruction with two typed read effects.
			if index+1 >= len(fields) || field.Type != I64 || fields[index+1] != I64 ||
				declaration.Frame.Results[index+1].Type != I64 || declaration.Frame.Results[index+1].Index != index+1 ||
				declaration.Frame.Results[index+1].Offset != field.Offset+8 ||
				len(frameSlotFields(declaration.Frame.Results[index+1])) != 0 || instruction&31 != uint32(index) ||
				instruction>>10&31 != uint32(index+1) || field.Offset+24 > allocation {
				return fail()
			}
			offset := int64(instruction >> 15 & 127)
			if offset&64 != 0 {
				offset -= 128
			}
			if offset*8 != field.Offset+8 {
				return fail()
			}
			observed.resultLoads = append(observed.resultLoads,
				arm64ByteCarrierMemory{PC: pc, Offset: field.Offset + 8, Bytes: 8, Instruction: instruction},
				arm64ByteCarrierMemory{PC: pc, Offset: field.Offset + 16, Bytes: 8, Instruction: instruction})
			previousEnd = field.Offset + 16
			index += 2
		} else {
			if instruction&0x3fc003e0 != 0x394003e0 || instruction&31 != uint32(index) ||
				int64(1)<<(instruction>>30) != width || int64(instruction>>10&4095)*width != field.Offset+8 {
				return fail()
			}
			observed.resultLoads = append(observed.resultLoads, arm64ByteCarrierMemory{
				PC: pc, Offset: field.Offset + 8, Bytes: width, Instruction: instruction,
			})
			previousEnd = field.Offset + width
			index++
		}
		pc += 4
	}
	if word(pc) != 0xf85f83fd || word(pc+4) != 0xf84007fe|uint32(allocation)<<12 || word(pc+8) != 0xd65f03c0 {
		return fail()
	}
	for pc += 12; pc < int64(len(text.Code)); pc += 4 {
		if word(pc) != 0 {
			return fail()
		}
	}
	return observed, nil
}
