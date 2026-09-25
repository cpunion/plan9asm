package plan9asm

import (
	"fmt"
	"strings"
)

func (c *arm64Ctx) lowerARM64SVEFloatDivideScale(op Op, ins Instr) (ok bool, terminated bool, err error) {
	if op != "ZFDIV" && op != "ZFDIVR" && op != "ZFSCALE" {
		return false, false, nil
	}
	if strings.ToUpper(string(ins.Op)) != string(op) || len(ins.Args) != 4 {
		return true, false, fmt.Errorf("arm64 %s expects Zm.H|S|D, Zdn.H|S|D, P0..P7/M, Zdn.H|S|D without a suffix: %q", op, ins.Raw)
	}
	second, secondBits, secondOK := arm64SVEFloatElementReg(ins.Args[0])
	first, firstBits, firstOK := arm64SVEFloatElementReg(ins.Args[1])
	predicate, predicateOK := arm64ParseSVEPredicateMode(ins.Args[2], "M", 7)
	destination, destinationBits, destinationOK := arm64SVEFloatElementReg(ins.Args[3])
	if !secondOK || !firstOK || !predicateOK || !destinationOK || secondBits != firstBits || destinationBits != firstBits || destination != first {
		return true, false, fmt.Errorf("arm64 %s requires matching-width destructive H/S/D operands and P0..P7/M: %q", op, ins.Raw)
	}
	firstValue, vectorType, err := c.loadRawSVEFloatVector(first, firstBits)
	if err != nil {
		return true, false, err
	}
	predicateValue, predicateType, err := c.loadPRegElements(predicate, firstBits)
	if err != nil {
		return true, false, err
	}
	_, _, lanes, err := arm64SVEFloatType(firstBits)
	if err != nil {
		return true, false, err
	}
	mangle := map[int]string{16: "f16", 32: "f32", 64: "f64"}[firstBits]
	result := c.newTmp()
	if op == "ZFSCALE" {
		secondValue, integerType, err := c.loadZRegElements(second, secondBits)
		if err != nil {
			return true, false, err
		}
		fmt.Fprintf(c.b, "  %%%s = call %s @llvm.aarch64.sve.fscale.nxv%d%s(%s %s, %s %s, %s %s)\n", result, vectorType, lanes, mangle, predicateType, predicateValue, vectorType, firstValue, integerType, secondValue)
	} else {
		secondValue, _, err := c.loadRawSVEFloatVector(second, secondBits)
		if err != nil {
			return true, false, err
		}
		intrinsic := "fdiv"
		if op == "ZFDIVR" {
			intrinsic = "fdivr"
		}
		fmt.Fprintf(c.b, "  %%%s = call %s @llvm.aarch64.sve.%s.nxv%d%s(%s %s, %s %s, %s %s)\n", result, vectorType, intrinsic, lanes, mangle, predicateType, predicateValue, vectorType, firstValue, vectorType, secondValue)
	}
	selected := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = select %s %s, %s %%%s, %s %s\n", selected, predicateType, predicateValue, vectorType, result, vectorType, firstValue)
	return true, false, c.storeRawSVEFloatVector(destination, firstBits, "%"+selected, vectorType)
}
