package main

import (
	"strings"
	"testing"

	"github.com/xgo-dev/plan9asm"
)

func TestLogicalCompareCoverageRequiresNativeInputContext(t *testing.T) {
	names := strings.Fields("PCMPEQB PCMPEQW PCMPEQL PCMPEQQ PCMPGTB PCMPGTW PCMPGTL PCMPGTQ " +
		"VPCMPEQB VPCMPEQW VPCMPEQD VPCMPEQQ VPCMPGTB VPCMPGTW VPCMPGTD VPCMPGTQ " +
		"ANDPS ANDPD PAND VANDPS VANDPD VPAND VPANDD VPANDQ " +
		"ANDNPS ANDNPD PANDN VANDNPS VANDNPD VPANDN VPANDND VPANDNQ " +
		"ORPS ORPD POR VORPS VORPD VPOR VPORD VPORQ " +
		"XORPS XORPD PXOR VXORPS VXORPD VPXOR VPXORD VPXORQ")
	if len(names) != 48 {
		t.Fatal("complete independent logical/compare inventory changed", len(names))
	}
	for _, arch := range []string{"386", "amd64"} {
		for _, op := range names {
			t.Run(arch+"/"+op, func(t *testing.T) {
				memory, registers := op+" 3(AX)(DX*2),X1", op+" X0,X1"
				if strings.HasPrefix(op, "V") {
					memory, registers = op+" 3(AX)(DX*2),X1,X2", op+" X0,X1,X2"
				}
				valueProbeNeedsContext(t, arch, memory)
				if err := plan9asm.ProbeInstructionSequence(plan9asm.ArchAMD64, arch,
					compactProbeInstructions(t, "VZEROALL\n"+registers)); err != nil {
					t.Fatal("initialized registers became unsupported while checking native context", err)
				}
			})
		}
	}
}
