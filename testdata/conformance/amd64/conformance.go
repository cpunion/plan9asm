//go:build amd64

package amd64conformance

func byteMemory(p *[4]byte, value byte)

func byteFlags(p *[4]byte, flags *[8]byte)

func unpackLowQWords(dst, src *[2]uint64)

func unpackDuplicateLowQWord(dst *[2]uint64)

func shiftLegacyThreeOperand(src, dst, amount uint32) uint32

func clearTopBit(value uint64) uint64

func doubleShift32(out *[8]uint32, src, dst, amount uint32)

func doubleShift64(out *[8]uint64, src, dst, amount uint64)

func goHexVectorOps(out *[160]byte, a, b *[16]byte)

func goHexWordOps(value, count uint64) uint64

func packedArithmeticShift32(out *[8]int32, src *[4]int32)
