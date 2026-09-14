package plan9asm

import (
	"fmt"
	"strings"
)

// lowerPackedScalarExtract implements every Go 1.27 operand form from the
// legacy yextr/yextrw tables and the AVX _yvextractps/_yvpextrw tables.
func (c *amd64Ctx) lowerPackedScalarExtract(op Op, ins Instr) (ok bool, terminated bool, err error) {
	rawOp := strings.ToUpper(string(op))
	baseOp := rawOp
	if dot := strings.IndexByte(rawOp, '.'); dot >= 0 {
		baseOp = rawOp[:dot]
	}

	bits := 0
	vector := false
	switch baseOp {
	case "PEXTRB":
		bits = 8
	case "PEXTRW":
		bits = 16
	case "PEXTRD":
		bits = 32
	case "PEXTRQ":
		bits = 64
	case "VPEXTRB":
		bits, vector = 8, true
	case "VPEXTRW":
		bits, vector = 16, true
	case "VPEXTRD":
		bits, vector = 32, true
	case "VPEXTRQ":
		bits, vector = 64, true
	default:
		return false, false, nil
	}

	if rawOp != baseOp {
		return true, false, fmt.Errorf("amd64 %s has no instruction suffixes in Go 1.27's optab: %q", baseOp, ins.Raw)
	}
	if len(ins.Args) != 3 || ins.Args[0].Kind != OpImm || ins.Args[1].Kind != OpReg {
		return true, false, fmt.Errorf("amd64 %s expects $imm8, X source, GP-or-memory destination: %q", baseOp, ins.Raw)
	}

	sourceIndex, sourceOK := amd64ParseXReg(ins.Args[1].Reg)
	if !sourceOK || (!vector && sourceIndex > 15) {
		return true, false, fmt.Errorf("amd64 %s source is outside its Go 1.27 X-register class: %q", baseOp, ins.Raw)
	}

	// The VEX entries cover both Yi8 and Yu8, while the EVEX entries (which
	// are required for X16-X31) cover only Yu8. Legacy entries also use Yu8.
	minimumImmediate := int64(0)
	if vector && sourceIndex <= 15 {
		minimumImmediate = -128
	}
	if immediate := ins.Args[0].Imm; immediate < minimumImmediate || immediate > 255 {
		return true, false, fmt.Errorf("amd64 %s immediate is outside its Go 1.27 encoding table: %q", baseOp, ins.Raw)
	}

	destination := ins.Args[2]
	destinationReg := Reg("")
	if destination.Kind == OpReg {
		if isAMD64YrlRegister(destination.Reg) {
			destinationReg = destination.Reg
		} else if mmxIndex, mmx := amd64ParseMReg(destination.Reg); mmx && !vector && bits != 16 {
			// The legacy B/D/Q yextr table uses Ymm. Go therefore accepts M0-M7
			// here and encodes their ModRM numbers as AX/CX/DX/BX/SP/BP/SI/DI.
			destinationReg = amd64MMXEncodingGP(mmxIndex)
		} else {
			return true, false, fmt.Errorf("amd64 %s register destination is outside its Go 1.27 table: %q", baseOp, ins.Raw)
		}
	} else if !isAMD64MemoryOperand(destination) {
		return true, false, fmt.Errorf("amd64 %s destination must be a GP register or memory: %q", baseOp, ins.Raw)
	}

	source, err := c.loadX(ins.Args[1].Reg)
	if err != nil {
		return true, false, err
	}
	lanes := 128 / bits
	laneVector := c.newTmp()
	fmt.Fprintf(c.b, "  %%%s = bitcast <16 x i8> %s to <%d x i%d>\n", laneVector, source, lanes, bits)
	extracted := c.newTmp()
	lane := int(uint64(ins.Args[0].Imm) & uint64(lanes-1))
	fmt.Fprintf(c.b, "  %%%s = extractelement <%d x i%d> %%%s, i32 %d\n", extracted, lanes, bits, laneVector, lane)
	value := "%" + extracted

	if destination.Kind == OpReg {
		if bits < 32 {
			widened := c.newTmp()
			fmt.Fprintf(c.b, "  %%%s = zext i%d %s to i32\n", widened, bits, value)
			return true, false, c.storeRegSized(destinationReg, I32, "%"+widened)
		}
		return true, false, c.storeRegSized(destinationReg, amd64IntegerTypeForBits(bits), value)
	}
	return true, false, c.storePackedScalarExtractMemory(destination, amd64IntegerTypeForBits(bits), value)
}

func amd64MMXEncodingGP(index int) Reg {
	return [...]Reg{AX, CX, DX, BX, SP, BP, SI, DI}[index]
}

func (c *amd64Ctx) storePackedScalarExtractMemory(destination Operand, typ LLVMType, value string) error {
	switch destination.Kind {
	case OpMem:
		pointer, pointerType, err := c.ptrFromMem(destination.Mem)
		if err != nil {
			return err
		}
		fmt.Fprintf(c.b, "  store %s %s, %s %s, align 1\n", typ, value, pointerType, pointer)
		return nil
	case OpSym:
		pointer, err := c.ptrFromSB(destination.Sym)
		if err != nil {
			return err
		}
		fmt.Fprintf(c.b, "  store %s %s, ptr %s, align 1\n", typ, value, pointer)
		return nil
	case OpFP:
		return c.storeFPResult(destination.FPOffset, typ, value)
	default:
		return fmt.Errorf("expected packed scalar extract memory destination, got %s", destination.String())
	}
}
