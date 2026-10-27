package goasmrules

import (
	"fmt"
	"strconv"
	"strings"
)

// X86PackedBinaryRow retains the owning official table row. Witnesses are
// expectations, not evidence that Go assembled or plan9asm translated them.
type X86PackedBinaryRow struct {
	ID, Opcode string
	Witnesses  []X86PackedBinaryWitness
}

// Unary AVX rows leave Second empty; legacy rows retain their literal second
// operand even when it is destination-only. These syntax fields do not certify
// semantic reads: typed effects own mask and old-destination dependencies.
type X86PackedBinaryWitness struct {
	Instruction, Generation    string
	First, Second, Destination string
	Width                      int
	Memory, Masked, Zeroing    bool
	Broadcast                  bool
	MaskResult                 bool
	Encoding                   []byte
}

// X86PackedIntegerMinMaxRows retains all amd64 legacy and AVX min/max rows.
func X86PackedIntegerMinMaxRows(legacy X86LegacyCatalog, avx []X86AVXRule) ([]X86PackedBinaryRow, error) {
	return x86PackedBinaryRows(legacy, avx, `
		PMAXSB PMAXSW PMAXSD PMAXUB PMAXUW PMAXUD
		PMINSB PMINSW PMINSD PMINUB PMINUW PMINUD
		VPMAXSB VPMAXSW VPMAXSD VPMAXSQ VPMAXUB VPMAXUW VPMAXUD VPMAXUQ
		VPMINSB VPMINSW VPMINSD VPMINSQ VPMINUB VPMINUW VPMINUD VPMINUQ`)
}

// X86PackedAverageMultiplyRows retains the XMM/VEX/EVEX rounded-average,
// word-product, adjacent multiply-add and widened dword-product families.
// MMX rows keep their separate region/EMMS contract and are not included.
func X86PackedAverageMultiplyRows(legacy X86LegacyCatalog, avx []X86AVXRule) ([]X86PackedBinaryRow, error) {
	return x86PackedBinaryRows(legacy, avx, `
		PAVGB PAVGW VPAVGB VPAVGW
		PMULLW PMULHW PMULHUW PMULHRSW VPMULLW VPMULHW VPMULHUW VPMULHRSW
		PMADDUBSW PMADDWL VPMADDUBSW VPMADDWD
		PMULDQ PMULULQ VPMULDQ VPMULUDQ`)
}

// X86PackedLogicalRows retains packed integer and PS/PD bitwise families,
// including the separate EVEX D/Q names. MMX regions remain a separate owner.
func X86PackedLogicalRows(legacy X86LegacyCatalog, avx []X86AVXRule) ([]X86PackedBinaryRow, error) {
	return x86PackedBinaryRows(legacy, avx, `
		PAND PANDN POR PXOR ANDPS ANDPD ANDNPS ANDNPD ORPS ORPD XORPS XORPD
		VPAND VPANDN VPOR VPXOR VANDPS VANDPD VANDNPS VANDNPD VORPS VORPD VXORPS VXORPD
		VPANDD VPANDQ VPANDND VPANDNQ VPORD VPORQ VPXORD VPXORQ`)
}

// X86PackedUnpackRows retains integer and PS/PD low/high interleave families.
// The six MMX rows retain their separate region and memory-access contracts.
func X86PackedUnpackRows(legacy X86LegacyCatalog, avx []X86AVXRule) ([]X86PackedBinaryRow, error) {
	return x86PackedBinaryRows(legacy, avx, `
		PUNPCKLBW PUNPCKHBW PUNPCKLWL PUNPCKHWL PUNPCKLLQ PUNPCKHLQ PUNPCKLQDQ PUNPCKHQDQ
		UNPCKLPS UNPCKHPS UNPCKLPD UNPCKHPD
		VPUNPCKLBW VPUNPCKHBW VPUNPCKLWD VPUNPCKHWD VPUNPCKLDQ VPUNPCKHDQ VPUNPCKLQDQ VPUNPCKHQDQ
		VUNPCKLPS VUNPCKHPS VUNPCKLPD VUNPCKHPD`)
}

// X86PackedNarrowRows retains signed/unsigned saturating dword-to-word and
// word-to-byte packs. The three MMX rows keep their separate region owner.
func X86PackedNarrowRows(legacy X86LegacyCatalog, avx []X86AVXRule) ([]X86PackedBinaryRow, error) {
	return x86PackedBinaryRows(legacy, avx, `
		PACKSSLW PACKSSWB PACKUSDW PACKUSWB VPACKSSDW VPACKSSWB VPACKUSDW VPACKUSWB`)
}

