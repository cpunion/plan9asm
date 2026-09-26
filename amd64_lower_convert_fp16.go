package plan9asm

import (
	"fmt"
	"strings"
)

// FP16 map5/map6 conversions are both load forms. The lane types determine
// access and result widths; encoded VL distinguishes the two X destinations.
func (c *amd64Ctx) lowerRawFP16Conversion(ins Instr, spec amd64HalfConversionSpec) (bool, bool, error) {
	if !ins.x86Encoded {
		return true, false, fmt.Errorf("%s has no named Go assembler form: %q", ins.Op, ins.Raw)
	}
	vl := ins.x86VectorBytes
	if vl != 16 && vl != 32 && vl != 64 {
		return true, false, fmt.Errorf("%s has invalid encoded vector length %d", ins.Op, vl)
	}
	if len(ins.Args) != 2 && len(ins.Args) != 3 {
		return true, false, fmt.Errorf("%s requires source, [K mask,] destination", ins.Op)
	}
	base := strings.SplitN(string(ins.Op), ".", 2)[0]
	suffix := strings.TrimPrefix(string(ins.Op), base)
	sae := suffix == ".SAE" || suffix == ".SAE.Z"
	if sae {
		suffix = strings.TrimPrefix(suffix, ".SAE")
	}
	properties, valid := parseAMD64FMA3Suffix(strings.TrimPrefix(suffix, "."))
	if !valid || spec.inputBits == 16 && properties.rounding != "" || spec.inputBits == 32 && sae {
		return true, false, fmt.Errorf("%s has invalid conversion controls", ins.Op)
	}
	masked := len(ins.Args) == 3
	if properties.zeroing && !masked || masked && !amd64NonzeroKOperand(ins.Args[1]) {
		return true, false, fmt.Errorf("%s requires K1-K7 for masking", ins.Op)
	}
	source, destination := ins.Args[0], ins.Args[len(ins.Args)-1]
	lanes := vl / 4
	inputBytes, outputBytes := lanes*spec.inputBits/8, lanes*spec.outputBits/8
	sourceRegisterBytes, destinationRegisterBytes := inputBytes, outputBytes
	if sourceRegisterBytes < 16 {
		sourceRegisterBytes = 16
	}
	if destinationRegisterBytes < 16 {
		destinationRegisterBytes = 16
	}
	if !c.isGoEVEXVectorRegister(destination, destinationRegisterBytes) {
		return true, false, fmt.Errorf("%s has an invalid destination register", ins.Op)
	}
	if source.Kind == OpReg {
		if !c.isGoEVEXVectorRegister(source, sourceRegisterBytes) || properties.broadcast {
			return true, false, fmt.Errorf("%s has invalid source width or register broadcast", ins.Op)
		}
	} else if !isAMD64MemoryOperand(source) || sae || properties.rounding != "" {
		return true, false, fmt.Errorf("%s has invalid memory or rounding operands", ins.Op)
	}
	if (sae || properties.rounding != "") && vl != 64 {
		return true, false, fmt.Errorf("%s explicit rounding/SAE requires 512-bit VL", ins.Op)
	}

	mask := ""
	var err error
	if masked {
		mask, err = c.loadK(ins.Args[1].Reg)
		if err != nil {
			return true, false, err
		}
	}
	var input string
	if source.Kind == OpReg || !masked && !properties.broadcast {
		input, err = c.loadPackedExtendInputs(source, sourceRegisterBytes, lanes, spec.inputBits)
	} else {
		input, err = c.loadMaskedPackedCompareLanes(source, inputBytes, spec.inputBits, properties.broadcast, mask)
	}
	if err != nil {
		return true, false, err
	}
	inputType, outputType, conversion := "half", "float", "fpext"
	if spec.inputBits == 32 {
		inputType, outputType, conversion = "float", "half", "fptrunc"
	}
	floats, converted := c.newTmp(), c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = bitcast <%d x i%d> %s to <%d x %s>\n", floats, lanes, spec.inputBits, input, lanes, inputType)
	fmt.Fprintf(c.b, "  %%%s = %s <%d x %s> %%%s to <%d x %s>\n", converted, conversion, lanes, inputType, floats, lanes, outputType)
	bits := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = bitcast <%d x %s> %%%s to <%d x i%d>\n", bits, lanes, outputType, converted, lanes, spec.outputBits)
	result := "%" + bits
	if spec.inputBits == 32 {
		result = c.adjustFP16NarrowRounding(lanes, "%"+floats, "%"+converted, result, properties.rounding)
	}
	if masked {
		old, err := c.loadPackedExtendInputs(destination, destinationRegisterBytes, lanes, spec.outputBits)
		if err != nil {
			return true, false, err
		}
		result = amd64ApplyIntegerLaneMask(c, lanes, spec.outputBits, result, old, mask, properties.zeroing)
	}
	bytes := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = bitcast <%d x i%d> %s to <%d x i8>\n", bytes, lanes, spec.outputBits, result, outputBytes)
	return true, false, c.storePackedHalfResult(destination, destinationRegisterBytes, outputBytes, "%"+bytes)
}

// Directed rounding differs from nearest-even by at most one half ULP.
// Comparing the exactly widened result lets integer correction preserve
// signed zero, subnormals, NaNs and directed overflow without changing the
// host FP environment. LLVM 22's AArch64 constrained fptrunc does not honor
// these static rounding modes for half, so it cannot serve as this oracle.
func (c *amd64Ctx) adjustFP16NarrowRounding(lanes int, original, halves, bits, rounding string) string {
	if rounding == "" || rounding == "RN_SAE" {
		return bits
	}
	widened, below, above, negative := c.newTmp(), c.newTmp(), c.newTmp(), c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = fpext <%d x half> %s to <%d x float>\n", widened, lanes, halves, lanes)
	fmt.Fprintf(c.b, "  %%%s = fcmp olt <%d x float> %%%s, %s\n", below, lanes, widened, original)
	fmt.Fprintf(c.b, "  %%%s = fcmp ogt <%d x float> %%%s, %s\n", above, lanes, widened, original)
	fmt.Fprintf(c.b, "  %%%s = icmp slt <%d x i16> %s, zeroinitializer\n", negative, lanes, bits)
	one := amd64SplatInteger(c, lanes, 16, "1")
	plus, minus := c.newTmp(), c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = add <%d x i16> %s, %s\n", plus, lanes, bits, one)
	fmt.Fprintf(c.b, "  %%%s = sub <%d x i16> %s, %s\n", minus, lanes, bits, one)
	condition, adjusted := "%"+above, "%"+minus
	choice := c.newTmp()
	switch rounding {
	case "RD_SAE":
		fmt.Fprintf(c.b, "  %%%s = select <%d x i1> %%%s, <%d x i16> %%%s, <%d x i16> %%%s\n", choice, lanes, negative, lanes, plus, lanes, minus)
		adjusted = "%" + choice
	case "RU_SAE":
		fmt.Fprintf(c.b, "  %%%s = select <%d x i1> %%%s, <%d x i16> %%%s, <%d x i16> %%%s\n", choice, lanes, negative, lanes, minus, lanes, plus)
		condition, adjusted = "%"+below, "%"+choice
	case "RZ_SAE":
		fmt.Fprintf(c.b, "  %%%s = select <%d x i1> %%%s, <%d x i1> %%%s, <%d x i1> %%%s\n", choice, lanes, negative, lanes, below, lanes, above)
		condition = "%" + choice
	}
	result := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = select <%d x i1> %s, <%d x i16> %s, <%d x i16> %s\n", result, lanes, condition, lanes, adjusted, lanes, bits)
	return "%" + result
}
