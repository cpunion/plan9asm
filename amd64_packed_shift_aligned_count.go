package plan9asm

import "fmt"

func x86PackedShiftHasAlignedCount(form x86PackedShiftForm) bool {
	return form.legacy && !form.mmx && form.memory == form.count
}

// Legacy SSE uniform shifts require an aligned m128 count read, including the
// ignored upper qword. A plain LLVM aligned load makes misalignment undefined;
// this actual SSE2 read instead preserves the native alignment/page fault.
// Its vector output and memory clobber keep both the complete read and value
// dependency observable. It does not grant the source address any permission.
func (c *amd64Ctx) loadX86PackedShiftAlignedCount(source Operand) (string, error) {
	pointer, pointerType, _, err := c.packedVectorMemoryPointer(source)
	if err != nil {
		return "", err
	}
	value := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = call <16 x i8> asm sideeffect %q, %q(%s %s)\n",
		value, "movdqa ($1), $0", "=x,r,~{memory}", pointerType, pointer)
	return "%" + value, nil
}
