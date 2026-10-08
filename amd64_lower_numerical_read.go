package plan9asm

import "fmt"

func (c *amd64Ctx) lowerNumericalRead(ins Instr) (bool, error) {
	if ins.x86NumericalRead == nil {
		return false, nil
	}
	read, form, valid := numericalCertifiedRead(ins, c.goarch)
	cert := ins.x86NumericalRead
	if !valid || c.sig.Name != read.Function || c.targetTriple != cert.binding.options.TargetTriple {
		return true, numericalSBContext("emitted function/profile/read certificate changed")
	}
	if c.fixedLocal == nil {
		if err := c.scalarFPResultInitialized(); err != nil {
			return true, err
		}
	} else if c.fixedLocal.binding == nil || c.fixedLocal.binding != c.sig.x86FixedLocal ||
		c.fixedLocal.binding.numericalReads == nil || c.fixedLocal.binding.numericalReads.source != cert.binding.source {
		return true, numericalSBContext("numerical read lost its independent complete fixed-local proof")
	}
	// No ptrFromSB call: only this exact typed read can consume the immutable
	// native source offset. The global remains externally mutable, and every
	// read is volatile even when its result/condition is later overwritten.
	pointer := llvmGlobal(read.Variable.NativeName)
	if read.Offset != 0 {
		address := c.newTmp()
		fmt.Fprintf(c.b, "  %%%s = getelementptr i8, ptr %s, i64 %d\n", address, pointer, read.Offset)
		pointer = "%" + address
	}
	sourceType := amd64IntegerTypeForBits(int(read.Width * 8))
	loaded := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = load volatile %s, ptr %s, align 1\n", loaded, sourceType, pointer)
	value := "%" + loaded
	if form.spec.compare {
		partner := ins.Args[form.partner]
		// Go's CMPQ imm32 is sign-extended by the actual native encoding.
		// An accepted unsigned Yi32 spelling is not a positive i64 value.
		if sourceType == I64 && partner.Kind == OpImm {
			partner.Imm = int64(int32(partner.Imm))
		}
		other, err := c.evalIntSized(partner, sourceType)
		if err != nil {
			return true, err
		}
		if form.operand == 0 {
			c.setCmpFlagsSized(sourceType, value, other)
		} else {
			c.setCmpFlagsSized(sourceType, other, value)
		}
		return true, nil
	}
	destinationType := amd64IntegerTypeForBits(int(form.spec.gpBytes * 8))
	if read.Width < form.spec.gpBytes {
		extension := "zext"
		if amd64ScalarExtensionMoveSpecs[ins.Op].signed {
			extension = "sext"
		}
		extended := c.newTmp()
		fmt.Fprintf(c.b, "  %%%s = %s %s %s to %s\n", extended, extension, sourceType, value, destinationType)
		value = "%" + extended
	}
	return true, c.storeRegSized(ins.Args[form.partner].Reg, destinationType, value)
}
