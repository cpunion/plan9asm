//go:build !llgo

package plan9asm

import (
	"strings"
	"testing"
)

// These forms come from github.com/golang/snappy@v1.0.0. The source module
// is cross-assembled with the native Go toolchain before it enters the
// third-party corpus, so these are positive assembler cases rather than
// speculative opcode probes.
func TestTranslateSnappyScalarForms(t *testing.T) {
	t.Run("amd64", func(t *testing.T) {
		ir := translateEcosystemScalarForms(t, ArchAMD64, "x86_64-unknown-linux-gnu", "amd64", `
TEXT snappy(SB),NOSPLIT,$0-0
	MOVQ $9, AX
	SUBB $4, AX
	CALL emitLiteral(SB)
	RET
`)
		for _, want := range []string{"sub i8", "insertvalue { ptr, i64, i64 }"} {
			if !strings.Contains(ir, want) {
				t.Fatalf("snappy amd64 forms are missing %q:\n%s", want, ir)
			}
		}
	})

	t.Run("arm64", func(t *testing.T) {
		ir := translateEcosystemScalarForms(t, ArchARM64, "aarch64-unknown-linux-gnu", "arm64", `
TEXT snappy(SB),NOSPLIT,$0-0
	MOVW $0xa7bd, R16
	MOVKW $0x3905, R16
	MOVKW $(0x1e35<<16), R16
	MULW R16, R11, R12
	MULW R16, R11
	RET
`)
		if got := strings.Count(ir, "mul i32"); got != 2 {
			t.Fatalf("MULW i32 multiplication count = %d, want 2:\n%s", got, ir)
		}
	})
}

// These forms come from github.com/pierrec/lz4/v4@v4.1.29. Closely related
// spellings are kept together so adding one member of an instruction family
// cannot leave the other accepted Go assembler forms silently unsupported.
func TestTranslateLZ4ScalarForms(t *testing.T) {
	t.Run("arm64", func(t *testing.T) {
		ir := translateEcosystemScalarForms(t, ArchARM64, "aarch64-unknown-linux-gnu", "arm64", `
TEXT lz4(SB),NOSPLIT,$0-0
	CMP $4, R1
	CCMP HS, R2, $4, $0
	CCMPW LO, R2, R3, $1
	CCMN EQ, R4, $5, $2
	CCMNW NE, R5, R6, $3
	BVS overflow
	BVC clear
	MADDW R3, R4, R5, R6
	MSUBW R3, R4, R5, R6
	clear:
	overflow:
	RET
`)
		for want, count := range map[string]int{"mul i32": 2, "select i1": 16, "br i1": 2} {
			if got := strings.Count(ir, want); got < count {
				t.Fatalf("ARM64 lz4 %q count = %d, want at least %d:\n%s", want, got, count, ir)
			}
		}
	})

	t.Run("arm", func(t *testing.T) {
		ir := translateEcosystemScalarForms(t, ArchARM, "armv7-unknown-linux-gnueabihf", "arm", `
TEXT lz4(SB),NOSPLIT,$0-0
	SUB.S $1, R1
	BPL positive
	BVS overflow
	BVC clear
	positive:
	MOVH (R2), R3
	MOVHU (R2), R3
	clear:
	overflow:
	RET
`)
		if got := strings.Count(ir, "br i1"); got != 3 {
			t.Fatalf("ARM BPL/BVS/BVC branch count = %d, want 3:\n%s", got, ir)
		}
	})
}

// These memory-source bit scans come from github.com/dgryski/go-bits. Both
// widths and directions are covered because the Go x86 encoder exposes the
// same register-or-memory source family for each form.
func TestTranslateGoBitsScalarForms(t *testing.T) {
	ir := translateEcosystemScalarForms(t, ArchAMD64, "x86_64-unknown-linux-gnu", "amd64", `
TEXT bits(SB),NOSPLIT,$0-16
	BSFQ x+0(FP), AX
	BSRQ x+0(FP), BX
	BSFL x+0(FP), CX
	BSRL x+0(FP), DX
	MOVQ AX, ret+8(FP)
	RET
`)
	for want, count := range map[string]int{"call i64 @llvm.cttz.i64": 1, "call i64 @llvm.ctlz.i64": 1, "call i32 @llvm.cttz.i32": 1, "call i32 @llvm.ctlz.i32": 1} {
		if got := strings.Count(ir, want); got != count {
			t.Fatalf("go-bits %q count = %d, want %d:\n%s", want, got, count, ir)
		}
	}
}

func translateEcosystemScalarForms(t *testing.T, arch Arch, triple, goarch, src string) string {
	t.Helper()
	file, err := Parse(arch, src)
	if err != nil {
		t.Fatal(err)
	}
	ir, err := Translate(file, Options{
		TargetTriple: triple,
		Sigs: map[string]FuncSig{
			"snappy": {Name: "snappy", Ret: Void},
			"lz4":    {Name: "lz4", Ret: Void},
			"bits": {
				Name: "bits",
				Args: []LLVMType{I64},
				Ret:  I64,
				Frame: FrameLayout{
					Params:  []FrameSlot{{Offset: 0, Type: I64, Index: 0, Field: -1}},
					Results: []FrameSlot{{Offset: 8, Type: I64, Index: 0, Field: -1}},
				},
			},
			"emitLiteral": {Name: "emitLiteral", Args: []LLVMType{"{ ptr, i64, i64 }", "{ ptr, i64, i64 }"}, Ret: I64},
		},
		Goarch: goarch,
	})
	if err != nil {
		t.Fatal(err)
	}
	return ir
}
