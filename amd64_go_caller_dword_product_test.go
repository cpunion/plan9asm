package plan9asm

import (
	"strings"
	"testing"
)

func TestX86GoCallerDwordProductIndependentSourceEffect(t *testing.T) {
	for _, row := range []string{"PMULDQ X0,X1", "PMULULQ X0,X1", "VPMULDQ Y0,Y1,Y2", "VPMULUDQ Y0,Y1,Y2"} {
		t.Run(row, func(t *testing.T) {
			result := "X2"
			if !strings.HasPrefix(row, "V") {
				result = "X1"
			}
			body := "VMOVUPS (DI),Z0\nVMOVUPS (SI),Z1\n" + row + "\nVMOVSS " + result + ",(DX)"
			assembly := strings.Replace(x86CallerFixtureAssembly, "MOVL AX,(DX)", body, 1)
			requireX86GoAssemblerResult(t, "amd64", assembly, true)
			pkg, _ := x86CallerCoreFixture(t, x86CallerFixtureSource, assembly)
			if err := x86CallerCoreTranslate(t, pkg, assembly, "x86_64-unknown-linux-gnu"); err != nil {
				t.Fatal(err)
			}
		})
	}
}
