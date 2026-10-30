package plan9asm

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

// This derives an ordinary language signature, not a compiler/native entry,
// a physical frame, a pointer extent or any initialized memory contents.
func amd64RegisterVoidSource(t *testing.T, declaration, body string) (Func, FuncSig, string) {
	t.Helper()
	_, sig, source := x86ScalarFPVoidSource(t, declaration, body, "amd64")
	source = strings.Replace(source, "·view(SB)", "·view<ABIInternal>(SB)", 1)
	file, err := Parse(ArchAMD64, source)
	if err != nil || len(file.Funcs) != 1 {
		t.Fatal("ordinary register-void fixture lost its complete source", err)
	}
	return file.Funcs[0], sig, source
}

func TestAMD64RegisterVoidKeepsDeclaredInputsAndFullFlow(t *testing.T) {
	registers := []Reg{AX, BX, CX, DI, SI, "R8", "R9", "R10", "R11"}
	cases := 0
	for _, width := range []struct {
		typ, move string
		bytes     int64
	}{
		{"uint8", "MOVB", 1}, {"uint16", "MOVW", 2},
		{"uint32", "MOVL", 4}, {"uint64", "MOVQ", 8},
	} {
		for index, register := range registers {
			t.Run(width.typ+"/"+string(register), func(t *testing.T) {
				declaration := "func view(" + strings.TrimSuffix(strings.Repeat(width.typ+",", index+1), ",") + ")\n"
				body := fmt.Sprintf("%s %s,R12", width.move, register)
				fn, sig, _ := amd64RegisterVoidSource(t, declaration, body)
				seen := 0
				err := checkAMD64RegisterResultInitialization(fn, sig,
					func(_ int, ins Instr, state *x86ResultState, owned []bool, resultOffset int64, arch string) bool {
						if ins.Raw != body {
							t.Fatal("unexpected source effect in the input counterexample", ins.Raw)
						}
						seen++
						for carrier, reg := range registers {
							position, _, valid := x86ResultGP(reg, width.bytes, arch)
							if !valid {
								t.Fatal("invalid original argument carrier", reg)
							}
							want := uint8(0)
							if carrier <= index {
								want = uint8(1<<uint(width.bytes)) - 1
							}
							if state.gp[position] != want || state.pointerGP[position] {
								t.Fatal("Void changed declared carrier bytes or pointer roles", reg, state.gp[position], want)
							}
						}
						return x86ScalarFPResultTransfer(ins, state, owned, resultOffset, arch)
					})
				if err != nil || seen != 1 {
					t.Fatal("ordinary Void must consume its actual input/effect without an AX result", seen, err)
				}
			})
			cases++
		}
	}
	if cases != 36 {
		t.Fatal("incomplete ordinary argument-width/carrier matrix", cases)
	}

	t.Run("pointer-is-not-numerical", func(t *testing.T) {
		fn, sig, _ := amd64RegisterVoidSource(t, "func view(input *uint64)\n", "MOVQ AX,BX\nMOVQ BX,CX")
		seen := 0
		err := checkAMD64RegisterResultInitialization(fn, sig,
			func(_ int, ins Instr, state *x86ResultState, owned []bool, resultOffset int64, arch string) bool {
				register := AX
				if seen != 0 {
					register = BX
				}
				position, _, _ := x86ResultGP(register, 8, arch)
				if !state.pointerGP[position] || !state.callerAddressGP[position] || state.gp[position] != 0 {
					t.Fatal("Void discarded the original opaque pointer role or invented numerical bytes")
				}
				seen++
				return x86ScalarFPResultTransfer(ins, state, owned, resultOffset, arch)
			})
		if err != nil || seen != 2 {
			t.Fatal("Void pointer copies require the complete original flow", seen, err)
		}
	})

	t.Run("no-AX-result", func(t *testing.T) {
		fn, sig, _ := amd64RegisterVoidSource(t, "func view()\n", "MOVQ $7,BX")
		seen := 0
		err := checkAMD64RegisterResultInitialization(fn, sig,
			func(_ int, ins Instr, state *x86ResultState, owned []bool, resultOffset int64, arch string) bool {
				position, _, _ := x86ResultGP(AX, 8, arch)
				if state.gp[position] != 0 {
					t.Fatal("the Void entry fabricated AX initialization")
				}
				seen++
				return x86ScalarFPResultTransfer(ins, state, owned, resultOffset, arch)
			})
		if err != nil || seen != 1 {
			t.Fatal("Void has no AX return obligation but retains the source effect", seen, err)
		}
	})
}

