package plan9asm

import (
	"fmt"
	"math"
)

// A named GP move's FP operand describes instruction bytes, not a coercion to
// the first logical field. Keep exact natural scalar transfers on their old
// routes (in particular, opaque pointers). A new partial/interior/cross-field
// view must be handled here, including a failed proof; it cannot fall back to
// the legacy first-slot conversion or an unresolved-offset zero.
func (c *amd64Ctx) integerFPByteView(off, width int64, write bool) (bool, error) {
	if c.goarch == "386" && width == 8 {
		return true, fmt.Errorf("%w: mode32 has no GP qword FP move; X/m64 retains its separate typed consumer", ErrProbeNeedsContext)
	}
	if c.fpScalarSource.x86SourceStack != nil || c.fpScalarSource.x86FixedLocal != nil ||
		len(c.sig.Frame.Params)+len(c.fpResults) == 0 {
		return false, nil // Independently bound physical storage retains its consumer.
	}
	for _, slots := range [][]FrameSlot{c.sig.Frame.Params, c.fpResults} {
		for _, slot := range slots {
			if slot.Offset == off && x86FrameScalarBytes(slot.Type, c.goarch) == width {
				return false, nil
			}
		}
	}
	fail := func(reason string) (bool, error) {
		return true, fmt.Errorf("%w: integer FP byte view +%d/%d: %s", ErrProbeNeedsContext, off, width, reason)
	}
	if width != 1 && width != 2 && width != 4 && width != 8 || off < 0 || off > math.MaxInt64-width {
		return fail("invalid or overflowing span")
	}
	for current := off; current < off+width; current++ {
		matches := 0
		for _, group := range []struct {
			slots  []FrameSlot
			result bool
		}{{c.sig.Frame.Params, false}, {c.fpResults, true}} {
			for _, slot := range group.slots {
				size := x86FrameScalarBytes(slot.Type, c.goarch)
				if size <= 0 || slot.Offset < 0 || slot.Offset > math.MaxInt64-size {
					return fail("invalid scalar field extent")
				}
				if current < slot.Offset || current >= slot.Offset+size {
					continue
				}
				if slot.Type == Ptr || write && !group.result {
					return fail("pointer bits and parameter writes retain independent roles")
				}
				matches++
			}
		}
		if matches != 1 {
			return fail(fmt.Sprintf("byte %d has %d scalar fields", current, matches))
		}
	}
	// Numerical definedness, canonical bool bytes and all source/native roles
	// are checked once for the complete original function, not inferred from the
	// layout or from zero-initialized implementation allocas.
	if err := c.scalarFPResultInitialized(); err != nil {
		return true, err
	}
	for current := off; current < off+width; current++ {
		slot, _, _ := c.fpVectorSlot(current)
		if slot.Type == I1 && !c.abi0BoolResultByteStorage() {
			return fail("physical bool bytes lack a checked canonical role")
		}
	}
	return true, nil
}

// This dispatch is only a cost/transport selector for the existing named GP
// move lowerer. Vector/MMX forms keep their own grammar and permissions; Go's
// MOVD/MOVDQ2Q aliases retain their original instruction identity.
func x86IntegerFPMoveWidth(original Instr, arch string) int64 {
	ins := x86GoMOVQSourceInstruction(original)
	for _, operand := range ins.Args {
		if operand.Kind != OpReg {
			continue
		}
		if _, ok := amd64ParseXReg(operand.Reg); ok {
			return 0
		}
		if _, ok := amd64ParseYReg(operand.Reg); ok {
			return 0
		}
		if _, ok := amd64ParseZReg(operand.Reg); ok {
			return 0
		}
		if _, ok := amd64ParseMReg(operand.Reg); ok {
			return 0
		}
	}
	switch ins.Op {
	case "MOVB":
		return 1
	case "MOVW":
		return 2
	case "MOVL":
		return 4
	case "MOVQ":
		if arch != "386" {
			return 8
		}
	}
	return 0
}

func (c *amd64Ctx) loadIntegerFPBytes(off int64, typ LLVMType) (string, bool, error) {
	bits, _ := amd64IntegerTypeBits(typ)
	handled, err := c.integerFPByteView(off, int64(bits/8), false)
	if !handled || err != nil {
		return "", handled, err
	}
	bytes, err := c.loadFPVectorBytesMode(off, bits/8, true)
	if err != nil {
		return "", true, err
	}
	value := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = bitcast <%d x i8> %s to %s\n", value, bits/8, bytes, typ)
	return "%" + value, true, nil
}

func (c *amd64Ctx) storeIntegerFPBytes(off int64, typ LLVMType, value string) (bool, error) {
	bits, _ := amd64IntegerTypeBits(typ)
	handled, err := c.integerFPByteView(off, int64(bits/8), true)
	if !handled || err != nil {
		return handled, err
	}
	bytes := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = bitcast %s %s to <%d x i8>\n", bytes, typ, value, bits/8)
	return true, c.storeFPVectorBytesMode(off, bits/8, "%"+bytes, "", true)
}
