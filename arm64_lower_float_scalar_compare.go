package plan9asm

import (
	"fmt"
	"strings"
)

func (c *arm64Ctx) lowerARM64ScalarFloatCompare(op Op, ins Instr) (ok bool, terminated bool, err error) {
	switch op {
	case "FCMPS", "FCMPD", "FCMPES", "FCMPED":
	default:
		return false, false, nil
	}
	if strings.ToUpper(string(ins.Op)) != string(op) || len(ins.Args) != 2 ||
		ins.Args[1].Kind != OpReg || !isARM64FReg(ins.Args[1].Reg) ||
		!arm64ScalarFloatCompareSource(ins.Args[0]) {
		return true, false, fmt.Errorf("arm64 %s expects Fsrc or $(0.0), Fdst and no suffix: %q", op, ins.Raw)
	}

	bits := 64
	if op == "FCMPS" || op == "FCMPES" {
		bits = 32
	}
	right := ""
	if ins.Args[0].Kind == OpReg {
		right, err = c.loadARM64ScalarFloatReg(ins.Args[0].Reg, bits)
		if err != nil {
			return true, false, err
		}
	}
	left, err := c.loadARM64ScalarFloatReg(ins.Args[1].Reg, bits)
	if err != nil {
		return true, false, err
	}
	c.emitARM64ScalarFloatCompare(bits, op == "FCMPES" || op == "FCMPED", left, right)
	return true, false, nil
}

func arm64ScalarFloatCompareSource(operand Operand) bool {
	if operand.Kind == OpReg {
		return isARM64FReg(operand.Reg)
	}
	// Go's C_FCON optab class accepts any floating constant here even though
	// AArch64's FCMP-immediate encoding always compares against zero. Preserve
	// that assembler behavior instead of treating the source spelling as the
	// value encoded by the instruction.
	return operand.Kind == OpImm && operand.ImmRaw == "" && operand.ImmIsFloat
}

// DDI0602 FCMP/FCMPE differ for quiet NaNs and both consume FPCR and update
// FPSR even when NZCV is dead. Execute the exact operation, capturing NZCV in
// the same asm block. The memory barrier orders existing native MRS/MSR;
// it grants no memory, source-layout or machine-entry permission. All virtual
// source reads were recorded by loadARM64ScalarFloatReg, and no SIMD bits change.
// An empty right operand represents the architecture's immediate-zero form.
func (c *arm64Ctx) emitARM64ScalarFloatCompare(bits int, signaling bool, left, right string) {
	opcode := "fcmp"
	if signaling {
		opcode = "fcmpe"
	}
	typeName, modifier := "double", "d"
	if bits == 32 {
		typeName, modifier = "float", "s"
	} else if bits == 16 {
		typeName, modifier = "half", "h"
	}
	assembly := fmt.Sprintf("%s ${1:%s}, #0.0\\0Amrs $0, nzcv", opcode, modifier)
	constraints := "=r,w,~{cc},~{memory}"
	arguments := typeName + " " + left
	if right != "" {
		assembly = fmt.Sprintf("%s ${1:%s}, ${2:%s}\\0Amrs $0, nzcv", opcode, modifier, modifier)
		constraints = "=r,w,w,~{cc},~{memory}"
		arguments += ", " + typeName + " " + right
	}
	nzcv := c.newTmp()
	// Assembly already contains an LLVM newline escape rather than a literal
	// newline: do not quote it through %q a second time.
	c.emitMachineNeutralNativeIR("  %%%s = call i64 asm sideeffect \"%s\", %q(%s)\n", nzcv, assembly, constraints, arguments)
	c.flagsWritten = true
	for index, slot := range []string{c.flagsVSlot, c.flagsCSlot, c.flagsZSlot, c.flagsNSlot} {
		shifted, flag := c.newTmp(), c.newTmp()
		fmt.Fprintf(c.b, "  %%%s = lshr i64 %%%s, %d\n", shifted, nzcv, 28+index)
		fmt.Fprintf(c.b, "  %%%s = trunc i64 %%%s to i1\n", flag, shifted)
		c.storeFlag(slot, "%"+flag)
	}
}
