package main

import (
	"errors"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/xgo-dev/plan9asm"
	"golang.org/x/tools/go/packages"
)

func TestWASMABIForGoPackageTarget(t *testing.T) {
	if got := wasmABIForGoPackageTarget("wasm"); got != plan9asm.WASMABIGo {
		t.Fatalf("wasmABIForGoPackageTarget(wasm) = %v, want Go stack ABI", got)
	}
	for _, goarch := range []string{"amd64", "386", "arm", "arm64"} {
		if got := wasmABIForGoPackageTarget(goarch); got != plan9asm.WASMABIDirect {
			t.Fatalf("wasmABIForGoPackageTarget(%s) = %v, want direct ABI", goarch, got)
		}
	}
}

func TestContainsTestAssembly(t *testing.T) {
	for _, name := range []string{"pkg/routine_test_amd64.s", "pkg/routine_test.s"} {
		if !containsTestAssembly([]string{name}) {
			t.Errorf("containsTestAssembly(%q) = false, want true", name)
		}
	}
	if containsTestAssembly([]string{"pkg/contest_amd64.s", "pkg/routine_amd64.s"}) {
		t.Fatal("containsTestAssembly accepted non-test assembly")
	}
}

func TestIsTestVariantPackage(t *testing.T) {
	base := &packages.Package{ID: "example.com/p", PkgPath: "example.com/p"}
	variant := &packages.Package{ID: "example.com/p [example.com/p.test]", PkgPath: "example.com/p"}
	if isTestVariantPackage(base) || !isTestVariantPackage(variant) {
		t.Fatalf("test variant classification: base=%v variant=%v", isTestVariantPackage(base), isTestVariantPackage(variant))
	}
}

func TestRunOneTargetUsesTestDeclarationForExactTestAssembly(t *testing.T) {
	report, _, err := runOneTarget(
		targetSpec{Goos: "linux", Goarch: "amd64"},
		[]string{"./testdata/testsignature"},
		nil,
		[]string{"testdata/testsignature/convert_test_amd64.s"},
		"github.com/xgo-dev/plan9asm/cmd/plan9asmll",
		t.TempDir(),
		false,
		0,
		true,
		false,
		false,
		filepath.Join("..", ".."),
		compileConfig{},
	)
	if err != nil {
		t.Fatal(err)
	}
	if report.TotalAsm != 1 || report.Success != 1 || report.Failed != 0 {
		t.Fatalf("test assembly report = %#v, want one successful translation", report)
	}
}

func TestTryDeclSigResolvesSamePackageQualifiedPlan9Symbol(t *testing.T) {
	pkg := types.NewPackage("runtime/internal/atomic", "atomic")
	params := types.NewTuple(
		types.NewParam(token.NoPos, pkg, "ptr", types.NewPointer(types.Typ[types.Uint64])),
		types.NewParam(token.NoPos, pkg, "delta", types.Typ[types.Int64]),
	)
	results := types.NewTuple(types.NewParam(token.NoPos, pkg, "ret", types.Typ[types.Uint64]))
	signature := types.NewSignatureType(nil, nil, nil, params, results, false)
	pkg.Scope().Insert(types.NewFunc(token.NoPos, pkg, "Xadd64", signature))

	got, _, ok, err := tryDeclSig(
		pkg.Scope(),
		"runtime∕internal∕atomic·Xadd64",
		"runtime/internal/atomic.Xadd64",
		nil,
		"386",
		types.SizesFor("gc", "386"),
	)
	if err != nil {
		t.Fatal(err)
	}
	if !ok || got.Ret != plan9asm.I64 || len(got.Args) != 2 || got.Args[0] != plan9asm.Ptr || got.Args[1] != plan9asm.I64 {
		t.Fatalf("qualified declaration signature = %#v, ok=%v; want (ptr, i64) i64", got, ok)
	}
}

func TestFallbackSigIncludesAddressedFrameParameters(t *testing.T) {
	fn := plan9asm.Func{
		Sym: "kernelCAS64<>",
		Instrs: []plan9asm.Instr{
			{Op: "MOVW", Args: []plan9asm.Operand{{Kind: plan9asm.OpFP, FPName: "addr", FPOffset: 0}, {Kind: plan9asm.OpReg, Reg: "R2"}}},
			{Op: "MOVW", Args: []plan9asm.Operand{{Kind: plan9asm.OpFPAddr, FPName: "oldval", FPOffset: 4}, {Kind: plan9asm.OpReg, Reg: "R0"}}},
			{Op: "MOVW", Args: []plan9asm.Operand{{Kind: plan9asm.OpFPAddr, FPName: "newval", FPOffset: 12}, {Kind: plan9asm.OpReg, Reg: "R1"}}},
			{Op: "MOVW", Args: []plan9asm.Operand{{Kind: plan9asm.OpReg, Reg: "R0"}, {Kind: plan9asm.OpFP, FPName: "ret", FPOffset: 20}}},
		},
	}
	sig := fallbackSigForAsmFunc(fn, "example.kernelCAS64$local")
	var offsets []int64
	for _, slot := range sig.Frame.Params {
		offsets = append(offsets, slot.Offset)
	}
	want := []int64{0, 4, 12}
	if !reflect.DeepEqual(offsets, want) {
		t.Fatalf("fallback frame parameter offsets = %v, want %v", offsets, want)
	}
}

func TestSigsForAsmFileDiscoversRETTailTarget(t *testing.T) {
	typesPkg := types.NewPackage("example.com/retjmp", "retjmp")
	voidSig := types.NewSignatureType(nil, nil, nil, nil, nil, false)
	typesPkg.Scope().Insert(types.NewFunc(token.NoPos, typesPkg, "f", voidSig))
	typesPkg.Scope().Insert(types.NewFunc(token.NoPos, typesPkg, "f2", voidSig))
	pkg := &packages.Package{
		PkgPath:    typesPkg.Path(),
		Types:      typesPkg,
		TypesSizes: types.SizesFor("gc", "wasm"),
	}
	file, err := plan9asm.Parse(plan9asm.ArchWASM, "TEXT ·f(SB), NOSPLIT, $0-0\n\tRET ·f2(SB)\n")
	if err != nil {
		t.Fatal(err)
	}
	resolve := resolveSymFunc(pkg.PkgPath)
	sigs, _, err := sigsForAsmFile(pkg, file, resolve, "wasm")
	if err != nil {
		t.Fatal(err)
	}
	want := pkg.PkgPath + ".f2"
	if sig, ok := sigs[want]; !ok || sig.Name != want || sig.Ret != plan9asm.Void {
		t.Fatalf("RET tail target signature = %#v, present=%v; want void signature for %q", sig, ok, want)
	}
}

