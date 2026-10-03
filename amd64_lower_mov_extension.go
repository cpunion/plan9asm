package plan9asm

import "fmt"

func (c *amd64Ctx) lowerScalarExtensionMove(op Op, ins Instr) error {
	form, recognized, err := x86ScalarExtensionProperties(op, ins, c.goarch)
	if err != nil {
		return err
	}
	if !recognized {
		return fmt.Errorf("unrecognized scalar extension opcode %s: %q", op, ins.Raw)
	}
	spec := form.spec
	sourceType := amd64IntegerTypeForBits(spec.sourceBits)
	destinationType := amd64IntegerTypeForBits(spec.destinationBits)
	value, err := c.evalIntSized(form.source, sourceType)
	if err != nil {
		return err
	}
	if spec.sourceBits < spec.destinationBits {
		extended := c.newTmp()
		extension := "zext"
		if spec.signed {
			extension = "sext"
		}
		fmt.Fprintf(c.b, "  %%%s = %s %s %s to %s\n", extended, extension, sourceType, value, destinationType)
		value = "%" + extended
	}
	return c.storeRegSized(form.result, destinationType, value)
}
