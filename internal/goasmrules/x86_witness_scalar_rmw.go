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

// X86ConditionalMoveRows reuses the scalar syntax/byte containers and
// transport. Conditional reads, old destinations and FLAGS remain owned by
// the typed family; this projection supplies no numerical or storage facts.
func X86ConditionalMoveRows(catalog X86LegacyCatalog, arch string) ([]X86ScalarRMWRow, error) {
	if arch != "386" && arch != "amd64" {
		return nil, fmt.Errorf("conditional move requires a Go x86 mode")
	}
	var result []X86ScalarRMWRow
	seen := map[string]bool{}
	for _, entry := range catalog.Optab {
		if !strings.HasPrefix(entry.Opcode, "CMOV") {
			continue
		}
		if seen[entry.Opcode] || len(entry.Opcode) != 7 || len(entry.Rows) != 1 {
			return nil, fmt.Errorf("duplicate or malformed conditional-move source member")
		}
		seen[entry.Opcode] = true
		operand := entry.Rows[0]
		witnesses, err := x86ScalarRMWRowWitnesses(entry, operand, arch)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", entry.Opcode, err)
		}
		row := X86ScalarRMWRow{ID: fmt.Sprintf("optab/%d/%s/%s/%d", entry.TableIndex, entry.Opcode, entry.Ytab, operand.RowIndex), Opcode: entry.Opcode}
		if arch == "386" && entry.Prefix == "Pw" {
			row.ModeFiltered = "Pw requires amd64"
		} else {
			row.Witnesses = witnesses
		}
		result = append(result, row)
	}
	if len(seen) != 48 || len(result) != 48 {
		return nil, fmt.Errorf("incomplete conditional-move source closure: names=%d rows=%d", len(seen), len(result))
	}
	return result, nil
}

// X86ScalarRMRegisterRows retains the non-CMOV yml_rl source closure. Mode
// filtering follows the independently observed original Go REX.W refusals;
// operand spellings and bytes still come from the shared scalar transport.
func X86ScalarRMRegisterRows(catalog X86LegacyCatalog, arch string) ([]X86ScalarRMWRow, error) {
	if arch != "386" && arch != "amd64" {
		return nil, fmt.Errorf("scalar rm-to-register requires a Go x86 mode")
	}
	var result []X86ScalarRMWRow
	seen := map[string]bool{}
	for _, entry := range catalog.Optab {
		if entry.Ytab != "yml_rl" || strings.HasPrefix(entry.Opcode, "CMOV") {
			continue
		}
		if seen[entry.Opcode] || len(entry.Rows) != 1 {
			return nil, fmt.Errorf("duplicate or malformed scalar rm-to-register source")
		}
		seen[entry.Opcode] = true
		witnesses, err := x86ScalarRMWRowWitnesses(entry, entry.Rows[0], arch)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", entry.Opcode, err)
		}
		row := X86ScalarRMWRow{ID: fmt.Sprintf("optab/%d/%s/%s/0", entry.TableIndex, entry.Opcode, entry.Ytab), Opcode: entry.Opcode}
		if arch == "386" && (entry.Prefix == "Pw" || entry.Prefix == "Pfw" || entry.Prefix == "Pq4w" || entry.Prefix == "Pq5w") {
			row.ModeFiltered = "original Go source rejects REX.W in 386 mode"
		} else {
			row.Witnesses = witnesses
		}
		result = append(result, row)
	}
	if len(result) != 28 || len(seen) != 28 {
		return nil, fmt.Errorf("incomplete scalar rm-to-register closure: names=%d rows=%d", len(seen), len(result))
	}
	return result, nil
}

