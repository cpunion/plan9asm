package plan9asm

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

func x86SegmentMOVRegisters(arch string, width int) []Reg {
	registers := []Reg{AX, CX, DX, BX, SP, BP, SI, DI}
	if arch == "amd64" {
		registers = append(registers, R8, R9, R10, R11, R12, R13, R14, R15)
	}
	if width == 1 {
		registers = append(registers, AL, CL, DL, BL, AH, CH, DH, BH)
		if arch == "386" {
			// Go synthesizes BP/SI/DI byte rows but rejects direct SP.
			registers = append(append([]Reg{}, registers[:4]...), registers[5:]...)
		} else {
			registers = append(registers, BPB, SIB, DIB, R8B, R9B, R10B, R11B, R12B, R13B, R14B, R15B)
		}
	}
	return registers
}

func TestX86SegmentMOVCompleteGoGrammarAndAliases(t *testing.T) {
	// This existing independent oracle pins every row of all four actual
	// ymov tables, including the distinct historical signed-imm32 table.
	TestX86GPSIMDTransferCompleteLegacyOriginalTables(t)
	x86OriginalMOVQAliases(t)
	source, err := os.ReadFile(filepath.Join(runtime.GOROOT(), "src/cmd/internal/obj/x86/asm6.go"))
	if err != nil {
		t.Fatal(err)
	}
	for _, table := range []struct{ name, row string }{
		{"ymovb", "{Zm_r,1,argList{Ymb,Yrb}}"},
		{"ymovw", "{Zm_r,1,argList{Yml,Yrl}}"},
		{"ymovl", "{Zm_r,1,argList{Yml,Yrl}}"},
		{"ymovq", "{Zm_r,1,argList{Yml,Yrl}}"},
	} {
		definition := regexp.MustCompile(`(?s)var ` + table.name + ` = \[\]ytab\{(.*?)\n\}`).FindStringSubmatch(string(source))
		if len(definition) != 2 {
			t.Fatal("missing complete actual MOV table", table.name)
		}
		rows := regexp.MustCompile(`//[^\n]*`).ReplaceAllString(definition[1], "")
		rows = strings.Join(strings.Fields(rows), "")
		if strings.Count(rows, table.row) != 1 {
			t.Fatal("actual native memory-to-GP row is missing or duplicated", table.name)
		}
	}
	for _, arch := range []string{"386", "amd64"} {
		for _, row := range x86SegmentMOVRows(arch) {
			for _, segment := range []string{"FS", "GS"} {
				for _, address := range []string{"fixed", "fixed-zero", "base", "index"} {
					_, memory := x86SegmentMOVAddress(arch, segment, address)
					for _, destination := range x86SegmentMOVRegisters(arch, row.width) {
						t.Run(fmt.Sprintf("%s/%s/%s/%s/%s", arch, row.op, segment, address, destination), func(t *testing.T) {
							ins := x86ScalarLaneInstruction(t, row.op+" "+memory+","+string(destination))
							original := ins
							original.Args = append([]Operand(nil), ins.Args...)
							form, known, err := x86ScalarNativeLoadProperties(ins.Op, ins, arch)
							if !known || err != nil || form.bytes != int64(row.width) || form.destination != destination || !reflect.DeepEqual(form.source, ins.Args[0]) || !reflect.DeepEqual(ins, original) {
								t.Fatal("complete typed segment MOV row/alias changed", form, known, err)
							}
						})
					}
				}
				for _, mutation := range []string{"OffRaw", "IndexExt", "scale", "wide-offset", "unknown-segment"} {
					t.Run(arch+"/"+row.op+"/"+segment+"/invalid-"+mutation, func(t *testing.T) {
						ins := x86ScalarLaneInstruction(t, row.op+" 24("+segment+"),DX")
						switch mutation {
						case "OffRaw":
							ins.Args[0].Mem.OffRaw = "unresolved"
						case "IndexExt":
							ins.Args[0].Mem.Index, ins.Args[0].Mem.IndexExt, ins.Args[0].Mem.Scale = CX, "UXTW", 4
						case "scale":
							ins.Args[0].Mem.Index, ins.Args[0].Mem.Scale = CX, 3
						case "wide-offset":
							ins.Args[0].Mem.Off = 1 << 32
						case "unknown-segment":
							ins.Args[0].Mem.Segment = "CS"
						}
						if _, known, err := x86ScalarNativeLoadProperties(ins.Op, ins, arch); known && err == nil {
							t.Fatal("typed source grammar accepted a malformed segment memory operand", mutation)
						}
					})
				}
			}
		}
	}
	for _, alias := range []Reg{AH, CH, DH, BH} {
		want := map[Reg]Reg{AH: SP, CH: BP, DH: SI, BH: DI}[alias]
		ins := x86ScalarLaneInstruction(t, "MOVB -8(R8)(FS),"+string(alias))
		form, known, err := x86ScalarNativeLoadProperties(ins.Op, ins, "amd64")
		if !known || err != nil || form.destination != want {
			t.Fatal("Go's actual REX byte-register number was not preserved", alias, form, err)
		}
	}
}

