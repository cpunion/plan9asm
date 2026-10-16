package plan9asm

import "fmt"

// Go's Zr_m GP store rows have the same width/register/address classes as
// Zm_r loads. Keep the operand projection local: source identity, FP/AUTO,
// static symbols, segments and the independent SIMD/MMX rows are unchanged.
type x86ScalarNativeStoreForm struct {
	source      Reg
	destination Operand
	bytes       int64
}

func x86ScalarNativeStoreProperties(op Op, ins Instr, arch string) (x86ScalarNativeStoreForm, bool, error) {
	if len(ins.Args) != 2 || ins.Args[0].Kind != OpReg || ins.Args[1].Kind != OpMem ||
		ins.Args[1].Mem.Sym != "" || ins.Args[1].Mem.Segment != "" {
		return x86ScalarNativeStoreForm{}, false, nil
	}
	projected := ins
	projected.Args = []Operand{ins.Args[1], ins.Args[0]}
	form, known, err := x86ScalarNativeLoadProperties(op, projected, arch)
	return x86ScalarNativeStoreForm{form.destination, form.source, form.bytes}, known, err
}

func (c *amd64Ctx) lowerScalarNativeStore(form x86ScalarNativeStoreForm, ins Instr) error {
	if form.source != ins.Args[0].Reg && (form.source == SP || form.source == BP) {
		return fmt.Errorf("%w: REX byte alias needs its physical SP/BP read contract", ErrProbeNeedsContext)
	}
	if ins.x86AddressBits != 0 || ins.x86UnboundAddressOverride {
		return fmt.Errorf("%w: scalar MOV needs its original native address size", ErrProbeNeedsContext)
	}
	typ := amd64IntegerTypeForBits(int(form.bytes * 8))
	value, err := c.evalIntSized(Operand{Kind: OpReg, Reg: form.source}, typ)
	if err != nil {
		return err
	}
	pointer, pointerType, _, err := c.packedVectorMemoryPointer(form.destination)
	if err != nil {
		return err
	}
	// Preserve even a dead write/fault, with no alignment, extent or alias
	// assumptions. Continuing this one effect grants no reusable memory fact.
	fmt.Fprintf(c.b, "  store volatile %s %s, %s %s, align 1\n", typ, value, pointerType, pointer)
	return nil
}

func x86LoweredScalarNativeStore(ins Instr, state *x86ResultState, owned []bool, arch string) (bool, bool) {
	form, known, err := x86ScalarNativeStoreProperties(ins.Op, ins, arch)
	if !known || err != nil {
		return false, known
	}
	if !x86ResultNamedPredicateInstruction(ins) || ins.x86NumericalRead != nil ||
		x86LogicalSourceFrameRegister(form.source, arch) ||
		!x86ResultIncomingPointerAddressIdentity(form.destination, state, arch) {
		return false, true
	}
	// A native NUM store changes no GP/vector/FLAGS/DF or logical bytes.
	// Pointer-as-data requires its separate GC/escape contract. Standalone
	// declaration, caller, raw and frame observers cannot use this consumer.
	return x86ResultRead(Operand{Kind: OpReg, Reg: form.source}, form.bytes, state, owned, arch), true
}
