package plan9asm

import (
	"fmt"
	"strings"
)

var arm64UnsignedWideningAddOps = map[Op]struct{}{
	"VUADDW":  {},
	"VUADDW2": {},
}

func (c *arm64Ctx) lowerARM64UnsignedWideningAdd(op Op, ins Instr) (ok bool, terminated bool, err error) {
	if _, ok := arm64UnsignedWideningAddOps[op]; !ok {
		return false, false, nil
	}
	if strings.ToUpper(string(ins.Op)) != string(op) || len(ins.Args) != 3 {
		return true, false, fmt.Errorf("arm64 %s expects narrow source, wide addend, and wide destination: %q", op, ins.Raw)
	}
	for _, arg := range ins.Args {
		if arg.Kind != OpReg {
			return true, false, fmt.Errorf("arm64 %s accepts only vector registers: %q", op, ins.Raw)
		}
	}
	narrow, narrowOK := parseARM64VectorArrangement(ins.Args[0].Reg)
	addendArrangement, addendOK := parseARM64VectorArrangement(ins.Args[1].Reg)
	destinationArrangement, destinationOK := parseARM64VectorArrangement(ins.Args[2].Reg)
	highHalf := op == "VUADDW2"
	wantNarrowLanes := addendArrangement.lanes
	if highHalf {
		wantNarrowLanes *= 2
	}
	if !narrowOK || !addendOK || !destinationOK || addendArrangement != destinationArrangement ||
		narrow.elementBits*2 != addendArrangement.elementBits || narrow.lanes != wantNarrowLanes ||
		addendArrangement.lanes*addendArrangement.elementBits != 128 {
		return true, false, fmt.Errorf("arm64 %s arrangements do not form a valid unsigned widening-add row: %q", op, ins.Raw)
	}
	narrowValue, err := c.loadARM64VectorInteger(ins.Args[0].Reg, narrow)
	if err != nil {
		return true, false, err
	}
	if highHalf {
		selected := c.newTmp()
		fmt.Fprintf(c.b, "  %%%s = shufflevector <%d x i%d> %s, <%d x i%d> poison, <%d x i32> <", selected, narrow.lanes, narrow.elementBits, narrowValue, narrow.lanes, narrow.elementBits, addendArrangement.lanes)
		for lane := 0; lane < addendArrangement.lanes; lane++ {
			if lane != 0 {
				c.b.WriteString(", ")
			}
			fmt.Fprintf(c.b, "i32 %d", addendArrangement.lanes+lane)
		}
		c.b.WriteString(">\n")
		narrowValue = "%" + selected
	}
	extended := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = zext <%d x i%d> %s to <%d x i%d>\n", extended, addendArrangement.lanes, narrow.elementBits, narrowValue, addendArrangement.lanes, addendArrangement.elementBits)
	addend, err := c.loadARM64VectorInteger(ins.Args[1].Reg, addendArrangement)
	if err != nil {
		return true, false, err
	}
	result := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = add <%d x i%d> %s, %%%s\n", result, addendArrangement.lanes, addendArrangement.elementBits, addend, extended)
	return true, false, c.storeARM64VectorInteger(ins.Args[2].Reg, destinationArrangement, "%"+result)
}
