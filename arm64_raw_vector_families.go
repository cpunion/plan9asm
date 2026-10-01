package plan9asm

// These existing Advanced SIMD integer and crypto decoders accept only scalar
// FP/SIMD or vector registers. Their encoded register 30 is not GP R30, and
// register 31 is an ordinary vector register rather than XZR/SP. Retain each
// complete decoder's arrangement, lane and reserved-encoding validation.
var arm64RawVectorFamilies = [...]arm64RawRegisterFamily{
	arm64RawVectorOnlyForm(decodeARM64RawAES, (*arm64Ctx).lowerRawAES),
	arm64RawVectorOnlyForm(decodeARM64RawSM4, (*arm64Ctx).lowerRawSM4),
	arm64RawVectorOnlyForm(decodeARM64RawRDMA, (*arm64Ctx).lowerRawRDMA),
	arm64RawVectorOnlyForm(decodeARM64RawScalarADDP, (*arm64Ctx).lowerRawScalarADDP),
	arm64RawVectorOnlyForm(decodeARM64RawDotProduct, (*arm64Ctx).lowerRawDotProduct),
	arm64RawVectorOnlyForm(decodeARM64RawMatrixMultiply, (*arm64Ctx).lowerRawMatrixMultiply),
	arm64RawVectorOnlyForm(decodeARM64RawTableLookup, (*arm64Ctx).lowerRawTableLookup),
	arm64RawVectorOnlyForm(decodeARM64RawSQRDMULH, (*arm64Ctx).lowerRawSQRDMULH),
	arm64RawVectorOnlyForm(decodeARM64RawMixedSaturatingAdd, (*arm64Ctx).lowerRawMixedSaturatingAdd),
	arm64RawVectorOnlyForm(decodeARM64RawSQDMULH, (*arm64Ctx).lowerRawSQDMULH),
	arm64RawVectorOnlyForm(decodeARM64RawMUL, (*arm64Ctx).lowerRawMUL),
	arm64RawVectorOnlyForm(decodeARM64RawHalvingAddSub, (*arm64Ctx).lowerRawHalvingAddSub),
	arm64RawVectorOnlyForm(decodeARM64RawIntegerCompare, (*arm64Ctx).lowerRawIntegerCompare),
	arm64RawVectorOnlyForm(decodeARM64RawMLS, (*arm64Ctx).lowerRawMLS),
	arm64RawVectorOnlyForm(decodeARM64RawPairwiseAddLong, (*arm64Ctx).lowerRawPairwiseAddLong),
	arm64RawVectorOnlyForm(decodeARM64RawUMULL, (*arm64Ctx).lowerRawUMULL),
	arm64RawVectorOnlyForm(decodeARM64RawAddHighNarrow, (*arm64Ctx).lowerRawAddHighNarrow),
	arm64RawVectorOnlyForm(decodeARM64RawUZP, (*arm64Ctx).lowerRawUZP),
	arm64RawVectorOnlyForm(decodeARM64RawSaturatingShiftNarrow, (*arm64Ctx).lowerRawSaturatingShiftNarrow),
	arm64RawVectorOnlyForm(decodeARM64RawSQSHLU, (*arm64Ctx).lowerRawSQSHLU),
	arm64RawVectorOnlyForm(decodeARM64RawUSHLL, (*arm64Ctx).lowerRawUSHLL),
	arm64RawVectorOnlyForm(decodeARM64RawUMLAL, (*arm64Ctx).lowerRawUMLAL),
	arm64RawVectorOnlyForm(decodeARM64RawDUPElement, (*arm64Ctx).lowerRawDUPElement),
	arm64RawVectorOnlyForm(decodeARM64RawMLA, (*arm64Ctx).lowerRawMLA),
	arm64RawVectorOnlyForm(decodeARM64RawSHA3, (*arm64Ctx).lowerRawSHA3),
	arm64RawVectorOnlyForm(decodeARM64RawLogical, (*arm64Ctx).lowerRawLogical),
	arm64RawVectorOnlyForm(decodeARM64RawModifiedImmediate, (*arm64Ctx).lowerRawModifiedImmediate),
}

func decodeARM64RawVectorFamily(word uint32) (arm64RawRegisterFamily, arm64RawRegisterControlEffects, bool) {
	for _, family := range arm64RawVectorFamilies {
		if effects, ok := family.effects(word); ok {
			return family, effects, true
		}
	}
	return arm64RawRegisterFamily{}, arm64RawRegisterControlEffects{}, false
}