func TestX86SegmentMOVOriginalGoEncodingWitnesses(t *testing.T) {
	for _, arch := range []string{"386", "amd64"} {
		for _, segment := range []string{"FS", "GS"} {
			prefix := byte(0x64)
			if segment == "GS" {
				prefix = 0x65
			}
			for _, row := range x86SegmentMOVRows(arch) {
				t.Run(arch+"/"+row.op+"/"+segment+"/fixed-bytes", func(t *testing.T) {
					want := []byte{prefix}
					if row.width == 2 {
						want = append(want, 0x66)
					} else if row.width == 8 {
						want = append(want, 0x48)
					}
					opcode := byte(0x8b)
					if row.width == 1 {
						opcode = 0x8a
					}
					if arch == "386" {
						want = append(want, opcode, 0x0d, 24, 0, 0, 0, 0xc3)
					} else {
						want = append(want, opcode, 0x0c, 0x25, 24, 0, 0, 0, 0xc3)
					}
					source := "TEXT original(SB),4,$0-0\n" + row.op + " 24(" + segment + "),CX\nRET\n"
					if got := x86OrderingHintGoBytes(t, arch, source); !bytes.Equal(got, want) {
						t.Fatalf("actual full original byte witness differs: %x != %x", got, want)
					}
				})
			}
			for _, witness := range []struct {
				name, instruction string
				bytes             []byte
			}{
				{"base", "MOVB -8(BX)(" + segment + "),CX", []byte{prefix, 0x8a, 0x4b, 0xf8, 0xc3}},
				{"index", "MOVL -8(" + segment + ")(CX*4),DX", []byte{prefix, 0x8b, 0x14, 0x8d, 0xf8, 0xff, 0xff, 0xff, 0xc3}},
			} {
				t.Run(arch+"/"+segment+"/"+witness.name+"-bytes", func(t *testing.T) {
					got := x86OrderingHintGoBytes(t, arch, "TEXT original(SB),4,$0-0\n"+witness.instruction+"\nRET\n")
					if !bytes.Equal(got, witness.bytes) {
						t.Fatalf("actual complete dynamic-address bytes differ: %x != %x", got, witness.bytes)
					}
				})
			}
			if arch == "amd64" {
				for index, alias := range []Reg{AH, CH, DH, BH} {
					t.Run(arch+"/"+segment+"/REX-"+string(alias), func(t *testing.T) {
						source := "TEXT original(SB),4,$0-0\nMOVB -8(R8)(" + segment + ")," + string(alias) + "\nRET\n"
						want := []byte{prefix, 0x41, 0x8a, byte(0x60 + 8*index), 0xf8, 0xc3}
						if got := x86OrderingHintGoBytes(t, arch, source); !bytes.Equal(got, want) {
							t.Fatalf("actual REX high-byte alias differs: %x != %x", got, want)
						}
					})
				}
			}
		}
		t.Run(arch+"/complete-register-address-banks", func(t *testing.T) {
			var source strings.Builder
			count := 0
			for _, row := range x86SegmentMOVRows(arch) {
				for _, segment := range []string{"FS", "GS"} {
					for _, address := range []string{"fixed", "fixed-zero", "base", "index"} {
						_, memory := x86SegmentMOVAddress(arch, segment, address)
						for _, destination := range x86SegmentMOVRegisters(arch, row.width) {
							fmt.Fprintf(&source, "TEXT original%d(SB),4,$0-0\n%s %s,%s\nRET\n", count, row.op, memory, destination)
							count++
						}
					}
				}
			}
			triple := "x86_64-unknown-linux-gnu"
			want := 920
			if arch == "386" {
				triple, want = "i386-unknown-linux-gnu", 248
			}
			if count != want || strings.Count(source.String(), "TEXT ") != want {
				t.Fatal("independent complete grammar denominator changed", count, want)
			}
			x86LEAActualTargetGoObject(t, arch, triple, source.String())
		})
	}
	for _, op := range []string{"MOVQ", "MOVD", "MOVDQ2Q"} {
		t.Run("386/no-GP-qword/"+op, func(t *testing.T) {
			requireLEANativeGoObject(t, "386", "TEXT invalid(SB),4,$0-0\n"+op+" 24(FS),AX\nRET\n", false)
		})
	}
}
