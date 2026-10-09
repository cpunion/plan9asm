package plan9asm

import (
	"fmt"
	"strings"
)

func armWordFloatNativeTarget(triple string) bool {
	parts := strings.Split(triple, "-")
	if len(parts) < 3 {
		return false
	}
	// Exact little-endian 32-bit ARM architecture components. In particular,
	// an "arm" prefix is not proof: arm64/armeb and invented variants differ.
	switch parts[0] {
	case "arm", "armv5te", "armv6", "armv7", "thumb", "thumbv7":
		return true
	default:
		return false
	}
}

func validateARMWordFloatNativeTarget(file *File, options Options) error {
	if file == nil || file.Arch != ArchARM {
		return nil
	}
	needsNative := false
	for _, function := range file.Funcs {
		for _, source := range function.Instrs {
			instruction := source
			if instruction.Op == OpWORD {
				decoded, err := decodeARMRawWordInstruction(instruction)
				if err != nil {
					continue // The ordinary raw validator retains its exact diagnostic.
				}
				instruction = decoded
			}
			op, _, _, _ := armDecodeOp(string(instruction.Op))
			_, matched, err := parseARMWordFloatConversion(op, instruction)
			if err != nil {
				return err // Invalid source is never reclassified as target context.
			}
			needsNative = needsNative || matched
		}
	}
	if needsNative && (!armWordFloatNativeTarget(options.TargetTriple) || options.Goarch != "" && options.Goarch != "arm") {
		return fmt.Errorf("%w: ARM word/float conversion requires an explicit native 32-bit little-endian ARM/Thumb FPSCR target, got %q/%q", ErrProbeNeedsContext, options.Goarch, options.TargetTriple)
	}
	return nil
}

// emitARMWordFloatNativeConversion preserves the complete VCVT effect, including
// FPSCR rounding, cumulative exceptions, exceptional inputs and dead outputs.
// LLVM's ordinary casts are pure and can instead fold to constants or poison.
func (c *armCtx) emitARMWordFloatNativeConversion(form armWordFloatConversionForm, input string) string {
	floatType, floatKind := "float", "f32"
	if form.floatBits == 64 {
		floatType, floatKind = "double", "f64"
	}
	wordKind := "s32"
	if form.unsigned {
		wordKind = "u32"
	}
	sourceType, resultType := floatType, "float"
	sourceKind, resultKind := floatKind, wordKind
	if form.wordToFloat {
		carrier := c.newTmp()
		fmt.Fprintf(c.b, "  %%%s = bitcast i32 %s to float\n", carrier, input)
		input = "%" + carrier
		sourceType, resultType = "float", floatType
		sourceKind, resultKind = wordKind, floatKind
	}
	result := c.newTmp()
	// LLVM22 ARM constraint t selects S0..31 for float and D0..15 for double.
	// This stays in the VFP2 register bank. The word is a bit carrier in an S
	// register, never a numeric floating cast. Sideeffect retains exceptions;
	// memory orders this call with the existing native VMRS/VMSR operations.
	operation := "vcvt." + resultKind + "." + sourceKind + " $0, $1"
	fmt.Fprintf(c.b, "  %%%s = call %s asm sideeffect %q, %q(%s %s)\n", result, resultType, operation, "=t,t,~{memory}", sourceType, input)
	if form.wordToFloat {
		return "%" + result
	}
	integer := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = bitcast float %%%s to i32\n", integer, result)
	return "%" + integer
}
