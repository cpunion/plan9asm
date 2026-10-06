package goasmrules

import (
	"fmt"
	"strings"
)

// X86PackedBinarySyntaxWitnesses expands the complete _yvandnpd grammar used
// by packed binary families. These are source-syntax candidates: low-bank X/Y
// spellings can select VEX even when an EVEX table row also matches. Actual Go
// objects and semantic oracles must independently establish encoding identity.
// Frontend architecture restrictions are likewise checked by the actual Go
// assembler, never silently pruned here.
func X86PackedBinarySyntaxWitnesses(rule X86AVXRule) ([]string, error) {
	if rule.Ytab != "_yvandnpd" || rule.Prefix != "Pavx" {
		return nil, fmt.Errorf("%s is outside the audited packed binary table grammar", rule.Opcode)
	}
	masked := rule.ZCase == "Zevex_rm_v_k_r"
	evex := masked || rule.ZCase == "Zevex_rm_v_r"
	if !evex && rule.ZCase != "Zvex_rm_v_r" {
		return nil, fmt.Errorf("%s has an unmodeled packed binary zcase %s", rule.Opcode, rule.ZCase)
	}
	count := 3
	if masked {
		count = 4
	}
	if len(rule.Classes) != count {
		return nil, fmt.Errorf("%s packed binary arity differs from its zcase", rule.Opcode)
	}
	bank := ""
	for _, candidate := range []struct {
		memory, register, bank string
	}{
		{"Yxm", "Yxr", "X"},
		{"Yym", "Yyr", "Y"},
		{"YxmEvex", "YxrEvex", "X"},
		{"YymEvex", "YyrEvex", "Y"},
		{"Yzm", "Yzr", "Z"},
	} {
		if rule.Classes[0] == candidate.memory && rule.Classes[1] == candidate.register &&
			rule.Classes[count-1] == candidate.register {
			bank = candidate.bank
			break
		}
	}
	if bank == "" || masked && rule.Classes[2] != "Yknot0" {
		return nil, fmt.Errorf("%s has unmodeled packed binary operand classes %v", rule.Opcode, rule.Classes)
	}
	stride := 2
	if evex {
		stride = 3
	}
	if rule.EncodingOffset < 0 || rule.EncodingOffset+stride > len(rule.OpBytes) {
		return nil, fmt.Errorf("%s encoding offset %d is outside its complete opBytes", rule.Opcode, rule.EncodingOffset)
	}
	encoding := rule.OpBytes[rule.EncodingOffset]
	width := map[string]string{"X": "128", "Y": "256", "Z": "512"}[bank]
	kind := "vex"
	if evex {
		kind = "evex"
	}
	if !hasEncodingTerm(encoding, kind+width) {
		return nil, fmt.Errorf("%s operand width %s differs from encoding %s", rule.Opcode, bank, encoding)
	}
	features := ""
	if evex {
		features = rule.OpBytes[rule.EncodingOffset+1]
	}
	if hasEncodingTerm(features, "evexRoundingEnabled") || hasEncodingTerm(features, "evexSaeEnabled") {
		return nil, fmt.Errorf("%s adds unaudited rounding/SAE to the packed binary grammar", rule.Opcode)
	}
	var witnesses []string
	emit := func(source, suffix string) {
		operands := []string{source, bank + "1"}
		if masked {
			operands = append(operands, "K1")
		}
		operands = append(operands, bank+"2")
		witnesses = append(witnesses, rule.Opcode+suffix+" "+strings.Join(operands, ","))
	}
	for _, source := range []string{bank + "0", "(AX)"} {
		emit(source, "")
		if masked && hasEncodingTerm(features, "evexZeroingEnabled") {
			emit(source, ".Z")
		}
	}
	if hasEncodingTerm(features, "evexBcstN4") || hasEncodingTerm(features, "evexBcstN8") {
		emit("(AX)", ".BCST")
		if masked && hasEncodingTerm(features, "evexZeroingEnabled") {
			emit("(AX)", ".BCST.Z")
		}
	}
	return witnesses, nil
}

func hasEncodingTerm(expression, term string) bool {
	for _, candidate := range strings.FieldsFunc(expression, func(r rune) bool {
		return r == '|' || r == '(' || r == ')' || r == ' ' || r == '\t' || r == '\n'
	}) {
		if candidate == term {
			return true
		}
	}
	return false
}
