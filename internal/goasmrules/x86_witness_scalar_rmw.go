package goasmrules

import (
	"fmt"
	"strconv"
	"strings"
)

// X86ScalarRMWRow retains every source row, including mode-filtered rows.
// Candidates describe selected syntax/bytes, not numerical or storage facts.
type X86ScalarRMWRow struct {
	ID, Opcode, ModeFiltered string
	Witnesses                []X86ScalarRMWWitness
}

type X86ScalarRMWWitness struct {
	Instruction, Destination string
	Encoding                 []byte
}

// X86ScalarRMWRows projects the six complete legacy scalar RMW tables. The
// original reader supplies all ordered attributes and opcode payloads; this
// adapter does not carry a second per-mnemonic opcode table.
func X86ScalarRMWRows(catalog X86LegacyCatalog, arch string) ([]X86ScalarRMWRow, error) {
	if arch != "386" && arch != "amd64" {
		return nil, fmt.Errorf("scalar RMW requires a Go x86 mode")
	}
	var result []X86ScalarRMWRow
	seen := map[string]bool{}
	for _, entry := range catalog.Optab {
		name := entry.Opcode
		if len(name) < 3 {
			continue
		}
		stem, width := name[:len(name)-1], name[len(name)-1]
		if !strings.Contains(" ADC ADD AND OR SBB SUB XOR INC DEC NEG NOT BTC BTR BTS ", " "+stem+" ") ||
			!strings.ContainsRune("BWLQ", rune(width)) || strings.HasPrefix(stem, "BT") && width == 'B' {
			continue
		}
		if seen[name] {
			return nil, fmt.Errorf("duplicate scalar RMW source member %s", name)
		}
		seen[name] = true
		for _, operand := range entry.Rows {
			row := X86ScalarRMWRow{ID: fmt.Sprintf("optab/%d/%s/%s/%d", entry.TableIndex, name, entry.Ytab, operand.RowIndex), Opcode: name}
			witnesses, err := x86ScalarRMWRowWitnesses(entry, operand, arch)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", row.ID, err)
			}
			switch {
			case arch == "386" && entry.Prefix == "Pw":
				row.ModeFiltered = "Pw requires amd64"
			case arch == "amd64" && entry.Prefix == "Px1" && operand.RowIndex == 0:
				row.ModeFiltered = "amd64 skips Px1 short-register row"
			default:
				row.Witnesses = witnesses
			}
			result = append(result, row)
		}
	}
	if len(seen) != 53 || len(result) != 169 {
		return nil, fmt.Errorf("incomplete scalar RMW source closure: names=%d rows=%d", len(seen), len(result))
	}
	return result, nil
}

