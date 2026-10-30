package plan9asm

import (
	"errors"
	"fmt"
	"reflect"
	"testing"
)

func TestGoAMD64FixedLocalOriginalMoveAddressDomains(t *testing.T) {
	for _, width := range []string{"W", "L", "Q"} {
		for _, old := range []uint8{0, 3, 0x55, 0xff} {
			t.Run(fmt.Sprintf("%s/old%x", width, old), func(t *testing.T) {
				ins := x86ScalarLaneInstruction(t, "MOV"+width+" $slot-16(SP),BX")
				original := fmt.Sprintf("%#v", ins)
				state := x86FixedLocalState{pointers: map[Reg]int64{SP: 0, BX: 8}, known: map[Reg]uint8{BX: old},
					opaque: map[Reg]x86FixedLocalOrigins{}, constants: map[Reg]int64{BX: 7}, booleans: map[Reg]uint8{}}
				proof := &x86FixedLocalProof{frame: 16}
				if known, err := fixedLocalPrivateAddressStep(ins, &state, proof); !known || err != nil {
					t.Fatal("original source address did not reach its existing private LEA consumer", known, err)
				}
				if original != fmt.Sprintf("%#v", ins) {
					t.Fatal("source-address projection changed the original MOV identity")
				}
				value, pointer := state.pointers[BX]
				want := old &^ 3
				if width == "L" {
					want = 0xf0
				} else if width == "Q" {
					want = 0
				}
				if pointer != (width == "Q") || pointer && value != 0 || state.known[BX] != want ||
					len(state.constants) != 0 || len(state.opaque) != 0 {
					t.Fatal("partial address became NUM or retained a full pointer identity", state)
				}
			})
		}
		for _, control := range []string{"missing-proof", "changed-raw", "changed-operand", "physical-origin", "address-size", "SP-output"} {
			t.Run(width+"/"+control, func(t *testing.T) {
				ins := x86ScalarLaneInstruction(t, "MOV"+width+" $slot-16(SP),BX")
				proof := &x86FixedLocalProof{frame: 16}
				switch control {
				case "missing-proof":
					proof = nil
				case "changed-raw":
					ins.Raw = "MOV" + width + " $other-8(SP),BX"
				case "changed-operand":
					ins.Args[0].Sym = "$other-8(SP)"
				case "physical-origin":
					ins.x86Encoded = true
				case "address-size":
					ins.x86AddressBits = 32
				case "SP-output":
					ins = x86ScalarLaneInstruction(t, "MOV"+width+" $slot-16(SP),SP")
				}
				state := x86FixedLocalState{pointers: map[Reg]int64{SP: 0}, known: map[Reg]uint8{},
					opaque: map[Reg]x86FixedLocalOrigins{}, constants: map[Reg]int64{}, booleans: map[Reg]uint8{}}
				before := fmt.Sprintf("%#v", state)
				known, err := fixedLocalPrivateAddressStep(ins, &state, proof)
				if known && !errors.Is(err, ErrProbeNeedsContext) || !reflect.DeepEqual(fmt.Sprintf("%#v", state), before) {
					t.Fatal("missing/changed source or physical role supplied address permission", known, err)
				}
			})
		}
	}
}
