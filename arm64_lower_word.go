package plan9asm

import "fmt"

// lowerRawWord models raw encodings only when their register and flag effects
// are known. Silently emitting nothing for an unknown machine instruction
// would report success while changing program semantics.
func (c *arm64Ctx) lowerRawWord(ins Instr) error {
	if len(ins.Args) != 1 || ins.Args[0].Kind != OpImm || ins.Args[0].ImmRaw != "" {
		return fmt.Errorf("arm64 WORD expects exactly one resolved integer constant: %q", ins.Raw)
	}

	word := uint32(ins.Args[0].Imm)
	if family, _, ok := decodeARM64RawFloatFamily(word); ok {
		return family.lower(c, word)
	}
	if form, ok := decodeARM64RawException(word); ok {
		return c.lowerRawException(form)
	}
	if form, ok := decodeARM64RawTLBI(word); ok {
		return c.lowerRawTLBI(form)
	}
	if form, ok := decodeARM64RawStreamingModeControl(word); ok {
		return c.lowerRawStreamingModeControl(form)
	}
	if mask, ok := decodeARM64RawZAZero(word); ok {
		return c.lowerRawZAZero(mask)
	}
	if form, ok := decodeARM64RawSMEOuterProduct(word); ok {
		return c.lowerRawSMEOuterProduct(form)
	}
	if form, ok := decodeARM64RawSMETileRead(word); ok {
		return c.lowerRawSMETileRead(form)
	}
	if form, ok := decodeARM64RawSMETileWrite(word); ok {
		return c.lowerRawSMETileWrite(form)
	}
	if form, ok := decodeARM64RawSMETileMemory(word); ok {
		return c.lowerRawSMETileMemory(form)
	}
	if family, decoded, ok := decodeARM64RawSVEFamily(word); ok {
		return family.lower(c, decoded)
	}
	for _, family := range arm64RawSVEVectorFamilies {
		if family.matches(word) {
			return family.lower(c, word)
		}
	}
	if reg, ok := decodeARM64RawICIVAU(word); ok {
		return c.lowerRawICIVAU(reg)
	}
	if _, reservedLoad := arm64RawSVEUnsignedLoadRows[word&0xffe0e000]; reservedLoad {
		return fmt.Errorf("reserved ARM64 SVE load encoding %#08x: %q", word, ins.Raw)
	}
	if form, ok := decodeARM64RawSVELD1B(word); ok {
		return c.lowerRawSVELD1B(form)
	}
	if form, ok := decodeARM64RawRNDR(word); ok {
		return c.lowerRawRNDR(form)
	}
	if form, ok := decodeARM64RawSystemRegister(word); ok {
		return c.lowerRawSystemRegister(form)
	}
	if immediate, ok := decodeARM64RawHint(word); ok {
		c.emitARM64Hint(immediate)
		return nil
	}
	if form, ok := decodeARM64RawCASP(word); ok {
		return c.lowerRawCASP(form)
	}
	if form, ok := decodeARM64RawAES(word); ok {
		return c.lowerRawAES(form)
	}
	if form, ok := decodeARM64RawSM4(word); ok {
		return c.lowerRawSM4(form)
	}
	if form, ok := decodeARM64RawRDMA(word); ok {
		return c.lowerRawRDMA(form)
	}
	if form, ok := decodeARM64RawScalarADDP(word); ok {
		return c.lowerRawScalarADDP(form)
	}
	if form, ok := decodeARM64RawDotProduct(word); ok {
		return c.lowerRawDotProduct(form)
	}
	if form, ok := decodeARM64RawMatrixMultiply(word); ok {
		return c.lowerRawMatrixMultiply(form)
	}
	if form, ok := decodeARM64RawTableLookup(word); ok {
		return c.lowerRawTableLookup(form)
	}
	if form, ok := decodeARM64RawSQRDMULH(word); ok {
		return c.lowerRawSQRDMULH(form)
	}
	if form, ok := decodeARM64RawMixedSaturatingAdd(word); ok {
		return c.lowerRawMixedSaturatingAdd(form)
	}
	if form, ok := decodeARM64RawSQDMULH(word); ok {
		return c.lowerRawSQDMULH(form)
	}
	if form, ok := decodeARM64RawMUL(word); ok {
		return c.lowerRawMUL(form)
	}
	if form, ok := decodeARM64RawHalvingAddSub(word); ok {
		return c.lowerRawHalvingAddSub(form)
	}
	if form, ok := decodeARM64RawIntegerCompare(word); ok {
		return c.lowerRawIntegerCompare(form)
	}
	if form, ok := decodeARM64RawMLS(word); ok {
		return c.lowerRawMLS(form)
	}
	if form, ok := decodeARM64RawPairwiseAddLong(word); ok {
		return c.lowerRawPairwiseAddLong(form)
	}
	if form, ok := decodeARM64RawUMULL(word); ok {
		return c.lowerRawUMULL(form)
	}
	if form, ok := decodeARM64RawAddHighNarrow(word); ok {
		return c.lowerRawAddHighNarrow(form)
	}
	if form, ok := decodeARM64RawUZP(word); ok {
		return c.lowerRawUZP(form)
	}
	if form, ok := decodeARM64RawSaturatingShiftNarrow(word); ok {
		return c.lowerRawSaturatingShiftNarrow(form)
	}
	if form, ok := decodeARM64RawSQSHLU(word); ok {
		return c.lowerRawSQSHLU(form)
	}
	if form, ok := decodeARM64RawUSHLL(word); ok {
		return c.lowerRawUSHLL(form)
	}
	if form, ok := decodeARM64RawStructureLane(word); ok {
		return c.lowerRawStructureLane(form)
	}
	if form, ok := decodeARM64RawLDnR(word); ok {
		return c.lowerRawLDnR(form)
	}
	if form, ok := decodeARM64RawUMLAL(word); ok {
		return c.lowerRawUMLAL(form)
	}
	if form, ok := decodeARM64RawDUPElement(word); ok {
		return c.lowerRawDUPElement(form)
	}
	if form, ok := decodeARM64RawMLA(word); ok {
		return c.lowerRawMLA(form)
	}
	if form, ok := decodeARM64RawSHA3(word); ok {
		return c.lowerRawSHA3(form)
	}
	if form, ok := decodeARM64RawLogical(word); ok {
		return c.lowerRawLogical(form)
	}
	if form, ok := decodeARM64RawModifiedImmediate(word); ok {
		return c.lowerRawModifiedImmediate(form)
	}
	if form, ok := decodeARM64RawMoveWide(word); ok {
		return c.lowerRawMoveWide(form)
	}
	if form, ok := decodeARM64RawSVEAddress(word); ok {
		return c.lowerRawSVEAddress(form)
	}
	if form, ok := decodeARM64RawSVECnt(word); ok {
		return c.lowerRawSVECnt(form)
	}
	if form, ok := decodeARM64RawBitfield(word); ok {
		return c.lowerRawBitfield(form)
	}
	if form, ok := decodeARM64RawSVEPTrue(word); ok {
		return c.lowerRawSVEPTrue(form)
	}
	if form, ok := decodeARM64RawSVELDST1D(word); ok {
		return c.lowerRawSVELDST1D(form)
	}
	if form, ok := decodeARM64RawSVELDST1W(word); ok {
		return c.lowerRawSVELDST1W(form)
	}
	if reduction, ok := decodeARM64RawSVEIntegerAddReduction(word); ok {
		return c.lowerARM64SVEAddReductionForm(
			reduction.spec,
			reduction.form,
			reduction.destination,
		)
	}
	if form, ok := decodeARM64RawSVEFloat(word); ok {
		return c.lowerRawSVEFloat(form)
	}
	if form, ok := decodeARM64RawSVELoadStore(word); ok {
		return c.lowerRawSVELoadStore(form)
	}
	if form, ok := decodeARM64RawSVEDupImmediate(word); ok {
		return c.lowerRawSVEDupImmediate(form)
	}
	if form, ok := decodeARM64RawSVEDupGeneral(word); ok {
		return c.lowerRawSVEDupGeneral(form)
	}
	if form, ok := decodeARM64RawSVEDupElement(word); ok {
		return c.lowerRawSVEDupElement(form)
	}
	if form, ok := decodeARM64RawSVEAdd(word); ok {
		return c.lowerRawSVEAdd(form)
	}
	if op, form, ok := decodeARM64RawSVEAddSub(word); ok {
		return c.lowerRawSVEAddSub(arm64SVEAddSubSpecs[op], form)
	}
	if form, ok := decodeARM64RawSVEShift(word); ok {
		return c.lowerRawSVELSR(form)
	}
	if form, ok := decodeARM64RawSVESelect(word); ok {
		return c.lowerRawSVESelect(form)
	}
	if form, ok := decodeARM64RawSVEEOR(word); ok {
		return c.lowerRawSVEEOR(form)
	}
	if form, ok := decodeARM64RawSVEPermute(word); ok {
		return c.lowerRawSVEPermute(form)
	}
	switch {
	case word == 0xea00001f: // TST X0, X0 (ANDS XZR, X0, X0)
		value, err := c.loadReg("R0")
		if err != nil {
			return err
		}
		c.setFlagsLogic(value)
		return nil

	case word&0xff800000 == 0xf1000000: // SUBS Xd, Xn, #imm{, LSL #12}
		rn := (word >> 5) & 31
		rd := word & 31
		imm := int64((word >> 10) & 0xfff)
		if word&(1<<22) != 0 {
			imm <<= 12
		}

		srcReg := Reg(fmt.Sprintf("R%d", rn))
		if rn == 31 {
			srcReg = SP
		}
		src, err := c.loadReg(srcReg)
		if err != nil {
			return err
		}

		resTmp := c.newTmp()
		fmt.Fprintf(c.b, "  %%%s = sub i64 %s, %d\n", resTmp, src, imm)
		res := "%" + resTmp
		if rd != 31 {
			if err := c.storeReg(Reg(fmt.Sprintf("R%d", rd)), res); err != nil {
				return err
			}
		}
		c.setFlagsSub(src, fmt.Sprintf("%d", imm), res)
		return nil
	}

	return fmt.Errorf("unsupported ARM64 WORD encoding %#08x: %q", word, ins.Raw)
}
