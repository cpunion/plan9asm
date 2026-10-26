package goasmrules

import (
	"bytes"
	"runtime"
	"strings"
	"testing"
)

func TestX86PackedAverageMultiplyRowsShareBinaryWitnesses(t *testing.T) {
	legacy, err := LoadX86LegacyRules(runtime.GOROOT())
	if err != nil {
		t.Fatal(err)
	}
	avx, err := LoadX86AVXRules(runtime.GOROOT())
	if err != nil {
		t.Fatal(err)
	}
	rows, err := X86PackedAverageMultiplyRows(legacy, avx)
	if err != nil {
		t.Fatal(err)
	}
	ids, names := map[string]bool{}, map[string]bool{}
	generationRows, generationFormats := map[string]int{}, map[string]int{}
	widthFormats := map[int]int{}
	for _, row := range rows {
		if row.ID == "" || ids[row.ID] || len(row.Witnesses) < 2 {
			t.Fatal("missing original row identity or register/memory witness", row.ID)
		}
		ids[row.ID], names[row.Opcode] = true, true
		generationRows[row.Witnesses[0].Generation]++
		for _, witness := range row.Witnesses {
			generationFormats[witness.Generation]++
			widthFormats[witness.Width]++
			if witness.Generation == "evex" && (witness.Encoding[0] != 0x62 ||
				!strings.HasSuffix(witness.Second, "17") || !strings.HasSuffix(witness.Destination, "18")) {
				t.Fatal("official EVEX identity can fall back to VEX", row.ID)
			}
		}
	}
	if len(rows) != 90 || len(names) != 20 || generationRows["legacy"] != 10 ||
		generationRows["vex"] != 20 || generationRows["evex"] != 60 ||
		generationFormats["legacy"] != 20 || generationFormats["vex"] != 40 || generationFormats["evex"] != 198 ||
		widthFormats[128] != 106 || widthFormats[256] != 86 || widthFormats[512] != 66 {
		t.Fatalf("average/multiply denominator drift: rows=%d names=%d generations=%v formats=%v widths=%v",
			len(rows), len(names), generationRows, generationFormats, widthFormats)
	}
	mmx := 0
	for _, entry := range legacy.Optab {
		if !names[entry.Opcode] {
			continue
		}
		for _, row := range entry.Rows {
			if strings.Join(row.Classes, ",") == "Ymm,Ymr" {
				mmx++
			}
		}
	}
	if mmx != 7 {
		t.Fatalf("separate MMX region inventory drift: %d", mmx)
	}
	t.Logf("%s: 90 canonical rows / 258 register, memory, width and mask expectations; seven MMX rows remain outside this scope, no translation/runtime claim", runtime.Version())
}

