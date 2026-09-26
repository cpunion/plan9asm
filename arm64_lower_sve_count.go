package plan9asm

import (
	"fmt"
	"strings"
)

type arm64RawSVECnt struct {
	op          Op
	elementBits int
	destination int
	pattern     int
	multiplier  int
	operation   string
}

func decodeARM64RawSVECnt(word uint32) (arm64RawSVECnt, bool) {
	const variable = uint32(3<<22 | 15<<16 | 31<<5 | 31)
	operation := ""
	prefix := "CNT"
	switch word &^ variable {
	case 0x0420e000:
	case 0x0430e000:
		prefix, operation = "INC", "add"
	case 0x0430e400:
		prefix, operation = "DEC", "sub"
	default:
		return arm64RawSVECnt{}, false
	}
	size := word >> 22 & 3
	return arm64RawSVECnt{
		op: Op(prefix + string("BHWD"[size])), elementBits: 8 << size,
		destination: int(word & 31), pattern: int(word >> 5 & 31),
		multiplier: int(word>>16&15) + 1, operation: operation,
	}, true
}

func (c *arm64Ctx) lowerRawSVECnt(form arm64RawSVECnt) error {
	reg := Reg(fmt.Sprintf("R%d", form.destination))
	if form.destination == 31 {
		reg = ZR
	}
	return c.lowerARM64SVEElementCount(reg, form.elementBits, form.pattern, form.multiplier, form.operation)
}

// CNT/INC/DEC share predicate-pattern counts and the encoded 1..16 multiplier.
// LLVM's count intrinsics implement POW2, fixed VL, MUL3/MUL4, ALL and unknown
// pattern values (which architecturally select zero elements).
func (c *arm64Ctx) lowerARM64SVEElementCount(reg Reg, bits, pattern, multiplier int, operation string) error {
	count := c.newTmp()
	suffix := map[int]string{8: "b", 16: "h", 32: "w", 64: "d"}[bits]
	fmt.Fprintf(c.b, "  %%%s = call i64 @llvm.aarch64.sve.cnt%s(i32 %d)\n", count, suffix, pattern)
	value := "%" + count
	if multiplier != 1 {
		scaled := c.newTmp()
		fmt.Fprintf(c.b, "  %%%s = mul i64 %s, %d\n", scaled, value, multiplier)
		value = "%" + scaled
	}
	if operation != "" {
		previous, err := c.loadReg(reg)
		if err != nil {
			return err
		}
		updated := c.newTmp()
		fmt.Fprintf(c.b, "  %%%s = %s i64 %s, %s\n", updated, operation, previous, value)
		value = "%" + updated
	}
	return c.storeReg(reg, value)
}

func (c *arm64Ctx) lowerARM64SVECnt(op Op, ins Instr) (ok bool, terminated bool, err error) {
	multiplier, handled := map[Op]int{"CNTB": 16, "CNTH": 8, "CNTW": 4, "CNTD": 2}[op]
	if !handled {
		return false, false, nil
	}
	if strings.ToUpper(string(ins.Op)) != string(op) || len(ins.Args) != 1 || ins.Args[0].Kind != OpReg || !isARM64GeneralOrZeroReg(ins.Args[0].Reg) || ins.Args[0].Reg == ZR {
		return true, false, fmt.Errorf("arm64 %s expects one general destination register: %q", op, ins.Raw)
	}
	return true, false, c.lowerARM64SVEElementCount(ins.Args[0].Reg, 128/multiplier, 31, 1, "")
}
