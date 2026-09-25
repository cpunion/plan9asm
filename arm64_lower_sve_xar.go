package plan9asm

import (
	"fmt"
	"strings"
)

func (c *arm64Ctx) lowerARM64SVEXAR(op Op, ins Instr) (ok bool, terminated bool, err error) {
	if op != "ZXAR" {
		return false, false, nil
	}
	if strings.ToUpper(string(ins.Op)) != string(op) || len(ins.Args) != 4 || ins.Args[0].Kind != OpImm || ins.Args[0].ImmRaw != "" {
		return true, false, fmt.Errorf("arm64 ZXAR expects $shift, Zm.T, Zdn.T, Zdn.T without a suffix: %q", ins.Raw)
	}
	shift := ins.Args[0].Imm
	second, secondBits, secondOK := arm64ParseSVEZElementReg(ins.Args[1])
	destination, destinationBits, destinationOK := arm64ParseSVEZElementReg(ins.Args[2])
	repeatedDestination, repeatedBits, repeatedOK := arm64ParseSVEZElementReg(ins.Args[3])
	if !secondOK || !destinationOK || !repeatedOK || secondBits != destinationBits || repeatedBits != destinationBits || repeatedDestination != destination || shift < 1 || shift >= int64(destinationBits) {
		return true, false, fmt.Errorf("arm64 ZXAR requires matching destructive B/H/S/D operands and shift $1..$elementBits-1: %q", ins.Raw)
	}
	destinationValue, vectorType, err := c.loadZRegElements(destination, destinationBits)
	if err != nil {
		return true, false, err
	}
	secondValue, _, err := c.loadZRegElements(second, secondBits)
	if err != nil {
		return true, false, err
	}
	lanes := 128 / destinationBits
	result := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = call %s @llvm.aarch64.sve.xar.nxv%di%d(%s %s, %s %s, i32 %d)\n", result, vectorType, lanes, destinationBits, vectorType, destinationValue, vectorType, secondValue, shift)
	return true, false, c.storeZRegElements(destination, destinationBits, "%"+result)
}
