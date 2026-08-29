package plan9asm

import "fmt"

func (c *amd64Ctx) lowerX87(op Op, ins Instr) (ok bool, terminated bool, err error) {
	switch op {
	case "FMOVD", "FMOVDP", "FMOVV", "FMOVVP", "FXCHD", "FADDDP",
		"FDIVD", "FMULD", "FMULDP", "FSTCW", "FLDCW", "FRNDINT",
		"FABS", "FUCOMI", "FTST", "FSTSW", "FLD1", "FSQRT":
	default:
		return false, false, nil
	}
	if c.goarch != "386" {
		return true, false, fmt.Errorf("amd64: x87 instruction %s requires GOARCH=386", op)
	}

	switch op {
	case "FMOVD":
		if len(ins.Args) != 2 {
			return true, false, fmt.Errorf("386 FMOVD expects src, dst: %q", ins.Raw)
		}
		v, err := c.evalX87F64(ins.Args[0])
		if err != nil {
			return true, false, err
		}
		if idx, isF := x87OperandReg(ins.Args[1]); isF && idx == 0 && ins.Args[0].Kind != OpReg {
			return true, false, c.pushX87(v)
		}
		return true, false, c.storeX87F64(ins.Args[1], v)

	case "FMOVDP":
		if len(ins.Args) != 2 {
			return true, false, fmt.Errorf("386 FMOVDP expects src, dst: %q", ins.Raw)
		}
		v, err := c.evalX87F64(ins.Args[0])
		if err != nil {
			return true, false, err
		}
		if err := c.storeX87F64(ins.Args[1], v); err != nil {
			return true, false, err
		}
		c.popX87()
		return true, false, nil

	case "FMOVV":
		if len(ins.Args) != 2 {
			return true, false, fmt.Errorf("386 FMOVV expects src, dst: %q", ins.Raw)
		}
		if idx, ok := x87OperandReg(ins.Args[1]); !ok || idx != 0 {
			return true, false, fmt.Errorf("386 FMOVV expects F0 destination: %q", ins.Raw)
		}
		v, err := c.evalIntSized(ins.Args[0], I64)
		if err != nil {
			return true, false, err
		}
		f := c.newTmp()
		fmt.Fprintf(c.b, "  %%%s = sitofp i64 %s to double\n", f, v)
		return true, false, c.pushX87("%" + f)

	case "FMOVVP":
		if len(ins.Args) != 2 {
			return true, false, fmt.Errorf("386 FMOVVP expects src, dst: %q", ins.Raw)
		}
		v, err := c.evalX87F64(ins.Args[0])
		if err != nil {
			return true, false, err
		}
		rounded := c.roundX87(v)
		iv := c.newTmp()
		fmt.Fprintf(c.b, "  %%%s = fptosi double %s to i64\n", iv, rounded)
		if err := c.storeX87I64(ins.Args[1], "%"+iv); err != nil {
			return true, false, err
		}
		c.popX87()
		return true, false, nil

	case "FXCHD":
		if len(ins.Args) != 2 {
			return true, false, fmt.Errorf("386 FXCHD expects Fsrc, Fdst: %q", ins.Raw)
		}
		a, okA := x87OperandReg(ins.Args[0])
		bv, okB := x87OperandReg(ins.Args[1])
		if !okA || !okB {
			return true, false, fmt.Errorf("386 FXCHD expects x87 registers: %q", ins.Raw)
		}
		av := c.loadX87(a)
		b := c.loadX87(bv)
		c.storeX87(a, b)
		c.storeX87(bv, av)
		return true, false, nil

	case "FDIVD", "FMULD":
		if len(ins.Args) != 2 {
			return true, false, fmt.Errorf("386 %s expects src, dst: %q", op, ins.Raw)
		}
		src, err := c.evalX87F64(ins.Args[0])
		if err != nil {
			return true, false, err
		}
		dst, err := c.evalX87F64(ins.Args[1])
		if err != nil {
			return true, false, err
		}
		out := c.newTmp()
		if op == "FDIVD" {
			fmt.Fprintf(c.b, "  %%%s = fdiv double %s, %s\n", out, dst, src)
		} else {
			fmt.Fprintf(c.b, "  %%%s = fmul double %s, %s\n", out, dst, src)
		}
		return true, false, c.storeX87F64(ins.Args[1], "%"+out)

	case "FADDDP", "FMULDP":
		if len(ins.Args) != 2 {
			return true, false, fmt.Errorf("386 %s expects src, dst: %q", op, ins.Raw)
		}
		src, err := c.evalX87F64(ins.Args[0])
		if err != nil {
			return true, false, err
		}
		dst, err := c.evalX87F64(ins.Args[1])
		if err != nil {
			return true, false, err
		}
		out := c.newTmp()
		if op == "FADDDP" {
			fmt.Fprintf(c.b, "  %%%s = fadd double %s, %s\n", out, dst, src)
		} else {
			fmt.Fprintf(c.b, "  %%%s = fmul double %s, %s\n", out, dst, src)
		}
		if err := c.storeX87F64(ins.Args[1], "%"+out); err != nil {
			return true, false, err
		}
		c.popX87()
		return true, false, nil

	case "FSTCW":
		if len(ins.Args) != 1 {
			return true, false, fmt.Errorf("386 FSTCW expects dst: %q", ins.Raw)
		}
		cw := c.newTmp()
		fmt.Fprintf(c.b, "  %%%s = load i16, ptr %s\n", cw, c.x87ControlSlot)
		return true, false, c.storeX87I16(ins.Args[0], "%"+cw)

	case "FLDCW":
		if len(ins.Args) != 1 {
			return true, false, fmt.Errorf("386 FLDCW expects src: %q", ins.Raw)
		}
		cw, err := c.evalIntSized(ins.Args[0], I16)
		if err != nil {
			return true, false, err
		}
		fmt.Fprintf(c.b, "  store i16 %s, ptr %s\n", cw, c.x87ControlSlot)
		return true, false, nil

	case "FRNDINT":
		if len(ins.Args) != 0 {
			return true, false, fmt.Errorf("386 FRNDINT takes no operands: %q", ins.Raw)
		}
		return true, false, c.storeX87F64(Operand{Kind: OpReg, Reg: Reg("F0")}, c.roundX87(c.loadX87(0)))

	case "FABS", "FSQRT":
		if len(ins.Args) != 0 {
			return true, false, fmt.Errorf("386 %s takes no operands: %q", op, ins.Raw)
		}
		v := c.loadX87(0)
		out := c.newTmp()
		intrinsic := "llvm.fabs.f64"
		if op == "FSQRT" {
			intrinsic = "llvm.sqrt.f64"
		}
		fmt.Fprintf(c.b, "  %%%s = call double @%s(double %s)\n", out, intrinsic, v)
		c.storeX87(0, "%"+out)
		return true, false, nil

	case "FUCOMI":
		if len(ins.Args) != 2 {
			return true, false, fmt.Errorf("386 FUCOMI expects lhs, rhs: %q", ins.Raw)
		}
		lhs, err := c.evalX87F64(ins.Args[0])
		if err != nil {
			return true, false, err
		}
		rhs, err := c.evalX87F64(ins.Args[1])
		if err != nil {
			return true, false, err
		}
		z := c.newTmp()
		cf := c.newTmp()
		slt := c.newTmp()
		fmt.Fprintf(c.b, "  %%%s = fcmp ueq double %s, %s\n", z, lhs, rhs)
		fmt.Fprintf(c.b, "  %%%s = fcmp ult double %s, %s\n", cf, lhs, rhs)
		fmt.Fprintf(c.b, "  %%%s = fcmp olt double %s, %s\n", slt, lhs, rhs)
		fmt.Fprintf(c.b, "  store i1 %%%s, ptr %s\n", z, c.flagsZSlot)
		fmt.Fprintf(c.b, "  store i1 %%%s, ptr %s\n", cf, c.flagsCFSlot)
		fmt.Fprintf(c.b, "  store i1 %%%s, ptr %s\n", slt, c.flagsSltSlot)
		return true, false, nil

	case "FTST":
		if len(ins.Args) != 0 {
			return true, false, fmt.Errorf("386 FTST takes no operands: %q", ins.Raw)
		}
		value := c.loadX87(0)
		equal := c.newTmp()
		less := c.newTmp()
		unordered := c.newTmp()
		fmt.Fprintf(c.b, "  %%%s = fcmp oeq double %s, 0.000000e+00\n", equal, value)
		fmt.Fprintf(c.b, "  %%%s = fcmp olt double %s, 0.000000e+00\n", less, value)
		fmt.Fprintf(c.b, "  %%%s = fcmp uno double %s, 0.000000e+00\n", unordered, value)
		c3 := c.newTmp()
		c0 := c.newTmp()
		fmt.Fprintf(c.b, "  %%%s = or i1 %%%s, %%%s\n", c3, equal, unordered)
		fmt.Fprintf(c.b, "  %%%s = or i1 %%%s, %%%s\n", c0, less, unordered)
		c3word := c.newTmp()
		c2word := c.newTmp()
		c0word := c.newTmp()
		fmt.Fprintf(c.b, "  %%%s = select i1 %%%s, i16 16384, i16 0\n", c3word, c3)
		fmt.Fprintf(c.b, "  %%%s = select i1 %%%s, i16 1024, i16 0\n", c2word, unordered)
		fmt.Fprintf(c.b, "  %%%s = select i1 %%%s, i16 256, i16 0\n", c0word, c0)
		status01 := c.newTmp()
		status := c.newTmp()
		fmt.Fprintf(c.b, "  %%%s = or i16 %%%s, %%%s\n", status01, c3word, c2word)
		fmt.Fprintf(c.b, "  %%%s = or i16 %%%s, %%%s\n", status, status01, c0word)
		fmt.Fprintf(c.b, "  store i16 %%%s, ptr %s\n", status, c.x87StatusSlot)
		return true, false, nil

	case "FSTSW":
		if len(ins.Args) != 1 {
			return true, false, fmt.Errorf("386 FSTSW expects dst: %q", ins.Raw)
		}
		status := c.newTmp()
		fmt.Fprintf(c.b, "  %%%s = load i16, ptr %s\n", status, c.x87StatusSlot)
		return true, false, c.storeX87I16(ins.Args[0], "%"+status)

	case "FLD1":
		if len(ins.Args) != 0 {
			return true, false, fmt.Errorf("386 FLD1 takes no operands: %q", ins.Raw)
		}
		return true, false, c.pushX87("1.000000e+00")
	}
	return false, false, nil
}

