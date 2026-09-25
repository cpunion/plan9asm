package plan9asm

import (
	"fmt"
	"strings"
)

var arm64SVEMultiplyAccumulateIntrinsics = map[Op]string{
	"ZMLA": "mla",
	"ZMLS": "mls",
}

type arm64SVEMultiplyAccumulateForm struct {
	intrinsic    string
	elementBits  int
	multiplier   int
	multiplicand int
	predicate    int
	lane         int
	destination  int
	indexed      bool
}

func arm64SVEMultiplyAccumulateNeedsSVE2(ins Instr) bool {
	return len(ins.Args) == 3
}

func (c *arm64Ctx) lowerARM64SVEMultiplyAccumulate(op Op, ins Instr) (ok bool, terminated bool, err error) {
	intrinsic, ok := arm64SVEMultiplyAccumulateIntrinsics[op]
	if !ok {
		return false, false, nil
	}
	if strings.ToUpper(string(ins.Op)) != string(op) || (len(ins.Args) != 3 && len(ins.Args) != 4) {
		return true, false, fmt.Errorf("arm64 %s expects one complete Go 1.27 multiply-accumulate form without a suffix: %q", op, ins.Raw)
	}
	form := arm64SVEMultiplyAccumulateForm{intrinsic: intrinsic}
	if len(ins.Args) == 4 {
		multiplier, multiplierBits, multiplierOK := arm64ParseSVEZElementReg(ins.Args[0])
		multiplicand, multiplicandBits, multiplicandOK := arm64ParseSVEZElementReg(ins.Args[1])
		predicate, predicateOK := arm64ParseSVEPredicateMode(ins.Args[2], "M", 7)
		destination, destinationBits, destinationOK := arm64ParseSVEZElementReg(ins.Args[3])
		if !multiplierOK || !multiplicandOK || !predicateOK || !destinationOK || multiplierBits != multiplicandBits || multiplicandBits != destinationBits {
			return true, false, fmt.Errorf("arm64 %s predicated operands must use one B/H/S/D width and P0..P7.M: %q", op, ins.Raw)
		}
		form.elementBits = multiplierBits
		form.multiplier = multiplier
		form.multiplicand = multiplicand
		form.predicate = predicate
		form.destination = destination
	} else {
		multiplier, multiplierBits, lane, multiplierOK := arm64ParseSVEZIndexedElementReg(ins.Args[0])
		multiplicand, multiplicandBits, multiplicandOK := arm64ParseSVEZElementReg(ins.Args[1])
		destination, destinationBits, destinationOK := arm64ParseSVEZElementReg(ins.Args[2])
		if !multiplierOK || !multiplicandOK || !destinationOK || multiplierBits != multiplicandBits || multiplicandBits != destinationBits {
			return true, false, fmt.Errorf("arm64 %s indexed operands must use one H/S/D width and an encodable lane: %q", op, ins.Raw)
		}
		form.elementBits = multiplierBits
		form.multiplier = multiplier
		form.multiplicand = multiplicand
		form.lane = lane
		form.destination = destination
		form.indexed = true
	}
	return true, false, c.lowerARM64SVEMultiplyAccumulateForm(form)
}

func (c *arm64Ctx) lowerARM64SVEMultiplyAccumulateForm(form arm64SVEMultiplyAccumulateForm) error {
	accumulator, vectorType, err := c.loadZRegElements(form.destination, form.elementBits)
	if err != nil {
		return err
	}
	multiplicand, _, err := c.loadZRegElements(form.multiplicand, form.elementBits)
	if err != nil {
		return err
	}
	multiplier, _, err := c.loadZRegElements(form.multiplier, form.elementBits)
	if err != nil {
		return err
	}
	_, lanes, err := arm64SVEVectorType(form.elementBits)
	if err != nil {
		return err
	}
	result := c.newTmp()
	if form.indexed {
		fmt.Fprintf(c.b, "  %%%s = call %s @llvm.aarch64.sve.%s.lane.nxv%di%d(%s %s, %s %s, %s %s, i32 %d)\n",
			result, vectorType, form.intrinsic, lanes, form.elementBits, vectorType, accumulator, vectorType, multiplicand, vectorType, multiplier, form.lane)
	} else {
		predicate, predicateType, err := c.loadPRegElements(form.predicate, form.elementBits)
		if err != nil {
			return err
		}
		fmt.Fprintf(c.b, "  %%%s = call %s @llvm.aarch64.sve.%s.nxv%di%d(%s %s, %s %s, %s %s, %s %s)\n",
			result, vectorType, form.intrinsic, lanes, form.elementBits, predicateType, predicate, vectorType, accumulator, vectorType, multiplicand, vectorType, multiplier)
	}
	return c.storeZRegElements(form.destination, form.elementBits, "%"+result)
}
