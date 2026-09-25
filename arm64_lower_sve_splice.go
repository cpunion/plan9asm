package plan9asm

import (
	"fmt"
	"strings"
)

func (c *arm64Ctx) lowerARM64SVESplice(op Op, ins Instr) (ok bool, terminated bool, err error) {
	if op != "ZSPLICE" {
		return false, false, nil
	}
	if strings.ToUpper(string(ins.Op)) != string(op) || len(ins.Args) < 3 || len(ins.Args) > 4 {
		return true, false, fmt.Errorf("arm64 ZSPLICE expects one Go 1.27 vector form without a suffix: %q", ins.Raw)
	}

	var first, second, predicate, destination, elementBits int
	if len(ins.Args) == 4 {
		var firstOK, secondOK, predicateOK, destinationOK bool
		second, elementBits, secondOK = arm64ParseSVEZElementReg(ins.Args[0])
		first, _, firstOK = arm64ParseSVEZElementReg(ins.Args[1])
		predicate, predicateOK = arm64ParseSVEPredicateBare(ins.Args[2], 15)
		destination, _, destinationOK = arm64ParseSVEZElementReg(ins.Args[3])
		_, firstBits, _ := arm64ParseSVEZElementReg(ins.Args[1])
		_, destinationBits, _ := arm64ParseSVEZElementReg(ins.Args[3])
		if !firstOK || !secondOK || !predicateOK || !destinationOK ||
			firstBits != elementBits || destinationBits != elementBits || first != destination {
			return true, false, fmt.Errorf("arm64 ZSPLICE destructive form requires Zm.T, Zdn.T, P0..P15, Zdn.T with one B/H/S/D width: %q", ins.Raw)
		}
	} else {
		if ins.Args[0].Kind != OpRegList || len(ins.Args[0].RegList) != 2 {
			return true, false, fmt.Errorf("arm64 ZSPLICE list form requires exactly two consecutive vectors: %q", ins.Raw)
		}
		var firstOK, secondOK, predicateOK, destinationOK bool
		first, elementBits, firstOK = arm64ParseSVEZElementReg(Operand{Kind: OpReg, Reg: ins.Args[0].RegList[0]})
		second, _, secondOK = arm64ParseSVEZElementReg(Operand{Kind: OpReg, Reg: ins.Args[0].RegList[1]})
		predicate, predicateOK = arm64ParseSVEPredicateBare(ins.Args[1], 15)
		destination, _, destinationOK = arm64ParseSVEZElementReg(ins.Args[2])
		_, secondBits, _ := arm64ParseSVEZElementReg(Operand{Kind: OpReg, Reg: ins.Args[0].RegList[1]})
		_, destinationBits, _ := arm64ParseSVEZElementReg(ins.Args[2])
		if !firstOK || !secondOK || !predicateOK || !destinationOK || first == 31 || second != first+1 ||
			secondBits != elementBits || destinationBits != elementBits {
			return true, false, fmt.Errorf("arm64 ZSPLICE list form requires [Zn.T, Z(n+1).T], P0..P15, Zd.T with one B/H/S/D width: %q", ins.Raw)
		}
	}

	predicateValue, predicateType, err := c.loadPRegElements(predicate, elementBits)
	if err != nil {
		return true, false, err
	}
	firstValue, vectorType, err := c.loadZRegElements(first, elementBits)
	if err != nil {
		return true, false, err
	}
	secondValue, _, err := c.loadZRegElements(second, elementBits)
	if err != nil {
		return true, false, err
	}
	_, lanes, err := arm64SVEVectorType(elementBits)
	if err != nil {
		return true, false, err
	}
	result := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = call %s @llvm.aarch64.sve.splice.nxv%di%d(%s %s, %s %s, %s %s)\n",
		result, vectorType, lanes, elementBits, predicateType, predicateValue, vectorType, firstValue, vectorType, secondValue)
	return true, false, c.storeZRegElements(destination, elementBits, "%"+result)
}
