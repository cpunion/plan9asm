package plan9asm

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

type x86ImplicitMultiplyCase struct {
	x86ScalarFPTupleSpanCase
	op    string
	width int
}

func x86ImplicitMultiplyCases(arch string) []x86ImplicitMultiplyCase {
	var cases []x86ImplicitMultiplyCase
	for _, stem := range []string{"MUL", "IMUL"} {
		for _, width := range []int{1, 2, 4, 8} {
			if arch == "386" && width == 8 {
				continue // No Go mode32 Q encoding, not a skipped target.
			}
			suffix := map[int]string{1: "B", 2: "W", 4: "L", 8: "Q"}[width]
			move, input, result, declaration := "MOV"+suffix, "7", 2*width, fmt.Sprintf("func Value() (uint%d,uint%d)", width*8, width*8)
			want := []uint64{21, 0}
			if stem == "IMUL" {
				input = "-2"
				mask := ^uint64(0)
				if width != 8 {
					mask = (uint64(1) << uint(width*8)) - 1
				}
				want = []uint64{mask - 5, mask}
			}
			stores := fmt.Sprintf("%s AX,ret+0(FP)\n%s DX,ret+%d(FP)\n", move, move, width)
			if width == 1 {
				declaration, result, stores = "func Value() uint16", 2, "MOVW AX,ret+0(FP)\n"
				want = []uint64{21}
				if stem == "IMUL" {
					want = []uint64{0xfffa}
				}
			}
			op := stem + suffix
			body := fmt.Sprintf("%s $%s,AX\n%s $3,BX\n%s BX\n%sRET\n", move, input, move, op, stores)
			cases = append(cases, x86ImplicitMultiplyCase{x86ScalarFPTupleSpanCase{
				name: op, declaration: declaration,
				source: fmt.Sprintf("TEXT ·Value(SB),4,$0-%d\n%s", result, body), want: want,
			}, op, width})
		}
	}
	return cases
}

func TestX86ImplicitMultiplyWholeSourceThreeAPIs(t *testing.T) {
	for _, target := range x86RawReturnTargets {
		for _, row := range x86ImplicitMultiplyCases(target.arch) {
			for _, api := range []string{"text", "module", "Go"} {
				for _, missing := range []string{"", "accumulator", "source"} {
					t.Run(target.triple+"/"+row.op+"/"+api+"/missing="+missing, func(t *testing.T) {
						source := row.source
						if missing != "" {
							lines := strings.Split(source, "\n")
							index := 1
							if missing == "source" {
								index = 2
							}
							lines[index] = "NOP"
							source = strings.Join(lines, "\n")
						}
						ir, err := x86ScalarFPBitViewAPI(t, source, row.declaration, target.arch, target.triple, api)
						if missing != "" {
							if !errors.Is(err, ErrProbeNeedsContext) || ir != "" {
								t.Fatalf("implicit operand was invented: output=%t error=%v", ir != "", err)
							}
							return
						}
						if err != nil || ir == "" {
							t.Fatal("complete implicit product rejected", err)
						}
						x86ScalarFPTupleSpanConstantResult(t, ir, row.want)
					})
				}
			}
		}
	}
}

func TestX86ImplicitMultiplyNUMBytesAndStatus(t *testing.T) {
	for _, arch := range []string{"386", "amd64"} {
		for _, row := range x86ImplicitMultiplyCases(arch) {
			t.Run(arch+"/"+row.op, func(t *testing.T) {
				ins := x86YMMReadSource(t, row.op+" BX")
				width := int64(row.width)
				input := uint8((uint16(1) << uint(width)) - 1)
				state := x86ResultState{flags: x86ByteFlagsAll}
				ax, _, _ := x86ResultGP(AX, width, arch)
				bx, _, _ := x86ResultGP(BX, width, arch)
				dx, _, _ := x86ResultGP(DX, width, arch)
				state.gp[ax], state.gp[bx], state.gp[dx] = input, input, 0x80
				before := state.clone()
				if !x86ScalarFPResultTransfer(ins, &state, nil, 0, arch) {
					t.Fatal("missing implicit numerical effect")
				}
				wantAX, wantDX := input, input|0x80
				if width == 1 {
					wantAX, wantDX = 3, before.gp[dx] // B writes AX16, never DX.
				} else if width == 4 {
					wantAX, wantDX = 0xff, 0xff // Dword writes clear both upper32 halves.
				}
				if state.gp[ax] != wantAX || state.gp[dx] != wantDX || state.flags != x86ByteFlagCF|x86ByteFlagOF {
					t.Fatal("incorrect implicit widths or undefined status retained", state.gp, state.flags)
				}
				unknown := before.clone()
				unknown.gp[ax] = 0
				original := unknown.clone()
				if x86ScalarFPResultTransfer(ins, &unknown, nil, 0, arch) || !reflect.DeepEqual(unknown, original) {
					t.Fatal("unknown accumulator acquired initialized bytes")
				}
			})
		}
	}
}

func TestX86ImplicitMultiplyDoesNotRegisterRawEntryOrDivide(t *testing.T) {
	for _, arch := range []string{"amd64", "386"} {
		for _, row := range x86ImplicitMultiplyCases(arch) {
			ins := x86YMMReadSource(t, row.op+" BX")
			if _, known, _ := x86TypedGPByteEffect(ins.Op, ins, arch); known || x86SourceStackInstructionKnown(ins, "MOVQ", "Q", arch) {
				t.Fatal("NUM-only effect widened an independent raw/source-stack registry", arch, row.op)
			}
			ins.x86Encoded = true
			if _, known, _ := x86TypedInstructionByteEffect(ins.Op, ins, arch); known {
				t.Fatal("named implicit product supplied a physical instruction effect")
			}
		}
		for _, row := range []string{"DIVB BX", "DIVW BX", "DIVL BX", "DIVQ BX", "IDIVB BX", "IDIVW BX", "IDIVL BX", "IDIVQ BX"} {
			ins := x86YMMReadSource(t, row)
			if _, known, _ := x86TypedInstructionByteEffect(ins.Op, ins, arch); known {
				t.Fatal("implicit multiply granted division/poison authority", arch, row)
			}
		}
	}
}
