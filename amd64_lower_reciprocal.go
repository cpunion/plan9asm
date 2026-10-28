package plan9asm

import (
	"fmt"
	"strings"
)

type amd64Reciprocal14Spec struct {
	laneBits int
	scalar   bool
}

// amd64Reciprocal14Specs covers every opcode backed by Go 1.27's
// _yvexpandpd and _yvgetexpsd reciprocal-estimate tables.
var amd64Reciprocal14Specs = map[Op]amd64Reciprocal14Spec{
	"VRCP14PS":   {laneBits: 32},
	"VRCP14PD":   {laneBits: 64},
	"VRCP14SS":   {laneBits: 32, scalar: true},
	"VRCP14SD":   {laneBits: 64, scalar: true},
	"VRSQRT14PS": {laneBits: 32},
	"VRSQRT14PD": {laneBits: 64},
	"VRSQRT14SS": {laneBits: 32, scalar: true},
	"VRSQRT14SD": {laneBits: 64, scalar: true},
}

// The complete Go RCP14/RSQRT14 family uses the existing typed native helper.
// Its approximation observes dynamic DAZ/FTZ, and active native memory reads
// must remain observable even when the numerical result is dead.
func (c *amd64Ctx) lowerReciprocal14(op Op, ins Instr) (bool, bool, error) {
	base, _, _ := strings.Cut(strings.ToUpper(string(op)), ".")
	if _, known := amd64Reciprocal14Specs[Op(base)]; !known {
		return false, false, nil
	}
	return c.lowerNativeFloatingHelper(op, ins)
}

// lowerReciprocalEstimate implements the complete Go 1.27 legacy and VEX
// RCP/RSQRT estimate family: packed RCPPS/RSQRTPS/VRCPPS/VRSQRTPS plus scalar
// RCPSS/RSQRTSS/VRCPSS/VRSQRTSS. Packed estimates retain their target
// intrinsics; scalar estimates preserve their complete source opcode,
// approximation, native m32 access and independent copied upper lanes.
func (c *amd64Ctx) lowerReciprocalEstimate(rawOp Op, ins Instr) (ok bool, terminated bool, err error) {
	raw := strings.ToUpper(string(rawOp))
	base := raw
	if dot := strings.IndexByte(raw, '.'); dot >= 0 {
		base = raw[:dot]
	}
	switch base {
	case "RCPPS", "RSQRTPS", "VRCPPS", "VRSQRTPS",
		"RCPSS", "RSQRTSS", "VRCPSS", "VRSQRTSS":
	default:
		return false, false, nil
	}
	if raw != base {
		return true, false, fmt.Errorf("%s %s has no suffixed form in Go 1.27: %q", c.goarch, base, ins.Raw)
	}
	if strings.HasSuffix(base, "SS") {
		return c.lowerScalarReciprocalEstimate(base, ins)
	}
	if len(ins.Args) != 2 {
		return true, false, fmt.Errorf("%s %s expects vector-or-memory source and vector destination: %q", c.goarch, base, ins.Raw)
	}

	source, destination := ins.Args[0], ins.Args[1]
	byteWidth := amd64VectorByteWidth(destination.Reg)
	legacy := base == "RCPPS" || base == "RSQRTPS"
	if legacy {
		if destination.Kind != OpReg || byteWidth != 16 || !c.isGoLegacyXReg(destination.Reg) {
			return true, false, fmt.Errorf("%s %s destination is outside Go 1.27's legacy X register class: %q", c.goarch, base, ins.Raw)
		}
		if source.Kind == OpReg && !c.isGoLegacyXReg(source.Reg) {
			return true, false, fmt.Errorf("%s %s source must be an in-range X register or memory: %q", c.goarch, base, ins.Raw)
		}
	} else {
		if destination.Kind != OpReg || byteWidth != 16 && byteWidth != 32 || !amd64VEXVectorRegister(destination, byteWidth) {
			return true, false, fmt.Errorf("%s %s destination is outside Go 1.27's VEX X/Y register classes: %q", c.goarch, base, ins.Raw)
		}
		if source.Kind == OpReg && !amd64VEXVectorRegister(source, byteWidth) {
			return true, false, fmt.Errorf("%s %s source must match the destination width: %q", c.goarch, base, ins.Raw)
		}
	}
	if source.Kind != OpReg && !isAMD64MemoryOperand(source) {
		return true, false, fmt.Errorf("%s %s source must be a matching vector register or memory: %q", c.goarch, base, ins.Raw)
	}

	lanes := byteWidth / 4
	value, err := c.loadPackedReciprocalSource(source, byteWidth, lanes)
	if err != nil {
		return true, false, err
	}
	intrinsic := "llvm.x86.sse.rcp.ps"
	if strings.Contains(base, "RSQRT") {
		intrinsic = "llvm.x86.sse.rsqrt.ps"
	}
	if byteWidth == 32 {
		intrinsic = "llvm.x86.avx.rcp.ps.256"
		if strings.Contains(base, "RSQRT") {
			intrinsic = "llvm.x86.avx.rsqrt.ps.256"
		}
	}
	vectorType := fmt.Sprintf("<%d x float>", lanes)
	result := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = call %s @%s(%s %s)\n", result, vectorType, intrinsic, vectorType, value)
	bytesValue := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = bitcast %s %%%s to <%d x i8>\n", bytesValue, vectorType, result, byteWidth)
	if legacy {
		return true, false, c.storeX(destination.Reg, "%"+bytesValue)
	}
	return true, false, c.storePackedMoveOperand(destination, byteWidth, "%"+bytesValue)
}

