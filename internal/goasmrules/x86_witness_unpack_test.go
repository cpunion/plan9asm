package goasmrules

import (
	"bytes"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

func TestX86PackedUnpackRowsShareBinaryWitnesses(t *testing.T) {
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
	ids, names := map[string]bool{}, map[string]bool{}
	generationRows, formats := map[string]int{}, map[string]int{}
	widths := map[int]int{}
	broadcast := 0
	for _, row := range rows {
		if row.ID == "" || ids[row.ID] || len(row.Witnesses) < 2 {
			t.Fatal("lost official row or register/memory witness", row.ID)
		}
		ids[row.ID], names[row.Opcode] = true, true
		generationRows[row.Witnesses[0].Generation]++
		for _, witness := range row.Witnesses {
			formats[witness.Generation]++
			widths[witness.Width]++
			if witness.Generation == "evex" && (witness.Encoding[0] != 0x62 ||
				!strings.HasSuffix(witness.Second, "17") || !strings.HasSuffix(witness.Destination, "18")) {
				t.Fatal("EVEX row could silently select VEX", row.ID)
			}
			if witness.Broadcast {
				broadcast++
				if !witness.Memory || witness.Generation != "evex" ||
					!(strings.HasSuffix(row.Opcode, "DQ") || strings.HasSuffix(row.Opcode, "PS") || strings.HasSuffix(row.Opcode, "PD")) {
					t.Fatal("broadcast escaped its official D/Q/PS/PD memory rows", row.ID)
				}
			}
		}
	}
	if len(rows) != 108 || len(names) != 24 || generationRows["legacy"] != 12 ||
		generationRows["vex"] != 24 || generationRows["evex"] != 72 ||
		formats["legacy"] != 24 || formats["vex"] != 48 || formats["evex"] != 288 ||
		widths[128] != 144 || widths[256] != 120 || widths[512] != 96 || broadcast != 72 {
		t.Fatalf("unpack denominator drift: rows=%d names=%d generations=%v formats=%v widths=%v broadcast=%d",
			len(rows), len(names), generationRows, formats, widths, broadcast)
	}
	member := regexp.MustCompile(`^V?P?UNPCK[HL](BW|WL|LQ|WD|DQ|QDQ|PS|PD)$`)
	check := func(name string) {
		if member.MatchString(name) && !names[name] {
			t.Fatal("official unpack family member omitted", name)
		}
	}
	mmx := 0
	for _, entry := range legacy.Optab {
		check(entry.Opcode)
		for _, row := range entry.Rows {
			if names[entry.Opcode] && strings.Join(row.Classes, ",") == "Ymm,Ymr" {
				mmx++
			}
		}
	}
	for _, row := range avx {
		check(row.Opcode)
	}
	if mmx != 6 {
		t.Fatal("separate MMX region inventory drift", mmx)
	}
	t.Logf("%s: 108 original rows / 360 expectations; six MMX rows separate, no translation/runtime claim", runtime.Version())
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
