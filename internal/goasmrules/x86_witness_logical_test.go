package goasmrules

import (
	"bytes"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

func TestX86PackedLogicalRowsShareBinaryWitnesses(t *testing.T) {
	legacy, err := LoadX86LegacyRules(runtime.GOROOT())
	if err != nil {
		t.Fatal(err)
	}
	avx, err := LoadX86AVXRules(runtime.GOROOT())
	if err != nil {
		t.Fatal(err)
	}
	rows, err := X86PackedLogicalRows(legacy, avx)
	if err != nil {
		t.Fatal(err)
	}
	ids, names := map[string]bool{}, map[string]bool{}
	generationRows, formats := map[string]int{}, map[string]int{}
	widths := map[int]int{}
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
		}
	}
	if len(rows) != 132 || len(names) != 32 || generationRows["legacy"] != 12 ||
		generationRows["vex"] != 24 || generationRows["evex"] != 96 ||
		formats["legacy"] != 24 || formats["vex"] != 48 || formats["evex"] != 432 ||
		widths[128] != 192 || widths[256] != 168 || widths[512] != 144 {
		t.Fatalf("logical denominator drift: rows=%d names=%d generations=%v formats=%v widths=%v",
			len(rows), len(names), generationRows, formats, widths)
	}
	member := regexp.MustCompile(`^(V?P(ANDN?|OR|XOR)(D|Q)?|V?(ANDN?|OR|XOR)(PS|PD))$`)
	check := func(name string) {
		if member.MatchString(name) && !names[name] {
			t.Fatal("official logical family member omitted", name)
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
	if mmx != 4 {
		t.Fatal("separate MMX region inventory drift", mmx)
	}
	t.Logf("%s: 132 original rows / 504 expectations; four MMX rows separate, no translation/runtime claim", runtime.Version())
}

func TestX86PackedLogicalLegacyPrefixAxes(t *testing.T) {
	legacy, err := LoadX86LegacyRules(runtime.GOROOT())
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range []struct {
		name string
		want []byte
	}{
		{"ANDPS", []byte{0x0f, 0x54, 0xc8}},
		{"ANDPD", []byte{0x66, 0x0f, 0x54, 0xc8}},
		{"XORPD", []byte{0x66, 0x0f, 0x57, 0xc8}},
		{"PAND", []byte{0x66, 0x0f, 0xdb, 0xc8}},
	} {
		t.Run(row.name, func(t *testing.T) {
			for _, entry := range legacy.Optab {
				if entry.Opcode != row.name {
					continue
				}
				operand := entry.Rows[len(entry.Rows)-1]
				witnesses, err := x86PackedLegacyRowWitnesses(entry, operand)
				if err != nil || len(witnesses) != 2 || !bytes.Equal(witnesses[0].Encoding, row.want) {
					t.Fatalf("official legacy prefix bytes: %+v / %v, want %x", witnesses, err, row.want)
				}
				entry.Prefix = "Pf3"
				if _, err := x86PackedLegacyRowWitnesses(entry, operand); err == nil {
					t.Fatal("unmodeled mandatory prefix accepted")
				}
				return
			}
			t.Fatal("missing official opcode", row.name)
		})
	}
}

func TestX86PackedLogicalAVXPrefixAndTableAxes(t *testing.T) {
	avx, err := LoadX86AVXRules(runtime.GOROOT())
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range []struct {
		name string
		row  int
		want []byte
	}{
		{"VANDPS", 0, []byte{0xc5, 0xf0, 0x54, 0xd0}},
		{"VANDPD", 0, []byte{0xc5, 0xf1, 0x54, 0xd0}},
		{"VPAND", 0, []byte{0xc5, 0xf1, 0xdb, 0xd0}},
		{"VANDPS", 2, []byte{0x62, 0xa1, 0x74, 0x00, 0x54, 0xd0}},
		{"VANDPD", 2, []byte{0x62, 0xa1, 0xf5, 0x00, 0x54, 0xd0}},
	} {
		t.Run(row.name+"/"+strconv.Itoa(row.row), func(t *testing.T) {
			for _, rule := range avx {
				if rule.Opcode != row.name || rule.RowIndex != row.row {
					continue
				}
				witnesses, err := x86PackedAVXRowWitnesses(rule)
				if err != nil || len(witnesses) < 2 || !bytes.Equal(witnesses[0].Encoding, row.want) {
					t.Fatalf("official AVX prefix/table bytes: %+v / %v, want %x", witnesses, err, row.want)
				}
				for _, suffix := range []string{" | futurePrefix", " | vexF3", " | evexF2"} {
					changed := rule
					changed.OpBytes = append([]string(nil), rule.OpBytes...)
					changed.OpBytes[rule.EncodingOffset] += suffix
					if _, err := x86PackedAVXRowWitnesses(changed); err == nil {
						t.Fatal("unmodeled or contradictory mandatory prefix accepted", suffix)
					}
				}
				if row.row == 2 {
					rule.Ytab = "_yvaddsubpd"
					if _, err := x86PackedAVXRowWitnesses(rule); err == nil {
						t.Fatal("VEX-only grammar acquired EVEX")
					}
				}
				return
			}
			t.Fatal("missing official row", row.name, row.row)
		})
	}
}

func TestX86PackedBinaryRejectsVEX512(t *testing.T) {
	rule := X86AVXRule{
		Opcode: "VPAND", Ytab: "_yvaddsubpd", Prefix: "Pavx", ZCase: "Zvex_rm_v_r", ZOffset: 2,
		Classes: []string{"Yzm", "Yzr", "Yzr"},
		OpBytes: []string{"avxEscape | vex512 | vex66 | vex0F | vexW0", "0xdb"},
	}
	if witnesses, err := x86PackedAVXRowWitnesses(rule); err == nil || len(witnesses) != 0 {
		t.Fatal("VEX acquired an impossible Z/512-bit encoding", witnesses, err)
	}
}
