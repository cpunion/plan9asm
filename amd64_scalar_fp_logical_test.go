package plan9asm

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

// Independent Go yxm/ymm and AVX _yvaddsubpd/_yvandnpd/_yvblendmpd
// witnesses. Do not derive the private-consumer denominator from its registry.
var x86ScalarFPLogicalOpcodes = []string{
	"ANDPS", "ANDPD", "ANDNPS", "ANDNPD", "ORPS", "ORPD", "XORPS", "XORPD",
	"PAND", "PANDN", "POR", "PXOR",
	"VANDPS", "VANDPD", "VANDNPS", "VANDNPD", "VORPS", "VORPD", "VXORPS", "VXORPD",
	"VPAND", "VPANDN", "VPOR", "VPXOR",
	"VPANDD", "VPANDQ", "VPANDND", "VPANDNQ", "VPORD", "VPORQ", "VPXORD", "VPXORQ",
}

func x86ScalarFPLogicalRow(op string) (row string, destination int) {
	if strings.HasPrefix(op, "V") {
		return op + " X0,X1,X2", 2
	}
	return op + " X0,X1", 1
}

func TestX86ScalarFPLogicalCompleteXFormsThreeAPIs(t *testing.T) {
	llc := findLLVM22Tool("llc")
	if llc == "" {
		t.Fatal("LLVM 22 llc required")
	}
	if len(x86ScalarFPLogicalOpcodes) != 32 {
		t.Fatal("independent Go logical family denominator changed")
	}
	var source, declarations strings.Builder
	for index, op := range x86ScalarFPLogicalOpcodes {
		row, destination := x86ScalarFPLogicalRow(op)
		fmt.Fprintf(&source, "TEXT ·logical%d(SB),4,$0-8\n", index)
		source.WriteString("MOVL $1,AX\nMOVL AX,X0\nMOVL $3,BX\nMOVL BX,X1\n")
		fmt.Fprintf(&source, "%s\nMOVQ X%d,ret+0(FP)\nRET\n", row, destination)
		fmt.Fprintf(&declarations, "func logical%d() uint64\n", index)
	}
	for _, target := range x86RawReturnTargets {
		t.Run(target.triple, func(t *testing.T) {
			x86LEAActualTargetGoObject(t, target.arch, target.triple, source.String())
			for _, api := range []string{"text", "module", "Go"} {
				t.Run(api, func(t *testing.T) {
					ir, err := x86ScalarFPBitViewAPI(t, source.String(), declarations.String(), target.arch, target.triple, api)
					if err != nil || ir == "" {
						t.Fatalf("complete logical inputs lost result initialization: %v", err)
					}
					compileLLVMToObject(t, llc, target.triple, "logical-results-"+api+".ll", "logical-results-"+api+".o", ir)
				})
			}
		})
	}
}

func TestX86ScalarFPLogicalEveryByteAndExactZeroIdiom(t *testing.T) {
	for _, arch := range []string{"386", "amd64"} {
		for _, op := range x86ScalarFPLogicalOpcodes {
			t.Run(arch+"/"+op, func(t *testing.T) {
				row, destination := x86ScalarFPLogicalRow(op)
				file, err := Parse(ArchAMD64, "TEXT f(SB),4,$0-0\n"+row+"\nRET\n")
				if err != nil {
					t.Fatal(err)
				}
				ins := file.Funcs[0].Instrs[1]
				for input := 0; input < 2; input++ {
					for missing := -1; missing < 16; missing++ {
						state := x86ResultState{flags: x86ByteFlagsAll}
						state.xmm[0], state.xmm[1] = 0xffff, 0xffff
						if missing >= 0 {
							state.xmm[input] &^= uint16(1) << uint(missing)
						}
						want := state.xmm[0] & state.xmm[1]
						before := state.clone()
						if !x86ScalarFPResultTransfer(ins, &state, nil, 0, arch) || state.xmm[destination] != want {
							t.Fatalf("%s input%d missing%d: initialized bytes=%04x, want %04x", row, input, missing, state.xmm[destination], want)
						}
						before.xmm[destination] = want
						if !reflect.DeepEqual(before, state) {
							t.Fatal("logical definedness changed unrelated GP/FP/flags/mask/pointer state")
						}
					}
				}
				zero := strings.Contains(op, "XOR") || strings.Contains(op, "ANDN")
				if strings.HasPrefix(op, "V") {
					ins.Args[1] = ins.Args[0]
				} else {
					ins.Args[0] = ins.Args[1]
				}
				state := x86ResultState{}
				want := uint16(0)
				if zero {
					want = 0xffff
				}
				if !x86ScalarFPResultTransfer(ins, &state, nil, 0, arch) || state.xmm[destination] != want {
					t.Fatalf("exact register zero idiom %s: got %04x want %04x", op, state.xmm[destination], want)
				}
			})
		}
	}
}