// X86PackedPerLaneShiftRows retains the complete W/D/Q logical-left,
// logical-right and arithmetic-right per-lane shift families.
func X86PackedPerLaneShiftRows(legacy X86LegacyCatalog, avx []X86AVXRule) ([]X86PackedBinaryRow, error) {
	return x86PackedBinaryRows(legacy, avx, `
		VPSLLVW VPSLLVD VPSLLVQ VPSRLVW VPSRLVD VPSRLVQ VPSRAVW VPSRAVD VPSRAVQ`)
}

// X86PackedSignRows retains all B/W/D conditional-sign rows. Go's legacy
// table is XMM-only; VEX adds X/Y forms, without MMX, EVEX or broadcast.
func X86PackedSignRows(legacy X86LegacyCatalog, avx []X86AVXRule) ([]X86PackedBinaryRow, error) {
	return x86PackedBinaryRows(legacy, avx, `PSIGNB PSIGNW PSIGND VPSIGNB VPSIGNW VPSIGND`)
}

// X86PackedAbsRows retains all unary B/W/D/Q ABS rows, including each
// overlapping VEX/EVEX row and the original mask/broadcast attributes.
func X86PackedAbsRows(legacy X86LegacyCatalog, avx []X86AVXRule) ([]X86PackedBinaryRow, error) {
	return x86PackedBinaryRows(legacy, avx, `PABSB PABSW PABSD VPABSB VPABSW VPABSD VPABSQ`)
}

// X86PackedHorizontalIntegerRows retains the XMM and VEX horizontal add/sub
// rows. PHADDD's distinct MMX row keeps its independent region/EMMS owner.
func X86PackedHorizontalIntegerRows(legacy X86LegacyCatalog, avx []X86AVXRule) ([]X86PackedBinaryRow, error) {
	return x86PackedBinaryRows(legacy, avx, `
		PHADDW PHADDD PHADDSW PHSUBW PHSUBD PHSUBSW
		VPHADDW VPHADDD VPHADDSW VPHSUBW VPHSUBD VPHSUBSW`)
}

// X86PackedByteShuffleRows retains byte-indexed vector permutations. The
// separate VPSHUFBITQMB bit-to-mask family does not share this operand grammar.
func X86PackedByteShuffleRows(legacy X86LegacyCatalog, avx []X86AVXRule) ([]X86PackedBinaryRow, error) {
	return x86PackedBinaryRows(legacy, avx, `PSHUFB VPSHUFB`)
}

// X86PackedIntegerCompareRows retains both vector and mask results. MMX's
// six legacy rows retain their separate region and EMMS contract.
func X86PackedIntegerCompareRows(legacy X86LegacyCatalog, avx []X86AVXRule) ([]X86PackedBinaryRow, error) {
	return x86PackedBinaryRows(legacy, avx, `
		PCMPEQB PCMPEQW PCMPEQL PCMPEQQ PCMPGTB PCMPGTW PCMPGTL PCMPGTQ
		VPCMPEQB VPCMPEQW VPCMPEQD VPCMPEQQ VPCMPGTB VPCMPGTW VPCMPGTD VPCMPGTQ`)
}

// X86PackedMaskBlendRows retains the floating-point and integer mask-blend
// families. Only the PS/PD/D/Q members permit scalar-memory broadcast.
func X86PackedMaskBlendRows(legacy X86LegacyCatalog, avx []X86AVXRule) ([]X86PackedBinaryRow, error) {
	return x86PackedBinaryRows(legacy, avx, `VBLENDMPS VBLENDMPD VPBLENDMB VPBLENDMW VPBLENDMD VPBLENDMQ`)
}

// X86PackedBinaryPilotRows covers the shared packed add/subtract and D/Q
// low-multiply pilot. MMX and the separate W multiply family are not included.
// PADDD is a frontend alias, not an additional official encoder row.
func X86PackedBinaryPilotRows(legacy X86LegacyCatalog, avx []X86AVXRule) ([]X86PackedBinaryRow, error) {
	return x86PackedBinaryRows(legacy, avx, `
		PADDB PADDW PADDL PADDQ PADDSB PADDSW PADDUSB PADDUSW
		VPADDB VPADDW VPADDD VPADDQ VPADDSB VPADDSW VPADDUSB VPADDUSW
		PSUBB PSUBW PSUBL PSUBQ PSUBSB PSUBSW PSUBUSB PSUBUSW
		VPSUBB VPSUBW VPSUBD VPSUBQ VPSUBSB VPSUBSW VPSUBUSB VPSUBUSW
		PMULLD VPMULLD VPMULLQ`)
}

