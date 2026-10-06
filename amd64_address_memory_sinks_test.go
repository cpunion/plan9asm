package plan9asm

import (
	"fmt"
	"strings"
	"testing"
)

func TestX86SymbolMemoryPointerRejectsAddressIdentityBeforeEmission(t *testing.T) {
	for _, source := range []string{
		"$·cell(SB)", "$·cell+8(SB)", " \t$·cell-8(SB)",
	} {
		t.Run(source, func(t *testing.T) {
			ctx, output := newAMD64CtxWithFuncForTest(t, Func{}, FuncSig{Ret: Void}, nil)
			before, temporary := output.String(), ctx.tmp
			pointer, err := ctx.ptrFromSB(source)
			if err == nil || pointer != "" || !strings.Contains(err.Error(), "address immediate") {
				t.Fatalf("address acquired a memory pointer: %q: %v", pointer, err)
			}
			if output.String() != before || ctx.tmp != temporary {
				t.Fatal("rejected address emitted IR or consumed a temporary")
			}
		})
	}

	for _, source := range []string{"·cell(SB)", "·cell+8(SB)", "·cell-8(SB)"} {
		t.Run("memory/"+source, func(t *testing.T) {
			ctx, _ := newAMD64CtxWithFuncForTest(t, Func{}, FuncSig{Ret: Void}, nil)
			if pointer, err := ctx.ptrFromSB(source); err != nil || pointer == "" {
				t.Fatalf("ordinary symbolic memory lost: %q: %v", pointer, err)
			}
		})
	}
}

func x86AddressMemorySinkRows(arch string) []string {
	// These ordinary stores and x87 loads use direct OpSym dispatch, not the
	// shared memory-class predicate. Their Go encoder rows still require actual
	// memory, including the complete scalar and x87 storage widths below.
	rows := []string{
		"MOVB AL,%s", "MOVW CX,%s", "MOVL CX,%s",
		"MOVSS X0,%s", "MOVSD X0,%s", "MOVOU X0,%s",
		"FMOVF F0,%s", "FMOVFP F0,%s", "FMOVD F0,%s", "FMOVDP F0,%s",
		"FMOVW F0,%s", "FMOVWP F0,%s", "FMOVL F0,%s", "FMOVLP F0,%s",
		"FMOVVP F0,%s", "FMOVBP F0,%s", "FMOVXP F0,%s",
		"FMOVF %s,F0", "FMOVD %s,F0", "FMOVW %s,F0", "FMOVL %s,F0",
		"FMOVV %s,F0", "FMOVB %s,F0", "FMOVX %s,F0",
	}
	if arch == "amd64" {
		rows = append(rows, "MOVQ CX,%s")
	}
	return rows
}

func x86AddressMemorySinkSource(row, operand string) string {
	// These are compile-only syntax/operand tests. Initialize explicit register
	// inputs, but do not claim execution of the external cell or x87 state.
	return "TEXT ·Value(SB),4,$0-0\nMOVL $7,AX\nMOVL $9,CX\n" +
		"PXOR X0,X0\nFLDZ\n" + fmt.Sprintf(row, operand) + "\nRET\n"
}

func TestX86AddressImmediateDirectMemorySinksRejectAcrossAPIs(t *testing.T) {
	for _, arch := range []string{"386", "amd64"} {
		for _, row := range x86AddressMemorySinkRows(arch) {
			for _, operand := range []string{"$·cell(SB)", "$·cell+8(SB)"} {
				t.Run(arch+"/"+fmt.Sprintf(row, operand), func(t *testing.T) {
					ordinary := x86AddressMemorySinkSource(row, strings.TrimPrefix(operand, "$"))
					goAssemblerOracleOutput(t, arch, ordinary, "", true)
					source := x86AddressMemorySinkSource(row, operand)
					goAssemblerOracleOutput(t, arch, source, "", false)
					for _, target := range x86AddressMemoryTargets(arch) {
						for _, api := range []string{"text", "module", "go"} {
							t.Run(target+"/"+api, func(t *testing.T) {
								if _, err := x86AddressMemoryAPI(t, api, source, arch, target); err == nil {
									t.Fatal("Go-rejected address bypassed memory-role validation")
								}
							})
						}
					}
				})
			}
		}
	}
}

func TestX86DirectSymbolMemorySinkObjectsRemainSupported(t *testing.T) {
	llc := findLLVM22Tool("llc")
	if llc == "" {
		t.Fatal("LLVM22 llc required")
	}
	for _, arch := range []string{"386", "amd64"} {
		for _, row := range x86AddressMemorySinkRows(arch) {
			source := x86AddressMemorySinkSource(row, "·cell+8(SB)")
			goAssemblerOracleOutput(t, arch, source, "", true)
			for _, target := range x86AddressMemoryTargets(arch) {
				for _, api := range []string{"text", "module", "go"} {
					t.Run(arch+"/"+row+"/"+target+"/"+api, func(t *testing.T) {
						ir, err := x86AddressMemoryAPI(t, api, source, arch, target)
						if err != nil {
							t.Fatal("ordinary memory sink lost:", err)
						}
						x86ResultInitializationObject(t, llc, target, ir, "memory-sink")
					})
				}
			}
		}
	}
}