func TestX86PackedBinaryPilotRows(t *testing.T) {
	legacy, err := LoadX86LegacyRules(runtime.GOROOT())
	if err != nil {
		t.Fatal(err)
	}
	avx, err := LoadX86AVXRules(runtime.GOROOT())
	if err != nil {
		t.Fatal(err)
	}
	rows, err := X86PackedBinaryPilotRows(legacy, avx)
	if err != nil {
		t.Fatal(err)
	}
	seen, names := map[string]bool{}, map[string]bool{}
	legacyRows, avxRows, witnesses := 0, 0, 0
	for _, row := range rows {
		if seen[row.ID] || row.ID == "" {
			t.Fatalf("lost distinct official row identity: %q", row.ID)
		}
		seen[row.ID], names[row.Opcode] = true, true
		if strings.HasPrefix(row.ID, "optab/") {
			legacyRows++
		} else {
			avxRows++
		}
		if len(row.Witnesses) < 2 {
			t.Fatalf("%s lacks register/memory witnesses", row.ID)
		}
		for _, witness := range row.Witnesses {
			witnesses++
			if len(witness.Encoding) < 4 || witness.Instruction == "" {
				t.Fatalf("%s has an incomplete witness: %+v", row.ID, witness)
			}
			if witness.Generation == "evex" && (!strings.HasSuffix(witness.Second, "17") ||
				!strings.HasSuffix(witness.Destination, "18") || witness.Encoding[0] != 0x62) {
				t.Fatalf("%s can accidentally select VEX: %+v", row.ID, witness)
			}
		}
	}
	if legacyRows != 17 || avxRows != 142 || len(names) != 35 || witnesses != 480 {
		t.Fatalf("pilot denominator changed: legacy=%d avx=%d names=%d witnesses=%d", legacyRows, avxRows, len(names), witnesses)
	}
	for _, name := range []string{"PADDD", "PMULLW", "VPMULLW"} {
		if names[name] {
			t.Fatalf("alias or unmodified word multiply entered pilot: %s", name)
		}
	}
	mmx, wordMultiply := 0, 0
	for _, entry := range legacy.Optab {
		for _, row := range entry.Rows {
			if names[entry.Opcode] && strings.Join(row.Classes, ",") == "Ymm,Ymr" {
				mmx++
			}
			if entry.Opcode == "PMULLW" {
				wordMultiply++
			}
		}
	}
	for _, row := range avx {
		if row.Opcode == "VPMULLW" {
			wordMultiply++
		}
	}
	if mmx != 7 || wordMultiply != 10 {
		t.Fatalf("outside-pilot inventory changed: MMX add=%d, word multiply=%d", mmx, wordMultiply)
	}
	t.Logf("%s: 159 original rows, 35 canonical opcodes, 480 syntax/encoding expectations; not translation coverage", runtime.Version())
}

func TestX86PackedBinaryRowEncodingAndFailClosed(t *testing.T) {
	rule := X86AVXRule{
		Opcode: "VPMULLD", Ytab: "_yvandnpd", Prefix: "Pavx", RowIndex: 3,
		Classes: []string{"YxmEvex", "YxrEvex", "Yknot0", "YxrEvex"},
		ZCase:   "Zevex_rm_v_k_r", ZOffset: 3,
		OpBytes: []string{"avxEscape | evex128 | evex66 | evex0F38 | evexW0", "evexN16 | evexBcstN4 | evexZeroingEnabled", "0x40"},
	}
	witnesses, err := x86PackedAVXRowWitnesses(rule)
	if err != nil {
		t.Fatal(err)
	}
	if len(witnesses) != 6 || witnesses[0].Instruction != "VPMULLD X16,X17,K1,X18" ||
		!bytes.Equal(witnesses[0].Encoding, []byte{0x62, 0xa2, 0x75, 0x01, 0x40, 0xd0}) ||
		witnesses[5].Instruction != "VPMULLD.BCST.Z (AX),X17,K1,X18" ||
		!bytes.Equal(witnesses[5].Encoding, []byte{0x62, 0xe2, 0x75, 0x91, 0x40, 0x10}) {
		t.Fatalf("independent fixed-register EVEX expectation: %+v", witnesses)
	}
	for _, mutate := range []func(*X86AVXRule){
		func(r *X86AVXRule) { r.OpBytes[0] += " | futureFlag" },
		func(r *X86AVXRule) { r.OpBytes[0] += " | evexW1" },
		func(r *X86AVXRule) { r.OpBytes[1] += " | evexSaeEnabled" },
		func(r *X86AVXRule) { r.OpBytes[1] = strings.Replace(r.OpBytes[1], "evexN16", "evexN64", 1) },
		func(r *X86AVXRule) { r.OpBytes[2] = "opcodeConstant" },
		func(r *X86AVXRule) { r.EncodingOffset = 1 },
		func(r *X86AVXRule) { r.ZOffset = 2 },
		func(r *X86AVXRule) { r.Classes[0] = "Yxm" },
		func(r *X86AVXRule) { r.Ytab = "unknown" },
	} {
		local := rule
		local.OpBytes = append([]string(nil), rule.OpBytes...)
		local.Classes = append([]string(nil), rule.Classes...)
		mutate(&local)
		if _, err := x86PackedAVXRowWitnesses(local); err == nil {
			t.Fatalf("unmodeled official row silently accepted: %+v", local)
		}
	}
}

