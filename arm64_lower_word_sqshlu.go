package plan9asm

import "fmt"

// Go 1.27 has no named fixed-width VSQSHLU optab row, but WORD accepts the
// architectural scalar and vector encodings. Keep this decoder raw-only so a
// named spelling rejected by Go is not incorrectly advertised as supported.
type arm64RawSQSHLU struct {
	scalar      bool
	arrangement arm64VectorArrangement
	shift       int
	source      int
	destination int
}

func decodeARM64RawSQSHLU(word uint32) (arm64RawSQSHLU, bool) {
	scalar := false
	switch {
	case word&0xff80fc00 == 0x7f006400:
		scalar = true
	case word&0xbf80fc00 == 0x2f006400:
	default:
		return arm64RawSQSHLU{}, false
	}

	encodedImmediate := int(word>>16) & 0x7f
	elementBits := 0
	switch {
	case encodedImmediate >= 64:
		elementBits = 64
	case encodedImmediate >= 32:
		elementBits = 32
	case encodedImmediate >= 16:
		elementBits = 16
	case encodedImmediate >= 8:
		elementBits = 8
	default:
		return arm64RawSQSHLU{}, false
	}
	shift := encodedImmediate - elementBits
	if shift < 0 || shift >= elementBits {
		return arm64RawSQSHLU{}, false
	}

	lanes := 1
	if !scalar {
		vectorBits := 64
		if word&(1<<30) != 0 {
			vectorBits = 128
		}
		if vectorBits < elementBits*2 {
			return arm64RawSQSHLU{}, false
		}
		lanes = vectorBits / elementBits
	}
	return arm64RawSQSHLU{
		scalar:      scalar,
		arrangement: arm64VectorArrangement{elementBits: elementBits, lanes: lanes},
		shift:       shift,
		source:      int(word>>5) & 31,
		destination: int(word) & 31,
	}, true
}

func (c *arm64Ctx) lowerRawSQSHLU(form arm64RawSQSHLU) error {
	value, err := c.loadRawARM64VectorOperand(form.source, form.arrangement, 0, form.scalar)
	if err != nil {
		return err
	}
	arrangement := form.arrangement
	vectorType := fmt.Sprintf("<%d x i%d>", arrangement.lanes, arrangement.elementBits)
	shift := arm64VectorIntegerSplat(arrangement, int64(form.shift))
	result := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = call %s @llvm.aarch64.neon.sqshlu.v%di%d(%s %s, %s %s)\n",
		result, vectorType, arrangement.lanes, arrangement.elementBits,
		vectorType, value, vectorType, shift)
	return c.storeRawARM64VectorResult(form.destination, arrangement, "%"+result, form.scalar)
}
