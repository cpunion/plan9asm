package plan9asm

import "testing"

func TestParseDataOnlyFile(t *testing.T) {
	file, err := Parse(ArchARM64, `DATA ·value(SB)/8, $42
GLOBL ·value(SB),RODATA,$8
`)
	if err != nil {
		t.Fatalf("Parse(data-only) error = %v", err)
	}
	if len(file.Funcs) != 0 || len(file.Data) != 1 || len(file.Globl) != 1 {
		t.Fatalf("Parse(data-only) = funcs:%d data:%d globl:%d", len(file.Funcs), len(file.Data), len(file.Globl))
	}
}

func TestParseDataAndGloblDirectives(t *testing.T) {
	file, err := Parse(ArchARM64, `TEXT ·Fn(SB),NOSPLIT,$0-0
	RET

DATA ·tab<>+8(SB)/PTRSIZE, $1
DATA ·str<>(SB)/8, $"hello"
DATA ·symptr<>(SB)/8, $runtime·main(SB)
GLOBL ·tab<>(SB), RODATA, $16
GLOBL ·symptr<>(SB), NOPTR, $(machTimebaseInfo__size)
`)
	if err != nil {
		t.Fatal(err)
	}

	if got, want := len(file.Data), 3; got != want {
		t.Fatalf("len(Data) = %d, want %d", got, want)
	}
	if got, want := len(file.Globl), 2; got != want {
		t.Fatalf("len(Globl) = %d, want %d", got, want)
	}

	if ds := file.Data[0]; ds.Sym != "·tab<>" || ds.Off != 8 || ds.Width != 8 || ds.Value != 1 {
		t.Fatalf("unexpected first DATA: %#v", ds)
	}
	if ds := file.Data[1]; ds.Sym != "·str<>" || string(ds.Payload) != "hello" {
		t.Fatalf("unexpected string DATA payload: %#v", ds)
	}
	payload, err := dataStmtPayload(file.Data[1])
	if err != nil || string(payload[:5]) != "hello" || len(payload) != 8 || payload[5] != 0 {
		t.Fatalf("padded string DATA payload = (%v, %v)", payload, err)
	}
	if ds := file.Data[2]; ds.Sym != "·symptr<>" || ds.Value != 0 {
		t.Fatalf("unexpected symbol DATA placeholder: %#v", ds)
	}

	if gs := file.Globl[0]; gs.Sym != "·tab<>" || gs.Flags != "RODATA" || gs.Size != 16 {
		t.Fatalf("unexpected first GLOBL: %#v", gs)
	}
	if gs := file.Globl[1]; gs.Sym != "·symptr<>" || gs.Flags != "NOPTR" || gs.Size != 64 {
		t.Fatalf("unexpected macro-sized GLOBL: %#v", gs)
	}

	comma, err := parseDATAStmt(ArchARM64, `·comma(SB)/12, $"hello, world"`)
	if err != nil || string(comma.Payload) != "hello, world" {
		t.Fatalf("parse comma string DATA = (%#v, %v)", comma, err)
	}
	if _, err := parseDATAStmt(ArchARM64, `·short(SB)/4, $"hello"`); err == nil {
		t.Fatal("oversized string DATA unexpectedly parsed")
	}
}

func TestParseDataRejectsMalformedPayloads(t *testing.T) {
	for _, stmt := range []string{
		`·missing(SB)/8 $1`,
		`, $1`,
		`·empty(SB)/8,`,
		`·quote(SB)/8, $"unterminated`,
	} {
		if _, err := parseDATAStmt(ArchARM64, stmt); err == nil {
			t.Errorf("parseDATAStmt(%q) unexpectedly succeeded", stmt)
		}
	}
}
