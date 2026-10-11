package plan9asm

import (
	"fmt"
	"strings"
)

func (c *amd64Ctx) lowerPackedZeroExtendMove(op Op, ins Instr) (bool, bool, error) {
	if ins.Op != "" && !strings.EqualFold(string(ins.Op), string(op)) {
		return true, false, fmt.Errorf("packed zero-extension dispatch/source opcode mismatch: %s/%s", op, ins.Op)
	}
	ins.Op = op // Preserve the pre-existing helper's empty-op dispatch contract.
	form, recognized, err := x86PackedZeroExtendProperties(ins, c.goarch)
	if !recognized || err != nil {
		return recognized, false, err
	}
	source, destination := ins.Args[0], ins.Args[form.destination]
	lanes := form.byteWidth * 8 / form.spec.outputBits
	if source.Kind == OpFP {
		if err := c.validateFPVectorSpan(source.FPOffset, form.memoryBytes); err != nil {
			return true, false, err
		}
	}
	var mask, inputs string
	if form.mask >= 0 {
		mask, err = c.loadK(ins.Args[form.mask].Reg)
		if err != nil {
			return true, false, err
		}
	}
	switch {
	case source.Kind == OpFP:
		// Independent ABI0 fields are not a contiguous native vector. Read
		// each declared byte and only then assemble the integer source lanes.
		var bytes string
		bytes, err = c.loadFPVectorBytes(source.FPOffset, form.memoryBytes)
		if err == nil {
			inputs = c.bitcastVectorBytesToIntegerLanes(form.memoryBytes, lanes, form.spec.inputBits, bytes)
		}
	case form.memory >= 0:
		// E5 inactive lanes suppress reads; active accesses remain observable
		// even if the widened result is dead. Read the shrinking source tuple,
		// not its larger physical vector container or widened destination.
		inputs, err = c.loadX86PackedShiftLanes(source, form.memoryBytes, form.spec.inputBits, false, mask)
	default:
		inputs, err = c.loadPackedExtendInputs(source, form.sourceRegisterBytes, lanes, form.spec.inputBits)
	}
	if err != nil {
		return true, false, err
	}
	result := "zeroinitializer"
	for lane := 0; lane < lanes; lane++ {
		input, extended, inserted := c.newTmp(), c.newTmp(), c.newTmp()
		fmt.Fprintf(c.b, "  %%%s = extractelement <%d x i%d> %s, i32 %d\n", input, lanes, form.spec.inputBits, inputs, lane)
		fmt.Fprintf(c.b, "  %%%s = zext i%d %%%s to i%d\n", extended, form.spec.inputBits, input, form.spec.outputBits)
		fmt.Fprintf(c.b, "  %%%s = insertelement <%d x i%d> %s, i%d %%%s, i32 %d\n", inserted, lanes, form.spec.outputBits, result, form.spec.outputBits, extended, lane)
		result = "%" + inserted
	}
	if form.mask >= 0 {
		old := "zeroinitializer"
		if !form.zeroing {
			var bytes string
			bytes, err = c.loadPackedCompareBytes(destination, form.byteWidth)
			if err != nil {
				return true, false, err
			}
			old = c.bitcastVectorBytesToIntegerLanes(form.byteWidth, lanes, form.spec.outputBits, bytes)
		}
		result = amd64ApplyIntegerLaneMask(c, lanes, form.spec.outputBits, result, old, mask, form.zeroing)
	}
	out := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = bitcast <%d x i%d> %s to <%d x i8>\n", out, lanes, form.spec.outputBits, result, form.byteWidth)
	if !form.spec.vector {
		return true, false, c.storeLegacyXViews(destination.Reg, "%"+out)
	}
	return true, false, c.storePackedMoveOperand(destination, form.byteWidth, "%"+out)
}
