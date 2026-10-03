package plan9asm

import (
	"fmt"
	"strings"
	"testing"
)

func x86CallerUnpackOpcodes() []string {
	return []string{
		"PUNPCKLBW", "PUNPCKHBW", "PUNPCKLWL", "PUNPCKHWL",
		"PUNPCKLLQ", "PUNPCKHLQ", "PUNPCKLQDQ", "PUNPCKHQDQ",
		"UNPCKLPS", "UNPCKHPS", "UNPCKLPD", "UNPCKHPD",
		"VPUNPCKLBW", "VPUNPCKHBW", "VPUNPCKLWD", "VPUNPCKHWD",
		"VPUNPCKLDQ", "VPUNPCKHDQ", "VPUNPCKLQDQ", "VPUNPCKHQDQ",
		"VUNPCKLPS", "VUNPCKHPS", "VUNPCKLPD", "VUNPCKHPD",
	}
}

func TestX86GoCallerUnpackCompleteFamilyOriginalSource(t *testing.T) {
	for _, opcode := range x86CallerUnpackOpcodes() {
		t.Run(opcode, func(t *testing.T) {
			operands, destination := "X0,X1", "X1"
			if strings.HasPrefix(opcode, "V") {
				operands, destination = "Y0,Y1,Y2", "X2"
			}
			body := fmt.Sprintf("VMOVUPS (DI),Y0\nVMOVUPS (SI),Y1\n%s %s\nVMOVSS %s,(DX)", opcode, operands, destination)
			assembly := strings.Replace(x86CallerFixtureAssembly, "MOVL AX,(DX)", body, 1)
			requireX86GoAssemblerResult(t, "amd64", assembly, true)
			pkg, _ := x86CallerCoreFixture(t, x86CallerFixtureSource, assembly)
			if err := x86CallerCoreTranslate(t, pkg, assembly, "x86_64-unknown-linux-gnu"); err != nil {
				t.Fatal(err)
			}
		})
	}
}