func TestSigsForAsmFileInfersUndeclaredTailForwarderFromDeclaredTarget(t *testing.T) {
	typesPkg := types.NewPackage("example.com/fakecgo", "fakecgo")
	params := types.NewTuple(
		types.NewParam(token.NoPos, typesPkg, "g", types.NewPointer(types.Typ[types.Uint8])),
		types.NewParam(token.NoPos, typesPkg, "setg", types.Typ[types.Uintptr]),
	)
	voidSig := types.NewSignatureType(nil, nil, nil, params, nil, false)
	typesPkg.Scope().Insert(types.NewFunc(token.NoPos, typesPkg, "x_cgo_init", voidSig))
	pkg := &packages.Package{
		PkgPath:    typesPkg.Path(),
		Types:      typesPkg,
		TypesSizes: types.SizesFor("gc", "386"),
	}
	file, err := plan9asm.Parse(plan9asm.ArchAMD64, "TEXT x_cgo_init_trampoline(SB),NOSPLIT|NOFRAME,$0\n\tJMP ·x_cgo_init(SB)\n\tRET\n")
	if err != nil {
		t.Fatal(err)
	}
	resolve := resolveSymFunc(pkg.PkgPath)
	sigs, _, err := sigsForAsmFile(pkg, file, resolve, "386")
	if err != nil {
		t.Fatal(err)
	}
	caller := sigs["x_cgo_init_trampoline"]
	target := sigs[pkg.PkgPath+".x_cgo_init"]
	if caller.Ret != plan9asm.Void || len(caller.Args) != 2 || caller.Args[0] != plan9asm.Ptr || caller.Args[1] != plan9asm.I32 {
		t.Fatalf("tail-forwarder signature = %#v, target = %#v; want (ptr, i32) void inherited from declared target", caller, target)
	}
}

func TestValidateDeclaredTextArgSizesClassifiesOnlyExplicitABIMismatches(t *testing.T) {
	resolve := func(sym string) string { return "example.com/ext." + strings.TrimPrefix(sym, "·") }
	declared := map[string]int64{"example.com/ext.StructFieldB": 17}

	tests := []struct {
		name    string
		text    string
		argSize int64
		wantErr bool
	}{
		{name: "wrong explicit size", text: "TEXT ·StructFieldB(SB), $0-25", argSize: 25, wantErr: true},
		{name: "matching explicit size", text: "TEXT ·StructFieldB(SB), $0-17", argSize: 17},
		{name: "omitted argument size", text: "TEXT ·StructFieldB(SB), $0"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			file := &plan9asm.File{Funcs: []plan9asm.Func{{
				Sym:     "·StructFieldB",
				ArgSize: test.argSize,
				Instrs:  []plan9asm.Instr{{Op: plan9asm.OpTEXT, Raw: test.text}},
			}}}
			err := validateDeclaredTextArgSizes(file, resolve, declared, "386")
			var mismatch *asmABINotApplicableError
			if test.wantErr {
				if !errors.As(err, &mismatch) {
					t.Fatalf("validateDeclaredTextArgSizes() error = %v, want asmABINotApplicableError", err)
				}
				if mismatch.Symbol != "example.com/ext.StructFieldB" || mismatch.DeclaredArgSize != test.argSize || mismatch.ExpectedArgSize != 17 || mismatch.Goarch != "386" {
					t.Fatalf("mismatch = %#v", mismatch)
				}
				return
			}
			if err != nil {
				t.Fatalf("validateDeclaredTextArgSizes() error = %v", err)
			}
		})
	}
}

func TestReadAsmSourceExpandsLocalIncludesRecursively(t *testing.T) {
	dir := t.TempDir()
	write := func(name, contents string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(dir, name), []byte(contents), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	write("ops.h", "#define STEP(a, b) ADDQ a, b\n")
	write("macros.h", "#include \"ops.h\"\n#define ROUND(a, b) STEP(a, b)\n")
	write("macro_amd64.s", "#include \"textflag.h\"\n#include \"macros.h\"\nTEXT ·macro(SB),NOSPLIT,$0-0\nROUND(AX, BX)\nRET\n")

	src, err := readAsmSource(filepath.Join(dir, "macro_amd64.s"), dir)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(src), `#include "macros.h"`) || !strings.Contains(string(src), "#define STEP") {
		t.Fatalf("local include was not expanded:\n%s", src)
	}
	if strings.Contains(string(src), `#include "textflag.h"`) || !strings.Contains(string(src), "#define NOSPLIT") {
		t.Fatalf("standard textflag.h was not expanded from GOROOT:\n%s", src)
	}
	file, err := plan9asm.Parse(plan9asm.ArchAMD64, string(src))
	if err != nil {
		t.Fatal(err)
	}
	if got := file.Funcs[0].Instrs[1].Op; got != plan9asm.Op("ADDQ") {
		t.Fatalf("expanded opcode = %s, want ADDQ", got)
	}
}

func TestReadAsmSourceExpandsStandardFuncdataMacros(t *testing.T) {
	dir := t.TempDir()
	asm := filepath.Join(dir, "macro_amd64.s")
	src := `#include "funcdata.h"
TEXT ·macro(SB),NOSPLIT,$0-0
GO_ARGS
GO_RESULTS_INITIALIZED
NO_LOCAL_POINTERS
RET
`
	if err := os.WriteFile(asm, []byte(src), 0o600); err != nil {
		t.Fatal(err)
	}
	expanded, err := readAsmSource(asm, dir)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(expanded), `#include "funcdata.h"`) || !strings.Contains(string(expanded), "#define GO_ARGS") {
		t.Fatalf("standard funcdata.h was not expanded from GOROOT:\n%s", expanded)
	}
	file, err := plan9asm.Parse(plan9asm.ArchAMD64, string(expanded))
	if err != nil {
		t.Fatal(err)
	}
	var got []plan9asm.Op
	for _, ins := range file.Funcs[0].Instrs {
		got = append(got, ins.Op)
	}
	want := []plan9asm.Op{"TEXT", "FUNCDATA", "PCDATA", "FUNCDATA", plan9asm.OpRET}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expanded funcdata ops = %v, want %v", got, want)
	}
}

