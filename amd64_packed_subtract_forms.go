package plan9asm

import (
	"fmt"
	"strings"
)

type x86PackedSubtractForm struct {
	x86PackedBinaryForm
	laneBits int
	mode     amd64PackedSubtractMode
}

// Go's legacy subtract rows are yxm, never the add family's MMX rows.
// The V rows share _yvandnpd, but only D/Q enable broadcast in avx_optabs.
func x86PackedSubtractProperties(ins Instr, goarch string) (x86PackedSubtractForm, bool, error) {
	base, _, _ := strings.Cut(strings.ToUpper(string(ins.Op)), ".")
	bits, mode, vector, known := amd64PackedSubtractProperties(base)
	if !known {
		return x86PackedSubtractForm{}, false, nil
	}
	spec := x86PackedBinarySpec{encoding: x86PackedLogicalLegacyFP}
	if vector {
		spec.encoding, spec.zeroing = x86PackedLogicalVEXOrEVEX, true
		if base == "VPSUBD" || base == "VPSUBQ" {
			spec.broadcastBytes = bits / 8
		}
	}
	form, err := x86PackedBinaryProperties(ins, goarch, spec)
	if err == nil && goarch == "386" && form.byteWidth == 64 {
		err = fmt.Errorf("386 %s has no Z-width form in Go's assembler frontend: %q", ins.Op, ins.Raw)
	}
	if err == nil {
		err = form.arithmeticMode32RegisterContext(ins, goarch)
	}
	return x86PackedSubtractForm{form, bits, mode}, true, err
}
