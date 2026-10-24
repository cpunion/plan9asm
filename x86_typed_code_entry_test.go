package plan9asm

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

// The original failure was MOVQ $·compare(SB), AX followed by a uintptr FP
// result store. Its code address is not numerical DATA. The matching callback
// consumer uses a typed LLVM entry, not Go's physical ABI0 stack convention.
func TestX86TypedCodeEntryReturn(t *testing.T) {
	for _, arch := range []string{"amd64", "386"} {
		addresses := []string{"MOV", "LEA"}
		if arch == "amd64" {
			addresses = append(addresses, "MOVD", "MOVDQ2Q")
		}
		for _, address := range addresses {
			for _, copies := range []int{0, 2} {
				t.Run(fmt.Sprintf("%s/%s/copies%d", arch, address, copies), func(t *testing.T) {
					pkg, source, options := x86TypedCodeEntryFixture(t, arch, address, copies)
					result, err := TranslateGoModule(pkg, []byte(source), options)
					if err != nil {
						t.Fatal(err)
					}
					defer result.Module.Dispose()
					ir := result.Module.String()
					if !strings.Contains(ir, `ptrtoint ptr @entry.Target to i64`) ||
						!strings.Contains(ir, `define i32 @entry.Target(`) {
						t.Fatal("missing real typed entry address/body")
					}
				})
			}
		}
	}
}

func TestX86TypedCodeEntryRequiresSameRetainedModule(t *testing.T) {
	for _, arch := range []string{"amd64", "386"} {
		pkg, source, options := x86TypedCodeEntryFixture(t, arch, "MOV", 0)
		file, err := Parse(ArchAMD64, source)
		if err != nil {
			t.Fatal(err)
		}
		sigs, err := goSigsForAsmFile(pkg, file, options.ResolveSym, arch, nil, nil, nil)
		if err != nil {
			t.Fatal(err)
		}
		// The same public declarations without the Go binding must not supply
		// an emitted-entry permission to standalone Translate or the FP query.
		if err := CheckX86ABI0ResultInitialization(file.Funcs[0], sigs["entry.PC"], arch); err == nil {
			t.Fatal("declaration-only observer acquired a code address")
		}
		if _, err := Translate(file, Options{Goarch: arch, Sigs: sigs, ResolveSym: options.ResolveSym}); err == nil {
			t.Fatal("standalone signature directory acquired a Go declaration")
		}
		entries := bindGoX86TypedCodeReturns(pkg, file, sigs, options, options.ResolveSym)
		if len(entries) != 1 {
			t.Fatal("original complete source did not bind")
		}
		original := Options{Goarch: arch, TargetTriple: options.TargetTriple, Sigs: sigs,
			ResolveSym: options.ResolveSym, x86TypedCodeReturns: entries}
		if err := validateX86TypedCodeReturns(file, original, options.ResolveSym); err != nil {
			t.Fatalf("unmodified source/signature/target binding failed: %v", err)
		}
		for _, mutation := range []string{"source", "target", "signature", "filtered", "coalesced", "continuation", "target-profile", "same-width-target-profile"} {
			t.Run(arch+"/"+mutation, func(t *testing.T) {
				changed, err := Parse(ArchAMD64, source)
				if err != nil {
					t.Fatal(err)
				}
				request := original
				switch mutation {
				case "source":
					changed.Funcs[0].Instrs[1].Args[0].Sym = "$·Target+1(SB)"
				case "target":
					changed.Funcs[1].Instrs[1].Args[0].Imm = 18
				case "signature":
					request.Sigs = make(map[string]FuncSig)
					for name, sig := range sigs {
						request.Sigs[name] = sig
					}
					sig := request.Sigs["entry.Target"]
					sig.Ret = I64
					request.Sigs["entry.Target"] = sig
				case "filtered":
					changed.Funcs = changed.Funcs[:1]
				case "coalesced":
					changed.Funcs = append(changed.Funcs, changed.Funcs[1])
				case "continuation":
					changed.Funcs[1].x86ContinuationAddresses = map[string]x86Continuation{"·Target": {}}
				case "target-profile":
					request.TargetTriple = "i386-unknown-linux-gnu"
					if arch == "386" {
						request.TargetTriple = "x86_64-unknown-linux-gnu"
					}
				case "same-width-target-profile":
					request.TargetTriple = "x86_64-pc-windows-msvc"
					if arch == "386" {
						request.TargetTriple = "i686-pc-windows-msvc"
					}
				}
				if err := validateX86TypedCodeReturns(changed, request, options.ResolveSym); err == nil {
					t.Fatal("changed module reused its original code-entry binding")
				}
			})
		}
	}
}

