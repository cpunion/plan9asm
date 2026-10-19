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
		return c.emitNativeFP16ThreeOperand(ins, x86NativeFP16ThreeOperand{
			byteWidth: 16, lanes: 1, memoryBytes: f.memoryBytes, second: f.upper,
			destination: f.destination, mask: f.mask,
			controls: f.controls, sae: f.sae,
		})
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

// FP16 scalar conversion/SQRT and scalar/packed arithmetic share a three-operand
// native encoding, not arithmetic semantics. Preserve the actual opcode and
// memory operand so masking, rounding, exceptions and dead accesses survive.
type x86NativeFP16ThreeOperand struct {
	byteWidth, lanes, memoryBytes int
	second, destination, mask     int
	controls                      amd64FMA3Suffix
	sae, readDestination          bool
}

func (c *amd64Ctx) emitNativeFP16ThreeOperand(ins Instr, f x86NativeFP16ThreeOperand) error {
	source := ins.Args[0]
	if source.Kind == OpFP || ins.x86AddressBits != 0 || source.Kind == OpMem &&
		!x86OrdinaryNativeMemoryOperand(source, c.goarch) {
		return fmt.Errorf("%w: FP16 instruction needs original native backing", ErrProbeNeedsContext)
	}
	typ := fmt.Sprintf("<%d x float>", f.byteWidth/4) // Bit carrier, never arithmetic precision.
	load := func(index int) (string, error) {
		bytes, err := c.loadPackedCompareBytes(ins.Args[index], f.byteWidth)
		if err != nil {
			return "", err
		}
		value := c.newTmp()
		fmt.Fprintf(c.b, "  %%%s = bitcast <%d x i8> %s to %s\n", value, f.byteWidth, bytes, typ)
		return "%" + value, nil
	}
	second, err := load(f.second)
	if err != nil {
		return err
	}
	old := "zeroinitializer"
	if f.readDestination || f.mask >= 0 && !f.controls.zeroing {
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
	args := fmt.Sprintf("%s %s, %s %s, %s %s", typ, old, typ, second, inputType, input)
	constraints := "=v,0,v," + inputConstraint
	mnemonic, _, _ := strings.Cut(strings.ToLower(string(ins.Op)), ".")
	assembly := mnemonic + " "
	if f.controls.rounding != "" {
		assembly += "{" + strings.ToLower(strings.TrimSuffix(f.controls.rounding, "_SAE")) + "-sae}, "
	} else if f.sae {
		assembly += "{sae}, "
	}
	assembly += inputTemplate
	if f.controls.broadcast {
		assembly += fmt.Sprintf("{1to%d}", f.lanes)
	}
	assembly += ", $2, $0"
	if f.mask >= 0 {
		mask, err := c.loadK(ins.Args[f.mask].Reg)
		if err != nil {
			return err
		}
		maskMove := "kmovw"
		if f.lanes > 16 {
			maskMove = "kmovd"
		}
		assembly = maskMove + " ${4:k}, %k1; " + assembly + " {%k1}"
		if f.controls.zeroing {
			assembly += " {z}"
		}
		constraints += ",r,~{k1}"
		args += ", i32 " + c.approximateMathMask32(mask)
	}
	result, bytes := c.newTmp(), c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = call %s asm sideeffect %q, %q(%s)\n", result, typ, assembly, constraints+",~{memory},~{fpsr}", args)
	fmt.Fprintf(c.b, "  %%%s = bitcast %s %%%s to <%d x i8>\n", bytes, typ, result, f.byteWidth)
	return c.storePackedMoveOperand(ins.Args[f.destination], f.byteWidth, "%"+bytes)
}
