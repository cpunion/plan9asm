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
	if rule.Ytab != "_yvandnpd" {
		return nil, fmt.Errorf("%s is outside the audited packed binary table grammar", rule.Opcode)
	}
	return x86PackedBinarySyntaxWitnesses(rule)
}

func x86PackedBinarySyntaxWitnesses(rule X86AVXRule) ([]string, error) {
	return x86PackedSyntaxWitnesses(rule, false)
}

func x86PackedMoveTable(table string) bool {
	return table == "_yvmovdqa" || table == "_yvmovapd" || table == "_yvmovdqa32"
}

func x86PackedMoveStoreRow(rule X86AVXRule) bool {
	return x86PackedMoveTable(rule.Ytab) && (rule.ZCase == "Zvex_r_v_rm" ||
		rule.ZCase == "Zevex_r_v_rm" || rule.ZCase == "Zevex_r_k_rm")
}

func x86PackedSyntaxWitnesses(rule X86AVXRule, unary bool) ([]string, error) {
	move, store := x86PackedMoveTable(rule.Ytab), x86PackedMoveStoreRow(rule)
	floating := rule.Ytab == "_yvaddpd" || rule.Ytab == "_yvaddsd"
	scalar := rule.Ytab == "_yvaddsd"
	binaryTable := rule.Ytab == "_yvandnpd" || rule.Ytab == "_yvblendmpd" ||
		rule.Ytab == "_yvaddsubpd" || rule.Ytab == "_yvpcmpeqb" || floating
	unaryTable := rule.Ytab == "_yvmovddup" || rule.Ytab == "_yvexpandpd" || move
	if (!unary && !binaryTable || unary && !unaryTable) || rule.Prefix != "Pavx" {
		return nil, fmt.Errorf("%s is outside the audited packed vector table grammar", rule.Opcode)
	}
	maskCase := "Zevex_rm_v_k_r"
	if unary {
		maskCase = "Zevex_rm_k_r"
	}
	if store {
		maskCase = "Zevex_r_k_rm"
	}
	masked := rule.ZCase == maskCase
	evex := masked || rule.ZCase == "Zevex_rm_v_r" || store && rule.ZCase == "Zevex_r_v_rm"
	maskResult := rule.Ytab == "_yvpcmpeqb" && evex
	if (rule.Ytab == "_yvaddsubpd" || rule.Ytab == "_yvmovdqa") && evex ||
		(rule.Ytab == "_yvexpandpd" || rule.Ytab == "_yvmovdqa32") && !evex {
		return nil, fmt.Errorf("%s encoding generation differs from its packed table", rule.Opcode)
	}
	if !evex && rule.ZCase != "Zvex_rm_v_r" && !(store && rule.ZCase == "Zvex_r_v_rm") {
		return nil, fmt.Errorf("%s has an unmodeled packed vector zcase %s", rule.Opcode, rule.ZCase)
	}
	count := 3
	if masked {
		count = 4
	}
	if unary {
		count--
	}
	if len(rule.Classes) != count {
		return nil, fmt.Errorf("%s packed vector arity differs from its zcase", rule.Opcode)
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
		resultClass := candidate.register
		if maskResult {
			resultClass = "Yk"
		}
		firstClass := candidate.memory
		if store {
			firstClass, resultClass = candidate.register, candidate.memory
		}
		if rule.Classes[0] == firstClass && (unary || rule.Classes[1] == candidate.register) &&
			rule.Classes[count-1] == resultClass {
			bank = candidate.bank
			break
		}
	}
	if bank == "" || !evex && bank == "Z" || scalar && bank != "X" || masked && rule.Classes[count-2] != "Yknot0" {
		return nil, fmt.Errorf("%s has unmodeled packed vector operand classes %v", rule.Opcode, rule.Classes)
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
	rounding := hasEncodingTerm(features, "evexRoundingEnabled")
	if hasEncodingTerm(features, "evexSaeEnabled") || rounding && (!floating || !scalar && bank != "Z") {
		return nil, fmt.Errorf("%s adds unaudited rounding/SAE to the packed binary grammar", rule.Opcode)
	}
	if maskResult && hasEncodingTerm(features, "evexZeroingEnabled") {
		return nil, fmt.Errorf("%s adds vector zeroing to a mask result", rule.Opcode)
	}
	destination := bank + "2"
	if maskResult {
		destination = "K2"
	}
	var witnesses []string
	emit := func(source, suffix string) {
		last := destination
		if store {
			source, last = bank+"0", source
		}
		operands := []string{source}
		if !unary {
			operands = append(operands, bank+"1")
		}
		if masked {
			operands = append(operands, "K1")
		}
		operands = append(operands, last)
		witnesses = append(witnesses, rule.Opcode+suffix+" "+strings.Join(operands, ","))
	}
	sources := []string{bank + "0", "(AX)"}
	if store {
		sources[0] = destination
	}
	for _, source := range sources {
		if move && !store && source != "(AX)" {
			continue // Register alternatives select the earlier store row.
		}
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
	if rounding {
		for _, control := range []string{"RN_SAE", "RD_SAE", "RU_SAE", "RZ_SAE"} {
			emit(bank+"0", "."+control)
			if masked && hasEncodingTerm(features, "evexZeroingEnabled") {
				emit(bank+"0", "."+control+".Z")
			}
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
