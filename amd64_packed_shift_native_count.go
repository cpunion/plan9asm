package plan9asm

import "fmt"

func x86PackedShiftHasNativeCount(form x86PackedShiftForm) bool {
	return form.legacy && !form.mmx && form.memory == form.count
}

// Preserve the original SSE2 m128 operation, including its complete native
// read and fault behavior. Substituting MOVDQA would impose its alignment
// contract on a different instruction. The tied data input, vector output
// and memory clobber keep the original operation observable without granting
// the source address any permission or claiming an alignment guarantee.
func (c *amd64Ctx) emitX86PackedShiftNativeCount(form x86PackedShiftForm, source Operand, data string) (string, error) {
	if !x86PackedShiftHasNativeCount(form) || form.byteWidth != 16 || form.countMode != x86PackedShiftUniform {
		return "", fmt.Errorf("expected legacy SSE m128 packed count form")
	}
	name := "psrl"
	if form.left {
		name = "psll"
	}
	switch form.laneBits {
	case 16:
		name += "w"
	case 32:
		name += "d"
	case 64:
		name += "q"
	default:
		return "", fmt.Errorf("unsupported legacy packed count lane width %d", form.laneBits)
	}
	pointer, pointerType, _, err := c.packedVectorMemoryPointer(source)
	if err != nil {
		return "", err
	}
	segment := ""
	switch pointerType {
	case "ptr addrspace(256)":
		segment = "%gs:"
	case "ptr addrspace(257)":
		segment = "%fs:"
	}
	value := c.newTmp()
	typeName := fmt.Sprintf("<%d x i%d>", 128/form.laneBits, form.laneBits)
	fmt.Fprintf(c.b, "  %%%s = call %s asm sideeffect %q, %q(%s %s, %s %s)\n",
		value, typeName, name+" "+segment+"($2), $0", "=x,0,r,~{memory}", typeName, data, pointerType, pointer)
	return "%" + value, nil
}
