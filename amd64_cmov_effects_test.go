package plan9asm

import (
	"strings"
	"testing"
)

func TestX86CMOVFamilyIndependentCallerSourceEffect(t *testing.T) {
	for _, row := range []string{"CMOVWCC BX,AX", "CMOVLLE BX,AX", "CMOVQPS BX,AX", "CMOVLLE (DI),AX"} {
		t.Run(row, func(t *testing.T) {
			body := "MOVQ $9,BX\nCMPQ AX,BX\n" + row + "\nMOVL AX,(DX)"
			assembly := strings.Replace(x86CallerFixtureAssembly, "MOVL AX,(DX)", body, 1)
			goAssemblerOracleOutput(t, "amd64", assembly, "", true)
			pkg, _ := x86CallerCoreFixture(t, x86CallerFixtureSource, assembly)
			if err := x86CallerCoreTranslate(t, pkg, assembly, "x86_64-unknown-linux-gnu"); err != nil {
				t.Fatal("actual-Go conditional move has no complete caller source effect", err)
			}
		})
	}
}

func TestX86CMOVFamilyIndependentResultByteEffect(t *testing.T) {
	for _, row := range []string{"CMOVWCC AX,BX", "CMOVLLE AX,BX", "CMOVQPS AX,BX", "CMOVLLE AX,AX"} {
		t.Run(row, func(t *testing.T) {
			result := "BX"
			if row == "CMOVLLE AX,AX" {
				result = "AX"
			}
			body := "MOVQ x+0(FP),AX\nMOVQ $9,BX\nCMPQ AX,BX\n" + row + "\nMOVQ " + result + ",ret0+8(FP)\nRET"
			_, file, sig, source := x86ResultInitializationFixture(t, "amd64", "func Value(x uint64) uint64", body)
			goAssemblerOracleOutput(t, "amd64", source, "", true)
			if err := CheckX86ABI0ResultInitialization(file.Funcs[0], sig, "amd64"); err != nil {
				t.Fatal("actual-Go conditional move has no complete result byte effect", err)
			}
		})
	}
}

func TestX86CMOVFamilyIndependentDeadNativeRead(t *testing.T) {
	for _, row := range []string{"CMOVWCC (DI),BX", "CMOVLLE (DI),BX", "CMOVQPS (DI),BX"} {
		t.Run(row, func(t *testing.T) {
			source := "TEXT ·Dead(SB),4,$0-0\nMOVQ $7,AX\nMOVQ $9,BX\nCMPQ AX,BX\n" + row + "\nRET\n"
			goAssemblerOracleOutput(t, "amd64", source, "", true)
			ir := x86NarrowOracleIR(t, source, "package cmov\nfunc Dead()\n", "x86_64-unknown-linux-gnu")
			optimized := x86DwordProductIROptimize(t, ir)
			t.Logf("original loads=%t; actual IR default<O2> retains any read=%t", strings.Contains(ir, " = load "), strings.Contains(optimized, " = load "))
			if !strings.Contains(ir, "load volatile") || !strings.Contains(optimized, "load volatile") {
				t.Fatal("unconditional source read can disappear with a dead CMOV result under real IR default<O2>")
			}
		})
	}
}
