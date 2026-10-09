package plan9asm

import (
	"strings"
	"testing"

	"github.com/xgo-dev/llvm"
)

// This directory is independent of the typed effect descriptor. It exercises
// all explicit yimul/yimul3 source classes; one-operand IMUL is not in it.
var x86CallerExplicitIMULRows = []struct {
	name, instruction string
	three             bool
}{
	{"register", "IMUL@ AX,BX", false},
	{"memory", "IMUL@ (DI),BX", false},
	{"immediate8", "IMUL@ $-7,BX", false},
	{"immediate32", "IMUL@ $65537,BX", false},
	{"three-register8", "IMUL3@ $-7,AX,BX", true},
	{"three-register32", "IMUL3@ $65537,AX,BX", true},
	{"three-memory8", "IMUL3@ $-7,(DI),BX", true},
	{"three-memory32", "IMUL3@ $65537,(DI),BX", true},
}

func x86CallerExplicitIMULFixture(t *testing.T, body string) (GoPackage, *File, string, Options) {
	t.Helper()
	declaration := strings.ReplaceAll(x86CallerFixtureSource, "float32", "uint64")
	assembly := strings.Replace(x86CallerFixtureAssembly, "MOVL $7,AX\nMOVL AX,(DX)", body+"\nMOVQ BX,(DX)", 1)
	goAssemblerOracleOutput(t, "amd64", assembly, "", true)
	pkg, file := x86CallerCoreFixture(t, declaration, assembly)
	resolve := func(symbol string) string { return strings.Replace(symbol, "·", "caller.", 1) }
	sigs, err := goSigsForAsmFile(pkg, file, resolve, "amd64", nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	sigs, err = BindGoX86RawReturnCallerRegions(pkg, file, "amd64", resolve, sigs)
	if err != nil || sigs["caller.kernel"].x86CallerRegions == nil {
		t.Fatalf("actual complete typed caller precondition failed: %v", err)
	}
	return pkg, file, assembly, Options{Goarch: "amd64", TargetTriple: "x86_64-unknown-linux-gnu", ResolveSym: resolve, Sigs: sigs}
}

func TestX86GoCallerExplicitIMULRawReturnRegisteredForms(t *testing.T) {
	llc := findLLVM22Tool("llc")
	if llc == "" {
		t.Fatal("LLVM22 llc required")
	}
	for _, suffix := range []string{"W", "L", "Q"} {
		for _, row := range x86CallerExplicitIMULRows {
			t.Run(suffix+"/"+row.name, func(t *testing.T) {
				body := "MOVQ $7,AX\n"
				// W preserves the old upper bytes. Two-operand IMUL also reads
				// the old low result; IMUL3L/Q must not require that old result.
				if suffix == "W" || !row.three {
					body += "MOVQ $3,BX\n"
				}
				body += strings.ReplaceAll(row.instruction, "@", suffix)
				pkg, file, assembly, options := x86CallerExplicitIMULFixture(t, body)
				for _, triple := range []string{"x86_64-apple-darwin", "x86_64-unknown-linux-gnu", "x86_64-pc-windows-msvc"} {
					options.TargetTriple = triple
					t.Run(triple+"/text", func(t *testing.T) {
						ir, err := Translate(file, options)
						if err != nil {
							t.Fatalf("Go-accepted explicit multiply lacks source/caller effect: %v", err)
						}
						compileLLVMToObject(t, llc, triple, "imul-caller.ll", "imul-caller.o", ir)
					})
					t.Run(triple+"/module", func(t *testing.T) {
						ctx := llvm.NewContext()
						defer ctx.Dispose()
						module, err := TranslateModuleInContext(ctx, file, options)
						if !module.IsNil() {
							defer module.Dispose()
						}
						if err != nil {
							t.Fatal(err)
						}
						compileLLVMToObject(t, llc, triple, "imul-caller-module.ll", "imul-caller-module.o", module.String())
					})
					t.Run(triple+"/go", func(t *testing.T) {
						ctx := llvm.NewContext()
						defer ctx.Dispose()
						translated, err := translateGoModuleInContext(ctx, pkg, []byte(assembly), GoModuleOptions{
							FileName: "kernel_amd64.s", GOARCH: "amd64", TargetTriple: triple, ResolveSym: options.ResolveSym,
						})
						if translated != nil && !translated.Module.IsNil() {
							defer translated.Module.Dispose()
						}
						if err != nil {
							t.Fatal(err)
						}
						compileLLVMToObject(t, llc, triple, "imul-caller-go.ll", "imul-caller-go.o", translated.Module.String())
					})
				}
			})
		}
	}
}
