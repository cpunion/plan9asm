package goasmrules

import (
	"fmt"
	"strconv"
	"strings"
)

// X86X87Row retains an original row, not an initialized x87 value or native
// entry. Register carrier, memory type/width, rounding and state are separate.
type X86X87Row struct {
	ID, Opcode string
	Witnesses  []X86ScalarRMWWitness
}

func x86X87RowSchema(table string) []string {
	switch table {
	case "yfmvx":
		return []string{"Zm_o/2/Ym,Yf0"}
	case "yfadd":
		return []string{"Zm_o/2/Ym,Yf0", "Zm_o/2/Yrf,Yf0", "Zo_m/2/Yf0,Yrf"}
	case "yfcmv":
		return []string{"Zm_o/2/Yrf,Yf0"}
	case "ycompp":
		return []string{"Zo_m/2/Yf0,Yrf"}
	default:
		return nil
	}
}

// X86X87Rows projects four complete ordered tables through two directions and
// three carrier shapes. All bytes come from the original selected payload.
func X86X87Rows(catalog X86LegacyCatalog, arch string) ([]X86X87Row, error) {
	if arch != "amd64" && arch != "386" {
		return nil, fmt.Errorf("x87 source transport requires an explicit Go x86 mode")
	}
	var rows []X86X87Row
	for _, entry := range catalog.Optab {
		if x86X87RowSchema(entry.Ytab) == nil {
			continue
		}
		for _, operand := range entry.Rows {
			witnesses, err := x86X87RowWitnesses(entry, operand)
			if err != nil {
				return nil, fmt.Errorf("%s/%d: %w", entry.Opcode, operand.RowIndex, err)
			}
			rows = append(rows, X86X87Row{
				ID:     fmt.Sprintf("optab/%d/%s/%s/%d", entry.TableIndex, entry.Opcode, entry.Ytab, operand.RowIndex),
				Opcode: entry.Opcode, Witnesses: witnesses,
			})
		}
	}
	if len(rows) != 79 {
		return nil, fmt.Errorf("incomplete original x87 table closure: rows=%d", len(rows))
	}
	return rows, nil
}

func x86X87RowWitnesses(entry X86LegacyOpcode, row X86LegacyOperand) ([]X86ScalarRMWWitness, error) {
	schema := x86X87RowSchema(entry.Ytab)
	if schema == nil || entry.Kind != "encoding" || entry.Prefix != "Px" || entry.EscapeAdjustment != 0 ||
		len(entry.Rows) != len(schema) || len(entry.OpBytes) != 2*len(schema) || row.RowIndex < 0 || row.RowIndex >= len(schema) {
		return nil, fmt.Errorf("unmodeled complete x87 source schema")
	}
	var payload [6]byte
	for index, original := range entry.Rows {
		shape := fmt.Sprintf("%s/%d/%s", original.ZCase, original.ZOffset, strings.Join(original.Classes, ","))
		if original.RowIndex != index || original.UnmatchedOffset != 2*index || shape != schema[index] {
			return nil, fmt.Errorf("ordered x87 source row changed")
		}
		for atom := 0; atom < 2; atom++ {
			value, err := strconv.ParseUint(entry.OpBytes[2*index+atom], 0, 8)
			if err != nil || atom == 0 && (value < 0xd8 || value > 0xdf) ||
				atom == 1 && original.Classes[0] == "Ym" && value > 7 {
				return nil, fmt.Errorf("x87 source opcode/ModRM atom changed")
			}
			payload[2*index+atom] = byte(value)
		}
	}
	if row.UnmatchedOffset != 2*row.RowIndex ||
		fmt.Sprintf("%s/%d/%s", row.ZCase, row.ZOffset, strings.Join(row.Classes, ",")) != schema[row.RowIndex] {
		return nil, fmt.Errorf("selected x87 source row differs")
	}
	opcode, fragment := payload[row.UnmatchedOffset], payload[row.UnmatchedOffset+1]
	var witnesses []X86ScalarRMWWitness
	if row.Classes[0] == "Ym" {
		for _, displacement := range []bool{false, true} {
			memory, code := "(AX)", []byte{opcode, fragment << 3}
			if displacement {
				memory, code = "7(AX)", []byte{opcode, 0x40 | fragment<<3, 7}
			}
			witnesses = append(witnesses, X86ScalarRMWWitness{Instruction: entry.Opcode + " " + memory + ",F0", Destination: "F0", Encoding: code})
		}
	} else {
		for _, lane := range []byte{1, 7} {
			first, last := fmt.Sprint("F", lane), "F0"
			if row.ZCase == "Zo_m" {
				first, last = last, first
			}
			// Go asmando ORs its full field fragment with the fixed register
			// mode. In particular FUCOMPP's decimal 13 is not a REX bit.
			witnesses = append(witnesses, X86ScalarRMWWitness{Instruction: entry.Opcode + " " + first + "," + last,
				Destination: last, Encoding: []byte{opcode, 0xc0 | lane | fragment<<3}})
		}
	}
	return witnesses, nil
}
