package plan9asm

import (
	"fmt"
	"os/exec"
	"reflect"
	"regexp"
	"strings"
	"testing"
)

func x86PackedShiftRawDecode(code []byte, mode int) (Instr, int, bool, error) {
	for _, decode := range []func([]byte, int) (Instr, int, bool, error){
		decodedX86PerLaneVariableShiftInstruction,
		decodedX86PackedUniformShiftInstruction,
		decodedX86RawVectorByteShiftInstruction,
	} {
		if ins, consumed, known, err := decode(code, mode); known {
			return ins, consumed, known, err
		}
	}
	return Instr{}, 0, false, nil
}

func TestX86PackedShiftRawCompletePhysicalLLVM22AndWIG(t *testing.T) {
	mc := findLLVM22Tool("llvm-mc")
	if mc == "" {
		t.Fatal("independent packed-shift WIG validation requires LLVM22 llvm-mc")
	}
	var rows []x86PackedShiftRawCase
	var input strings.Builder
	wantMnemonics := map[string]int{}
	for _, row := range x86PackedShiftRawCases() {
		values := []bool{false}
		if row.wig() {
			rows = append(rows, row)
			values = append(values, true)
		}
		for _, w := range values {
			for _, value := range row.code(w) {
				fmt.Fprintf(&input, "0x%02x ", value)
			}
			input.WriteByte('\n')
			wantMnemonics[strings.ToLower(row.member.op)]++
		}
	}
	if len(rows) != 172 || len(wantMnemonics) != 20 {
		t.Fatal("complete uniform-word/byte WIG directory changed", len(rows), len(wantMnemonics))
	}
	for _, mode := range []int{32, 64} {
		triple := map[int]string{32: "i386-unknown-linux-gnu", 64: "x86_64-unknown-linux-gnu"}[mode]
		t.Run(triple, func(t *testing.T) {
			// An external decoder must recognize every exact W0/W1 byte first.
			// Go choosing W0 is not evidence that W1 is an invalid instruction.
			stage, cancel := runtimeOracleStageContext(t)
			command := exec.CommandContext(stage, mc, "--disassemble", "--triple="+triple,
				"--mattr=+avx,+avx2,+avx512f,+avx512bw,+avx512vl")
			command.Stdin = strings.NewReader(input.String())
			output, err := runtimeOracleCombinedOutput(stage, command)
			cancel()
			if err != nil || strings.Contains(string(output), "warning") || strings.Contains(string(output), "invalid") {
				t.Fatalf("independent LLVM22 rejected complete WIG source bytes: %v\n%s", err, output)
			}
			for mnemonic, want := range wantMnemonics {
				if got := len(regexp.MustCompile(`(?m)\b`+mnemonic+`[ \t]+`).FindAll(output, -1)); got != want {
					t.Fatalf("independent LLVM22 WIG row count for %s: %d != %d\n%s", mnemonic, got, want, output)
				}
			}
			for _, row := range rows {
				t.Run(row.name(), func(t *testing.T) {
					var control Instr
					for _, w := range []bool{false, true} {
						code := row.code(w)
						ins, consumed, known, err := x86PackedShiftRawDecode(code, mode)
						if !known || err != nil || consumed != len(code) {
							t.Fatal("ISA WIG does not close the complete typed shift grammar", w, known, consumed, err)
						}
						if !w {
							control = ins
						} else if ins.Op != control.Op || !reflect.DeepEqual(ins.Args, control.Args) {
							t.Fatal("ignored W changed the original shift's semantic operands", ins.Raw, control.Raw)
						}
					}
				})
			}
		})
	}
}

func TestX86PackedShiftRawPhysicalRestrictionsStayLocal(t *testing.T) {
	for _, row := range x86PackedShiftRawCases() {
		if row.input != "register" || row.mask != 0 {
			continue
		}
		t.Run(row.name(), func(t *testing.T) {
			valid := row.code(false)
			if _, consumed, known, err := x86PackedShiftRawDecode(valid, 32); !known || err != nil || consumed != len(valid) {
				t.Fatal("legitimate low-bank mode32 physical shift rejected", err)
			}
			controls := map[string][]byte{
				"address-override": append([]byte{0x67}, valid...),
				"truncated":        append([]byte(nil), valid[:len(valid)-1]...),
			}
			if row.evex {
				for _, name := range []string{"fixed-bit", "reserved-length", "zero-without-mask", "physical-register8"} {
					code := append([]byte(nil), valid...)
					switch name {
					case "fixed-bit":
						code[2] &^= 4
					case "reserved-length":
						code[3] |= 0x60
					case "zero-without-mask":
						code[3] |= 0x80
					case "physical-register8":
						code[1] &^= 0x20 // ModRM.r/m is physically register8.
					}
					controls[name] = code
				}
			} else {
				code := append([]byte(nil), valid...)
				code[1] &^= 0x20
				controls["physical-register8"] = code
			}
			if row.member.mode == "lane" && row.member.bits == 16 {
				code := append([]byte(nil), valid...)
				code[2] &^= 0x80
				controls["per-lane-word-W0"] = code
			}
			if row.member.mode == "bytes" && row.evex {
				code := append([]byte(nil), valid...)
				code[3] |= 1
				controls["byte-shift-mask"] = code
			}
			for name, code := range controls {
				if _, _, known, err := x86PackedShiftRawDecode(code, 32); known && err == nil {
					t.Fatal("malformed or unavailable physical shift became valid", name, code)
				}
			}
		})
	}
}
