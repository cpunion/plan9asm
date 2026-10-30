package goasmrules

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

// This inventory is independent of a new formatter. It names the four whole
// source tables, not an opcode-byte registry or a numerical x87 permission.
var x86X87SourceTableNames = map[string]string{
	"yfmvx": `FMOVB FMOVV FMOVX FCOMF FCOMFP FCOML FCOMLP FCOMW FCOMWP
		FADDW FADDL FADDF FMULW FMULL FMULF FSUBW FSUBL FSUBF
		FSUBRW FSUBRL FSUBRF FDIVW FDIVL FDIVF FDIVRW FDIVRL FDIVRF`,
	"yfadd": "FCOMD FCOMDP FADDD FMULD FSUBD FSUBRD FDIVD FDIVRD",
	"yfcmv": `FCMOVCC FCMOVCS FCMOVEQ FCMOVHI FCMOVLS FCMOVB FCMOVBE
		FCMOVNB FCMOVNBE FCMOVE FCMOVNE FCMOVNU FCMOVU FCMOVUN FCOMI FCOMIP`,
	"ycompp": `FCOMDPP FUCOM FUCOMI FUCOMIP FUCOMP FUCOMPP
		FADDDP FMULDP FSUBDP FSUBRDP FDIVDP FDIVRDP`,
}

func x86X87SourceTableMember(entry X86LegacyOpcode) bool {
	_, found := x86X87SourceTableNames[entry.Ytab]
	return found
}

func TestX86LegacyX87CompleteOriginalSourceInventory(t *testing.T) {
	catalog, err := LoadX86LegacyRules(runtime.GOROOT())
	if err != nil {
		t.Fatal(err)
	}
	schemas := map[string][]string{
		"yfmvx":  {"Zm_o/2/Ym,Yf0"},
		"yfadd":  {"Zm_o/2/Ym,Yf0", "Zm_o/2/Yrf,Yf0", "Zo_m/2/Yf0,Yrf"},
		"yfcmv":  {"Zm_o/2/Yrf,Yf0"},
		"ycompp": {"Zo_m/2/Yf0,Yrf"},
	}
	counts, names, identities := map[string]int{}, map[string]bool{}, map[string]bool{}
	fragments := 0
	for _, entry := range catalog.Optab {
		if !x86X87SourceTableMember(entry) {
			continue
		}
		if names[entry.Opcode] || !strings.Contains(" "+strings.Join(strings.Fields(x86X87SourceTableNames[entry.Ytab]), " ")+" ", " "+entry.Opcode+" ") {
			t.Fatal("missing, repeated or foreign x87 source member", entry)
		}
		names[entry.Opcode] = true
		if entry.Kind != "encoding" || entry.Prefix != "Px" || entry.EscapeAdjustment != 0 ||
			len(entry.Rows) != len(schemas[entry.Ytab]) || len(entry.OpBytes) != 2*len(entry.Rows) {
			t.Fatal("complete x87 source schema changed", entry)
		}
		for index, row := range entry.Rows {
			id := fmt.Sprintf("optab/%d/%s/%s/%d", entry.TableIndex, entry.Opcode, entry.Ytab, row.RowIndex)
			if identities[id] || row.RowIndex != index || row.UnmatchedOffset != 2*index ||
				fmt.Sprintf("%s/%d/%s", row.ZCase, row.ZOffset, strings.Join(row.Classes, ",")) != schemas[entry.Ytab][index] {
				t.Fatal("original ordered x87 row identity changed", id, row)
			}
			identities[id] = true
			counts[entry.Ytab]++
			for atom := 0; atom < 2; atom++ {
				value, err := strconv.ParseUint(entry.OpBytes[2*index+atom], 0, 8)
				if err != nil || atom == 0 && (value < 0xd8 || value > 0xdf) {
					t.Fatal("x87 row lost an actual opcode/ModRM source atom", id, err)
				}
				if atom == 1 && value > 7 {
					fragments++
					if entry.Opcode != "FUCOMPP" || value != 13 {
						t.Fatal("source ModRM field-fragment inventory changed", id, value)
					}
				}
			}
		}
	}
	for table, expected := range map[string]int{"yfmvx": 27, "yfadd": 24, "yfcmv": 16, "ycompp": 12} {
		if counts[table] != expected {
			t.Fatal("incomplete x87 original row family", table, counts[table], expected)
		}
	}
	if len(identities) != 79 || len(names) != 63 || fragments != 1 {
		t.Fatal("x87 row/name/field-fragment union changed", len(identities), len(names), fragments)
	}
}

// The recorded feature RED called the old helper. The new explicit x87
// projection preserves that packed-only refusal and all its original views.
func TestX86LegacyX87ExplicitCarrierTransportRequired(t *testing.T) {
	catalog, err := LoadX86LegacyRules(runtime.GOROOT())
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range catalog.Optab {
		if !x86X87SourceTableMember(entry) {
			continue
		}
		for _, row := range entry.Rows {
			t.Run(fmt.Sprintf("%d/%s/%s/%d", entry.TableIndex, entry.Opcode, entry.Ytab, row.RowIndex), func(t *testing.T) {
				witnesses, err := x86X87RowWitnesses(entry, row)
				if err != nil || len(witnesses) != 2 {
					t.Fatal("complete original x87 row has no explicit carrier transport", err)
				}
				if previous, err := x86PackedLegacyRowWitnesses(entry, row); err == nil || len(previous) != 0 {
					t.Fatal("x87 projection widened the old packed-only contract", previous, err)
				}
			})
		}
	}
}

func TestX86LegacyX87TransportRejectsForeignSourceAxes(t *testing.T) {
	catalog, err := LoadX86LegacyRules(runtime.GOROOT())
	if err != nil {
		t.Fatal(err)
	}
	for _, original := range catalog.Optab {
		if !x86X87SourceTableMember(original) {
			continue
		}
		for _, selected := range original.Rows {
			for _, control := range []string{"kind", "table", "prefix", "escape", "index", "classes", "zcase", "stride", "offset", "extent", "selected-byte", "unselected-byte"} {
				t.Run(fmt.Sprintf("%s/%d/%s", original.Opcode, selected.RowIndex, control), func(t *testing.T) {
					entry, row := original, selected
					entry.OpBytes = append([]string(nil), original.OpBytes...)
					switch control {
					case "kind":
						entry.Kind = "special"
					case "table":
						entry.Ytab = "unmodeled"
					case "prefix":
						entry.Prefix = "Pw"
					case "escape":
						entry.EscapeAdjustment++
					case "index":
						row.RowIndex = len(entry.Rows)
					case "classes":
						row.Classes = []string{"Yxm", "Yxr"}
					case "zcase":
						row.ZCase = "Zm_r_xm"
					case "stride":
						row.ZOffset++
					case "offset":
						row.UnmatchedOffset++
					case "extent":
						entry.OpBytes = append(entry.OpBytes, "0")
					case "selected-byte":
						entry.OpBytes[row.UnmatchedOffset] = "Pe"
					case "unselected-byte":
						entry.OpBytes[len(entry.OpBytes)-1] = "invalid"
					}
					if witnesses, err := x86X87RowWitnesses(entry, row); err == nil || len(witnesses) != 0 {
						t.Fatal("foreign source attributes supplied an x87 carrier transport", control, witnesses, err)
					}
				})
			}
		}
	}
}
