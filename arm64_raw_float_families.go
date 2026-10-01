package plan9asm

import "fmt"

type arm64RawFloatControlEffects struct {
	gpReads  uint32
	gpWrites uint32
}

// These families are register-only. Their concrete decoder determines whether
// a scalar register belongs to the floating/vector bank or the GP bank. There
// are no memory operations, implicit GP clobbers or control transfers here.
type arm64RawFloatFamily struct {
	effects func(uint32) (arm64RawFloatControlEffects, bool)
	lower   func(*arm64Ctx, uint32) error
}

func arm64RawFloatForm[T any](decode func(uint32) (T, bool), lower func(*arm64Ctx, T) error, effects func(T) arm64RawFloatControlEffects) arm64RawFloatFamily {
	return arm64RawFloatFamily{
		effects: func(word uint32) (arm64RawFloatControlEffects, bool) {
			form, ok := decode(word)
			if !ok {
				return arm64RawFloatControlEffects{}, false
			}
			return effects(form), true
		},
		lower: func(c *arm64Ctx, word uint32) error {
			form, ok := decode(word)
			if !ok {
				return fmt.Errorf("ARM64 raw floating family rejected its selected encoding %#08x", word)
			}
			return lower(c, form)
		},
	}
}

// Only explicitly registered floating/vector-only forms use this constructor.
// GP conversions must instead supply their typed source/destination effects.
func arm64RawFloatVectorOnly[T any](decode func(uint32) (T, bool), lower func(*arm64Ctx, T) error) arm64RawFloatFamily {
	return arm64RawFloatForm(decode, lower, func(T) arm64RawFloatControlEffects {
		return arm64RawFloatControlEffects{}
	})
}

func arm64RawFloatGPBit(index int) uint32 {
	if index >= 0 && index < 31 {
		return 1 << uint(index)
	}
	return 0 // Encoding 31 is WZR/XZR, not SP.
}