func (c *amd64Ctx) lowerScalarReciprocalEstimate(base string, ins Instr) (bool, bool, error) {
	form, _, err := x86ScalarReciprocalProperties(Op(base), ins, c.goarch)
	if err != nil {
		return true, false, err
	}
	return true, false, c.emitScalarReciprocalEstimate(ins, form)
}

func (c *amd64Ctx) emitScalarReciprocalEstimate(ins Instr, form x86ScalarReciprocalForm) error {
	if ins.x86AddressBits != 0 {
		return fmt.Errorf("%w: %s needs its original address-size binding", ErrProbeNeedsContext, ins.Op)
	}
	upper, err := c.loadFloatingVectorOperand(ins.Args[form.base], 16, 4, 32)
	if err != nil {
		return err
	}
	value, sourceType, constraint, template, err := c.scalarReciprocalNativeSource(ins.Args[0])
	if err != nil {
		return err
	}
	// Keep the scalar hardware approximation, not a packed estimate or generic
	// division/sqrt. The original opcode has no SIMD exceptions, ignores RC,
	// treats denormals as signed zero and leaves all MXCSR status untouched.
	assembly := strings.ToLower(string(ins.Op)) + " "
	constraints, args := "=x,", ""
	if form.spec.vector {
		assembly += template + ", $2, $0"
		constraints += constraint + ",x"
		args = fmt.Sprintf("%s %s, <4 x float> %s", sourceType, value, upper)
	} else {
		assembly += strings.Replace(template, "$1", "$2", 1) + ", $0"
		constraints += "0," + constraint
		args = fmt.Sprintf("<4 x float> %s, %s %s", upper, sourceType, value)
	}
	result, bytes := c.newTmp(), c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = call <4 x float> asm sideeffect %q, %q(%s)\n", result, assembly, constraints+",~{memory},~{fpsr}", args)
	fmt.Fprintf(c.b, "  %%%s = bitcast <4 x float> %%%s to <16 x i8>\n", bytes, result)
	if !form.spec.vector {
		return c.storeLegacyXViews(ins.Args[form.destination].Reg, "%"+bytes)
	}
	return c.storePackedMoveOperand(ins.Args[form.destination], 16, "%"+bytes)
}

func (c *amd64Ctx) scalarReciprocalNativeSource(source Operand) (value, typ, constraint, template string, err error) {
	if source.Kind == OpReg {
		value, err = c.loadFloatingVectorOperand(source, 16, 4, 32)
		return value, "<4 x float>", "x", "$1", err
	}
	if source.Kind == OpFP {
		bits, loadErr := c.loadScalarLaneFPBits(source.FPOffset, 32)
		if loadErr != nil {
			return "", "", "", "", loadErr
		}
		scalar, inserted := c.newTmp(), c.newTmp()
		fmt.Fprintf(c.b, "  %%%s = bitcast i32 %s to float\n", scalar, bits)
		fmt.Fprintf(c.b, "  %%%s = insertelement <4 x float> zeroinitializer, float %%%s, i32 0\n", inserted, scalar)
		return "%" + inserted, "<4 x float>", "x", "$1", nil
	}
	// A sideeffecting m32 operand keeps the original access even for Void.
	// This pointer is not a widened virtual FP slot or invented backing.
	pointer, pointerType, _, err := c.packedVectorMemoryPointer(source)
	if err != nil {
		return "", "", "", "", err
	}
	segment := ""
	if source.Kind == OpMem && source.Mem.Segment != "" {
		segment = "%" + strings.ToLower(string(source.Mem.Segment)) + ":"
	}
	return pointer, pointerType + " elementtype([4 x i8])", "*m", segment + "$1", nil
}

func (c *amd64Ctx) loadPackedReciprocalSource(source Operand, byteWidth, lanes int) (string, error) {
	if source.Kind == OpSym {
		// As in Go's x86 assembler, a bare integer is an absolute memory
		// operand. '$' is required for an immediate.
		if address, parseErr := parseInt(source.Sym); parseErr == nil {
			bytesValue := c.newTmp()
			fmt.Fprintf(c.b, "  %%%s = load <%d x i8>, ptr %s, align 1\n", bytesValue, byteWidth, c.ptrFromAddrI64(fmt.Sprintf("%d", address)))
			value := c.newTmp()
			fmt.Fprintf(c.b, "  %%%s = bitcast <%d x i8> %%%s to <%d x float>\n", value, byteWidth, bytesValue, lanes)
			return "%" + value, nil
		}
	}
	return c.loadFloatingVectorOperand(source, byteWidth, lanes, 32)
}
