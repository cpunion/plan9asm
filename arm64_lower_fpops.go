package plan9asm

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

func (c *arm64Ctx) lowerFP(op Op, ins Instr) (ok bool, terminated bool, err error) {
	if ok, terminated, err := c.lowerARM64ScalarFloatBinary(op, ins); ok {
		return ok, terminated, err
	}
	switch op {
	case "FMOVS", "FMOVD":
		return c.lowerScalarFloatMove(op, ins)

	case "FCVTSD", "FCVTDS":
		if strings.ToUpper(string(ins.Op)) != string(op) || len(ins.Args) != 2 ||
			ins.Args[0].Kind != OpReg || !isARM64FReg(ins.Args[0].Reg) ||
			ins.Args[1].Kind != OpReg || !isARM64FReg(ins.Args[1].Reg) {
			return true, false, fmt.Errorf("arm64 %s expects Fsrc, Fdst and no suffix: %q", op, ins.Raw)
		}
		sourceBits, destinationBits := 32, 64
		if op == "FCVTDS" {
			sourceBits, destinationBits = 64, 32
		}
		source, err := c.loadARM64ScalarFloatReg(ins.Args[0].Reg, sourceBits)
		if err != nil {
			return true, false, err
		}
		converted := c.newTmp()
		if sourceBits == 32 {
			fmt.Fprintf(c.b, "  %%%s = fpext float %s to double\n", converted, source)
		} else {
			fmt.Fprintf(c.b, "  %%%s = fptrunc double %s to float\n", converted, source)
		}
		return true, false, c.storeARM64ScalarFloatReg(ins.Args[1].Reg, destinationBits, "%"+converted)

	case "FCMPD":
		// FCMPD src, dst => compare dst ? src (same operand order convention as CMP/SUB).
		if len(ins.Args) != 2 {
			return true, false, fmt.Errorf("arm64 FCMPD expects 2 operands: %q", ins.Raw)
		}
		src, err := c.evalF64(ins.Args[0])
		if err != nil {
			return true, false, err
		}
		dst, err := c.evalF64(ins.Args[1])
		if err != nil {
			return true, false, err
		}
		eq := c.newTmp()
		lt := c.newTmp()
		gt := c.newTmp()
		uno := c.newTmp()
		fmt.Fprintf(c.b, "  %%%s = fcmp oeq double %s, %s\n", eq, dst, src)
		fmt.Fprintf(c.b, "  %%%s = fcmp olt double %s, %s\n", lt, dst, src)
		fmt.Fprintf(c.b, "  %%%s = fcmp ogt double %s, %s\n", gt, dst, src)
		fmt.Fprintf(c.b, "  %%%s = fcmp uno double %s, %s\n", uno, dst, src)
		// AArch64 FCMP flags model:
		// - unordered: N=0 Z=0 C=1 V=1
		// - lt: N=1 Z=0 C=0 V=0
		// - eq: N=0 Z=1 C=1 V=0
		// - gt: N=0 Z=0 C=1 V=0
		c01 := c.newTmp()
		cf := c.newTmp()
		fmt.Fprintf(c.b, "  %%%s = or i1 %%%s, %%%s\n", c01, gt, eq)
		fmt.Fprintf(c.b, "  %%%s = or i1 %%%s, %%%s\n", cf, c01, uno)
		fmt.Fprintf(c.b, "  store i1 %%%s, ptr %s\n", lt, c.flagsNSlot)
		fmt.Fprintf(c.b, "  store i1 %%%s, ptr %s\n", eq, c.flagsZSlot)
		fmt.Fprintf(c.b, "  store i1 %%%s, ptr %s\n", cf, c.flagsCSlot)
		fmt.Fprintf(c.b, "  store i1 %%%s, ptr %s\n", uno, c.flagsVSlot)
		c.flagsWritten = true
		return true, false, nil

	case "FCSELD":
		// FCSELD cond, a, b, dst
		if len(ins.Args) != 4 || ins.Args[0].Kind != OpIdent {
			return true, false, fmt.Errorf("arm64 FCSELD expects cond, a, b, dst: %q", ins.Raw)
		}
		a, err := c.evalF64(ins.Args[1])
		if err != nil {
			return true, false, err
		}
		bv, err := c.evalF64(ins.Args[2])
		if err != nil {
			return true, false, err
		}
		cv, err := c.condValue(ins.Args[0].Ident)
		if err != nil {
			return true, false, err
		}
		t := c.newTmp()
		fmt.Fprintf(c.b, "  %%%s = select i1 %s, double %s, double %s\n", t, cv, a, bv)
		return true, false, c.storeF64(ins.Args[3], "%"+t)

	case "FMADDS", "FMADDD", "FMSUBS", "FMSUBD",
		"FNMADDS", "FNMADDD", "FNMSUBS", "FNMSUBD":
		return c.lowerFusedMultiplyAdd(op, ins)

	case "FABSD":
		if len(ins.Args) != 2 {
			return true, false, fmt.Errorf("arm64 FABSD expects 2 operands: %q", ins.Raw)
		}
		src, err := c.evalF64(ins.Args[0])
		if err != nil {
			return true, false, err
		}
		t := c.newTmp()
		fmt.Fprintf(c.b, "  %%%s = call double @llvm.fabs.f64(double %s)\n", t, src)
		return true, false, c.storeF64(ins.Args[1], "%"+t)

	case "FSQRTS", "FSQRTD":
		if strings.ToUpper(string(ins.Op)) != string(op) || len(ins.Args) != 2 ||
			ins.Args[0].Kind != OpReg || ins.Args[1].Kind != OpReg {
			return true, false, fmt.Errorf("arm64 %s expects Fsrc, Fdst and no suffixes: %q", op, ins.Raw)
		}
		if _, ok := arm64ParseFReg(ins.Args[0].Reg); !ok {
			return true, false, fmt.Errorf("arm64 %s expects an F-register source: %q", op, ins.Raw)
		}
		if _, ok := arm64ParseFReg(ins.Args[1].Reg); !ok {
			return true, false, fmt.Errorf("arm64 %s expects an F-register destination: %q", op, ins.Raw)
		}
		if op == "FSQRTD" {
			src, err := c.evalF64(ins.Args[0])
			if err != nil {
				return true, false, err
			}
			result := c.newTmp()
			fmt.Fprintf(c.b, "  %%%s = call double @llvm.sqrt.f64(double %s)\n", result, src)
			return true, false, c.storeF64(ins.Args[1], "%"+result)
		}
		bits64, err := c.loadReg(ins.Args[0].Reg)
		if err != nil {
			return true, false, err
		}
		bits32 := c.newTmp()
		fmt.Fprintf(c.b, "  %%%s = trunc i64 %s to i32\n", bits32, bits64)
		src := c.newTmp()
		fmt.Fprintf(c.b, "  %%%s = bitcast i32 %%%s to float\n", src, bits32)
		result := c.newTmp()
		fmt.Fprintf(c.b, "  %%%s = call float @llvm.sqrt.f32(float %%%s)\n", result, src)
		resultBits32 := c.newTmp()
		fmt.Fprintf(c.b, "  %%%s = bitcast float %%%s to i32\n", resultBits32, result)
		resultBits64 := c.newTmp()
		fmt.Fprintf(c.b, "  %%%s = zext i32 %%%s to i64\n", resultBits64, resultBits32)
		return true, false, c.storeReg(ins.Args[1].Reg, "%"+resultBits64)

	case "FRINTZD", "FRINTMD", "FRINTPD":
		if len(ins.Args) != 2 {
			return true, false, fmt.Errorf("arm64 %s expects 2 operands: %q", op, ins.Raw)
		}
		src, err := c.evalF64(ins.Args[0])
		if err != nil {
			return true, false, err
		}
		t := c.newTmp()
		switch op {
		case "FRINTZD":
			fmt.Fprintf(c.b, "  %%%s = call double @llvm.trunc.f64(double %s)\n", t, src)
		case "FRINTMD":
			fmt.Fprintf(c.b, "  %%%s = call double @llvm.floor.f64(double %s)\n", t, src)
		case "FRINTPD":
			fmt.Fprintf(c.b, "  %%%s = call double @llvm.ceil.f64(double %s)\n", t, src)
		}
		return true, false, c.storeF64(ins.Args[1], "%"+t)

	case "FCVTZSD":
		// FCVTZSD src, dstReg
		if len(ins.Args) != 2 || ins.Args[1].Kind != OpReg {
			return true, false, fmt.Errorf("arm64 FCVTZSD expects src, dstReg: %q", ins.Raw)
		}
		src, err := c.evalF64(ins.Args[0])
		if err != nil {
			return true, false, err
		}
		t := c.newTmp()
		fmt.Fprintf(c.b, "  %%%s = fptosi double %s to i64\n", t, src)
		return true, false, c.storeReg(ins.Args[1].Reg, "%"+t)

	case "SCVTFD", "SCVTFS", "SCVTFWD", "SCVTFWS",
		"UCVTFD", "UCVTFS", "UCVTFWD", "UCVTFWS":
		if strings.ToUpper(string(ins.Op)) != string(op) || len(ins.Args) != 2 ||
			ins.Args[0].Kind != OpReg || !isARM64GeneralOrZeroReg(ins.Args[0].Reg) ||
			ins.Args[1].Kind != OpReg || !isARM64FReg(ins.Args[1].Reg) {
			return true, false, fmt.Errorf("arm64 %s expects one R/ZR source, one F destination, and no suffix: %q", op, ins.Raw)
		}
		src, err := c.loadReg(ins.Args[0].Reg)
		if err != nil {
			return true, false, err
		}
		integerType := "i64"
		if strings.Contains(string(op), "W") {
			narrow := c.newTmp()
			fmt.Fprintf(c.b, "  %%%s = trunc i64 %s to i32\n", narrow, src)
			src = "%" + narrow
			integerType = "i32"
		}
		floatType := "double"
		bits := 64
		if strings.HasSuffix(string(op), "S") {
			floatType = "float"
			bits = 32
		}
		conversion := "sitofp"
		if strings.HasPrefix(string(op), "U") {
			conversion = "uitofp"
		}
		converted := c.newTmp()
		fmt.Fprintf(c.b, "  %%%s = %s %s %s to %s\n", converted, conversion, integerType, src, floatType)
		if bits == 64 {
			encoded := c.newTmp()
			fmt.Fprintf(c.b, "  %%%s = bitcast double %%%s to i64\n", encoded, converted)
			return true, false, c.storeReg(ins.Args[1].Reg, "%"+encoded)
		}
		encoded := c.newTmp()
		wide := c.newTmp()
		fmt.Fprintf(c.b, "  %%%s = bitcast float %%%s to i32\n", encoded, converted)
		fmt.Fprintf(c.b, "  %%%s = zext i32 %%%s to i64\n", wide, encoded)
		return true, false, c.storeReg(ins.Args[1].Reg, "%"+wide)
	}
	return false, false, nil
}

