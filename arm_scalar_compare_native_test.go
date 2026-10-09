package plan9asm

import (
	"errors"
	"strings"
	"testing"
)

func TestARMScalarCompareNativeEffects(t *testing.T) {
	for _, body := range []string{"CMPF F0", "CMPD F0,F1", "WORD $0xeeb50a40", "WORD $0xeeb50ac0"} {
		source := "TEXT conversion(SB),4,$0-0\nMOVD $1.5,F0\nMOVD $2.5,F1\n" + body + "\nRET\n"
		file, err := Parse(ArchARM, source)
		if err != nil {
			t.Fatal(err)
		}
		ir, err := Translate(file, Options{Goarch: "arm", TargetTriple: "armv7-unknown-linux-gnueabihf",
			Sigs: map[string]FuncSig{"conversion": {Name: "conversion", Ret: Void}},
		})
		if err != nil {
			t.Fatal(err)
		}
		operation := "vcmpe."
		if body == "WORD $0xeeb50a40" {
			operation = "vcmp."
		}
		if strings.Contains(ir, "fcmp ") || !strings.Contains(ir, "asm sideeffect") || !strings.Contains(ir, operation) {
			t.Errorf("%s must retain the native signaling/FPSCR effect:\n%s", body, ir)
		}
	}
}

func TestARMScalarCompareNativeTargetBoundary(t *testing.T) {
	for _, api := range []string{"text", "module", "Go"} {
		for _, body := range []string{"CMPF F0", "CMPD F0,F1", "WORD $0xeeb50a40", "WORD $0xeeb50bc0"} {
			for _, triple := range []string{"", "aarch64-unknown-linux-gnu", "x86_64-unknown-linux-gnu", "armeb-unknown-linux-gnueabi", "armanything-unknown-linux-gnu"} {
				source := "TEXT conversion(SB),4,$0-0\n" + body + "\nRET\n"
				if err := armWordFloatBoundaryAPI(t, api, source, triple); !errors.Is(err, ErrProbeNeedsContext) {
					t.Errorf("%s %s %q needs native FPSCR target: %v", api, body, triple, err)
				}
			}
			for _, invalid := range []string{"CMPF R0", "CMPD.S F0", "WORD $0xffffffff"} {
				for _, dead := range []string{"", "RET\n"} {
					source := "TEXT conversion(SB),4,$0-0\n" + body + "\n" + dead + invalid + "\nRET\n"
					if err := armWordFloatBoundaryAPI(t, api, source, ""); err == nil || errors.Is(err, ErrProbeNeedsContext) {
						t.Errorf("%s invalid source cannot become target Context: %v", api, err)
					}
				}
			}
		}
	}
}
