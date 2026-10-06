package plan9asm

import (
	"fmt"
	"strings"
	"testing"

	"github.com/xgo-dev/llvm"
)

func TestX86PackedSubtractLowZThreeAPIFiveTargetObjects(t *testing.T) {
	llc := findLLVM22Tool("llc")
	if llc == "" {
		t.Fatal("LLVM 22 llc required")
	}
	var parameters []string
	for _, prefix := range []string{"a", "b"} {
		for index := 0; index < 8; index++ {
			parameters = append(parameters, fmt.Sprintf("%s%d uint64", prefix, index))
		}
	}
	declaration := "func Value(" + strings.Join(parameters, ",") + ") uint64"
	for _, target := range []struct{ arch, triple string }{
		{"amd64", "x86_64-apple-darwin"},
		{"amd64", "x86_64-unknown-linux-gnu"},
		{"amd64", "x86_64-pc-windows-msvc"},
		{"386", "i386-unknown-linux-gnu"},
		{"386", "i686-pc-windows-msvc"},
	} {
		for _, opcode := range []string{"VPSUBB", "VPSUBW", "VPSUBD", "VPSUBQ", "VPSUBSB", "VPSUBSW", "VPSUBUSB", "VPSUBUSW"} {
			forms := []struct{ name, instruction string }{
				{"register", opcode + " Z0,Z1,Z2"},
				{"alias", opcode + " Z1,Z1,Z2"},
				{"memory", opcode + " a0+0(FP),Z1,Z2"},
			}
			if opcode == "VPSUBD" || opcode == "VPSUBQ" {
				forms = append(forms, struct{ name, instruction string }{
					"broadcast", opcode + ".BCST a0+0(FP),Z1,Z2",
				})
			}
			for _, form := range forms {
				t.Run(target.triple+"/"+opcode+"/"+form.name, func(t *testing.T) {
					body := "VMOVUPS a0+0(FP),Z0\nVMOVUPS b0+64(FP),Z1\n" +
						form.instruction + "\nMOVQ X2,ret0+128(FP)\nRET"
					pkg, file, sig, source := x86ResultInitializationFixture(t, target.arch, declaration, body)
					goAssemblerOracleOutput(t, target.arch, source, "", true)
					if err := CheckX86ABI0ResultInitialization(file.Funcs[0], sig, target.arch); err != nil {
						t.Fatal("declared full-width inputs lost initialized low-Z arithmetic", err)
					}
					options := Options{Goarch: target.arch, TargetTriple: target.triple,
						Sigs:       map[string]FuncSig{"Value": sig},
						ResolveSym: func(symbol string) string { return strings.TrimPrefix(symbol, "·") }}
					before := fmt.Sprintf("%#v %#v", file, sig)
					ir, err := Translate(file, options)
					if err != nil {
						t.Fatal(err)
					}
					x86ResultInitializationObject(t, llc, target.triple, ir, "text")
					context := llvm.NewContext()
					defer context.Dispose()
					module, err := TranslateModuleInContext(context, file, options)
					if !module.IsNil() {
						defer module.Dispose()
					}
					if err != nil {
						t.Fatal(err)
					}
					x86ResultInitializationObject(t, llc, target.triple, module.String(), "module")
					translated, err := translateGoModuleInContext(context, pkg, []byte(source), GoModuleOptions{
						GOARCH: target.arch, TargetTriple: target.triple, ResolveSym: options.ResolveSym,
					})
					if translated != nil && !translated.Module.IsNil() {
						defer translated.Module.Dispose()
					}
					if err != nil {
						t.Fatal(err)
					}
					x86ResultInitializationObject(t, llc, target.triple, translated.Module.String(), "typed-go")
					if before != fmt.Sprintf("%#v %#v", file, sig) {
						t.Fatal("translation mutated original typed inputs")
					}
				})
			}
		}
	}
}
