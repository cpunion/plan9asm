package plan9asm

import "fmt"

// The complete shared grammar retains each source/data/mask role. Native
// Type4 and E4 reads are exact, observable and unaligned; no speculative
// eager memory load may replace EVEX active-lane suppression.
func (c *amd64Ctx) lowerPackedAbsSign(op Op, ins Instr) (bool, bool, error) {
	form, known, err := x86PackedAbsSignProperties(op, ins, c.goarch)
	if !known || err != nil {
		return known, false, err
	}
	if err := c.validateConflictFunnelMemory(ins, form.x86PackedBinaryForm); err != nil {
		return true, false, err
	}
	mask := ""
	if form.mask >= 0 {
		mask, err = c.loadK(ins.Args[form.mask].Reg)
		if err != nil {
			return true, false, err
		}
	}
	bits, width := form.spec.laneBits, form.byteWidth
	first, err := c.loadX86PackedShiftLanes(ins.Args[form.first], width, bits, form.broadcast, mask)
	if err != nil {
		return true, false, err
	}
	result := ""
	if form.spec.sign {
		data, err := c.loadPackedCompareLanes(ins.Args[form.second], width, bits, false)
		if err != nil {
			return true, false, err
		}
		result = c.emitPackedSign(first, data, width, bits)
	} else {
		result = c.emitPackedAbs(first, width, bits)
	}
	lanes := width * 8 / bits
	destination := ins.Args[form.destination]
	if form.mask >= 0 {
		old := "zeroinitializer"
		if !form.zeroing {
			oldBytes, err := c.loadPackedCompareBytes(destination, width)
			if err != nil {
				return true, false, err
			}
			old = c.bitcastVectorBytesToIntegerLanes(width, lanes, bits, oldBytes)
		}
		result = amd64ApplyIntegerLaneMask(c, lanes, bits, result, old, mask, form.zeroing)
	}
	out := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = bitcast <%d x i%d> %s to <%d x i8>\n", out, lanes, bits, result, width)
	if form.legacy {
		return true, false, c.storeLegacyXViews(destination.Reg, "%"+out)
	}
	return true, false, c.storePackedMoveOperand(destination, width, "%"+out)
}

func (c *amd64Ctx) emitPackedAbs(values string, byteWidth, laneBits int) string {
	lanes := byteWidth * 8 / laneBits
	negative := c.newTmp()
	negated := c.newTmp()
	result := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = icmp slt <%d x i%d> %s, zeroinitializer\n", negative, lanes, laneBits, values)
	// No nsw: ABS of the most negative signed value wraps to itself.
	fmt.Fprintf(c.b, "  %%%s = sub <%d x i%d> zeroinitializer, %s\n", negated, lanes, laneBits, values)
	fmt.Fprintf(c.b, "  %%%s = select <%d x i1> %%%s, <%d x i%d> %%%s, <%d x i%d> %s\n", result, lanes, negative, lanes, laneBits, negated, lanes, laneBits, values)
	return "%" + result
}

func (c *amd64Ctx) emitPackedSign(control, data string, width, bits int) string {
	lanes := width * 8 / bits
	positive := c.newTmp()
	negative := c.newTmp()
	negated := c.newTmp()
	nonpositive := c.newTmp()
	result := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = icmp sgt <%d x i%d> %s, zeroinitializer\n", positive, lanes, bits, control)
	fmt.Fprintf(c.b, "  %%%s = icmp slt <%d x i%d> %s, zeroinitializer\n", negative, lanes, bits, control)
	fmt.Fprintf(c.b, "  %%%s = sub <%d x i%d> zeroinitializer, %s\n", negated, lanes, bits, data)
	fmt.Fprintf(c.b, "  %%%s = select <%d x i1> %%%s, <%d x i%d> %%%s, <%d x i%d> zeroinitializer\n", nonpositive, lanes, negative, lanes, bits, negated, lanes, bits)
	fmt.Fprintf(c.b, "  %%%s = select <%d x i1> %%%s, <%d x i%d> %s, <%d x i%d> %%%s\n", result, lanes, positive, lanes, bits, data, lanes, bits, nonpositive)
	return "%" + result
}
