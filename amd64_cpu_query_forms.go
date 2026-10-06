package plan9asm

import (
	"fmt"
	"strings"
)

type x86CPUQuerySpec struct {
	native  string
	inputs  []Reg
	outputs []Reg
	opBytes []byte // Complete Go ynone/Pm opcode row, after its 0x0f prefix.
}

var x86CPUQuerySpecs = map[Op]x86CPUQuerySpec{
	OpCPUID:  {"cpuid", []Reg{AX, CX}, []Reg{AX, BX, CX, DX}, []byte{0xa2}},
	OpXGETBV: {"xgetbv", []Reg{CX}, []Reg{AX, DX}, []byte{0x01, 0xd0}},
}

func x86CPUQueryProperties(op Op, ins Instr, arch string) (x86CPUQuerySpec, bool, error) {
	name, _, dotted := strings.Cut(strings.ToUpper(string(op)), ".")
	spec, known := x86CPUQuerySpecs[Op(name)]
	if !known {
		return x86CPUQuerySpec{}, false, nil
	}
	if len(ins.Args) != 0 || dotted || arch != "amd64" && arch != "386" ||
		ins.Op != "" && !strings.EqualFold(string(ins.Op), name) {
		return spec, true, fmt.Errorf("%s %s requires the exact zero-operand Go CPU-query form: %q", arch, op, ins.Raw)
	}
	return spec, true, nil
}

func (spec x86CPUQuerySpec) resultType() LLVMType {
	return LLVMType("{ " + strings.TrimSuffix(strings.Repeat("i32, ", len(spec.outputs)), ", ") + " }")
}

func (spec x86CPUQuerySpec) constraints() string {
	var fields []string
	for _, register := range spec.outputs {
		fields = append(fields, "={"+strings.ToLower(string(register))+"}")
	}
	for _, register := range spec.inputs {
		fields = append(fields, "{"+strings.ToLower(string(register))+"}")
	}
	// Preserve source ordering at the compiler boundary. These conservative
	// clobbers do not assert that XGETBV serializes hardware or changes flags.
	fields = append(fields, "~{dirflag}", "~{fpsr}", "~{flags}", "~{memory}")
	return strings.Join(fields, ",")
}

func emitX86CPUQuery(b *strings.Builder, spec x86CPUQuerySpec, inputs []string, newTmp func() string) []string {
	arguments := make([]string, len(inputs))
	for index, input := range inputs {
		arguments[index] = "i32 " + input
	}
	call := newTmp()
	fmt.Fprintf(b, "  %%%s = call %s asm sideeffect %q, %q(%s)\n",
		call, spec.resultType(), spec.native, spec.constraints(), strings.Join(arguments, ", "))
	outputs := make([]string, len(spec.outputs))
	for index := range outputs {
		part := newTmp()
		fmt.Fprintf(b, "  %%%s = extractvalue %s %%%s, %d\n", part, spec.resultType(), call, index)
		outputs[index] = "%" + part
	}
	return outputs
}

func (c *amd64Ctx) lowerX86CPUQuery(op Op, ins Instr) (bool, bool, error) {
	spec, known, err := x86CPUQueryProperties(op, ins, c.goarch)
	if !known || err != nil {
		return known, false, err
	}
	inputs := make([]string, len(spec.inputs))
	for index, register := range spec.inputs {
		value, err := c.loadReg(register)
		if err != nil {
			return true, false, err
		}
		inputs[index] = c.truncI64(value, I32)
	}
	outputs := emitX86CPUQuery(c.b, spec, inputs, c.newTmp)
	for index, register := range spec.outputs {
		if err := c.storeRegSized(register, I32, outputs[index]); err != nil {
			return true, false, err
		}
	}
	return true, false, nil
}
