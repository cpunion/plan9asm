package plan9asm

import (
	"fmt"
	"strings"
)

var arm64SVEMADMSBIntrinsics = map[Op]string{
	"ZMAD": "mad",
	"ZMSB": "msb",
}

type arm64SVEMADMSBForm struct {
	intrinsic   string
	elementBits int
	addend      int
	multiplier  int
	predicate   int
	destination int
}

func (c *arm64Ctx) lowerARM64SVEMADMSB(op Op, ins Instr) (ok bool, terminated bool, err error) {
	intrinsic, ok := arm64SVEMADMSBIntrinsics[op]
	if !ok {
		return false, false, nil
	}
	if strings.ToUpper(string(ins.Op)) != string(op) || len(ins.Args) != 4 {
		return true, false, fmt.Errorf("arm64 %s expects one complete Go 1.27 SVE MAD/MSB form without a suffix: %q", op, ins.Raw)
	}
	addend, addendBits, addendOK := arm64ParseSVEZElementReg(ins.Args[0])
	multiplier, multiplierBits, multiplierOK := arm64ParseSVEZElementReg(ins.Args[1])
	predicate, predicateOK := arm64ParseSVEPredicateMode(ins.Args[2], "M", 7)
	destination, destinationBits, destinationOK := arm64ParseSVEZElementReg(ins.Args[3])
	if !addendOK || !multiplierOK || !predicateOK || !destinationOK || addendBits != multiplierBits || multiplierBits != destinationBits {
		return true, false, fmt.Errorf("arm64 %s operands must use one B/H/S/D width and P0..P7.M: %q", op, ins.Raw)
	}
	form := arm64SVEMADMSBForm{
		intrinsic:   intrinsic,
		elementBits: addendBits,
		addend:      addend,
		multiplier:  multiplier,
		predicate:   predicate,
		destination: destination,
	}
	return true, false, c.lowerARM64SVEMADMSBForm(form)
}

func (c *arm64Ctx) lowerARM64SVEMADMSBForm(form arm64SVEMADMSBForm) error {
	destination, vectorType, err := c.loadZRegElements(form.destination, form.elementBits)
	if err != nil {
		return err
	}
	multiplier, _, err := c.loadZRegElements(form.multiplier, form.elementBits)
	if err != nil {
		return err
	}
	addend, _, err := c.loadZRegElements(form.addend, form.elementBits)
	if err != nil {
		return err
	}
	predicate, predicateType, err := c.loadPRegElements(form.predicate, form.elementBits)
	if err != nil {
		return err
	}
	_, lanes, err := arm64SVEVectorType(form.elementBits)
	if err != nil {
		return err
	}
	result := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = call %s @llvm.aarch64.sve.%s.nxv%di%d(%s %s, %s %s, %s %s, %s %s)\n",
		result, vectorType, form.intrinsic, lanes, form.elementBits, predicateType, predicate, vectorType, destination, vectorType, multiplier, vectorType, addend)
	return c.storeZRegElements(form.destination, form.elementBits, "%"+result)
}
