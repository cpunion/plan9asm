package plan9asm

import (
	"fmt"
	"strings"
	"testing"

	"golang.org/x/arch/x86/x86asm"
)

func x86MOVSXDBytes(encoding []byte) string {
	var source strings.Builder
	for _, value := range encoding {
		fmt.Fprintf(&source, "BYTE $0x%02x\n", value)
	}
	return source.String()
}

func TestX86RawMOVSXDNativeWidths(t *testing.T) {
	for _, test := range []struct {
		name     string
		encoding []byte
		want     string
	}{
		{"long register", []byte{0x63, 0xc3}, "MOVL BX, AX"},
		{"long high source", []byte{0x41, 0x63, 0xc1}, "MOVL R9, AX"},
		{"long high result", []byte{0x44, 0x63, 0xc3}, "MOVL BX, R8"},
		{"long alias", []byte{0x63, 0xc0}, "MOVL AX, AX"},
		{"long memory", []byte{0x63, 0x07}, "MOVL 0(DI), AX"},
		{"long indexed", []byte{0x63, 0x44, 0x8f, 0x03}, "MOVL 0x3(DI)(CX*4), AX"},
		{"word register", []byte{0x66, 0x63, 0xc3}, "MOVW BX, AX"},
		{"word physical SI", []byte{0x66, 0x63, 0xc6}, "MOVW SI, AX"},
		{"word high source", []byte{0x66, 0x41, 0x63, 0xc1}, "MOVW R9, AX"},
		{"word high result", []byte{0x66, 0x44, 0x63, 0xc3}, "MOVW BX, R8"},
		{"quad register", []byte{0x48, 0x63, 0xc3}, "MOVLQSX BX, AX"},
		{"quad memory", []byte{0x48, 0x63, 0x07}, "MOVLQSX 0(DI), AX"},
	} {
		t.Run(test.name, func(t *testing.T) {
			inst, err := x86asm.Decode(test.encoding, 64)
			if err != nil || inst.Op != x86asm.MOVSXD || inst.Len != len(test.encoding) {
				t.Fatalf("physical decode: %+v %v", inst, err)
			}
			got, err := decodedX86GoSyntax(inst, test.encoding)
			if err != nil || got != test.want {
				t.Fatalf("physical width: got %q, want %q: %v", got, test.want, err)
			}
		})
	}
}

func TestX86RawMOVSXDLongCallerOriginalBytes(t *testing.T) {
	assembly := strings.Replace(x86CallerFixtureAssembly, "MOVL $7,AX\nMOVL AX,(DX)",
		"MOVQ $0x80000000,BX\nWORD $0xc363\nMOVL AX,(DX)", 1)
	requireX86GoAssemblerResult(t, "amd64", assembly, true)
	pkg, _ := x86CallerCoreFixture(t, x86CallerFixtureSource, assembly)
	if err := x86CallerCoreTranslate(t, pkg, assembly, "x86_64-unknown-linux-gnu"); err != nil {
		t.Fatal(err)
	}
}
