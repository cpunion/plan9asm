package plan9asm

import "testing"

// These four names and the complete encoder rows are independently checked in
// TestX86DwordProductCompleteIndependentEncoderRows. Reuse the common native
// input matrix, not the implementation's registry.
func x86DwordProductResultRows(arch string) []struct{ op, bank, mode string } {
	return x86FixedBinaryValueRows(arch, []string{"PMULDQ", "PMULULQ", "VPMULDQ", "VPMULUDQ"})
}

func TestX86DwordProductPairedNativeReadCompleteFamily(t *testing.T) {
	x86FixedBinaryCompleteFamilyObjects(t, x86DwordProductResultRows, 20, 8)
}

func TestX86DwordProductPairedNativeReadCannotBorrowInputs(t *testing.T) {
	x86FixedBinaryNativeInputBoundaries(t, x86DwordProductResultRows)
}

func TestX86DwordProductVoidCannotInventRegisterInputs(t *testing.T) {
	x86FixedBinaryVoidInputBoundaries(t, x86DwordProductResultRows)
}
