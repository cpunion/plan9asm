package plan9asm

import (
	"fmt"
	"strings"
	"testing"
)

func TestX86StatusFlagMaskClearsExactlyItsPhysicalSubset(t *testing.T) {
	// Independent architectural bit positions, not the production slot list.
	flags := []struct {
		bit  uint
		slot string
	}{
		{0, "%flags_cf"}, {2, "%flags_pf"}, {4, "%flags_af"},
		{6, "%flags_z"}, {7, "%flags_slt"}, {11, "%flags_of"},
	}
	for _, arch := range []string{"386", "amd64"} {
		for subset := 0; subset < 64; subset++ {
			t.Run(fmt.Sprintf("%s/subset=%02x", arch, subset), func(t *testing.T) {
				ctx, output := newAMD64CtxWithFuncForTest(t, Func{}, FuncSig{Ret: Void}, nil)
				ctx.goarch = arch
				output.Reset()
				temporary := ctx.tmp
				var mask x86StatusFlagMask
				var want strings.Builder
				for index, flag := range flags {
					if subset&(1<<uint(index)) != 0 {
						mask |= 1 << flag.bit
						fmt.Fprintf(&want, "  store i1 false, ptr %s\n", flag.slot)
					}
				}
				ctx.clearX86StatusFlags(mask)
				if output.String() != want.String() || ctx.tmp != temporary {
					t.Fatal("status subset changed another bit, allocated a value, or lost a selected bit",
						output.String(), want.String())
				}
			})
		}
	}
}

// Each combined object contains the complete affected family. This avoids one
// compiler process per spelling while retaining all Go sources and three APIs.
// These are object proofs, not execution of WAITPKG/random/enqueue instructions.
func TestX86DefinedStatusThreeAPIsFiveTargetsObjects(t *testing.T) {
	llc := findLLVM22Tool("llc")
	if llc == "" {
		t.Fatal("required LLVM 22 llc missing")
	}
	for _, target := range []struct{ arch, triple string }{
		{"amd64", "x86_64-apple-darwin"},
		{"amd64", "x86_64-unknown-linux-gnu"},
		{"amd64", "x86_64-pc-windows-msvc"},
		{"386", "i386-unknown-linux-gnu"},
		{"386", "i686-pc-windows-msvc"},
	} {
		var declarations, assembly strings.Builder
		word, valueType, valueBytes := "Q", "uint64", 8
		if target.arch == "386" {
			word, valueType, valueBytes = "L", "uint32", 4
		}
		writers := x86DefinedZeroAFWriters()
		for _, op := range []string{"COMISS", "COMISD", "UCOMISS", "UCOMISD", "VCOMISS", "VCOMISD", "VUCOMISS", "VUCOMISD"} {
			writers = append(writers, x86DefinedZeroAFCase{op, op + " X0,X1", false})
		}
		count := 0
		for _, row := range writers {
			if target.arch == "386" && row.amd64Only {
				continue // Absent 64-bit Go encoding, not a skipped target.
			}
			name := fmt.Sprintf("status%d", count)
			fmt.Fprintf(&declarations, "func %s(seed %s) %s\n", name, valueType, valueType)
			// Initialize each instruction's actual complete inputs. MOVL to X
			// clears X128; KMOVW zero-extends K. Both precede the FLAGS seed.
			setup := ""
			if strings.Contains(row.instruction, "X0") {
				setup += "MOVL AX,X0\nMOVL AX,X1\n"
			}
			if strings.Contains(row.instruction, "K1") {
				setup += "KMOVW AX,K1\nKMOVW AX,K2\n"
			}
			fmt.Fprintf(&assembly, "TEXT ·%s(SB),4,$0-%d\nMOVL $0,AX\nMOVL $0,DX\nMOVL $0,BX\n"+
				"%s"+
				"PUSH%s seed+0(FP)\nPOPF%s\n%s\nPUSHF%s\nPOP%s CX\nMOV%s CX,ret+%d(FP)\nRET\n",
				name, 2*valueBytes, setup, word, word, row.instruction, word, word, word, valueBytes)
			count++
		}
		wantCount := 42
		if target.arch == "386" {
			wantCount = 39
		}
		if count != wantCount {
			t.Fatal("complete independent defined-status family changed", target.arch, count)
		}
		goAssemblerOracleOutput(t, target.arch, assembly.String(), "", true)
		for _, api := range []string{"text", "module", "Go"} {
			t.Run(target.triple+"/"+api, func(t *testing.T) {
				ir := x86CPUQueryIRFromDeclaration(t, assembly.String(), declarations.String(),
					target.arch, target.triple, api)
				for _, optimized := range []bool{false, true} {
					t.Run(fmt.Sprint("O2=", optimized), func(t *testing.T) {
						input := ir
						if optimized {
							input = x86UndefinedTrapOptimizedIR(t, input)
						}
						compileLLVMToObject(t, llc, target.triple, "status.ll", "status.o", input)
						t.Logf("actual LLVM object contains %d declared status writers", count)
					})
				}
			})
		}
	}
}
