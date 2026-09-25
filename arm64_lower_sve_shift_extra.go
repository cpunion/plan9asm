package plan9asm

import (
	"fmt"
	"strings"
)

type arm64SVEImmediateShiftSpec struct {
	intrinsic  string
	predicated bool
	minimum    int64
	sve2       bool
}

var arm64SVEImmediateShiftSpecs = map[Op]arm64SVEImmediateShiftSpec{
	"ZASRD":   {intrinsic: "asrd", predicated: true, minimum: 1},
	"ZSLI":    {intrinsic: "sli", sve2: true},
	"ZSQSHLU": {intrinsic: "sqshlu", predicated: true, sve2: true},
	"ZSRI":    {intrinsic: "sri", minimum: 1, sve2: true},
	"ZSRSHR":  {intrinsic: "srshr", predicated: true, minimum: 1, sve2: true},
	"ZSRSRA":  {intrinsic: "srsra", minimum: 1, sve2: true},
	"ZSSRA":   {intrinsic: "ssra", minimum: 1, sve2: true},
	"ZURSHR":  {intrinsic: "urshr", predicated: true, minimum: 1, sve2: true},
	"ZURSRA":  {intrinsic: "ursra", minimum: 1, sve2: true},
	"ZUSRA":   {intrinsic: "usra", minimum: 1, sve2: true},
}

var arm64SVEReverseShiftIntrinsics = map[Op]string{
	"ZASRR": "asr",
	"ZLSLR": "lsl",
	"ZLSRR": "lsr",
}

func (c *arm64Ctx) lowerARM64SVEExtraShift(op Op, ins Instr) (ok bool, terminated bool, err error) {
	if spec, ok := arm64SVEImmediateShiftSpecs[op]; ok {
		return c.lowerARM64SVEExtraImmediateShift(op, ins, spec)
	}
	intrinsic, ok := arm64SVEReverseShiftIntrinsics[op]
	if !ok {
		return false, false, nil
	}
	if strings.ToUpper(string(ins.Op)) != string(op) || len(ins.Args) != 4 {
		return true, false, fmt.Errorf("arm64 %s expects Zm.T, Zdn.T, Pg/M, Zdn.T without a suffix: %q", op, ins.Raw)
	}
	data, dataBits, dataOK := arm64ParseSVEZElementReg(ins.Args[0])
	oldDestination, oldBits, oldOK := arm64ParseSVEZElementReg(ins.Args[1])
	predicate, predicateOK := arm64ParseSVEPredicateMerge(ins.Args[2])
	destination, destinationBits, destinationOK := arm64ParseSVEZElementReg(ins.Args[3])
	if !dataOK || !oldOK || !predicateOK || !destinationOK || dataBits != oldBits || oldBits != destinationBits || oldDestination != destination {
		return true, false, fmt.Errorf("arm64 %s requires matching vectors and a repeated destructive destination: %q", op, ins.Raw)
	}
	dataValue, vectorType, err := c.loadZRegElements(data, dataBits)
	if err != nil {
		return true, false, err
	}
	oldValue, _, err := c.loadZRegElements(destination, dataBits)
	if err != nil {
		return true, false, err
	}
	predicateValue, predicateType, err := c.loadPRegElements(predicate, dataBits)
	if err != nil {
		return true, false, err
	}
	allTrue, _, err := c.allTruePRegElements(dataBits)
	if err != nil {
		return true, false, err
	}
	_, lanes, _ := arm64SVEVectorType(dataBits)
	active := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = call %s @llvm.aarch64.sve.%s.nxv%di%d(%s %s, %s %s, %s %s)\n", active, vectorType, intrinsic, lanes, dataBits, predicateType, allTrue, vectorType, dataValue, vectorType, oldValue)
	result := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = select %s %s, %s %%%s, %s %s\n", result, predicateType, predicateValue, vectorType, active, vectorType, oldValue)
	return true, false, c.storeZRegElements(destination, dataBits, "%"+result)
}

func (c *arm64Ctx) lowerARM64SVEExtraImmediateShift(op Op, ins Instr, spec arm64SVEImmediateShiftSpec) (ok bool, terminated bool, err error) {
	wantArgs := 3
	if spec.predicated {
		wantArgs = 4
	}
	if strings.ToUpper(string(ins.Op)) != string(op) || len(ins.Args) != wantArgs || ins.Args[0].Kind != OpImm || ins.Args[0].ImmRaw != "" {
		return true, false, fmt.Errorf("arm64 %s expects its Go 1.27 immediate form without a suffix: %q", op, ins.Raw)
	}
	source, sourceBits, sourceOK := arm64ParseSVEZElementReg(ins.Args[1])
	destinationIndex := wantArgs - 1
	destination, destinationBits, destinationOK := arm64ParseSVEZElementReg(ins.Args[destinationIndex])
	shift := ins.Args[0].Imm
	if !sourceOK || !destinationOK || sourceBits != destinationBits || shift < spec.minimum || shift >= int64(sourceBits) {
		return true, false, fmt.Errorf("arm64 %s immediate is outside Go 1.27's element-width range: %q", op, ins.Raw)
	}
	var predicateValue, predicateType string
	if spec.predicated {
		predicate, predicateOK := arm64ParseSVEPredicateMerge(ins.Args[2])
		if !predicateOK || source != destination {
			return true, false, fmt.Errorf("arm64 %s requires a repeated destructive destination and Pg/M: %q", op, ins.Raw)
		}
		predicateValue, predicateType, err = c.loadPRegElements(predicate, sourceBits)
		if err != nil {
			return true, false, err
		}
	}
	sourceValue, vectorType, err := c.loadZRegElements(source, sourceBits)
	if err != nil {
		return true, false, err
	}
	_, lanes, _ := arm64SVEVectorType(sourceBits)
	result := c.newTmp()
	if spec.predicated {
		fmt.Fprintf(c.b, "  %%%s = call %s @llvm.aarch64.sve.%s.nxv%di%d(%s %s, %s %s, i32 %d)\n", result, vectorType, spec.intrinsic, lanes, sourceBits, predicateType, predicateValue, vectorType, sourceValue, shift)
	} else {
		oldDestination, _, err := c.loadZRegElements(destination, sourceBits)
		if err != nil {
			return true, false, err
		}
		fmt.Fprintf(c.b, "  %%%s = call %s @llvm.aarch64.sve.%s.nxv%di%d(%s %s, %s %s, i32 %d)\n", result, vectorType, spec.intrinsic, lanes, sourceBits, vectorType, oldDestination, vectorType, sourceValue, shift)
	}
	return true, false, c.storeZRegElements(destination, sourceBits, "%"+result)
}
