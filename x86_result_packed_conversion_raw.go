package plan9asm

// This value snapshot is copied with its sealed raw origin, including immutable
// whole-source replay clones. No default false value identifies a conversion:
// arch is present only after its complete typed physical shape validates.
type x86RawPackedConversionIdentity struct {
	arch               string
	encoded, evex, mmx bool
}

func x86PackedConversionRawIdentity(ins Instr, arch string) x86RawPackedConversionIdentity {
	_, signed, signedErr := x86SignedPackedFloatProperties(ins, arch)
	_, numeric, numericErr := x86PackedConversionProperties(ins, arch)
	if !signed && !numeric || signed && signedErr != nil || numeric && numericErr != nil {
		return x86RawPackedConversionIdentity{}
	}
	return x86RawPackedConversionIdentity{
		arch: arch, encoded: ins.x86Encoded,
		evex: ins.x86PackedNumericEVEX, mmx: ins.x86MMXSourceConversion,
	}
}
