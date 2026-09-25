package plan9asm

import (
	"fmt"
	"strings"
)

var arm64SVEMultiplyHighIntrinsics = map[Op]string{
	"ZSMULH": "smulh",
	"ZUMULH": "umulh",
}

func arm64SVEMultiplyHighNeedsSVE2(ins Instr) bool {
	return len(ins.Args) == 3
}

func (c *arm64Ctx) lowerARM64SVEMultiplyHigh(op Op, ins Instr) (ok bool, terminated bool, err error) {
	intrinsic, ok := arm64SVEMultiplyHighIntrinsics[op]
	if !ok {
		return false, false, nil
	}
	if strings.ToUpper(string(ins.Op)) != string(op) || (len(ins.Args) != 3 && len(ins.Args) != 4) {
		return true, false, fmt.Errorf("arm64 %s expects one complete Go 1.27 vector form without a suffix: %q", op, ins.Raw)
	}
	second, secondBits, secondOK := arm64ParseSVEZElementReg(ins.Args[0])
	first, firstBits, firstOK := arm64ParseSVEZElementReg(ins.Args[1])
	if !secondOK || !firstOK || secondBits != firstBits {
		return true, false, fmt.Errorf("arm64 %s sources must use one B/H/S/D width: %q", op, ins.Raw)
	}
	destinationIndex := 2
	predicate := 0
	predicated := len(ins.Args) == 4
	if predicated {
		var predicateOK bool
		predicate, predicateOK = arm64ParseSVEPredicateMode(ins.Args[2], "M", 7)
		if !predicateOK {
			return true, false, fmt.Errorf("arm64 %s predicated form requires P0..P7.M: %q", op, ins.Raw)
		}
		destinationIndex = 3
	}
	destination, destinationBits, destinationOK := arm64ParseSVEZElementReg(ins.Args[destinationIndex])
	if !destinationOK || destinationBits != firstBits || (predicated && destination != first) {
		return true, false, fmt.Errorf("arm64 %s destination width and destructive operands do not match: %q", op, ins.Raw)
	}
	firstValue, vectorType, err := c.loadZRegElements(first, firstBits)
	if err != nil {
		return true, false, err
	}
	secondValue, _, err := c.loadZRegElements(second, firstBits)
	if err != nil {
		return true, false, err
	}
	var predicateValue, predicateType string
	if predicated {
		predicateValue, predicateType, err = c.loadPRegElements(predicate, firstBits)
	} else {
		predicateValue, predicateType, err = c.allTruePRegElements(firstBits)
		intrinsic += ".u"
	}
	if err != nil {
		return true, false, err
	}
	_, lanes, err := arm64SVEVectorType(firstBits)
	if err != nil {
		return true, false, err
	}
	result := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = call %s @llvm.aarch64.sve.%s.nxv%di%d(%s %s, %s %s, %s %s)\n",
		result, vectorType, intrinsic, lanes, firstBits, predicateType, predicateValue, vectorType, firstValue, vectorType, secondValue)
	return true, false, c.storeZRegElements(destination, firstBits, "%"+result)
}
