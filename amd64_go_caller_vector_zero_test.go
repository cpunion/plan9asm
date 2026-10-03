package plan9asm

import (
	"strings"
	"testing"
)

func TestX86GoCallerVectorZeroOriginalSource(t *testing.T) {
	for _, body := range []string{
		"VMOVUPS (DI),Z0\nVZEROUPPER\nVMOVSS X0,(DX)",
		"VZEROALL\nVMOVSS X0,(DX)",
		"VMOVUPS (DI),Z0\nWORD $0xf8c5\nBYTE $0x77\nVMOVSS X0,(DX)",
		"WORD $0xfcc5\nBYTE $0x77\nVMOVSS X0,(DX)",
	} {
		t.Run(strings.ReplaceAll(body, "\n", "_"), func(t *testing.T) {
			assembly := strings.Replace(x86CallerFixtureAssembly, "MOVL AX,(DX)", body, 1)
			goAssemblerOracleOutput(t, "amd64", assembly, "", true)
			pkg, _ := x86CallerCoreFixture(t, x86CallerFixtureSource, assembly)
			if err := x86CallerCoreTranslate(t, pkg, assembly, "x86_64-unknown-linux-gnu"); err != nil {
				t.Fatal(err)
			}
		})
	}
}
