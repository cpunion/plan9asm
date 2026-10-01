package plan9asm

import (
	"encoding/binary"
	"fmt"
	"strings"
)

type x86RawNearReturnForm struct {
	nativeWidth bool
	cleanup     uint16
}

// C3 and C2 imm16 use the target's native return-PC width. Operand-size
// overrides and other prefixes require their own physical entry contract;
// successful decoding alone does not make them ordinary Go/C returns.
func x86RawNearReturnEncoding(code []byte) x86RawNearReturnForm {
	if len(code) == 1 && code[0] == 0xc3 {
		return x86RawNearReturnForm{nativeWidth: true}
	}
	if len(code) == 3 && code[0] == 0xc2 {
		return x86RawNearReturnForm{nativeWidth: true, cleanup: binary.LittleEndian.Uint16(code[1:])}
	}
	return x86RawNearReturnForm{}
}

func validateX86RawNearReturns(fn Func) error {
	hasReturn := false
	for _, ins := range fn.Instrs {
		form := ins.x86RawNearReturn
		if form == nil {
			continue
		}
		hasReturn = true
		if !form.nativeWidth || form.cleanup != 0 {
			return fmt.Errorf("%w: raw near RET needs a physical return-width and caller-stack cleanup contract: %q", ErrProbeNeedsContext, ins.Raw)
		}
	}
	if !hasReturn {
		return nil
	}
	// The raw instruction bypasses obj6's RET epilogue. A zero-sized,
	// stack-unobserving leaf has the same return-PC location as its caller.
	// All other cases need a native frame/continuation contract, not a guessed
	// epilogue or an inline RET inside LLVM's independently allocated frame.
	if fn.FrameSize != 0 || !x86RawReturnLeafStackUnobserved(fn) {
		return fmt.Errorf("%w: raw near RET requires a proved source stack/continuation; Go's TEXT frame epilogue is not encoded in raw bytes", ErrProbeNeedsContext)
	}
	return nil
}

func x86RawReturnLeafStackUnobserved(fn Func) bool {
	hasText := false
	for _, ins := range fn.Instrs {
		if ins.Op == OpTEXT {
			hasText = true
			_, rest := splitOpcode(ins.Raw)
			parts := strings.Split(rest, ",")
			if len(parts) == 3 {
				expression := globlFlagName.ReplaceAllStringFunc(parts[1], func(name string) string {
					if value, ok := goTextFlagValues[name]; ok {
						return value
					}
					return name
				})
				flags, known := parseImmExpr(strings.TrimSpace(expression))
				if !known || flags & ^uint64(2|4|512) != 0 {
					return false
				}
			} else if len(parts) != 2 {
				return false
			}
			continue
		}
		op := strings.ToUpper(string(ins.Op))
		if strings.HasPrefix(op, "PUSH") || strings.HasPrefix(op, "POP") ||
			strings.HasPrefix(op, "IRET") || strings.HasPrefix(op, "RETF") {
			return false
		}
		switch Op(op) {
		case OpCALL, "ADJSP", "ENTER", "ENTERL", "ENTERQ", "ENTERW", "LEAVEL", "LEAVEQ", "LEAVEW":
			return false
		case OpRET:
			if len(ins.Args) != 0 {
				return false
			}
		case OpJMP:
			if len(ins.Args) != 1 || (ins.Args[0].Kind != OpIdent && ins.Args[0].Kind != OpLabel) {
				return false
			}
		}
		for _, operand := range ins.Args {
			if operand.Kind == OpFPAddr || (operand.Kind == OpFP && operand.FPOffset < 0) {
				return false
			}
			if operand.Kind == OpReg && x86RawReturnStackRegister(operand.Reg) {
				return false
			}
			if operand.Kind == OpMem && x86RawReturnStackMemory(operand.Mem) {
				return false
			}
			// Go's effective-address immediate is OpSym, not OpMem. Use the
			// same parser and typed register roots as its actual lowering.
			if operand.Kind == OpSym {
				symbol := strings.TrimSpace(operand.Sym)
				if strings.HasPrefix(symbol, "$") {
					address, ok := parseMem(strings.TrimSpace(strings.TrimPrefix(symbol, "$")))
					if ok && x86RawReturnStackMemory(address) {
						return false
					}
				}
			}
		}
	}
	return hasText
}

func x86RawReturnStackRegister(reg Reg) bool {
	root, known := amd64FullRegBase(reg)
	return known && (root == SP || root == BP)
}

func x86RawReturnStackMemory(memory MemRef) bool {
	return x86RawReturnStackRegister(memory.Base) || x86RawReturnStackRegister(memory.Index)
}
