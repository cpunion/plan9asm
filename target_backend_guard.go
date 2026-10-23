package plan9asm

import (
	"fmt"
	"strings"

	"github.com/xgo-dev/llvm"
)

// Goarch selects the source encoder/ABI rules. It is not a backend selector:
// portable lowering may legitimately target a different LLVM architecture.
func validateBackendSourceGoarch(file *File, options Options) error {
	if file == nil || options.Goarch == "" {
		return nil
	}
	valid := true
	switch file.Arch {
	case ArchAMD64:
		valid = options.Goarch == "amd64" || options.Goarch == "386"
	case ArchARM:
		valid = options.Goarch == "arm"
	case ArchARM64:
		valid = options.Goarch == "arm64"
	case ArchWASM:
		valid = options.Goarch == "wasm"
	}
	if !valid {
		return fmt.Errorf("%w: source architecture %q conflicts with Goarch %q",
			ErrProbeNeedsContext, file.Arch, options.Goarch)
	}
	return nil
}

// validateTranslatedBackend checks the actual returned module, including early
// native-source returns. Keep target-independent source lowering portable.
func validateTranslatedBackend(file *File, options Options, module llvm.Module) error {
	triple := module.Target()
	architecture, _, _ := strings.Cut(triple, "-")
	if architecture == "wasm32" || architecture == "wasm64" {
		for function := module.FirstFunction(); !function.IsNil(); function = llvm.NextFunction(function) {
			for block := function.FirstBasicBlock(); !block.IsNil(); block = llvm.NextBasicBlock(block) {
				// A BasicBlock is not an ordinary LLVM constant operand. Its
				// constant users are BlockAddress values, whether those addresses
				// are nested in expressions, aggregates, aliases or instructions.
				// Inspect their actual def-use edges rather than printed names,
				// source annotations or the spelling of a particular lowerer.
				for use := block.AsValue().FirstUse(); !use.IsNil(); use = use.NextUse() {
					if !use.User().IsAConstant().IsNil() {
						return fmt.Errorf("%w: LLVM target %q cannot represent blockaddress in %q",
							ErrProbeNeedsContext, triple, function.Name())
					}
				}
			}
		}
	}
	if arm64ScalarConversionNativeTarget(triple) {
		return nil
	}
	if file != nil && file.Arch == ArchARM64 {
		// Reuse the complete named/raw family classifier even when a caller
		// supplies custom attributes. Raw literal-pool words are data, not SVE
		// instructions. Pure scalable-vector operations also need this check;
		// they need not contain an llvm.aarch64 intrinsic call.
		view := arm64NativeTargetCodeView(file)
		for _, function := range view.Funcs {
			features := inferFuncTargetFeaturesForGOARCH(view.Arch, options.Goarch, function)
			for _, feature := range strings.Split(features, ",") {
				if strings.HasPrefix(feature, "+sve") {
					return fmt.Errorf("%w: ARM64 SVE requires an explicit AArch64 LLVM target, got %q",
						ErrProbeNeedsContext, triple)
				}
			}
		}
	}
	// Declarations in an architecture prelude are not executed requirements.
	// An actually referenced target intrinsic must nevertheless stay on its
	// backend, including any future lowering not yet in the source classifier.
	for function := module.FirstFunction(); !function.IsNil(); function = llvm.NextFunction(function) {
		if strings.HasPrefix(function.Name(), "llvm.aarch64.sve.") && !function.FirstUse().IsNil() {
			return fmt.Errorf("%w: used ARM64 SVE intrinsic %q requires an AArch64 LLVM target, got %q",
				ErrProbeNeedsContext, function.Name(), triple)
		}
	}
	return nil
}
