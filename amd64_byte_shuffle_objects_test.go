package plan9asm

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestX86ByteShuffleThreeAPIsFiveTargetObjects(t *testing.T) {
	llc := findLLVM22Tool("llc")
	if llc == "" {
		t.Fatal("LLVM 22 llc required")
	}
	for _, target := range x86RawReturnTargets {
		t.Run(target.triple, func(t *testing.T) {
			rows := x86ByteShuffleFormatRows(target.arch)
			for begin := 0; begin < len(rows); begin += 64 {
				end := begin + 64
				if end > len(rows) {
					end = len(rows)
				}
				t.Run(fmt.Sprint(begin), func(t *testing.T) {
					var source, declarations strings.Builder
					for index, row := range rows[begin:end] {
						if strings.HasPrefix(row, "PSHUFB ") && (strings.Contains(row, "(FP)") || strings.Contains(row, "(SP)")) {
							// Explicit native-storage rejection controls below.
							continue
						}
						fmt.Fprintf(&declarations, "func Shuffle%d(value [64]byte)\n", index)
						fmt.Fprintf(&source, "TEXT ·Shuffle%d(SB),4,$64-64\n%s\nRET\n", index, row)
					}
					x86LEAActualTargetGoObject(t, target.arch, target.triple, source.String())
					for _, api := range []string{"text", "module", "Go"} {
						ir, err := x86ScalarFPBitViewAPI(t, source.String(), declarations.String(), target.arch, target.triple, api)
						if err != nil {
							t.Fatal(api, err)
						}
						x86ResultInitializationObject(t, llc, target.triple, ir, "byte-shuffle-"+api)
						x86ResultInitializationObject(t, llc, target.triple, x86MulDivRoundOptimize(t, ir), "byte-shuffle-"+api+"-optimized")
					}
				})
			}
		})
	}
}

func TestX86ByteShuffleVirtualLegacyStorageRemainsContext(t *testing.T) {
	for _, row := range []string{"PSHUFB value+0(FP),X1", "PSHUFB local-16(SP),X1", "PSHUFB (SP),X1", "PSHUFB (BP),X1"} {
		source := "TEXT ·Shuffle(SB),4,$64-64\n" + row + "\nRET\n"
		for _, target := range x86RawReturnTargets {
			x86LEAActualTargetGoObject(t, target.arch, target.triple, source)
			for _, api := range []string{"text", "module", "Go"} {
				_, err := x86ScalarFPBitViewAPI(t, source, "func Shuffle(value [64]byte)", target.arch, target.triple, api)
				if !errors.Is(err, ErrProbeNeedsContext) {
					t.Fatal("logical storage acquired original SSE alignment", row, target.triple, api, err)
				}
			}
		}
	}
}
