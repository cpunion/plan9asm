package plan9asm

import (
	"fmt"
	"strings"
)

func validateX86TLSMoveFile(file *File, arch string) error {
	if file == nil || file.Arch != ArchAMD64 {
		return nil
	}
	if arch == "" {
		arch = "amd64"
	}
	for _, function := range file.Funcs {
		for _, instruction := range function.Instrs {
			if err := x86SourceTLSMoveError(instruction, arch); err != nil {
				return fmt.Errorf("%s: %w", function.Sym, err)
			}
		}
	}
	return nil
}

// Validate TLS's complete ymovb/w/l/q memory rows before scalar, MMX or XMM
// dispatch. This does not rewrite TLS to a GP register or a guessed segment.
// No API here owns Go's runtime TLS relocation, g layout or thread transport.
func x86SourceTLSMoveError(ins Instr, arch string) error {
	op := strings.ToUpper(string(ins.Op))
	base := strings.SplitN(op, ".", 2)[0]
	if base != "MOVB" && base != "MOVW" && base != "MOVL" && base != "MOVQ" {
		return nil
	}
	present := false
	for _, operand := range ins.Args {
		present = present || operand.Kind == OpReg && operand.Reg == TLS ||
			operand.Kind == OpMem && (operand.Mem.Base == TLS || operand.Mem.Index == TLS)
	}
	if !present {
		return nil
	}
	bad := func(reason string) error {
		return fmt.Errorf("%s %s %s: %q", arch, op, reason, ins.Raw)
	}
	if arch != "386" && arch != "amd64" || op != base || len(ins.Args) != 2 {
		return bad("has an invalid mode, suffix or operand count")
	}
	source, destination := ins.Args[0], ins.Args[1]
	if source.Kind == OpReg && source.Reg == TLS {
		if base != "MOVL" && base != "MOVQ" || destination.Kind != OpReg || !isX86YrlRegisterForArch(destination.Reg, arch) {
			return bad("requires the original Ytls-to-Yrl row")
		}
		// progedit can NOP either width on direct-TLS systems; initial-exec
		// movTLSReg instead checks the native pointer width. Do not guess.
		return fmt.Errorf("%w: %s needs its actual Go OS/arch/shared TLS source contract: %q", ErrProbeNeedsContext, op, ins.Raw)
	}
	memoryIndex := -1
	for index, operand := range ins.Args {
		if operand.Kind == OpMem && (operand.Mem.Base == TLS || operand.Mem.Index == TLS) {
			if memoryIndex >= 0 || !x86SourceTLSMemoryRole(operand, arch) {
				return bad("requires one encodable TLS memory role")
			}
			memoryIndex = index
		}
	}
	if memoryIndex < 0 {
		return bad("has no row with a TLS destination register")
	}
	other := ins.Args[1-memoryIndex]
	valid := false
	switch other.Kind {
	case OpImm:
		valid = memoryIndex == 1 && !other.ImmIsFloat && other.ImmRaw == "" &&
			(base != "MOVQ" || arch == "amd64") &&
			(arch == "386" || other.Imm >= -1<<31 && other.Imm <= 1<<32-1)
	case OpReg:
		valid = isX86YrlRegisterForArch(other.Reg, arch) && (base != "MOVQ" || arch == "amd64")
		if base == "MOVB" {
			valid = isGoYmbRegisterForArch(other.Reg, arch)
		}
		if base == "MOVW" {
			switch other.Reg {
			case CS, SS, DS, ES, FS, GS:
				valid = true // The complete movtab Yseg/Yml rows.
			}
		}
		if base == "MOVL" || base == "MOVQ" {
			if _, mmx := amd64ParseMReg(other.Reg); mmx {
				valid = true
			}
			if index, xmm := amd64ParseXReg(other.Reg); xmm && index < 16 && (arch == "amd64" || index < 8) {
				valid = true
			}
		}
	case OpSym:
		if memoryIndex == 1 && arch == "386" && base != "MOVQ" && strings.HasPrefix(other.Sym, "$") {
			_, _, valid = x86ScalarCompareSBAddress(Operand{Kind: OpSym, Sym: strings.TrimPrefix(other.Sym, "$")}, arch)
		}
	}
	if !valid {
		return bad("operands do not match a complete Go move row")
	}
	return fmt.Errorf("%w: %s TLS memory needs its actual Go OS/arch/shared progedit, relocation and native TLS address contract: %q", ErrProbeNeedsContext, op, ins.Raw)
}