func TestX86TypedCodeEntryMatchesIndirectForwarder(t *testing.T) {
	pkg := mustGoPackage(t, "entry", "package entry\n"+x86TypedCodeCallbackDeclarations)
	options := GoModuleOptions{GOARCH: "amd64", TargetTriple: "x86_64-unknown-linux-gnu", ResolveSym: testResolveSym("entry")}
	result, err := TranslateGoModule(pkg, []byte(x86TypedCodeCallbackAssembly), options)
	if err != nil {
		t.Fatal(err)
	}
	defer result.Module.Dispose()
	file, err := Parse(ArchAMD64, x86TypedCodeCallbackAssembly)
	if err != nil {
		t.Fatal(err)
	}
	want := result.Signatures["entry.Target"]
	want.Name = ""
	got := inferX86ABI0Forwarder(file.Funcs[2], result.Signatures["entry.Forward"], "amd64")
	if got == nil || !reflect.DeepEqual(*got, want) {
		t.Fatalf("code entry and consumer have different typed transports: got %+v, want %+v", got, want)
	}
}

func TestX86TypedCodePointerBridgeTranslation(t *testing.T) {
	pkg := mustGoPackage(t, "entry", "package entry\n"+x86TypedCodePointerDeclarations)
	assembly := x86TypedCodePointerAssembly(t)
	goAssemblerOracleOutput(t, "amd64", assembly, "", true)
	result, err := TranslateGoModule(pkg, []byte(assembly), GoModuleOptions{
		GOARCH: "amd64", TargetTriple: "x86_64-unknown-linux-gnu", ResolveSym: testResolveSym("entry"),
	})
	if err != nil {
		t.Fatal(err)
	}
	defer result.Module.Dispose()
	if want := []LLVMType{Ptr, I64, I64}; !reflect.DeepEqual(result.Signatures["entry.Target"].Args, want) {
		t.Fatal("code-address entry lost the indirect caller's exact integer transport")
	}
	if want := []LLVMType{Ptr, Ptr, Ptr}; !reflect.DeepEqual(result.Signatures["entry.TypedCompare"].Args, want) {
		t.Fatal("direct callee lost its actual pointer declaration")
	}
	ir := result.Module.String()
	calledWithPointers := false
	for _, line := range strings.Split(ir, "\n") {
		if strings.Contains(line, "call i32 @entry.TypedCompare(") {
			calledWithPointers = strings.Count(line, "ptr ") == 3
		}
	}
	if !calledWithPointers || strings.Count(ir, "load volatile i64") != 2 {
		t.Fatal("bridge lacks its typed direct call or exact two native pointer reads")
	}
}

func TestX86TypedCodePointerBridgeCannotInventPointerInputs(t *testing.T) {
	declarations := strings.Replace(x86TypedCodePointerDeclarations, "left, right *int64", "left, right uintptr", 1)
	pkg := mustGoPackage(t, "entry", "package entry\n"+declarations)
	result, err := TranslateGoModule(pkg, []byte(x86TypedCodePointerAssembly(t)), GoModuleOptions{
		GOARCH: "amd64", TargetTriple: "x86_64-unknown-linux-gnu", ResolveSym: testResolveSym("entry"),
	})
	if result != nil {
		result.Module.Dispose()
	}
	if err == nil {
		t.Fatal("direct-call bridge manufactured pointer provenance for a uintptr declaration")
	}
}

const x86TypedCodePointerDeclarations = x86TypedCodeCallbackDeclarations + `
func TypedCompare(tls *uint64, left, right *int64) int32
`

