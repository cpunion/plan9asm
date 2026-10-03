package plan9asm

import (
	"fmt"
	"strings"
	"testing"
)

func TestX86VectorZeroCompleteGoFormsFiveObjects(t *testing.T) {
	llc := findLLVM22Tool("llc")
	if llc == "" {
		t.Fatal("LLVM22 llc required")
	}
	const source = "TEXT zeroforms(SB),4,$0-0\nVZEROUPPER\nVZEROALL\nWORD $0xf8c5\nBYTE $0x77\nWORD $0xfcc5\nBYTE $0x77\nRET\n"
	for _, target := range x86RawReturnTargets {
		t.Run(target.triple, func(t *testing.T) {
			requireX86GoAssemblerResult(t, target.arch, source, true)
			file, err := Parse(ArchAMD64, source)
			if err != nil {
				t.Fatal(err)
			}
			ir, err := Translate(file, Options{Goarch: target.arch, TargetTriple: target.triple, Sigs: map[string]FuncSig{"zeroforms": {Ret: Void}}})
			if err != nil {
				t.Fatal(err)
			}
			compileLLVMToObject(t, llc, target.triple, "vector-zero.ll", "vector-zero.o", ir)
		})
	}
}

func TestX86VectorZeroRejectsMalformedGoFormsBeforeEmission(t *testing.T) {
	for _, arch := range []string{"amd64", "386"} {
		for _, opcode := range []string{"VZEROUPPER", "VZEROALL"} {
			for _, tail := range []string{" X0", " AX", " K1", " $1", ".Z", ".BCST", ".SAE", ".RN_SAE"} {
				row := opcode + tail
				t.Run(arch+"/"+row, func(t *testing.T) {
					source := "TEXT f(SB),4,$0-0\n" + row + "\nRET\n"
					requireX86GoAssemblerResult(t, arch, source, false)
					file, err := Parse(ArchAMD64, source)
					if err != nil {
						t.Fatal(err)
					}
					ctx, output := newAMD64CtxWithFuncForTest(t, file.Funcs[0], FuncSig{Ret: Void}, nil)
					ctx.goarch = arch
					before := output.String()
					ok, _, err := ctx.lowerVec(file.Funcs[0].Instrs[1].Op, file.Funcs[0].Instrs[1])
					if !ok || err == nil || output.String() != before {
						t.Fatalf("malformed zero form emitted/accepted: %t/%v", ok, err)
					}
				})
			}
		}
	}
}

func TestX86VectorZeroRawOpcodeLengthIdentity(t *testing.T) {
	for _, mode := range []int{32, 64} {
		for _, prefix := range [][]byte{{0xc5, 0xf8}, {0xc5, 0xfc}, {0xc4, 0xe1, 0x78}, {0xc4, 0xe1, 0x7c}, {0xc4, 0xe1, 0xf8}, {0xc4, 0xe1, 0xfc}} {
			code := append(append([]byte(nil), prefix...), 0x77)
			t.Run(fmt.Sprintf("%d/%x", mode, code), func(t *testing.T) {
				instructions, err := decodeX86RawDirectiveGroup(code, mode, 0, "actual vector-zero bytes", nil)
				if err != nil || len(instructions) != 1 {
					t.Fatalf("raw zero failed: %+v/%v", instructions, err)
				}
				want := Op("VZEROUPPER")
				if prefix[len(prefix)-1]&4 != 0 {
					want = "VZEROALL"
				}
				if instructions[0].Op != want || len(instructions[0].Args) != 0 {
					t.Fatalf("raw %x became %s, want %s without operands", code, instructions[0].Op, want)
				}
			})
		}
	}
}

func TestX86VectorZeroReservedRawFormsFailClosed(t *testing.T) {
	for _, mode := range []int{32, 64} {
		for _, code := range [][]byte{
			{0xc5, 0xf9, 0x77}, {0xc5, 0xfb, 0x77}, {0xc5, 0xe8, 0x77},
			{0xc4, 0xe1, 0xf9, 0x77}, {0xc4, 0xe1, 0xec, 0x77},
			{0x66, 0xc5, 0xf8, 0x77}, {0xf2, 0xc5, 0xfc, 0x77}, {0xf0, 0xc5, 0xf8, 0x77},
			{0xc5}, {0xc5, 0xf8}, {0xc4, 0xe1, 0xf8},
		} {
			t.Run(fmt.Sprintf("%d/%x", mode, code), func(t *testing.T) {
				instructions, err := decodeX86RawDirectiveGroup(code, mode, 0, "reserved zero encoding", nil)
				if err == nil {
					t.Fatalf("reserved/truncated raw %x accepted: %+v", code, instructions)
				}
				if strings.Contains(fmt.Sprint(instructions), "EMMS") {
					t.Fatalf("VEX zero was silently decoded as legacy EMMS: %+v", instructions)
				}
			})
		}
	}
}