func TestReadAsmSourceResolvesGoIncludePathRelativeToPkgInclude(t *testing.T) {
	dir := t.TempDir()
	asm := filepath.Join(dir, "tls_386.s")
	if err := os.WriteFile(asm, []byte("#include \"../../src/runtime/go_tls.h\"\nTEXT ·tls(SB),NOSPLIT,$0-0\nRET\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	expanded, err := readAsmSource(asm, dir)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(expanded), `#include "../../src/runtime/go_tls.h"`) || !strings.Contains(string(expanded), "get_tls") {
		t.Fatalf("GOROOT-relative runtime header was not expanded:\n%s", expanded)
	}
}

func TestExtractSupportedOpsFindsCompleteAddedInstructionFamilies(t *testing.T) {
	repoRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	supported, err := extractSupportedOps(repoRoot, "amd64")
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		"MOVD", "VMOVD", "VMOVQ", "PREFETCHNTA", "PREFETCHT0", "PREFETCHT1", "PREFETCHT2", "UD2",
		"CVTSL2SS", "CVTSL2SD", "CVTSQ2SS", "CVTSQ2SD", "CVTPL2PS", "CVTPL2PD",
		"VBROADCASTSS", "VBROADCASTSD",
		"PMULDQ", "PMULULQ", "VPMULDQ", "VPMULUDQ",
		"MOVDDUP", "MOVSHDUP", "MOVSLDUP",
		"VMOVDDUP", "VMOVSHDUP", "VMOVSLDUP",
		"PMULLD", "VPMULLD",
		"PMULLW", "PMULHW", "PMULHUW", "PMULHRSW",
		"VPMULLW", "VPMULHW", "VPMULHUW", "VPMULHRSW",
		"PMADDWL", "VPMADDWD", "PMADDUBSW", "VPMADDUBSW",
		"VPDPBUSD", "VPDPBUSDS", "VPDPWSSD", "VPDPWSSDS",
		"PAVGB", "PAVGW", "VPAVGB", "VPAVGW",
		"PCMPGTB", "PCMPGTW", "PCMPGTL", "PCMPGTQ",
		"PSRAW", "PSRAL", "VPSRAW", "VPSRAD", "VPSRAQ",
		"VPMINSB", "VPMINUB", "VPMAXSB", "VPMAXUB",
		"VPMINSW", "VPMINUW", "VPMAXSW", "VPMAXUW",
		"VPMINSD", "VPMINUD", "VPMAXSD", "VPMAXUD",
		"VPMINSQ", "VPMINUQ", "VPMAXSQ", "VPMAXUQ",
		"SHUFPS", "SHUFPD", "VSHUFPS", "VSHUFPD",
		"PBLENDW", "BLENDPS", "BLENDPD", "VPBLENDW", "VPBLENDD", "VBLENDPS", "VBLENDPD",
		"PUNPCKLBW", "PUNPCKHBW", "PUNPCKLWL", "PUNPCKHWL",
		"PUNPCKLLQ", "PUNPCKHLQ", "PUNPCKLQDQ", "PUNPCKHQDQ",
		"VPUNPCKLBW", "VPUNPCKHBW", "VPUNPCKLWD", "VPUNPCKHWD",
		"VPUNPCKLDQ", "VPUNPCKHDQ", "VPUNPCKLQDQ", "VPUNPCKHQDQ",
		"UNPCKLPS", "UNPCKHPS", "UNPCKLPD", "UNPCKHPD",
		"VUNPCKLPS", "VUNPCKHPS", "VUNPCKLPD", "VUNPCKHPD",
		"VPERMQ", "VPERMPD",
		"VPERMILPD", "VPERMILPS",
		"VMOVAPD", "VMOVAPS", "VMOVUPD", "VMOVUPS",
		"VMOVDQA32", "VMOVDQA64", "VMOVDQU8", "VMOVDQU16", "VMOVDQU32", "VMOVDQU64",
		"VINSERTF128", "VINSERTI128",
		"VINSERTF32X4", "VINSERTF64X2", "VINSERTI32X4", "VINSERTI64X2",
		"VINSERTF32X8", "VINSERTF64X4", "VINSERTI32X8", "VINSERTI64X4",
		"PSHUFB", "VPSHUFB",
		"ADDSUBPS", "ADDSUBPD", "VADDSUBPS", "VADDSUBPD",
		"VCVTDQ2PS", "VCVTPS2DQ", "VCVTTPS2DQ", "VSQRTPD", "VSQRTPS",
		"PMOVSXBW", "PMOVSXBD", "PMOVSXBQ", "PMOVSXWD", "PMOVSXWQ", "PMOVSXDQ",
		"VPMOVSXBW", "VPMOVSXBD", "VPMOVSXBQ", "VPMOVSXWD", "VPMOVSXWQ", "VPMOVSXDQ",
		"PMOVZXBW", "PMOVZXBD", "PMOVZXBQ", "PMOVZXWD", "PMOVZXWQ", "PMOVZXDQ",
		"VPMOVZXBW", "VPMOVZXBD", "VPMOVZXBQ", "VPMOVZXWD", "VPMOVZXWQ", "VPMOVZXDQ",
		"PHADDD", "PHADDSW", "PHADDW", "PHSUBD", "PHSUBSW", "PHSUBW",
		"VPHADDD", "VPHADDSW", "VPHADDW", "VPHSUBD", "VPHSUBSW", "VPHSUBW",
		"PHMINPOSUW", "VPHMINPOSUW",
		"MASKMOVQ", "MASKMOVOU", "MASKMOVDQU", "VMASKMOVDQU",
		"JCXZW", "JCXZL", "JCXZQ",
		"PALIGNR", "VPALIGNR",
		"PCLMULQDQ", "VPCLMULQDQ",
		"PSLLO", "PSLLDQ", "PSRLO", "PSRLDQ", "VPSLLDQ", "VPSRLDQ",
		"RCLB", "RCLW", "RCLL", "RCLQ", "RCRB", "RCRW", "RCRL", "RCRQ",
		"PSHUFHW", "PSHUFLW", "VPSHUFHW", "VPSHUFLW",
		"AESENC", "AESENCLAST", "AESDEC", "AESDECLAST", "AESIMC", "AESKEYGENASSIST",
		"VAESENC", "VAESENCLAST", "VAESDEC", "VAESDECLAST", "VAESIMC", "VAESKEYGENASSIST",
		"CVTSS2SL", "CVTSS2SQ", "CVTSD2SL", "CVTSD2SQ",
		"CVTTSS2SL", "CVTTSS2SQ", "CVTTSD2SL", "CVTTSD2SQ",
		"VCVTSS2SI", "VCVTSS2SIQ", "VCVTSD2SI", "VCVTSD2SIQ",
		"VCVTTSS2SI", "VCVTTSS2SIQ", "VCVTTSD2SI", "VCVTTSD2SIQ",
		"VCVTSS2USIL", "VCVTSS2USIQ", "VCVTSD2USIL", "VCVTSD2USIQ",
		"VCVTTSS2USIL", "VCVTTSS2USIQ", "VCVTTSD2USIL", "VCVTTSD2USIQ",
		"ROUNDPS", "ROUNDPD", "ROUNDSS", "ROUNDSD",
		"VROUNDPS", "VROUNDPD", "VROUNDSS", "VROUNDSD",
		"BOUNDW", "BOUNDL",
		"CLC", "STC", "CMC",
		"PSLLW", "PSLLL", "PSLLQ", "PSRLW", "PSRLL", "PSRLQ",
		"VPSLLW", "VPSLLD", "VPSLLQ", "VPSRLW", "VPSRLD", "VPSRLQ",
		"PMULLW", "PMULHW", "PMULHUW", "PMULHRSW",
		"VPMULLW", "VPMULHW", "VPMULHUW", "VPMULHRSW",
		"LDMXCSR", "VLDMXCSR", "STMXCSR", "VSTMXCSR",
		"CVTPS2PL", "CVTTPS2PL", "CVTPS2PD", "VCVTPS2PD",
		"VPCMPEQB", "VPCMPEQW", "VPCMPEQD", "VPCMPEQQ",
		"VPCMPGTB", "VPCMPGTW", "VPCMPGTD", "VPCMPGTQ",
		"PANDN", "VPANDN", "VPANDND", "VPANDNQ",
		"VPBROADCASTB", "VPBROADCASTW", "VPBROADCASTD", "VPBROADCASTQ",
		"PADDB", "PADDL", "PADDQ", "PADDSB", "PADDSW", "PADDUSB", "PADDUSW", "PADDW",
		"VPADDB", "VPADDD", "VPADDQ", "VPADDSB", "VPADDSW", "VPADDUSB", "VPADDUSW", "VPADDW",
		"PSUBB", "PSUBL", "PSUBQ", "PSUBSB", "PSUBSW", "PSUBUSB", "PSUBUSW", "PSUBW",
		"VPSUBB", "VPSUBD", "VPSUBQ", "VPSUBSB", "VPSUBSW", "VPSUBUSB", "VPSUBUSW", "VPSUBW",
		"PSADBW", "VPSADBW",
		"VEXTRACTF128", "VEXTRACTI128",
		"VEXTRACTF32X4", "VEXTRACTF64X2", "VEXTRACTI32X4", "VEXTRACTI64X2",
		"VEXTRACTF32X8", "VEXTRACTF64X4", "VEXTRACTI32X8", "VEXTRACTI64X4",
		"PEXTRB", "PEXTRW", "PEXTRD", "PEXTRQ",
		"VPEXTRB", "VPEXTRW", "VPEXTRD", "VPEXTRQ",
		"MOVHLPS", "MOVLHPS", "VMOVHLPS", "VMOVLHPS",
		"POPCNTW", "POPCNTL", "POPCNTQ",
		"NOTB", "NOTW", "NOTL", "NOTQ",
		"PDEPL", "PDEPQ", "PEXTL", "PEXTQ",
		"BTW", "BTL", "BTQ", "BTCW", "BTCL", "BTCQ",
		"BTRW", "BTRL", "BTRQ", "BTSW", "BTSL", "BTSQ",
		"RDMSR", "WRMSR",
		"LGDT", "LIDT", "SGDT", "SIDT",
		"LLDT", "LTR", "LMSW",
		"SLDTW", "SLDTL", "SLDTQ",
		"SMSWW", "SMSWL", "SMSWQ",
		"STRW", "STRL", "STRQ",
		"CMPXCHGB", "CMPXCHGW", "CMPXCHGL", "CMPXCHGQ", "CMPXCHG8B", "CMPXCHG16B",
		"TZCNTW", "TZCNTL", "TZCNTQ",
		"CMPPD", "CMPPS", "CMPSD", "CMPSS", "VCMPPD", "VCMPPS", "VCMPSD", "VCMPSS",
		"PACKSSLW", "PACKSSWB", "PACKUSDW", "PACKUSWB",
		"VPACKSSDW", "VPACKSSWB", "VPACKUSDW", "VPACKUSWB",
		"VPERMD", "VPERMPS",
		"SETCC", "SETCS", "SETEQ", "SETGE", "SETGT", "SETHI", "SETLE", "SETLS",
		"SETLT", "SETMI", "SETNE", "SETOC", "SETOS", "SETPC", "SETPL", "SETPS",
		"VMOVSD", "VMOVSS",
		"VCOMISD", "VCOMISS", "VUCOMISD", "VUCOMISS",
		"INCB", "INCW", "INCL", "INCQ", "DECB", "DECW", "DECL", "DECQ",
		"MOVBWSX", "MOVBWZX", "MOVBLSX", "MOVBLZX", "MOVBQSX", "MOVBQZX",
		"MOVWLSX", "MOVWLZX", "MOVWQSX", "MOVWQZX", "MOVLQSX", "MOVLQZX",
		"MOVSWW", "MOVZWW",
		"KMOVB", "KMOVW", "KMOVD", "KMOVQ",
		"VGATHERDPS", "VGATHERQPD", "VPGATHERDD", "VPGATHERQQ",
		"VGATHERDPD", "VPGATHERDQ", "VGATHERQPS", "VPGATHERQD",
		"VGATHERPF0DPD", "VGATHERPF0DPS", "VGATHERPF0QPD", "VGATHERPF0QPS",
		"VGATHERPF1DPD", "VGATHERPF1DPS", "VGATHERPF1QPD", "VGATHERPF1QPS",
		"VPTERNLOGD", "VPTERNLOGQ",
		"VPROLD", "VPROLQ", "VPROLVD", "VPROLVQ",
		"VPRORD", "VPRORQ", "VPRORVD", "VPRORVQ",
		"VPANDD", "VPANDQ", "VPANDND", "VPANDNQ",
		"VPORD", "VPORQ", "VPXORD", "VPXORQ",
		"ANDPS", "ANDPD", "ANDNPS", "ANDNPD", "ORPS", "ORPD", "XORPS", "XORPD",
		"VANDPS", "VANDPD", "VANDNPS", "VANDNPD", "VORPS", "VORPD", "VXORPS", "VXORPD",
		"HADDPS", "HADDPD", "HSUBPS", "HSUBPD", "VHADDPS", "VHADDPD", "VHSUBPS", "VHSUBPD",
		"FMOVB", "FMOVBP", "FMOVD", "FMOVDP", "FMOVF", "FMOVFP", "FMOVL", "FMOVLP",
		"FMOVV", "FMOVVP", "FMOVW", "FMOVWP", "FMOVX", "FMOVXP",
		"FCOMD", "FCOMDP", "FCOMDPP", "FCOMF", "FCOMFP", "FCOMI", "FCOMIP",
		"FCOML", "FCOMLP", "FCOMW", "FCOMWP", "FUCOM", "FUCOMI", "FUCOMIP", "FUCOMP", "FUCOMPP",
		"FBLD", "FBSTP", "FLDCW", "FLDENV", "FRSTOR", "FSAVE", "FSTCW", "FSTENV", "FSTSW",
		"F2XM1", "FABS", "FCHS", "FCLEX", "FCOS", "FDECSTP", "FINCSTP", "FINIT",
		"FLD1", "FLDL2E", "FLDL2T", "FLDLG2", "FLDLN2", "FLDPI", "FLDZ", "FNOP",
		"FPATAN", "FPREM", "FPREM1", "FPTAN", "FRNDINT", "FSCALE", "FSIN", "FSINCOS",
		"FSQRT", "FTST", "FXAM", "FXTRACT", "FYL2X", "FYL2XP1", "FXCHD", "LAHF", "SAHF",
	}
	for _, stem := range []string{"FADD", "FMUL", "FSUB", "FSUBR", "FDIV", "FDIVR"} {
		for _, suffix := range []string{"W", "L", "F", "D", "DP"} {
			want = append(want, stem+suffix)
		}
	}
	for _, stem := range []string{"ROL", "ROR", "SAR", "SAL", "SHL", "SHR"} {
		for _, width := range []string{"B", "W", "L", "Q"} {
			want = append(want, stem+width)
		}
	}
	for _, stem := range []string{"KAND", "KANDN", "KOR", "KXNOR", "KXOR", "KNOT", "KTEST", "KORTEST"} {
		for _, width := range []string{"B", "W", "D", "Q"} {
			want = append(want, stem+width)
		}
	}
	for _, family := range []string{"VFMADD", "VFMSUB", "VFNMADD", "VFNMSUB", "VFMADDSUB", "VFMSUBADD"} {
		for _, order := range []string{"132", "213", "231"} {
			for _, element := range []string{"PS", "PD"} {
				want = append(want, family+order+element)
			}
			if family != "VFMADDSUB" && family != "VFMSUBADD" {
				for _, element := range []string{"SS", "SD"} {
					want = append(want, family+order+element)
				}
			}
		}
	}
	for _, family := range []string{"VADD", "VSUB", "VMUL", "VDIV", "VMAX", "VMIN"} {
		for _, element := range []string{"PS", "PD", "SS", "SD"} {
			want = append(want, family+element)
		}
	}
	for _, width := range []string{"W", "L", "Q"} {
		for _, condition := range []string{"CC", "CS", "EQ", "GE", "GT", "HI", "LE", "LS", "LT", "MI", "NE", "OC", "OS", "PC", "PL", "PS"} {
			want = append(want, "CMOV"+width+condition)
		}
	}
	for _, op := range want {
		if _, ok := supported[op]; !ok {
			t.Errorf("supported opcode extraction omitted %s", op)
		}
	}
	for _, op := range []string{
		"VPOPCNTB", "VPOPCNTW", "VPOPCNTD", "VPOPCNTQ",
		"VPCOMPRESSB", "VPCOMPRESSW", "VPCOMPRESSD", "VPCOMPRESSQ",
	} {
		if _, ok := supported[op]; !ok {
			t.Errorf("supported opcode extraction omitted %s", op)
		}
	}
	arm64Supported, err := extractSupportedOps(repoRoot, "arm64")
	if err != nil {
		t.Fatal(err)
	}
	for _, op := range []string{
		"VZIP1", "VZIP2", "VUZP1", "VUZP2", "VTRN1", "VTRN2",
		"VUSHLL", "VUSHLL2", "VSSHLL", "VSSHLL2",
		"VUXTL", "VUXTL2", "VSXTL", "VSXTL2",
		"SXTB", "SXTBW", "SXTH", "SXTHW", "SXTW",
		"UXTB", "UXTBW", "UXTH", "UXTHW", "UXTW",
		"VCNT",
		"VUADDW", "VUADDW2",
	} {
		if _, ok := arm64Supported[op]; !ok {
			t.Errorf("ARM64 supported opcode extraction omitted %s", op)
		}
	}
}

