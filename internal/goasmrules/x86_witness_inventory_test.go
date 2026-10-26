package goasmrules

import (
	"regexp"
	"runtime"
	"strings"
	"testing"
)

type x86PackedRowCounts struct {
	names, mmx, broadcast int
	rows, formats         [3]int // Legacy, VEX, EVEX.
	widths                [3]int // 128, 256, 512 bits.
}

type x86PackedRowInventory struct {
	rows   []X86PackedBinaryRow
	legacy X86LegacyCatalog
	avx    []X86AVXRule
	names  map[string]bool
}

// Both the family membership and expected counts are independent test data.
// The complete official catalogs, not the translator registry, own the rows.
func x86CheckPackedRowInventory(t *testing.T,
	build func(X86LegacyCatalog, []X86AVXRule) ([]X86PackedBinaryRow, error),
	member, broadcast string, want x86PackedRowCounts) x86PackedRowInventory {
	t.Helper()
	legacy, err := LoadX86LegacyRules(runtime.GOROOT())
	if err != nil {
		t.Fatal(err)
	}
	avx, err := LoadX86AVXRules(runtime.GOROOT())
	if err != nil {
		t.Fatal(err)
	}
	rows, err := build(legacy, avx)
	if err != nil {
		t.Fatal(err)
	}
	family, allowedBroadcast := regexp.MustCompile(member), regexp.MustCompile(broadcast)
	ids, names := map[string]bool{}, map[string]bool{}
	generation := map[string]int{"legacy": 0, "vex": 1, "evex": 2}
	width := map[int]int{128: 0, 256: 1, 512: 2}
	var got x86PackedRowCounts
	for _, row := range rows {
		if row.ID == "" || ids[row.ID] || len(row.Witnesses) < 2 || !family.MatchString(row.Opcode) {
			t.Fatal("lost original row, family or register/memory witness", row.ID, row.Opcode)
		}
		ids[row.ID], names[row.Opcode] = true, true
		kind := row.Witnesses[0].Generation
		g, valid := generation[kind]
		if !valid || strings.HasPrefix(row.ID, "optab/") != (kind == "legacy") {
			t.Fatal("original row and encoding generation disagree", row.ID, kind)
		}
		got.rows[g]++
		for _, witness := range row.Witnesses {
			w, valid := width[witness.Width]
			// Legacy PS rows without a mandatory prefix are three bytes.
			if !valid || witness.Generation != kind || witness.Instruction == "" || len(witness.Encoding) < 3 {
				t.Fatal("incomplete or inconsistent original encoding witness", row.ID, witness)
			}
			got.formats[g]++
			got.widths[w]++
			if kind == "evex" && (witness.Encoding[0] != 0x62 ||
				!strings.HasSuffix(witness.Second, "17") || !strings.HasSuffix(witness.Destination, "18")) {
				t.Fatal("official EVEX identity can fall back to VEX", row.ID)
			}
			if witness.Broadcast {
				got.broadcast++
				if kind != "evex" || !witness.Memory || !allowedBroadcast.MatchString(row.Opcode) {
					t.Fatal("broadcast escaped the independent family contract", row.ID)
				}
			}
		}
	}
	check := func(name string) {
		if family.MatchString(name) && !names[name] {
			t.Fatal("official family member omitted", name)
		}
	}
	for _, entry := range legacy.Optab {
		check(entry.Opcode)
		for _, row := range entry.Rows {
			if names[entry.Opcode] && strings.Join(row.Classes, ",") == "Ymm,Ymr" {
				got.mmx++
			}
		}
	}
	for _, row := range avx {
		check(row.Opcode)
	}
	got.names = len(names)
	if got != want {
		t.Fatalf("official family inventory drift:\n got %+v\nwant %+v", got, want)
	}
	t.Logf("%s: original rows %v, format expectations %v; %d MMX rows separate, no translation/runtime claim",
		runtime.Version(), got.rows, got.formats, got.mmx)
	return x86PackedRowInventory{rows: rows, legacy: legacy, avx: avx, names: names}
}
