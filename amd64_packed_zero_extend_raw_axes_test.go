package plan9asm

import (
	"fmt"
	"os/exec"
	"strings"
	"testing"
)

func TestX86PackedExtendRawWBitMatchesIndependentLLVMDecoder(t *testing.T) {
	mc := findLLVM22Tool("llvm-mc")
	if mc == "" {
		t.Fatal("LLVM22 llvm-mc required")
	}
	for _, spec := range rawPackedExtendTestSpecs {
		for _, evex := range []bool{false, true} {
			for width := 0; width < 3; width++ {
				if !evex && width == 2 {
					continue
				}
				code := encodeX86VEXPackedExtend(spec.opcode, width != 0, 1, 0)
				code[2] |= 0x80 // All VEX forms are WIG, not W0.
				if evex {
					code = encodeX86EVEXPackedExtend(spec.opcode, byte(width), 1, true, 1, 0)
					code[2] |= 0x80 // EVEX is WIG except signed/unsigned DQ.
				}
				var bytes strings.Builder
				for _, value := range code {
					fmt.Fprintf(&bytes, "0x%02x ", value)
				}
				ctx, cancel := runtimeOracleStageContext(t)
				command := exec.CommandContext(ctx, mc, "--disassemble", "--triple=x86_64", "--mattr=+avx,+avx2,+avx512f,+avx512bw,+avx512vl")
				command.Stdin = strings.NewReader(bytes.String() + "\n")
				output, err := runtimeOracleCombinedOutput(ctx, command)
				cancel()
				invalid := evex && spec.inputBits == 32
				if invalid {
					if !strings.Contains(string(output), "invalid instruction encoding") {
						t.Fatalf("LLVM22 did not reject EVEX DQ W1 %x: %v\n%s", code, err, output)
					}
				} else if err != nil || !strings.Contains(strings.ToUpper(string(output)), string(spec.op)) {
					t.Fatalf("LLVM22 independent WIG decode %x: %v\n%s", code, err, output)
				}
				for _, mode := range []int{32, 64} {
					ins, length, recognized, err := decodedX86PackedExtendInstruction(code, mode)
					if invalid {
						if !recognized || err == nil {
							t.Fatalf("mode%d illegal EVEX DQ W1 acquired widening semantics: %x %+v/%v", mode, code, ins, err)
						}
					} else if !recognized || err != nil || length != len(code) || strings.TrimSuffix(string(ins.Op), ".Z") != string(spec.op) {
						t.Fatalf("mode%d legal WIG %x lost exact family semantics: %+v/%v", mode, code, ins, err)
					}
				}
			}
		}
	}
}

func TestX86PackedZeroExtendActualGoLowMaskedMode32(t *testing.T) {
	for _, ratio := range x86ZeroExtendTestRatios {
		for _, bank := range []string{"X", "Y", "Z"} {
			source := "X0"
			if bank == "Z" && ratio.output == 2*ratio.input {
				source = "Y0"
			}
			for _, input := range []string{source, "(AX)"} {
				for _, suffix := range []string{"", ".Z"} {
					row := fmt.Sprintf("VPMOVZX%s%s %s,K7,%s7", ratio.suffix, suffix, input, bank)
					listing, err := x86ExtensionGoListing(t, "386", row)
					if err != nil {
						t.Fatalf("actual mode32 Go acceptance: %s: %v\n%s", row, err, listing)
					}
					file, err := Parse(ArchAMD64, "TEXT f(SB),4,$0-0\n"+row+"\nRET\n")
					if err != nil {
						t.Fatal(err)
					}
					if _, err := Translate(file, Options{Goarch: "386", TargetTriple: "i386-unknown-linux-gnu", Sigs: map[string]FuncSig{"f": {Ret: Void}}}); err != nil {
						t.Fatalf("actual Go low-bank mask form incorrectly rejected: %s: %v", row, err)
					}
				}
			}
		}
	}
}
