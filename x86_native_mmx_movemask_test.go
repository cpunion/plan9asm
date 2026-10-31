package plan9asm

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/xgo-dev/plan9asm/internal/goasmrules"
)

func TestX86NativeMMXMoveMaskByteEffects(t *testing.T) {
	for _, arch := range []string{"amd64", "386"} {
		registers := []Reg{AX, BX}
		if arch == "amd64" {
			registers = append(registers, "R8")
		}
		for bank := 0; bank < 8; bank++ {
			for _, destination := range registers {
				t.Run(fmt.Sprintf("%s/M%d/%s", arch, bank, destination), func(t *testing.T) {
					instruction := fmt.Sprintf("PMOVMSKB M%d,%s", bank, destination)
					_, file, sig, _ := x86ResultInitializationFixture(t, arch, "func Value(a uint64) uint32",
						fmt.Sprintf("MOVQ a+0(FP),M%d\n%s\nEMMS\nMOVL %s,ret+8(FP)\nRET", bank, instruction, destination))
					plan, err := newX86NativeMMXRegion(file.Funcs[0], sig, arch, x86Legacy127Targets(arch)[0])
					if err != nil || plan == nil {
						t.Fatal("MMX_MOVE_MASK_TYPED_REGION_GAP", err)
					}
					ins := file.Funcs[0].Instrs[1]
					index, _, valid := x86ResultGP(destination, 4, arch)
					if !valid {
						t.Fatal("independent GP fixture invalid")
					}
					state := x86ResultState{flags: x86ByteFlagsCF | x86ByteFlagsZF, direction: x86ResultDirectionBackward}
					state.nativeMMX[bank] = 0xff
					state.canonicalGP[index] = 0xff
					state.pointerGP[index], state.callerAddressGP[index], state.callerAddressOrigin[index] = true, true, 7
					state.resultAddressGP[index] = x86ResultSlotAddress{start: 8, end: 12, valid: true}
					state.boolFrameGP = new([16]x86ResultSlotAddress)
					state.boolFrameGP[index] = state.resultAddressGP[index]
					before := state.clone()
					if accepted, known := plan.transfer(1, ins, &state, nil, 8, arch); !accepted || !known {
						t.Fatal("defined physical MMX source not transferred")
					}
					expected := before.clone()
					expected.gp[index] = 0xff
					expected.canonicalGP[index] = 0xf0
					expected.pointerGP[index], expected.callerAddressGP[index], expected.callerAddressOrigin[index] = false, false, 0
					expected.resultAddressGP[index], expected.boolFrameGP[index] = x86ResultSlotAddress{}, x86ResultSlotAddress{}
					if !reflect.DeepEqual(expected, state) {
						t.Fatal("r32 result changed unrelated source/FLAGS or retained pointer/bool facts", state, expected)
					}
					for missing := 0; missing < 8; missing++ {
						state = before.clone()
						state.nativeMMX[bank] &^= 1 << uint(missing)
						unchanged := state.clone()
						if accepted, known := plan.transfer(1, ins, &state, nil, 8, arch); accepted || !known || !reflect.DeepEqual(state, unchanged) {
							t.Fatal("missing physical source byte supplied a value or mutated state", missing)
						}
					}
					state = before.clone()
					if accepted, known := x86ResultMoveMaskStep(ins, &state, nil, 8, arch); accepted || !known {
						t.Fatal("ordinary observer acquired physical MMX")
					}
					changed := ins
					changed.Raw += " // changed"
					state = before.clone()
					if accepted, known := plan.transfer(1, changed, &state, nil, 8, arch); accepted || !known {
						t.Fatal("private source identity changed")
					}
				})
			}
		}
	}
}

func TestX86NativeMMXMoveMaskRoleBoundaries(t *testing.T) {
	for _, arch := range []string{"amd64", "386"} {
		for _, body := range []string{
			"PMOVMSKB M0,AX\nEMMS",                              // No physical source definition.
			"MOVQ a+0(FP),M0\nPMOVMSKB M0,AX",                   // No real tag cleanup.
			"MOVQ a+0(FP),M0\nPMOVMSKB M0,AX\nMOVL $1,CX\nEMMS", // Interrupted physical region.
		} {
			_, _, _, source := x86ResultInitializationFixture(t, arch, "func Value(a uint64) uint32", body+"\nMOVL AX,ret+8(FP)\nRET")
			if _, err := x86ScalarLiteralGoListing(t, arch, source); err != nil {
				t.Fatal("negative is not real original Go", err)
			}
			for _, api := range []string{"text", "module", "Go"} {
				ir, err := x86ScalarFPBitViewAPI(t, source, "func Value(a uint64) uint32", arch, x86Legacy127Targets(arch)[0], api)
				if !errors.Is(err, ErrProbeNeedsContext) || ir != "" {
					t.Fatal("MMX state/input boundary widened", arch, api, body, err)
				}
			}
		}
	}
}

func TestX86NativeMMXMoveMaskOriginalGoRegisterBytes(t *testing.T) {
	for _, arch := range []string{"amd64", "386"} {
		registers := []string{"AX", "CX", "DX", "BX", "SP", "BP", "SI", "DI"}
		if arch == "amd64" {
			registers = append(registers, "R8", "R9", "R10", "R11", "R12", "R13", "R14", "R15")
		}
		var fixtures []x86OfficialPackedCase
		for bank := 0; bank < 8; bank++ {
			for reg, destination := range registers {
				code := []byte{0x0f, 0xd7, 0xc0 | byte(reg&7)<<3 | byte(bank)}
				if reg >= 8 {
					code = append([]byte{0x44}, code...)
				}
				fixtures = append(fixtures, x86OfficialPackedCase{witness: goasmrules.X86PackedBinaryWitness{Instruction: fmt.Sprintf("PMOVMSKB M%d,%s", bank, destination), Encoding: code, Destination: destination}})
			}
		}
		for start := 0; start < len(fixtures); start += 32 {
			end := start + 32
			if end > len(fixtures) {
				end = len(fixtures)
			}
			batch := append([]x86OfficialPackedCase(nil), fixtures[start:end]...)
			var source strings.Builder
			for i := range batch {
				batch[i].symbol = fmt.Sprintf("MMXMaskBytes%d", start+i)
				fmt.Fprintf(&source, "TEXT ·%s(SB),4,$0-0\n", batch[i].symbol)
				batch[i].line = strings.Count(source.String(), "\n") + 1
				fmt.Fprintln(&source, batch[i].witness.Instruction)
				source.WriteString("RET\n")
			}
			listing, err := x86ScalarLiteralGoListing(t, arch, source.String())
			if err != nil {
				t.Fatal(err)
			}
			if err := x86OfficialPackedVerifyListing(listing, batch); err != nil {
				t.Fatal(err)
			}
		}
	}
}
