package plan9asm

import (
	"fmt"
	"strings"
)

type arm64RawSVECnt struct {
	op          Op
	elementBits int
	destination int
}

func decodeARM64RawSVECnt(word uint32) (arm64RawSVECnt, bool) {
	base := word & 0xffe0ffe0
	forms := map[uint32]struct {
		op   Op
		bits int
	}{
		0x0420e3e0: {"CNTB", 8},
		0x0460e3e0: {"CNTH", 16},
		0x04a0e3e0: {"CNTW", 32},
		0x04e0e3e0: {"CNTD", 64},
	}
	form, ok := forms[base]
	if !ok {
		return arm64RawSVECnt{}, false
	}
	return arm64RawSVECnt{op: form.op, elementBits: form.bits, destination: int(word & 31)}, true
}

func (c *arm64Ctx) lowerRawSVECnt(form arm64RawSVECnt) error {
	ins := Instr{Op: form.op, Raw: fmt.Sprintf("decoded ARM64 WORD as %s", form.op), Args: []Operand{{Kind: OpReg, Reg: Reg(fmt.Sprintf("R%d", form.destination))}}}
	ok, _, err := c.lowerARM64SVECnt(form.op, ins)
	if !ok && err == nil {
		return fmt.Errorf("arm64 raw %s decoder reached no semantic lowerer", form.op)
	}
	return err
}

func (c *arm64Ctx) lowerARM64SVECnt(op Op, ins Instr) (ok bool, terminated bool, err error) {
	multiplier, handled := map[Op]int{"CNTB": 16, "CNTH": 8, "CNTW": 4, "CNTD": 2}[op]
	if !handled {
		return false, false, nil
	}
	if strings.ToUpper(string(ins.Op)) != string(op) || len(ins.Args) != 1 || ins.Args[0].Kind != OpReg || !isARM64GeneralOrZeroReg(ins.Args[0].Reg) || ins.Args[0].Reg == ZR {
		return true, false, fmt.Errorf("arm64 %s expects one general destination register: %q", op, ins.Raw)
	}
	vscale := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = call i64 @llvm.vscale.i64()\n", vscale)
	count := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = mul i64 %%%s, %d\n", count, vscale, multiplier)
	return true, false, c.storeReg(ins.Args[0].Reg, "%"+count)
}