func x87OperandReg(op Operand) (int, bool) {
	if op.Kind != OpReg {
		return 0, false
	}
	return amd64ParseX87Reg(op.Reg)
}

func (c *amd64Ctx) loadX87(index int) string {
	t := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = load double, ptr %s\n", t, c.x87Slot[index])
	return "%" + t
}

func (c *amd64Ctx) storeX87(index int, value string) {
	fmt.Fprintf(c.b, "  store double %s, ptr %s\n", value, c.x87Slot[index])
}

func (c *amd64Ctx) pushX87(value string) error {
	for i := len(c.x87Slot) - 1; i > 0; i-- {
		c.storeX87(i, c.loadX87(i-1))
	}
	c.storeX87(0, value)
	return nil
}

func (c *amd64Ctx) popX87() {
	for i := 0; i < len(c.x87Slot)-1; i++ {
		c.storeX87(i, c.loadX87(i+1))
	}
	c.storeX87(len(c.x87Slot)-1, "0.000000e+00")
}

func (c *amd64Ctx) evalX87F64(op Operand) (string, error) {
	if index, ok := x87OperandReg(op); ok {
		return c.loadX87(index), nil
	}
	return c.evalF64(op)
}

func (c *amd64Ctx) storeX87F64(dst Operand, value string) error {
	if index, ok := x87OperandReg(dst); ok {
		c.storeX87(index, value)
		return nil
	}
	switch dst.Kind {
	case OpFP:
		return c.storeFPResult(dst.FPOffset, LLVMType("double"), value)
	case OpMem:
		p, ptrType, err := c.ptrFromMem(dst.Mem)
		if err != nil {
			return err
		}
		fmt.Fprintf(c.b, "  store double %s, %s %s, align 1\n", value, ptrType, p)
		return nil
	case OpSym:
		p, err := c.ptrFromSB(dst.Sym)
		if err != nil {
			return err
		}
		fmt.Fprintf(c.b, "  store double %s, ptr %s, align 1\n", value, p)
		return nil
	default:
		return fmt.Errorf("386: unsupported x87 double destination %s", dst.String())
	}
}

