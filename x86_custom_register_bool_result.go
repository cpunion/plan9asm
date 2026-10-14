package plan9asm

import (
	"fmt"
	"strings"
)

// Explicit ArgRegs is a logical translator transport, not Go ABIInternal or
// ABI0. It supplies only the declared register bytes and an opaque pointer
// identity; no FP/AUTO, native entry, caller backing or spill storage exists.
func checkX86CustomRegisterBoolResultInitialization(fn Func, sig FuncSig, arch string, transfer x86ResultSourceTransfer) error {
	fail := func(reason string) error {
		return fmt.Errorf("%w: custom register bool result: %s", ErrProbeNeedsContext, reason)
	}
	if arch != "386" || sig.Ret != I1 || len(sig.ArgRegs) == 0 || len(sig.ArgRegs) != len(sig.Args) || len(sig.Args) > 6 ||
		len(sig.Frame.Params) != 0 || len(sig.Frame.Results) != 0 || fn.FrameSize != 0 || fn.ArgSize != 0 ||
		sig.ARMEntry != nil || sig.ARM64GoRegisterABI != nil || sig.ARM64ClosureABI != nil || sig.WASMContext != "" || sig.WASMNative ||
		strings.Contains(fn.Sym, "<") || len(fn.X86RawText) != 0 || fn.x86SourceStack != nil || fn.x86FixedLocal != nil ||
		fn.x86DirectTailRoot != nil || fn.x86BoolCalls != nil || sig.x86BorrowedBoolTail {
		return fail("requires a complete distinct ordinary mode32 logical register transport")
	}
	for _, ins := range fn.Instrs {
		if ins.x86Encoded || ins.x86SourceRaw != nil || ins.x86RawNearReturn != nil ||
			ins.x86AddressBits != 0 || ins.x86UnboundAddressOverride || ins.x86SegmentPrefix != 0 {
			return fail("raw/native source requires its separate contract")
		}
	}
	initial := x86ResultState{}
	occupied := [16]bool{}
	for index, typ := range sig.Args {
		width := int64(0)
		switch typ {
		case I1, I8:
			width = 1
		case I16:
			width = 2
		case I32, Ptr:
			width = 4
		}
		reg := sig.ArgRegs[index]
		root, _, valid := x86ResultGP(reg, 4, arch)
		if !valid || occupied[root] || width == 0 {
			return fail("arguments require unique whole-word register assignments and scalar types")
		}
		occupied[root] = true
		if typ == Ptr {
			initial.pointerGP[root], initial.callerAddressGP[root] = true, true
			continue
		}
		_, mask, valid := x86ResultGP(reg, width, arch)
		if !valid {
			return fail("narrow argument has no registered physical byte view")
		}
		initial.gp[root] = mask // No unspecified upper byte becomes initialized.
		if typ == I1 {
			initial.canonicalGP[root] = mask
		}
	}
	original := transfer
	transfer = func(pc int, ins Instr, state *x86ResultState, owned []bool, resultOffset int64, arch string) bool {
		before := state.canonicalGP
		if !original(pc, ins, state, owned, resultOffset, arch) {
			return false
		}
		state.canonicalGP = x86RegisterCanonicalBytesAfter(ins, before, arch)
		for index := range state.gp {
			state.canonicalGP[index] &= state.gp[index]
		}
		return true
	}
	return checkX86ResultSourceFlow(fn, initial, nil, 0, arch, transfer,
		func(state x86ResultState, pc int) error {
			root, mask, _ := x86ResultGP(AX, 1, arch)
			if state.gp[root]&mask != mask || state.canonicalGP[root]&mask != mask {
				return fail(fmt.Sprintf("AX low result byte is not initialized and canonical on every path to RET %d", pc))
			}
			return nil
		})
}

// This lowering-only companion consumes the complete existing CMPXCHG rows.
// Memory forms retain the native sideeffect read/write opcode, including faults
// and unaligned behavior. Continued execution supplies result/FLAGS bytes, not
// validity, alignment, bounds, reusable memory or native-entry permission.
func x86CustomRegisterBoolCompareExchangeStep(fn Func, ordinal int, ins Instr, state *x86ResultState, arch string) (bool, bool) {
	if ins.Op == "LOCK" {
		if len(ins.Args) != 0 || ordinal < 0 || ordinal+1 >= len(fn.Instrs) {
			return false, true
		}
		form, known, err := x86CompareExchangeProperties(fn.Instrs[ordinal+1].Op, fn.Instrs[ordinal+1], arch)
		return known && err == nil && form.destination.Kind == OpMem, true
	}
	form, known, err := x86CompareExchangeProperties(ins.Op, ins, arch)
	if !known {
		return false, false
	}
	if err != nil || arch != "386" {
		return false, true
	}
	width := int64(form.spec.bits / 8)
	inputs := []Operand{form.source, {Kind: OpReg, Reg: AX}}
	if form.spec.pair {
		width = 4
		inputs = []Operand{{Kind: OpReg, Reg: AX}, {Kind: OpReg, Reg: DX}, {Kind: OpReg, Reg: BX}, {Kind: OpReg, Reg: CX}}
	}
	for _, operand := range inputs {
		if !x86ResultRead(operand, width, state, nil, arch) {
			return false, true
		}
	}
	if form.destination.Kind == OpReg {
		if !x86ResultRead(form.destination, width, state, nil, arch) {
			return false, true
		}
	} else if !x86ResultIncomingPointerAddressIdentity(form.destination, state, arch) {
		return false, true // No SB, FP, AUTO, segment or unknown base is inferred.
	}
	// Accumulator writes are conditional: a successful dword CAS preserves
	// its prior upper half. Define only the actual low bytes, never all eight.
	outputs := []Reg{AX}
	if form.spec.pair {
		outputs = append(outputs, DX)
	} else if form.destination.Kind == OpReg {
		outputs = append(outputs, form.destination.Reg)
	}
	for _, reg := range outputs {
		root, mask, valid := x86ResultGP(reg, width, arch)
		if !valid {
			return false, true
		}
		state.gp[root] |= mask
		state.pointerGP[root] = false
	}
	if form.spec.pair {
		state.flags |= x86ByteFlagZF // Other status flags are unchanged by 8B.
	} else {
		state.flags = x86ByteFlagsAll
	}
	return true, true
}
