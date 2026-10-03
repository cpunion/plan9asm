package plan9asm

import (
	"fmt"
	"strings"
)

type amd64HorizontalFloatingSpec struct {
	laneBits int
	subtract bool
	vector   bool
}

// amd64HorizontalFloatingSpecs is the complete Go 1.27 horizontal
// floating-point add/subtract family. Legacy forms use yxm; V-prefixed forms
// use _yvaddsubpd and are deliberately limited to VEX X/Y registers.
var amd64HorizontalFloatingSpecs = map[Op]amd64HorizontalFloatingSpec{
	"HADDPS":  {laneBits: 32},
	"HADDPD":  {laneBits: 64},
	"HSUBPS":  {laneBits: 32, subtract: true},
	"HSUBPD":  {laneBits: 64, subtract: true},
	"VHADDPS": {laneBits: 32, vector: true},
	"VHADDPD": {laneBits: 64, vector: true},
	"VHSUBPS": {laneBits: 32, subtract: true, vector: true},
	"VHSUBPD": {laneBits: 64, subtract: true, vector: true},
}

func (c *amd64Ctx) lowerHorizontalFloating(op Op, ins Instr) (ok bool, terminated bool, err error) {
	baseOp, _, _ := strings.Cut(strings.ToUpper(string(op)), ".")
	_, recognized := amd64HorizontalFloatingSpecs[Op(baseOp)]
	if !recognized {
		return false, false, nil
	}
	if ins.Op != "" && !strings.EqualFold(string(ins.Op), string(op)) {
		return true, false, fmt.Errorf("horizontal floating dispatch/source opcode mismatch: %s/%s", op, ins.Op)
	}
	ins.Op = op
	form, recognized, err := x86HorizontalFloatProperties(ins, c.goarch)
	if err != nil || !recognized {
		return recognized, false, err
	}
	byteWidth, spec := form.byteWidth, form.spec
	if ins.Args[form.first].Kind == OpFP {
		if err := c.validateFPVectorSpan(ins.Args[form.first].FPOffset, byteWidth); err != nil {
			return true, false, err
		}
	}
	first, err := c.loadHorizontalFloatingOperand(ins.Args[form.first], byteWidth, spec.laneBits)
	if err != nil {
		return true, false, err
	}
	second, err := c.loadHorizontalFloatingOperand(ins.Args[form.second], byteWidth, spec.laneBits)
	if err != nil {
		return true, false, err
	}
	result := c.emitHorizontalFloating(spec, byteWidth, first, second)
	out := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = bitcast %s %s to <%d x i8>\n", out, amd64FMA3LLVMType(byteWidth*8/spec.laneBits, spec.laneBits), result, byteWidth)
	if !spec.vector {
		return true, false, c.storeLegacyPackedLogical(ins.Args[form.destination], "%"+out)
	}
	return true, false, c.storePackedMoveOperand(ins.Args[form.destination], byteWidth, "%"+out)
}

func (c *amd64Ctx) loadHorizontalFloatingOperand(operand Operand, byteWidth, laneBits int) (string, error) {
	bytesValue, err := c.loadPackedCompareBytes(operand, byteWidth)
	if err != nil {
		return "", err
	}
	lanes := byteWidth * 8 / laneBits
	result := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = bitcast <%d x i8> %s to %s\n", result, byteWidth, bytesValue, amd64FMA3LLVMType(lanes, laneBits))
	return "%" + result, nil
}

func (c *amd64Ctx) emitHorizontalFloating(spec amd64HorizontalFloatingSpec, byteWidth int, first, second string) string {
	lanes := byteWidth * 8 / spec.laneBits
	lanesPer128 := 128 / spec.laneBits
	typeName := amd64FMA3LLVMType(lanes, spec.laneBits)
	leftIndices, rightIndices := make([]string, lanes), make([]string, lanes)
	for lane := 0; lane < lanes; lane++ {
		position := lane % lanesPer128
		block := lane - position
		pair, sourceBase := position, 0
		if position >= lanesPer128/2 {
			sourceBase = lanes
			pair -= lanesPer128 / 2
		}
		leftIndex := sourceBase + block + 2*pair
		leftIndices[lane] = fmt.Sprintf("i32 %d", leftIndex)
		rightIndices[lane] = fmt.Sprintf("i32 %d", leftIndex+1)
	}
	left, right := c.newTmp(), c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = shufflevector %s %s, %s %s, <%d x i32> <%s>\n", left, typeName, second, typeName, first, lanes, strings.Join(leftIndices, ", "))
	fmt.Fprintf(c.b, "  %%%s = shufflevector %s %s, %s %s, <%d x i32> <%s>\n", right, typeName, second, typeName, first, lanes, strings.Join(rightIndices, ", "))
	// Project each independent 128-bit block without additional arithmetic.
	// One constrained vector operation retains joint exception semantics;
	// the reviewed integer-only binary helper preserves pair NaN priority.
	form := x86FloatAddSubForm{spec: x86FloatAddSubSpec{bits: spec.laneBits, subtract: spec.subtract}}
	return c.emitFloatAddSub(form, lanes, "%"+left, "%"+right, "")
}
