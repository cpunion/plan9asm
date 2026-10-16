package plan9asm

import (
	"errors"
	"fmt"
	"testing"
)

func TestX86PoppedFlagWordImmediateMaskThreeAPIs(t *testing.T) {
	for _, target := range x86RawReturnTargets {
		for _, row := range []struct {
			suffix, typ string
			width       int
		}{
			{"W", "uint16", 2}, {"L", "uint32", 4}, {"Q", "uint64", 8},
		} {
			if target.arch == "amd64" && row.suffix == "L" || target.arch == "386" && row.suffix == "Q" {
				continue // These stack-width spellings are Go-rejected in this mode.
			}
			for _, control := range []struct {
				name, between string
				mask          int
				pass          bool
			}{
				{"CF-ZF", "", 65, true}, {"CF", "", 1, true}, {"ZF", "", 64, true}, {"zero", "", 0, true},
				{"unknown-PF", "", 69, false}, {"unknown-DF", "", 1024, false}, {"unknown-OF", "", 2048, false},
				{"overwritten", "MOV" + row.suffix + " BX,AX\n", 65, false},
			} {
				body := fmt.Sprintf("MOVL $1,DX\nCMPL DX,DX\nBTW $0,DX\nPUSHF%s\nPOP%s AX\n%sAND%s $%d,AX\nMOV%s AX,ret+0(FP)\nRET\n", row.suffix, row.suffix, control.between, row.suffix, control.mask, row.suffix)
				source := fmt.Sprintf("TEXT ·Value(SB),4,$0-%d\n%s", row.width, body)
				goAssemblerOracleOutput(t, target.arch, source, "", true)
				for _, api := range []string{"text", "module", "Go"} {
					t.Run(target.triple+"/"+row.suffix+"/"+control.name+"/"+api, func(t *testing.T) {
						ir, err := x86ScalarFPBitViewAPI(t, source, "func Value() "+row.typ, target.arch, target.triple, api)
						if control.pass && (err != nil || ir == "") || !control.pass && (!errors.Is(err, ErrProbeNeedsContext) || ir != "") {
							t.Fatal("captured status mask lost or invented defined bits", control.pass, err)
						}
					})
				}
			}
		}
	}
}

func TestX86NativeBitTestCompleteReachingInputsThreeAPIs(t *testing.T) {
	for _, target := range x86RawReturnTargets {
		word, move := 8, "MOVQ"
		if target.arch == "386" {
			word, move = 4, "MOVL"
		}
		for _, op := range x86BitTestOracleNames(target.arch) {
			for _, index := range []string{"$-128", "$127", "CX"} {
				body := fmt.Sprintf("%s p+0(FP),DX\n%s index+%d(FP),CX\n%s %s,(DX)\nSETCS AL\nMOVB AL,ret+%d(FP)\nRET\n", move, move, word, op, index, word*2)
				for _, control := range []struct {
					name, body string
					pass       bool
				}{
					{"complete", body, true},
					{"unknown-base", "NOP\n" + body[len(move+" p+0(FP),DX\n"):], false},
				} {
					source := fmt.Sprintf("TEXT ·Value(SB),4,$0-%d\n%s", word*2+1, control.body)
					goAssemblerOracleOutput(t, target.arch, source, "", true)
					for _, api := range []string{"text", "module", "Go"} {
						t.Run(target.triple+"/"+op+"/"+index+"/"+control.name+"/"+api, func(t *testing.T) {
							ir, err := x86ScalarFPBitViewAPI(t, source, "func Value(p *byte,index uintptr) uint8", target.arch, target.triple, api)
							if control.pass && (err != nil || ir == "") || !control.pass && (!errors.Is(err, ErrProbeNeedsContext) || ir != "") {
								t.Fatal("native BIT reaching input class mismatch", control.pass, err)
							}
							if control.pass {
								llc := findLLVM22Tool("llc")
								if llc == "" {
									t.Fatal("LLVM22 llc required")
								}
								for _, level := range []string{"-O0", "-O2"} {
									x86PackedStringCompileObject(t, llc, target.triple, ir, level)
								}
							}
						})
					}
				}
			}
		}
	}
}