// Keep the original indirect consumer unchanged. Only the address-taken
// callback inserts a declared direct-call bridge: identical ABI0 word slots,
// explicitly loaded as Ptr by the typed callee, not a mismatched indirect call.
func x86TypedCodePointerAssembly(t *testing.T) string {
	t.Helper()
	const target = "TEXT ·Target(SB), 4, $0-28"
	const compare = "\tCMPQ AX, CX"
	if strings.Count(x86TypedCodeCallbackAssembly, target) != 1 || strings.Count(x86TypedCodeCallbackAssembly, compare) != 1 {
		t.Fatal("pointer oracle lost its exact numeric callback source")
	}
	assembly := strings.Replace(x86TypedCodeCallbackAssembly, target, `TEXT ·Target(SB), 0, $40-28
	MOVQ tls+0(FP), AX
	MOVQ AX, 0(SP)
	MOVQ left+8(FP), AX
	MOVQ AX, 8(SP)
	MOVQ right+16(FP), AX
	MOVQ AX, 16(SP)
	CALL ·TypedCompare(SB)
	MOVL 24(SP), AX
	MOVL AX, ret+24(FP)
	RET
TEXT ·TypedCompare(SB), 4, $0-28`, 1)
	return strings.Replace(assembly, compare, "\tMOVQ (AX), AX\n\tMOVQ (CX), CX\n"+compare, 1)
}

const x86TypedCodeCallbackDeclarations = `
func PC() uintptr
func Target(tls *uint64, left, right uintptr) int32
func Forward(tls *uint64, left, right, code uintptr) int32
`

const x86TypedCodeCallbackAssembly = `TEXT ·PC(SB), 4, $0-8
	MOVQ $·Target(SB), AX
	MOVQ AX, CX
	MOVQ CX, ret+0(FP)
	RET
TEXT ·Target(SB), 4, $0-28
	MOVQ left+8(FP), AX
	MOVQ right+16(FP), CX
	CMPQ AX, CX
	JL less
	JG greater
	MOVL $0, ret+24(FP)
	RET
less:
	MOVL $-1, ret+24(FP)
	RET
greater:
	MOVL $1, ret+24(FP)
	RET
TEXT ·Forward(SB), 0, $40-36
	MOVQ tls+0(FP), AX
	MOVQ AX, 0(SP)
	MOVQ left+8(FP), AX
	MOVQ AX, 8(SP)
	MOVQ right+16(FP), AX
	MOVQ AX, 16(SP)
	MOVQ code+24(FP), AX
	CALL *AX
	MOVL 24(SP), AX
	MOVL AX, ret+32(FP)
	RET
`

func TestX86TypedCodeEntryFiveTargetObjects(t *testing.T) {
	llc := findLLVM22Tool("llc")
	if llc == "" {
		t.Fatal("LLVM22 llc required for code-entry relocations")
	}
	for _, target := range x86RawReturnTargets {
		t.Run(target.triple, func(t *testing.T) {
			pkg, source, options := x86TypedCodeEntryFixture(t, target.arch, "MOV", 2)
			options.TargetTriple = target.triple
			goAssemblerOracleOutput(t, target.arch, source, "", true)
			result, err := TranslateGoModule(pkg, []byte(source), options)
			if err != nil {
				t.Fatal(err)
			}
			ir := result.Module.String()
			result.Module.Dispose()
			compileLLVMToObject(t, llc, target.triple, "code-entry.ll", "code-entry.o", ir)
		})
	}
}

