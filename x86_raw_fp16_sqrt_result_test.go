package plan9asm

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

// Intel FP16 sections 5.74/5.75: independent packed/scalar, VL/LLIG,
// memory/broadcast, merge/zero and all four embedded rounding controls.
// https://cdrdv2-public.intel.com/678970/intel-avx512-fp16.pdf
func x86FP16SqrtResultEncoding(scalar bool, length byte, memory, control, zero bool) []byte {
	code := rawFP16PackedSqrt(length, 1, 0, 1, memory, control, zero)
	if scalar {
		code = encodeRawFP16ScalarBinary(0x51, length, 2, 1, 0, 1, memory, control, zero)
	}
	if memory {
		code[len(code)-1] = 0
	}
	return code
}

func TestX86RawFP16SqrtInitializedNativeObjects(t *testing.T) {
	llc := findLLVM22Tool("llc")
	if llc == "" {
		t.Fatal("LLVM22 llc required")
	}
	for _, target := range x86RawReturnTargets {
		t.Run(target.triple, func(t *testing.T) {
			word, move := 8, "MOVQ"
			if target.arch == "386" {
				word, move = 4, "MOVL"
			}
			var source, declarations strings.Builder
			count := 0
			for _, scalar := range []bool{false, true} {
				for length := byte(0); length < 4; length++ {
					for _, memory := range []bool{false, true} {
						for _, control := range []bool{false, true} {
							if !scalar && length == 3 && (memory || !control) || scalar && memory && control {
								continue
							}
							for _, zero := range []bool{false, true} {
								name := fmt.Sprintf("halfSqrt%d", count)
								fmt.Fprintf(&source, "TEXT %s(SB),4,$0-%d\n%s p+0(FP),AX\n", name, word+4, move)
								source.WriteString("MOVL $1,CX\nKMOVD CX,K1\nVPXORD Z0,Z0,Z0\nVPXORD Z1,Z1,Z1\nVPXORD Z2,Z2,Z2\n")
								source.WriteString(x86RawDisplacementDirectives(x86FP16SqrtResultEncoding(scalar, length, memory, control, zero)))
								fmt.Fprintf(&source, "MOVL X0,DX\nMOVL DX,ret+%d(FP)\nRET\n", word)
								fmt.Fprintf(&declarations, "func %s(p *byte) uint32\n", name)
								count++
							}
						}
					}
				}
			}
			if count != 50 {
				t.Fatal("incomplete independent FP16 SQRT physical forms", count)
			}
			x86LEAActualTargetGoObject(t, target.arch, target.triple, source.String())
			for _, api := range []string{"text", "module", "Go"} {
				ir, err := x86ScalarFPBitViewAPI(t, source.String(), declarations.String(), target.arch, target.triple, api)
				if err != nil {
					t.Fatal("initialized raw FP16 SQRT family", api, err)
				}
				for _, text := range []string{ir, x86DwordProductIROptimize(t, ir)} {
					if strings.Count(text, "asm sideeffect") != count || !strings.Contains(text, "vsqrtph ") || !strings.Contains(text, "vsqrtsh ") {
						t.Fatal("FP16 SQRT lost its native rounding/exception effect", api)
					}
					for _, level := range []string{"-O0", "-O2"} {
						x86PackedStringCompileObject(t, llc, target.triple, text, level)
					}
				}
			}
		})
	}
}

func TestX86RawFP16SqrtCannotInventInputs(t *testing.T) {
	for _, arch := range []string{"386", "amd64"} {
		for _, scalar := range []bool{false, true} {
			for _, memory := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/scalar%t/memory%t", arch, scalar, memory), func(t *testing.T) {
					code := x86RawDisplacementDirectives(x86FP16SqrtResultEncoding(scalar, 0, memory, false, false))
					bare := "TEXT Value(SB),4,$0-0\n" + code + "RET\n"
					for _, api := range []string{"text", "module", "Go"} {
						ir, err := x86ScalarFPBitViewAPI(t, bare, "func Value()", arch, x86AddressMemoryTargets(arch)[0], api)
						if ir != "" || !errors.Is(err, ErrProbeNeedsContext) {
							t.Fatal("Void FP16 SQRT borrowed unknown input or mask", api, err)
						}
					}
				})
			}
		}
	}
}
