package plan9asm

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestX86IncomingPointerResidueWholeSourceThreeAPIs(t *testing.T) {
	for _, target := range x86RawReturnTargets {
		word, copy := goWordSize(target.arch), "MOVQ"
		if target.arch == "386" {
			copy = "MOVL"
		}
		for _, row := range []struct {
			op, store, typ string
			width          int
		}{{"ANDB", "MOVB", "uint8", 1}, {"ANDW", "MOVW", "uint16", 2}, {"ANDL", "MOVL", "uint32", 4}, {"ANDQ", "MOVQ", "uint64", 8}} {
			if target.arch == "386" && row.width == 8 {
				continue // No Go Q instruction in this source mode.
			}
			for _, scenario := range []string{"actual-pointer", "copied-pointer", "unknown", "all-bits", "sparse-mask"} {
				for _, api := range []string{"text", "module", "Go"} {
					t.Run(target.triple+"/"+row.op+"/"+scenario+"/"+api, func(t *testing.T) {
						body := ""
						if scenario != "unknown" {
							body = copy + " p+0(FP),AX\n"
						}
						if scenario == "copied-pointer" {
							body += copy + " AX,BX\n"
							body += copy + " BX,AX\n"
						}
						mask := "7"
						if scenario == "all-bits" {
							mask = "-1"
						} else if scenario == "sparse-mask" {
							mask = "5"
						}
						body += row.op + " $" + mask + ",AX\n" + fmt.Sprintf("%s AX,ret+%d(FP)\nRET\n", row.store, word)
						source := fmt.Sprintf("TEXT ·Value(SB),4,$0-%d\n%s", word+row.width, body)
						ir, err := x86ScalarFPBitViewAPI(t, source, "func Value(p *byte) "+row.typ, target.arch, target.triple, api)
						positive := scenario == "actual-pointer" || scenario == "copied-pointer"
						if positive && (err != nil || ir == "") {
							t.Fatal("actual pointer low-bit projection did not produce a defined scalar", err)
						}
						if !positive && (!errors.Is(err, ErrProbeNeedsContext) || ir != "") {
							t.Fatal("unregistered pointer representation became general NUM", scenario, err)
						}
					})
				}
			}
		}
		for _, row := range []struct{ name, params, body string }{
			{"native-after-projection", "p *byte", copy + " p+0(FP),SI\nAND" + strings.TrimPrefix(copy, "MOV") + " $7,SI\nMOVL $0,AX\nCRC32B (SI),AX\n"},
			{"numeric-native-address", "p uintptr", copy + " p+0(FP),SI\nAND" + strings.TrimPrefix(copy, "MOV") + " $7,SI\nMOVL $0,AX\nCRC32B (SI),AX\n"},
			{"unknown-join", "p *byte", "MOVL $0,CX\nCMPL CX,$0\nJE unknown\n" + copy + " p+0(FP),AX\nJMP join\nunknown:\nNOP\njoin:\nAND" + strings.TrimPrefix(copy, "MOV") + " $7,AX\n"},
			{"result-address", "p *byte", "LEA" + strings.TrimPrefix(copy, "MOV") + fmt.Sprintf(" ret+%d(FP),AX\n", word) + "AND" + strings.TrimPrefix(copy, "MOV") + " $7,AX\n"},
		} {
			for _, api := range []string{"text", "module", "Go"} {
				t.Run(target.triple+"/"+row.name+"/"+api, func(t *testing.T) {
					source := fmt.Sprintf("TEXT ·Value(SB),4,$0-%d\n%sMOVL AX,ret+%d(FP)\nRET\n", word+4, row.body, word)
					ir, err := x86ScalarFPBitViewAPI(t, source, "func Value("+row.params+") uint32", target.arch, target.triple, api)
					if !errors.Is(err, ErrProbeNeedsContext) || ir != "" {
						t.Fatal("residue manufactured an address/unknown/padding authority", err)
					}
				})
			}
		}
	}
}