func TestExtractSupportedOpsFindsPackageLevelSpecTableWithoutOpcodeName(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "amd64_table.go"), []byte(`package sample
var packedFamilySpecs = map[string]int{
	"VTABLEOP": 1,
}
`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "translate.go"), []byte("package sample\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	supported, err := extractSupportedOps(dir, "amd64")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := supported["VTABLEOP"]; !ok {
		t.Fatal("package-level table-driven opcode was not extracted")
	}
}

func TestExplicitSingleTargetUsesMatrixReport(t *testing.T) {
	tests := []struct {
		name       string
		allTargets bool
		targets    string
		reports    int
		want       bool
	}{
		{name: "legacy-default-single", reports: 1, want: false},
		{name: "explicit-single", targets: "linux/amd64", reports: 1, want: true},
		{name: "explicit-multiple", targets: "linux/amd64,windows/amd64", reports: 2, want: true},
		{name: "all-targets", allTargets: true, reports: 9, want: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := useMatrixReport(test.allTargets, test.targets, test.reports); got != test.want {
				t.Fatalf("useMatrixReport(%v, %q, %d) = %v, want %v", test.allTargets, test.targets, test.reports, got, test.want)
			}
		})
	}
}

func TestExtractSupportedOpsFindsCompleteARM64AddedFamilies(t *testing.T) {
	repoRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	supported, err := extractSupportedOps(repoRoot, "arm64")
	if err != nil {
		t.Fatal(err)
	}
	for _, op := range []string{
		"FSQRTS", "FSQRTD", "FMOVS", "FMOVD",
		"FMADDS", "FMADDD", "FMSUBS", "FMSUBD", "FNMADDS", "FNMADDD", "FNMSUBS", "FNMSUBD",
		"SCVTFD", "SCVTFS", "SCVTFWD", "SCVTFWS", "UCVTFD", "UCVTFS", "UCVTFWD", "UCVTFWS",
		"FADDS", "FADDD", "FSUBS", "FSUBD", "FMULS", "FMULD", "FNMULS", "FNMULD", "FDIVS", "FDIVD",
		"FMAXS", "FMAXD", "FMINS", "FMIND", "FMAXNMS", "FMAXNMD", "FMINNMS", "FMINNMD",
		"CASPW", "CASPD",
		"LDXPW", "LDXP", "LDAXPW", "LDAXP",
		"STXPW", "STXP", "STLXPW", "STLXP",
	} {
		if _, ok := supported[op]; !ok {
			t.Errorf("supported opcode extraction omitted %s", op)
		}
	}
}

