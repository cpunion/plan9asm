package goasmrules

import (
	"bytes"
	"runtime"
	"testing"
)

func TestX86PackedUnpackRowsShareBinaryWitnesses(t *testing.T) {
	x86CheckPackedRowInventory(t, X86PackedUnpackRows,
		`^V?P?UNPCK[HL](BW|WL|LQ|WD|DQ|QDQ|PS|PD)$`,
		`(DQ|PS|PD)$`, x86PackedRowCounts{
			names: 24, mmx: 6, broadcast: 72,
			rows: [3]int{12, 24, 72}, formats: [3]int{24, 48, 288}, widths: [3]int{144, 120, 96},
		})
}

func TestX86PackedUnpackEncodingAxes(t *testing.T) {
	legacy, err := LoadX86LegacyRules(runtime.GOROOT())
	if err != nil {
		t.Fatal(err)
	}
	avx, err := LoadX86AVXRules(runtime.GOROOT())
	if err != nil {
		t.Fatal(err)
	}
	rows, err := X86PackedUnpackRows(legacy, avx)
	if err != nil {
		t.Fatal(err)
	}
	observed := map[string][]byte{}
	for _, row := range rows {
		for _, witness := range row.Witnesses {
			if _, exists := observed[witness.Instruction]; exists {
				t.Fatal("distinct official rows lost their physical witness identity", witness.Instruction)
			}
			observed[witness.Instruction] = witness.Encoding
		}
	}
	// Literal bytes independently cover the legacy L/W versus AVX D/W
	// spellings, pp=0/66, VEX W0 versus EVEX W1, mask, width and broadcast.
	for _, row := range []struct {
		instruction string
		encoding    []byte
	}{
		{"PUNPCKLWL X0,X1", []byte{0x66, 0x0f, 0x61, 0xc8}},
		{"PUNPCKHLQ (AX),X1", []byte{0x66, 0x0f, 0x6a, 0x08}},
		{"UNPCKLPS X0,X1", []byte{0x0f, 0x14, 0xc8}},
		{"UNPCKHPD (AX),X1", []byte{0x66, 0x0f, 0x15, 0x08}},
		{"VPUNPCKLWD X0,X1,X2", []byte{0xc5, 0xf1, 0x61, 0xd0}},
		{"VPUNPCKLQDQ Y0,Y1,Y2", []byte{0xc5, 0xf5, 0x6c, 0xd0}},
		{"VPUNPCKLQDQ X16,X17,X18", []byte{0x62, 0xa1, 0xf5, 0x00, 0x6c, 0xd0}},
		{"VUNPCKLPS X16,X17,X18", []byte{0x62, 0xa1, 0x74, 0x00, 0x14, 0xd0}},
		{"VPUNPCKHDQ.BCST.Z (AX),Y17,K1,Y18", []byte{0x62, 0xe1, 0x75, 0xb1, 0x6a, 0x10}},
		{"VUNPCKHPD.BCST (AX),Z17,Z18", []byte{0x62, 0xe1, 0xf5, 0x50, 0x15, 0x10}},
	} {
		t.Run(row.instruction, func(t *testing.T) {
			if got := observed[row.instruction]; !bytes.Equal(got, row.encoding) {
				t.Fatalf("official encoding is %x, want %x", got, row.encoding)
			}
		})
	}
}
