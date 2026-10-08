package plan9asm

import (
	"errors"
	"fmt"
	"reflect"
	"testing"
)

func TestX86ABI0ScalarImportOpaquePointerIndependentResult(t *testing.T) {
	for _, arch := range []string{"amd64", "386"} {
		word := x86ResultPointerWord(arch)
		for name, body := range x86ResultIncomingPointerIdentityBodies(arch) {
			t.Run(arch+"/"+name, func(t *testing.T) {
				finish := fmt.Sprintf("MOVL $42,DX\nMOVL DX,X1\nPEXTRD $0,X1,BX\nMOVL BX,ret0+%d(FP)\nRET", word)
				_, file, sig, source := x86ResultInitializationFixture(t, arch,
					"func Value(x *[8]uint32) uint32", body+finish)
				goAssemblerOracleOutput(t, arch, source, "", true)
				before := fmt.Sprintf("%#v %#v", file.Funcs[0], sig)
				if err := checkX86ScalarFPResultInitialization(file.Funcs[0], sig, arch); err != nil {
					t.Fatalf("opaque pointer transport contaminated independent numerical result: %v", err)
				}
				if before != fmt.Sprintf("%#v %#v", file.Funcs[0], sig) {
					t.Fatal("import observation changed original source or declaration")
				}
			})
		}
	}
}

func TestX86ABI0ScalarImportSliceThreeAPIFiveTargetObjects(t *testing.T) {
	llc := findLLVM22Tool("llc")
	if llc == "" {
		t.Fatal("LLVM22 llc is required")
	}
	for _, target := range x86RawReturnTargets {
		t.Run(target.triple, func(t *testing.T) {
			word, move := 8, "MOVQ"
			if target.arch == "386" {
				word, move = 4, "MOVL"
			}
			// This independently derived gc declaration has the crc32 shape:
			// an integer followed by slice data/len/cap and one scalar result.
			// The pointer is transported, never dereferenced or used as bits.
			source := fmt.Sprintf("TEXT ·view(SB),4,$0-%d\nMOVL crc+0(FP),X0\n%s p+%d(FP),SI\n%s p_len+%d(FP),CX\n%s SI,DI\nPEXTRD $0,X0,AX\nMOVL AX,ret+%d(FP)\nRET\n",
				4*word+4, move, word, move, 2*word, move, 4*word)
			x86LEAActualTargetGoObject(t, target.arch, target.triple, source)
			for _, api := range []string{"text", "module", "Go"} {
				t.Run(api, func(t *testing.T) {
					ir, err := x86ScalarFPBitViewAPI(t, source,
						"func view(crc uint32,p []byte) uint32\n", target.arch, target.triple, api)
					if err != nil || ir == "" {
						t.Fatalf("declared scalar import with independent extracted result: %v", err)
					}
					compileLLVMToObject(t, llc, target.triple, "scalar-import.ll", "scalar-import.o", ir)
				})
			}
		})
	}
}

func TestX86ABI0ScalarImportDoesNotGrantNumericOrMemoryPermission(t *testing.T) {
	for _, arch := range []string{"amd64", "386"} {
		word := x86ResultPointerWord(arch)
		move, lea := "MOVQ", "LEAQ"
		if arch == "386" {
			move, lea = "MOVL", "LEAL"
		}
		load := move + " x+0(FP),AX\n"
		finish := fmt.Sprintf("MOVL $42,DX\nMOVL DX,X1\nPEXTRD $0,X1,BX\nMOVL BX,ret0+%d(FP)\nRET", word)
		for name, body := range map[string]string{
			"pointee":             load + "MOVL (AX),BX\n",
			"pointee vector":      load + "MOVOU (AX),X0\n",
			"pointer GP bits":     load + "MOVL AX,CX\n",
			"pointer X bits":      load + move + " AX,X0\n",
			"pointer FP bits":     "MOVL x+0(FP),X0\n",
			"pointer lane":        load + "PINSRD $0,AX,X0\n",
			"input store":         load + move + " AX,x+0(FP)\n",
			"result identity":     load + fmt.Sprintf("MOVL AX,ret0+%d(FP)\n", word),
			"unknown source":      move + " AX,CX\n",
			"unknown LEA index":   load + lea + " (AX)(CX*4),DI\n",
			"pointer LEA index":   load + lea + " (AX)(AX*4),DI\n",
			"unknown GP result":   load + "PINSRD $0,CX,X0\n",
			"unknown result lane": load + "MOVL $7,DX\nPINSRD $1,DX,X0\nPEXTRD $0,X0,BX\n",
		} {
			t.Run(arch+"/"+name, func(t *testing.T) {
				_, file, sig, source := x86ResultInitializationFixture(t, arch, "func Value(x *uint32) uint32", body+finish)
				goAssemblerOracleOutput(t, arch, source, "", true)
				if err := checkX86ScalarFPResultInitialization(file.Funcs[0], sig, arch); !errors.Is(err, ErrProbeNeedsContext) {
					t.Fatalf("opaque import granted numerical/native authority: %v", err)
				}
			})
		}
	}
	state := x86ResultState{pointerFP: map[int64]bool{0: true}, fp: make([]bool, 8)}
	owned := []bool{true, true, true, true, true, true, true, true}
	ins := x86YMMReadSource(t, "MOVQ x+0(FP),AX")
	if !x86ScalarFPResultTransfer(ins, &state, owned, 8, "amd64") || !state.pointerGP[0] || state.gp[0] != 0 ||
		!reflect.DeepEqual(state.fp, make([]bool, 8)) {
		t.Fatal("opaque import initialized numerical pointer bytes")
	}
}
