package plan9asm

import (
	"fmt"
	"strings"
	"testing"

	"github.com/xgo-dev/llvm"
)

func TestX86UnnamedAddressImmediateGSourceModeActualGo(t *testing.T) {
	for _, arch := range []string{"386", "amd64"} {
		for _, annotation := range []string{"g", "R14"} {
			for _, address := range []string{
				"1(" + annotation + ")",
				"1(" + annotation + "*1)",
				"1(AX)(" + annotation + "*2)",
				"1(" + annotation + ")(CX)",
			} {
				t.Run(arch+"/"+address, func(t *testing.T) {
					source := "TEXT addr(SB),4,$0-0\nADDL $" + address + ",DX\nRET\n"
					accepted := arch == "amd64" || annotation == "R14"
					x86ScalarLiteralExpectedGoListing(t, arch, source, accepted)
				})
			}
		}
	}
}

func TestX86UnnamedAddressImmediateGSourceModeThreeAPIs(t *testing.T) {
	for _, target := range x86RawReturnTargets {
		for _, statement := range []string{
			"MOVW $1(g),DX",
			"ADDL $1(g),DX",
			"ADDL $1(AX)(g*2),DX",
			"CMPL DX,$1(g)(CX)",
		} {
			t.Run(target.triple+"/"+statement, func(t *testing.T) {
				source := "TEXT ·addr(SB),4,$0-0\nMOVL $2,DX\n" + statement + "\nRET\n"
				file, err := Parse(ArchAMD64, source)
				if err != nil {
					t.Fatal(err)
				}
				pkg := x86AddSubLiteralPackage(t, target.arch, "func addr()\n")
				resolve := func(symbol string) string { return strings.Replace(symbol, "·", "literal.", 1) }
				sigs, err := goSigsForAsmFile(pkg, file, resolve, target.arch, nil, nil, nil)
				if err != nil {
					t.Fatal(err)
				}
				options := Options{Goarch: target.arch, TargetTriple: target.triple, ResolveSym: resolve, Sigs: sigs}
				for _, api := range []string{"text", "module", "GoModule"} {
					t.Run(api, func(t *testing.T) {
						context := llvm.NewContext()
						defer context.Dispose()
						var translationErr error
						switch api {
						case "text":
							_, translationErr = Translate(file, options)
						case "module":
							var module llvm.Module
							module, translationErr = TranslateModuleInContext(context, file, options)
							if !module.IsNil() {
								module.Dispose()
							}
						case "GoModule":
							var result *GoModuleTranslation
							result, translationErr = translateGoModuleInContext(context, pkg, []byte(source), GoModuleOptions{GOARCH: target.arch, TargetTriple: target.triple, ResolveSym: resolve})
							if result != nil && !result.Module.IsNil() {
								result.Module.Dispose()
							}
						}
						if target.arch == "386" && translationErr == nil {
							t.Error(fmt.Sprintf("%s accepted unregistered source g after ignored R14 normalization", api))
						}
						if target.arch == "amd64" && translationErr != nil {
							t.Fatalf("actual registered g constant annotation: %v", translationErr)
						}
					})
				}
			})
		}
	}
}
