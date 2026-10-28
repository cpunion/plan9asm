package goasmrules

import (
	"fmt"
	"strconv"
	"strings"
)

// Vector input width/tuple and K result/writemask are independent roles.
// Matching this physical format grants no input, memory or numeric authority.
type x86ImmediateIntegerCompareFormat struct {
	width, broadcast int
	masked           bool
	encoding         x86PackedAVXOpcode
}

func x86ImmediateIntegerCompareFormatForRule(rule X86AVXRule) (x86ImmediateIntegerCompareFormat, error) {
	f := x86ImmediateIntegerCompareFormat{masked: rule.ZCase == "Zevex_i_rm_v_k_r"}
	fail := func() (x86ImmediateIntegerCompareFormat, error) {
		return f, fmt.Errorf("immediate integer compare input/tuple/K-result differs from its official format")
	}
	count := 4
	if f.masked {
		count++
	}
	if rule.Ytab != "_yvpcmpb" || rule.Prefix != "Pavx" || rule.ZCase != "Zevex_i_rm_v_r" && !f.masked ||
		len(rule.Classes) != count || rule.Classes[0] != "Yu8" || rule.Classes[count-1] != "Yk" ||
		f.masked && rule.Classes[3] != "Yknot0" {
		return fail()
	}
	f.width = map[string]int{"YxmEvex": 128, "YymEvex": 256, "Yzm": 512}[rule.Classes[1]]
	second := map[int]string{128: "YxrEvex", 256: "YyrEvex", 512: "Yzr"}[f.width]
	member, valid := strings.CutPrefix(rule.Opcode, "VPCMP")
	member = strings.TrimPrefix(member, "U")
	if !valid || len(member) != 1 || f.width == 0 || rule.Classes[2] != second {
		return fail()
	}
	bits := map[byte]int{'B': 8, 'W': 16, 'D': 32, 'Q': 64}[member[0]]
	if bits == 0 {
		return fail()
	}
	widthIndex := map[int]int{128: 0, 256: 1, 512: 2}[f.width]
	maskIndex := 0
	if f.masked {
		maskIndex = 1
	}
	if rule.RowIndex != widthIndex*2+maskIndex || rule.ZOffset != maskIndex*3 ||
		rule.EncodingOffset != widthIndex*3 || len(rule.OpBytes) != 9 {
		return fail()
	}
	w := "W0"
	if bits == 16 || bits == 64 {
		w = "W1"
	}
	offset := rule.EncodingOffset
	if _, err := x86PackedEncodingAtoms(rule.OpBytes[offset], []string{
		"avxEscape", "evex" + strconv.Itoa(f.width), "evex66", "evex0F3A", "evex" + w,
	}, nil); err != nil {
		return fail()
	}
	required := []string{"evexN" + strconv.Itoa(f.width/8)}
	if bits >= 32 {
		f.broadcast = bits / 8
		required = append(required, "evexBcstN"+strconv.Itoa(f.broadcast))
	}
	if _, err := x86PackedEncodingAtoms(rule.OpBytes[offset+1], required, nil); err != nil {
		return fail()
	}
	opcode, err := strconv.ParseUint(rule.OpBytes[offset+2], 0, 8)
	if err != nil {
		return fail()
	}
	f.encoding = x86PackedAVXOpcode{kind: "evex", mapByte: 3, pp: 1, length: byte(widthIndex), opcode: byte(opcode)}
	if w == "W1" {
		f.encoding.wBit = 0x80
	}
	return f, nil
}

func x86ImmediateIntegerCompareRowWitnesses(rule X86AVXRule) ([]X86PackedBinaryWitness, error) {
	f, err := x86ImmediateIntegerCompareFormatForRule(rule)
	if err != nil {
		return nil, err
	}
	bank := map[int]string{128: "X", 256: "Y", 512: "Z"}[f.width]
	var witnesses []X86PackedBinaryWitness
	add := func(first string, destination byte, broadcast bool) {
		result, suffix := "K"+strconv.Itoa(int(destination)), ""
		operands, p0, p2, modrm := "$7,"+first+","+bank+"17", byte(0xb0), f.encoding.length<<5, byte(0xc0)|destination<<3
		if first == "(AX)" {
			p0, modrm = 0xf0, destination<<3
		}
		if f.masked {
			operands, p2 = operands+",K1", p2|1
		}
		if broadcast {
			suffix, p2 = ".BCST", p2|0x10
		}
		witnesses = append(witnesses, X86PackedBinaryWitness{
			Instruction: rule.Opcode + suffix + " " + operands + "," + result, Generation: "evex",
			First: first, Second: bank + "17", Destination: result, Width: f.width,
			Memory: first == "(AX)", Masked: f.masked, Broadcast: broadcast, MaskResult: true,
			Encoding: append(f.encoding.bytes(0x70, p0, p2, modrm), 7),
		})
	}
	// K0 is legal data output, independently of the nonzero-only writemask.
	for _, destination := range []byte{2, 0} {
		add(bank+"16", destination, false)
		add("(AX)", destination, false)
		if f.broadcast != 0 {
			add("(AX)", destination, true)
		}
	}
	return witnesses, nil
}
