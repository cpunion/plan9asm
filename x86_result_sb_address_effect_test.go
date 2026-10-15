package plan9asm

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

func x86LEASBUnknownAddressSource(t *testing.T, arch, declaration, body string) string {
	t.Helper()
	return "DATA ·table<>+0(SB)/8,$7\nGLOBL ·table<>(SB),16,$16\n" +
		x86ResultGPSIMDAdmissionSource(t, arch, declaration, body)
}

func TestX86LEASBUnknownAddressCompleteThreeAPIs(t *testing.T) {
	for _, target := range x86RawReturnTargets {
		widths := []string{"W", "L"}
		registers := []string{"AX", "CX", "DX", "BX", "SI", "DI", "BP"}
		if target.arch == "amd64" {
			widths = append(widths, "Q")
			registers = []string{"AX", "CX", "DX", "BX", "SI", "DI", "R8", "R9", "R10", "R11", "R12", "R13", "R15"}
		}
		for _, width := range widths {
			for _, register := range registers {
				for _, api := range []string{"text", "module", "Go"} {
					t.Run(target.triple+"/"+width+"/"+register+"/"+api, func(t *testing.T) {
						// Neither the address nor its uninitialized W remainder
						// contributes data to this independently defined result.
						body := fmt.Sprintf("MOVL $1,AX\nCMPL AX,$1\nLEA%s ·table<>-1(SB),%s\nSETEQ CL\nMOVL $7,ret+RESULT(FP)", width, register)
						source := x86LEASBUnknownAddressSource(t, target.arch, "func view() uint32", body)
						ir, err := x86ScalarFPBitViewAPI(t, source, "func view() uint32", target.arch, target.triple, api)
						if err != nil || ir == "" {
							t.Fatalf("discarded valid SB address blocked independent NUM result: %v", err)
						}
					})
				}
			}
		}
	}
}

func TestX86LEASBUnknownAddressDoesNotGrantFacts(t *testing.T) {
	for _, target := range x86RawReturnTargets {
		widths := []string{"W", "L"}
		move := "MOVL"
		if target.arch == "amd64" {
			widths, move = append(widths, "Q"), "MOVQ"
		}
		for _, width := range widths {
			for _, test := range []struct{ name, before, after string }{
				{"NUM-result", "", "MOVL DX,ret+RESULT(FP)"},
				{"NUM-compare", "", "TESTW DX,DX"},
				{"native-read", move + " p+0(FP),DX\n", "MOVB (DX),CL"},
				{"native-write", move + " p+0(FP),DX\n", "MOVB $1,(DX)"},
				{"caller-address", move + " p+0(FP),DX\n", "CMPW DX,$0"},
				{"old-result-address", "LEA" + strings.TrimPrefix(move, "MOV") + " ret+RESULT(FP),DX\n", "MOVL $7,(DX)"},
			} {
				for _, api := range []string{"text", "module", "Go"} {
					t.Run(target.triple+"/"+width+"/"+test.name+"/"+api, func(t *testing.T) {
						declaration := "func view(p *byte) uint32"
						body := "MOVL $7,ret+RESULT(FP)\n" + test.before + "LEA" + width + " ·table<>(SB),DX\n" + test.after
						source := x86LEASBUnknownAddressSource(t, target.arch, declaration, body)
						ir, err := x86ScalarFPBitViewAPI(t, source, declaration, target.arch, target.triple, api)
						if !errors.Is(err, ErrProbeNeedsContext) || ir != "" {
							t.Fatalf("SB address created numerical/caller/result/native authority: IR=%d error=%v", len(ir), err)
						}
					})
				}
			}
		}
		for _, address := range []string{"local+0(SP)", "local+0(SP)(CX*1)"} {
			for _, api := range []string{"text", "module", "Go"} {
				t.Run(target.triple+"/physical/"+address+"/"+api, func(t *testing.T) {
					body := "MOVL $7,ret+RESULT(FP)\nMOVL $0,CX\nLEAL " + address + ",DX"
					source := x86LEASBUnknownAddressSource(t, target.arch, "func view() uint32", body)
					if ir, err := x86ScalarFPBitViewAPI(t, source, "func view() uint32", target.arch, target.triple, api); !errors.Is(err, ErrProbeNeedsContext) || ir != "" {
						t.Fatalf("physical source stack acquired backing: %v", err)
					}
				})
			}
		}
	}
}
