package plan9asm

import (
	"errors"
	"strings"
	"testing"
)

func TestX86COMIModeContextCannotHideMalformedSource(t *testing.T) {
	for _, name := range []string{"VCOMISS", "VCOMISD", "VUCOMISS", "VUCOMISD"} {
		for _, body := range []string{
			name + " X8,Y1\n",
			name + " X8,X1\nRET\n" + name + " X0,Y1\n",
			name + " X8,X1\nRET\nTEXT ·later(SB),4,$0-0\n" + name + " X0,Y1\n",
		} {
			for _, api := range []string{"text", "module", "Go"} {
				t.Run(name+"/"+api+"/"+body, func(t *testing.T) {
					_, err := x86ScalarFPBitViewAPI(t, "TEXT ·view(SB),4,$0-0\n"+body+"RET\n",
						"func view()\nfunc later()\n", "386", "i386-unknown-linux-gnu", api)
					if err == nil || errors.Is(err, ErrProbeNeedsContext) {
						t.Fatalf("mode32 context hid malformed original source: %v", err)
					}
				})
			}
		}
	}
}

func TestX86COMIDirectLowererRequiresNativeTarget(t *testing.T) {
	for _, triple := range []string{"", "aarch64-unknown-linux-gnu", "i386-unknown-linux-gnu"} {
		var output strings.Builder
		ctx := amd64Ctx{b: &output, goarch: "amd64", targetTriple: triple}
		instruction := x86ScalarLaneInstruction(t, "COMISS X0,X1")
		known, terminated, err := ctx.lowerVectorScalarFlagCompare(instruction.Op, instruction)
		if !known || terminated || !errors.Is(err, ErrProbeNeedsContext) || output.Len() != 0 {
			t.Fatal("direct helper emitted without native target", triple, known, terminated, err, output.String())
		}
	}
}

func TestX86COMIRawVEXLengthIsNotPortableIgnoredBit(t *testing.T) {
	for _, name := range []string{"VCOMISS", "VCOMISD", "VUCOMISS", "VUCOMISD"} {
		code := x86PackedShiftActualGoBytes(t, "amd64", "TEXT probe(SB),4,$0-0\n"+name+" X0,X1\nRET\n")
		if len(code) != 5 || code[0] != 0xc5 || code[1]&4 != 0 {
			t.Fatalf("actual Go must retain its VEX.L0 encoding: %s %x", name, code)
		}
		code[1] |= 4
		for _, mode := range []int{32, 64} {
			if _, _, matched, err := decodedX86ScalarFlagCompareInstruction(code[:4], mode); !matched || !errors.Is(err, ErrProbeNeedsContext) {
				t.Fatalf("raw VEX.L1 normalized to L0 without physical contract: %s mode%d %x %v", name, mode, code, err)
			}
		}
	}
}
