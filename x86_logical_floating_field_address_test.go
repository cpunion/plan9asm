package plan9asm

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

func x86LogicalFloatingFieldCopy(arch, kind string, extent int) (string, string) {
	lea, result := "LEAQ", 8
	if arch == "386" {
		lea = "LEAL"
		if kind == "float32" {
			result = 4
		}
	}
	move := "MOVSS"
	if kind == "float64" {
		move = "MOVSD"
	}
	return fmt.Sprintf("TEXT Value(SB),4,$0-%d\n%s x+0(FP),AX\n%s (AX),X0\n"+
			"%s ret+%d(FP),BX\n%s X0,(BX)\nRET\n", extent, lea, move, lea, result, move),
		fmt.Sprintf("func Value(x %s) %s", kind, kind)
}

// Isolate logical storage from SQRT's arithmetic effect in the unchanged
// SqrtByAddress reproducer. One field is not a contiguous incoming ABI0 frame.
func TestX86LogicalFloatingFieldAddressCopyThreeAPIsFiveTargets(t *testing.T) {
	llc := findLLVM22Tool("llc")
	if llc == "" {
		t.Fatal("LLVM22 llc required")
	}
	for _, target := range x86RawReturnTargets {
		for _, kind := range []string{"float32", "float64"} {
			extents := []int{16}
			if kind == "float32" {
				if target.arch == "386" {
					extents = []int{8}
				} else {
					extents = []int{12, 16} // Logical end and rounded native allocation.
				}
			}
			for _, extent := range extents {
				t.Run(fmt.Sprintf("%s/%s/extent-%d", target.triple, kind, extent), func(t *testing.T) {
					source, declaration := x86LogicalFloatingFieldCopy(target.arch, kind, extent)
					x86LEAActualTargetGoObject(t, target.arch, target.triple, source)
					for _, api := range []string{"text", "module", "Go"} {
						ir, err := x86ScalarFPBitViewAPI(t, source, declaration, target.arch, target.triple, api)
						if err != nil {
							t.Error("exact typed floating fields lost their private logical storage", api, err)
							continue
						}
						for _, text := range []string{ir, x86MulDivRoundOptimize(t, ir)} {
							for _, level := range []string{"-O0", "-O2"} {
								x86PackedStringCompileObject(t, llc, target.triple, text, level)
							}
						}
					}
				})
			}
		}
	}
}

func TestX86LogicalFloatingFieldAddressKeepsIndependentBoundaries(t *testing.T) {
	source, declaration := x86LogicalFloatingFieldCopy("amd64", "float32", 16)
	for _, mutation := range []struct{ name, before, after string }{
		{"parameter-padding", "x+0(FP),AX", "x+4(FP),AX"},
		{"result-padding", "ret+8(FP),BX", "ret+12(FP),BX"},
		{"uninitialized-result", "x+0(FP),AX", "ret+8(FP),AX"},
		{"cross-field-read", "MOVSS (AX),X0", "MOVSD (AX),X0"},
		{"input-store", "LEAQ ret+8(FP),BX", "MOVSS X0,(AX)\nLEAQ ret+8(FP),BX"},
		{"physical-SP", "LEAQ x+0(FP),AX\nMOVSS (AX),X0", "LEAQ x+0(FP),SP\nMOVSS (SP),X0"},
	} {
		t.Run(mutation.name, func(t *testing.T) {
			changed := strings.Replace(source, mutation.before, mutation.after, 1)
			if changed == source {
				t.Fatal("negative source mutation was not applied")
			}
			x86LEAActualTargetGoObject(t, "amd64", "x86_64-unknown-linux-gnu", changed)
			for _, api := range []string{"text", "module", "Go"} {
				ir, err := x86ScalarFPBitViewAPI(t, changed, declaration, "amd64", "x86_64-unknown-linux-gnu", api)
				if ir != "" || !errors.Is(err, ErrProbeNeedsContext) {
					t.Fatal("floating field acquired padding, input-write or native state authority", api, err)
				}
			}
		})
	}
}
