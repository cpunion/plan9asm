package plan9asm

import (
	"fmt"
	"strings"
)

func arm64ScalarConversionNativeTarget(triple string) bool {
	parts := strings.Split(triple, "-")
	return len(parts) >= 3 && (parts[0] == "aarch64" || parts[0] == "arm64")
}

func arm64RawScalarConversion(word uint32) bool {
	if _, ok := decodeARM64RawScalarFCVTToInt(word); ok {
		return true
	}
	if _, ok := decodeARM64RawScalarIntToFloat(word); ok {
		return true
	}
	if _, ok := decodeARM64RawScalarVectorFCVTZ(word); ok {
		return true
	}
	// The four pre-existing scalar SIMD SCVTF/UCVTF formats reach lowerFP
	// through the architecture decoder, not the arranged-vector registry.
	switch word &^ uint32(31<<5|31) {
	case 0x5e21d800, 0x5e61d800, 0x7e21d800, 0x7e61d800:
		return true
	}
	return false
}

func validateARM64ScalarConversionNativeTarget(file *File, options Options) error {
	if file == nil || file.Arch != ArchARM64 {
		return nil
	}
	needsNative := false
	var rawError error
	for _, function := range file.Funcs {
		for _, instruction := range function.Instrs {
			if instruction.Op == OpWORD {
				if len(instruction.Args) != 1 || instruction.Args[0].Kind != OpImm ||
					instruction.Args[0].ImmRaw != "" || instruction.Args[0].ImmIsFloat {
					if rawError == nil {
						rawError = fmt.Errorf("arm64 WORD expects exactly one resolved integer constant: %q", instruction.Raw)
					}
					continue
				}
				word := uint32(instruction.Args[0].Imm)
				if arm64RawScalarConversion(word) {
					needsNative = true
					continue
				}
				if _, valid := arm64RawContinuationEffectsForWord(word); valid {
					continue
				}
				if _, _, valid := decodeARM64RawSVEFamily(word); valid {
					continue
				}
				// A syntax-only placeholder lets the decoder check actual PC
				// encodings here. It grants no edge/layout proof: the original
				// source still takes the normal complete PC normalization path.
				if _, err := decodeARM64RawWordInstructionTarget(instruction, "preflight"); err != nil && rawError == nil {
					rawError = err
				}
				continue
			}
			_, matched, err := parseARM64ScalarConversion(arm64ControlOp(instruction), instruction)
			if err != nil {
				return err
			}
			needsNative = needsNative || matched
		}
	}
	if !needsNative {
		return nil
	}
	if rawError != nil {
		return rawError // Illegal/dead WORDs are not target-context successes.
	}
	if !arm64ScalarConversionNativeTarget(options.TargetTriple) || options.Goarch != "" && options.Goarch != "arm64" {
		return fmt.Errorf("%w: ARM64 scalar conversion requires an explicit native little-endian AArch64 FPCR/FPSR target, got %q/%q",
			ErrProbeNeedsContext, options.Goarch, options.TargetTriple)
	}
	return nil
}
