package plan9asm

import (
	"fmt"
	"strings"
)

var arm64SVEPredicateSelectOps = map[Op]struct{}{
	"PSEL": {},
}

func (c *arm64Ctx) lowerARM64SVEPredicateSelect(op Op, ins Instr) (ok bool, terminated bool, err error) {
	if _, ok := arm64SVEPredicateSelectOps[op]; !ok {
		return false, false, nil
	}
	if strings.ToUpper(string(ins.Op)) != string(op) || len(ins.Args) != 4 {
		return true, false, fmt.Errorf("arm64 PSEL expects Pm.B, Pn.B, Pg, Pd.B: %q", ins.Raw)
	}
	first, firstBits, firstOK := arm64ParseSVEPredicateElement(ins.Args[0])
	second, secondBits, secondOK := arm64ParseSVEPredicateElement(ins.Args[1])
	governing, governingOK := arm64ParseSVEPredicateBare(ins.Args[2], 15)
	destination, destinationBits, destinationOK := arm64ParseSVEPredicateElement(ins.Args[3])
	if !firstOK || !secondOK || !governingOK || !destinationOK || firstBits != 8 || secondBits != 8 || destinationBits != 8 {
		return true, false, fmt.Errorf("arm64 PSEL only accepts the Go 1.27 Pm.B, Pn.B, Pg, Pd.B form: %q", ins.Raw)
	}
	firstValue, err := c.loadPReg(first)
	if err != nil {
		return true, false, err
	}
	secondValue, err := c.loadPReg(second)
	if err != nil {
		return true, false, err
	}
	governingValue, err := c.loadPReg(governing)
	if err != nil {
		return true, false, err
	}
	result := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = select <vscale x 16 x i1> %s, <vscale x 16 x i1> %s, <vscale x 16 x i1> %s\n", result, governingValue, secondValue, firstValue)
	return true, false, c.storePReg(destination, "%"+result)
}
