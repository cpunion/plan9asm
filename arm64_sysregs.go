package plan9asm

import "fmt"

//go:generate go run ./cmd/plan9asmgenregs -out arm64_sysregs_generated.go

// A named Go register and a raw MRS/MSR word share the same physical operand
// grammar. Canonical encodings avoid LLVM's feature-gated aliases without
// replacing a system-register read with an invented constant.
type arm64SystemRegisterSpec struct {
	encoding           uint16
	readable, writable bool
}

func arm64CheckedSystemRegister(name string, read bool) (string, error) {
	if spec, ok := arm64GoSystemRegisters[name]; ok {
		if (read && !spec.readable) || (!read && !spec.writable) {
			access := "writable"
			if read {
				access = "readable"
			}
			return "", fmt.Errorf("arm64 system register %s is not %s", name, access)
		}
		return arm64EncodedSystemRegisterName(spec.encoding), nil
	}
	// Encoded names also occur in decoded raw instructions. Their grammar is
	// validated by the raw decoder and LLVM's assembler, not a named alias table.
	return arm64CanonicalSysReg(name), nil
}
