package plan9asm

import (
	"fmt"
	"strings"
)

func (c *amd64Ctx) emitNativeFP16Conversion(ins Instr, f x86FP16ConversionForm) error {
	source := ins.Args[0]
	if source.Kind == OpFP || ins.x86AddressBits != 0 || source.Kind == OpMem &&
		!x86OrdinaryNativeMemoryOperand(source, c.goarch) {
		return fmt.Errorf("%w: FP16 conversion needs original native backing", ErrProbeNeedsContext)
	}
	if f.spec.scalar {
		return c.emitNativeScalarFP16Conversion(ins, f)
	}
	mnemonic, _, _ := strings.Cut(strings.ToLower(string(ins.Op)), ".")
	if source.Kind != OpReg && f.spec.outputBits == 16 && !f.controls.broadcast && f.sourceBytes < 64 {
		mnemonic += map[int]string{16: "x", 32: "y"}[f.sourceBytes]
	}
	return c.emitNativePackedConversion(ins, x86NativePackedConversion{
		byteWidth: f.byteWidth, sourceBytes: f.sourceBytes, memoryBytes: f.memoryBytes,
		outputBits:  32, // Bit carrier only: the original opcode selects FP16/FP32.
		destination: f.destination, mask: f.mask, lanes: f.lanes, evex: true,
		broadcast: f.controls.broadcast, zeroing: f.controls.zeroing,
		rounding: f.controls.rounding, sae: f.sae, mnemonic: mnemonic,
	})
}

// Scalar conversion copies upper X bytes from its independent first source.
// Keep the original native memory operand so masking and exceptions, including
// dead results, retain the instruction's own access and arithmetic behavior.
func (c *amd64Ctx) emitNativeScalarFP16Conversion(ins Instr, f x86FP16ConversionForm) error {
	const typ = "<4 x float>"
	load := func(index int) (string, error) {
		bytes, err := c.loadX(ins.Args[index].Reg)
		if err != nil {
			return "", err
		}
		value := c.newTmp()
		fmt.Fprintf(c.b, "  %%%s = bitcast <16 x i8> %s to %s\n", value, bytes, typ)
		return "%" + value, nil
	}
	upper, err := load(f.upper)
	if err != nil {
		return err
	}
	old := "zeroinitializer"
	if f.mask >= 0 && !f.controls.zeroing {
		old, err = load(f.destination)
		if err != nil {
			return err
		}
	}
	inputType, inputConstraint, inputTemplate := typ, "v", "$3"
	var input string
	if ins.Args[0].Kind == OpReg {
		input, err = load(0)
	} else {
		var pointerType string
		input, pointerType, _, err = c.packedVectorMemoryPointer(ins.Args[0])
		inputType = fmt.Sprintf("%s elementtype([%d x i8])", pointerType, f.memoryBytes)
		inputConstraint = "*m"
		if pointerType == "ptr addrspace(256)" {
			inputTemplate = "%gs:" + inputTemplate
		} else if pointerType == "ptr addrspace(257)" {
			inputTemplate = "%fs:" + inputTemplate
		}
	}
	if err != nil {
		return err
	}
	args := fmt.Sprintf("%s %s, %s %s, %s %s", typ, old, typ, upper, inputType, input)
	constraints := "=v,0,v," + inputConstraint
	mnemonic, _, _ := strings.Cut(strings.ToLower(string(ins.Op)), ".")
	assembly := mnemonic + " "
	if f.controls.rounding != "" {
		assembly += "{" + strings.ToLower(strings.TrimSuffix(f.controls.rounding, "_SAE")) + "-sae}, "
	} else if f.sae {
		assembly += "{sae}, "
	}
	assembly += inputTemplate + ", $2, $0"
	if f.mask >= 0 {
		mask, err := c.loadK(ins.Args[f.mask].Reg)
		if err != nil {
			return err
		}
		assembly = "kmovw ${4:k}, %k1; " + assembly + " {%k1}"
		if f.controls.zeroing {
			assembly += " {z}"
		}
		constraints += ",r,~{k1}"
		args += ", i32 " + c.approximateMathMask32(mask)
	}
	result, bytes := c.newTmp(), c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = call %s asm sideeffect %q, %q(%s)\n", result, typ, assembly, constraints+",~{memory},~{fpsr}", args)
	fmt.Fprintf(c.b, "  %%%s = bitcast %s %%%s to <16 x i8>\n", bytes, typ, result)
	return c.storePackedMoveOperand(ins.Args[f.destination], 16, "%"+bytes)
}