func (c *arm64Ctx) lowerFusedMultiplyAdd(op Op, ins Instr) (ok bool, terminated bool, err error) {
	if strings.ToUpper(string(ins.Op)) != string(op) || len(ins.Args) != 4 {
		return true, false, fmt.Errorf("arm64 %s expects four F-register operands and no suffixes: %q", op, ins.Raw)
	}
	for _, arg := range ins.Args {
		if arg.Kind != OpReg {
			return true, false, fmt.Errorf("arm64 %s expects four F-register operands: %q", op, ins.Raw)
		}
		if _, ok := arm64ParseFReg(arg.Reg); !ok {
			return true, false, fmt.Errorf("arm64 %s expects four F-register operands: %q", op, ins.Raw)
		}
	}

	floatType := "double"
	intrinsic := "@llvm.fma.f64"
	eval := c.evalF64
	store := c.storeF64
	if strings.HasSuffix(string(op), "S") {
		floatType = "float"
		intrinsic = "@llvm.fma.f32"
		eval = c.evalF32
		store = c.storeF32
	}

	// Go's Plan 9 order is Fm, Fa, Fn, Fd. The architectural operations are:
	//   FMADD:   Fn*Fm + Fa       FMSUB:  Fa - Fn*Fm
	//   FNMADD: -(Fn*Fm + Fa)     FNMSUB: Fn*Fm - Fa
	fm, err := eval(ins.Args[0])
	if err != nil {
		return true, false, err
	}
	fa, err := eval(ins.Args[1])
	if err != nil {
		return true, false, err
	}
	fn, err := eval(ins.Args[2])
	if err != nil {
		return true, false, err
	}

	if strings.HasPrefix(string(op), "FMSUB") {
		negFM := c.newTmp()
		fmt.Fprintf(c.b, "  %%%s = fneg %s %s\n", negFM, floatType, fm)
		fm = "%" + negFM
	} else if strings.HasPrefix(string(op), "FNMSUB") {
		negFA := c.newTmp()
		fmt.Fprintf(c.b, "  %%%s = fneg %s %s\n", negFA, floatType, fa)
		fa = "%" + negFA
	}

	result := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = call %s %s(%s %s, %s %s, %s %s)\n",
		result, floatType, intrinsic, floatType, fm, floatType, fn, floatType, fa)
	value := "%" + result
	if strings.HasPrefix(string(op), "FNMADD") {
		negResult := c.newTmp()
		fmt.Fprintf(c.b, "  %%%s = fneg %s %s\n", negResult, floatType, value)
		value = "%" + negResult
	}
	return true, false, store(ins.Args[3], value)
}

