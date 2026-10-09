package plan9asm

import (
	"errors"
	"fmt"
	"go/types"
	"strings"
	"testing"
)

func fixedLocalTransportAssembly(name string, args int, body string) string {
	return fmt.Sprintf(`TEXT ·%s(SB),4,$256-%d
MOVQ SP,BP
ADDQ $64,BP
ANDQ $-64,BP
MOVQ n+40(FP),AX
MOVQ AX,64(BP)
MOVQ AX,72(BP)
MOVQ AX,80(BP)
MOVQ AX,88(BP)
VMOVDQA 64(BP),Y14
VMOVDQA Y14,96(BP)
%s
RET
`, name, args, body)
}

func fixedLocalTransportFixture(t *testing.T, assembly, declarations, target, triple string) (GoPackage, GoAMD64FrameCaptureRequest, *File, Options) {
	t.Helper()
	pkg, request := goAMD64FrameTypedFixture(t, declarations, assembly)
	request.Target = target
	file, err := Parse(ArchAMD64, assembly)
	if err != nil {
		t.Fatal(err)
	}
	sigs := make(map[string]FuncSig)
	for _, name := range pkg.Types.Scope().Names() {
		function, ok := pkg.Types.Scope().Lookup(name).(*types.Func)
		if !ok {
			continue
		}
		sig, err := DeriveGoFuncABI0Signature(pkg.Path+"."+name, function, "amd64", types.SizesFor("gc", "amd64"))
		if err != nil {
			t.Fatal(err)
		}
		sigs[sig.Name] = sig
	}
	return pkg, request, file, Options{Goarch: "amd64", TargetTriple: triple, Sigs: sigs,
		ResolveSym: func(symbol string) string { return strings.Replace(symbol, "·", pkg.Path+".", 1) }}
}

const fixedLocalTransportParams = "a []byte, p *uint64, b bool, n uint64"

const fixedLocalTransportBoolBody = `MOVQ a_base+0(FP),CX
MOVQ CX,128(BP)
MOVQ 128(BP),DX
MOVQ p+24(FP),R8
MOVB b+32(FP),BX
MOVB BX,160(BP)
MOVB 160(BP),BX
MOVB BX,ret+48(FP)`

func TestGoAMD64FixedLocalTransportActualObjects(t *testing.T) {
	llc := findLLVM22Tool("llc")
	if llc == "" {
		t.Fatal("LLVM22 required")
	}
	var assembly, declarations strings.Builder
	cases := []struct {
		result string
		args   int
		body   string
	}{
		{"bool", 49, fixedLocalTransportBoolBody},
		{"*uint64", 56, "MOVQ p+24(FP),CX\nMOVQ CX,128(BP)\nMOVQ 128(BP),DX\nMOVQ DX,ret+48(FP)"},
		{"[]byte", 72, "MOVQ a_base+0(FP),CX\nMOVQ a_len+8(FP),DX\nMOVQ a_cap+16(FP),BX\nMOVQ CX,128(BP)\nMOVQ 128(BP),CX\nMOVQ CX,ret_base+48(FP)\nMOVQ DX,ret_len+56(FP)\nMOVQ BX,ret_cap+64(FP)"},
		{"(r []byte, ok bool)", 73, "MOVQ a_base+0(FP),CX\nMOVQ a_len+8(FP),DX\nMOVQ a_cap+16(FP),BX\nMOVQ CX,r_base+48(FP)\nMOVQ DX,r_len+56(FP)\nMOVQ BX,r_cap+64(FP)\nMOVB b+32(FP),CX\nMOVB CX,ok+72(FP)"},
		{"", 48, "MOVQ a_base+0(FP),CX\nMOVQ CX,128(BP)\nMOVQ 128(BP),CX\nMOVQ p+24(FP),R8"},
	}
	for _, condition := range strings.Fields("SETCC SETCS SETEQ SETGE SETGT SETHI SETLE SETLS SETLT SETMI SETNE SETOC SETOS SETPC SETPL SETPS") {
		cases = append(cases, struct {
			result string
			args   int
			body   string
		}{"bool", 49, "CMPQ AX,$0\n" + condition + " BX\nMOVB BX,ret+48(FP)"})
	}
	for i, test := range cases {
		name := fmt.Sprintf("Value%d", i)
		assembly.WriteString(fixedLocalTransportAssembly(name, test.args, test.body))
		fmt.Fprintf(&declarations, "func %s(%s) %s\n", name, fixedLocalTransportParams, test.result)
	}
	for _, target := range []struct{ name, triple string }{
		{"linux/amd64", "x86_64-unknown-linux-gnu"},
		{"darwin/amd64", "x86_64-apple-darwin"},
		{"windows/amd64", "x86_64-pc-windows-msvc"},
	} {
		t.Run(target.name, func(t *testing.T) {
			pkg, request, _, options := fixedLocalTransportFixture(t, assembly.String(), declarations.String(), target.name, target.triple)
			source := requireGoAMD64FrameCapture(t, pkg, request)
			if len(source.functions) != len(cases) {
				t.Fatal("missing complete original object functions")
			}
			for i, test := range cases {
				observed := source.functions[fmt.Sprintf("%s.Value%d", pkg.Path, i)]
				if observed.args != int64(test.args) || observed.locals != 264 || len(observed.code) == 0 {
					t.Fatalf("actual Go ABI0 frame differs: %+v", observed)
				}
				wantOffsets := []int64{0, 8, 16, 24, 32, 40}
				if len(observed.signature.Frame.Params) != len(wantOffsets) {
					t.Fatal("slice ABI0 component inventory changed")
				}
				for j, slot := range observed.signature.Frame.Params {
					if slot.Offset != wantOffsets[j] {
						t.Fatal("actual component offset differs", slot)
					}
				}
			}
			expanded, err := source.preprocess(request.AssemblyFile, []byte(assembly.String()))
			if err != nil {
				t.Fatal(err)
			}
			file, err := Parse(ArchAMD64, expanded)
			if err != nil {
				t.Fatal(err)
			}
			ctx, cancel := runtimeOracleStageContext(t)
			defer cancel()
			bound, err := source.BindFixedLocal(ctx, pkg, file, options)
			if err != nil {
				t.Fatal("complete first-party source lacks typed transport", err)
			}
			fixedLocalAPIs(t, pkg, file, bound, func(t *testing.T, ir string, err error) {
				if err != nil || strings.Count(ir, "%fixed_local_native = alloca") != len(cases) {
					t.Fatal("typed transport lost original private-frame proof", err)
				}
				compileLLVMToObject(t, llc, target.triple, "transport.ll", "transport.o", ir)
				compileLLVMToObject(t, llc, target.triple, "transport-opt.ll", "transport-opt.o", x86MulDivRoundOptimize(t, ir))
			})
		})
	}
}

