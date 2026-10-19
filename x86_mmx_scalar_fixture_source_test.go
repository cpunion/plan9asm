package plan9asm

import "testing"

const x86LegacyPackedLogicalMMXRuntimeSource = `TEXT legacypackedlogicalsemantics(SB),$0-24
	MOVQ out+0(FP), AX
	MOVQ a+8(FP), M0
	MOVQ b+16(FP), M1
	STC
	PAND M0, M1
	MOVQ M1, 0(AX)
	MOVQ a+8(FP), M0
	MOVQ b+16(FP), M1
	PANDN M0, M1
	MOVQ M1, 8(AX)
	MOVQ a+8(FP), M0
	MOVQ b+16(FP), M1
	POR M0, M1
	MOVQ M1, 16(AX)
	MOVQ a+8(FP), M0
	MOVQ b+16(FP), M1
	PXOR M0, M1
	MOVQ M1, 24(AX)
	SETCS 32(AX)
	EMMS
	RET
`

const x86PADDDGoAliasRuntimeSource = `TEXT padddaliassemantics(SB),$0-24
	MOVQ out+0(FP), AX
	MOVQ a+8(FP), M0
	MOVQ b+16(FP), M1
	STC
	PADDD M0, M1
	MOVQ M1, 0(AX)
	SETCS 8(AX)
	EMMS
	RET
`

func TestX86MMXOriginalScalarLogicalAndPADDDObjects(t *testing.T) {
	for _, fixture := range []struct{ name, source, declaration string }{
		{"logical", x86LegacyPackedLogicalMMXRuntimeSource, "func legacypackedlogicalsemantics(out *byte,a,b uint64)"},
		{"paddd", x86PADDDGoAliasRuntimeSource, "func padddaliassemantics(out *byte,a,b uint64)"},
	} {
		t.Run(fixture.name, func(t *testing.T) {
			x86MMXOriginalFixtureObjects(t, fixture.source, fixture.declaration)
		})
	}
}
