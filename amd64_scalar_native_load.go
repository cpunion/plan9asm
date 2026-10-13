package plan9asm

import "fmt"

// Register-addressed memory-to-GP rows of Go's ymovb/ymovw/ymovl/ymovq tables.
// Register, immediate, static-symbol, FP/AUTO, MMX/X and system-register rows
// keep their own consumers, including Go's distinct symbol-address roles.
// This is an ordinary native load shape, not a memory-ownership certificate.
type x86ScalarNativeLoadForm struct {
	source      Operand
	destination Reg
	bytes       int64
}

func x86ScalarNativeLoadProperties(op Op, ins Instr, arch string) (x86ScalarNativeLoadForm, bool, error) {
	local := ins
	local.Op = op
	canonical := x86GoMOVQSourceInstruction(local).Op
	width := x86RawReturnScalarMoveBytes(canonical)
	if width == 0 || len(ins.Args) != 2 || ins.Args[0].Kind != OpMem || ins.Args[1].Kind != OpReg {
		return x86ScalarNativeLoadForm{}, false, nil
	}
	source, destination := ins.Args[0], ins.Args[1].Reg
	if source.Mem.Sym != "" {
		return x86ScalarNativeLoadForm{}, false, nil
	}
	validRegister := isX86YrlRegisterForArch(destination, arch)
	if width == 1 {
		validRegister = isGoYmbRegisterForArch(destination, arch)
	}
	if !validRegister || !x86OrdinaryNativeMemoryOperand(source, arch) {
		return x86ScalarNativeLoadForm{}, false, nil
	}
	if arch == "386" && width == 8 || !x86PackedShiftMemoryValid(source, arch) {
		return x86ScalarNativeLoadForm{}, true, fmt.Errorf("invalid %s native scalar load: %q", arch, ins.Raw)
	}
	if width == 1 && arch != "386" &&
		(x86ExtensionHighRegister(source.Mem.Base) || x86ExtensionHighRegister(source.Mem.Index)) {
		// Go emits the actual REX byte even with AH/CH/DH/BH source
		// spellings. Their ModRM numbers then name SP/BP/SI/DI low bytes.
		switch destination {
		case AH, CH, DH, BH:
			destination = amd64YmbEffectiveRegister(destination, 16)
		}
	}
	return x86ScalarNativeLoadForm{source, destination, width}, true, nil
}

func (c *amd64Ctx) lowerScalarNativeLoad(form x86ScalarNativeLoadForm, ins Instr) error {
	if form.destination != ins.Args[1].Reg && (form.destination == SP || form.destination == BP) {
		return fmt.Errorf("%w: REX byte alias needs its physical SP/BP write contract", ErrProbeNeedsContext)
	}
	if ins.x86AddressBits != 0 || ins.x86UnboundAddressOverride {
		return fmt.Errorf("%w: scalar MOV needs its original native address size", ErrProbeNeedsContext)
	}
	pointer, pointerType, _, err := c.packedVectorMemoryPointer(form.source)
	if err != nil {
		return err
	}
	typ := amd64IntegerTypeForBits(int(form.bytes * 8))
	loaded := c.newTmp()
	// The fault/read remains observable even when the numerical result is
	// discarded. This asserts no pointee extent, alignment or nonaliasing.
	fmt.Fprintf(c.b, "  %%%s = load volatile %s, %s %s, align 1\n", loaded, typ, pointerType, pointer)
	return c.storeRegSized(form.destination, typ, "%"+loaded)
}

func x86LoweredScalarNativeLoad(ins Instr, state *x86ResultState, owned []bool, resultOffset int64, arch string) (bool, bool) {
	form, known, err := x86ScalarNativeLoadProperties(ins.Op, ins, arch)
	if !known || err != nil {
		return false, known
	}
	if form.destination == SP || form.destination == BP {
		return false, true
	}
	if !x86ResultReadWithNativeContinuation(ins, form.source, form.bytes, state, owned, arch, true) {
		return false, true
	}
	return x86ResultStore(Operand{Kind: OpReg, Reg: form.destination}, form.bytes, state, owned, resultOffset, arch), true
}