func x86PackedBinaryRows(legacy X86LegacyCatalog, avx []X86AVXRule, names string) ([]X86PackedBinaryRow, error) {
	// Scope names are independent of the translator's registry. All operand,
	// width, suffix and encoding axes still come from the original Go rows.
	selected := map[string]bool{}
	for _, name := range strings.Fields(names) {
		selected[name] = false
	}
	var rows []X86PackedBinaryRow
	for _, entry := range legacy.Optab {
		if _, ok := selected[entry.Opcode]; !ok {
			continue
		}
		for _, operand := range entry.Rows {
			if operand.RowIndex == 0 && strings.Join(operand.Classes, ",") == "Ymm,Ymr" && operand.UnmatchedOffset == 0 &&
				(entry.Ytab == "ymm" && operand.ZCase == "Zm_r_xm" && operand.ZOffset == 1 ||
					entry.Ytab == "ymmxmm0f38" && entry.Prefix == "Px" && entry.EscapeAdjustment == 1 &&
						operand.ZCase == "Zlitm_r" && operand.ZOffset == 3) {
				continue // MMX regions have a separate owner and EMMS contract.
			}
			witnesses, err := x86PackedLegacyRowWitnesses(entry, operand)
			if err != nil {
				return nil, fmt.Errorf("%s/%d: %w", entry.Opcode, operand.RowIndex, err)
			}
			rows = append(rows, X86PackedBinaryRow{
				ID:     fmt.Sprintf("optab/%d/%s/%s/%d", entry.TableIndex, entry.Opcode, entry.Ytab, operand.RowIndex),
				Opcode: entry.Opcode, Witnesses: witnesses,
			})
			selected[entry.Opcode] = true
		}
	}
	for index, rule := range avx {
		if _, ok := selected[rule.Opcode]; !ok {
			continue
		}
		witnesses, err := x86PackedAVXRowWitnesses(rule)
		if err != nil {
			return nil, fmt.Errorf("%s/%d: %w", rule.Opcode, rule.RowIndex, err)
		}
		rows = append(rows, X86PackedBinaryRow{
			ID:     fmt.Sprintf("avx/%d/%s/%s/%d", index, rule.Opcode, rule.Ytab, rule.RowIndex),
			Opcode: rule.Opcode, Witnesses: witnesses,
		})
		selected[rule.Opcode] = true
	}
	for name, found := range selected {
		if !found {
			return nil, fmt.Errorf("official packed opcode %s has no retained row", name)
		}
	}
	return rows, nil
}