func TestExtractSupportedOpsFindsCompleteARMAddedFamilies(t *testing.T) {
	repoRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	supported, err := extractSupportedOps(repoRoot, "arm")
	if err != nil {
		t.Fatal(err)
	}
	for _, op := range []string{
		"MOVF", "MOVD",
		"NEGF", "NEGD", "ABSF", "ABSD", "SQRTF", "SQRTD", "MOVFD", "MOVDF", "CMPF", "CMPD",
		"ADDF", "ADDD", "SUBF", "SUBD", "MULF", "MULD", "NMULF", "NMULD", "DIVF", "DIVD",
		"MULAF", "MULAD", "MULSF", "MULSD", "NMULAF", "NMULAD", "NMULSF", "NMULSD",
		"FMULAF", "FMULAD", "FMULSF", "FMULSD", "FNMULAF", "FNMULAD", "FNMULSF", "FNMULSD",
		"MOVWF", "MOVWD", "MOVFW", "MOVDW", "PLD",
		"DIV", "DIVU", "MOD", "MODU",
	} {
		if _, ok := supported[op]; !ok {
			t.Errorf("supported opcode extraction omitted %s", op)
		}
	}
}

func TestExtractSupportedOpsFindsCompleteWasmFloatUnaryFamily(t *testing.T) {
	repoRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	supported, err := extractSupportedOps(repoRoot, "wasm")
	if err != nil {
		t.Fatal(err)
	}
	for _, width := range []string{"F32", "F64"} {
		for _, operation := range []string{"ABS", "NEG", "CEIL", "FLOOR", "TRUNC", "NEAREST", "SQRT"} {
			op := width + operation
			if _, ok := supported[op]; !ok {
				t.Errorf("supported opcode extraction omitted %s", op)
			}
		}
	}
}

