package goasmrules

import (
	"bytes"
	"fmt"
	"runtime"
	"strings"
	"testing"
)

func TestX86ScalarRMWCompleteOrderedRowsAndModes(t *testing.T) {
	catalog, err := LoadX86LegacyRules(runtime.GOROOT())
	if err != nil {
		t.Fatal(err)
	}
	for _, mode := range []struct {
		arch                string
		filtered, witnesses int
	}{{"amd64", 2, 278}, {"386", 45, 202}} {
		rows, err := X86ScalarRMWRows(catalog, mode.arch)
		if err != nil {
			t.Fatal(err)
		}
		names, ids := map[string]bool{}, map[string]bool{}
		filtered, witnesses := 0, 0
		for _, row := range rows {
			if row.ID == "" || ids[row.ID] || row.ModeFiltered == "" && len(row.Witnesses) == 0 {
				t.Fatal("lost source identity or selected syntax", row)
			}
			ids[row.ID], names[row.Opcode] = true, true
			if row.ModeFiltered != "" {
				filtered++
				if len(row.Witnesses) != 0 {
					t.Fatal("mode-filtered row acquired a selected witness", row)
				}
			}
			for _, witness := range row.Witnesses {
				if !strings.HasPrefix(witness.Instruction, row.Opcode+" ") || len(witness.Encoding) == 0 || witness.Destination == "" {
					t.Fatal("incomplete selected scalar witness", row)
				}
				witnesses++
			}
		}
		if len(rows) != 169 || len(names) != 53 || filtered != mode.filtered || witnesses != mode.witnesses {
			t.Fatal("source/mode/selected grammar denominator changed", mode, len(rows), len(names), filtered, witnesses)
		}
	}
	if _, err := X86ScalarRMWRows(catalog, "arm64"); err == nil {
		t.Fatal("foreign mode entered x86 scalar projection")
	}
	if _, err := X86ScalarRMWRows(X86LegacyCatalog{}, "amd64"); err == nil {
		t.Fatal("missing source closure was treated as successful empty coverage")
	}
	for index, entry := range catalog.Optab {
		if entry.Opcode != "ADCB" {
			continue
		}
		changed := catalog
		changed.Optab = append(append([]X86LegacyOpcode(nil), catalog.Optab[:index]...), catalog.Optab[index+1:]...)
		if _, err := X86ScalarRMWRows(changed, "amd64"); err == nil {
			t.Fatal("missing family member admitted")
		}
		changed.Optab = append(append([]X86LegacyOpcode(nil), catalog.Optab...), entry)
		if _, err := X86ScalarRMWRows(changed, "amd64"); err == nil {
			t.Fatal("duplicate source member admitted")
		}
		break
	}
}

func TestX86ScalarRMWSelectedPayloadAnchors(t *testing.T) {
	catalog, err := LoadX86LegacyRules(runtime.GOROOT())
	if err != nil {
		t.Fatal(err)
	}
	anchors := []struct {
		mode, instruction string
		code              []byte
	}{
		{"amd64", "ADDB $128,AL", []byte{0x04, 0x80}},
		{"amd64", "ADDW $128,AX", []byte{0x66, 0x05, 0x80, 0}},
		{"amd64", "ADDL $1,BX", []byte{0x83, 0xc3, 1}},
		{"amd64", "ADDQ $128,BX", []byte{0x48, 0x81, 0xc3, 0x80, 0, 0, 0}},
		{"amd64", "SUBL CX,0(SI)", []byte{0x29, 0x0e}},
		{"amd64", "SUBL 0(SI),BX", []byte{0x2b, 0x1e}},
		{"amd64", "BTSQ $1,BX", []byte{0x48, 0x0f, 0xba, 0xeb, 1}},
		{"amd64", "BTCW CX,0(SI)", []byte{0x66, 0x0f, 0xbb, 0x0e}},
		{"amd64", "INCL BX", []byte{0xff, 0xc3}},
		{"386", "INCL BX", []byte{0x43}},
		{"386", "DECL 0(SI)", []byte{0xff, 0x0e}},
		{"amd64", "NOTB BL", []byte{0xf6, 0xd3}},
	}
	for _, anchor := range anchors {
		t.Run(anchor.mode+"/"+anchor.instruction, func(t *testing.T) {
			rows, err := X86ScalarRMWRows(catalog, anchor.mode)
			if err != nil {
				t.Fatal(err)
			}
			matches := 0
			for _, row := range rows {
				for _, witness := range row.Witnesses {
					if witness.Instruction == anchor.instruction {
						matches++
						if !bytes.Equal(witness.Encoding, anchor.code) {
							t.Fatalf("payload %x want=%x", witness.Encoding, anchor.code)
						}
					}
				}
			}
			if matches != 1 {
				t.Fatal("selected source must have one row owner", matches)
			}
		})
	}
}

func TestX86ScalarRMWRejectsChangedPhysicalAxes(t *testing.T) {
	catalog, err := LoadX86LegacyRules(runtime.GOROOT())
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range catalog.Optab {
		if entry.Opcode != "ADDQ" {
			continue
		}
		for _, mutation := range []struct {
			name   string
			change func(*X86LegacyOpcode, *X86LegacyOperand)
		}{
			{"class", func(_ *X86LegacyOpcode, r *X86LegacyOperand) { r.Classes = []string{"Yu8", "Yml"} }},
			{"zcase", func(_ *X86LegacyOpcode, r *X86LegacyOperand) { r.ZCase = "Zm_r" }},
			{"stride", func(_ *X86LegacyOpcode, r *X86LegacyOperand) { r.ZOffset++ }},
			{"offset", func(_ *X86LegacyOpcode, r *X86LegacyOperand) { r.UnmatchedOffset++ }},
			{"index", func(_ *X86LegacyOpcode, r *X86LegacyOperand) { r.RowIndex = 1 }},
			{"prefix", func(e *X86LegacyOpcode, _ *X86LegacyOperand) { e.Prefix = "Pe" }},
			{"escape", func(e *X86LegacyOpcode, _ *X86LegacyOperand) { e.EscapeAdjustment = 1 }},
			{"truncated", func(e *X86LegacyOpcode, _ *X86LegacyOperand) { e.OpBytes = e.OpBytes[:1] }},
			{"extra", func(e *X86LegacyOpcode, _ *X86LegacyOperand) {
				e.OpBytes = append(append([]string(nil), e.OpBytes...), "0")
			}},
			{"symbol", func(e *X86LegacyOpcode, _ *X86LegacyOperand) {
				e.OpBytes = append([]string(nil), e.OpBytes...)
				e.OpBytes[0] = "opcode"
			}},
			{"extension", func(e *X86LegacyOpcode, _ *X86LegacyOperand) {
				e.OpBytes = append([]string(nil), e.OpBytes...)
				e.OpBytes[1] = "8"
			}},
		} {
			t.Run(mutation.name, func(t *testing.T) {
				row := entry.Rows[0]
				if witnesses, err := x86ScalarRMWRowWitnesses(entry, row, "amd64"); err != nil || len(witnesses) != 2 {
					t.Fatal("valid original failed before negative", err)
				}
				changed := entry
				mutation.change(&changed, &row)
				if _, err := x86ScalarRMWRowWitnesses(changed, row, "amd64"); err == nil {
					t.Fatal("changed physical axis admitted", fmt.Sprint(changed, row))
				}
			})
		}
		return
	}
	t.Fatal("actual Go omitted ADDQ")
}
