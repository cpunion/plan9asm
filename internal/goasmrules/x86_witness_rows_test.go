package goasmrules

import (
	"bytes"
	"strings"
	"testing"
)

func TestX86PackedNarrowRowsShareBinaryWitnesses(t *testing.T) {
	x86CheckPackedRowInventory(t, X86PackedNarrowRows, `^V?PACK`, `DW$`, x86PackedRowCounts{
		names: 8, mmx: 3, broadcast: 18,
		rows: [3]int{4, 8, 24}, formats: [3]int{8, 16, 90}, widths: [3]int{46, 38, 30},
	})
}

func TestX86PackedAverageMultiplyRowsShareBinaryWitnesses(t *testing.T) {
	x86CheckPackedRowInventory(t, X86PackedAverageMultiplyRows,
		`^V?(PAVG[BW]|PMUL(LW|HW|HUW|HRSW|DQ|ULQ|UDQ)|PMADD(UBSW|WL|WD))$`,
		`^VPMUL(DQ|UDQ)$`, x86PackedRowCounts{
			names: 20, mmx: 7, broadcast: 18,
			rows: [3]int{10, 20, 60}, formats: [3]int{20, 40, 198}, widths: [3]int{106, 86, 66},
		})
}

func TestX86PackedBinaryPilotRows(t *testing.T) {
	inventory := x86CheckPackedRowInventory(t, X86PackedBinaryPilotRows,
		`^V?(P(ADD|SUB)(B|W|L|D|Q|SB|SW|USB|USW)|PMULL[DQ])$`,
		`[DQ]$`, x86PackedRowCounts{
			names: 35, mmx: 7, broadcast: 54,
			rows: [3]int{17, 34, 108}, formats: [3]int{34, 68, 378}, widths: [3]int{194, 160, 126},
		})
	for _, row := range inventory.rows {
		for _, witness := range row.Witnesses {
			if len(witness.Encoding) < 4 {
				t.Fatal("integer pilot lost its mandatory prefix or vector encoding", row.ID, witness)
			}
		}
	}
	for _, name := range []string{"PADDD", "PMULLW", "VPMULLW"} {
		if inventory.names[name] {
			t.Fatalf("alias or unmodified word multiply entered pilot: %s", name)
		}
	}
	wordMultiply := 0
	for _, entry := range inventory.legacy.Optab {
		if entry.Opcode == "PMULLW" {
			wordMultiply += len(entry.Rows)
		}
	}
	for _, row := range inventory.avx {
		if row.Opcode == "VPMULLW" {
			wordMultiply++
		}
	}
	if wordMultiply != 10 {
		t.Fatal("outside-pilot word multiply inventory changed", wordMultiply)
	}
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
	inventory := x86CheckPackedRowInventory(t, X86PackedIntegerMinMaxRows,
		`^V?P(MIN|MAX)`, `[DQ]$`, x86PackedRowCounts{
			names: 28, broadcast: 72,
			rows: [3]int{12, 24, 96}, formats: [3]int{24, 48, 360}, widths: [3]int{168, 144, 120},
		})
	narrowEVEX := 0
	for _, row := range inventory.rows {
		for _, witness := range row.Witnesses {
			if witness.Generation == "evex" && witness.Width < 512 && !witness.Masked &&
				!witness.Broadcast && !strings.HasSuffix(row.Opcode, "Q") {
				narrowEVEX++
			}
		}
	}
	if narrowEVEX != 48 {
		t.Fatal("unmasked narrow non-Q EVEX inventory drift", narrowEVEX)
	}
}
