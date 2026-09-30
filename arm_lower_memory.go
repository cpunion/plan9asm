package plan9asm

import (
	"fmt"
	"strings"
	"unicode"
)

// Go's integer load/store family shares one address grammar. Signed byte and
// halfword loads, and halfword stores, have only an unshifted register offset;
// word and the other byte forms also admit all four immediate shift kinds.
type armIntegerMemorySpec struct {
	bits   int
	signed bool
}

var armIntegerMemorySpecs = map[string]armIntegerMemorySpec{
	"MOVW": {bits: 32}, "MOVB": {bits: 8, signed: true},
	"MOVBS": {bits: 8, signed: true}, "MOVBU": {bits: 8},
	"MOVH": {bits: 16, signed: true}, "MOVHS": {bits: 16, signed: true}, "MOVHU": {bits: 16},
}

type armIntegerMemoryForm struct {
	spec      armIntegerMemorySpec
	base, reg Reg
	offset    Operand
	displace  int64
	load      bool
	postIndex bool
	writeback bool
	subtract  bool
}

type armIntegerSymbolForm struct {
	spec    armIntegerMemorySpec
	symbol  string
	reg     Reg
	load    bool
	address bool
}

func parseARMIntegerSymbolForm(op string, ins Instr) (armIntegerSymbolForm, error) {
	f := armIntegerSymbolForm{spec: armIntegerMemorySpecs[op]}
	if len(ins.Args) != 2 {
		return f, fmt.Errorf("arm %s symbol form expects two operands: %q", op, ins.Raw)
	}
	symbol, register := ins.Args[1], ins.Args[0]
	if ins.Args[0].Kind == OpSym {
		symbol, register, f.load = ins.Args[0], ins.Args[1], true
	}
	if symbol.Kind != OpSym || register.Kind != OpReg || !isARMGeneralReg(register.Reg) {
		return f, fmt.Errorf("arm %s symbol form requires a symbol and a general register: %q", op, ins.Raw)
	}
	f.symbol, f.reg = strings.TrimSpace(symbol.Sym), register.Reg
	f.address = strings.HasPrefix(f.symbol, "$")
	if f.address {
		if op != "MOVW" || !f.load {
			return f, fmt.Errorf("arm symbol address constants require MOVW $symbol(SB),register: %q", ins.Raw)
		}
		if err := armRequireConditionOnlySuffix(ins); err != nil {
			return f, err
		}
	} else if _, _, _, err := armMemoryModifiers(ins); err != nil {
		return f, err
	}
	if _, _, ok := parseSBRef(f.symbol); !ok || !strings.HasSuffix(f.symbol, "(SB)") {
		return f, fmt.Errorf("arm %s requires a valid symbol(SB) reference: %q", op, ins.Raw)
	}
	return f, nil
}

func (c *armCtx) lowerIntegerSymbolMove(op, cond string, ins Instr) (bool, bool, error) {
	if _, ok := armIntegerMemorySpecs[op]; !ok {
		return false, false, nil
	}
	hasSymbol := false
	for _, operand := range ins.Args {
		hasSymbol = hasSymbol || operand.Kind == OpSym
	}
	if !hasSymbol {
		return false, false, nil
	}
	f, err := parseARMIntegerSymbolForm(op, ins)
	if err != nil {
		return true, false, err
	}
	err = c.emitConditionalEffect(cond, func() error {
		ptr, err := c.ptrFromSB(f.symbol)
		if err != nil {
			return err
		}
		address := c.newTmp()
		fmt.Fprintf(c.b, "  %%%s = ptrtoint ptr %s to i32\n", address, ptr)
		if f.address {
			return c.storeReg(f.reg, "%"+address)
		}
		// Go's C_ADDR rows materialize the address in REGTMP (R11) before
		// the actual load/store. P/W/U operate on a zero displacement from
		// that scratch address, not on the symbol's relocation addend.
		if err := c.storeReg("R11", "%"+address); err != nil {
			return err
		}
		if f.load {
			value, err := c.loadMemoryPointer(ptr, f.spec.bits, f.spec.signed)
			if err != nil {
				return err
			}
			return c.storeReg(f.reg, value)
		}
		value, err := c.loadReg(f.reg)
		if err != nil {
			return err
		}
		return c.storeMemoryPointer(ptr, f.spec.bits, value)
	})
	return true, false, err
}

