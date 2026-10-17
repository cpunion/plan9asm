package plan9asm

import (
	"fmt"
	"reflect"
	"testing"
)

// Independent sibling inventory: MOVS/STOS/LODS share the existing twenty
// Go ynone rows, but not the comparison family's FLAGS or conditional exit.
func TestX86StringLoadStorePairedEffectsCompleteFamily(t *testing.T) {
	for _, arch := range []string{"386", "amd64"} {
		widths := "BWLQ"
		if arch == "386" {
			widths = "BWL"
		}
		for _, stem := range []string{"MOVS", "STOS", "LODS"} {
			for index, suffix := range widths {
				width := []int{1, 2, 4, 8}[index]
				for _, prefix := range []string{"", "REP", "REPN"} {
					for _, count := range []x86ResultStringCount{x86StringCountUnknown, x86StringCountZero, x86StringCountNonzero} {
						for _, direction := range []x86ResultDirection{x86ResultDirectionForward, x86ResultDirectionBackward, x86ResultDirectionEither} {
							t.Run(fmt.Sprintf("%s/%s%c/%s/count%d/DF%d", arch, stem, suffix, prefix, count, direction), func(t *testing.T) {
								source := "TEXT strings(SB),4,$0-0\n"
								if prefix != "" {
									source += prefix + "\n"
								}
								file, err := Parse(ArchAMD64, source+stem+string(suffix)+"\nRET\n")
								if err != nil {
									t.Fatal(err)
								}
								plan, err := x86ResultStringMemorySourcePlan(file.Funcs[0], arch)
								if err != nil || plan == nil || len(plan.bodies) != 1 {
									t.Fatal("shared typed string grammar omitted the complete load/store sibling", err)
								}
								for ordinal, site := range plan.bodies {
									state := x86ResultState{flags: x86ByteFlagCF | x86ByteFlagAF, direction: direction, stringCount: count}
									state.gp[0], state.gp[1], state.gp[3] = 0xff, 0xff, 0x55
									if stem == "LODS" {
										state.gp[0] = 0xf0 // Unspecified accumulator low bytes stay unspecified on count0.
									}
									state.canonicalGP[0] = 0xa0
									for _, register := range []int{6, 7} {
										state.pointerGP[register], state.callerAddressGP[register] = true, true
										state.callerAddressOrigin[register] = uint16(register + 1)
									}
									expected := state.clone()
									if prefix == "" || count != x86StringCountZero {
										indexes := []int{7}
										if stem == "LODS" {
											indexes = []int{6}
											mask := uint8((uint16(1) << uint(width)) - 1)
											if prefix == "" || count == x86StringCountNonzero {
												expected.gp[0] |= mask
											}
											expected.canonicalGP[0] &^= mask
											if width == 4 && (prefix == "" || count == x86StringCountNonzero) {
												expected.gp[0] |= 0xf0
												expected.canonicalGP[0] |= 0xf0
											}
										} else if stem == "MOVS" {
											indexes = append(indexes, 6)
										}
										for _, register := range indexes {
											expected.pointerGP[register], expected.callerAddressGP[register] = false, false
											expected.callerAddressOrigin[register] = 0
										}
										if prefix != "" {
											// Unlike SCAS/CMPS, every completed repeated transfer
											// consumes the entire counter; no predicate exits early.
											expected.gp[1], expected.canonicalGP[1] = 0xff, 0xff
											expected.stringCount = x86StringCountZero
										}
									}
									if accepted, known := plan.transfer(ordinal, site.instruction, &state, arch); !accepted || !known || !reflect.DeepEqual(state, expected) {
										t.Fatal("native transfer lost conditional bytes, zero counter or independent FLAGS/DF", accepted, known)
									}
								}
							})
						}
					}
				}
			}
		}
	}
}
