package plan9asm

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

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

func TestX86DwordProductRawPairedNativeObjects(t *testing.T) {
	llc := findLLVM22Tool("llc")
	if llc == "" {
		t.Fatal("LLVM22 llc required")
	}
	for _, target := range x86RawReturnTargets {
		t.Run(target.triple, func(t *testing.T) {
			var source, declarations strings.Builder
			count := 0
			for _, row := range x86FixedBinaryValueRows("amd64", []string{
				"VPMULDQ", "VPMULUDQ", "VPMULDQ.BCST", "VPMULUDQ.BCST",
			}) {
				body, declaration := x86FixedBinaryValueFixture(target.arch, row.op, row.bank, row.mode, false)
				file, err := Parse(ArchAMD64, body)
				if err != nil {
					t.Fatal(err)
				}
				var instruction string
				for _, ins := range file.Funcs[0].Instrs {
					if strings.HasPrefix(string(ins.Op), "VPMULDQ") || strings.HasPrefix(string(ins.Op), "VPMULUDQ") {
						instruction = ins.Raw
					}
				}
				if instruction == "" {
					t.Fatal("independent source lost its arithmetic instruction")
				}
				code := x86OrderingHintGoBytes(t, "amd64", "TEXT original(SB),4,$0-0\n"+instruction+"\nRET\n")
				body = strings.Replace(body, instruction, x86DwordProductDirectives(code[:len(code)-1]), 1)
				name := fmt.Sprintf("ProductRaw%d", count)
				source.WriteString(strings.ReplaceAll(body, "Value", name))
				declarations.WriteString(strings.ReplaceAll(declaration, "Value", name) + "\n")
				count++
			}
			if count != 36 {
				t.Fatal("complete raw full/broadcast/mask denominator changed", count)
			}
			x86LEAActualTargetGoObject(t, target.arch, target.triple, source.String())
			for _, api := range []string{"text", "module", "Go"} {
				ir, err := x86ScalarFPBitViewAPI(t, source.String(), declarations.String(), target.arch, target.triple, api)
				if err != nil {
					t.Fatal("exact raw input source lost its paired native continuation", api, err)
				}
				for _, text := range []string{ir, x86DwordProductIROptimize(t, ir)} {
					for _, level := range []string{"-O0", "-O2"} {
						x86PackedStringCompileObject(t, llc, target.triple, text, level)
					}
				}
			}
		})
	}
}

func TestX86DwordProductRawVoidCannotInventInputs(t *testing.T) {
	for _, row := range x86FixedBinaryValueRows("386", []string{"VPMULDQ", "VPMULUDQ"}) {
		instruction := fmt.Sprintf("%s %s0,%s1,%s2", row.op, row.bank, row.bank, row.bank)
		code := x86OrderingHintGoBytes(t, "amd64", "TEXT original(SB),4,$0-0\n"+instruction+"\nRET\n")
		source := "TEXT ·Value(SB),4,$0-0\n" + x86DwordProductDirectives(code[:len(code)-1]) + "RET\n"
		for _, target := range x86RawReturnTargets {
			for _, api := range []string{"text", "module", "Go"} {
				ir, err := x86ScalarFPBitViewAPI(t, source, "func Value()", target.arch, target.triple, api)
				if ir != "" || !errors.Is(err, ErrProbeNeedsContext) {
					t.Fatal("raw Void opcode borrowed zeroed implementation registers", instruction, target.triple, api, err)
				}
			}
		}
	}
}
