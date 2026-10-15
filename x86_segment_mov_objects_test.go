package plan9asm

import (
	"fmt"
	"strings"
	"testing"
)

type x86SegmentMOVFunction struct {
	name                string
	width, addressSpace int
}

func x86SegmentMOVCompleteSource(t *testing.T, arch string) (string, string, []x86SegmentMOVFunction) {
	t.Helper()
	var source, declarations strings.Builder
	var functions []x86SegmentMOVFunction
	appendFunction := func(assembly, declaration string, width, space int) {
		name := fmt.Sprintf("segment%d", len(functions))
		if strings.Count(assembly, "·view(SB)") != 1 || strings.Count(declaration, "view(") != 1 {
			t.Fatal("complete segment fixture lost its one source/type identity")
		}
		source.WriteString(strings.Replace(assembly, "·view(SB)", "·"+name+"(SB)", 1))
		declarations.WriteString(strings.Replace(declaration, "view(", name+"(", 1) + "\n")
		functions = append(functions, x86SegmentMOVFunction{name, width, space})
	}
	addresses := []string{"fixed", "fixed-zero", "base", "segment-index-scale1", "segment-index-scale2", "segment-index-scale4", "segment-index-scale8"}
	if arch == "386" {
		addresses = append(addresses, "mode32-BP")
	}
	for _, row := range x86SegmentMOVRows(arch) {
		for _, segment := range []string{"FS", "GS"} {
			space := 257
			if segment == "GS" {
				space = 256
			}
			for _, address := range addresses {
				for _, dead := range []bool{false, true} {
					assembly, declaration := x86SegmentMOVFixture(t, arch, row, segment, address, dead)
					appendFunction(assembly, declaration, row.width, space)
				}
			}
			assembly, declaration := x86SegmentMOVVoidFixture(arch, row, segment)
			appendFunction(assembly, declaration, row.width, space)
			for _, address := range []string{"fixed", "base"} {
				assembly, declaration := x86SegmentMOVVoidAddressFixture(arch, row, segment, address)
				appendFunction(assembly, declaration, row.width, space)
			}
		}
	}
	if arch == "amd64" {
		for _, segment := range []string{"FS", "GS"} {
			space := 257
			if segment == "GS" {
				space = 256
			}
			for _, alias := range []string{"DH", "BH"} {
				for _, dead := range []bool{false, true} {
					assembly, declaration := x86SegmentMOVREXFixture(t, segment, alias, dead)
					appendFunction(assembly, declaration, 1, space)
				}
			}
		}
	}
	want := 212
	if arch == "386" {
		want = 114
	}
	if len(functions) != want || strings.Count(source.String(), "TEXT ") != want || strings.Count(declarations.String(), "func ") != want {
		t.Fatal("independent complete segment source denominator changed", len(functions), want)
	}
	return source.String(), declarations.String(), functions
}

func TestX86SegmentMOVFiveTargetThreeAPICompleteObjects(t *testing.T) {
	llc := findLLVM22Tool("llc")
	if llc == "" {
		t.Fatal("LLVM22 llc is required")
	}
	for _, target := range x86RawReturnTargets {
		t.Run(target.triple, func(t *testing.T) {
			source, declarations, functions := x86SegmentMOVCompleteSource(t, target.arch)
			x86LEAActualTargetGoObject(t, target.arch, target.triple, source)
			for _, api := range []string{"text", "module", "Go"} {
				ir, err := x86ScalarFPBitViewAPI(t, source, declarations, target.arch, target.triple, api)
				if err != nil || ir == "" {
					t.Fatal("complete uncut segment MOV source was rejected", err)
				}
				for _, irLevel := range []string{"0", "2"} {
					input := ir
					if irLevel == "2" {
						input = x86DwordProductIROptimize(t, input)
					}
					if strings.Count(input, "\ndefine ") != len(functions) {
						t.Fatal("whole-file translation added or cropped a function", api, irLevel)
					}
					for _, function := range functions {
						x86SegmentMOVNamedReadIR(t, input, function.name, function.width, function.addressSpace)
					}
					for _, backendLevel := range []string{"0", "2"} {
						t.Run(api+"/IR"+irLevel+"/backend"+backendLevel, func(t *testing.T) {
							x86PackedStringCompileObject(t, llc, target.triple, input, "-O"+backendLevel)
						})
					}
				}
			}
		})
	}
}
