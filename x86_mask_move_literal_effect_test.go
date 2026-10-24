package plan9asm

import (
	"fmt"
	"reflect"
	"testing"
)

func TestX86MaskMoveLiteralProofExactReadWriteRoles(t *testing.T) {
	forms := []struct {
		operands string
		gpWrite  bool
	}{
		{"AX,K1", false}, {"K1,K7", false}, {"K1,BX", true},
		{"(DX),K1", false}, {"K1,(DX)", false},
	}
	for _, arch := range []string{"386", "amd64"} {
		for _, row := range x86ResultMaskMoveWidthOracle {
			for _, form := range forms {
				t.Run(arch+"/"+row.op+"/"+form.operands, func(t *testing.T) {
					ins := x86ScalarLaneInstruction(t, row.op+" "+form.operands)
					effect, known, err := x86TypedInstructionByteEffect(ins.Op, ins, arch)
					physicalQ := arch == "386" && row.bits == 64 &&
						(amd64IsYrlOperand(ins.Args[0], arch) || amd64IsYrlOperand(ins.Args[1], arch))
					if physicalQ {
						if !known || err == nil {
							t.Fatal("mode32 physical qword role lost", known, err)
						}
						return
					}
					width := int64(row.bits / 8)
					writeWidth := width
					if form.gpWrite && writeWidth < x86ResultPointerWord(arch) {
						writeWidth = x86ResultPointerWord(arch)
					}
					want := x86TypedByteEffect{requiresResultProof: true, operands: []x86TypedByteOperand{
						{index: 0, operand: ins.Args[0], width: width, read: true},
						{index: 1, operand: ins.Args[1], width: writeWidth, write: true},
					}}
					if !known || err != nil || !reflect.DeepEqual(effect, want) {
						t.Fatal("mask move lost exact data width, destination zeroing or preserved FLAGS", effect, known, err)
					}
					state := x86SDKLiteralState{comparison: 1, constant: 16, compared: true}
					state.gp[0] = x86SDKLiteralUnknownNumber(x86ResultPointerWord(arch))
					state.gp[1] = x86SDKLiteralValue{width: x86ResultPointerWord(arch), minimum: 0, maximum: 15}
					state.gp[3] = x86SDKLiteralValue{symbol: "table<>", minimum: 0, maximum: 0}
					wantState := state
					if form.gpWrite {
						wantState.gp[3] = x86SDKLiteralValue{}
					}
					if err := x86SDKLiteralTransfer(ins, &state, FuncSig{}, arch, nil); err != nil || !reflect.DeepEqual(state, wantState) {
						t.Fatal("mask move changed unrelated range, retained overwritten literal or invented memory facts", state, err)
					}
				})
			}
		}
	}
}

func TestX86MaskMoveLiteralCannotEscapeOrPreserveOverwrittenRange(t *testing.T) {
	for _, arch := range []string{"386", "amd64"} {
		for _, row := range x86ResultMaskMoveWidthOracle {
			if arch == "386" && row.bits == 64 {
				continue // The independent effect test retains the rejected GP-qword boundary.
			}
			t.Run(arch+"/"+row.op, func(t *testing.T) {
				state := x86SDKLiteralState{}
				state.gp[0] = x86SDKLiteralValue{symbol: "table<>"}
				ins := x86ScalarLaneInstruction(t, fmt.Sprintf("%s AX,K1", row.op))
				if err := x86SDKLiteralTransfer(ins, &state, FuncSig{}, arch, nil); err == nil {
					t.Fatal("literal address escaped as mask data")
				}
				state = x86SDKLiteralState{comparison: 1, constant: 16, compared: true}
				state.gp[1] = x86SDKLiteralValue{width: x86ResultPointerWord(arch), maximum: 15}
				ins = x86ScalarLaneInstruction(t, row.op+" K1,CX")
				if err := x86SDKLiteralTransfer(ins, &state, FuncSig{}, arch, nil); err != nil ||
					state.compared || state.gp[1] != (x86SDKLiteralValue{}) {
					t.Fatal("mask destination borrowed the previous index envelope", state, err)
				}
			})
		}
	}
}