func TestX86TypedCodeEntryRejectsUnboundOrEscapingRoles(t *testing.T) {
	for _, arch := range []string{"amd64", "386"} {
		for _, mutation := range []string{
			"filtered", "external", "manual only", "signature changed", "non uintptr result",
			"partial copy", "partial result", "addend", "indexed address", "dereference",
			"arithmetic", "indirect call", "store data", "undefined copy", "ABIInternal",
			"raw target", "target fails", "Go body", "missing syntax",
		} {
			t.Run(arch+"/"+mutation, func(t *testing.T) {
				pkg, source, options := x86TypedCodeEntryFixture(t, arch, "MOV", 0)
				word, move, _ := x86TypedCodeEntryMode(arch)
				switch mutation {
				case "filtered":
					options.KeepFunc = func(_, name string) bool { return name != "entry.Target" }
				case "external":
					source = source[:strings.Index(source, "TEXT ·Target")]
				case "manual only":
					pkg = mustGoPackage(t, "entry", "package entry\nfunc PC() uintptr\n")
					options.ManualSig = func(name string) (FuncSig, bool) {
						return FuncSig{Name: name, Ret: I32, Args: []LLVMType{Ptr}, Frame: FrameLayout{
							Params:  []FrameSlot{{Offset: 0, Type: Ptr, Index: 0, Field: -1}},
							Results: []FrameSlot{{Offset: word, Type: I32, Index: 0, Field: -1}},
						}}, name == "entry.Target"
					}
				case "signature changed":
					options.ManualSig = func(name string) (FuncSig, bool) {
						return FuncSig{Name: name, Ret: I32, Args: []LLVMType{I64}, Frame: FrameLayout{
							Params:  []FrameSlot{{Offset: 0, Type: I64, Index: 0, Field: -1}},
							Results: []FrameSlot{{Offset: word, Type: I32, Index: 0, Field: -1}},
						}}, name == "entry.Target"
					}
				case "non uintptr result":
					pkg = mustGoPackage(t, "entry", "package entry\nfunc PC() uint\nfunc Target(p *int32) int32\n")
				case "partial copy":
					source = strings.Replace(source, move+" AX, ret+0(FP)", "MOVW AX, CX\n\t"+move+" CX, ret+0(FP)", 1)
				case "partial result":
					source = strings.Replace(source, move+" AX, ret+0(FP)", "MOVW AX, ret+0(FP)", 1)
				case "addend":
					source = strings.Replace(source, "$·Target(SB)", "$·Target+1(SB)", 1)
				case "indexed address":
					source = strings.Replace(source, move+" $·Target(SB), AX", "LEA"+move[3:]+" ·Target(SB)(CX*1), AX", 1)
				case "dereference":
					source = strings.Replace(source, move+" AX, ret+0(FP)", move+" (AX), AX\n\t"+move+" AX, ret+0(FP)", 1)
				case "arithmetic":
					source = strings.Replace(source, move+" AX, ret+0(FP)", "ADD"+move[3:]+" $1, AX\n\t"+move+" AX, ret+0(FP)", 1)
				case "indirect call":
					source = strings.Replace(source, move+" AX, ret+0(FP)", "CALL AX\n\t"+move+" AX, ret+0(FP)", 1)
				case "store data":
					source = "GLOBL box<>(SB), 16, $8\n" + strings.Replace(source, move+" AX, ret+0(FP)", move+" AX, box<>(SB)\n\t"+move+" AX, ret+0(FP)", 1)
				case "undefined copy":
					source = strings.Replace(source, move+" AX, ret+0(FP)", move+" CX, ret+0(FP)", 1)
				case "ABIInternal":
					source = strings.ReplaceAll(source, "·Target(SB)", "·Target<ABIInternal>(SB)")
				case "raw target":
					source = strings.Replace(source, "MOVL $17, AX", "BYTE $0x90", 1)
				case "target fails":
					source = strings.Replace(source, "MOVL $17, AX", "UNSUPPORTED AX", 1)
				case "Go body":
					pkg = mustGoPackage(t, "entry", "package entry\nfunc PC() uintptr\nfunc Target(p *int32) int32 { return 17 }\n")
				case "missing syntax":
					pkg.Syntax = nil
				}
				result, err := TranslateGoModule(pkg, []byte(source), options)
				if result != nil {
					result.Module.Dispose()
				}
				if err == nil {
					t.Fatal("unbound or escaping code entry was accepted")
				}
			})
		}
	}
}

func x86TypedCodeEntryMode(arch string) (word int64, move, triple string) {
	if arch == "386" {
		return 4, "MOVL", "i386-unknown-linux-gnu"
	}
	return 8, "MOVQ", "x86_64-unknown-linux-gnu"
}

func x86TypedCodeEntryFixture(t *testing.T, arch, address string, copies int) (GoPackage, string, GoModuleOptions) {
	t.Helper()
	word, move, triple := x86TypedCodeEntryMode(arch)
	pkg := mustGoPackage(t, "entry", "package entry\nfunc PC() uintptr\nfunc Target(p *int32) int32\n")
	producer := fmt.Sprintf("%s $·Target(SB), AX", move)
	if address == "LEA" {
		producer = "LEA" + move[3:] + " ·Target(SB), AX"
	} else if address != "MOV" {
		move = address
		producer = move + " $·Target(SB), AX"
	}
	for i := 0; i < copies; i++ {
		producer += "\n\t" + move + " AX, CX\n\t" + move + " CX, AX"
	}
	source := fmt.Sprintf(`TEXT ·PC(SB), 4, $0-%d
	%s
	%s AX, ret+0(FP)
	RET
TEXT ·Target(SB), 4, $0-%d
	MOVL $17, AX
	MOVL AX, ret+%d(FP)
	RET
`, word, producer, move, word+4, word)
	return pkg, source, GoModuleOptions{GOARCH: arch, TargetTriple: triple, ResolveSym: testResolveSym("entry")}
}