func armMemoryShift(mem MemRef) (Operand, bool) {
	reg, shift, amount, countReg, ok := parseRegShift(mem.OffRaw)
	return Operand{Kind: OpRegShift, Reg: reg, ShiftOp: shift, ShiftAmount: amount, ShiftReg: countReg}, ok
}

// Go's C_AUTO address names annotate SP slots; they are not unresolved
// displacement expressions. Keep the already parsed numeric offset intact.
func armNamedStackOffset(mem MemRef) bool {
	if mem.Base != SP || mem.Index != "" || mem.OffRaw == "" {
		return false
	}
	name, offset := splitSymPlusOff(mem.OffRaw)
	if name == "" || offset != mem.Off {
		return false
	}
	for i, ch := range name {
		if ch != '_' && ch != '·' && !unicode.IsLetter(ch) && (i == 0 || !unicode.IsDigit(ch)) {
			return false
		}
	}
	return true
}

func parseARMIntegerMemoryForm(op string, ins Instr) (armIntegerMemoryForm, error) {
	f := armIntegerMemoryForm{spec: armIntegerMemorySpecs[op]}
	if len(ins.Args) != 2 {
		return f, fmt.Errorf("arm %s memory form expects two operands: %q", op, ins.Raw)
	}
	memory, register := ins.Args[1], ins.Args[0]
	if ins.Args[0].Kind == OpMem {
		memory, register, f.load = ins.Args[0], ins.Args[1], true
	}
	if memory.Kind != OpMem || register.Kind != OpReg || !isARMGeneralReg(register.Reg) || !isARMGeneralReg(memory.Mem.Base) {
		return f, fmt.Errorf("arm %s memory form requires an address and a general register: %q", op, ins.Raw)
	}
	f.base, f.reg = memory.Mem.Base, register.Reg
	if armNamedStackOffset(memory.Mem) {
		memory.Mem.OffRaw = ""
	}
	post, write, down, err := armMemoryModifiers(ins)
	if err != nil {
		return f, err
	}
	f.postIndex, f.writeback = post, post || write
	f.offset = Operand{Kind: OpImm, Imm: memory.Mem.Off}
	fullShift := op == "MOVW" || f.load && op == "MOVBU" || !f.load && f.spec.bits == 8
	if memory.Mem.OffRaw != "" {
		if memory.Mem.Index != "" || memory.Mem.Off != 0 {
			return f, fmt.Errorf("arm %s mixes incompatible address offset forms: %q", op, ins.Raw)
		}
		shift, ok := armMemoryShift(memory.Mem)
		if !ok || !isARMGeneralReg(shift.Reg) || shift.ShiftReg != "" || shift.ShiftAmount < 0 || shift.ShiftAmount > 31 {
			return f, fmt.Errorf("arm %s has an invalid immediate register-offset shift: %q", op, ins.Raw)
		}
		if !fullShift && (shift.ShiftOp != ShiftLeft || shift.ShiftAmount != 0) {
			return f, fmt.Errorf("arm %s signed-byte/halfword address only permits R<<0: %q", op, ins.Raw)
		}
		f.offset, f.subtract = shift, down
	} else if memory.Mem.Index != "" {
		// Preserve the existing structured MemRef API; source Go ARM spellings
		// reach the shift branch above, not x86/ARM64 parenthesized indexes.
		if !isARMGeneralReg(memory.Mem.Index) || memory.Mem.IndexExt != "" {
			return f, fmt.Errorf("arm %s has an invalid address index: %q", op, ins.Raw)
		}
		amount := int64(0)
		for scale := memory.Mem.Scale; scale > 1 && scale%2 == 0; scale /= 2 {
			amount++
		}
		if memory.Mem.Scale != 0 && memory.Mem.Scale != int64(1)<<uint(amount) || amount > 31 {
			return f, fmt.Errorf("arm %s has an invalid address scale: %q", op, ins.Raw)
		}
		f.offset = Operand{Kind: OpRegShift, Reg: memory.Mem.Index, ShiftOp: ShiftLeft, ShiftAmount: amount}
		f.displace, f.subtract = memory.Mem.Off, down
	} else {
		// Go's olr rejects .U on negative immediate offsets. olhr (signed
		// byte / halfword immediates) ignores .U, unlike its C_SHIFTADDR rows.
		if down && fullShift && memory.Mem.Off < 0 {
			return f, fmt.Errorf("arm %s .U on negative offset: %q", op, ins.Raw)
		}
		f.subtract = down && fullShift
	}
	return f, nil
}