// X86ScalarMultiplyMOVBERows joins the complete ordered integer multiply and
// byte-order-move tables. Implicit accumulator, old destination, FLAGS and
// memory inputs remain the existing typed families' responsibility.
func X86ScalarMultiplyMOVBERows(catalog X86LegacyCatalog, arch string) ([]X86ScalarRMWRow, error) {
	if arch != "386" && arch != "amd64" {
		return nil, fmt.Errorf("scalar multiply/MOVBE requires a Go x86 mode")
	}
	var result []X86ScalarRMWRow
	seen := map[string]bool{}
	for _, entry := range catalog.Optab {
		if entry.Ytab != "yimul" && entry.Ytab != "yimul3" && entry.Ytab != "ymovbe" {
			continue
		}
		if seen[entry.Opcode] {
			return nil, fmt.Errorf("duplicate scalar multiply/MOVBE source member")
		}
		seen[entry.Opcode] = true
		for _, operand := range entry.Rows {
			witnesses, err := x86ScalarRMWRowWitnesses(entry, operand, arch)
			if err != nil {
				return nil, fmt.Errorf("%s/%d: %w", entry.Opcode, operand.RowIndex, err)
			}
			row := X86ScalarRMWRow{ID: fmt.Sprintf("optab/%d/%s/%s/%d", entry.TableIndex, entry.Opcode, entry.Ytab, operand.RowIndex),
				Opcode: entry.Opcode, Witnesses: witnesses}
			if arch == "386" && entry.Prefix == "Pw" {
				row.ModeFiltered, row.Witnesses = "original Go rejects REX.W in 386 mode", nil
			}
			result = append(result, row)
		}
	}
	if len(seen) != 9 || len(result) != 24 {
		return nil, fmt.Errorf("incomplete scalar multiply/MOVBE closure: names=%d rows=%d", len(seen), len(result))
	}
	return result, nil
}

// X86ScalarImplicitMulDivRows retains the four complete integer families.
// Other users of ydivb/ydivl (descriptor/state operations and NOPs) have their
// own contracts; sharing an operand table grants no numerical authority.
func X86ScalarImplicitMulDivRows(catalog X86LegacyCatalog, arch string) ([]X86ScalarRMWRow, error) {
	if arch != "amd64" && arch != "386" {
		return nil, fmt.Errorf("implicit integer rows require a Go x86 mode")
	}
	var result []X86ScalarRMWRow
	seen := map[string]bool{}
	for _, entry := range catalog.Optab {
		stem, width, known := x86ScalarImplicitFamily(entry.Opcode)
		if !known {
			continue
		}
		table := "ydivl"
		if width == 'B' {
			table = "ydivb"
		} else if stem == "IMUL" {
			table = "yimul"
		}
		if seen[entry.Opcode] || entry.Ytab != table || len(entry.Rows) == 0 {
			return nil, fmt.Errorf("duplicate or changed implicit integer source table")
		}
		seen[entry.Opcode] = true
		witnesses, err := x86ScalarRMWRowWitnesses(entry, entry.Rows[0], arch)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", entry.Opcode, err)
		}
		row := X86ScalarRMWRow{ID: fmt.Sprintf("optab/%d/%s/%s/0", entry.TableIndex, entry.Opcode, entry.Ytab),
			Opcode: entry.Opcode, Witnesses: witnesses}
		if arch == "386" && width == 'Q' {
			row.ModeFiltered, row.Witnesses = "original Go rejects REX.W in 386 mode", nil
		}
		result = append(result, row)
	}
	if len(seen) != 16 {
		return nil, fmt.Errorf("incomplete implicit integer source families: names=%d", len(seen))
	}
	return result, nil
}

func x86ScalarImplicitFamily(name string) (string, byte, bool) {
	if len(name) < 4 || !strings.ContainsRune("BWLQ", rune(name[len(name)-1])) {
		return "", 0, false
	}
	stem, width := name[:len(name)-1], name[len(name)-1]
	return stem, width, stem == "MUL" || stem == "IMUL" || stem == "DIV" || stem == "IDIV"
}

