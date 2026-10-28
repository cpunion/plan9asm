package goasmrules

import (
	"fmt"
	"strconv"
	"strings"
)

// Physical immediate/source/result roles do not combine the four families'
// numerical, old-destination or native-access contracts.
type x86LeadingImmediateVectorFormat struct {
	width             int
	masked, broadcast bool
	encoding          x86PackedAVXOpcode
}

func x86LeadingImmediateVectorFormatForRule(rule X86AVXRule) (x86LeadingImmediateVectorFormat, error) {
	f := x86LeadingImmediateVectorFormat{masked: rule.ZCase == "Zevex_i_rm_v_k_r"}
	fail := func() (x86LeadingImmediateVectorFormat, error) {
		return f, fmt.Errorf("leading-immediate two-vector result roles differ from their official format")
	}
	count, maskIndex := 4, 0
	if f.masked {
		count, maskIndex = 5, 1
	}
	if rule.Ytab != "_yvalignd" || rule.Prefix != "Pavx" ||
		rule.ZCase != "Zevex_i_rm_v_r" && !f.masked || len(rule.Classes) != count ||
		rule.Classes[0] != "Yu8" || f.masked && rule.Classes[3] != "Yknot0" {
		return fail()
	}
	f.width = map[string]int{"YxmEvex": 128, "YymEvex": 256, "Yzm": 512}[rule.Classes[1]]
	register := map[int]string{128: "YxrEvex", 256: "YyrEvex", 512: "Yzr"}[f.width]
	widthIndex := map[int]int{128: 0, 256: 1, 512: 2}[f.width]
	if f.width == 0 || rule.Classes[2] != register || rule.Classes[count-1] != register ||
		rule.RowIndex != widthIndex*2+maskIndex || rule.ZOffset != maskIndex*3 ||
		rule.EncodingOffset != widthIndex*3 || len(rule.OpBytes) != 9 {
		return fail()
	}
	offset := rule.EncodingOffset
	w := "evexW0"
	if strings.HasSuffix(rule.Opcode, "Q") || strings.HasSuffix(rule.Opcode, "DW") {
		w = "evexW1"
	}
	if _, err := x86PackedEncodingAtoms(rule.OpBytes[offset], []string{
		"avxEscape", "evex" + strconv.Itoa(f.width), "evex66", "evex0F3A", w,
	}, nil); err != nil {
		return fail()
	}
	f.broadcast = strings.HasSuffix(rule.Opcode, "D") || strings.HasSuffix(rule.Opcode, "Q")
	features := []string{"evexN" + strconv.Itoa(f.width/8), "evexZeroingEnabled"}
	if f.broadcast {
		bytes := 4
		if w == "evexW1" {
			bytes = 8
		}
		features = append(features, "evexBcstN"+strconv.Itoa(bytes))
	}
	if _, err := x86PackedEncodingAtoms(rule.OpBytes[offset+1], features, nil); err != nil {
		return fail()
	}
	opcode, err := strconv.ParseUint(rule.OpBytes[offset+2], 0, 8)
	if err != nil {
		return fail()
	}
	f.encoding = x86PackedAVXOpcode{kind: "evex", mapByte: 3, pp: 1, length: byte(widthIndex), opcode: byte(opcode)}
	if w == "evexW1" {
		f.encoding.wBit = 0x80
	}
	return f, nil
}

func x86LeadingImmediateVectorRowWitnesses(rule X86AVXRule) ([]X86PackedBinaryWitness, error) {
	f, err := x86LeadingImmediateVectorFormatForRule(rule)
	if err != nil {
		return nil, err
	}
	bank := map[int]string{128: "X", 256: "Y", 512: "Z"}[f.width]
	var witnesses []X86PackedBinaryWitness
	for _, source := range []string{bank + "16", "(AX)"} {
		for _, broadcast := range []bool{false, true} {
			if broadcast && (source != "(AX)" || !f.broadcast) {
				continue
			}
			for _, zeroing := range []bool{false, true} {
				if zeroing && !f.masked {
					continue
				}
				operands, suffix := "$7,"+source+","+bank+"17", ""
				p0, p2, modrm := byte(0xa0), f.encoding.length<<5, byte(0xd0)
				if source == "(AX)" {
					p0, modrm = 0xe0, 0x10
				}
				if f.masked {
					operands, p2 = operands+",K1", p2|1
				}
				if broadcast {
					suffix, p2 = ".BCST", p2|0x10
				}
				if zeroing {
					suffix, p2 = suffix+".Z", p2|0x80
				}
				witnesses = append(witnesses, X86PackedBinaryWitness{
					Instruction: rule.Opcode + suffix + " " + operands + "," + bank + "18", Generation: "evex",
					First: source, Second: bank + "17", Destination: bank + "18", Width: f.width,
					Memory: source == "(AX)", Masked: f.masked, Zeroing: zeroing, Broadcast: broadcast,
					Encoding: append(f.encoding.bytes(0x70, p0, p2, modrm), 7),
				})
			}
		}
	}
	return witnesses, nil
}
