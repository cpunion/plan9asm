package plan9asm

import (
	"strings"
	"testing"
)

func TestTranslateARM64RawSVEFloatImmediateGoHighwayRegression(t *testing.T) {
	const source = `
TEXT rawSVEFloatImmediate(SB),$0-0
	WORD $0x2579cc00 // FMOV Z0.H, #0.5
	WORD $0x25b9cc1b // FMOV Z27.S, #0.5
	WORD $0x25f9ce1f // FMOV Z31.D, #1.0
	RET
`
	requireARM64GoAssemblerResult(t, source, true)
	file, err := Parse(ArchARM64, source)
	if err != nil {
		t.Fatal(err)
	}
	llc := findLLVM22Tool("llc")
	if llc == "" {
		t.Fatal("LLVM 22 llc not found")
	}
	for _, triple := range []string{
		"aarch64-apple-darwin",
		"aarch64-unknown-linux-gnu",
		"aarch64-pc-windows-msvc",
	} {
		t.Run(triple, func(t *testing.T) {
			ir, err := Translate(file, Options{
				TargetTriple: triple,
				Goarch:       "arm64",
				Sigs: map[string]FuncSig{
					"rawSVEFloatImmediate": {Name: "rawSVEFloatImmediate", Ret: Void},
				},
			})
			if err != nil {
				t.Fatal(err)
			}
			for _, want := range []string{"<vscale x 8 x half>", "<vscale x 4 x float>", "<vscale x 2 x double>"} {
				if !strings.Contains(ir, want) {
					t.Fatalf("SVE FMOV immediate omitted %q:\n%s", want, ir)
				}
			}
			compileLLVMToObject(t, llc, triple, "arm64-raw-sve-fmov.ll", "arm64-raw-sve-fmov.o", ir)
		})
	}
}
