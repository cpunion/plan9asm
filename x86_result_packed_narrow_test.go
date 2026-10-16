package plan9asm

import (
	"strings"
	"testing"
)

// Independent ymm/yxm_q4 and _yvandnpd families, including each opcode's
// separate N4 broadcast attribute. Neither inventory reads production specs.
func x86NarrowValueRows(arch string) []struct{ op, bank, mode string } {
	rows := x86FixedBinaryValueRows(arch, []string{
		"PACKSSLW", "PACKSSWB", "PACKUSDW", "PACKUSWB",
		"VPACKSSDW", "VPACKSSWB", "VPACKUSDW", "VPACKUSWB",
	})
	broadcast := x86FixedBinaryValueRows(arch, []string{"VPACKSSDW", "VPACKUSDW"})
	for _, row := range broadcast {
		row.op += ".BCST"
		rows = append(rows, row)
	}
	return rows
}

func TestX86PackedNarrowValueCompleteFamilyFiveTargetsThreeAPIs(t *testing.T) {
	x86FixedBinaryCompleteFamilyObjects(t, x86NarrowValueRows, 58, 22)
}

func TestX86PackedNarrowVoidCannotInventRegisterInputs(t *testing.T) {
	x86FixedBinaryVoidInputBoundaries(t, func(arch string) []struct{ op, bank, mode string } {
		var rows []struct{ op, bank, mode string }
		for _, row := range x86NarrowValueRows(arch) {
			if !strings.Contains(row.op, ".") {
				rows = append(rows, row)
			}
		}
		return rows
	})
}

func TestX86PackedNarrowNativeInputsStayIndependent(t *testing.T) {
	x86FixedBinaryNativeInputBoundaries(t, x86NarrowValueRows)
}

func TestX86PackedNarrowResultEveryOutputByteKeepsExactDependencies(t *testing.T) {
	for _, arch := range []string{"amd64", "386"} {
		for _, row := range x86NarrowValueRows(arch) {
			if strings.Contains(row.op, ".") {
				continue // Broadcast's complete memory tuple has its own matrix.
			}
			op, operands := row.op, "X0,X1"
			legacy := !strings.HasPrefix(op, "V")
			if !legacy {
				operands = row.bank + "0," + row.bank + "1,"
				if row.mode != "plain" {
					operands += "K1,"
				}
				operands += row.bank + "2"
				if row.mode == "zero" {
					op += ".Z"
				}
			}
			t.Run(arch+"/"+op+" "+operands, func(t *testing.T) {
				ins := x86ScalarLaneInstruction(t, op+" "+operands)
				width, destination := amd64VectorByteWidth(Reg(row.bank+"0")), 2
				if legacy {
					destination = 1
				}
				inputBytes := 2
				if strings.HasSuffix(row.op, "DW") || row.op == "PACKSSLW" {
					inputBytes = 4
				}
				for _, first := range []uint64{0, 1, 3, 15, 0xffff, 0xff00ff00ff00ff00, ^uint64(0)} {
					for _, second := range []uint64{0, 7, 0xffff, 0x5555555555555555, ^uint64(0)} {
						old := uint64(0xaaaaaaaaaaaaaaaa)
						state := x86ResultState{flags: x86ByteFlagsAll}
						for index, bits := range []uint64{first, second, old} {
							state.xmm[index], state.vectorUpper[index] = uint16(bits), bits&^0xffff
						}
						state.masks[1] = true
						if legacy {
							old = second
						}
						// An output element depends on every byte of exactly one
						// signed input element, independently in each 128-bit lane.
						var want uint64
						for output := 0; output < width; output += inputBytes / 2 {
							chunk, local := output/16*16, output%16
							input, source := chunk+(local%8)/(inputBytes/2)*inputBytes, second
							if local >= 8 {
								source = first
							}
							needed := uint64(1<<uint(inputBytes)-1) << uint(input)
							if source&needed == needed {
								want |= uint64(1<<uint(inputBytes/2)-1) << uint(output)
							}
						}
						if row.mode == "merge" {
							want &= old
						}
						active := x86CallerVectorByteMask(int64(width))
						if legacy {
							want |= old &^ active
						} else {
							want |= ^active
						}
						if !x86ResultTransfer(ins, &state, nil, 0, arch) {
							t.Fatal("complete typed narrow has no result dependency consumer")
						}
						if got := state.vectorBytes(destination); got != want || state.flags != x86ByteFlagsAll {
							t.Fatalf("first=%x second=%x got=%x want=%x flags=%x", first, second, got, want, state.flags)
						}
					}
				}
			})
		}
	}
}