func (c *arm64Ctx) evalFMOVDBits(op Operand) (string, error) {
	if op.Kind == OpSym && strings.HasPrefix(op.Sym, "$") {
		if fv, ok := arm64ParseDollarFloat(op.Sym); ok {
			u := math.Float64bits(fv)
			return strconv.FormatInt(int64(u), 10), nil
		}
		if iv, ok := arm64ParseDollarInt64(op.Sym); ok {
			return strconv.FormatInt(iv, 10), nil
		}
	}
	return c.eval64(op, false)
}

func (c *arm64Ctx) evalF64(op Operand) (string, error) {
	switch op.Kind {
	case OpReg:
		v64, err := c.loadReg(op.Reg)
		if err != nil {
			return "", err
		}
		t := c.newTmp()
		fmt.Fprintf(c.b, "  %%%s = bitcast i64 %s to double\n", t, v64)
		return "%" + t, nil
	case OpImm:
		return formatLLVMFloat64Literal(float64(op.Imm)), nil
	case OpSym:
		if strings.HasPrefix(op.Sym, "$") {
			if fv, ok := arm64ParseDollarFloat(op.Sym); ok {
				return formatLLVMFloat64Literal(fv), nil
			}
			if iv, ok := arm64ParseDollarInt64(op.Sym); ok {
				return formatLLVMFloat64Literal(float64(iv)), nil
			}
		}
		return "", fmt.Errorf("arm64: unsupported f64 immediate %q", op.String())
	case OpFP:
		slot, ok := c.fpParams[op.FPOffset]
		if !ok {
			return "", fmt.Errorf("arm64: unsupported FP param slot: %s", op.String())
		}
		idx := slot.Index
		if idx < 0 || idx >= len(c.sig.Args) {
			return "", fmt.Errorf("arm64: FP slot %s invalid arg index %d", op.String(), idx)
		}
		arg := fmt.Sprintf("%%arg%d", idx)
		if fields := frameSlotFields(slot); len(fields) != 0 {
			aggTy := c.sig.Args[idx]
			t := c.newTmp()
			fmt.Fprintf(c.b, "  %%%s = extractvalue %s %s%s\n", t, aggTy, arg, frameSlotExtractSuffix(slot))
			arg = "%" + t
		}
		switch slot.Type {
		case LLVMType("double"):
			return arg, nil
		case LLVMType("float"):
			t := c.newTmp()
			fmt.Fprintf(c.b, "  %%%s = fpext float %s to double\n", t, arg)
			return "%" + t, nil
		case I64:
			t := c.newTmp()
			fmt.Fprintf(c.b, "  %%%s = bitcast i64 %s to double\n", t, arg)
			return "%" + t, nil
		case I32, I16, I8, I1:
			t := c.newTmp()
			fmt.Fprintf(c.b, "  %%%s = zext %s %s to i64\n", t, slot.Type, arg)
			b := c.newTmp()
			fmt.Fprintf(c.b, "  %%%s = bitcast i64 %%%s to double\n", b, t)
			return "%" + b, nil
		case Ptr:
			t := c.newTmp()
			fmt.Fprintf(c.b, "  %%%s = ptrtoint ptr %s to i64\n", t, arg)
			b := c.newTmp()
			fmt.Fprintf(c.b, "  %%%s = bitcast i64 %%%s to double\n", b, t)
			return "%" + b, nil
		default:
			return "", fmt.Errorf("arm64: unsupported FP slot type %s", slot.Type)
		}
	default:
		return "", fmt.Errorf("arm64: unsupported f64 operand %s", op.String())
	}
}

