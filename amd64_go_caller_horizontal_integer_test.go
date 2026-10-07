package plan9asm

import (
	"strings"
	"testing"

	"github.com/xgo-dev/llvm"
)

// This first-party caller closes its own source/ABI boundary. The horizontal
// operation has no exemption from the original raw-return stack certificate.
func TestX86GoCallerHorizontalIntegerOriginalSourceThreeAPIs(t *testing.T) {
	for _, body := range []string{
		"VMOVDQU (DI),Y0\nVMOVDQU (SI),Y1\nVPHADDD Y0,Y1,Y2\nVMOVSS X2,(DX)",
		"VMOVDQU (SI),Y1\nVPHADDD (DI),Y1,Y2\nVMOVSS X2,(DX)",
	} {
		t.Run(strings.Split(body, "\n")[1], func(t *testing.T) {
			assembly := strings.Replace(x86CallerFixtureAssembly, "MOVL AX,(DX)", body, 1)
			goAssemblerOracleOutput(t, "amd64", assembly, "", true)
			pkg, file := x86CallerCoreFixture(t, x86CallerFixtureSource, assembly)
			resolve := func(symbol string) string { return strings.Replace(symbol, "·", "caller.", 1) }
			sigs, err := goSigsForAsmFile(pkg, file, resolve, "amd64", nil, nil, nil)
			if err != nil {
				t.Fatal(err)
			}
			options := Options{Goarch: "amd64", TargetTriple: "x86_64-unknown-linux-gnu", Sigs: sigs, ResolveSym: resolve}
			if _, err := Translate(file, options); err != nil {
				t.Error("text", err)
			}
			ctx := llvm.NewContext()
			defer ctx.Dispose()
			module, err := TranslateModuleInContext(ctx, file, options)
			if !module.IsNil() {
				defer module.Dispose()
			}
			if err != nil {
				t.Error("module", err)
			}
			if err := x86CallerCoreTranslate(t, pkg, assembly, options.TargetTriple); err != nil {
				t.Error("selected Go caller", err)
			}
		})
	}
}
