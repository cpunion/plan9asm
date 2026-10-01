package plan9asm

import "strings"

// Source RET is expanded by Go obj7 before encoding. Only raw AWORD RET is
// always one instruction. The source leaf/frame contract must be inspected
// before any raw AWORD BL has been rewritten into a semantic BL instruction.
func arm64SourceReturnWidth(fn Func, ins Instr) (int64, bool) {
	if ins.Op != OpRET || len(ins.Args) > 1 {
		return 0, false
	}
	if len(ins.Args) == 1 {
		target := ins.Args[0]
		switch target.Kind {
		case OpReg:
			if !isARM64GeneralOrZeroReg(target.Reg) && target.Reg != SP && target.Reg != "RSP" {
				return 0, false
			}
		case OpMem:
			if !isARM64GeneralOrZeroReg(target.Mem.Base) && target.Mem.Base != SP && target.Mem.Base != "RSP" {
				return 0, false
			}
		case OpSym:
			if !strings.HasSuffix(target.Sym, "(SB)") {
				return 0, false
			}
		default:
			return 0, false
		}
	}
	// Go's autosize is int32. Stay within its non-overflowing, aligned range;
	// invalid declarations cannot establish a raw instruction boundary.
	if fn.FrameSize < 0 && fn.FrameSize != -8 || fn.FrameSize > (1<<31)-25 || fn.FrameSize&7 != 0 {
		return 0, false
	}
	if !arm64SourceFrameFlagsKnown(fn) {
		return 0, false
	}
	frame := arm64SourceGoFrame(fn)
	if !frame.present {
		return 4, fn.FrameSize <= 0
	}
	autosize := fn.FrameSize + 8
	if autosize&15 == 8 {
		autosize += 8
	} else {
		autosize += 16
	}
	if !frame.restoresLink {
		// Leaf epilogue: ADD $(autosize-8),SP,FP; ADD $autosize,SP;
		// RET or symbol B. Large ADD constants may expand to several words.
		return arm64PositiveFrameAddWidth(autosize-8) + arm64PositiveFrameAddWidth(autosize) + 4, true
	}
	if autosize < 1<<12 {
		// MOVD FP + MOVD.P LR/SP for small frames; LDP FP/LR + ADD SP
		// above 0xf0. Both sequences are two words before the branch.
		return 12, true
	}
	// Large nonleaf: LDP FP/LR; MOVD $autosize,R27; ADD R27,SP;
	// RET or symbol B. Positive int32 constants require at most two MOV words.
	return 12 + arm64PositiveFrameMoveWidth(autosize), true
}

func arm64SourceFrameFlagsKnown(fn Func) bool {
	for _, ins := range fn.Instrs {
		if ins.Op != OpTEXT {
			continue
		}
		_, rest := splitOpcode(ins.Raw)
		parts := strings.Split(rest, ",")
		if len(parts) == 2 {
			return true
		}
		if len(parts) != 3 {
			return false
		}
		if _, resolved := parseImmExpr(strings.TrimSpace(parts[1])); resolved {
			return true
		}
		// Canonical textflag.h names have known NOFRAME membership. An
		// unresolved user flag may contain NOFRAME and cannot fix the width.
		for _, flag := range strings.Split(parts[1], "|") {
			switch strings.TrimSpace(flag) {
			case "NOPROF", "DUPOK", "NOSPLIT", "RODATA", "NOPTR", "WRAPPER", "NEEDCTXT", "TLSBSS", "NOFRAME", "REFLECTMETHOD", "TOPFRAME", "ABIWRAPPER":
			default:
				return false
			}
		}
		return true
	}
	return false
}

func arm64PositiveFrameMoveWidth(value int64) int64 {
	if value&0xffff == 0 || value>>16 == 0 || arm64LogicalBitmaskImmediate(uint64(value)) {
		return 4
	}
	return 8
}

func arm64PositiveFrameAddWidth(value int64) int64 {
	// asm7 C_ADDCON, C_MOVCON/C_BITCON, C_ADDCON2 and C_MOVCON2
	// are the complete positive int32 ADD constant classes used here.
	if value <= 0xfff || value&0xfff == 0 && value>>12 <= 0xfff {
		return 4
	}
	if arm64PositiveFrameMoveWidth(value) == 4 || value <= 0xffffff {
		return 8
	}
	return 12
}
