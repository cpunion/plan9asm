package plan9asm

import (
	"errors"
	"fmt"
	"reflect"
	"testing"
)

func TestX86SelfCMPCompleteFlagsWithoutNumericalInputs(t *testing.T) {
	for _, arch := range []string{"386", "amd64"} {
		widths := "BWLQ"
		if arch == "386" {
			widths = "BWL"
		}
		for _, width := range widths {
			registers := []string{"AX", "CX", "DX", "BX", "SI", "DI"}
			if arch == "386" {
				registers = append(registers, "BP")
			} else {
				registers = append(registers, "R8", "R9", "R10", "R11", "R12", "R13", "R14", "R15")
			}
			if width == 'B' {
				registers = []string{"AX", "CX", "DX", "BX", "AL", "CL", "DL", "BL", "AH", "CH", "DH", "BH"}
				if arch == "amd64" {
					registers = append(registers, "SI", "DI", "R8", "R9", "R10", "R11", "R12", "R13", "R14", "R15")
				}
			}
			for _, register := range registers {
				t.Run(fmt.Sprintf("%s/CMP%c/%s", arch, width, register), func(t *testing.T) {
					ins := x86ScalarLaneInstruction(t, fmt.Sprintf("CMP%c %s,%s", width, register, register))
					state := x86ResultState{direction: x86ResultDirectionBackward}
					state.pointerGP[0], state.callerAddressGP[0], state.callerAddressOrigin[0] = true, true, 5
					expected := state.clone()
					expected.flags = x86ByteFlagsAll
					if !x86ResultTransfer(ins, &state, nil, 0, arch) || !reflect.DeepEqual(state, expected) {
						t.Fatal("self comparison needs no source value and must define only the six status flags")
					}
				})
			}
		}
	}
}

func TestX86SelfCMPThreeAPIsKeepNumericalAndPhysicalBoundaries(t *testing.T) {
	for _, target := range x86RawReturnTargets {
		widths := "BWLQ"
		if target.arch == "386" {
			widths = "BWL"
		}
		for _, width := range widths {
			for _, row := range []struct {
				name, body string
				accepted   bool
			}{
				{"flags-only", fmt.Sprintf("CMP%c AX,AX\nMOVL $0,BX\nSETEQ BL\nMOVL BX,ret+RESULT(FP)", width), true},
				{"not-NUM", fmt.Sprintf("CMP%c AX,AX\nMOVL AX,ret+RESULT(FP)", width), false},
				{"different-inputs", fmt.Sprintf("CMP%c AX,BX\nMOVL $7,ret+RESULT(FP)", width), false},
				{"TEST-reads-input", fmt.Sprintf("TEST%c AX,AX\nMOVL $7,ret+RESULT(FP)", width), false},
				{"SP-still-physical", fmt.Sprintf("CMP%c SP,SP\nMOVL $7,ret+RESULT(FP)", width), false},
			} {
				for _, api := range []string{"text", "module", "Go"} {
					t.Run(fmt.Sprintf("%s/CMP%c/%s/%s", target.triple, width, row.name, api), func(t *testing.T) {
						source := x86ResultGPSIMDAdmissionSource(t, target.arch, "func view() uint32", row.body)
						ir, err := x86ScalarFPBitViewAPI(t, source, "func view() uint32", target.arch, target.triple, api)
						if row.accepted {
							if err != nil || ir == "" {
								t.Fatal("self-CMP status was not available", err)
							}
						} else if !errors.Is(err, ErrProbeNeedsContext) || ir != "" {
							t.Fatal("self comparison granted an input value or a different operand role", err)
						}
					})
				}
			}
		}
	}
}
