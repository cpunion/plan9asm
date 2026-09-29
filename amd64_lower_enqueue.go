package plan9asm

import (
	"fmt"
	"strings"
)

var x86EnqueueOps = map[string]byte{
	"ENQCMD":  0xf2,
	"ENQCMDS": 0xf3,
}

// Keep the 64-byte device enqueue operation in hardware. An ordinary memory
// copy cannot implement PASID handling, atomic submission, or retry status.
func (c *amd64Ctx) lowerEnqueue(op Op, ins Instr) (bool, bool, error) {
	name := strings.ToUpper(string(op))
	base := strings.SplitN(name, ".", 2)[0]
	if _, ok := x86EnqueueOps[base]; !ok {
		return false, false, nil
	}
	if name != base || len(ins.Args) != 2 || !isAMD64MemoryOperand(ins.Args[0]) ||
		ins.Args[1].Kind != OpReg || !isX86YrlRegisterForArch(ins.Args[1].Reg, c.goarch) {
		return true, false, fmt.Errorf("%s %s expects a memory source and GP address register without suffixes: %q", c.goarch, base, ins.Raw)
	}
	if ins.Args[0].Kind == OpMem && !x86MemoryRegistersValidForArch(ins.Args[0].Mem, c.goarch) {
		return true, false, fmt.Errorf("%s %s source uses an invalid address register: %q", c.goarch, base, ins.Raw)
	}
	source := ins.Args[0]
	assemblySource := "$1"
	if source.Kind == OpMem && source.Mem.Segment != "" {
		switch source.Mem.Segment {
		case FS, GS:
			// LLVM inline-asm memory operands do not preserve the pointer's
			// segment address space. Spell the prefix explicitly and pass
			// its offset, not an ordinary unsegmented effective address.
			assemblySource = "%" + strings.ToLower(string(source.Mem.Segment)) + ":$1"
			source.Mem.Segment = ""
		default:
			return true, false, fmt.Errorf("unsupported enqueue source segment %s", source.Mem.Segment)
		}
	}
	ptr, ptrType, err := c.x86DescriptorMemoryPointer(source)
	if err != nil {
		return true, false, err
	}
	wordType, wordBits := I64, 64
	if c.goarch == "386" {
		wordType, wordBits = I32, 32
	}
	destination, err := c.evalIntSized(ins.Args[1], wordType)
	if err != nil {
		return true, false, err
	}
	if bits := ins.x86AddressBits; bits != 0 && bits != wordBits {
		if bits != wordBits/2 {
			return true, false, fmt.Errorf("invalid %s address size %d", c.goarch, bits)
		}
		// Truncate the effective offset before hardware applies any explicit
		// source-segment override.
		address := c.newTmp()
		fmt.Fprintf(c.b, "  %%%s = ptrtoint %s %s to i%d\n", address, ptrType, ptr, bits)
		wrapped := c.newTmp()
		fmt.Fprintf(c.b, "  %%%s = inttoptr i%d %%%s to %s\n", wrapped, bits, address, ptrType)
		ptr = "%" + wrapped
		low, extended := c.newTmp(), c.newTmp()
		fmt.Fprintf(c.b, "  %%%s = trunc %s %s to i%d\n", low, wordType, destination, bits)
		fmt.Fprintf(c.b, "  %%%s = zext i%d %%%s to %s\n", extended, bits, low, wordType)
		destination = "%" + extended
	}

	status := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = call i8 asm sideeffect %q, %q(%s elementtype([64 x i8]) %s, %s %s)\n",
		status, strings.ToLower(base)+" "+assemblySource+", $2; sete $0",
		"=q,*m,r,~{memory},~{dirflag},~{fpsr},~{flags}", ptrType, ptr, wordType, destination)
	c.storeSegmentQueryZF("%" + status)
	for _, slot := range []string{c.flagsCFSlot, c.flagsOFSlot, c.flagsSltSlot, c.flagsPFSlot} {
		fmt.Fprintf(c.b, "  store i1 false, ptr %s\n", slot)
	}
	return true, false, nil
}
