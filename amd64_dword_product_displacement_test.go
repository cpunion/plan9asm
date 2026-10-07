package plan9asm

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

type x86DwordProductDisplacementOracle struct {
	row   string
	width int
	tuple int
	units int
	mask  bool
	high  bool
	mem   MemRef
	code  []byte
}

// These tuples and offsets are independently prescribed from the actual
// Go encoder's Full/N16,N32,N64 and BcstN8 rows, not production properties.
func x86DwordProductDisplacementOracles(t *testing.T) []x86DwordProductDisplacementOracle {
	t.Helper()
	var result []x86DwordProductDisplacementOracle
	addresses := []MemRef{
		{Base: AX}, {Base: BP}, {Base: SP, Index: CX, Scale: 2},
		{Base: AX, Index: CX, Scale: 4}, {Base: "R13", Index: "R12", Scale: 8},
	}
	for _, name := range []string{"VPMULDQ", "VPMULUDQ"} {
		for _, width := range []int{16, 32, 64} {
			for _, broadcast := range []bool{false, true} {
				for masking := 0; masking < 3; masking++ {
					for _, units := range []int{-129, -128, -1, 1, 127, 128} {
						for _, memory := range addresses {
							tuple, op := width, name
							if broadcast {
								tuple, op = 8, name+".BCST"
							}
							if masking == 2 {
								op += ".Z"
							}
							memory.Off = int64(units * tuple)
							bank := map[int]string{16: "X", 32: "Y", 64: "Z"}[width]
							second, destination := bank+"1", bank+"2"
							high := masking == 0 && !broadcast && width != 64
							if high {
								second, destination = bank+"17", bank+"18" // Force actual EVEX.
							}
							args := (Operand{Kind: OpMem, Mem: memory}).String() + "," + second
							if masking == 1 {
								args += ",K1"
							} else if masking == 2 {
								args += ",K7"
							}
							row := op + " " + args + "," + destination
							code := x86OrderingHintGoBytes(t, "amd64", "TEXT probe(SB),4,$0-0\n"+row+"\nRET\n")
							if len(code) < 8 || code[0] != 0x62 || code[2]&0x80 == 0 || code[len(code)-1] != 0xc3 {
								t.Fatalf("independent original EVEX.W1 object changed: %s %x", row, code)
							}
							mod, end := code[5]>>6, 6
							if code[5]&7 == 4 {
								end++ // Independently observe SIB, not decoded source.
							}
							if units >= -128 && units <= 127 {
								if mod != 1 || int(int8(code[end])) != units || len(code) != end+2 {
									t.Fatalf("original compressed disp8/%d changed: %s %x", tuple, row, code)
								}
							} else if mod != 2 || len(code) != end+5 {
								t.Fatalf("original unscaled disp32 boundary changed: %s %x", row, code)
							}
							result = append(result, x86DwordProductDisplacementOracle{row, width, tuple, units, masking != 0, high, memory, code[:len(code)-1]})
						}
					}
				}
			}
		}
	}
	return result
}

func TestX86DwordProductCompressedDisp8IndependentPhysicalSource(t *testing.T) {
	fixtures := x86DwordProductDisplacementOracles(t)
	for _, fixture := range fixtures {
		for _, mode := range []int{32, 64} {
			for _, prefix := range []struct {
				byte    byte
				segment Reg
			}{{0, ""}, {0x64, FS}, {0x65, GS}} {
				t.Run(fmt.Sprintf("%s/mode%d/prefix%x", fixture.row, mode, prefix.byte), func(t *testing.T) {
					code := append([]byte(nil), fixture.code...)
					if prefix.byte != 0 {
						code = append([]byte{prefix.byte}, code...)
					}
					ins, length, known, err := decodedX86DwordToQwordMultiplyInstruction(code, mode)
					if mode == 32 && (fixture.high || fixture.mem.Base == "R13") {
						if !known || err == nil {
							t.Fatal("extended physical mode32 bank/address acquired permission", ins, err)
						}
						return
					}
					want := fixture.mem
					want.Segment = prefix.segment
					if !known || err != nil || length != len(code) || !ins.x86Encoded || len(ins.Args) == 0 || ins.Args[0].Kind != OpMem || !reflect.DeepEqual(ins.Args[0].Mem, want) {
						t.Fatalf("native physical tuple source changed: %x got=%+v length%d known%t err%v want=%+v", code, ins, length, known, err, want)
					}
				})
			}
			for _, bad := range [][]byte{
				append([]byte{0x67}, fixture.code...),
				append([]byte{0x64, 0x67}, fixture.code...),
				append([]byte{0x65, 0x67}, fixture.code...),
			} {
				if _, _, known, err := decodedX86DwordToQwordMultiplyInstruction(bad, mode); !known || err == nil {
					t.Fatal("address override acquired unbound source permission", fixture.row, mode, err)
				}
			}
		}
		wrongW := append([]byte(nil), fixture.code...)
		wrongW[2] &^= 0x80
		if _, _, known, err := decodedX86DwordToQwordMultiplyInstruction(wrongW, 64); !known || err == nil {
			t.Fatal("full/broadcast EVEX.W0 accepted", fixture.row, err)
		}
	}
	t.Logf("%d independently assembled full/broadcast,unmasked/merge/zero,disp8/disp32,SIB/bank witnesses ×two modes ×three segment prefixes; address/W rejection retained", len(fixtures))
}

func x86DwordProductDirectives(code []byte) string {
	var out strings.Builder
	for _, value := range code {
		fmt.Fprintf(&out, "BYTE $%d\n", value)
	}
	return out.String()
}
