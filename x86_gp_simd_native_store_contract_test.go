package plan9asm

import (
	"errors"
	"fmt"
	"reflect"
	"testing"
)

// Independent named store inventory: Go's MOVD and MOVDQ2Q alias MOVQ,
// while VMOVQ's memory width stays eight bytes even in mode32.
var x86GPSIMDNativeStoreRows = []struct {
	op    string
	bytes int64
}{
	{"MOVL", 4}, {"MOVQ", 8}, {"MOVD", 8}, {"MOVDQ2Q", 8},
	{"VMOVD", 4}, {"VMOVQ", 8},
}

func TestX86GPSIMDNativeStoreNumericalContinuation(t *testing.T) {
	for _, arch := range []string{"amd64", "386"} {
		for _, row := range x86GPSIMDNativeStoreRows {
			t.Run(arch+"/"+row.op, func(t *testing.T) {
				ins := x86YMMReadSource(t, row.op+" X0,(BX)")
				form, known, err := x86GPSIMDTransferProperties(ins, arch)
				if !known || err != nil || !form.memoryWrite || int64(form.bits/8) != row.bytes {
					t.Fatal("complete named store grammar changed", form, err)
				}
				state := x86ResultState{flags: x86ByteFlagsAll, direction: x86ResultDirectionBackward}
				state.writeVector(0, 64, ^uint64(0), false)
				state.pointerGP[3], state.gp[3], state.masks[1] = true, 0xff, true
				before := state.clone()
				if !x86ResultGPSIMDTransferWithRead(ins, form, &state, nil, 0, arch, true) || !reflect.DeepEqual(state, before) {
					t.Fatal("paired exact native store lost its input or changed numerical/storage facts")
				}
				if x86ResultGPSIMDTransfer(ins, form, &state, nil, 0, arch) {
					t.Fatal("standalone observer acquired native write permission")
				}
				for missing := int64(-1); missing < row.bytes; missing++ {
					state = before.clone()
					if missing < 0 {
						state.pointerGP[3] = false
					} else {
						state.writeVector(0, 64, ^(uint64(1) << uint(missing)), false)
					}
					original := state.clone()
					if x86ResultGPSIMDTransferWithRead(ins, form, &state, nil, 0, arch, true) || !reflect.DeepEqual(state, original) {
						t.Fatal("missing address/data acquired a native store or changed state", missing)
					}
				}
			})
		}
	}
}

func TestX86GPSIMDNativeStoreVoidCannotInventInputs(t *testing.T) {
	for _, arch := range []string{"amd64", "386"} {
		triple, move, word := "x86_64-unknown-linux-gnu", "MOVQ", 8
		if arch == "386" {
			triple, move, word = "i386-unknown-linux-gnu", "MOVL", 4
		}
		for _, row := range x86GPSIMDNativeStoreRows {
			for _, unused := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/%s/unused-bool=%t", arch, row.op, unused), func(t *testing.T) {
					declaration, extent := "func Value(p *byte)", word
					if unused {
						declaration, extent = "func Value(p *byte,unused bool)", word+1
					}
					source := fmt.Sprintf("TEXT ·Value(SB),4,$0-%d\n%s p+0(FP),BX\n%s X0,(BX)\nRET\n", extent, move, row.op)
					x86LEAActualTargetGoObject(t, arch, triple, source)
					for _, api := range []string{"text", "module", "Go"} {
						ir, err := x86ScalarFPBitViewAPI(t, source, declaration, arch, triple, api)
						if ir != "" || !errors.Is(err, ErrProbeNeedsContext) {
							t.Fatal("Void or unused bool invented the stored vector bytes", api, err)
						}
					}
				})
			}
		}
	}
}
