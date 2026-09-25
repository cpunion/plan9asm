package plan9asm

import "fmt"

type arm64RawSVELoadStore struct {
	load      bool
	immediate int
	base      int
	vector    int
}

// decodeARM64RawSVELoadStore covers the complete unpredicated SVE LDR/STR
// scalar-base format. Its signed imm9 is measured in whole vector lengths.
func decodeARM64RawSVELoadStore(word uint32) (arm64RawSVELoadStore, bool) {
	form := arm64RawSVELoadStore{}
	switch word & 0xffc0e000 {
	case 0x85804000:
		form.load = true
	case 0xe5804000:
		form.load = false
	default:
		return arm64RawSVELoadStore{}, false
	}
	imm9 := int(word>>16&0x3f)<<3 | int(word>>10&7)
	if imm9&0x100 != 0 {
		imm9 -= 0x200
	}
	form.immediate = imm9
	form.base = int(word>>5) & 31
	form.vector = int(word) & 31
	return form, true
}

func (c *arm64Ctx) lowerRawSVELoadStore(form arm64RawSVELoadStore) error {
	baseReg := Reg(fmt.Sprintf("R%d", form.base))
	if form.base == 31 {
		baseReg = SP
	}
	base, err := c.loadReg(baseReg)
	if err != nil {
		return err
	}
	address := base
	if form.immediate != 0 {
		vscale := c.newTmp()
		offset := c.newTmp()
		adjusted := c.newTmp()
		fmt.Fprintf(c.b, "  %%%s = call i64 @llvm.vscale.i64()\n", vscale)
		fmt.Fprintf(c.b, "  %%%s = mul i64 %%%s, %d\n", offset, vscale, form.immediate*16)
		fmt.Fprintf(c.b, "  %%%s = add i64 %s, %%%s\n", adjusted, base, offset)
		address = "%" + adjusted
	}
	pointer := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = inttoptr i64 %s to ptr\n", pointer, address)
	if form.load {
		loaded := c.newTmp()
		fmt.Fprintf(c.b, "  %%%s = load <vscale x 16 x i8>, ptr %%%s, align 1\n", loaded, pointer)
		return c.storeZReg(form.vector, "%"+loaded)
	}
	value, err := c.loadZReg(form.vector)
	if err != nil {
		return err
	}
	fmt.Fprintf(c.b, "  store <vscale x 16 x i8> %s, ptr %%%s, align 1\n", value, pointer)
	return nil
}
