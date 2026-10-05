package plan9asm

import (
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestTranslateGoModule_UsesDeclsAndLinkname(t *testing.T) {
	pkg := mustGoPackage(t, "test/pkg", `package testpkg
func Compare(a, b int) int
//go:linkname cmp runtime.cmp
func cmp(a, b int) int
`)

	asm := []byte(`TEXT ·Compare(SB),NOSPLIT,$0-24
	CALL runtime·cmp(SB)
	MOVD $0, R0
	RET
`)

	tr, err := TranslateGoModule(pkg, asm, GoModuleOptions{
		FileName:     "compare_arm64.s",
		GOARCH:       "arm64",
		TargetTriple: "aarch64-unknown-linux-gnu",
		ResolveSym:   testResolveSym("test/pkg"),
	})
	if err != nil {
		t.Fatal(err)
	}
	defer tr.Module.Dispose()

	if _, ok := tr.Signatures["test/pkg.Compare"]; !ok {
		t.Fatalf("missing package function signature")
	}
	if _, ok := tr.Signatures["runtime.cmp"]; !ok {
		t.Fatalf("missing go:linkname target signature")
	}
	if got := tr.Functions[0].ResolvedSymbol; got != "test/pkg.Compare" {
		t.Fatalf("resolved symbol: got %q", got)
	}
}

func TestTranslateGoModule_ExpandsGeneratedWasmOffsets(t *testing.T) {
	pkg := mustGoPackage(t, "runtime", `package runtime
type gobuf struct {
	sp   uintptr
	pc   uintptr
	g    uintptr
	ctxt uintptr
}
func gogo(buf *gobuf)
`)
	tr, err := TranslateGoModule(pkg, []byte(`#include "go_asm.h"
TEXT ·gogo(SB),NOSPLIT,$0-8
	MOVD buf+0(FP), R0
	MOVD gobuf_g(R0), R1
	RET
`), GoModuleOptions{
		FileName:     "asm_wasm.s",
		GOARCH:       "wasm",
		TargetTriple: "wasm32-unknown-unknown",
		ResolveSym:   testResolveSym("runtime"),
	})
	if err != nil {
		t.Fatal(err)
	}
	defer tr.Module.Dispose()
	ir := tr.Module.String()
	if !strings.Contains(ir, "getelementptr i8") || !strings.Contains(ir, "i32 16") ||
		!strings.Contains(ir, "load i64") {
		t.Fatalf("generated gobuf_g offset was not resolved through go_asm.h:\n%s", ir)
	}
}

func TestExpandGoAssemblySourcePreservesIncludedMacroDefinitionNames(t *testing.T) {
	pkg := mustGoPackage(t, "example.com/ffi", `package ffi
const type64 = 0
type frame struct { value uintptr }
func asmcall()
`)
	source := []byte(`// runtime has #include "go_asm.h"
#define const_type64 0x0
#define frame__size 8
TEXT ·asmcall(SB),NOSPLIT,$0-0
	CMPB AX, $const_type64
	MOVQ $frame__size, BX
	RET
`)
	expanded := ExpandGoAssemblySource(pkg, source, "amd64")
	if strings.Contains(string(expanded), "#define 0 ") {
		t.Fatalf("constant expansion rewrote the macro definition name:\n%s", expanded)
	}
	if !strings.Contains(string(expanded), "#define const_type64 0x0") {
		t.Fatalf("constant expansion did not preserve the macro definition:\n%s", expanded)
	}
	if !strings.Contains(string(expanded), "#define frame__size 8") {
		t.Fatalf("type-layout expansion did not preserve the macro definition:\n%s", expanded)
	}
	if _, err := ParseWithDefines(ArchAMD64, string(expanded), GoAssemblerDefines("linux", "amd64")); err != nil {
		t.Fatalf("ParseWithDefines(expanded) error = %v\n%s", err, expanded)
	}
}

func TestTranslateGoModule_UsesCallerWasmTypeSizes(t *testing.T) {
	pkg := mustGoPackage(t, "example.com/wasm32", `package wasm32
func Consume(value int, data []byte) int
`)
	tr, err := TranslateGoModule(pkg, []byte(`TEXT ·Consume(SB),NOSPLIT,$0-16
	I64Const $0
	I64Store ret+32(FP)
	RET
`), GoModuleOptions{
		FileName:     "consume_wasm.s",
		GOARCH:       "wasm",
		Sizes:        &types.StdSizes{WordSize: 4, MaxAlign: 4},
		FrameSizes:   &types.StdSizes{WordSize: 8, MaxAlign: 8},
		TargetTriple: "wasm32-unknown-unknown",
		ResolveSym:   testResolveSym("example.com/wasm32"),
	})
	if err != nil {
		t.Fatal(err)
	}
	defer tr.Module.Dispose()
	sig := tr.Signatures["example.com/wasm32.Consume"]
	want := []LLVMType{I32, "{ ptr, i32, i32 }"}
	if len(sig.Args) != len(want) {
		t.Fatalf("Consume args = %v, want %v", sig.Args, want)
	}
	for i := range want {
		if sig.Args[i] != want[i] {
			t.Fatalf("Consume arg %d = %s, want %s", i, sig.Args[i], want[i])
		}
	}
	if got := sig.Frame.Params[3].Offset; got != 24 {
		t.Fatalf("slice capacity FP offset = %d, want official Go wasm offset 24", got)
	}
	if got := sig.Frame.Results[0].Offset; got != 32 {
		t.Fatalf("result FP offset = %d, want official Go wasm offset 32", got)
	}
}

// A manual signature resolves a name and type. It does not provide the
// original Go declaration/caller evidence required by an ABI0 tail.
func TestTranslateGoModule_ManualSigCannotInventABI0TailDeclaration(t *testing.T) {
	pkg := mustGoPackage(t, "test/pkg", `package testpkg
func IndexByte(b []byte, c byte) int
`)
	manualCalled := false
	asm := []byte(`TEXT ·IndexByte(SB),NOSPLIT,$0-40
	B indexbytebody<>(SB)

TEXT indexbytebody<>(SB),NOSPLIT,$0
	MOVD $0, R0
	RET
`)

	tr, err := TranslateGoModule(pkg, asm, GoModuleOptions{
		FileName:     "indexbyte_arm64.s",
		GOARCH:       "arm64",
		TargetTriple: "aarch64-unknown-linux-gnu",
		ResolveSym:   testResolveSym("test/pkg"),
		ManualSig: func(resolved string) (FuncSig, bool) {
			if resolved != "test/pkg.indexbytebody" {
				return FuncSig{}, false
			}
			manualCalled = true
			return FuncSig{Name: resolved, Args: []LLVMType{"{ ptr, i64, i64 }", "i8"}, Ret: I64}, true
		},
	})
	if tr != nil {
		tr.Module.Dispose()
	}
	if !errors.Is(err, ErrProbeNeedsContext) || !strings.Contains(err.Error(), "actual selected bodyless declaration") {
		t.Fatalf("manual signature supplied an unbound ABI0 tail: %v", err)
	}
	if !manualCalled {
		t.Fatalf("manual signature callback was not used")
	}
}

func TestTranslateGoModule_UsesManualSigForPlainLocalHelper(t *testing.T) {
	pkg := mustGoPackage(t, "test/pkg", `package testpkg
func Call()
`)
	// This fixture isolates signature resolution with an ordinary no-argument
	// call. It does not invent slice transport or a result stored only in R0.
	asm := []byte(`TEXT ·Call(SB),4,$0-0
	CALL localbody(SB)
	RET

TEXT localbody(SB),4,$0-0
	MOVD $0, R0
	RET
`)
	llc := findLLVM22Tool("llc")
	if llc == "" {
		t.Fatal("LLVM 22 llc required")
	}
	for _, triple := range []string{
		"aarch64-unknown-linux-gnu", "aarch64-apple-darwin", "aarch64-pc-windows-msvc",
	} {
		t.Run(triple, func(t *testing.T) {
			requireManualHelperGoObject(t, triple, asm)
			manualCalled := false
			tr, err := TranslateGoModule(pkg, asm, GoModuleOptions{
				FileName:     "call_arm64.s",
				GOARCH:       "arm64",
				TargetTriple: triple,
				ResolveSym:   testResolveSym("test/pkg"),
				ManualSig: func(resolved string) (FuncSig, bool) {
					if resolved != "test/pkg.localbody" {
						return FuncSig{}, false
					}
					manualCalled = true
					return FuncSig{Name: resolved, Ret: Void}, true
				},
			})
			if err != nil {
				t.Fatal(err)
			}
			defer tr.Module.Dispose()
			if !manualCalled {
				t.Fatal("manual signature callback was not used")
			}
			sig, ok := tr.Signatures["test/pkg.localbody"]
			if !ok || len(sig.Args) != 0 || sig.Ret != Void {
				t.Fatalf("manual helper signature = %+v, present = %v", sig, ok)
			}
			compileLLVMToObject(t, llc, triple, "manual-helper.ll", "manual-helper.o", tr.Module.String())
		})
	}
}

func requireManualHelperGoObject(t *testing.T, triple string, source []byte) {
	t.Helper()
	goos := "linux"
	if strings.Contains(triple, "apple") {
		goos = "darwin"
	} else if strings.Contains(triple, "windows") {
		goos = "windows"
	}
	dir := t.TempDir()
	input := filepath.Join(dir, "call_arm64.s")
	object := filepath.Join(dir, "call.o")
	if err := os.WriteFile(input, source, 0600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := runtimeOracleStageContext(t)
	defer cancel()
	command := exec.CommandContext(ctx, "go", "tool", "asm", "-p", "test/pkg", "-o", object, input)
	command.Env = append(os.Environ(), "GOOS="+goos, "GOARCH=arm64", "CGO_ENABLED=0", "GOTOOLCHAIN=local", "GOENV=off", "GOWORK=off")
	if output, err := runtimeOracleCombinedOutput(ctx, command); err != nil {
		t.Fatalf("original Go local helper object: %v\n%s", err, output)
	} else if len(output) != 0 {
		t.Fatalf("unexpected original Go object diagnostic: %s", output)
	}
	info, err := os.Lstat(object)
	if err != nil || !info.Mode().IsRegular() || info.Size() == 0 {
		t.Fatalf("missing regular original Go object: %v", err)
	}
	data, err := os.ReadFile(object)
	if err != nil || !strings.HasPrefix(string(data), "go object "+goos+" arm64 ") {
		t.Fatalf("original Go object target mismatch: %v", err)
	}
}

func TestTranslateGoModule_UsesManualSigExternalName(t *testing.T) {
	pkg := mustGoPackage(t, "test/pkg", `package testpkg
func Call()
`)
	// This checks external-name resolution, not runtime.memmove's ABI. An
	// uninitialized pointer call cannot prove that the saved caller LR survives.
	asm := []byte(`TEXT ·Call(SB),NOSPLIT,$0-0
	CALL external·namedSink(SB)
	RET
`)

	for _, goarch := range []string{"amd64", "arm64"} {
		t.Run(goarch, func(t *testing.T) {
			tr, err := TranslateGoModule(pkg, asm, GoModuleOptions{
				FileName:   "call_" + goarch + ".s",
				GOARCH:     goarch,
				ResolveSym: testResolveSym("test/pkg"),
				ManualSig: func(resolved string) (FuncSig, bool) {
					if resolved != "external.namedSink" {
						return FuncSig{}, false
					}
					return FuncSig{Name: "namedSink", Ret: Void}, true
				},
			})
			if err != nil {
				t.Fatal(err)
			}
			defer tr.Module.Dispose()
			ir := tr.Module.String()
			if !strings.Contains(ir, "@namedSink") || strings.Contains(ir, "external.namedSink") {
				t.Fatalf("manual external name not applied:\n%s", ir)
			}
		})
	}
}

func TestTranslateGoModule_X87Mode(t *testing.T) {
	pkg := mustGoPackage(t, "test/pkg", `package testpkg
func Round(x float64) int64
`)
	asm := []byte(`TEXT ·Round(SB),NOSPLIT,$0-16
	FMOVD x+0(FP), F0
	FRNDINT
	FMOVVP F0, ret+8(FP)
	RET
`)

	tr, err := TranslateGoModule(pkg, asm, GoModuleOptions{
		FileName:     "round_386.s",
		GOARCH:       "386",
		TargetTriple: "i386-unknown-linux-gnu",
		X87Mode:      X87Software,
		ResolveSym:   testResolveSym("test/pkg"),
	})
	if err != nil {
		t.Fatal(err)
	}
	defer tr.Module.Dispose()

	ir := tr.Module.String()
	if !strings.Contains(ir, "@llvm.floor.f64") || strings.Contains(ir, "frndint") {
		t.Fatalf("software x87 mode was not propagated:\n%s", ir)
	}
}

func mustGoPackage(t *testing.T, pkgPath, src string) GoPackage {
	t.Helper()
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "pkg.go", src, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}
	conf := types.Config{}
	pkg, err := conf.Check(pkgPath, fset, []*ast.File{f}, nil)
	if err != nil {
		t.Fatal(err)
	}
	return GoPackage{Path: pkgPath, Types: pkg, Syntax: []*ast.File{f}}
}

func testResolveSym(pkgPath string) func(string) string {
	return func(sym string) string {
		sym = goStripABISuffix(sym)
		if strings.HasPrefix(sym, "·") {
			return pkgPath + "." + strings.TrimPrefix(sym, "·")
		}
		if !strings.Contains(sym, "·") && !strings.Contains(sym, ".") && !strings.Contains(sym, "/") {
			return pkgPath + "." + sym
		}
		sym = strings.ReplaceAll(sym, "∕", "/")
		return strings.ReplaceAll(sym, "·", ".")
	}
}
