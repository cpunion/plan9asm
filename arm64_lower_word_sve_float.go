package plan9asm

import "fmt"

type arm64RawSVEFloatKind uint8

const (
	arm64RawSVEFloatAdd arm64RawSVEFloatKind = iota
	arm64RawSVEFloatSub
	arm64RawSVEFloatMul
	arm64RawSVEFloatFMLA
	arm64RawSVEFloatFMAD
	arm64RawSVEFloatFADDV
	arm64RawSVEFloatFADDA
)

type arm64RawSVEFloat struct {
	kind        arm64RawSVEFloatKind
	elementBits int
	destination int
	first       int
	second      int
	predicate   int
	predicated  bool
}

func (c *arm64Ctx) loadRawSVEFloatVector(index, elementBits int) (string, string, error) {
	bytes, err := c.loadZReg(index)
	if err != nil {
		return "", "", err
	}
	floatingType := map[int]string{16: "half", 32: "float", 64: "double"}[elementBits]
	if floatingType == "" {
		return "", "", fmt.Errorf("unsupported ARM64 SVE floating element width %d", elementBits)
	}
	vectorType := fmt.Sprintf("<vscale x %d x %s>", 128/elementBits, floatingType)
	converted := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = bitcast <vscale x 16 x i8> %s to %s\n", converted, bytes, vectorType)
	return "%" + converted, vectorType, nil
}

func (c *arm64Ctx) storeRawSVEFloatVector(index, elementBits int, value string, vectorType string) error {
	bytes := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = bitcast %s %s to <vscale x 16 x i8>\n", bytes, vectorType, value)
	return c.storeZReg(index, "%"+bytes)
}

// decodeARM64RawSVEFloat covers the scalar-vector SVE floating arithmetic
// encodings emitted by Go's generated assembly. The destructive predicated
// forms keep Zd as their first operand; FMLA/FMAD use the two encoded source
// vectors and accumulate into Zd.
func decodeARM64RawSVEFloat(word uint32) (arm64RawSVEFloat, bool) {
	form := arm64RawSVEFloat{elementBits: 8 << (int(word>>22) & 3)}
	switch {
	case word&0xff20fc00 == 0x65000000:
		form.kind = arm64RawSVEFloatAdd
		form.first = int(word>>5) & 31
		form.second = int(word>>16) & 31
		form.destination = int(word) & 31
	case word&0xff20fc00 == 0x65000400:
		form.kind = arm64RawSVEFloatSub
		form.first = int(word>>5) & 31
		form.second = int(word>>16) & 31
		form.destination = int(word) & 31
	case word&0xff20fc00 == 0x65000800:
		form.kind = arm64RawSVEFloatMul
		form.first = int(word>>5) & 31
		form.second = int(word>>16) & 31
		form.destination = int(word) & 31
	case word&0xff20e000 == 0x65008000:
		form.kind = arm64RawSVEFloatAdd
		form.predicated = true
		form.first = int(word>>5) & 31
		form.destination = int(word) & 31
		form.predicate = int(word>>10) & 7
	case word&0xff20e000 == 0x65008400:
		form.kind = arm64RawSVEFloatSub
		form.predicated = true
		form.first = int(word>>5) & 31
		form.destination = int(word) & 31
		form.predicate = int(word>>10) & 7
	case word&0xff20e000 == 0x65008800:
		form.kind = arm64RawSVEFloatMul
		form.predicated = true
		form.first = int(word>>5) & 31
		form.destination = int(word) & 31
		form.predicate = int(word>>10) & 7
	case word&0xff20e000 == 0x65200000:
		form.kind = arm64RawSVEFloatFMLA
		form.predicated = true
		form.first = int(word>>5) & 31
		form.second = int(word>>16) & 31
		form.destination = int(word) & 31
		form.predicate = int(word>>10) & 7
	case word&0xff20e000 == 0x65208000:
		form.kind = arm64RawSVEFloatFMAD
		form.predicated = true
		form.first = int(word>>5) & 31
		form.second = int(word>>16) & 31
		form.destination = int(word) & 31
		form.predicate = int(word>>10) & 7
	case word&0xff20e000 == 0x65002000:
		form.destination = int(word) & 31
		form.predicate = int(word>>10) & 7
		if word&0x001f0000 == 0 {
			form.kind = arm64RawSVEFloatFADDV
			form.first = form.destination
		} else if word&0x001f0000 == 0x00180000 {
			form.kind = arm64RawSVEFloatFADDA
			form.first = int(word>>5) & 31
		} else {
			return arm64RawSVEFloat{}, false
		}
	default:
		return arm64RawSVEFloat{}, false
	}
	return form, true
}

