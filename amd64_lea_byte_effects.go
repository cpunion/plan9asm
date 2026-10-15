package plan9asm

import (
	"fmt"
	"strings"
)

// This describes only the pure numerical effective-address subset of Go's
// complete ym_rl grammar. The register dependencies are read before the
// destination is overwritten; no memory is read and FLAGS are unchanged.
// Modulo-W/L/Q arithmetic needs only those low input bytes. A W write keeps
// the other bytes, while L clears upper32 through the existing write consumer.
// Actual initialization and pointer exclusion remain consumers' obligations.
func x86TypedLEAByteEffect(op Op, original Instr, arch string) (x86TypedByteEffect, bool, error) {
	name, _, dotted := strings.Cut(strings.ToUpper(string(op)), ".")
	if name != "LEAW" && name != "LEAL" && name != "LEAQ" {
		return x86TypedByteEffect{}, false, nil
	}
	// A decoded physical instruction belongs to its existing separate effect
	// consumers. Do not intercept it with this named-source registration.
	if original.x86Encoded || original.x86SourceRaw != nil || original.x86RawNearReturn != nil ||
		original.x86AddressBits != 0 || original.x86UnboundAddressOverride || original.x86SegmentPrefix != 0 {
		return x86TypedByteEffect{}, false, nil
	}
	fail := func() (x86TypedByteEffect, bool, error) {
		return x86TypedByteEffect{}, true, fmt.Errorf("%w: %s %s needs a pure numerical Go effective address: %q", ErrProbeNeedsContext, arch, op, original.Raw)
	}
	if arch != "amd64" && arch != "386" || dotted ||
		original.Op != "" && !strings.EqualFold(string(original.Op), name) {
		return fail()
	}
	ins := original
	ins.Op = Op(name)
	ins = x86LEASourceInstruction(ins)
	width, err := x86LEAOperandWidth(ins, arch)
	if err != nil {
		return x86TypedByteEffect{}, true, err
	}
	if ins.Args[0].Kind != OpMem {
		return fail()
	}
	memory := ins.Args[0].Mem
	if memory.Sym != "" || memory.OffRaw != "" || memory.Segment != "" || memory.IndexExt != "" ||
		memory.Index == "" && memory.Scale != 0 {
		return fail()
	}
	// A syntactically valid physical stack address may already have an exact
	// source-stack consumer. Leave it to that independent role; the numerical
	// effect must neither steal its FLAGS handling nor grant any stack fact.
	for _, register := range []Reg{memory.Base, memory.Index, ins.Args[1].Reg} {
		if register == SP || register == BP && arch != "386" {
			return x86TypedByteEffect{}, false, nil
		}
	}
	if _, _, valid := x86ResultGP(ins.Args[1].Reg, width, arch); !valid {
		return fail()
	}
	effect := x86TypedByteEffect{}
	for _, register := range []Reg{memory.Base, memory.Index} {
		if register == "" {
			continue
		}
		if _, _, valid := x86ResultGP(register, width, arch); !valid {
			return fail()
		}
		effect.operands = append(effect.operands, x86TypedByteOperand{
			index: 0, operand: Operand{Kind: OpReg, Reg: register}, width: width, read: true,
		})
	}
	effect.operands = append(effect.operands, x86TypedByteOperand{
		index: 1, operand: ins.Args[1], width: width, write: true,
	})
	return effect, true, nil
}
