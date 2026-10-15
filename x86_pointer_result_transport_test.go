package plan9asm

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

// Retain the three complete TEXT roles from the standard-library FIPS pointer
// check: one local code definition and two Go-declared opaque pointer results.
// Register aliases and copies exercise the same complete native-word grammar.
func x86PointerResultTransportSource(arch string) (string, string) {
	word, move, moves := 8, "MOVQ", []string{"MOVQ", "MOVD", "MOVDQ2Q"}
	if arch == "386" {
		word, move, moves = 4, "MOVL", []string{"MOVL"}
	}
	var source, declarations strings.Builder
	source.WriteString("DATA StaticData<>(SB)/4,$10\nGLOBL StaticData<>(SB),16,$4\n")
	source.WriteString("TEXT StaticText<>(SB),$0\nRET\n")
	for _, opcode := range moves {
		for _, kind := range []string{"Data", "Text"} {
			for _, copied := range []bool{false, true} {
				name := fmt.Sprintf("Pointer%s%sCopy%t", opcode, kind, copied)
				fmt.Fprintf(&source, "TEXT ·%s(SB),4,$0-%d\n%s $Static%s<>(SB),AX\n", name, word, opcode, kind)
				register := "AX"
				if copied {
					fmt.Fprintf(&source, "%s AX,BX\n", move)
					register = "BX"
				}
				fmt.Fprintf(&source, "%s %s,ret+0(FP)\nRET\n", move, register)
				fmt.Fprintf(&declarations, "func %s() *byte\n", name)
			}
		}
	}
	return source.String(), declarations.String()
}

func x86PointerResultTransportAPI(t *testing.T, source, declarations, arch, triple, api string) (string, error) {
	t.Helper()
	pkg := mustGoPackage(t, "bitview", "package bitview\n"+declarations)
	// A file-local TEXT does not imply a signature. This exact fixture owns
	// the complete zero-argument RET body explicitly; do not infer a helper
	// ABI from another function's Go declaration or remove the original TEXT.
	manual := func(name string) (FuncSig, bool) {
		return FuncSig{Name: name, Ret: Void}, name == "StaticText"
	}
	return x86ScalarFPBitViewPackageAPIWithManual(t, source, pkg, arch, triple, api, manual)
}

func TestX86OpaquePointerResultAddressFamilyCompleteObjects(t *testing.T) {
	llc := findLLVM22Tool("llc")
	if llc == "" {
		t.Fatal("LLVM 22 llc is required")
	}
	for _, target := range dataRelocationTargets {
		if target.arch != ArchAMD64 {
			continue
		}
		t.Run(target.triple, func(t *testing.T) {
			source, declarations := x86PointerResultTransportSource(target.goarch)
			dataRelocationGoObject(t, source, target.goos, target.goarch, true)
			for _, api := range []string{"text", "module", "Go"} {
				t.Run(api, func(t *testing.T) {
					ir, err := x86PointerResultTransportAPI(t, source, declarations, target.goarch, target.triple, api)
					if err != nil || ir == "" {
						t.Fatal("complete code/data address lost its existing opaque pointer transport", err)
					}
					for _, optimized := range []bool{false, true} {
						objectIR := ir
						if optimized {
							objectIR = x86DwordProductIROptimize(t, ir)
						}
						for _, level := range []string{"-O0", "-O2"} {
							t.Run(fmt.Sprintf("IR-O2=%t/backend%s", optimized, level), func(t *testing.T) {
								x86PackedStringCompileObject(t, llc, target.triple, objectIR, level)
							})
						}
					}
				})
			}
		})
	}
}

func TestX86OpaquePointerAddressCannotBorrowNumericalOrPartialStorage(t *testing.T) {
	for _, arch := range []string{"386", "amd64"} {
		word, move, triple := 8, "MOVQ", "x86_64-unknown-linux-gnu"
		if arch == "386" {
			word, move, triple = 4, "MOVL", "i386-unknown-linux-gnu"
		}
		for _, row := range []struct {
			name, declaration, body string
		}{
			{"code-is-not-a-number", "func Value() uintptr", move + " $StaticText<>(SB),AX\n" + move + " AX,ret+0(FP)"},
			{"missing-result", "func Value() *byte", move + " $StaticData<>(SB),AX"},
			{"partial-result", "func Value() *byte", move + " $StaticData<>(SB),AX\nMOVB AX,ret+0(FP)"},
			{"unknown-copy", "func Value() *byte", move + " $StaticData<>(SB),AX\n" + move + " BX,ret+0(FP)"},
			{"no-pointee-permission", "func Value() *byte", move + " $StaticData<>(SB),AX\nMOVB (AX),CX\n" + move + " AX,ret+0(FP)"},
		} {
			for _, api := range []string{"text", "module", "Go"} {
				t.Run(arch+"/"+row.name+"/"+api, func(t *testing.T) {
					source := "DATA StaticData<>(SB)/4,$10\nGLOBL StaticData<>(SB),16,$4\n" +
						"TEXT StaticText<>(SB),$0\nRET\n" +
						fmt.Sprintf("TEXT ·Value(SB),4,$0-%d\n%s\nRET\n", word, row.body)
					goAssemblerOracleOutput(t, arch, source, "", true)
					ir, err := x86PointerResultTransportAPI(t, source, row.declaration, arch, triple, api)
					if !errors.Is(err, ErrProbeNeedsContext) || ir != "" {
						t.Fatal("opaque pointer transport acquired numerical, partial or native permissions", err)
					}
				})
			}
		}
	}
}
