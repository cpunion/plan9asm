package plan9asm

import (
	"fmt"
	"testing"
)

// Independent of the production shift directory. Uniform counts select their
// low qword; variable counts have one complete unsigned integer per data lane.
func x86ArithmeticShiftTestNames() []string {
	return []string{"PSRAW", "PSRAL", "VPSRAW", "VPSRAD", "VPSRAQ", "VPSRAVW", "VPSRAVD", "VPSRAVQ"}
}

func x86ArithmeticShiftTestBits(name string) int {
	switch name[len(name)-1] {
	case 'W':
		return 16
	case 'Q':
		return 64
	default:
		return 32
	}
}

func TestX86ArithmeticShiftSharedTypedResultGrammar(t *testing.T) {
	for _, name := range x86ArithmeticShiftTestNames() {
		for _, bank := range []string{"X", "Y", "Z"} {
			legacy := name[0] != 'V'
			if legacy && bank != "X" {
				continue
			}
			perLane := len(name) > 5 && name[5] == 'V'
			count := "X0"
			if perLane {
				count = bank + "0"
			}
			row := name + " " + count + "," + bank + "1," + bank + "2"
			if legacy {
				row = name + " X0,X1"
			}
			ins := x86ScalarLaneInstruction(t, row)
			for _, arch := range []string{"386", "amd64"} {
				form, known, err := x86PackedShiftProperties(ins.Op, ins, arch)
				if !known || err != nil || form.laneBits != x86ArithmeticShiftTestBits(name) {
					t.Errorf("complete arithmetic family missing from shared result grammar: %s/%s: %v", arch, row, err)
					continue
				}
				state := x86ResultState{flags: x86ByteFlagsAll}
				state.writeVector(0, 64, ^uint64(0), false)
				state.writeVector(1, 64, ^uint64(0), false)
				if accepted, known := x86ResultPackedShiftStep(ins, &state, nil, arch); !known || !accepted {
					t.Errorf("initialized arithmetic result rejected: %s/%s", arch, row)
				}
			}
		}
	}
}

// Enumerate individual bits rather than borrowing the interval implementation.
// Oversized arithmetic counts select the sign bit, never a constant zero.
func TestX86ArithmeticShiftEverySignBitDependency(t *testing.T) {
	for _, name := range x86ArithmeticShiftTestNames() {
		legacy := name[0] != 'V'
		perLane := len(name) > 5 && name[5] == 'V'
		for _, width := range []int{16, 32, 64} {
			if legacy && width != 16 {
				continue
			}
			bank := map[int]string{16: "X", 32: "Y", 64: "Z"}[width]
			counts := []int{-1}
			if !perLane {
				counts = append(counts, 0, 1, 7, 8, 15, 16, 31, 32, 63, 64, 127, 255)
			}
			for _, count := range counts {
				if legacy && count == 255 {
					continue
				}
				countSource := "X0"
				if count >= 0 {
					countSource = fmt.Sprintf("$%d", count)
				} else if perLane {
					countSource = bank + "0"
				}
				row := name + " " + countSource + "," + bank + "1," + bank + "2"
				if legacy {
					row = name + " " + countSource + ",X1"
				}
				ins := x86ScalarLaneInstruction(t, row)
				form, known, err := x86PackedShiftProperties(ins.Op, ins, "amd64")
				if !known || err != nil {
					t.Errorf("arithmetic dependency grammar missing: %s: %v", row, err)
					continue
				}
				bits := x86ArithmeticShiftTestBits(name)
				for output := 0; output < width; output++ {
					var data, counter uint64
					for bit := output * 8; bit < output*8+8; bit++ {
						first, end := count, count+1
						if count < 0 {
							first, end = 0, bits+1
						}
						for shift := first; shift < end; shift++ {
							from := bit%bits + shift
							if from >= bits {
								from = bits - 1
							}
							data |= uint64(1) << uint((bit/bits*bits+from)/8)
						}
					}
					if count < 0 {
						counter = 255
						if perLane {
							counter = x86CallerVectorByteMask(int64(bits/8)) << uint(output/(bits/8)*(bits/8))
						}
					}
					gotData, gotCount := x86PackedShiftByteInputs(form, output)
					if gotData != data || gotCount != counter {
						t.Fatalf("%s byte%d: dependencies %x/%x want %x/%x", row, output, gotData, gotCount, data, counter)
					}
				}
			}
		}
	}
}
