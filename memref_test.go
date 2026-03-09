package plan9asm

import "testing"

func TestParseMemAllowsParenExpandedOffsetBeforeBase(t *testing.T) {
	mem, ok := parseMem("1*8+(0)(SP)")
	if !ok {
		t.Fatal("parseMem failed")
	}
	if mem.Base != Reg("SP") {
		t.Fatalf("base=%q, want SP", mem.Base)
	}
	if mem.Off != 8 {
		t.Fatalf("off=%d, want 8", mem.Off)
	}
}
