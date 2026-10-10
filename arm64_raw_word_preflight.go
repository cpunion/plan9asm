package plan9asm

import "fmt"

// Native-target preflights must not disguise malformed WORDs as missing
// target context. This is syntax-only: registered forms still require their
// original source, machine-entry, continuation and layout proofs later.
func arm64RawWordPreflightError(instruction Instr) error {
	if instruction.Op != OpWORD {
		return nil
	}
	if len(instruction.Args) != 1 || instruction.Args[0].Kind != OpImm ||
		instruction.Args[0].ImmRaw != "" || instruction.Args[0].ImmIsFloat {
		return fmt.Errorf("arm64 WORD expects exactly one resolved integer constant: %q", instruction.Raw)
	}
	word := uint32(instruction.Args[0].Imm)
	if arm64RawScalarConversion(word) {
		return nil
	}
	if _, valid := arm64RawContinuationEffectsForWord(word); valid {
		return nil
	}
	if _, valid := arm64RawSVEEffects(word); valid {
		return nil
	}
	// The placeholder only lets the decoder check PC-relative syntax. It is
	// never an edge or source-location witness consumed by translation.
	_, err := decodeARM64RawWordInstructionTarget(instruction, "preflight")
	return err
}
