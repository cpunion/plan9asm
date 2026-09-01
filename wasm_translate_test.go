package plan9asm

import (
	"strings"
	"testing"
)

func TestTranslateOfficialWasmFloorFunctions(t *testing.T) {
	file, err := Parse(ArchWASM, `
TEXT ·archFloor(SB),NOSPLIT,$0
	Get SP
	F64Load x+0(FP)
	F64Floor
	F64Store ret+8(FP)
	RET

TEXT ·archCeil(SB),NOSPLIT,$0
	Get SP
	F64Load x+0(FP)
	F64Ceil
	F64Store ret+8(FP)
	RET

TEXT ·archTrunc(SB),NOSPLIT,$0
	Get SP
	F64Load x+0(FP)
	F64Trunc
	F64Store ret+8(FP)
	RET
`)
	if err != nil {
		t.Fatal(err)
	}
	sigs := make(map[string]FuncSig)
	for _, name := range []string{"archFloor", "archCeil", "archTrunc"} {
		sigs[name] = FuncSig{
			Name: name,
			Args: []LLVMType{LLVMType("double")},
			Ret:  LLVMType("double"),
			Frame: FrameLayout{
				Params:  []FrameSlot{{Offset: 0, Type: LLVMType("double"), Index: 0, Field: -1}},
				Results: []FrameSlot{{Offset: 8, Type: LLVMType("double"), Index: 0, Field: -1}},
			},
		}
	}
	ir, err := Translate(file, Options{
		TargetTriple: "wasm32-unknown-unknown",
		ResolveSym: func(sym string) string {
			return strings.TrimPrefix(sym, "·")
		},
		Sigs:   sigs,
		Goarch: "wasm",
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"call double @llvm.floor.f64(double %arg0)",
		"call double @llvm.ceil.f64(double %arg0)",
		"call double @llvm.trunc.f64(double %arg0)",
		"ret double %wasm1",
	} {
		if !strings.Contains(ir, want) {
			t.Fatalf("translated IR is missing %q:\n%s", want, ir)
		}
	}
}

func TestTranslateWasmTailWrapper(t *testing.T) {
	file, err := Parse(ArchWASM, `TEXT ·addVV(SB),NOSPLIT,$0
	JMP ·addVV_g(SB)
`)
	if err != nil {
		t.Fatal(err)
	}
	resolve := func(sym string) string { return "math/big." + strings.TrimPrefix(sym, "·") }
	sig := FuncSig{Name: "math/big.addVV", Args: []LLVMType{Ptr, Ptr, Ptr}, Ret: I64}
	target := FuncSig{Name: "math/big.addVV_g", Args: []LLVMType{Ptr, Ptr, Ptr}, Ret: I64}
	ir, err := Translate(file, Options{
		TargetTriple: "wasm32-unknown-unknown",
		ResolveSym:   resolve,
		Sigs: map[string]FuncSig{
			"math/big.addVV":   sig,
			"math/big.addVV_g": target,
		},
		Goarch: "wasm",
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		`declare i64 @"math/big.addVV_g"(ptr, ptr, ptr)`,
		`tail call i64 @"math/big.addVV_g"(ptr %arg0, ptr %arg1, ptr %arg2)`,
		`ret i64 %wasm1`,
	} {
		if !strings.Contains(ir, want) {
			t.Fatalf("translated IR is missing %q:\n%s", want, ir)
		}
	}
}

func TestTranslateOfficialWasmMemoryPrimitives(t *testing.T) {
	file, err := Parse(ArchWASM, `
TEXT runtime·memmove(SB), NOSPLIT, $0-24
	MOVD to+0(FP), R0
	MOVD from+8(FP), R1
	MOVD n+16(FP), R2
	Get R0
	I32WrapI64
	Get R1
	I32WrapI64
	Get R2
	I32WrapI64
	MemoryCopy
	RET

TEXT runtime·memclrNoHeapPointers(SB), NOSPLIT, $0-16
	MOVD ptr+0(FP), R0
	MOVD n+8(FP), R1
	Get R0
	I32WrapI64
	I32Const $0
	Get R1
	I32WrapI64
	MemoryFill
	RET
`)
	if err != nil {
		t.Fatal(err)
	}
	resolve := func(sym string) string { return strings.ReplaceAll(sym, "·", ".") }
	ir, err := Translate(file, Options{
		TargetTriple: "wasm32-unknown-unknown",
		ResolveSym:   resolve,
		Sigs: map[string]FuncSig{
			"runtime.memmove": {
				Name: "runtime.memmove", Args: []LLVMType{Ptr, Ptr, I64}, Ret: Void,
				Frame: FrameLayout{Params: []FrameSlot{
					{Offset: 0, Type: Ptr, Index: 0, Field: -1},
					{Offset: 8, Type: Ptr, Index: 1, Field: -1},
					{Offset: 16, Type: I64, Index: 2, Field: -1},
				}},
			},
			"runtime.memclrNoHeapPointers": {
				Name: "runtime.memclrNoHeapPointers", Args: []LLVMType{Ptr, I64}, Ret: Void,
				Frame: FrameLayout{Params: []FrameSlot{
					{Offset: 0, Type: Ptr, Index: 0, Field: -1},
					{Offset: 8, Type: I64, Index: 1, Field: -1},
				}},
			},
		},
		Goarch: "wasm",
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"call void @llvm.memmove.p0.p0.i32",
		"call void @llvm.memset.p0.i32",
		"define void @runtime.memmove(ptr %arg0, ptr %arg1, i64 %arg2)",
		"define void @runtime.memclrNoHeapPointers(ptr %arg0, i64 %arg1)",
	} {
		if !strings.Contains(ir, want) {
			t.Fatalf("translated IR is missing %q:\n%s", want, ir)
		}
	}
}

func TestWasmStoreRequiresStackAddress(t *testing.T) {
	file, err := Parse(ArchWASM, `TEXT ·floor(SB),NOSPLIT,$0
	F64Load x+0(FP)
	F64Floor
	F64Store ret+8(FP)
	RET
`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = Translate(file, Options{ResolveSym: func(sym string) string {
		return strings.TrimPrefix(sym, "·")
	}, Sigs: map[string]FuncSig{
		"floor": {
			Name: "floor", Args: []LLVMType{LLVMType("double")}, Ret: LLVMType("double"),
			Frame: FrameLayout{
				Params:  []FrameSlot{{Offset: 0, Type: LLVMType("double"), Index: 0, Field: -1}},
				Results: []FrameSlot{{Offset: 8, Type: LLVMType("double"), Index: 0, Field: -1}},
			},
		},
	}, Goarch: "wasm"})
	if err == nil || !strings.Contains(err.Error(), "preceding Get SP") {
		t.Fatalf("Translate error = %v, want missing stack-address diagnostic", err)
	}
}

func TestParseWasmRegistersKeepsGoAndWasmNamesDistinct(t *testing.T) {
	file, err := Parse(ArchWASM, `TEXT ·regs(SB),NOSPLIT,$0
	Get g
	Get CTXT
	Get R15
	Get F31
	Get V15
	Get PC_B
	RET
`)
	if err != nil {
		t.Fatal(err)
	}
	want := []Reg{"G", "CTXT", "R15", "F31", "V15", "PC_B"}
	for i, reg := range want {
		got := file.Funcs[0].Instrs[i+1].Args[0]
		if got.Kind != OpReg || got.Reg != reg {
			t.Fatalf("operand %d = %#v, want wasm register %s", i, got, reg)
		}
	}
}
