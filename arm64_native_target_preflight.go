package plan9asm

import "errors"

// A native-target requirement is deferred until every registered family's
// source grammar has been checked. Otherwise a valid conversion can hide an
// invalid compare or system-register transfer later in the same source file.
// This changes error precedence only; it never supplies a native target or
// any entry, memory, flag, source-encoder or continuation permission.
func validateARM64NativeTargets(file *File, options Options) error {
	if file == nil || file.Arch != ArchARM64 {
		return nil
	}
	var targetError error
	for _, validate := range []func(*File, Options) error{
		validateARM64ScalarConversionNativeTarget,
		validateARM64ScalarCompareTarget,
		validateARM64FPControlTarget,
		validateARM64ScalarBinaryTarget,
		validateARM64ScalarMinMaxTarget,
		validateARM64RNDRTarget,
	} {
		if err := validate(file, options); err != nil {
			if !errors.Is(err, ErrProbeNeedsContext) {
				return err
			}
			if targetError == nil {
				targetError = err
			}
		}
	}
	if targetError != nil {
		// Some native families do not themselves claim unrelated WORDs.
		// Their target error still cannot hide a malformed instruction.
		for _, function := range file.Funcs {
			for _, instruction := range function.Instrs {
				if err := arm64RawWordPreflightError(instruction); err != nil && !errors.Is(err, ErrProbeNeedsContext) {
					return err
				}
			}
		}
	}
	return targetError
}
