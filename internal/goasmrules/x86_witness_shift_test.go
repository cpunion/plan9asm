package goasmrules

import (
	"bytes"
	"testing"
)

var x86PerLaneShiftRowCounts = x86PackedRowCounts{
	names: 9, broadcast: 54,
	rows: [3]int{0, 10, 54}, formats: [3]int{0, 20, 216}, widths: [3]int{82, 82, 72},
}

func TestX86PackedPerLaneShiftRowsShareBinaryWitnesses(t *testing.T) {
	x86CheckPackedRowInventory(t, X86PackedPerLaneShiftRows,
		`^VPS(LL|RL|RA)V`, `[DQ]$`, x86PerLaneShiftRowCounts)
}

func TestX86PackedPerLaneShiftEncodingAxes(t *testing.T) {
	inventory := x86CheckPackedRowInventory(t, X86PackedPerLaneShiftRows,
		`^VPS(LL|RL|RA)V`, `[DQ]$`, x86PerLaneShiftRowCounts)
	observed := map[string][]byte{}
	for _, row := range inventory.rows {
		for _, witness := range row.Witnesses {
			if _, exists := observed[witness.Instruction]; exists {
				t.Fatal("distinct official rows lost their physical witness identity", witness.Instruction)
			}
			observed[witness.Instruction] = witness.Encoding
		}
	}
	// Independent literal bytes span all nine names, X/Y/Z widths,
	// W0/W1, masks, zeroing, complete memory and D/Q-only broadcast.
	for _, row := range []struct {
		instruction string
		encoding    []byte
	}{
		{"VPSLLVD X0,X1,X2", []byte{0xc4, 0xe2, 0x71, 0x47, 0xd0}},
		{"VPSLLVQ Y0,Y1,Y2", []byte{0xc4, 0xe2, 0xf5, 0x47, 0xd0}},
		{"VPSLLVW X16,X17,X18", []byte{0x62, 0xa2, 0xf5, 0x00, 0x12, 0xd0}},
		{"VPSRAVD X0,X1,X2", []byte{0xc4, 0xe2, 0x71, 0x46, 0xd0}},
		{"VPSRAVQ.BCST.Z (AX),Z17,K1,Z18", []byte{0x62, 0xe2, 0xf5, 0xd1, 0x46, 0x10}},
		{"VPSRAVW.Z (AX),Y17,K1,Y18", []byte{0x62, 0xe2, 0xf5, 0xa1, 0x11, 0x10}},
		{"VPSRLVD.BCST (AX),X17,X18", []byte{0x62, 0xe2, 0x75, 0x10, 0x45, 0x10}},
		{"VPSRLVQ Y16,Y17,K1,Y18", []byte{0x62, 0xa2, 0xf5, 0x21, 0x45, 0xd0}},
		{"VPSRLVW Z16,Z17,Z18", []byte{0x62, 0xa2, 0xf5, 0x40, 0x10, 0xd0}},
	} {
		t.Run(row.instruction, func(t *testing.T) {
			if got := observed[row.instruction]; !bytes.Equal(got, row.encoding) {
				t.Fatalf("official encoding is %x, want %x", got, row.encoding)
			}
		})
	}
}
