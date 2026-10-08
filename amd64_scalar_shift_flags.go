package plan9asm

import "fmt"

// Emit only architecturally defined flag values. Undefined flags keep an old
// representative here but lose availability in the shared typed effect. This
// is numerical virtual state, never native flags/entry restoration authority.
func (c *amd64Ctx) emitScalarShiftFlags(stem string, bits int, before, result, count string, double bool) {
	ty := amd64IntegerTypeForBits(bits)
	predicate := func(comparison string, amount int) string {
		name := c.newTmp()
		fmt.Fprintf(c.b, "  %%%s = icmp %s i64 %s, %d\n", name, comparison, count, amount)
		return "%" + name
	}
	nonzero, one := predicate("ne", 0), predicate("eq", 1)
	valid := nonzero
	if double && bits == 16 {
		bounded := predicate("ule", bits)
		name := c.newTmp()
		fmt.Fprintf(c.b, "  %%%s = and i1 %s, %s\n", name, nonzero, bounded)
		valid = "%" + name
	}
	bit := func(value, index string) string {
		if bits != 64 {
			truncated := c.newTmp()
			fmt.Fprintf(c.b, "  %%%s = trunc i64 %s to %s\n", truncated, index, ty)
			index = "%" + truncated
		}
		shifted, flag := c.newTmp(), c.newTmp()
		fmt.Fprintf(c.b, "  %%%s = lshr %s %s, %s\n", shifted, ty, value, index)
		fmt.Fprintf(c.b, "  %%%s = trunc %s %%%s to i1\n", flag, ty, shifted)
		return "%" + flag
	}
	write := func(slot, value, when, old string) {
		if old == "" {
			old = c.loadFlag(slot)
		}
		selected := c.newTmp()
		fmt.Fprintf(c.b, "  %%%s = select i1 %s, i1 %s, i1 %s\n", selected, when, value, old)
		fmt.Fprintf(c.b, "  store i1 %%%s, ptr %s\n", selected, slot)
	}
	msb := bit(result, fmt.Sprint(bits-1))
	var cf, overflow string
	if !double && (stem == "ROL" || stem == "ROR") {
		if stem == "ROL" {
			cf = bit(result, "0")
		} else {
			cf = msb
		}
		write(c.flagsCFSlot, cf, nonzero, "")
		other := cf
		if stem == "ROR" {
			other = bit(result, fmt.Sprint(bits-2))
		}
		name := c.newTmp()
		fmt.Fprintf(c.b, "  %%%s = xor i1 %s, %s\n", name, msb, other)
		overflow = "%" + name
	} else {
		// Clamp only the bit-index calculation, so count0 and out-of-range
		// B/W counts cannot produce LLVM poison even on an inactive select.
		zero := predicate("eq", 0)
		minimum := c.newTmp()
		fmt.Fprintf(c.b, "  %%%s = select i1 %s, i64 1, i64 %s\n", minimum, zero, count)
		large := predicate("ugt", bits)
		bounded := c.newTmp()
		fmt.Fprintf(c.b, "  %%%s = select i1 %s, i64 %d, i64 %%%s\n", bounded, large, bits, minimum)
		index := c.newTmp()
		if stem == "SHL" || stem == "SAL" {
			fmt.Fprintf(c.b, "  %%%s = sub i64 %d, %%%s\n", index, bits, bounded)
		} else {
			fmt.Fprintf(c.b, "  %%%s = sub i64 %%%s, 1\n", index, bounded)
		}
		cf = bit(before, "%"+index)
		cfValid := valid
		if !double && stem != "SAR" && bits < 32 {
			bounded := predicate("ult", bits)
			name := c.newTmp()
			fmt.Fprintf(c.b, "  %%%s = and i1 %s, %s\n", name, valid, bounded)
			cfValid = "%" + name
		}
		write(c.flagsCFSlot, cf, cfValid, "")
		oldMSB := bit(before, fmt.Sprint(bits-1))
		if double {
			name := c.newTmp()
			fmt.Fprintf(c.b, "  %%%s = xor i1 %s, %s\n", name, oldMSB, msb)
			overflow = "%" + name
		} else if stem == "SAR" {
			overflow = "false"
		} else if stem == "SHR" {
			overflow = oldMSB
		} else {
			name := c.newTmp()
			fmt.Fprintf(c.b, "  %%%s = xor i1 %s, %s\n", name, msb, cf)
			overflow = "%" + name
		}
		zf, sf := c.newTmp(), c.newTmp()
		fmt.Fprintf(c.b, "  %%%s = icmp eq %s %s, 0\n", zf, ty, result)
		fmt.Fprintf(c.b, "  %%%s = icmp slt %s %s, 0\n", sf, ty, result)
		write(c.flagsZSlot, "%"+zf, valid, "")
		write(c.flagsSltSlot, "%"+sf, valid, "")
		oldPF := c.loadFlag(c.flagsPFSlot)
		c.setParityFlagSized(ty, result)
		write(c.flagsPFSlot, c.loadFlag(c.flagsPFSlot), valid, oldPF)
	}
	write(c.flagsOFSlot, overflow, one, "")
}
