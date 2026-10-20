package main

import (
	"fmt"
	"strings"
	"testing"

	"github.com/xgo-dev/plan9asm"
)

func TestInLanePermuteCoverageRetainsActualInputsAndVoidReads(t *testing.T) {
	for _, arch := range []string{"amd64", "386"} {
		for _, op := range []string{"VPERMILPS", "VPERMILPD"} {
			for _, bank := range []string{"X", "Y", "Z"} {
				for _, immediate := range []bool{false, true} {
					t.Run(fmt.Sprintf("%s/%s/%s/immediate=%t", arch, op, bank, immediate), func(t *testing.T) {
						zero := "VPXOR"
						if bank == "Z" {
							zero = "VPXORD"
						}
						var prefix strings.Builder
						for register := 0; register < 3; register++ {
							fmt.Fprintf(&prefix, "%s %s%d,%s%d,%s%d\n", zero, bank, register, bank, register, bank, register)
						}
						operands, memory := bank+"0,"+bank+"1", "(AX),"+bank+"1"
						if immediate {
							operands, memory = "$0x1b,"+bank+"1", "$0x1b,(AX)"
						}
						row := op + " " + operands + "," + bank + "2"
						// Observing the low result distinguishes unknown-byte
						// propagation from an initialized numerical value.
						observed := row + "\nMOVL X2,AX"
						valueProbeNeedsContext(t, arch, observed)
						if err := plan9asm.ProbeInstructionSequence(plan9asm.ArchAMD64, arch,
							compactProbeInstructions(t, prefix.String()+observed)); err != nil {
							t.Fatal("complete register data/selector input rejected", err)
						}
						// A numeric zero is not a native pointer, even when the
						// result is dead. Neither mask=0 nor broadcast repairs it.
						native := op + " " + memory + "," + bank + "2"
						valueProbeNeedsContext(t, arch, prefix.String()+native)
						valueProbeNeedsContext(t, arch, prefix.String()+"MOVL $0,AX\n"+native)
						if arch == "amd64" {
							masked := op + ".BCST.Z " + memory + ",K1," + bank + "2"
							valueProbeNeedsContext(t, arch, prefix.String()+"MOVL $0,AX\nKMOVD AX,K1\n"+masked)
						}
					})
				}
			}
		}
	}
}
