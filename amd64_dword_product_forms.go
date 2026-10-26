package plan9asm

import "strings"

type x86DwordProductSpec struct {
	x86PackedBinarySpec
	signed       bool
	nativeOpcode string
}

// Go names the legacy unsigned instruction PMULULQ, not Intel's PMULUDQ.
// Only the unsigned legacy ymm table has an MMX/m64 row. Both AVX tables
// permit qword broadcast and qword writemasking, never rounding or SAE.
var x86DwordProductSpecs = map[Op]x86DwordProductSpec{
	"PMULDQ":   {x86PackedBinarySpec{encoding: x86PackedLogicalLegacyFP}, true, "pmuldq"},
	"PMULULQ":  {x86PackedBinarySpec{encoding: x86PackedLogicalLegacyInteger}, false, "pmuludq"},
	"VPMULDQ":  {x86PackedBinarySpec{encoding: x86PackedLogicalVEXOrEVEX, broadcastBytes: 8, zeroing: true}, true, ""},
	"VPMULUDQ": {x86PackedBinarySpec{encoding: x86PackedLogicalVEXOrEVEX, broadcastBytes: 8, zeroing: true}, false, ""},
}

type x86DwordProductForm struct {
	x86PackedBinaryForm
	spec   x86DwordProductSpec
	source Operand
}

// The operand model grants no initialized bank, native extent or caller role.
// In particular a GP spelling admitted to Ymm still names the physical MMX
// bank, and a named high mode32 vector still needs its physical binding.
func x86DwordProductProperties(op Op, ins Instr, arch string) (x86DwordProductForm, bool, error) {
	base, _, _ := strings.Cut(strings.ToUpper(string(op)), ".")
	spec, known := x86DwordProductSpecs[Op(base)]
	if !known {
		return x86DwordProductForm{}, false, nil
	}
	binary, source, started, err := x86PackedBinaryPhysicalSource(op, ins, arch, spec.x86PackedBinarySpec, "dword product")
	if !started {
		return x86DwordProductForm{}, true, err
	}
	return x86DwordProductForm{binary, spec, source}, true, err
}
