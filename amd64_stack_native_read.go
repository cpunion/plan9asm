package plan9asm

import "fmt"

func (c *amd64Ctx) loadStackValue(source Operand, bits int) (string, error) {
	typ := amd64StackLLVMType(bits)
	if !x86RegisterAddressedNativeMemoryOperand(source, c.goarch) {
		return c.evalIntSized(source, typ)
	}
	pointer, pointerType, _, err := c.packedVectorMemoryPointer(source)
	if err != nil {
		return "", err
	}
	value := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = load volatile %s, %s %s, align 1\n", value, typ, pointerType, pointer)
	return "%" + value, nil
}

func x86LoweredStackNativeRead(ins Instr, state *x86ResultState, owned []bool,
	resultOffset int64, arch string) (bool, bool) {
	spec, known, err := x86StackWidthProperties(ins, arch)
	if !known || err != nil || spec.mode != amd64StackPushValue || len(ins.Args) != 1 ||
		!x86RegisterAddressedNativeMemoryOperand(ins.Args[0], arch) {
		return false, false
	}
	if !x86ResultNamedPredicateInstruction(ins) || ins.x86NumericalRead != nil ||
		!x86ResultIncomingPointerReadAddressIdentity(ins.Args[0], state, arch) {
		return false, true
	}
	// Only this exact read gets a transfer-local numerical view. The clone
	// and the defer both revoke it before any successor observes the state.
	state.literalMemoryOperand, state.literalMemoryBytes = &ins.Args[0], int64(spec.bits/8)
	defer func() { state.literalMemoryOperand, state.literalMemoryBytes = nil, 0 }()
	return x86ResultValueStackStep(ins, state, owned, resultOffset, arch)
}
