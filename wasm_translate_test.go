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
