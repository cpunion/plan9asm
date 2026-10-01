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

func validateX86RawNearReturns(fn Func, file *File, opt Options, bindFrame bool) error {
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
	if !x86RawReturnMemoryBound(fn, file, opt, bindFrame) {
		return fmt.Errorf("%w: raw near RET needs bounded typed FP/static memory; ordinary pointers have no native return-PC nonalias contract", ErrProbeNeedsContext)
	}
	return nil
}

// This bounded contract does not infer noalias from a Go pointer type or a
// register value. It proves only scalar MOV accesses to bound FP slots or a
// same-file static object. Other memory forms need their own complete effects
// and access-range proof. In particular, this is not a general store whitelist.
func x86RawReturnMemoryBound(fn Func, file *File, opt Options, bindFrame bool) bool {
	name := fn.Sym
	if opt.ResolveSym != nil {
		name = opt.ResolveSym(name)
	}
	sig := opt.Sigs[name]
	for _, ins := range fn.Instrs {
		op := Op(normalizeInstructionOpcode(ins.Op))
		if _, _, _, memory := x86StringProperties(op); memory {
			return false
		}
		if _, _, memory := x86PortStringProperties(op); memory {
			return false
		}
		if _, memory := amd64ImplicitMaskMoveSpecs[op]; memory || op == "XLAT" {
			return false
		}
		for _, arg := range ins.Args {
			switch arg.Kind {
			case OpMem:
				// Even a static symbol plus an index is not a bounded object.
				// Address-only register expressions have no dereference, but
				// their later uses still have to satisfy this same contract.
				if op != "LEAW" && op != "LEAL" && op != "LEAQ" {
					return false
				}
			case OpFP:
				if !bindFrame {
					continue // Pure prepartition normalization retains metadata.
				}
				width := x86RawReturnScalarMoveBytes(op)
				if width == 0 || !x86RawReturnFPBound(arg.FPOffset, width, sig, opt.Goarch) {
					return false
				}
			case OpSym:
				symbol := strings.TrimSpace(arg.Sym)
				if strings.HasPrefix(symbol, "$") || ins.Op == OpTEXT || ins.Op == OpLABEL {
					continue
				}
				base, offset, static := parseSBRef(symbol)
				if !static {
					continue // Non-memory control operands retain their own checks.
				}
				if !x86RawReturnStaticBound(base, offset, x86RawReturnScalarMoveBytes(op), file) {
					return false
				}
			}
		}
	}
	return true
}

func x86RawReturnScalarMoveBytes(op Op) int64 {
	switch op {
	case "MOVB":
		return 1
	case "MOVW":
		return 2
	case "MOVL":
		return 4
	case "MOVQ":
		return 8
	default:
		return 0
	}
}

func x86RawReturnFPBound(offset, width int64, sig FuncSig, goarch string) bool {
	if offset < 0 || width <= 0 {
		return false
	}
	for group, slots := range [][]FrameSlot{sig.Frame.Params, sig.Frame.Results} {
		for _, slot := range slots {
			if slot.Offset != offset || slot.Index < 0 {
				continue
			}
			if group == 0 && slot.Index >= len(sig.Args) {
				continue
			}
			if group == 1 {
				if sig.Ret == Void || sig.Ret == "" {
					continue
				}
				fields, aggregate := parseLiteralStructFields(sig.Ret)
				if aggregate && slot.Index >= len(fields) || !aggregate && slot.Index != 0 {
					continue
				}
			}
			var size int64
			switch slot.Type {
			case I1, I8:
				size = 1
			case I16:
				size = 2
			case I32, LLVMType("float"):
				size = 4
			case I64, LLVMType("double"):
				size = 8
			case Ptr:
				size = 8
				if goarch == "386" {
					size = 4
				}
			}
			if width <= size && offset <= int64(^uint64(0)>>1)-size {
				return true
			}
		}
	}
	return false
}

func x86RawReturnStaticBound(symbol string, offset, width int64, file *File) bool {
	if file == nil || offset < 0 || width <= 0 {
		return false
	}
	for _, object := range file.Globl {
		if object.Sym == symbol && object.SizeRaw == "" && object.Size >= width && offset <= object.Size-width {
			return true
		}
	}
	return false
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
