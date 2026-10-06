package plan9asm

import (
	"fmt"
	"strings"
)

type x86UndefinedTrapSpec struct {
	native  string
	opBytes []byte // Complete canonical Go ynone/Px row.
}

// UNDEF is Go's canonical spelling; UD2 is the physical decoder/internal
// spelling of the same opcode. Neither spelling is a returning debug trap.
var x86UndefinedTrapSpecs = map[Op]x86UndefinedTrapSpec{
	"UNDEF": {native: "ud2", opBytes: []byte{0x0f, 0x0b}},
}

func x86UndefinedTrapProperties(op Op, ins Instr, arch string) (x86UndefinedTrapSpec, bool, error) {
	name, _, suffix := strings.Cut(strings.ToUpper(string(op)), ".")
	canonical := name
	if name == "UD2" {
		canonical = "UNDEF"
	}
	spec, known := x86UndefinedTrapSpecs[Op(canonical)]
	if !known {
		return x86UndefinedTrapSpec{}, false, nil
	}
	if suffix || len(ins.Args) != 0 || arch != "386" && arch != "amd64" ||
		ins.Op != "" && !strings.EqualFold(string(ins.Op), name) {
		return spec, true, fmt.Errorf("%s %s requires an exact zero-operand undefined-instruction form: %q", arch, op, ins.Raw)
	}
	return spec, true, nil
}

func (c *amd64Ctx) lowerX86UndefinedTrap(op Op, ins Instr) (bool, bool, error) {
	spec, known, err := x86UndefinedTrapProperties(op, ins, c.goarch)
	if !known || err != nil {
		return known, false, err
	}
	// Use the exact architectural fault instruction, not merely unreachable
	// or a target-dependent generic trap that may select a different signal.
	fmt.Fprintf(c.b, "  call void asm sideeffect %q, %q()\n", spec.native, "~{memory}")
	c.b.WriteString("  unreachable\n")
	return true, true, nil
}