func x86PackedAVXRowWitnesses(rule X86AVXRule) ([]X86PackedBinaryWitness, error) {
	unary := rule.Ytab == "_yvmovddup" || rule.Ytab == "_yvexpandpd"
	syntax, err := x86PackedSyntaxWitnesses(rule, unary)
	if err != nil {
		return nil, err
	}
	masked := rule.ZCase == "Zevex_rm_v_k_r" || rule.ZCase == "Zevex_rm_k_r"
	kind, stride, width := "vex", 2, 128
	if strings.HasPrefix(rule.ZCase, "Zevex_") {
		kind, stride = "evex", 3
	}
	// The validated official source class owns vector width, independently
	// of register/memory witnesses and vector versus K result classes.
	bank := strings.ToUpper(rule.Classes[0][1:2])
	if bank == "Y" {
		width = 256
	} else if bank == "Z" {
		width = 512
	}
	wantClass := map[string]string{"X": "Yxm", "Y": "Yym", "Z": "Yzm"}[bank]
	if kind == "evex" && bank != "Z" {
		wantClass += "Evex"
	}
	wantOffset := stride
	if kind == "evex" && !masked {
		wantOffset = 0
	}
	if rule.Classes[0] != wantClass || rule.ZOffset != wantOffset {
		return nil, fmt.Errorf("generation/class/row stride mismatch")
	}
	header := rule.OpBytes[rule.EncodingOffset]
	atoms, err := x86PackedEncodingAtoms(header, []string{
		"avxEscape", kind + strconv.Itoa(width),
	}, []string{kind + "66", kind + "0F", kind + "0F38", kind + "W0", kind + "W1"})
	if err != nil || atoms[kind+"0F"] == atoms[kind+"0F38"] || atoms[kind+"W0"] == atoms[kind+"W1"] {
		return nil, fmt.Errorf("unmodeled packed encoding header %q", header)
	}
	mapByte, wBit, length, pp := byte(1), byte(0), byte(0), byte(0)
	if atoms[kind+"66"] {
		pp = 1
	}
	if atoms[kind+"0F38"] {
		mapByte = 2
	}
	if atoms[kind+"W1"] {
		wBit = 0x80
	}
	if width == 256 {
		length = 1
	} else if width == 512 {
		length = 2
	}
	if kind == "evex" {
		features := rule.OpBytes[rule.EncodingOffset+1]
		flags, err := x86PackedEncodingAtoms(features, []string{"evexN" + strconv.Itoa(width/8)},
			[]string{"evexZeroingEnabled", "evexBcstN4", "evexBcstN8"})
		if err != nil || flags["evexBcstN4"] && flags["evexBcstN8"] {
			return nil, fmt.Errorf("unmodeled packed encoding features %q", features)
		}
	}
	opcode, err := strconv.ParseUint(rule.OpBytes[rule.EncodingOffset+stride-1], 0, 8)
	if err != nil {
		return nil, fmt.Errorf("packed opcode is not a byte literal")
	}
	var witnesses []X86PackedBinaryWitness
	for _, instruction := range syntax {
		if kind == "evex" {
			// High banks force EVEX even for unmasked X/Y forms. Low-bank
			// syntax alone cannot identify overlapping VEX/EVEX rows.
			instruction = strings.NewReplacer(bank+"0", bank+"16", bank+"1", bank+"17", bank+"2", bank+"18").Replace(instruction)
		}
		fields := strings.Fields(instruction)
		operands := strings.Split(fields[1], ",")
		second, vvvv, vprime := operands[1], byte(0x70), byte(0)
		if unary {
			// Unary encodings reserve VEX.vvvv and EVEX.V' as ones;
			// the mask and merging destination remain separate inputs.
			second, vvvv, vprime = "", 0x78, 0x08
		}
		witness := X86PackedBinaryWitness{
			Instruction: instruction, Generation: kind, Width: width,
			First: operands[0], Second: second, Destination: operands[len(operands)-1],
			Memory: operands[0] == "(AX)", Masked: masked,
			Zeroing: strings.HasSuffix(fields[0], ".Z"), Broadcast: strings.Contains(fields[0], ".BCST"),
			MaskResult: rule.Classes[len(rule.Classes)-1] == "Yk",
		}
		modrm := byte(0xd0) // destination 2/18, r/m register 0/16.
		if witness.Memory {
			modrm = 0x10 // the only memory witness is the undisturbed (AX).
		}
		if kind == "evex" {
			p0, p2 := byte(0xa0)|mapByte, length<<5|vprime
			if witness.Memory {
				p0 = 0xe0 | mapByte
			}
			if witness.MaskResult {
				p0 |= 0x10 // K2 has no destination bit 4, unlike vector 18.
			}
			if masked {
				p2 |= 1 // K1; binary V' is zero for the second source 17.
			}
			if witness.Zeroing {
				p2 |= 0x80
			}
			if witness.Broadcast {
				p2 |= 0x10
			}
			witness.Encoding = []byte{0x62, p0, wBit | vvvv | 0x04 | pp, p2, byte(opcode), modrm}
		} else if mapByte == 1 && wBit == 0 {
			witness.Encoding = []byte{0xc5, 0x80 | vvvv | pp | length<<2, byte(opcode), modrm}
		} else {
			witness.Encoding = []byte{0xc4, 0xe0 | mapByte, wBit | vvvv | pp | length<<2, byte(opcode), modrm}
		}
		witnesses = append(witnesses, witness)
	}
	return witnesses, nil
}

func x86PackedEncodingAtoms(expression string, required, optional []string) (map[string]bool, error) {
	allowed, found := map[string]bool{}, map[string]bool{}
	for _, term := range append(append([]string(nil), required...), optional...) {
		allowed[term] = true
	}
	for _, term := range strings.FieldsFunc(expression, func(r rune) bool {
		return r == '|' || r == '(' || r == ')' || r == ' ' || r == '\t' || r == '\n'
	}) {
		if !allowed[term] || found[term] {
			return nil, fmt.Errorf("unrecognized or repeated encoding term %s", term)
		}
		found[term] = true
	}
	for _, term := range required {
		if !found[term] {
			return nil, fmt.Errorf("missing encoding term %s", term)
		}
	}
	return found, nil
}