// Preserve the previous lowering precedence. Lowering, continuation provenance
// and raw-pool GP effects all consume this same typed family registration.
var arm64RawFloatFamilies = [...]arm64RawFloatFamily{
	arm64RawFloatVectorOnly(decodeARM64RawBFloatDot, (*arm64Ctx).lowerRawBFloatDot),
	arm64RawFloatVectorOnly(decodeARM64RawBFloatMatrix, (*arm64Ctx).lowerRawBFloatMatrix),
	arm64RawFloatVectorOnly(decodeARM64RawFloatMultiplyLong, (*arm64Ctx).lowerRawFloatMultiplyLong),
	arm64RawFloatVectorOnly(decodeARM64RawCVTF, (*arm64Ctx).lowerRawCVTF),
	arm64RawFloatVectorOnly(decodeARM64RawFMLA, (*arm64Ctx).lowerRawFMLA),
	arm64RawFloatVectorOnly(decodeARM64RawHalfFMA, (*arm64Ctx).lowerRawHalfFMA),
	arm64RawFloatVectorOnly(decodeARM64RawFMULByElement, (*arm64Ctx).lowerRawFMULByElement),
	arm64RawFloatVectorOnly(decodeARM64RawScalarHalfUnary, (*arm64Ctx).lowerRawScalarHalfUnary),
	arm64RawFloatVectorOnly(decodeARM64RawScalarVectorFCVTZ, (*arm64Ctx).lowerRawScalarVectorFCVTZ),
	arm64RawFloatVectorOnly(decodeARM64RawScalarFloatBinary, (*arm64Ctx).lowerRawScalarFloatBinary),
	arm64RawFloatVectorOnly(decodeARM64RawScalarFloatCompare, (*arm64Ctx).lowerRawScalarFloatCompare),
	arm64RawFloatVectorOnly(decodeARM64RawScalarFloatSelect, (*arm64Ctx).lowerRawScalarFloatSelect),
	arm64RawFloatVectorOnly(decodeARM64RawScalarFloatImmediate, (*arm64Ctx).lowerRawScalarFloatImmediate),
	arm64RawFloatForm(decodeARM64RawScalarIntToFloat, (*arm64Ctx).lowerRawScalarIntToFloat,
		func(form arm64RawScalarIntToFloat) arm64RawFloatControlEffects {
			return arm64RawFloatControlEffects{gpReads: arm64RawFloatGPBit(form.source)}
		}),
	arm64RawFloatForm(decodeARM64RawFixedIntToFloat, (*arm64Ctx).lowerRawFixedIntToFloat,
		func(form arm64RawFixedIntToFloat) arm64RawFloatControlEffects {
			if form.vectorSource {
				return arm64RawFloatControlEffects{}
			}
			return arm64RawFloatControlEffects{gpReads: arm64RawFloatGPBit(form.source)}
		}),
	arm64RawFloatForm(decodeARM64RawFloatGPMove, (*arm64Ctx).lowerRawFloatGPMove,
		func(form arm64RawFloatGPMove) arm64RawFloatControlEffects {
			if form.toFloat {
				return arm64RawFloatControlEffects{gpReads: arm64RawFloatGPBit(form.gpReg)}
			}
			return arm64RawFloatControlEffects{gpWrites: arm64RawFloatGPBit(form.gpReg)}
		}),
	arm64RawFloatVectorOnly(decodeARM64RawBFloatConvert, (*arm64Ctx).lowerRawBFloatConvert),
	arm64RawFloatVectorOnly(decodeARM64RawVectorFloatRound, (*arm64Ctx).lowerRawVectorFloatRound),
	arm64RawFloatVectorOnly(decodeARM64RawVectorFloatNarrow, (*arm64Ctx).lowerRawVectorFloatNarrow),
	arm64RawFloatVectorOnly(decodeARM64RawVectorFloatWiden, (*arm64Ctx).lowerRawVectorFloatWiden),
	arm64RawFloatVectorOnly(decodeARM64RawFloatMinMaxAcross, (*arm64Ctx).lowerRawFloatMinMaxAcross),
	arm64RawFloatVectorOnly(decodeARM64RawFADDP, (*arm64Ctx).lowerRawFADDP),
	arm64RawFloatVectorOnly(decodeARM64RawFloatPairwiseMinMax, (*arm64Ctx).lowerRawFloatPairwiseMinMax),
	arm64RawFloatVectorOnly(decodeARM64RawFloatBinary, (*arm64Ctx).lowerRawFloatBinary),
	arm64RawFloatVectorOnly(decodeARM64RawReciprocalEstimate, (*arm64Ctx).lowerRawReciprocalEstimate),
	arm64RawFloatVectorOnly(decodeARM64RawFSQRT, (*arm64Ctx).lowerRawFSQRT),
	arm64RawFloatVectorOnly(decodeARM64RawFloatAbsNeg, (*arm64Ctx).lowerRawFloatAbsNeg),
	arm64RawFloatVectorOnly(decodeARM64RawFloatCompare, (*arm64Ctx).lowerRawFloatCompare),
	arm64RawFloatVectorOnly(decodeARM64RawFloatImmediate, (*arm64Ctx).lowerRawFloatImmediate),
	arm64RawFloatForm(decodeARM64RawScalarFCVTToInt, (*arm64Ctx).lowerRawScalarFCVTToInt,
		func(form arm64RawScalarFCVTToInt) arm64RawFloatControlEffects {
			return arm64RawFloatControlEffects{gpWrites: arm64RawFloatGPBit(form.destination)}
		}),
	arm64RawFloatVectorOnly(decodeARM64RawFCVTZ, (*arm64Ctx).lowerRawFCVTZ),
}

func decodeARM64RawFloatFamily(word uint32) (arm64RawFloatFamily, arm64RawFloatControlEffects, bool) {
	for _, family := range arm64RawFloatFamilies {
		if effects, ok := family.effects(word); ok {
			return family, effects, true
		}
	}
	return arm64RawFloatFamily{}, arm64RawFloatControlEffects{}, false
}
