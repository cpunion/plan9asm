package plan9asm

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestX86IncomingPointerResidueMasksRolesAndByteAliases(t *testing.T) {
	for _, arch := range []string{"amd64", "386"} {
		for _, row := range []struct {
			op    string
			width int64
		}{{"ANDB", 1}, {"ANDW", 2}, {"ANDL", 4}, {"ANDQ", 8}} {
			if arch == "386" && row.width == 8 {
				continue
			}
			for _, register := range strings.Fields("AX CX DX BX BP SI DI R8 R15 AL AH BL BH CL CH DL DH BPB SIB DIB R8B R15B") {
				for _, mask := range []int64{0, 1, 3, 7, 5, -1} {
					ins := x86ScalarLaneInstruction(t, fmt.Sprintf("%s $%d,%s", row.op, mask, register))
					// Byte identity is the separately validated existing grammar;
					// the expectation below independently selects role/mask bits.
					form, known, err := x86LockScalarRMWProperties(ins.Op, ins, arch)
					if !known || err != nil {
						continue
					}
					index, byteMask, valid := x86ResultGP(form.destination.Reg, row.width, arch)
					if !valid {
						continue // Physical SP/amd64 BP and mode32 high banks have no role.
					}
					for _, caller := range []bool{false, true} {
						for _, pointer := range []bool{false, true} {
							state := x86ResultState{flags: x86ByteFlagsAll, direction: x86ResultDirectionBackward}
							state.pointerGP[index], state.callerAddressGP[index] = pointer, caller
							before := state.clone()
							positive := caller && pointer && (mask == 0 || mask == 1 || mask == 3 || mask == 7)
							got := x86ScalarFPResultTransfer(ins, &state, nil, 0, arch)
							if got != positive {
								t.Fatalf("%s %s %s mask=%d caller=%t pointer=%t: got=%t want=%t", arch, row.op, register, mask, caller, pointer, got, positive)
							}
							if positive {
								before.gp[index] = byteMask
								if row.width >= 4 {
									before.gp[index] = 255
								}
								before.pointerGP[index], before.callerAddressGP[index] = false, false
								before.flags = x86ByteFlagsAll &^ x86ByteFlagAF
							}
							if !reflect.DeepEqual(state, before) {
								t.Fatal("projection changed unrelated bytes/flags/DF or retained address authority", arch, ins.Raw)
							}
						}
						}
					}
				}
			}
		}
	}
}

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