func (c *amd64Ctx) storeX87I16(dst Operand, value string) error {
	switch dst.Kind {
	case OpReg:
		return c.storeRegSized(dst.Reg, I16, value)
	case OpMem:
		p, ptrType, err := c.ptrFromMem(dst.Mem)
		if err != nil {
			return err
		}
		fmt.Fprintf(c.b, "  store i16 %s, %s %s, align 1\n", value, ptrType, p)
		return nil
	case OpSym:
		p, err := c.ptrFromSB(dst.Sym)
		if err != nil {
			return err
		}
		fmt.Fprintf(c.b, "  store i16 %s, ptr %s, align 1\n", value, p)
		return nil
	default:
		return fmt.Errorf("386: unsupported x87 word destination %s", dst.String())
	}
}

func (c *amd64Ctx) storeX87I64(dst Operand, value string) error {
	switch dst.Kind {
	case OpFP:
		return c.storeFPResult(dst.FPOffset, I64, value)
	case OpMem:
		p, ptrType, err := c.ptrFromMem(dst.Mem)
		if err != nil {
			return err
		}
		fmt.Fprintf(c.b, "  store i64 %s, %s %s, align 1\n", value, ptrType, p)
		return nil
	case OpSym:
		p, err := c.ptrFromSB(dst.Sym)
		if err != nil {
			return err
		}
		fmt.Fprintf(c.b, "  store i64 %s, ptr %s, align 1\n", value, p)
		return nil
	default:
		return fmt.Errorf("386: unsupported x87 integer destination %s", dst.String())
	}
}

