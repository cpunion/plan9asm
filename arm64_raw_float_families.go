package plan9asm

// Preserve the previous lowering precedence. Lowering, continuation provenance
// and raw-pool GP effects all consume this same typed family registration.
var arm64RawFloatFamilies = [...]arm64RawRegisterFamily{
	arm64RawVectorOnlyForm(decodeARM64RawBFloatDot, (*arm64Ctx).lowerRawBFloatDot),
	arm64RawVectorOnlyForm(decodeARM64RawBFloatMatrix, (*arm64Ctx).lowerRawBFloatMatrix),
	arm64RawVectorOnlyForm(decodeARM64RawFloatMultiplyLong, (*arm64Ctx).lowerRawFloatMultiplyLong),
	arm64RawVectorOnlyForm(decodeARM64RawCVTF, (*arm64Ctx).lowerRawCVTF),
	arm64RawVectorOnlyForm(decodeARM64RawFMLA, (*arm64Ctx).lowerRawFMLA),
	arm64RawVectorOnlyForm(decodeARM64RawHalfFMA, (*arm64Ctx).lowerRawHalfFMA),
	arm64RawVectorOnlyForm(decodeARM64RawFMULByElement, (*arm64Ctx).lowerRawFMULByElement),
	arm64RawVectorOnlyForm(decodeARM64RawScalarHalfUnary, (*arm64Ctx).lowerRawScalarHalfUnary),
	arm64RawVectorOnlyForm(decodeARM64RawScalarVectorFCVTZ, (*arm64Ctx).lowerRawScalarVectorFCVTZ),
	arm64RawVectorOnlyForm(decodeARM64RawScalarFloatBinary, (*arm64Ctx).lowerRawScalarFloatBinary),
	arm64RawVectorOnlyForm(decodeARM64RawScalarFloatCompare, (*arm64Ctx).lowerRawScalarFloatCompare),
	arm64RawVectorOnlyForm(decodeARM64RawScalarFloatSelect, (*arm64Ctx).lowerRawScalarFloatSelect),
	arm64RawVectorOnlyForm(decodeARM64RawScalarFloatImmediate, (*arm64Ctx).lowerRawScalarFloatImmediate),
	arm64RawRegisterForm(decodeARM64RawScalarIntToFloat, (*arm64Ctx).lowerRawScalarIntToFloat,
		func(form arm64RawScalarIntToFloat) arm64RawRegisterControlEffects {
			return arm64RawRegisterControlEffects{gpReads: arm64RawGPBit(form.source)}
		}),
	arm64RawRegisterForm(decodeARM64RawFixedIntToFloat, (*arm64Ctx).lowerRawFixedIntToFloat,
		func(form arm64RawFixedIntToFloat) arm64RawRegisterControlEffects {
			if form.vectorSource {
				return arm64RawRegisterControlEffects{}
			}
			return arm64RawRegisterControlEffects{gpReads: arm64RawGPBit(form.source)}
		}),
	arm64RawRegisterForm(decodeARM64RawFloatGPMove, (*arm64Ctx).lowerRawFloatGPMove,
		func(form arm64RawFloatGPMove) arm64RawRegisterControlEffects {
			if form.toFloat {
				return arm64RawRegisterControlEffects{gpReads: arm64RawGPBit(form.gpReg)}
			}
			return arm64RawRegisterControlEffects{gpWrites: arm64RawGPBit(form.gpReg)}
		}),
	arm64RawVectorOnlyForm(decodeARM64RawBFloatConvert, (*arm64Ctx).lowerRawBFloatConvert),
	arm64RawVectorOnlyForm(decodeARM64RawVectorFloatRound, (*arm64Ctx).lowerRawVectorFloatRound),
	arm64RawVectorOnlyForm(decodeARM64RawVectorFloatNarrow, (*arm64Ctx).lowerRawVectorFloatNarrow),
	arm64RawVectorOnlyForm(decodeARM64RawVectorFloatWiden, (*arm64Ctx).lowerRawVectorFloatWiden),
	arm64RawVectorOnlyForm(decodeARM64RawFloatMinMaxAcross, (*arm64Ctx).lowerRawFloatMinMaxAcross),
	arm64RawVectorOnlyForm(decodeARM64RawFADDP, (*arm64Ctx).lowerRawFADDP),
	arm64RawVectorOnlyForm(decodeARM64RawFloatPairwiseMinMax, (*arm64Ctx).lowerRawFloatPairwiseMinMax),
	arm64RawVectorOnlyForm(decodeARM64RawFloatBinary, (*arm64Ctx).lowerRawFloatBinary),
	arm64RawVectorOnlyForm(decodeARM64RawReciprocalEstimate, (*arm64Ctx).lowerRawReciprocalEstimate),
	arm64RawVectorOnlyForm(decodeARM64RawFSQRT, (*arm64Ctx).lowerRawFSQRT),
	arm64RawVectorOnlyForm(decodeARM64RawFloatAbsNeg, (*arm64Ctx).lowerRawFloatAbsNeg),
	arm64RawVectorOnlyForm(decodeARM64RawFloatCompare, (*arm64Ctx).lowerRawFloatCompare),
	arm64RawVectorOnlyForm(decodeARM64RawFloatImmediate, (*arm64Ctx).lowerRawFloatImmediate),
	arm64RawRegisterForm(decodeARM64RawScalarFCVTToInt, (*arm64Ctx).lowerRawScalarFCVTToInt,
		func(form arm64RawScalarFCVTToInt) arm64RawRegisterControlEffects {
			return arm64RawRegisterControlEffects{gpWrites: arm64RawGPBit(form.destination)}
		}),
	arm64RawVectorOnlyForm(decodeARM64RawFCVTZ, (*arm64Ctx).lowerRawFCVTZ),
}

func decodeARM64RawFloatFamily(word uint32) (arm64RawRegisterFamily, arm64RawRegisterControlEffects, bool) {
	for _, family := range arm64RawFloatFamilies {
		if effects, ok := family.effects(word); ok {
			return family, effects, true
		}
	}
	return arm64RawRegisterFamily{}, arm64RawRegisterControlEffects{}, false
}
