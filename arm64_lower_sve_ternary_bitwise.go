package plan9asm

import (
	"fmt"
	"strings"
)

var arm64SVETernaryBitwiseIntrinsics = map[Op]string{
	"ZBCAX":  "bcax",
	"ZBSL":   "bsl",
	"ZBSL1N": "bsl1n",
	"ZBSL2N": "bsl2n",
	"ZEOR3":  "eor3",
	"ZNBSL":  "nbsl",
}

func (c *arm64Ctx) lowerARM64SVETernaryBitwise(op Op, ins Instr) (ok bool, terminated bool, err error) {
	intrinsic, ok := arm64SVETernaryBitwiseIntrinsics[op]
	if !ok {
		return false, false, nil
	}
	if strings.ToUpper(string(ins.Op)) != string(op) || len(ins.Args) != 4 {
		return true, false, fmt.Errorf("arm64 %s expects Zk.D, Zm.D, Zdn.D, Zdn.D without a suffix: %q", op, ins.Raw)
	}
	third, thirdBits, thirdOK := arm64ParseSVEZElementReg(ins.Args[0])
	second, secondBits, secondOK := arm64ParseSVEZElementReg(ins.Args[1])
	first, firstBits, firstOK := arm64ParseSVEZElementReg(ins.Args[2])
	destination, destinationBits, destinationOK := arm64ParseSVEZElementReg(ins.Args[3])
	if !thirdOK || !secondOK || !firstOK || !destinationOK || thirdBits != 64 || secondBits != 64 || firstBits != 64 || destinationBits != 64 || first != destination {
		return true, false, fmt.Errorf("arm64 %s requires D-width operands and one destructive destination: %q", op, ins.Raw)
	}
	firstValue, vectorType, err := c.loadZRegElements(first, 64)
	if err != nil {
		return true, false, err
	}
	secondValue, _, err := c.loadZRegElements(second, 64)
	if err != nil {
		return true, false, err
	}
	thirdValue, _, err := c.loadZRegElements(third, 64)
	if err != nil {
		return true, false, err
	}
	result := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = call %s @llvm.aarch64.sve.%s.nxv2i64(%s %s, %s %s, %s %s)\n", result, vectorType, intrinsic, vectorType, firstValue, vectorType, secondValue, vectorType, thirdValue)
	return true, false, c.storeZRegElements(destination, 64, "%"+result)
}
