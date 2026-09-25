package plan9asm

import (
	"fmt"
	"strings"
)

func (c *arm64Ctx) lowerARM64SVEIndex(op Op, ins Instr) (ok bool, terminated bool, err error) {
	if op != "ZINDEX" && op != "ZINDEXW" {
		return false, false, nil
	}
	if strings.ToUpper(string(ins.Op)) != string(op) || len(ins.Args) != 3 {
		return true, false, fmt.Errorf("arm64 %s expects step, start, Zd.B/H/S/D without a suffix: %q", op, ins.Raw)
	}
	destination, writtenBits, destinationOK := arm64ParseSVEZElementReg(ins.Args[2])
	if !destinationOK {
		return true, false, fmt.Errorf("arm64 %s destination must be Z0..Z31.B/H/S/D: %q", op, ins.Raw)
	}
	stepIsReg := arm64SVEIndexScalarRegister(ins.Args[0])
	startIsReg := arm64SVEIndexScalarRegister(ins.Args[1])
	stepIsImm := arm64SVEIndexImmediate(ins.Args[0])
	startIsImm := arm64SVEIndexImmediate(ins.Args[1])
	if !stepIsReg && !stepIsImm || !startIsReg && !startIsImm || op == "ZINDEXW" && stepIsImm && startIsImm {
		return true, false, fmt.Errorf("arm64 %s operands do not match its Go 1.27 scalar/immediate forms: %q", op, ins.Raw)
	}

	// Go's ZINDEX register encodings carry fixed D size bits. The generated
	// assembler accepts B/H/S spellings too, but ORs their size field with D;
	// preserve the resulting machine semantics rather than the written suffix.
	elementBits := writtenBits
	if op == "ZINDEX" && (stepIsReg || startIsReg) {
		elementBits = 64
	}
	step, err := c.arm64SVEIndexScalar(ins.Args[0], elementBits)
	if err != nil {
		return true, false, err
	}
	start, err := c.arm64SVEIndexScalar(ins.Args[1], elementBits)
	if err != nil {
		return true, false, err
	}
	vectorType, lanes, err := arm64SVEVectorType(elementBits)
	if err != nil {
		return true, false, err
	}
	result := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = call %s @llvm.aarch64.sve.index.nxv%di%d(i%d %s, i%d %s)\n",
		result, vectorType, lanes, elementBits, elementBits, start, elementBits, step)
	return true, false, c.storeZRegElements(destination, elementBits, "%"+result)
}

func arm64SVEIndexScalarRegister(operand Operand) bool {
	return operand.Kind == OpReg && isARM64GeneralOrZeroReg(operand.Reg)
}

func arm64SVEIndexImmediate(operand Operand) bool {
	return operand.Kind == OpImm && operand.ImmRaw == "" && operand.Imm >= -16 && operand.Imm <= 15
}

func (c *arm64Ctx) arm64SVEIndexScalar(operand Operand, elementBits int) (string, error) {
	if arm64SVEIndexImmediate(operand) {
		return fmt.Sprintf("%d", operand.Imm), nil
	}
	if !arm64SVEIndexScalarRegister(operand) {
		return "", fmt.Errorf("arm64 SVE INDEX scalar must be R0..R30, ZR, or a signed 5-bit immediate")
	}
	value, err := c.loadReg(operand.Reg)
	if err != nil {
		return "", err
	}
	if elementBits == 64 || value == "0" {
		return value, nil
	}
	truncated := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = trunc i64 %s to i%d\n", truncated, value, elementBits)
	return "%" + truncated, nil
}