func (c *arm64Ctx) lowerRawSVEFloat(form arm64RawSVEFloat) error {
	switch form.kind {
	case arm64RawSVEFloatFADDV, arm64RawSVEFloatFADDA:
		return c.lowerRawSVEFloatReduction(form)
	}
	_, _, err := arm64SVEVectorType(form.elementBits)
	if err != nil {
		return err
	}
	result := c.newTmp()
	var vectorType string
	if form.kind == arm64RawSVEFloatFMLA || form.kind == arm64RawSVEFloatFMAD {
		first, vt, err := c.loadRawSVEFloatVector(form.first, form.elementBits)
		if err != nil {
			return err
		}
		vectorType = vt
		second, _, err := c.loadRawSVEFloatVector(form.second, form.elementBits)
		if err != nil {
			return err
		}
		old, _, err := c.loadRawSVEFloatVector(form.destination, form.elementBits)
		if err != nil {
			return err
		}
		fmt.Fprintf(c.b, "  %%%s = call %s @llvm.fma.nxv%df%d(%s %s, %s %s, %s %s)\n", result, vectorType, 128/form.elementBits, form.elementBits, vectorType, first, vectorType, second, vectorType, old)
	} else {
		first, vt, err := c.loadRawSVEFloatVector(form.first, form.elementBits)
		if err != nil {
			return err
		}
		vectorType = vt
		second := first
		if !form.predicated {
			second, _, err = c.loadRawSVEFloatVector(form.second, form.elementBits)
			if err != nil {
				return err
			}
		}
		op := map[arm64RawSVEFloatKind]string{
			arm64RawSVEFloatAdd: "fadd",
			arm64RawSVEFloatSub: "fsub",
			arm64RawSVEFloatMul: "fmul",
		}[form.kind]
		if form.predicated {
			old, _, err := c.loadRawSVEFloatVector(form.destination, form.elementBits)
			if err != nil {
				return err
			}
			fmt.Fprintf(c.b, "  %%%s = %s %s %s, %s\n", result, op, vectorType, old, second)
		} else {
			fmt.Fprintf(c.b, "  %%%s = %s %s %s, %s\n", result, op, vectorType, first, second)
		}
	}
	value := "%" + result
	if form.predicated {
		predicate, predicateType, err := c.loadPRegElements(form.predicate, form.elementBits)
		if err != nil {
			return err
		}
		old, _, err := c.loadRawSVEFloatVector(form.destination, form.elementBits)
		if err != nil {
			return err
		}
		selected := c.newTmp()
		fmt.Fprintf(c.b, "  %%%s = select %s %s, %s %s, %s %s\n", selected, predicateType, predicate, vectorType, value, vectorType, old)
		value = "%" + selected
	}
	return c.storeRawSVEFloatVector(form.destination, form.elementBits, value, vectorType)
}

func (c *arm64Ctx) lowerRawSVEFloatReduction(form arm64RawSVEFloat) error {
	sourceIndex := form.destination
	if form.kind == arm64RawSVEFloatFADDA {
		sourceIndex = form.first
	}
	source, vectorType, err := c.loadZRegElements(sourceIndex, form.elementBits)
	if err != nil {
		return err
	}
	predicate, predicateType, err := c.loadPRegElements(form.predicate, form.elementBits)
	if err != nil {
		return err
	}
	masked := c.newTmp()
	floatingType := map[int]string{16: "half", 32: "float", 64: "double"}[form.elementBits]
	if floatingType == "" {
		return fmt.Errorf("unsupported ARM64 SVE floating element width %d", form.elementBits)
	}
	floatingVectorType := fmt.Sprintf("<vscale x %d x %s>", 128/form.elementBits, floatingType)
	if vectorType != floatingVectorType {
		converted := c.newTmp()
		fmt.Fprintf(c.b, "  %%%s = bitcast %s %s to %s\n", converted, vectorType, source, floatingVectorType)
		source = "%" + converted
	}
	zero := fmt.Sprintf("%s zeroinitializer", floatingVectorType)
	fmt.Fprintf(c.b, "  %%%s = select %s %s, %s %s, %s\n", masked, predicateType, predicate, floatingVectorType, source, zero)
	initial := fmt.Sprintf("0.000000e+00")
	if form.kind == arm64RawSVEFloatFADDA {
		initial, err = c.loadARM64ScalarFloatReg(Reg(fmt.Sprintf("F%d", form.destination)), form.elementBits)
		if err != nil {
			return err
		}
	}
	result := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = call %s @llvm.vector.reduce.fadd.nxv%df%d(%s %s, %s %%%s)\n", result, floatingType, 128/form.elementBits, form.elementBits, floatingType, initial, floatingVectorType, masked)
	return c.storeARM64ScalarFloatReg(Reg(fmt.Sprintf("F%d", form.destination)), form.elementBits, "%"+result)
}