func TestAMD64RegisterVoidWholeFunctionThreeAPIsAndObjects(t *testing.T) {
	llc := findLLVM22Tool("llc")
	if llc == "" {
		t.Fatal("LLVM 22 llc is required")
	}
	var assembly, declarations strings.Builder
	fixtures := 0
	for index, fixture := range []struct {
		declaration, body string
	}{
		{"func view(input uint8)\n", "MOVB AX,R12"},
		{"func view(input uint16)\n", "MOVW AX,R12"},
		{"func view(input uint32)\n", "MOVL AX,R12"},
		{"func view(input uint64)\n", "MOVQ AX,R12"},
		{"func view(input *uint64)\n", "MOVQ AX,BX\nMOVQ BX,CX"},
		{"func view()\n", "MOVQ $7,BX"},
		{"func view(input *uint64)\n", "MOVQ (AX),BX"},
	} {
		_, _, source := amd64RegisterVoidSource(t, fixture.declaration, fixture.body)
		name := fmt.Sprintf("consume%d", index)
		assembly.WriteString(strings.Replace(source, "·view<", "·"+name+"<", 1))
		declarations.WriteString(strings.Replace(fixture.declaration, "view(", name+"(", 1))
		fixtures++
	}
	if fixtures != 7 {
		t.Fatal("incomplete integer/pointer/empty/native-read Void source matrix", fixtures)
	}
	requireAMD64RegisterResultGoObject(t, assembly.String())
	for _, triple := range []string{"x86_64-unknown-linux-gnu", "x86_64-apple-darwin", "x86_64-pc-windows-msvc"} {
		for _, api := range []string{"text", "module", "Go"} {
			t.Run(triple+"/"+api, func(t *testing.T) {
				ir, err := x86ScalarFPBitViewAPI(t, assembly.String(), declarations.String(), "amd64", triple, api)
				if err != nil || ir == "" {
					t.Fatal("ordinary typed Void entry requires no fabricated result", err)
				}
				for _, level := range []string{"-O0", "-O2"} {
					text := ir
					if level == "-O2" {
						text = x86MulDivRoundOptimize(t, text)
					}
					if strings.Count(text, "load volatile i64,") != 1 {
						t.Fatal("Void discarded or duplicated its exact native qword read", api, level)
					}
					x86PackedStringCompileObject(t, llc, triple, text, level)
				}
			})
		}
	}
}

func TestAMD64RegisterVoidCannotWaiveFaultingInputsAndSourceRoles(t *testing.T) {
	for _, fixture := range []struct {
		name, declaration, body string
	}{
		{"undefined-base", "func view(input uint64)\n", "MOVQ (R10),AX"},
		{"NUM-is-not-pointer", "func view(input uint64)\n", "MOVQ (AX),BX"},
		{"hidden-closure", "func view(input uint64)\n", "MOVQ 8(DX),BX"},
		{"hidden-g", "func view(input uint64)\n", "MOVQ (R14),AX"},
		{"physical-stack", "func view(input uint64)\n", "MOVQ (SP),AX"},
		{"ABI0-input", "func view(input uint64)\n", "MOVQ input+0(FP),AX"},
		{"ABI0-store", "func view(input uint64)\n", "MOVQ AX,input+0(FP)"},
		{"unbound-call", "func view(input uint64)\n", "CALL ·unknown(SB)"},
		{"one-path-fault", "func view(input uint64)\n", "CMPQ AX,$0\nJE done\nMOVQ (R10),BX\ndone:"},
	} {
		t.Run(fixture.name, func(t *testing.T) {
			_, _, source := amd64RegisterVoidSource(t, fixture.declaration, fixture.body)
			for _, api := range []string{"text", "module", "Go"} {
				ir, err := x86ScalarFPBitViewAPI(t, source, fixture.declaration, "amd64", "x86_64-unknown-linux-gnu", api)
				if !errors.Is(err, ErrProbeNeedsContext) || ir != "" {
					t.Error("Void waived a faulting read or its original source/entry role", api, err)
				}
			}
		})
	}
}

func TestAMD64RegisterVoidRejectsForgedResultsAndEntries(t *testing.T) {
	for _, change := range []string{"ABI0", "manual-registers", "result-without-slot", "slot-on-Void", "raw", "physical-address-mode", "WASM-entry"} {
		t.Run(change, func(t *testing.T) {
			fn, sig, _ := amd64RegisterVoidSource(t, "func view(input uint64)\n", "MOVQ AX,BX")
			switch change {
			case "ABI0":
				fn.Sym = strings.TrimSuffix(fn.Sym, "<ABIInternal>")
			case "manual-registers":
				sig.ArgRegs = []Reg{AX}
			case "result-without-slot":
				sig.Ret = I64
			case "slot-on-Void":
				sig.Frame.Results = append(sig.Frame.Results, sig.Frame.Params[0])
			case "raw":
				fn.Instrs[1].x86Encoded = true
			case "physical-address-mode":
				fn.Instrs[1].x86AddressBits = 32
			case "WASM-entry":
				sig.WASMContext = "go"
			}
			calls := 0
			err := checkAMD64RegisterResultInitialization(fn, sig,
				func(_ int, ins Instr, state *x86ResultState, owned []bool, resultOffset int64, arch string) bool {
					calls++
					return x86ScalarFPResultTransfer(ins, state, owned, resultOffset, arch)
				})
			if !errors.Is(err, ErrProbeNeedsContext) || calls != 0 {
				t.Fatal("Void admitted contradictory result/entry metadata", calls, err)
			}
		})
	}
}