func x86PackedLegacyRowWitnesses(entry X86LegacyOpcode, row X86LegacyOperand) ([]X86PackedBinaryWitness, error) {
	literal := entry.Ytab == "ymmxmm0f38" && entry.Prefix == "Px" && entry.EscapeAdjustment == 1 &&
		row.RowIndex == 1 && row.ZCase == "Zlitm_r" && row.ZOffset == 5 && row.UnmatchedOffset == 4
	if entry.Kind != "encoding" || entry.EscapeAdjustment != 0 && !literal || strings.Join(row.Classes, ",") != "Yxm,Yxr" {
		return nil, fmt.Errorf("unmodeled legacy packed operand row")
	}
	prefix, offset, opcodeBytes := []byte{0x66, 0x0f}, 0, 1
	terminator := 0
	switch {
	case literal:
		// Zlitm_r emits the selected zero-terminated sequence verbatim,
		// including its mandatory prefix and opcode-map bytes.
		prefix, offset, opcodeBytes, terminator = nil, row.UnmatchedOffset, row.ZOffset-1, 1
		if len(entry.OpBytes) != offset+opcodeBytes+terminator ||
			entry.OpBytes[offset-1] != "0" || entry.OpBytes[len(entry.OpBytes)-1] != "0" {
			return nil, fmt.Errorf("unmodeled legacy packed literal boundaries")
		}
	case entry.Ytab == "ymm" && entry.Prefix == "Py1" && row.RowIndex == 1 && row.ZCase == "Zm_r_xm" && row.ZOffset == 2 && row.UnmatchedOffset == 1:
		if len(entry.OpBytes) != 3 || entry.OpBytes[1] != "Pe" || entry.OpBytes[0] != entry.OpBytes[2] {
			return nil, fmt.Errorf("unmodeled Py1 XMM opcode payload")
		}
		offset = 2 // Py1 consumes the Pe at offset 1 before the opcode.
	case entry.Ytab == "yxm" && (entry.Prefix == "Pe" || entry.Prefix == "Pq" || entry.Prefix == "Pm") &&
		row.RowIndex == 0 && row.ZCase == "Zm_r_xm" && row.ZOffset == 1 && row.UnmatchedOffset == 0:
		if entry.Prefix == "Pm" {
			prefix = []byte{0x0f}
		}
	case entry.Ytab == "yxm_q4" && entry.Prefix == "Pq4" && row.RowIndex == 0 && row.ZCase == "Zm_r" && row.ZOffset == 1 && row.UnmatchedOffset == 0:
		prefix = append(prefix, 0x38)
	case entry.Ytab == "ymshufb" && entry.Prefix == "Pq" && row.RowIndex == 0 && row.ZCase == "Zm2_r" && row.ZOffset == 2 && row.UnmatchedOffset == 0:
		// Go's Zm2_r emits two selected opcode bytes before ModRM. They
		// are payload, not another mandatory prefix or a mnemonic lookup.
		opcodeBytes = 2
	default:
		return nil, fmt.Errorf("unmodeled legacy packed prefix/offset rule")
	}
	if len(entry.OpBytes) != offset+opcodeBytes+terminator {
		return nil, fmt.Errorf("unmodeled legacy packed opcode payload")
	}
	encoding := prefix
	for _, value := range entry.OpBytes[offset : offset+opcodeBytes] {
		opcode, err := strconv.ParseUint(value, 0, 8)
		if err != nil || literal && opcode == 0 {
			return nil, fmt.Errorf("legacy packed opcode is not a byte literal")
		}
		encoding = append(encoding, byte(opcode))
	}
	var witnesses []X86PackedBinaryWitness
	for _, first := range []string{"X0", "(AX)"} {
		modrm := byte(0xc8)
		if first == "(AX)" {
			modrm = 0x08
		}
		witnesses = append(witnesses, X86PackedBinaryWitness{
			Instruction: entry.Opcode + " " + first + ",X1", Generation: "legacy", Width: 128,
			First: first, Second: "X1", Destination: "X1", Memory: first == "(AX)",
			Encoding: append(append([]byte(nil), encoding...), modrm),
		})
	}
	return witnesses, nil
}