func TestAsmFilesOfPkgSkipsCommentOnlyAssembly(t *testing.T) {
	dir := t.TempDir()
	comments := filepath.Join(dir, "comments.s")
	code := filepath.Join(dir, "code.s")
	include := filepath.Join(dir, "include.s")
	for path, contents := range map[string]string{
		comments: "//go:build amd64\n\n/* license only */\n",
		code:     "// comment\nTEXT ·f(SB),0,$0-0\n",
		include:  "#include \"textflag.h\"\n",
	} {
		if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	missing := filepath.Join(dir, "missing.s")
	pkg := &packages.Package{OtherFiles: []string{comments, code, include, missing}}
	want := []string{code, include, missing}
	if got := asmFilesOfPkg(pkg); !reflect.DeepEqual(got, want) {
		t.Fatalf("asmFilesOfPkg() = %#v, want %#v", got, want)
	}
}

func TestCollectAsmTasksHonorsExactModuleRelativeAllowlist(t *testing.T) {
	dir := t.TempDir()
	pkgDir := filepath.Join(dir, "pkg")
	if err := os.MkdirAll(pkgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	amd64File := filepath.Join(pkgDir, "fast.s")
	arm64File := filepath.Join(pkgDir, "portable.s")
	for _, path := range []string{amd64File, arm64File} {
		if err := os.WriteFile(path, []byte("TEXT ·f(SB),0,$0-0\nRET\n"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	pkg := &packages.Package{
		PkgPath:    "example.com/root/pkg",
		OtherFiles: []string{amd64File, arm64File},
		Module:     &packages.Module{Path: "example.com/root", Dir: dir},
	}
	tasks, packages := collectAsmTasks([]*packages.Package{pkg}, "/out", []string{"pkg/portable.s"})
	if len(tasks) != 1 || tasks[0].AsmFile != arm64File {
		t.Fatalf("collectAsmTasks() tasks = %#v, want only %s", tasks, arm64File)
	}
	if !reflect.DeepEqual(packages, []string{"example.com/root/pkg"}) {
		t.Fatalf("collectAsmTasks() packages = %#v", packages)
	}
}

func TestFilterPackagesByModuleExcludesNestedModules(t *testing.T) {
	pkgs := []*packages.Package{
		{PkgPath: "example.com/root/pkg", Module: &packages.Module{Path: "example.com/root"}},
		{PkgPath: "example.com/root/v2", Module: &packages.Module{Path: "example.com/root/v2"}},
		{PkgPath: "example.com/root/vendorless", Module: nil},
	}
	want := []*packages.Package{pkgs[0]}
	if got := filterPackagesByModule(pkgs, "example.com/root"); !reflect.DeepEqual(got, want) {
		t.Fatalf("filterPackagesByModule() = %#v, want %#v", got, want)
	}
}

func TestDefaultMatrixTargetsCoversEveryPlan9Architecture(t *testing.T) {
	want := []targetSpec{
		{Goos: "darwin", Goarch: "amd64"},
		{Goos: "darwin", Goarch: "arm64"},
		{Goos: "linux", Goarch: "386"},
		{Goos: "linux", Goarch: "amd64"},
		{Goos: "linux", Goarch: "arm"},
		{Goos: "linux", Goarch: "arm64"},
		{Goos: "windows", Goarch: "386"},
		{Goos: "windows", Goarch: "amd64"},
		{Goos: "windows", Goarch: "arm64"},
		{Goos: "js", Goarch: "wasm"},
		{Goos: "wasip1", Goarch: "wasm"},
	}
	if got := defaultMatrixTargets(); !reflect.DeepEqual(got, want) {
		t.Fatalf("defaultMatrixTargets() = %#v, want %#v", got, want)
	}
}

func TestExternalCorpusTargetArchitectureAndTriple(t *testing.T) {
	tests := []struct {
		goos       string
		goarch     string
		wantArch   plan9asm.Arch
		wantTriple string
	}{
		{goos: "linux", goarch: "arm", wantArch: plan9asm.ArchARM, wantTriple: "armv7-unknown-linux-gnueabihf"},
		{goos: "js", goarch: "wasm", wantArch: plan9asm.ArchWASM, wantTriple: "wasm32-unknown-unknown"},
		{goos: "wasip1", goarch: "wasm", wantArch: plan9asm.ArchWASM, wantTriple: "wasm32-wasi"},
	}
	for _, test := range tests {
		t.Run(test.goos+"/"+test.goarch, func(t *testing.T) {
			arch, err := toPlan9Arch(test.goarch)
			if err != nil {
				t.Fatal(err)
			}
			if arch != test.wantArch {
				t.Fatalf("toPlan9Arch(%q) = %q, want %q", test.goarch, arch, test.wantArch)
			}
			if got := targetTriple(test.goos, test.goarch); got != test.wantTriple {
				t.Fatalf("targetTriple(%q, %q) = %q, want %q", test.goos, test.goarch, got, test.wantTriple)
			}
		})
	}
}

func TestLLVMArgsAndFrameSlotsForTupleSliceParam(t *testing.T) {
	tup := types.NewTuple(types.NewVar(token.NoPos, nil, "b", types.NewSlice(types.Typ[types.Byte])))
	sz := types.SizesFor("gc", "amd64")
	args, slots, nextOff, err := llvmArgsAndFrameSlotsForTuple(tup, "amd64", sz, 0, false)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(args, []plan9asm.LLVMType{"{ ptr, i64, i64 }"}) {
		t.Fatalf("args mismatch: %#v", args)
	}
	wantSlots := []plan9asm.FrameSlot{
		{Offset: 0, Type: plan9asm.Ptr, Index: 0, Field: 0},
		{Offset: 8, Type: plan9asm.I64, Index: 0, Field: 1},
		{Offset: 16, Type: plan9asm.I64, Index: 0, Field: 2},
	}
	if !reflect.DeepEqual(slots, wantSlots) {
		t.Fatalf("slots mismatch: got=%#v want=%#v", slots, wantSlots)
	}
	if nextOff != 24 {
		t.Fatalf("nextOff mismatch: got=%d want=24", nextOff)
	}
}

func TestLLVMArgsAndFrameSlotsForTupleSliceResultFlatten(t *testing.T) {
	tup := types.NewTuple(types.NewVar(token.NoPos, nil, "r", types.NewSlice(types.Typ[types.Byte])))
	sz := types.SizesFor("gc", "amd64")
	args, slots, nextOff, err := llvmArgsAndFrameSlotsForTuple(tup, "amd64", sz, 0, true)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(args, []plan9asm.LLVMType{plan9asm.Ptr, plan9asm.I64, plan9asm.I64}) {
		t.Fatalf("args mismatch: %#v", args)
	}
	wantSlots := []plan9asm.FrameSlot{
		{Offset: 0, Type: plan9asm.Ptr, Index: 0, Field: -1, Name: "r"},
		{Offset: 8, Type: plan9asm.I64, Index: 1, Field: -1, Name: "r"},
		{Offset: 16, Type: plan9asm.I64, Index: 2, Field: -1, Name: "r"},
	}
	if !reflect.DeepEqual(slots, wantSlots) {
		t.Fatalf("slots mismatch: got=%#v want=%#v", slots, wantSlots)
	}
	if nextOff != 24 {
		t.Fatalf("nextOff mismatch: got=%d want=24", nextOff)
	}
}

func TestLLVMArgsAndFrameSlotsForTupleInterfaceParam(t *testing.T) {
	iface := types.NewInterfaceType(nil, nil)
	iface.Complete()
	tup := types.NewTuple(types.NewVar(token.NoPos, nil, "v", iface))
	sz := types.SizesFor("gc", "amd64")
	args, slots, nextOff, err := llvmArgsAndFrameSlotsForTuple(tup, "amd64", sz, 0, false)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(args, []plan9asm.LLVMType{"{ ptr, ptr }"}) {
		t.Fatalf("args mismatch: %#v", args)
	}
	wantSlots := []plan9asm.FrameSlot{
		{Offset: 0, Type: plan9asm.Ptr, Index: 0, Field: 0},
		{Offset: 8, Type: plan9asm.Ptr, Index: 0, Field: 1},
	}
	if !reflect.DeepEqual(slots, wantSlots) {
		t.Fatalf("slots mismatch: got=%#v want=%#v", slots, wantSlots)
	}
	if nextOff != 16 {
		t.Fatalf("nextOff mismatch: got=%d want=16", nextOff)
	}
}

func TestLLVMArgsAndFrameSlotsForTupleNamedInterfaceParam(t *testing.T) {
	iface := types.NewInterfaceType(nil, nil)
	iface.Complete()
	named := types.NewNamed(types.NewTypeName(token.NoPos, nil, "Reader", nil), iface, nil)
	tup := types.NewTuple(types.NewVar(token.NoPos, nil, "v", named))
	sz := types.SizesFor("gc", "amd64")
	args, slots, _, err := llvmArgsAndFrameSlotsForTuple(tup, "amd64", sz, 0, false)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(args, []plan9asm.LLVMType{"{ ptr, ptr }"}) {
		t.Fatalf("args mismatch: %#v", args)
	}
	wantSlots := []plan9asm.FrameSlot{
		{Offset: 0, Type: plan9asm.Ptr, Index: 0, Field: 0},
		{Offset: 8, Type: plan9asm.Ptr, Index: 0, Field: 1},
	}
	if !reflect.DeepEqual(slots, wantSlots) {
		t.Fatalf("slots mismatch: got=%#v want=%#v", slots, wantSlots)
	}
}

func TestLLVMArgsAndFrameSlotsForNestedAggregateParam(t *testing.T) {
	slice := types.NewSlice(types.Typ[types.Byte])
	array := types.NewArray(slice, 2)
	tup := types.NewTuple(types.NewVar(token.NoPos, nil, "blocks", array))
	sz := types.SizesFor("gc", "amd64")
	args, slots, nextOff, err := llvmArgsAndFrameSlotsForTuple(tup, "amd64", sz, 0, false)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(args, []plan9asm.LLVMType{"[2 x { ptr, i64, i64 }]"}) {
		t.Fatalf("args mismatch: %#v", args)
	}
	wantSlots := []plan9asm.FrameSlot{
		{Offset: 0, Type: plan9asm.Ptr, Index: 0, Field: 0, Fields: []int{0, 0}},
		{Offset: 8, Type: plan9asm.I64, Index: 0, Field: 0, Fields: []int{0, 1}},
		{Offset: 16, Type: plan9asm.I64, Index: 0, Field: 0, Fields: []int{0, 2}},
		{Offset: 24, Type: plan9asm.Ptr, Index: 0, Field: 1, Fields: []int{1, 0}},
		{Offset: 32, Type: plan9asm.I64, Index: 0, Field: 1, Fields: []int{1, 1}},
		{Offset: 40, Type: plan9asm.I64, Index: 0, Field: 1, Fields: []int{1, 2}},
	}
	if !reflect.DeepEqual(slots, wantSlots) {
		t.Fatalf("slots mismatch: got=%#v want=%#v", slots, wantSlots)
	}
	if nextOff != 48 {
		t.Fatalf("nextOff mismatch: got=%d want=48", nextOff)
	}
}

func TestLLVMArgsAndFrameSlotsForStructParam(t *testing.T) {
	st := types.NewStruct([]*types.Var{
		types.NewVar(token.NoPos, nil, "A", types.NewArray(types.Typ[types.Uint16], 3)),
		types.NewVar(token.NoPos, nil, "B", types.Typ[types.Byte]),
		types.NewVar(token.NoPos, nil, "C", types.Typ[types.String]),
	}, nil)
	tup := types.NewTuple(types.NewVar(token.NoPos, nil, "v", st))
	sz := types.SizesFor("gc", "amd64")
	args, slots, nextOff, err := llvmArgsAndFrameSlotsForTuple(tup, "amd64", sz, 0, false)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(args, []plan9asm.LLVMType{"{ [3 x i16], i8, { ptr, i64 } }"}) {
		t.Fatalf("args mismatch: %#v", args)
	}
	wantSlots := []plan9asm.FrameSlot{
		{Offset: 0, Type: plan9asm.I16, Index: 0, Field: 0, Fields: []int{0, 0}},
		{Offset: 2, Type: plan9asm.I16, Index: 0, Field: 0, Fields: []int{0, 1}},
		{Offset: 4, Type: plan9asm.I16, Index: 0, Field: 0, Fields: []int{0, 2}},
		{Offset: 6, Type: plan9asm.I8, Index: 0, Field: 1},
		{Offset: 8, Type: plan9asm.Ptr, Index: 0, Field: 2, Fields: []int{2, 0}},
		{Offset: 16, Type: plan9asm.I64, Index: 0, Field: 2, Fields: []int{2, 1}},
	}
	if !reflect.DeepEqual(slots, wantSlots) {
		t.Fatalf("slots mismatch: got=%#v want=%#v", slots, wantSlots)
	}
	if nextOff != 24 {
		t.Fatalf("nextOff mismatch: got=%d want=24", nextOff)
	}
}

func TestLLVMArgsAndFrameSlotsFlattensArrayResult(t *testing.T) {
	array := types.NewArray(types.Typ[types.Byte], 7)
	tup := types.NewTuple(types.NewVar(token.NoPos, nil, "result", array))
	sz := types.SizesFor("gc", "amd64")
	args, slots, nextOff, err := llvmArgsAndFrameSlotsForTuple(tup, "amd64", sz, 0, true)
	if err != nil {
		t.Fatal(err)
	}
	wantArgs := []plan9asm.LLVMType{
		plan9asm.I8, plan9asm.I8, plan9asm.I8, plan9asm.I8,
		plan9asm.I8, plan9asm.I8, plan9asm.I8,
	}
	if !reflect.DeepEqual(args, wantArgs) {
		t.Fatalf("args mismatch: got=%#v want=%#v", args, wantArgs)
	}
	for i, slot := range slots {
		want := plan9asm.FrameSlot{Offset: int64(i), Type: plan9asm.I8, Index: i, Field: -1, Name: "result"}
		if !reflect.DeepEqual(slot, want) {
			t.Fatalf("slot %d = %#v, want %#v", i, slot, want)
		}
	}
	if nextOff != 7 {
		t.Fatalf("nextOff mismatch: got=%d want=7", nextOff)
	}
}

func TestLLVMArgsAndFrameSlotsForComplexParamsAndResults(t *testing.T) {
	for _, goarch := range []string{"386", "amd64", "arm", "arm64"} {
		t.Run(goarch, func(t *testing.T) {
			tup := types.NewTuple(
				types.NewVar(token.NoPos, nil, "c64", types.Typ[types.Complex64]),
				types.NewVar(token.NoPos, nil, "c128", types.Typ[types.Complex128]),
			)
			sz := types.SizesFor("gc", goarch)
			args, slots, nextOff, err := llvmArgsAndFrameSlotsForTuple(tup, goarch, sz, 0, false)
			if err != nil {
				t.Fatal(err)
			}
			wantArgs := []plan9asm.LLVMType{"{ float, float }", "{ double, double }"}
			if !reflect.DeepEqual(args, wantArgs) {
				t.Fatalf("parameter args mismatch: got=%#v want=%#v", args, wantArgs)
			}
			wantSlots := []plan9asm.FrameSlot{
				{Offset: 0, Type: plan9asm.LLVMType("float"), Index: 0, Field: 0},
				{Offset: 4, Type: plan9asm.LLVMType("float"), Index: 0, Field: 1},
				{Offset: 8, Type: plan9asm.LLVMType("double"), Index: 1, Field: 0},
				{Offset: 16, Type: plan9asm.LLVMType("double"), Index: 1, Field: 1},
			}
			if !reflect.DeepEqual(slots, wantSlots) {
				t.Fatalf("parameter slots mismatch: got=%#v want=%#v", slots, wantSlots)
			}
			if nextOff != 24 {
				t.Fatalf("parameter nextOff = %d, want 24", nextOff)
			}

			results, resultSlots, resultEnd, err := llvmArgsAndFrameSlotsForTuple(tup, goarch, sz, 0, true)
			if err != nil {
				t.Fatal(err)
			}
			wantResults := []plan9asm.LLVMType{"float", "float", "double", "double"}
			if !reflect.DeepEqual(results, wantResults) {
				t.Fatalf("flattened result args mismatch: got=%#v want=%#v", results, wantResults)
			}
			wantResultSlots := []plan9asm.FrameSlot{
				{Offset: 0, Type: plan9asm.LLVMType("float"), Index: 0, Field: -1, Name: "c64"},
				{Offset: 4, Type: plan9asm.LLVMType("float"), Index: 1, Field: -1, Name: "c64"},
				{Offset: 8, Type: plan9asm.LLVMType("double"), Index: 2, Field: -1, Name: "c128"},
				{Offset: 16, Type: plan9asm.LLVMType("double"), Index: 3, Field: -1, Name: "c128"},
			}
			if !reflect.DeepEqual(resultSlots, wantResultSlots) {
				t.Fatalf("result slots mismatch: got=%#v want=%#v", resultSlots, wantResultSlots)
			}
			if resultEnd != 24 {
				t.Fatalf("result nextOff = %d, want 24", resultEnd)
			}
		})
	}
}
