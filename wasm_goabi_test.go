package plan9asm

import (
	"strings"
	"testing"
)

func TestWasmGoABICallerStackAndUnwindLowering(t *testing.T) {
	file, err := Parse(ArchWASM, `TEXT caller(SB), NOSPLIT, $16-0
	CALL callee(SB)
	RET

TEXT callee(SB), NOSPLIT, $0-0
	RET

TEXT unwind(SB), NOSPLIT, $0-0
	RETUNWIND
`)
	if err != nil {
		t.Fatal(err)
	}
	if file.Funcs[0].FrameSize != 16 || file.Funcs[0].ArgSize != 0 {
		t.Fatalf("caller TEXT frame = (%d, %d), want (16, 0)", file.Funcs[0].FrameSize, file.Funcs[0].ArgSize)
	}
	sigs := map[string]FuncSig{
		"caller": {Name: "caller", Ret: Void},
		"callee": {Name: "callee", Ret: Void},
		"unwind": {Name: "unwind", Ret: Void},
	}
	ir, err := Translate(file, Options{
		TargetTriple: "wasm32-unknown-unknown",
		Sigs:         sigs,
		Goarch:       "wasm",
		WASMABI:      WASMABIGo,
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"@SP = external addrspace(1) global i32",
		"@CTXT = external addrspace(1) global i64",
		"define i32 @caller(i32 %pc_b)",
		"switch i32 %pc_b, label %wasm_entry_body [",
		"i32 1, label %wasm_resume_1",
		"sub i32",
		"ptrtoint ptr @caller to i64",
		"shl i64",
		"or i64",
		"store i64",
		"call i32 @callee(i32 0)",
		"wasm_call_unwind_",
		"ret i32 1",
		"add i32",
		"ret i32 0",
	} {
		if !strings.Contains(ir, want) {
			t.Fatalf("Go wasm ABI IR is missing %q:\n%s", want, ir)
		}
	}
}

func TestWasmRETUNWINDRequiresGoABI(t *testing.T) {
	file, err := Parse(ArchWASM, "TEXT unwind(SB), NOSPLIT, $0-0\nRETUNWIND\n")
	if err != nil {
		t.Fatal(err)
	}
	_, err = Translate(file, Options{
		Sigs:   map[string]FuncSig{"unwind": {Name: "unwind", Ret: Void}},
		Goarch: "wasm",
	})
	if err == nil || !strings.Contains(err.Error(), "requires the Go WebAssembly ABI") {
		t.Fatalf("Translate direct RETUNWIND error = %v", err)
	}
}

func TestWasmGoABICallImportUsesResolvedSymbol(t *testing.T) {
	file, err := Parse(ArchWASM, "TEXT ·host(SB), NOSPLIT, $0-0\nCallImport\nRET\n")
	if err != nil {
		t.Fatal(err)
	}
	resolve := func(sym string) string {
		return strings.ReplaceAll(sym, "·", "example.com/pkg.")
	}
	ir, err := Translate(file, Options{
		TargetTriple: "wasm32-unknown-unknown",
		ResolveSym:   resolve,
		Sigs: map[string]FuncSig{
			"example.com/pkg.host": {Name: "example.com/pkg.host", Ret: Void},
		},
		Goarch:  "wasm",
		WASMABI: WASMABIGo,
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		`declare void @"example.com/pkg.host$wasmimport"(i32)`,
		`call void @"example.com/pkg.host$wasmimport"(i32 %`,
	} {
		if !strings.Contains(ir, want) {
			t.Fatalf("Go wasm CallImport IR is missing %q:\n%s", want, ir)
		}
	}
}

func TestWasmGoWriteBarrierNativeSignatures(t *testing.T) {
	tests := []struct {
		name string
		args []LLVMType
		ret  LLVMType
	}{
		{name: "runtime.gcWriteBarrier", args: []LLVMType{I64, I64}, ret: Void},
		{name: "runtime.gcWriteBarrier$local", args: []LLVMType{I64}, ret: I64},
		{name: "runtime.gcWriteBarrier1", ret: I64},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			sig, ok := wasmGoNativeFuncSig(test.name)
			if !ok {
				t.Fatalf("wasmGoNativeFuncSig(%q) not found", test.name)
			}
			if sig.Ret != test.ret || len(sig.Args) != len(test.args) {
				t.Fatalf("wasmGoNativeFuncSig(%q) = (%v) %s, want (%v) %s", test.name, sig.Args, sig.Ret, test.args, test.ret)
			}
			for i := range test.args {
				if sig.Args[i] != test.args[i] {
					t.Fatalf("wasmGoNativeFuncSig(%q) arg %d = %s, want %s", test.name, i, sig.Args[i], test.args[i])
				}
			}
		})
	}
}
