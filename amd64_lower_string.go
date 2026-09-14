package plan9asm

import (
	"fmt"
)

type x86StringSpec struct {
	kind  string
	typ   LLVMType
	width int
}

var x86StringSpecs = map[Op]x86StringSpec{
	"MOVSB": {kind: "movs", typ: I8, width: 1},
	"MOVSW": {kind: "movs", typ: I16, width: 2},
	"MOVSL": {kind: "movs", typ: I32, width: 4},
	"MOVSQ": {kind: "movs", typ: I64, width: 8},
	"STOSB": {kind: "stos", typ: I8, width: 1},
	"STOSW": {kind: "stos", typ: I16, width: 2},
	"STOSL": {kind: "stos", typ: I32, width: 4},
	"STOSQ": {kind: "stos", typ: I64, width: 8},
	"SCASB": {kind: "scas", typ: I8, width: 1},
	"SCASW": {kind: "scas", typ: I16, width: 2},
	"SCASL": {kind: "scas", typ: I32, width: 4},
	"SCASQ": {kind: "scas", typ: I64, width: 8},
}

func x86StringProperties(op Op) (kind string, typ LLVMType, width int, ok bool) {
	spec, ok := x86StringSpecs[Op(normalizeInstructionOpcode(op))]
	if !ok {
		return "", "", 0, false
	}
	return spec.kind, spec.typ, spec.width, true
}

func (c *amd64Ctx) lowerString(op Op, ins Instr) (ok bool, terminated bool, err error) {
	kind, typ, width, ok := x86StringProperties(op)
	if !ok {
		return false, false, nil
	}
	prefix := c.repeatPrefix
	c.repeatPrefix = ""
	if c.goarch == "386" && typ == I64 {
		return true, false, fmt.Errorf("386 %s is not in the Go assembler's 32-bit instruction table", op)
	}
	if len(ins.Args) != 0 {
		return true, false, fmt.Errorf("%s %s takes no operands: %q", c.goarch, op, ins.Raw)
	}

	count := "1"
	if prefix != "" {
		cx, err := c.loadReg(CX)
		if err != nil {
			return true, false, err
		}
		if c.goarch == "386" {
			cx32 := c.truncI64(cx, I32)
			wide := c.newTmp()
			fmt.Fprintf(c.b, "  %%%s = zext i32 %s to i64\n", wide, cx32)
			count = "%" + wide
		} else {
			count = cx
		}
	}
	direction := c.loadDirectionFlag()

	switch kind {
	case "movs":
		si, err := c.loadReg(SI)
		if err != nil {
			return true, false, err
		}
		di, err := c.loadReg(DI)
		if err != nil {
			return true, false, err
		}
		fmt.Fprintf(c.b, "  call void @%s(i64 %s, i64 %s, i64 %s, i1 %s)\n", x86StringHelperName(kind, width), di, si, count, direction)
		bytes := c.newTmp()
		fmt.Fprintf(c.b, "  %%%s = mul i64 %s, %d\n", bytes, count, width)
		if err := c.storeStringIndex(SI, si, "%"+bytes, direction); err != nil {
			return true, false, err
		}
		if err := c.storeStringIndex(DI, di, "%"+bytes, direction); err != nil {
			return true, false, err
		}

	case "stos":
		di, err := c.loadReg(DI)
		if err != nil {
			return true, false, err
		}
		value, err := c.evalIntSized(Operand{Kind: OpReg, Reg: AX}, typ)
		if err != nil {
			return true, false, err
		}
		fmt.Fprintf(c.b, "  call void @%s(i64 %s, %s %s, i64 %s, i1 %s)\n", x86StringHelperName(kind, width), di, typ, value, count, direction)
		bytes := c.newTmp()
		fmt.Fprintf(c.b, "  %%%s = mul i64 %s, %d\n", bytes, count, width)
		if err := c.storeStringIndex(DI, di, "%"+bytes, direction); err != nil {
			return true, false, err
		}

	case "scas":
		di, err := c.loadReg(DI)
		if err != nil {
			return true, false, err
		}
		needle, err := c.evalIntSized(Operand{Kind: OpReg, Reg: AX}, typ)
		if err != nil {
			return true, false, err
		}
		whileEqual := prefix == "REP"
		call := c.newTmp()
		fmt.Fprintf(c.b, "  %%%s = call { i64, i64, i1 } @%s(i64 %s, %s %s, i64 %s, i1 %s, i1 %t)\n", call, x86StringHelperName(kind, width), di, typ, needle, count, direction, whileEqual)
		next := c.newTmp()
		remaining := c.newTmp()
		equal := c.newTmp()
		fmt.Fprintf(c.b, "  %%%s = extractvalue { i64, i64, i1 } %%%s, 0\n", next, call)
		fmt.Fprintf(c.b, "  %%%s = extractvalue { i64, i64, i1 } %%%s, 1\n", remaining, call)
		fmt.Fprintf(c.b, "  %%%s = extractvalue { i64, i64, i1 } %%%s, 2\n", equal, call)
		if err := c.storeReg(DI, "%"+next); err != nil {
			return true, false, err
		}
		zf := "%" + equal
		if prefix != "" {
			oldZF := c.loadFlag(c.flagsZSlot)
			executed := c.newTmp()
			fmt.Fprintf(c.b, "  %%%s = icmp ne i64 %s, 0\n", executed, count)
			preserved := c.newTmp()
			fmt.Fprintf(c.b, "  %%%s = select i1 %%%s, i1 %%%s, i1 %s\n", preserved, executed, equal, oldZF)
			zf = "%" + preserved
		}
		fmt.Fprintf(c.b, "  store i1 %s, ptr %s\n", zf, c.flagsZSlot)
		if prefix != "" {
			if err := c.storeStringCount("%" + remaining); err != nil {
				return true, false, err
			}
		}
	}

	if prefix != "" && kind != "scas" {
		if err := c.storeStringCount("0"); err != nil {
			return true, false, err
		}
	}
	return true, false, nil
}

func x86StringHelperName(kind string, width int) string {
	suffix := map[int]string{1: "b", 2: "w", 4: "l", 8: "q"}[width]
	if kind == "movs" {
		return "__plan9asm_movs" + suffix
	}
	return "__plan9asm_rep_" + kind + suffix
}

func (c *amd64Ctx) storeStringCount(value string) error {
	if c.goarch == "386" {
		return c.storeRegSized(CX, I32, c.truncI64(value, I32))
	}
	return c.storeRegSized(CX, I64, value)
}

func (c *amd64Ctx) loadDirectionFlag() string {
	t := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = load i1, ptr %s\n", t, c.directionSlot)
	return "%" + t
}

func (c *amd64Ctx) storeStringIndex(reg Reg, old, delta, backward string) error {
	forward := c.newTmp()
	reverse := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = add i64 %s, %s\n", forward, old, delta)
	fmt.Fprintf(c.b, "  %%%s = sub i64 %s, %s\n", reverse, old, delta)
	next := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = select i1 %s, i64 %%%s, i64 %%%s\n", next, backward, reverse, forward)
	return c.storeReg(reg, "%"+next)
}
