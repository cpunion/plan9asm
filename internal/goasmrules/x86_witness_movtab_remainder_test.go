package goasmrules

import (
	"fmt"
	"runtime"
	"testing"
)

func movtabRemainderOriginalRows(t *testing.T) []X86MoveRule {
	t.Helper()
	catalog, err := LoadX86LegacyRules(runtime.GOROOT())
	if err != nil {
		t.Fatal(err)
	}
	var rows []X86MoveRule
	for _, row := range catalog.Movtab {
		if row.Kind == "encoding" && row.Code != "movDoubleShift" && row.Code != "movTLSReg" {
			rows = append(rows, row)
		}
	}
	if len(rows) != 92 {
		t.Fatal("complete non-TLS source formatter denominator changed", len(rows))
	}
	return rows
}

// The old formatter is the existing X86MoveRule byte hook. Its rejection of
// the new roles records the initial feature gap; it must keep that boundary
// after a separate remainder formatter is introduced. Do not add these roles
// to double-shift or infer numerical/native permissions from this test.
func movtabExistingByteHook(row X86MoveRule) ([]X86ScalarRMWWitness, error) {
	return x86ScalarDoubleShiftWitnesses(row)
}

func TestX86MovtabRemainderCompleteSourceTransport(t *testing.T) {
	for _, row := range movtabRemainderOriginalRows(t) {
		t.Run(fmt.Sprintf("%d/%s/%s", row.TableIndex, row.Opcode, row.Code), func(t *testing.T) {
			before := fmt.Sprintf("%#v", row)
			witnesses, err := movtabExistingByteHook(row)
			if err != nil || len(witnesses) == 0 {
				t.Fatal("MOVTAB_REMAINDER_FORMATTER_FEATURE_GAP", row.TableIndex, row.Code, err)
			}
			for _, witness := range witnesses {
				if witness.Instruction == "" || len(witness.Encoding) == 0 {
					t.Fatal("source formatter returned an empty encoding", row, witness)
				}
			}
			if before != fmt.Sprintf("%#v", row) {
				t.Fatal("source formatter mutated its original capture")
			}
		})
	}
}

// Initially the old hook rejects these different source families wholesale.
// These controls become new-formatter metadata tests when that independent
// formatter is wired in; their initial PASS is not a metadata defect repair.
func TestX86MovtabRemainderCompleteSourceMetadataBoundaries(t *testing.T) {
	for _, original := range movtabRemainderOriginalRows(t) {
		t.Run(fmt.Sprintf("%d/%s/%s", original.TableIndex, original.Opcode, original.Code), func(t *testing.T) {
			for _, field := range []string{
				"kind", "negative-index", "empty-opcode", "opcode-expression",
				"from", "from3", "to", "joined-from", "code", "empty-payload", "overflow-capacity",
			} {
				row := original
				row.OpBytes = append([]string(nil), original.OpBytes...)
				switch field {
				case "kind":
					row.Kind = "sentinel"
				case "negative-index":
					row.TableIndex = -1
				case "empty-opcode":
					row.Opcode = ""
				case "opcode-expression":
					row.OpcodeExpression = "obj.ANOP"
				case "from":
					row.From = "Yinvalid"
				case "from3":
					row.From3 = "Yrl"
				case "to":
					row.To = "Yinvalid"
				case "joined-from":
					row.From += ",Ynone"
				case "code":
					row.Code = "movUnknown"
				case "empty-payload":
					row.OpBytes = nil
				case "overflow-capacity":
					row.OpBytes = []string{"0x90", "0", "0", "0", "0"}
				}
				movtabRequireMetadataRefusal(t, row, field)
			}
			for index := range original.OpBytes {
				for _, atom := range []string{"-1", "0x100", "not-a-byte"} {
					row := original
					row.OpBytes = append([]string(nil), original.OpBytes...)
					row.OpBytes[index] = atom
					movtabRequireMetadataRefusal(t, row, fmt.Sprintf("atom-%d-%s", index, atom))
				}
			}
			if original.Code == "movLit" {
				row := original
				row.OpBytes = []string{"0x90", "0x90", "0x90", "0x90"}
				movtabRequireMetadataRefusal(t, row, "unterminated-literal")
			}
		})
	}
}

func movtabRequireMetadataRefusal(t *testing.T, row X86MoveRule, field string) {
	t.Helper()
	defer func() {
		if caught := recover(); caught != nil {
			t.Errorf("MOVTAB_REMAINDER_METADATA_PANIC row=%d field=%s panic=%v", row.TableIndex, field, caught)
		}
	}()
	if got, err := movtabExistingByteHook(row); err == nil || len(got) != 0 {
		t.Errorf("MOVTAB_REMAINDER_METADATA_ACCEPTED row=%d field=%s witnesses=%v", row.TableIndex, field, got)
	}
}

func TestX86MovtabRemainderDoesNotWidenExistingDoubleShift(t *testing.T) {
	catalog, err := LoadX86LegacyRules(runtime.GOROOT())
	if err != nil {
		t.Fatal(err)
	}
	refused, accepted := 0, 0
	for _, row := range catalog.Movtab {
		if row.Kind != "encoding" {
			continue
		}
		got, err := x86ScalarDoubleShiftWitnesses(row)
		if row.Code == "movDoubleShift" {
			if err != nil || len(got) == 0 {
				t.Fatal("existing double-shift row regressed", row, err)
			}
			accepted++
		} else {
			if err == nil || len(got) != 0 {
				t.Fatal("new source family acquired old double-shift membership", row, got)
			}
			refused++
		}
	}
	if accepted != 18 || refused != 94 {
		t.Fatal("old formatter source-family boundary changed", accepted, refused)
	}
}