func (c *amd64Ctx) roundX87(value string) string {
	cw := c.newTmp()
	mode := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = load i16, ptr %s\n", cw, c.x87ControlSlot)
	fmt.Fprintf(c.b, "  %%%s = and i16 %%%s, 3072\n", mode, cw)
	down := c.newTmp()
	up := c.newTmp()
	zero := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = call double @llvm.floor.f64(double %s)\n", down, value)
	fmt.Fprintf(c.b, "  %%%s = call double @llvm.ceil.f64(double %s)\n", up, value)
	fmt.Fprintf(c.b, "  %%%s = call double @llvm.trunc.f64(double %s)\n", zero, value)
	nearest := c.roundNearestEvenX87(value, "%"+down)
	isDown := c.newTmp()
	isUp := c.newTmp()
	isZero := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = icmp eq i16 %%%s, 1024\n", isDown, mode)
	fmt.Fprintf(c.b, "  %%%s = icmp eq i16 %%%s, 2048\n", isUp, mode)
	fmt.Fprintf(c.b, "  %%%s = icmp eq i16 %%%s, 3072\n", isZero, mode)
	selectDown := c.newTmp()
	selectUp := c.newTmp()
	result := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = select i1 %%%s, double %%%s, double %s\n", selectDown, isDown, down, nearest)
	fmt.Fprintf(c.b, "  %%%s = select i1 %%%s, double %%%s, double %%%s\n", selectUp, isUp, up, selectDown)
	fmt.Fprintf(c.b, "  %%%s = select i1 %%%s, double %%%s, double %%%s\n", result, isZero, zero, selectUp)
	return "%" + result
}

func (c *amd64Ctx) roundNearestEvenX87(value, floor string) string {
	// Values at or above 2^52 are already integral in binary64. Treat unordered
	// values the same way so NaNs pass through without feeding fptosi.
	abs := c.newTmp()
	large := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = call double @llvm.fabs.f64(double %s)\n", abs, value)
	fmt.Fprintf(c.b, "  %%%s = fcmp uge double %%%s, 4.503599627370496e+15\n", large, abs)
	fraction := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = fsub double %s, %s\n", fraction, value, floor)
	aboveHalf := c.newTmp()
	tie := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = fcmp ogt double %%%s, 5.000000e-01\n", aboveHalf, fraction)
	fmt.Fprintf(c.b, "  %%%s = fcmp oeq double %%%s, 5.000000e-01\n", tie, fraction)
	floorInt := c.newTmp()
	oddBits := c.newTmp()
	odd := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = fptosi double %s to i64\n", floorInt, floor)
	fmt.Fprintf(c.b, "  %%%s = and i64 %%%s, 1\n", oddBits, floorInt)
	fmt.Fprintf(c.b, "  %%%s = icmp ne i64 %%%s, 0\n", odd, oddBits)
	tieUp := c.newTmp()
	roundUp := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = and i1 %%%s, %%%s\n", tieUp, tie, odd)
	fmt.Fprintf(c.b, "  %%%s = or i1 %%%s, %%%s\n", roundUp, aboveHalf, tieUp)
	ceil := c.newTmp()
	rounded := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = fadd double %s, 1.000000e+00\n", ceil, floor)
	fmt.Fprintf(c.b, "  %%%s = select i1 %%%s, double %%%s, double %s\n", rounded, roundUp, ceil, floor)
	// Arithmetic can turn a negative result rounded to zero into +0. Preserve
	// the source sign to match x87 FRNDINT.
	valueBits := c.newTmp()
	signedZeroBits := c.newTmp()
	signedZero := c.newTmp()
	isZero := c.newTmp()
	withSignedZero := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = bitcast double %s to i64\n", valueBits, value)
	fmt.Fprintf(c.b, "  %%%s = and i64 %%%s, -9223372036854775808\n", signedZeroBits, valueBits)
	fmt.Fprintf(c.b, "  %%%s = bitcast i64 %%%s to double\n", signedZero, signedZeroBits)
	fmt.Fprintf(c.b, "  %%%s = fcmp oeq double %%%s, 0.000000e+00\n", isZero, rounded)
	fmt.Fprintf(c.b, "  %%%s = select i1 %%%s, double %%%s, double %%%s\n", withSignedZero, isZero, signedZero, rounded)
	result := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = select i1 %%%s, double %s, double %%%s\n", result, large, value, withSignedZero)
	return "%" + result
}
