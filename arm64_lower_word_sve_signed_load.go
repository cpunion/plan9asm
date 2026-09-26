package plan9asm

import "fmt"

type arm64SVESignedLoadAddress uint8

const (
	arm64SVESignedLoadRegister arm64SVESignedLoadAddress = iota
	arm64SVESignedLoadImmediate
	arm64SVESignedLoadVectorOffset
	arm64SVESignedLoadVectorBase
)

type arm64RawSVESignedLoadRow struct {
	op          Op
	elementBits int
	address     arm64SVESignedLoadAddress
	extension   ExtendOp
	scale       int64
}

var arm64RawSVESignedLoadRows = arm64SVESignedLoadRawTable()

// LD1SB/SH/SW share the signed ordinary-load semantics. Expand only the
// encoding axes here: widening element size, scalar/vector address, index
// extension and scale. No opcode-specific lowering is needed.
func arm64SVESignedLoadRawTable() map[uint32]arm64RawSVESignedLoadRow {
	rows := make(map[uint32]arm64RawSVESignedLoadRow)
	for op, spec := range arm64SVEOrdinaryMemorySpecs {
		if !spec.load || !spec.signed {
			continue
		}
		memorySize := uint32(0)
		for 8<<memorySize < spec.memoryBits {
			memorySize++
		}
		for elementSize := memorySize + 1; elementSize < 4; elementSize++ {
			row := arm64RawSVESignedLoadRow{op: op, elementBits: 8 << elementSize, scale: int64(spec.memoryBits / 8)}
			base := uint32(0xa5804000) - memorySize<<23 + (3-elementSize)<<21
			row.address = arm64SVESignedLoadRegister
			rows[base] = row
			row.address = arm64SVESignedLoadImmediate
			rows[base+0x6000] = row
			if elementSize < 2 {
				continue
			}
			vectorBase := uint32(0x84208000) + memorySize<<23
			if elementSize == 3 {
				vectorBase |= 1 << 30
			}
			row.address = arm64SVESignedLoadVectorBase
			rows[vectorBase] = row
			row.address = arm64SVESignedLoadVectorOffset
			for scale := uint32(0); scale < 2; scale++ {
				if memorySize == 0 && scale != 0 {
					continue
				}
				row.scale = 1 << (scale * memorySize)
				if elementSize == 3 {
					row.extension = ""
					rows[0xc4408000+memorySize<<23+scale<<21] = row
				}
				for signed, extension := range []ExtendOp{ExtendUXTW, ExtendSXTW} {
					row.extension = extension
					base := uint32(0x84000000) + memorySize<<23 + uint32(signed)<<22 + scale<<21
					if elementSize == 3 {
						base |= 1 << 30
					}
					rows[base] = row
				}
			}
		}
	}
	return rows
}

func decodeARM64RawSVESignedLoad(word uint32) (Instr, bool) {
	row, ok := arm64RawSVESignedLoadRows[word&0xffe0e000]
	if !ok {
		return Instr{}, false
	}
	base, index := word>>5&31, word>>16&31
	width := map[int]byte{16: 'H', 32: 'S', 64: 'D'}[row.elementBits]
	memory := MemRef{Base: Reg(fmt.Sprintf("R%d", base))}
	if base == 31 {
		memory.Base = "RSP"
	}
	switch row.address {
	case arm64SVESignedLoadRegister:
		if index == 31 { // Xm excludes 31; this is not a zero-index alias.
			return Instr{}, false
		}
		memory.Index, memory.Scale = Reg(fmt.Sprintf("R%d", index)), row.scale
		if row.scale == 1 {
			memory.Base, memory.Index = memory.Index, memory.Base
		}
	case arm64SVESignedLoadImmediate:
		if index&16 != 0 {
			return Instr{}, false
		}
		offset := int(index)
		if offset >= 8 {
			offset -= 16
		}
		if offset < 0 {
			memory.OffRaw = fmt.Sprintf("-VL*%d", -offset)
		} else if offset > 0 {
			memory.OffRaw = fmt.Sprintf("VL*%d", offset)
		}
	case arm64SVESignedLoadVectorOffset:
		memory.Index = Reg(fmt.Sprintf("Z%d.%c", index, width))
		memory.IndexExt, memory.Scale = row.extension, row.scale
	case arm64SVESignedLoadVectorBase:
		memory.Base = Reg(fmt.Sprintf("Z%d.%c", base, width))
		memory.Off = int64(index) * row.scale
	}
	return Instr{Op: row.op, Raw: fmt.Sprintf("WORD $%#08x", word), Args: []Operand{
		{Kind: OpMem, Mem: memory},
		{Kind: OpReg, Reg: Reg(fmt.Sprintf("P%d.Z", word>>10&7))},
		{Kind: OpRegList, RegList: []Reg{Reg(fmt.Sprintf("Z%d.%c", word&31, width))}},
	}}, true
}
