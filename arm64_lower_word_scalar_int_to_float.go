package plan9asm

import "fmt"

type arm64RawScalarIntToFloat struct {
	floatBits   int
	integerBits int
	unsigned    bool
	source      int
	destination int
}

func decodeARM64RawScalarIntToFloat(word uint32) (arm64RawScalarIntToFloat, bool) {
	const (
		integerWidth   = uint32(1 << 31)
		floatType      = uint32(3 << 22)
		unsignedBit    = uint32(1 << 16)
		source         = uint32(31 << 5)
		destination    = uint32(31)
		variableFields = integerWidth | floatType | unsignedBit | source | destination
		conversionBase = uint32(0x1e220000)
	)
	if word&^variableFields != conversionBase {
		return arm64RawScalarIntToFloat{}, false
	}

	floatBits := 0
	switch word >> 22 & 3 {
	case 0:
		floatBits = 32
	case 1:
		floatBits = 64
	case 3:
		floatBits = 16
	default:
		return arm64RawScalarIntToFloat{}, false
	}
	integerBits := 32
	if word&integerWidth != 0 {
		integerBits = 64
	}
	return arm64RawScalarIntToFloat{
		floatBits:   floatBits,
		integerBits: integerBits,
		unsigned:    word&unsignedBit != 0,
		source:      int(word>>5) & 31,
		destination: int(word) & 31,
	}, true
}

func (c *arm64Ctx) lowerRawScalarIntToFloat(form arm64RawScalarIntToFloat) error {
	source := ZR
	if form.source != 31 {
		source = Reg(fmt.Sprintf("R%d", form.source))
	}
	return c.lowerARM64ScalarConversion(arm64ScalarConversionForm{
		arm64ScalarConversionSpec: arm64ScalarConversionSpec{
			floatBits: form.floatBits, integerBits: form.integerBits, toFloat: true, unsigned: form.unsigned,
		},
		source: source, destination: Reg(fmt.Sprintf("F%d", form.destination)),
	})
}