func TestX86PackedBinaryLegacyRowUsesSelectedPrefixOffset(t *testing.T) {
	entry := X86LegacyOpcode{Kind: "encoding", Opcode: "PADDL", Ytab: "ymm", Prefix: "Py1", OpBytes: []string{"0xfe", "Pe", "0xfe"}}
	row := X86LegacyOperand{RowIndex: 1, Classes: []string{"Yxm", "Yxr"}, ZCase: "Zm_r_xm", ZOffset: 2, UnmatchedOffset: 1}
	witnesses, err := x86PackedLegacyRowWitnesses(entry, row)
	if err != nil || len(witnesses) != 2 || !bytes.Equal(witnesses[0].Encoding, []byte{0x66, 0x0f, 0xfe, 0xc8}) ||
		!bytes.Equal(witnesses[1].Encoding, []byte{0x66, 0x0f, 0xfe, 0x08}) {
		t.Fatalf("Py1 selected XMM prefix/opcode: %+v/%v", witnesses, err)
	}
	for _, offset := range []int{0, 2, 3} {
		local := row
		local.UnmatchedOffset = offset
		if _, err := x86PackedLegacyRowWitnesses(entry, local); err == nil {
			t.Fatal("unmatched traversal offset was mistaken for a selected opcode offset", offset)
		}
	}
	entry.OpBytes[1] = "Pf2"
	if _, err := x86PackedLegacyRowWitnesses(entry, row); err == nil {
		t.Fatal("changed legacy mandatory prefix was ignored")
	}
}

func TestX86PackedIntegerMinMaxRowsShareBinaryWitnesses(t *testing.T) {
	legacy, err := LoadX86LegacyRules(runtime.GOROOT())
	if err != nil {
		t.Fatal(err)
	}
	avx, err := LoadX86AVXRules(runtime.GOROOT())
	if err != nil {
		t.Fatal(err)
	}
	rows, err := X86PackedIntegerMinMaxRows(legacy, avx)
	if err != nil {
		t.Fatal(err)
	}
	names, ids := map[string]bool{}, map[string]bool{}
	legacyRows, formats, narrowEVEX := 0, 0, 0
	for _, row := range rows {
		if row.ID == "" || ids[row.ID] {
			t.Fatal("missing or duplicate original min/max row", row.ID)
		}
		ids[row.ID], names[row.Opcode] = true, true
		if strings.HasPrefix(row.ID, "optab/") {
			legacyRows++
		}
		for _, witness := range row.Witnesses {
			formats++
			if witness.Generation == "evex" && witness.Width < 512 && !witness.Masked &&
				!witness.Broadcast && !strings.HasSuffix(row.Opcode, "Q") {
				narrowEVEX++
				if witness.Encoding[0] != 0x62 || !strings.HasSuffix(witness.Second, "17") {
					t.Fatal("unmasked narrow EVEX row can fall back to VEX", row.ID)
				}
			}
		}
	}
	if len(rows) != 132 || legacyRows != 12 || len(names) != 28 || formats != 432 || narrowEVEX != 48 {
		t.Fatalf("min/max inventory drift: rows=%d legacy=%d names=%d formats=%d narrowEVEX=%d",
			len(rows), legacyRows, len(names), formats, narrowEVEX)
	}
	checkMember := func(opcode string) {
		base := strings.TrimPrefix(opcode, "V")
		if (strings.HasPrefix(base, "PMIN") || strings.HasPrefix(base, "PMAX")) && !names[opcode] {
			t.Fatalf("new official min/max member is absent from the independent scope: %s", opcode)
		}
	}
	for _, entry := range legacy.Optab {
		checkMember(entry.Opcode)
	}
	for _, row := range avx {
		checkMember(row.Opcode)
	}
}
