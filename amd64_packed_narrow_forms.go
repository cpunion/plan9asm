package plan9asm

import (
	"fmt"
	"strings"
)

type x86PackedNarrowSpec struct {
	x86PackedBinarySpec
	inputBits, outputBits int
	mode                  amd64PackedNarrowMode
}

// Go's ymm rows admit MMX only in amd64 mode; PACKUSDW is yxm_q4.
// All four V rows use _yvandnpd, with N4 broadcast only for dword inputs.
// Memory exception behavior is E4NF (SSWB/USWB: E4NF.nb), independently
// of their output-lane writemask. No initialized bank or ABI is granted here.
var x86PackedNarrowSpecs = map[Op]x86PackedNarrowSpec{
	"PACKSSLW":  {x86PackedBinarySpec{encoding: x86PackedLogicalLegacyInteger}, 32, 16, amd64PackedNarrowSigned},
	"PACKSSWB":  {x86PackedBinarySpec{encoding: x86PackedLogicalLegacyInteger}, 16, 8, amd64PackedNarrowSigned},
	"PACKUSDW":  {x86PackedBinarySpec{encoding: x86PackedLogicalLegacyFP}, 32, 16, amd64PackedNarrowUnsigned},
	"PACKUSWB":  {x86PackedBinarySpec{encoding: x86PackedLogicalLegacyInteger}, 16, 8, amd64PackedNarrowUnsigned},
	"VPACKSSDW": {x86PackedBinarySpec{encoding: x86PackedLogicalVEXOrEVEX, broadcastBytes: 4, zeroing: true}, 32, 16, amd64PackedNarrowSigned},
	"VPACKSSWB": {x86PackedBinarySpec{encoding: x86PackedLogicalVEXOrEVEX, zeroing: true}, 16, 8, amd64PackedNarrowSigned},
	"VPACKUSDW": {x86PackedBinarySpec{encoding: x86PackedLogicalVEXOrEVEX, broadcastBytes: 4, zeroing: true}, 32, 16, amd64PackedNarrowUnsigned},
	"VPACKUSWB": {x86PackedBinarySpec{encoding: x86PackedLogicalVEXOrEVEX, zeroing: true}, 16, 8, amd64PackedNarrowUnsigned},
}

type x86PackedNarrowForm struct {
	x86PackedBinaryForm
	spec   x86PackedNarrowSpec
	source Operand
}

func x86PackedNarrowProperties(ins Instr, goarch string) (x86PackedNarrowForm, bool, error) {
	base, _, _ := strings.Cut(strings.ToUpper(string(ins.Op)), ".")
	spec, known := x86PackedNarrowSpecs[Op(base)]
	if !known {
		return x86PackedNarrowForm{}, false, nil
	}
	local := ins
	local.Args = append([]Operand(nil), ins.Args...)
	// Go's GP-to-Ymm compatibility encodes MMX ModRM bits, not a GP value.
	// Retain that identity for syntax; the emitter still needs its transport.
	if spec.encoding == x86PackedLogicalLegacyInteger && goarch == "amd64" && len(local.Args) == 2 {
		if _, mmx := amd64ParseMReg(local.Args[1].Reg); local.Args[1].Kind == OpReg && mmx &&
			local.Args[0].Kind == OpReg && isX86YrlRegisterForArch(local.Args[0].Reg, goarch) {
			for index, register := range []Reg{AX, CX, DX, BX, SP, BP, SI, DI, "R8", "R9", "R10", "R11", "R12", "R13", "R14", "R15"} {
				if local.Args[0].Reg == register {
					local.Args[0].Reg = Reg(fmt.Sprintf("M%d", index&7))
					break
				}
			}
		}
	}
	grammarArch := goarch
	if goarch == "386" && ins.x86Encoded && spec.encoding == x86PackedLogicalVEXOrEVEX {
		// Validated physical bytes may have masks which Go's mode32 named
		// frontend rejects. The private marker is not JSON-exportable.
		grammarArch = "amd64"
	}
	form, err := x86PackedBinaryProperties(local, grammarArch, spec.x86PackedBinarySpec)
	if err == nil {
		err = x86Mode32VectorRegisterContext(ins, goarch)
	}
	if err == nil && grammarArch != goarch && form.memory >= 0 && !x86PackedShiftMemoryValid(local.Args[form.memory], goarch) {
		err = fmt.Errorf("386 %s physical source has an unbound address", ins.Op)
	}
	if err == nil && form.memory >= 0 && (ins.x86AddressBits != 0 || ins.x86UnboundAddressOverride) {
		err = fmt.Errorf("%w: packed narrow has no bound address-size override", ErrProbeNeedsContext)
	}
	result := x86PackedNarrowForm{x86PackedBinaryForm: form, spec: spec}
	if err == nil {
		result.source = local.Args[form.first]
	}
	return result, true, err
}
