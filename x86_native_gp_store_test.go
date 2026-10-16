package plan9asm

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestX86NativeGPStoreCompleteWidthsThreeAPIs(t *testing.T) {
	llc := findLLVM22Tool("llc")
	if llc == "" {
		t.Fatal("LLVM22 llc required")
	}
	for _, target := range x86RawReturnTargets {
		word, move := 8, "MOVQ"
		if target.arch == "386" {
			word, move = 4, "MOVL"
		}
		for _, row := range []struct {
			op, register string
			width        int
		}{{"MOVB", "AL", 1}, {"MOVW", "AX", 2}, {"MOVL", "AX", 4}, {"MOVQ", "AX", 8}, {"MOVD", "AX", 8}, {"MOVDQ2Q", "AX", 8}} {
			if row.width > word {
				continue // Go386 has no ordinary qword GP-to-memory row.
			}
			source := fmt.Sprintf("TEXT ·Value(SB),4,$0-%d\n%s p+0(FP),DI\n%s value+%d(FP),AX\n%s %s,8(DI)\nMOVL $7,AX\nMOVL AX,ret+%d(FP)\nRET\n",
				word*2+4, move, move, word, row.op, row.register, word*2)
			x86LEAActualTargetGoObject(t, target.arch, target.triple, source)
			for _, api := range []string{"text", "module", "Go"} {
				t.Run(target.triple+"/"+row.op+"/"+api, func(t *testing.T) {
					ir, err := x86ScalarFPBitViewAPI(t, source, "func Value(p *byte,value uintptr) uint32", target.arch, target.triple, api)
					if err != nil || ir == "" {
						t.Fatal("complete numerical GP store lost its paired continuation", err)
					}
					if strings.Count(ir, fmt.Sprintf("store volatile i%d ", row.width*8)) != 1 {
						t.Fatal("the paired emitter did not retain exactly one full native store")
					}
					pkg := mustGoPackage(t, "store", "package store\nfunc Value(p *byte,value uintptr) uint32")
					file := mustParseX86RawReturn(t, source)
					sigs, err := goSigsForAsmFile(pkg, file, func(name string) string { return strings.TrimPrefix(name, "·") }, target.arch, nil, nil, nil)
					if err != nil {
						t.Fatal(err)
					}
					if err := CheckX86ABI0ResultInitialization(file.Funcs[0], sigs["Value"], target.arch); !errors.Is(err, ErrProbeNeedsContext) {
						t.Fatal("a declaration-only query acquired native write authority", err)
					}
					for _, level := range []string{"-O0", "-O2"} {
						x86PackedStringCompileObject(t, llc, target.triple, ir, level)
					}
				})
			}
		}
	}
}

func TestX86NativeGPStoreCannotBorrowUnknownOrPointerData(t *testing.T) {
	for _, target := range x86RawReturnTargets {
		word, move := 8, "MOVQ"
		if target.arch == "386" {
			word, move = 4, "MOVL"
		}
		body := fmt.Sprintf("%s p+0(FP),DI\n%s value+%d(FP),AX\n%s AX,(DI)\nMOVL $7,ret+%d(FP)\nRET\n", move, move, word, move, word*2)
		for _, mutation := range []struct{ name, body string }{
			{"unknown-address", strings.Replace(body, move+" p+0(FP),DI\n", "", 1)},
			{"numerical-address", strings.Replace(body, move+" p+0(FP),DI", fmt.Sprintf("%s value+%d(FP),DI", move, word), 1)},
			{"unknown-data", strings.Replace(body, fmt.Sprintf("%s value+%d(FP),AX\n", move, word), "", 1)},
			{"pointer-data", strings.Replace(body, fmt.Sprintf("%s value+%d(FP),AX", move, word), move+" p+0(FP),AX", 1)},
			{"partial-data", strings.Replace(body, fmt.Sprintf("%s value+%d(FP),AX", move, word), fmt.Sprintf("MOVB value+%d(FP),AL", word), 1)},
			{"raw-return", strings.Replace(body, "RET", "BYTE $0xc3\nRET", 1)},
		} {
			source := fmt.Sprintf("TEXT ·Value(SB),4,$0-%d\n%s", word*2+4, mutation.body)
			goAssemblerOracleOutput(t, target.arch, source, "", true)
			for _, api := range []string{"text", "module", "Go"} {
				t.Run(target.triple+"/"+mutation.name+"/"+api, func(t *testing.T) {
					if ir, err := x86ScalarFPBitViewAPI(t, source, "func Value(p *byte,value uintptr) uint32", target.arch, target.triple, api); !errors.Is(err, ErrProbeNeedsContext) || ir != "" {
						t.Fatal("unproved native write data/address/source was admitted", err)
					}
				})
			}
		}
	}
}
