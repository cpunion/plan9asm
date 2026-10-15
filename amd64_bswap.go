package plan9asm

import (
	"fmt"
	"strings"
)

// Go's complete ybswap row is a single Yrl destination encoded by Z_rp.
// There is no byte/word, memory, immediate, mask or suffix form.
var x86ByteSwapSpecs = map[Op]int{"BSWAPL": 32, "BSWAPQ": 64}

func x86ByteSwapProperties(op Op, ins Instr, arch string) (int, bool, error) {
	name, _, dotted := strings.Cut(strings.ToUpper(string(op)), ".")
	bits, known := x86ByteSwapSpecs[Op(name)]
	if !known {
		return 0, false, nil
	}
	if arch != "amd64" && arch != "386" || arch == "386" && bits == 64 || dotted ||
		ins.Op != "" && !strings.EqualFold(string(ins.Op), name) || len(ins.Args) != 1 ||
		ins.Args[0].Kind != OpReg || !isX86YrlRegisterForArch(ins.Args[0].Reg, arch) {
		return bits, true, fmt.Errorf("%s %s outside Go BSWAP register grammar: %q", arch, op, ins.Raw)
	}
	return bits, true, nil
}

func x86TypedByteSwapEffect(op Op, ins Instr, arch string) (x86TypedByteEffect, bool, error) {
	bits, known, err := x86ByteSwapProperties(op, ins, arch)
	if !known || err != nil {
		return x86TypedByteEffect{}, known, err
	}
	return x86TypedByteEffect{operands: []x86TypedByteOperand{
		{index: 0, operand: ins.Args[0], width: int64(bits / 8), read: true, write: true},
	}}, true, nil // BSWAP does not read or modify FLAGS, DF or any memory.
}

// Byte facts are permuted, not widened to a full initialized word. The same
// dependency transports independently proved canonical bytes, never values,
// pointer identity, native backing or a new canonicality witness.
func x86ByteSwapByteMask(before uint8, bits int) uint8 {
	var after uint8
	for index := 0; index < bits/8; index++ {
		if before&(1<<uint(index)) != 0 {
			after |= 1 << uint(bits/8-1-index)
		}
	}
	return after
}

func x86ResultByteSwapStep(ins Instr, state *x86ResultState, arch string) (bool, bool) {
	bits, known, err := x86ByteSwapProperties(ins.Op, ins, arch)
	if !known || err != nil {
		return false, known
	}
	index, _, valid := x86ResultGP(ins.Args[0].Reg, int64(bits/8), arch)
	if !valid || !x86ResultNamedPredicateInstruction(ins) {
		return false, true
	}
	state.gp[index] = x86ByteSwapByteMask(state.gp[index], bits)
	if bits == 32 {
		state.gp[index] |= 0xf0 // The internal i64 carrier zero-extends r32.
	}
	state.pointerGP[index] = false
	return true, true
}

func (c *amd64Ctx) byteSwapResultInitialized() error {
	if err := c.scalarFPResultInitialized(); err != nil {
		return err
	}
	// Generic result-byte initialization is not a canonical bool witness.
	// The ordinary386 consumer covers only a single logical I1 return, not
	// a boolean field in an aggregate. Other independently authenticated
	// consumers retain their own complete result/canonicality contracts.
	canonicalABI0 := c.goarch == "386" && c.sig.Ret == I1
	physicalBool := c.fpScalarSource.x86SourceStack != nil && c.sig.Ret == I1
	if canonicalABI0 || physicalBool ||
		c.fpScalarSource.x86FixedLocal != nil || x86RegisterBoolResultNeedsProof(c.fpScalarSource, c.sig, c.goarch) {
		return nil
	}
	for _, slot := range c.fpResults {
		if slot.Type == I1 {
			return fmt.Errorf("%w: BSWAP bool result requires an independent canonical result role", ErrProbeNeedsContext)
		}
	}
	return nil
}

func (c *amd64Ctx) lowerByteSwap(op Op, ins Instr) (bool, bool, error) {
	bits, known, err := x86ByteSwapProperties(op, ins, c.goarch)
	if !known || err != nil {
		return known, false, err
	}
	typ := amd64IntegerTypeForBits(bits)
	value, err := c.evalIntSized(ins.Args[0], typ)
	if err != nil {
		return true, false, err
	}
	result := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = call %s @llvm.bswap.%s(%s %s)\n", result, typ, typ, typ, value)
	return true, false, c.storeRegSized(ins.Args[0].Reg, typ, "%"+result)
}
