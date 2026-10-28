package goasmrules

import (
	"fmt"
	"strconv"
	"strings"
)

// Both families have byte-vector results; only affine has a leading imm8
// and m64 matrix broadcast. W1 never changes the byte mask's granularity.
type x86GFNIFormat struct {
	width                        int
	immediate, masked, broadcast bool
	encoding                     x86PackedAVXOpcode
}

func x86GFNIFormatForRule(rule X86AVXRule) (x86GFNIFormat, error) {
	f := x86GFNIFormat{immediate: rule.Ytab == "_yvgf2p8affineinvqb"}
	fail := func() (x86GFNIFormat, error) {
		return f, fmt.Errorf("GFNI data/matrix/tuple/control differs from its official format")
	}
	first, plain, masked, opcodeMap, mapByte, w := 0, "Zvex_rm_v_r", "Zevex_rm_v_k_r", "0F38", byte(2), "W0"
	if f.immediate {
		first, plain, masked, opcodeMap, mapByte, w = 1, "Zvex_i_rm_v_r", "Zevex_i_rm_v_k_r", "0F3A", 3, "W1"
		member, valid := strings.CutPrefix(rule.Opcode, "VGF2P8AFFINE")
		if !valid || member != "QB" && member != "INVQB" {
			return fail()
		}
	} else if rule.Opcode != "VGF2P8MULB" || rule.Ytab != "_yvandnpd" {
		return fail()
	}
	f.masked = rule.ZCase == masked
	evex := strings.HasPrefix(rule.ZCase, "Zevex_")
	count := first + 3
	if f.masked {
		count++
	}
	if evex {
		plain = strings.Replace(plain, "Zvex_", "Zevex_", 1)
	}
	if rule.Prefix != "Pavx" || rule.ZCase != plain && !f.masked || len(rule.Classes) != count ||
		f.immediate && rule.Classes[0] != "Yu8" || f.masked && rule.Classes[first+2] != "Yknot0" {
		return fail()
	}
	f.width = map[string]int{"Yxm": 128, "Yym": 256}[rule.Classes[first]]
	register := map[int]string{128: "Yxr", 256: "Yyr"}[f.width]
	if evex {
		f.width = map[string]int{"YxmEvex": 128, "YymEvex": 256, "Yzm": 512}[rule.Classes[first]]
		register = map[int]string{128: "YxrEvex", 256: "YyrEvex", 512: "Yzr"}[f.width]
	}
	if f.width == 0 || rule.Classes[first+1] != register || rule.Classes[count-1] != register || len(rule.OpBytes) != 13 {
		return fail()
	}
	widthIndex := map[int]int{128: 0, 256: 1, 512: 2}[f.width]
	index, offset, zoffset, kind, stride := widthIndex, widthIndex*2, 2, "vex", 2
	if evex {
		index, offset, zoffset, kind, stride = 2+widthIndex*2, 4+widthIndex*3, 0, "evex", 3
		if f.masked {
			index, zoffset = index+1, 3
		}
	}
	if rule.RowIndex != index || rule.EncodingOffset != offset || rule.ZOffset != zoffset {
		return fail()
	}
	if _, err := x86PackedEncodingAtoms(rule.OpBytes[offset], []string{
		"avxEscape", kind + strconv.Itoa(f.width), kind + "66", kind + opcodeMap, kind + w,
	}, nil); err != nil {
		return fail()
	}
	if evex {
		required := []string{"evexN" + strconv.Itoa(f.width/8), "evexZeroingEnabled"}
		if f.immediate {
			required = append(required, "evexBcstN8")
			f.broadcast = true
		}
		if _, err := x86PackedEncodingAtoms(rule.OpBytes[offset+1], required, nil); err != nil {
			return fail()
		}
	}
	opcode, err := strconv.ParseUint(rule.OpBytes[offset+stride-1], 0, 8)
	if err != nil {
		return fail()
	}
	f.encoding = x86PackedAVXOpcode{kind: kind, mapByte: mapByte, pp: 1, length: byte(widthIndex), opcode: byte(opcode)}
	if f.immediate {
		f.encoding.wBit = 0x80
	}
	return f, nil
}

func x86GFNIRowWitnesses(rule X86AVXRule) ([]X86PackedBinaryWitness, error) {
	f, err := x86GFNIFormatForRule(rule)
	if err != nil {
		return nil, err
	}
	bank := map[int]string{128: "X", 256: "Y", 512: "Z"}[f.width]
	first, second, destination := bank+"0", bank+"1", bank+"2"
	if f.encoding.kind == "evex" {
		first, second, destination = bank+"16", bank+"17", bank+"18"
	}
	controls := []string{""}
	if f.masked {
		controls = append(controls, ".Z")
	}
	var witnesses []X86PackedBinaryWitness
	add := func(source, suffix string) {
		operands, p0, p2, modrm := source+","+second, byte(0xa0), f.encoding.length<<5, byte(0xd0)
		if f.immediate {
			operands = "$7," + operands
		}
		if source == "(AX)" {
			p0, modrm = 0xe0, 0x10
		}
		if f.masked {
			operands, p2 = operands+",K1", p2|1
		}
		zeroing, broadcast := strings.HasSuffix(suffix, ".Z"), strings.Contains(suffix, ".BCST")
		if zeroing {
			p2 |= 0x80
		}
		if broadcast {
			p2 |= 0x10
		}
		encoding := f.encoding.bytes(0x70, p0, p2, modrm)
		if f.immediate {
			encoding = append(encoding, 7)
		}
		witnesses = append(witnesses, X86PackedBinaryWitness{
			Instruction: rule.Opcode + suffix + " " + operands + "," + destination, Generation: f.encoding.kind,
			First: source, Second: second, Destination: destination, Width: f.width,
			Memory: source == "(AX)", Masked: f.masked, Zeroing: zeroing, Broadcast: broadcast, Encoding: encoding,
		})
	}
	for _, source := range []string{first, "(AX)"} {
		for _, suffix := range controls {
			add(source, suffix)
		}
	}
	if f.broadcast {
		for _, suffix := range controls {
			add("(AX)", ".BCST"+suffix)
		}
	}
	return witnesses, nil
}
