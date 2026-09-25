package plan9asm

import (
	"fmt"
	"strings"
)

func arm64SVECompactNeedsSVE2P2(ins Instr) bool {
	if len(ins.Args) != 3 {
		return false
	}
	_, bits, ok := arm64ParseSVEZElementReg(ins.Args[0])
	return ok && bits <= 16
}

func (c *arm64Ctx) lowerARM64SVECompact(op Op, ins Instr) (ok bool, terminated bool, err error) {
	if op != "ZCOMPACT" {
		return false, false, nil
	}
	if strings.ToUpper(string(ins.Op)) != string(op) || len(ins.Args) != 3 {
		return true, false, fmt.Errorf("arm64 ZCOMPACT expects Zn.T, Pg, Zd.T without a suffix: %q", ins.Raw)
	}
	source, sourceBits, sourceOK := arm64ParseSVEZElementReg(ins.Args[0])
	predicate, predicateOK := arm64ParseSVEPredicateBare(ins.Args[1], 7)
	destination, destinationBits, destinationOK := arm64ParseSVEZElementReg(ins.Args[2])
	if !sourceOK || !predicateOK || !destinationOK || sourceBits != destinationBits {
		return true, false, fmt.Errorf("arm64 ZCOMPACT requires matching B/H/S/D vectors and P0..P7: %q", ins.Raw)
	}
	predicateValue, predicateType, err := c.loadPRegElements(predicate, sourceBits)
	if err != nil {
		return true, false, err
	}
	sourceValue, vectorType, err := c.loadZRegElements(source, sourceBits)
	if err != nil {
		return true, false, err
	}
	_, lanes, err := arm64SVEVectorType(sourceBits)
	if err != nil {
		return true, false, err
	}
	result := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = call %s @llvm.aarch64.sve.compact.nxv%di%d(%s %s, %s %s)\n",
		result, vectorType, lanes, sourceBits, predicateType, predicateValue, vectorType, sourceValue)
	return true, false, c.storeZRegElements(destination, sourceBits, "%"+result)
}