func x86ScalarRMWRowWitnesses(entry X86LegacyOpcode, row X86LegacyOperand, arch string) ([]X86ScalarRMWWitness, error) {
	// These are ordered physical schemas, not opcode values. In particular,
	// /digit immediate extension and ModRM direction have different transports.
	schemas := map[string][]string{
		"yxorb":  {"Zib_/1/Yi32,Yal", "Zibo_m/2/Yi32,Ymb", "Zr_m/1/Yrb,Ymb", "Zm_r/1/Ymb,Yrb"},
		"yaddl":  {"Zibo_m/2/Yi8,Yml", "Zil_/1/Yi32,Yax", "Zilo_m/2/Yi32,Yml", "Zr_m/1/Yrl,Yml", "Zm_r/1/Yml,Yrl"},
		"yincl":  {"Z_rp/1/Yrl", "Zo_m/2/Yml"},
		"yincq":  {"Zo_m/2/Yml"},
		"yscond": {"Zo_m/2/Ymb"},
		"ybtl":   {"Zibo_m/2/Yi8,Yml", "Zr_m/1/Yrl,Yml"},
	}
	schema := schemas[entry.Ytab]
	width := entry.Opcode[len(entry.Opcode)-1]
	expectedPrefix := map[byte]string{'B': "Pb", 'W': "Pe", 'L': "Px", 'Q': "Pw"}[width]
	if entry.Ytab == "ybtl" && width != 'Q' {
		expectedPrefix = map[byte]string{'W': "Pq", 'L': "Pm"}[width]
	} else if entry.Ytab == "yincl" {
		expectedPrefix = "Px1"
	}
	if entry.Kind != "encoding" || expectedPrefix == "" || entry.Prefix != expectedPrefix ||
		len(entry.Rows) != len(schema) || row.RowIndex < 0 || row.RowIndex >= len(schema) ||
		fmt.Sprintf("%s/%d/%s", row.ZCase, row.ZOffset, strings.Join(row.Classes, ",")) != schema[row.RowIndex] {
		return nil, fmt.Errorf("unmodeled complete scalar operand schema/prefix")
	}
	var payload []byte
	for _, token := range entry.OpBytes {
		value, err := strconv.ParseUint(token, 0, 8)
		if err != nil {
			return nil, fmt.Errorf("scalar opcode atom is not a byte")
		}
		payload = append(payload, byte(value))
	}
	adjustment := 0
	if len(payload) != 0 && payload[0] == 0x0f {
		adjustment = 1
	}
	offset := 0
	for _, previous := range entry.Rows[:row.RowIndex] {
		offset += previous.ZOffset + adjustment
	}
	last := entry.Rows[len(entry.Rows)-1]
	if entry.EscapeAdjustment != adjustment || row.UnmatchedOffset != offset ||
		last.UnmatchedOffset+last.ZOffset+adjustment != len(payload) || offset >= len(payload) {
		return nil, fmt.Errorf("scalar complete payload/ordered offset changed")
	}
	var prefix []byte
	switch entry.Prefix {
	case "Pe":
		prefix = []byte{0x66}
	case "Pm":
		prefix = []byte{0x0f}
	case "Pq":
		prefix = []byte{0x66, 0x0f}
	case "Pw":
		prefix = []byte{0x48}
	}
	if payload[offset] == 0x0f {
		prefix = append(prefix, 0x0f)
		offset++
	}
	if offset >= len(payload) {
		return nil, fmt.Errorf("missing scalar opcode")
	}
	byteOperand := width == 'B'
	src, dst, accumulator := "CX", "BX", "AX"
	if byteOperand {
		src, dst, accumulator = "CL", "BL", "AL"
	}
	// Earlier matching classes own register/register and small-immediate
	// shadows. The late direction is witnessed by memory source instead.
	var operands [][2]string
	immediate := 128
	switch row.ZCase {
	case "Zib_", "Zil_":
		operands = append(operands, [2]string{"$128", accumulator})
	case "Zibo_m", "Zilo_m":
		if row.Classes[0] == "Yi8" {
			immediate = 1
		}
		operands = append(operands, [2]string{fmt.Sprint("$", immediate), dst}, [2]string{fmt.Sprint("$", immediate), "0(SI)"})
	case "Zr_m":
		operands = append(operands, [2]string{src, dst}, [2]string{src, "0(SI)"})
	case "Zm_r":
		operands = append(operands, [2]string{"0(SI)", dst})
	case "Z_rp":
		operands = append(operands, [2]string{"", dst})
	case "Zo_m":
		if entry.Ytab != "yincl" || arch == "amd64" {
			operands = append(operands, [2]string{"", dst})
		}
		operands = append(operands, [2]string{"", "0(SI)"})
	}
	var witnesses []X86ScalarRMWWitness
	for _, args := range operands {
		code := append(append([]byte(nil), prefix...), payload[offset])
		memory := args[0] == "0(SI)" || args[1] == "0(SI)"
		rm := byte(3)
		if memory {
			rm = 6
		} else {
			rm |= 0xc0
		}
		switch row.ZCase {
		case "Zr_m":
			code = append(code, rm|1<<3)
		case "Zm_r":
			code = append(code, rm|3<<3)
		case "Zibo_m", "Zilo_m", "Zo_m":
			if offset+1 >= len(payload) || payload[offset+1] > 7 {
				return nil, fmt.Errorf("scalar /digit is not a three-bit extension")
			}
			code = append(code, rm|payload[offset+1]<<3)
		case "Z_rp":
			code[len(code)-1] += 3
		}
		if strings.HasPrefix(args[0], "$") {
			n := 1
			if row.ZCase == "Zil_" || row.ZCase == "Zilo_m" {
				n = 4
				if width == 'W' {
					n = 2
				}
			}
			for shift := 0; shift < n; shift++ {
				code = append(code, byte(immediate>>uint(shift*8)))
			}
		}
		instruction := entry.Opcode + " "
		if args[0] != "" {
			instruction += args[0] + ","
		}
		witnesses = append(witnesses, X86ScalarRMWWitness{Instruction: instruction + args[1], Destination: args[1], Encoding: code})
	}
	return witnesses, nil
}
