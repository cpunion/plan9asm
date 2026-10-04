package plan9asm

import (
	"fmt"
	"strings"
)

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
	if err == nil && goarch == "386" && !form.legacy {
		for _, operand := range ins.Args {
			index, valid := amd64VectorRegisterIndex(operand.Reg, form.byteWidth)
			if operand.Kind == OpReg && valid && index >= 8 {
				return form, true, fmt.Errorf("%w: 386 packed multiply-add register %s has no proved physical mode32 binding", ErrProbeNeedsContext, operand.Reg)
			}
		}
	}
	return form, true, err
}
