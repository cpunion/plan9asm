package plan9asm

import (
	"fmt"
	"testing"
)

var x86WordMultiplyNames = []string{
	"PMULLW", "PMULHW", "PMULHUW", "PMULHRSW",
	"VPMULLW", "VPMULHW", "VPMULHUW", "VPMULHRSW",
}

func TestX86PackedWordMultiplyMode32LowZActualGo(t *testing.T) {
	llc := findLLVM22Tool("llc")
	if llc == "" {
		t.Fatal("LLVM 22 llc required")
	}
	for _, name := range x86WordMultiplyNames[4:] {
		for _, first := range []string{"Z0", "(AX)"} {
			row := name + " " + first + ", Z1, Z2"
			t.Run(row, func(t *testing.T) {
				source := "TEXT wordmul(SB),4,$0-0\n" + row + "\nRET\n"
				goAssemblerOracleOutput(t, "386", source, "", true)
				file, err := Parse(ArchAMD64, source)
				if err != nil {
					t.Fatal(err)
				}
				ir, err := Translate(file, Options{Goarch: "386", TargetTriple: "i386-unknown-linux-gnu",
					Sigs: map[string]FuncSig{"wordmul": {Name: "wordmul", Ret: Void}}})
				if err != nil {
					t.Fatal(err)
				}
				x86ResultInitializationObject(t, llc, "i386-unknown-linux-gnu", ir, "word-multiply-low-z")
			})
		}
	}
}

func TestX86PackedWordMultiplyLegacyByteEffectActualGo(t *testing.T) {
	for _, name := range x86WordMultiplyNames[:4] {
		for _, arch := range []string{"386", "amd64"} {
			row := name + " X0, X1"
			t.Run(arch+"/"+name, func(t *testing.T) {
				source := "TEXT wordmul(SB),4,$0-0\n" + row + "\nRET\n"
				goAssemblerOracleOutput(t, arch, source, "", true)
				file, err := Parse(ArchAMD64, source)
				if err != nil {
					t.Fatal(err)
				}
				ins := file.Funcs[0].Instrs[1]
				effect, known, err := x86TypedInstructionByteEffect(ins.Op, ins, arch)
				if !known || err != nil || len(effect.operands) != 2 {
					t.Fatalf("accepted legacy format lacks complete typed byte effect: known=%t err=%v effect=%+v", known, err, effect)
				}
				if !effect.operands[0].read || effect.operands[0].write || !effect.operands[1].read ||
					!effect.operands[1].write || effect.operands[0].width != 16 || effect.operands[1].width != 16 ||
					effect.flagsDefined != 0 || effect.flagsKilled != 0 {
					t.Fatal(fmt.Sprintf("incorrect read/modify/write/flags effect: %+v", effect))
				}
			})
		}
	}
}
