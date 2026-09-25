package plan9asm

import (
	"fmt"
	"strings"
)

var arm64SVECopyVectorBaseSize = map[Op]int{
	"ZCPYB": 0,
	"ZCPYH": 1,
	"ZCPYS": 2,
	"ZCPYD": 3,
}

func (c *arm64Ctx) lowerARM64SVECopy(op Op, ins Instr) (ok bool, terminated bool, err error) {
	vectorBaseSize, vectorForm := arm64SVECopyVectorBaseSize[op]
	if op != "ZCPY" && op != "ZCPYW" && !vectorForm {
		return false, false, nil
	}
	if strings.ToUpper(string(ins.Op)) != string(op) || len(ins.Args) != 3 {
		return true, false, fmt.Errorf("arm64 %s expects scalar, predicate mode, Z.B/H/S/D without a suffix: %q", op, ins.Raw)
	}

	destination, writtenBits, destinationOK := arm64ParseSVEZElementReg(ins.Args[2])
	if !destinationOK {
		return true, false, fmt.Errorf("arm64 %s destination must be Z0..Z31.B/H/S/D: %q", op, ins.Raw)
	}
	writtenSize := map[int]int{8: 0, 16: 1, 32: 2, 64: 3}[writtenBits]
	elementBits := writtenBits
	predicateLimit := 7
	mode := "M"
	var scalar string

	switch {
	case op == "ZCPY" && ins.Args[0].Kind == OpImm:
		if !arm64SVECopyImmediate(ins.Args[0]) {
			return true, false, fmt.Errorf("arm64 ZCPY immediate must be signed 8-bit, optionally shifted left 8: %q", ins.Raw)
		}
		predicateLimit = 15
		if _, ok := arm64ParseSVEPredicateMode(ins.Args[1], "M", predicateLimit); !ok {
			mode = "Z"
		}
		scalar = arm64SVECopyImmediateValue(ins.Args[0].Imm, elementBits)
	case op == "ZCPY" || op == "ZCPYW":
		if !arm64SVECopyGeneralRegister(ins.Args[0]) {
			return true, false, fmt.Errorf("arm64 %s source must be R0..R30 or RSP: %q", op, ins.Raw)
		}
		baseSize := 3
		if op == "ZCPYW" {
			baseSize = 0
		}
		elementBits = 8 << (baseSize | writtenSize)
		scalar, err = c.loadReg(ins.Args[0].Reg)
		if err != nil {
			return true, false, err
		}
		scalar = c.arm64SVECopyTruncate(scalar, elementBits)
	case vectorForm:
		register, registerOK := arm64SVEInsertVRegister(ins.Args[0])
		if !registerOK {
			return true, false, fmt.Errorf("arm64 %s source must be a bare V0..V31 register: %q", op, ins.Raw)
		}
		elementBits = 8 << (vectorBaseSize | writtenSize)
		scalar, err = c.arm64SVEInsertVectorScalar(register, elementBits)
		if err != nil {
			return true, false, err
		}
	}

	predicate, predicateOK := arm64ParseSVEPredicateMode(ins.Args[1], mode, predicateLimit)
	if !predicateOK {
		return true, false, fmt.Errorf("arm64 %s predicate must be P0..P%d.%s: %q", op, predicateLimit, mode, ins.Raw)
	}
	predicateValue, predicateType, err := c.loadPRegElements(predicate, elementBits)
	if err != nil {
		return true, false, err
	}
	vectorType, lanes, err := arm64SVEVectorType(elementBits)
	if err != nil {
		return true, false, err
	}
	inserted := c.newTmp()
	splat := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = insertelement %s poison, i%d %s, i64 0\n", inserted, vectorType, elementBits, scalar)
	fmt.Fprintf(c.b, "  %%%s = shufflevector %s %%%s, %s poison, <vscale x %d x i32> zeroinitializer\n", splat, vectorType, inserted, vectorType, lanes)
	merge := "zeroinitializer"
	if mode == "M" {
		merge, _, err = c.loadZRegElements(destination, elementBits)
		if err != nil {
			return true, false, err
		}
	}
	result := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = select %s %s, %s %%%s, %s %s\n",
		result, predicateType, predicateValue, vectorType, splat, vectorType, merge)
	return true, false, c.storeZRegElements(destination, elementBits, "%"+result)
}

func arm64SVECopyImmediate(operand Operand) bool {
	if operand.Kind != OpImm || operand.ImmRaw != "" {
		return false
	}
	value := operand.Imm
	if value >= -128 && value <= 127 {
		return true
	}
	return value%256 == 0 && value/256 >= -128 && value/256 <= 127
}

func arm64SVECopyImmediateValue(value int64, elementBits int) string {
	mask := uint64(1)<<elementBits - 1
	return fmt.Sprintf("%d", uint64(value)&mask)
}

func arm64SVECopyGeneralRegister(operand Operand) bool {
	if operand.Kind != OpReg {
		return false
	}
	return operand.Reg == SP || operand.Reg == Reg("RSP") ||
		(isARM64GeneralOrZeroReg(operand.Reg) && operand.Reg != ZR)
}

func (c *arm64Ctx) arm64SVECopyTruncate(value string, elementBits int) string {
	if elementBits == 64 {
		return value
	}
	truncated := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = trunc i64 %s to i%d\n", truncated, value, elementBits)
	return "%" + truncated
}
