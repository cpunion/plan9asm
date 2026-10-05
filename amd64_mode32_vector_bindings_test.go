package plan9asm

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"testing"
)

type mode32VectorFamilyCase struct {
	name       string
	rows       []string
	properties func(Instr, string) (bool, error)
	lower      func(*amd64Ctx, Instr) (bool, bool, error)
}

// Existing SDK-table completeness tests independently bind these families to
// Go. The actual Go object here observes syntax acceptance, not register state.
func mode32VectorFamilyCases() []mode32VectorFamilyCase {
	logical := mode32VectorFamilyCase{
		name: "logical",
		properties: func(ins Instr, arch string) (bool, error) {
			_, recognized, err := x86PackedLogicalProperties(ins, arch)
			return recognized, err
		},
		lower: func(ctx *amd64Ctx, ins Instr) (bool, bool, error) {
			return ctx.lowerPackedLogicalFamily(ins.Op, ins)
		},
	}
	for _, opcode := range x86LogicalOpcodeNames() {
		if x86PackedLogicalSpecs[opcode].encoding > x86PackedLogicalLegacyInteger {
			logical.rows = append(logical.rows, x86LogicalAcceptedForms(opcode, "386")...)
		}
	}
	move := mode32VectorFamilyCase{
		name: "move",
		properties: func(ins Instr, arch string) (bool, error) {
			_, recognized, err := amd64PackedVectorMoveProperties(ins, arch)
			return recognized, err
		},
		lower: func(ctx *amd64Ctx, ins Instr) (bool, bool, error) {
			return ctx.lowerPackedVectorMove(ins.Op, ins)
		},
	}
	var moves []string
	for opcode := range amd64PackedVectorMoveSpecs {
		moves = append(moves, string(opcode))
	}
	sort.Strings(moves)
	for _, opcode := range moves {
		spec := amd64PackedVectorMoveSpecs[Op(opcode)]
		high := 31
		if spec.encoding == amd64PackedMoveVEX {
			high = 15
		}
		for _, bank := range []string{"X", "Y"} {
			reg := fmt.Sprintf("%s%d", bank, high)
			for _, args := range []string{"(AX)," + reg, reg + ",(AX)", reg + "," + bank + "0"} {
				move.rows = append(move.rows, opcode+" "+args)
			}
			if spec.encoding != amd64PackedMoveVEX {
				for _, args := range []string{"(AX),K7," + reg, reg + ",K7,(AX)", reg + ",K7," + bank + "0"} {
					move.rows = append(move.rows, opcode+" "+args, opcode+".Z "+args)
				}
			}
		}
	}
	fma := mode32VectorFamilyCase{
		name: "FMA3",
		properties: func(ins Instr, arch string) (bool, error) {
			_, recognized, err := amd64FMA3Properties(ins, arch)
			return recognized, err
		},
		lower: func(ctx *amd64Ctx, ins Instr) (bool, bool, error) {
			return ctx.lowerFMA3(ins.Op, ins)
		},
	}
	// Scalar rounding, memory and packed broadcast axes remain in the source.
	// Z0-Z7 are physically available; this regression targets the X/Y banks.
	for _, row := range x86FMA3NamedFormats("386") {
		changed := strings.NewReplacer("X2", "X16", "Y2", "Y16").Replace(row)
		if changed != row {
			fma.rows = append(fma.rows, changed)
		}
	}
	return []mode32VectorFamilyCase{logical, move, fma}
}

func TestX86Mode32VectorFamiliesRequirePhysicalBindings(t *testing.T) {
	for _, family := range mode32VectorFamilyCases() {
		for _, row := range family.rows {
			file, err := Parse(ArchAMD64, "TEXT f(SB),4,$0-0\n"+row+"\nRET\n")
			if err != nil {
				t.Fatal(err)
			}
			instruction := file.Funcs[0].Instrs[1]
			high := false
			for _, operand := range instruction.Args {
				width := amd64VectorByteWidth(operand.Reg)
				index, valid := amd64VectorRegisterIndex(operand.Reg, width)
				high = high || operand.Kind == OpReg && valid && index >= 8
			}
			if !high {
				continue
			}
			t.Run(family.name+"/"+strings.ReplaceAll(row, " ", "_"), func(t *testing.T) {
				requireX86GoAssemblerResult(t, "386", "TEXT f(SB),4,$0-0\n"+row+"\nRET\n", true)
				if recognized, err := family.properties(instruction, "386"); !recognized || !errors.Is(err, ErrProbeNeedsContext) {
					t.Fatalf("Go-accepted high bank became a physical mode32 binding: recognized=%v err=%v", recognized, err)
				}
				ctx, output := newAMD64CtxWithFuncForTest(t, file.Funcs[0], FuncSig{Ret: Void}, nil)
				ctx.goarch = "386"
				before := output.String()
				if recognized, _, err := family.lower(ctx, instruction); !recognized ||
					!errors.Is(err, ErrProbeNeedsContext) || output.String() != before {
					t.Fatalf("unproved high-bank source emitted IR: recognized=%v err=%v", recognized, err)
				}
			})
		}
	}
}
