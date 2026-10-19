package plan9asm

import "fmt"

type amd64ADXCarry uint8

const (
	amd64ADXCarryCF amd64ADXCarry = iota
	amd64ADXCarryOF
)

type amd64ADXSpec struct {
	bits   int
	prefix byte
	carry  amd64ADXCarry
}

// Go 1.27 asm6.go uses yml_rl for every ADX form. One spec binds the
// register/memory grammar, encoded mandatory prefix, width and carry chain.
var amd64ADXSpecs = map[Op]amd64ADXSpec{
	"ADCXL": {bits: 32, prefix: 0x66, carry: amd64ADXCarryCF},
	"ADCXQ": {bits: 64, prefix: 0x66, carry: amd64ADXCarryCF},
	"ADOXL": {bits: 32, prefix: 0xf3, carry: amd64ADXCarryOF},
	"ADOXQ": {bits: 64, prefix: 0xf3, carry: amd64ADXCarryOF},
}

// lowerADX implements ADCXL/Q and ADOXL/Q. All four opcodes use Go 1.27's
// yml_rl row: register/memory source and register destination. ADCX consumes
// and updates CF; ADOX independently consumes and updates OF.
func (c *amd64Ctx) lowerADX(op Op, ins Instr) (ok bool, terminated bool, err error) {
	form, known, err := x86ADXProperties(op, ins, c.goarch)
	if !known || err != nil {
		return known, false, err
	}
	bits, flagSlot := form.spec.bits, c.flagsCFSlot
	if form.spec.carry == amd64ADXCarryOF {
		flagSlot = c.flagsOFSlot
	}
	source, destination := form.source, form.destination
	typ := amd64IntegerTypeForBits(bits)
	sourceValue, err := c.evalADXSource(source, typ, ins.x86AddressBits)
	if err != nil {
		return true, false, err
	}
	destinationValue, err := c.evalIntSized(destination, typ)
	if err != nil {
		return true, false, err
	}
	carryIn := c.loadFlag(flagSlot)
	carryValue := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = zext i1 %s to %s\n", carryValue, carryIn, typ)
	sum := c.newTmp()
	result := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = add %s %s, %s\n", sum, typ, destinationValue, sourceValue)
	fmt.Fprintf(c.b, "  %%%s = add %s %%%s, %%%s\n", result, typ, sum, carryValue)
	carryFromOperands := c.newTmp()
	carryFromInput := c.newTmp()
	carryOut := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = icmp ult %s %%%s, %s\n", carryFromOperands, typ, sum, destinationValue)
	fmt.Fprintf(c.b, "  %%%s = icmp ult %s %%%s, %%%s\n", carryFromInput, typ, result, sum)
	fmt.Fprintf(c.b, "  %%%s = or i1 %%%s, %%%s\n", carryOut, carryFromOperands, carryFromInput)
	if err := c.storeRegSized(destination.Reg, typ, "%"+result); err != nil {
		return true, false, err
	}
	fmt.Fprintf(c.b, "  store i1 %%%s, ptr %s\n", carryOut, flagSlot)
	return true, false, nil
}

func (c *amd64Ctx) evalADXSource(source Operand, typ LLVMType, addressBits int) (string, error) {
	wordBits := 64
	if c.goarch == "386" {
		wordBits = 32
	}
	if source.Kind != OpMem && source.Kind != OpSym {
		return c.evalIntSized(source, typ)
	}
	if addressBits != 0 && addressBits != wordBits && addressBits != wordBits/2 {
		return "", fmt.Errorf("invalid ADX address size %d for %s", addressBits, c.goarch)
	}
	ptr, ptrType, _, err := c.packedVectorMemoryPointer(source)
	if err != nil {
		return "", err
	}
	if source.Kind == OpMem && addressBits != 0 && addressBits != wordBits {
		// Wrap only the effective offset, before the FS/GS hardware base.
		address, wrapped := c.newTmp(), c.newTmp()
		fmt.Fprintf(c.b, "  %%%s = ptrtoint %s %s to i%d\n", address, ptrType, ptr, addressBits)
		fmt.Fprintf(c.b, "  %%%s = inttoptr i%d %%%s to %s\n", wrapped, addressBits, address, ptrType)
		ptr = "%" + wrapped
	}
	value := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = load volatile %s, %s %s, align 1\n", value, typ, ptrType, ptr)
	return "%" + value, nil
}
