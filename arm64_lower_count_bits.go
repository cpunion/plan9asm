package plan9asm

import (
	"fmt"
	"strings"
)

var arm64VectorCountBitsOps = map[Op]struct{}{
	"VCNT": {},
}

func (c *arm64Ctx) lowerARM64VectorCountBits(op Op, ins Instr) (ok bool, terminated bool, err error) {
	if _, ok := arm64VectorCountBitsOps[op]; !ok {
		return false, false, nil
	}
	if strings.ToUpper(string(ins.Op)) != "VCNT" || len(ins.Args) != 2 || ins.Args[0].Kind != OpReg || ins.Args[1].Kind != OpReg {
		return true, false, fmt.Errorf("arm64 VCNT expects source and destination B8/B16 vector registers: %q", ins.Raw)
	}
	sourceArrangement, sourceOK := parseARM64VectorArrangement(ins.Args[0].Reg)
	destinationArrangement, destinationOK := parseARM64VectorArrangement(ins.Args[1].Reg)
	if !sourceOK || !destinationOK || sourceArrangement != destinationArrangement ||
		sourceArrangement.elementBits != 8 || sourceArrangement.lanes != 8 && sourceArrangement.lanes != 16 {
		return true, false, fmt.Errorf("arm64 VCNT requires matching B8 or B16 arrangements: %q", ins.Raw)
	}
	source, err := c.loadARM64VectorInteger(ins.Args[0].Reg, sourceArrangement)
	if err != nil {
		return true, false, err
	}
	result := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = call <%d x i8> @llvm.ctpop.v%di8(<%d x i8> %s)\n", result, sourceArrangement.lanes, sourceArrangement.lanes, sourceArrangement.lanes, source)
	return true, false, c.storeARM64VectorInteger(ins.Args[1].Reg, destinationArrangement, "%"+result)
}