// Validate every original row before selecting a byte offset. The selected
// row alone cannot establish the stride, arity or identity of earlier rows.
// Exact schemas and a 0/1 source escape bound all subsequent offset arithmetic.
func x86ScalarCompleteOperandSchema(entry X86LegacyOpcode, row X86LegacyOperand, schema []string) error {
	if len(schema) == 0 || len(entry.Rows) != len(schema) || row.RowIndex < 0 || row.RowIndex >= len(schema) ||
		entry.EscapeAdjustment < 0 || entry.EscapeAdjustment > 1 {
		return fmt.Errorf("unmodeled complete scalar operand schema/escape")
	}
	offset := 0
	for index, original := range entry.Rows {
		fields := strings.Split(schema[index], "/")
		if len(fields) != 3 {
			return fmt.Errorf("invalid internal scalar schema")
		}
		classes := strings.Split(fields[2], ",")
		stride, err := strconv.Atoi(fields[1])
		if err != nil || stride <= 0 || original.RowIndex != index || original.ZCase != fields[0] ||
			original.ZOffset != stride || original.UnmatchedOffset != offset || len(original.Classes) != len(classes) {
			return fmt.Errorf("scalar complete original row identity changed")
		}
		for position, class := range classes {
			if original.Classes[position] != class {
				return fmt.Errorf("scalar complete original operand class changed")
			}
		}
		offset += stride + entry.EscapeAdjustment
	}
	original := entry.Rows[row.RowIndex]
	if row.ZCase != original.ZCase || row.ZOffset != original.ZOffset || row.UnmatchedOffset != original.UnmatchedOffset ||
		len(row.Classes) != len(original.Classes) {
		return fmt.Errorf("scalar selected row differs from complete original source")
	}
	for index, class := range original.Classes {
		if row.Classes[index] != class {
			return fmt.Errorf("scalar selected operand classes differ from original row")
		}
	}
	return nil
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
		"yml_rl": {"Zm_r/1/Yml,Yrl"},
		"ycmpb":  {"Z_ib/1/Yal,Yi32", "Zm_ibo/2/Ymb,Yi32", "Zm_r/1/Ymb,Yrb", "Zr_m/1/Yrb,Ymb"},
		"ycmpl":  {"Zm_ibo/2/Yml,Yi8", "Z_il/1/Yax,Yi32", "Zm_ilo/2/Yml,Yi32", "Zm_r/1/Yml,Yrl", "Zr_m/1/Yrl,Yml"},
		"ytestl": {"Zil_/1/Yi32,Yax", "Zilo_m/2/Yi32,Yml", "Zr_m/1/Yrl,Yml", "Zm_r/1/Yml,Yrl"},
		"yimul":  {"Zm_o/2/Yml", "Zib_rr/1/Yi8,Yrl", "Zil_rr/1/Yi32,Yrl", "Zm_r/2/Yml,Yrl"},
		"yimul3": {"Zibm_r/2/Yi8,Yml,Yrl", "Zibm_r/2/Yi32,Yml,Yrl"},
		"ymovbe": {"Zlitm_r/3/Ym,Yrl", "Zlitr_m/3/Yrl,Ym"},
		"ydivb":  {"Zm_o/2/Ymb"},
		"ydivl":  {"Zm_o/2/Yml"},
	}
	schema := schemas[entry.Ytab]
	if entry.Opcode == "" {
		return nil, fmt.Errorf("missing scalar source opcode")
	}
	width := entry.Opcode[len(entry.Opcode)-1]
	if entry.Ytab == "ydivb" || entry.Ytab == "ydivl" {
		stem, actualWidth, known := x86ScalarImplicitFamily(entry.Opcode)
		if !known || actualWidth == 'B' && entry.Ytab != "ydivb" || actualWidth != 'B' &&
			(entry.Ytab != "ydivl" || stem == "IMUL") {
			return nil, fmt.Errorf("shared divide table member is outside the implicit integer family")
		}
	}
	conditional := strings.HasPrefix(entry.Opcode, "CMOV")
	remainingPrefix := ""
	if conditional {
		if len(entry.Opcode) != 7 || entry.Ytab != "yml_rl" || !strings.ContainsRune("WLQ", rune(entry.Opcode[4])) {
			return nil, fmt.Errorf("unmodeled conditional-move source shape")
		}
		width = entry.Opcode[4]
	} else if entry.Ytab == "yml_rl" {
		width, remainingPrefix = x86ScalarYmlRegisterPrefix(entry.Opcode)
		if remainingPrefix == "" {
			return nil, fmt.Errorf("scalar source family does not own this table")
		}
	}
	expectedPrefix := map[byte]string{'B': "Pb", 'W': "Pe", 'L': "Px", 'Q': "Pw"}[width]
	if remainingPrefix != "" {
		expectedPrefix = remainingPrefix
	} else if (entry.Ytab == "ybtl" || conditional) && width != 'Q' {
		expectedPrefix = map[byte]string{'W': "Pq", 'L': "Pm"}[width]
	} else if entry.Ytab == "yincl" {
		expectedPrefix = "Px1"
	} else if entry.Ytab == "ymovbe" {
		expectedPrefix = map[byte]string{'W': "Pq", 'L': "Pm", 'Q': "Pw"}[width]
	}
	if entry.Kind != "encoding" || expectedPrefix == "" || entry.Prefix != expectedPrefix {
		return nil, fmt.Errorf("unmodeled complete scalar operand schema/prefix")
	}
	if err := x86ScalarCompleteOperandSchema(entry, row, schema); err != nil {
		return nil, err
	}
	multiplyMove := entry.Ytab == "yimul" || entry.Ytab == "yimul3" || entry.Ytab == "ymovbe"
	if multiplyMove && !strings.ContainsRune("WLQ", rune(width)) {
		return nil, fmt.Errorf("scalar carrier is outside the complete table grammar")
	}
	var payload []byte
	for index, token := range entry.OpBytes {
		if entry.Ytab == "yimul" && index == 4 {
			if token != "Pm" {
				return nil, fmt.Errorf("scalar multiply source lost its symbolic escape")
			}
			payload = append(payload, 0x0f) // Go's Pm opcode-escape atom.
			continue
		}
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
	if entry.EscapeAdjustment != adjustment {
		return nil, fmt.Errorf("scalar original leading byte and source escape differ")
	}
	if multiplyMove {
		for _, original := range entry.Rows {
			start, size := original.UnmatchedOffset, original.ZOffset+entry.EscapeAdjustment
			if start < 0 || start+size > len(payload) {
				return nil, fmt.Errorf("scalar complete source payload is truncated")
			}
			for index, atom := range payload[start : start+size] {
				terminator := entry.Ytab != "yimul" && index == size-1
				if terminator != (atom == 0) || entry.Ytab == "yimul" && start == 0 && index == 1 && atom > 7 {
					return nil, fmt.Errorf("scalar literal terminator or extension changed")
				}
			}
			if entry.Ytab == "ymovbe" && (entry.EscapeAdjustment == 1) != (payload[start] == 0x0f) {
				return nil, fmt.Errorf("scalar literal source escape changed")
			}
		}
	}
	offset := 0
	for _, previous := range entry.Rows[:row.RowIndex] {
		offset += previous.ZOffset + adjustment
	}
	last := entry.Rows[len(entry.Rows)-1]
	if entry.EscapeAdjustment != adjustment || row.UnmatchedOffset != offset ||
		last.UnmatchedOffset+last.ZOffset+adjustment != len(payload) || offset < 0 || offset >= len(payload) {
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
	case "Pf3":
		prefix = []byte{0xf3, 0x0f}
	case "Pef3":
		prefix = []byte{0x66, 0xf3, 0x0f}
	case "Pfw":
		prefix = []byte{0xf3, 0x48, 0x0f}
	case "Pq4":
		prefix = []byte{0x66, 0x0f, 0x38}
	case "Pq4w":
		prefix = []byte{0x66, 0x48, 0x0f, 0x38}
	case "Pq5":
		prefix = []byte{0xf3, 0x0f, 0x38}
	case "Pq5w":
		prefix = []byte{0xf3, 0x48, 0x0f, 0x38}
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
	compare := entry.Ytab == "ycmpl" || entry.Ytab == "ycmpb"
	switch row.ZCase {
	case "Zib_", "Zil_":
		operands = append(operands, [2]string{"$128", accumulator})
	case "Z_ib", "Z_il":
		operands = append(operands, [2]string{accumulator, "$128"})
	case "Zm_ibo", "Zm_ilo":
		if row.Classes[1] == "Yi8" {
			immediate = 1
		}
		operands = append(operands, [2]string{dst, fmt.Sprint("$", immediate)}, [2]string{"0(SI)", fmt.Sprint("$", immediate)})
	case "Zibo_m", "Zilo_m":
		if row.Classes[0] == "Yi8" {
			immediate = 1
		}
		operands = append(operands, [2]string{fmt.Sprint("$", immediate), dst}, [2]string{fmt.Sprint("$", immediate), "0(SI)"})
	case "Zib_rr", "Zil_rr", "Zibm_r":
		if row.Classes[0] == "Yi8" {
			immediate = 1
		}
		input := fmt.Sprint("$", immediate)
		if row.ZCase == "Zibm_r" {
			operands = append(operands, [2]string{input + "," + src, dst}, [2]string{input + ",0(SI)", dst})
		} else {
			operands = append(operands, [2]string{input, dst})
		}
	case "Zr_m":
		if !compare {
			operands = append(operands, [2]string{src, dst})
		}
		operands = append(operands, [2]string{src, "0(SI)"})
	case "Zm_r":
		if entry.Ytab == "yml_rl" || entry.Ytab == "yimul" || compare {
			operands = append(operands, [2]string{src, dst})
		}
		operands = append(operands, [2]string{"0(SI)", dst})
	case "Zlitm_r":
		operands = append(operands, [2]string{"0(SI)", dst})
	case "Zlitr_m":
		operands = append(operands, [2]string{src, "0(SI)"})
	case "Zm_o":
		operands = append(operands, [2]string{"", src}, [2]string{"", "0(SI)"})
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
		memory := args[0] == "0(SI)" || args[1] == "0(SI)" || row.ZCase == "Zibm_r" && strings.HasSuffix(args[0], ",0(SI)")
		rm := byte(3)
		if memory {
			rm = 6
		} else {
			rm |= 0xc0
			if row.ZCase == "Zm_r" || row.ZCase == "Zibm_r" || row.ZCase == "Zm_o" {
				rm = 0xc1
			} // CX source, BX destination.
		}
		switch row.ZCase {
		case "Zr_m":
			code = append(code, rm|1<<3)
		case "Zm_r":
			code = append(code, rm|3<<3)
		case "Zib_rr", "Zil_rr", "Zibm_r":
			code = append(code, rm|3<<3)
		case "Zlitm_r", "Zlitr_m":
			end := offset + 1
			for end < len(payload) && payload[end] != 0 {
				code = append(code, payload[end])
				end++
			}
			register := byte(3)
			if row.ZCase == "Zlitr_m" {
				register = 1
			}
			code = append(code, rm|register<<3)
		case "Zibo_m", "Zilo_m", "Zm_ibo", "Zm_ilo", "Zo_m", "Zm_o":
			if offset+1 >= len(payload) || payload[offset+1] > 7 {
				return nil, fmt.Errorf("scalar /digit is not a three-bit extension")
			}
			code = append(code, rm|payload[offset+1]<<3)
		case "Z_rp":
			code[len(code)-1] += 3
		}
		if strings.HasPrefix(args[0], "$") || compare && strings.HasPrefix(args[1], "$") {
			n := 1
			if row.ZCase == "Zil_" || row.ZCase == "Zilo_m" || row.ZCase == "Z_il" || row.ZCase == "Zm_ilo" ||
				row.ZCase == "Zil_rr" || row.ZCase == "Zibm_r" && row.Classes[0] == "Yi32" {
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
		destination := args[1]
		if row.ZCase == "Zm_o" {
			destination = "AX" // B uses AL/AH; the separate DX output remains a semantic effect.
		}
		witnesses = append(witnesses, X86ScalarRMWWitness{Instruction: instruction + args[1], Destination: destination, Encoding: code})
	}
	return witnesses, nil
}

// This is family/prefix grammar, not an opcode-byte table. The original Go
// payload remains the only instruction-byte source. GP spellings use the Yrl
// carrier; source-memory tuples and result definedness belong to typed effects.
func x86ScalarYmlRegisterPrefix(name string) (byte, string) {
	if len(name) < 2 {
		return 0, ""
	}
	if strings.HasPrefix(name, "MOV") {
		if len(name) != 7 || name[5:] != "SX" && name[5:] != "ZX" ||
			!(name[3] == 'W' && (name[4] == 'L' || name[4] == 'Q') || name[3] == 'L' && name[4] == 'Q') {
			return 0, ""
		}
		if name[3] == 'L' && name[5:] == "ZX" {
			return 'L', "Px" // Ordinary MOVL encoding; AMD64 zero-extends its GP result.
		}
		return name[4], map[byte]string{'L': "Pm", 'Q': "Pw"}[name[4]]
	}
	width, stem := name[len(name)-1], name[:len(name)-1]
	switch stem {
	case "BSF", "BSR", "LAR", "LSL":
		return width, map[byte]string{'W': "Pq", 'L': "Pm", 'Q': "Pw"}[width]
	case "LZCNT", "POPCNT":
		return width, map[byte]string{'W': "Pef3", 'L': "Pf3", 'Q': "Pfw"}[width]
	case "ADCX", "ADOX":
		prefix := "Pq4"
		if stem == "ADOX" {
			prefix = "Pq5"
		}
		if width == 'Q' {
			return width, prefix + "w"
		}
		if width == 'L' {
			return width, prefix
		}
	}
	return 0, ""
}
