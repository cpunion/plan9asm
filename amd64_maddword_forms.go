package plan9asm

import "strings"

// Go calls Intel's legacy PMADDWD instruction PMADDWL. The raw decoder keeps
// that canonical spelling. ymm supplies MMX/m64 and X/m128; _yvandnpd supplies
// all VEX/EVEX widths. Each VPMADDWD row permits zeroing, never broadcasting.
var x86WordMultiplyAddSpecs = map[Op]x86PackedBinarySpec{
	"PMADDWL":  {encoding: x86PackedLogicalLegacyInteger},
	"VPMADDWD": {encoding: x86PackedLogicalVEXOrEVEX, zeroing: true},
}

func x86WordMultiplyAddProperties(ins Instr, goarch string) (x86PackedBinaryForm, bool, error) {
	base, _, _ := strings.Cut(strings.ToUpper(string(ins.Op)), ".")
	spec, recognized := x86WordMultiplyAddSpecs[Op(base)]
	if !recognized {
		return x86PackedBinaryForm{}, false, nil
	}
	form, err := x86PackedBinaryProperties(ins, goarch, spec)
	if err == nil {
		err = form.arithmeticMode32RegisterContext(ins, goarch)
	}
	return form, true, err
}