func (c *armCtx) integerMemoryAddress(f armIntegerMemoryForm) (address, updated string, err error) {
	base, err := c.loadReg(f.base)
	if err != nil {
		return "", "", err
	}
	var delta string
	if f.offset.Kind == OpRegShift && f.offset.ShiftAmount == 0 && f.offset.ShiftOp != ShiftLeft {
		value, loadErr := c.loadReg(f.offset.Reg)
		if loadErr != nil {
			return "", "", loadErr
		}
		switch f.offset.ShiftOp {
		case ShiftRight:
			delta = "0" // encoded LSR #0 means LSR #32
		case ShiftArith:
			tmp := c.newTmp()
			fmt.Fprintf(c.b, "  %%%s = ashr i32 %s, 31\n", tmp, value)
			delta = "%" + tmp
		case ShiftRotate:
			carry := c.loadFlagValue(c.flagsCSlot)
			ext, high, low, result := c.newTmp(), c.newTmp(), c.newTmp(), c.newTmp()
			fmt.Fprintf(c.b, "  %%%s = zext i1 %s to i32\n  %%%s = shl i32 %%%s, 31\n", ext, carry, high, ext)
			fmt.Fprintf(c.b, "  %%%s = lshr i32 %s, 1\n  %%%s = or i32 %%%s, %%%s\n", low, value, result, low, high)
			delta = "%" + result // encoded ROR #0 means RRX, without updating C
		}
	} else {
		delta, err = c.eval32(f.offset, false)
		if err != nil {
			return "", "", err
		}
	}
	if f.displace != 0 {
		tmp := c.newTmp()
		fmt.Fprintf(c.b, "  %%%s = add i32 %s, %d\n", tmp, delta, f.displace)
		delta = "%" + tmp
	}
	op := "add"
	if f.subtract {
		op = "sub"
	}
	tmp := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = %s i32 %s, %s\n", tmp, op, base, delta)
	updated = "%" + tmp
	address = updated
	if f.postIndex {
		address = base
	}
	return address, updated, nil
}

func (c *armCtx) lowerIntegerMemoryMove(op, cond string, ins Instr) (bool, bool, error) {
	if _, ok := armIntegerMemorySpecs[op]; !ok {
		return false, false, nil
	}
	hasMemory := false
	for _, operand := range ins.Args {
		hasMemory = hasMemory || operand.Kind == OpMem
	}
	if !hasMemory {
		return false, false, nil
	}
	f, err := parseARMIntegerMemoryForm(op, ins)
	if err != nil {
		return true, false, err
	}
	err = c.emitConditionalEffect(cond, func() error {
		address, updated, err := c.integerMemoryAddress(f)
		if err != nil {
			return err
		}
		var value string
		if f.load {
			value, err = c.loadMemAddress(address, f.spec.bits, f.spec.signed)
		} else {
			value, err = c.loadReg(f.reg)
			if err == nil {
				err = c.storeMemAddress(address, f.spec.bits, value)
			}
		}
		if err != nil {
			return err
		}
		if f.writeback {
			if err := c.storeReg(f.base, updated); err != nil {
				return err
			}
		}
		if f.load {
			return c.storeReg(f.reg, value)
		}
		return nil
	})
	return true, false, err
}