func TestX86ScalarFPLogicalKeepsUnknownResultAndNativeMemoryBoundaries(t *testing.T) {
	for name, body := range map[string]string{
		"unknown input":       "PXOR X0,X1\nMOVQ X1,ret+0(FP)",
		"partial low qword":   "MOVSS $1.0,X2\nMOVSS X2,X0\nPXOR X0,X1\nMOVQ X1,ret+0(FP)",
		"native pointer read": "MOVQ $1,AX\nPXOR (AX),X0\nMOVQ X0,ret+0(FP)",
		"unknown global read": "PXOR input(SB),X0\nMOVQ X0,ret+0(FP)",
		"missing result":      "PXOR X0,X0\nMOVQ X0,output(SB)",
		"unknown global data": "MOVQ X0,output(SB)\nMOVQ X0,ret+0(FP)",
		"unknown merge":       "PXOR X0,X0\nMOVL $1,AX\nKMOVW AX,K1\nVPXORD X0,X0,K1,X2\nMOVQ X2,ret+0(FP)",
		"unknown mask":        "PXOR X0,X0\nVPXORD.Z X0,X0,K1,X2\nMOVQ X2,ret+0(FP)",
		"wider bank":          "VPXOR Y0,Y0,Y1\nMOVQ X1,ret+0(FP)",
	} {
		t.Run(name, func(t *testing.T) {
			source := "TEXT ·view(SB),4,$0-8\n" + body + "\nRET\n"
			goAssemblerOracleOutput(t, "amd64", source, "", true)
			for _, api := range []string{"text", "module", "Go"} {
				ir, err := x86ScalarFPBitViewAPI(t, source, "func view() uint64\n", "amd64", "x86_64-unknown-linux-gnu", api)
				if !errors.Is(err, ErrProbeNeedsContext) || ir != "" {
					t.Fatalf("%s acquired result or memory permission: %v\n%s", api, err, ir)
				}
			}
		})
	}
}

func TestX86ScalarFPLogicalNumericalSBStoreKeepsIndependentResult(t *testing.T) {
	const source = "GLOBL slot<>(SB),16,$8\nTEXT ·view(SB),4,$0-8\nPXOR X0,X0\nMOVQ X0,slot<>(SB)\nMOVQ X0,ret+0(FP)\nRET\n"
	llc := findLLVM22Tool("llc")
	if llc == "" {
		t.Fatal("LLVM 22 llc required")
	}
	for _, target := range x86RawReturnTargets {
		t.Run(target.triple, func(t *testing.T) {
			x86LEAActualTargetGoObject(t, target.arch, target.triple, source)
			for _, api := range []string{"text", "module", "Go"} {
				t.Run(api, func(t *testing.T) {
					ir, err := x86ScalarFPBitViewAPI(t, source, "func view() uint64\n", target.arch, target.triple, api)
					if err != nil || !strings.Contains(ir, "store volatile i64") {
						t.Fatalf("numeric SB side effect lost before independent result: %v", err)
					}
					compileLLVMToObject(t, llc, target.triple, "logical-sb-"+api+".ll", "logical-sb-"+api+".o", ir)
				})
			}
		})
	}
}