func TestGoAMD64FixedLocalTransportAuthorityNegatives(t *testing.T) {
	base := fixedLocalTransportAssembly("Value", 49, fixedLocalTransportBoolBody)
	declaration := "func Value(" + fixedLocalTransportParams + ") bool"
	for name, body := range map[string]string{
		"incoming dereference":   "MOVQ (DX),AX\n" + fixedLocalTransportBoolBody,
		"incoming arithmetic":    strings.Replace(fixedLocalTransportBoolBody, "MOVQ p+24(FP),R8", "ADDQ $8,DX", 1),
		"incoming comparison":    strings.Replace(fixedLocalTransportBoolBody, "MOVQ p+24(FP),R8", "CMPQ DX,$0", 1),
		"incoming mask":          strings.Replace(fixedLocalTransportBoolBody, "MOVQ p+24(FP),R8", "KMOVQ DX,K1", 1),
		"incoming partial":       strings.Replace(fixedLocalTransportBoolBody, "MOVQ 128(BP),DX", "MOVL 128(BP),DX", 1),
		"incoming numeric spill": strings.Replace(fixedLocalTransportBoolBody, "MOVQ p+24(FP),R8", "VMOVDQA 128(BP),Y0", 1),
		"bool noncanonical":      strings.Replace(fixedLocalTransportBoolBody, "MOVB BX,ret+48(FP)", "MOVB $2,BX\nMOVB BX,ret+48(FP)", 1),
		"bool unknown":           strings.Replace(fixedLocalTransportBoolBody, "MOVB BX,ret+48(FP)", "MOVB AX,ret+48(FP)", 1),
		"bool wide FP":           strings.Replace(fixedLocalTransportBoolBody, "MOVB b+32(FP),BX", "MOVQ b+32(FP),BX", 1),
		"missing result":         strings.Replace(fixedLocalTransportBoolBody, "MOVB BX,ret+48(FP)", "", 1),
		"private pointer result": "MOVQ BP,ret+48(FP)",
		"false SET flags":        "SETEQ BX\nMOVB BX,ret+48(FP)",
	} {
		t.Run(name, func(t *testing.T) {
			assembly := strings.Replace(base, fixedLocalTransportBoolBody, body, 1)
			pkg, _, file, options := fixedLocalTransportFixture(t, assembly, declaration, "linux/amd64", "x86_64-unknown-linux-gnu")
			if _, err := proveX86FixedLocal(file.Funcs[0], options.Sigs[pkg.Path+".Value"]); !errors.Is(err, ErrProbeNeedsContext) {
				t.Fatal("typed transport granted unproved numerical/memory/result authority", err)
			}
		})
	}
}