func (c *arm64Ctx) evalF32(op Operand) (string, error) {
	if op.Kind != OpReg {
		return "", fmt.Errorf("arm64: unsupported f32 operand %s", op.String())
	}
	v64, err := c.loadReg(op.Reg)
	if err != nil {
		return "", err
	}
	v32 := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = trunc i64 %s to i32\n", v32, v64)
	value := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = bitcast i32 %%%s to float\n", value, v32)
	return "%" + value, nil
}

func (c *arm64Ctx) storeF64(dst Operand, v string) error {
	switch dst.Kind {
	case OpReg:
		b := c.newTmp()
		fmt.Fprintf(c.b, "  %%%s = bitcast double %s to i64\n", b, v)
		return c.storeReg(dst.Reg, "%"+b)
	case OpFP:
		b := c.newTmp()
		fmt.Fprintf(c.b, "  %%%s = bitcast double %s to i64\n", b, v)
		return c.storeFPResult64(dst.FPOffset, "%"+b)
	default:
		return fmt.Errorf("arm64: unsupported f64 dst operand %s", dst.String())
	}
}

func (c *arm64Ctx) storeF32(dst Operand, v string) error {
	if dst.Kind != OpReg {
		return fmt.Errorf("arm64: unsupported f32 dst operand %s", dst.String())
	}
	bits32 := c.newTmp()
	bits64 := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = bitcast float %s to i32\n", bits32, v)
	fmt.Fprintf(c.b, "  %%%s = zext i32 %%%s to i64\n", bits64, bits32)
	return c.storeReg(dst.Reg, "%"+bits64)
}

func arm64ParseDollarFloat(sym string) (float64, bool) {
	if !strings.HasPrefix(sym, "$") {
		return 0, false
	}
	s := strings.TrimSpace(strings.TrimPrefix(sym, "$"))
	if s == "" {
		return 0, false
	}
	// Integer-looking immediates are handled by arm64ParseDollarInt64.
	if strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X") {
		return 0, false
	}
	// Heuristic: require decimal/exponent marker.
	if !strings.ContainsAny(s, ".eE") {
		return 0, false
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, false
	}
	return v, true
}

func arm64ParseDollarInt64(sym string) (int64, bool) {
	if !strings.HasPrefix(sym, "$") {
		return 0, false
	}
	s := strings.TrimSpace(strings.TrimPrefix(sym, "$"))
	if s == "" {
		return 0, false
	}
	if v, err := strconv.ParseInt(s, 0, 64); err == nil {
		return v, true
	}
	if uv, err := strconv.ParseUint(s, 0, 64); err == nil {
		return int64(uv), true
	}
	return 0, false
}
