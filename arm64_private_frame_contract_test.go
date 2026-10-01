package plan9asm

import (
	"errors"
	"strings"
	"testing"
)

func TestCrossLinuxRuntimeMatrixARM64CallerFPAliasInvalidatesPointer(t *testing.T) {
	_, _, _, runner := arm64FPPairRuntimeTools(t)
	// aliasTarget's real Go frame is 48 bytes; its first FP pointer slot is
	// at entry SP+56. The caller explicitly points p at that pointer slot.
	// An aliased MOVD reconstructs p as the callee saved-LR address. A reload
	// then changes LR to the real caller continuation plus four bytes. Native
	// Go safely skips one witness instruction; synthesizing a normal return
	// would therefore have different semantics. No fabricated link is used.
	const target = `TEXT ·aliasTarget(SB),4,$32-8
CALL ·aliasAnchor(SB)
MOVD (RSP),R22
MOVD p+0(FP),R9
SUB $56,R9,R10
MOVD R10,(R9)
MOVD p+0(FP),R11
ADD $4,R22,R12
MOVD R12,(R11)
RET
`
	const anchor = "TEXT ·aliasAnchor(SB),4,$0-0\nRET\n"
	const caller = `TEXT ·aliasMeasure(SB),516,$0-8
MOVD RSP,R19
MOVD R30,R23
MOVD R29,R24
SUB $64,RSP
ADD $8,RSP,R9
MOVD R9,8(RSP)
MOVD $7,R25
BL ·aliasTarget(SB)
MOVD $11,R25
ADD $100,R25
MOVD $0,R9
MOVD R9,8(RSP)
MOVD R19,RSP
MOVD R24,R29
MOVD R23,R30
MOVD R25,ret+0(FP)
RET
`
	requireARM64GoAssemblerResult(t, caller+target+anchor, true)
	runARM64LocalRegisterGoOracle(t, caller+target+anchor, `package main
func aliasMeasure() uint64
func aliasTarget(*uint64)
func aliasAnchor()
func main() { if got:=aliasMeasure(); got!=107 { println(got); panic("caller FP alias witness mismatch") } }
`, len(runner) != 0)
	pkg := mustGoPackage(t, "test/fpalias", "package fpalias\nfunc aliasTarget(*uint64)\nfunc aliasAnchor()\n")
	tr, err := TranslateGoModule(pkg, []byte(target+anchor), GoModuleOptions{
		GOARCH: "arm64", ResolveSym: func(sym string) string { return strings.TrimPrefix(sym, "·") },
	})
	if err == nil {
		tr.Module.Dispose()
	}
	if !errors.Is(err, ErrProbeNeedsContext) {
		t.Fatalf("aliased FP mutation cannot retain the original typed Ptr category: %v", err)
	}
	t.Log("actual native Go alias redirects caller continuation by one instruction; translation must retain Context")
}

func TestCrossLinuxRuntimeMatrixARM64TypedPrivateFrameNeedsMemoryContract(t *testing.T) {
	_, _, _, runner := arm64FPPairRuntimeTools(t)
	for _, tc := range []struct {
		name, body, decl, goCallee, goChecks string
	}{
		{
			name:     "incoming_pointer_store",
			body:     "MOVD p+0(FP),R9\nMOVD a+8(FP),R10\nMOVD R10,(R9)\nCALL ·X(SB)",
			goChecks: "if data[0] != a { panic(\"typed data store mismatch\") }",
		},
		{
			name:     "declared_pointer_callee",
			body:     "MOVD p+0(FP),R9\nMOVD a+8(FP),R10\nMOVD R9,8(RSP)\nMOVD R10,16(RSP)\nCALL ·X(SB)\nMOVD p+0(FP),R9\nMOVD a+8(FP),R10\nMOVD R10,8(R9)",
			decl:     "func X(*uint64,uint64)",
			goCallee: "func X(p *uint64,a uint64) { *p=a*7+9 }",
			goChecks: "if data[0] != a*7+9 || data[1]!=a { panic(\"typed pointer callee mismatch\") }",
		},
		{
			name: "narrow_postindexed",
			body: "MOVD p+0(FP),R9\nMOVD a+8(FP),R10\nMOVH.P R10,2(R9)\nMOVW.P R10,4(R9)\nMOVD.P R10,8(R9)\nCALL ·X(SB)",
			goChecks: `bytes:=(*[32]byte)(unsafe.Pointer(&data[0]))
for j:=0;j<32;j++ {
  var want byte
  switch { case j<2: want=byte(a>>uint(j*8)); case j<6: want=byte(a>>uint((j-2)*8)); case j<14: want=byte(a>>uint((j-6)*8)) }
  if bytes[j]!=want { panic("typed postindex bytes mismatch") }
}`,
		},
		{
			name:     "dynamic_data_offset",
			body:     "MOVD p+0(FP),R9\nMOVD a+8(FP),R10\nAND $3,R10,R11\nLSL $3,R11\nADD R11,R9,R9\nMOVD R10,(R9)\nCALL ·X(SB)",
			goChecks: "for j,v:=range data { var want uint64; if j==int(a&3) { want=a }; if v!=want { panic(\"typed offset mismatch\") } }",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if tc.decl == "" {
				tc.decl, tc.goCallee = "func X()", "func X() {}"
			}
			source := "TEXT ·Y(SB),4,$48-24\n" + tc.body + "\nMOVD a+8(FP),R10\nADD $13,R10\nMOVD R10,ret+16(FP)\nRET\n"
			requireARM64GoAssemblerResult(t, source, true)
			imports := ""
			if strings.Contains(tc.goChecks, "unsafe.") {
				imports = "import \"unsafe\"\n"
			}
			goMain := "package main\n" + imports + "func Y(*uint64,uint64) uint64\n" + tc.goCallee + "\nfunc main() { for _,a:=range []uint64{0,1,123,0x8000000000000000,^uint64(0)} { var data [4]uint64; if Y(&data[0],a)!=a+13 { panic(\"typed frame result mismatch\") }; " + tc.goChecks + " } }\n"
			runARM64LocalRegisterGoOracle(t, source, goMain, len(runner) != 0)
			pkg := mustGoPackage(t, "test/privateframe", "package privateframe\nfunc Y(*uint64,uint64) uint64\n"+tc.decl+"\n")
			// Native Go establishes the supplied callsites' data semantics, but
			// the declaration alone does not prove in-bounds access or exclude
			// a caller FP alias. Every target must retain the missing contract.
			for _, target := range []string{
				"aarch64-unknown-linux-gnu", "aarch64-unknown-linux-musl",
				"aarch64-apple-darwin", "aarch64-pc-windows-msvc",
			} {
				tr, err := TranslateGoModule(pkg, []byte(source), GoModuleOptions{
					GOARCH: "arm64", TargetTriple: target,
					ResolveSym: func(sym string) string { return strings.TrimPrefix(sym, "·") },
				})
				if err == nil {
					tr.Module.Dispose()
				}
				if !errors.Is(err, ErrProbeNeedsContext) {
					t.Fatalf("%s: incomplete pointer bounds/alias contract: %v", target, err)
				}
			}
		})
	}
}
