package plan9asm

import (
	"strings"
	"testing"
)

func TestX86GoCallerSADOriginalRawEffect(t *testing.T) {
	// Unchanged instruction from Weaviate v1.39.6 hamming_avx256_amd64.s:77.
	// The independent fixture binds its own complete selected typed callers;
	// it does not borrow the external kernel's unresolved loop/frame contract.
	body := "VMOVDQU (DI),Y2\nVMOVDQU (SI),Y5\nLONG $0xeaf6d5c5\nVMOVSS X5,(DX)"
	assembly := strings.Replace(x86CallerFixtureAssembly, "MOVL AX,(DX)", body, 1)
	requireX86GoAssemblerResult(t, "amd64", assembly, true)
	pkg, _ := x86CallerCoreFixture(t, x86CallerFixtureSource, assembly)
	if err := x86CallerCoreTranslate(t, pkg, assembly, "x86_64-unknown-linux-gnu"); err != nil {
		t.Fatal(err)
	}
}

func TestX86GoCallerSADCompleteNamedFamily(t *testing.T) {
	for _, row := range []struct {
		instruction string
		result      string
	}{
		{"PSADBW X0,X1", "X1"},
		{"VPSADBW Y0,Y1,Y2", "X2"},
		{"MPSADBW $5,X0,X1", "X1"},
		{"VMPSADBW $37,Y0,Y1,Y2", "X2"},
		{"VDBPSADBW $228,Z0,Z1,K1,Z2", "X2"},
	} {
		t.Run(row.instruction, func(t *testing.T) {
			body := "VMOVUPS (DI),Z0\nVMOVUPS (SI),Z1\nVMOVUPS (DI),Z2\nMOVL $1,AX\nKMOVQ AX,K1\n" + row.instruction + "\nVMOVSS " + row.result + ",(DX)"
			assembly := strings.Replace(x86CallerFixtureAssembly, "MOVL AX,(DX)", body, 1)
			requireX86GoAssemblerResult(t, "amd64", assembly, true)
			pkg, _ := x86CallerCoreFixture(t, x86CallerFixtureSource, assembly)
			if err := x86CallerCoreTranslate(t, pkg, assembly, "x86_64-unknown-linux-gnu"); err != nil {
				t.Fatal(err)
			}
		})
	}
}
