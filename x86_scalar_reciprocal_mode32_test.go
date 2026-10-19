package plan9asm

import (
	"errors"
	"testing"
)

func TestX86ScalarReciprocalMode32FrontendAddressHasNoPhysicalPermission(t *testing.T) {
	for _, op := range []string{"VRCPSS", "VRSQRTSS"} {
		for _, memory := range []string{"16(R9)", "16(AX)(R9*2)", "16(R9)(CX*2)"} {
			t.Run(op+"/"+memory, func(t *testing.T) {
				source := "TEXT f(SB),4,$0-0\n" + op + " " + memory + ",X1,X2\nRET\n"
				goAssemblerOracleOutput(t, "386", source, "", true)
				ins := x86ScalarLaneInstruction(t, op+" "+memory+",X1,X2")
				if _, known, err := x86ScalarReciprocalProperties(ins.Op, ins, "386"); !known || !errors.Is(err, ErrProbeNeedsContext) {
					t.Fatal("Go-accepted VEX address was misclassified as rejected syntax or granted a physical mode32 binding", known, err)
				}
				state := x86ResultState{}
				state.xmm[1], state.xmm[2] = 0xffff, 0xffff
				for index := range state.gp {
					state.gp[index], state.pointerGP[index] = 0xff, true
				}
				if accepted, known := x86ResultScalarReciprocalStep(ins, &state, nil, "386", true); !known || accepted {
					t.Fatal("frontend extended address acquired numerical or native permission", accepted, known)
				}
			})
		}
	}
}
