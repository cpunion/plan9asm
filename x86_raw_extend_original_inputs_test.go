package plan9asm

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"strings"
	"testing"
)

func x86RawExtendOriginalInputSource(t *testing.T, block bool) string {
	t.Helper()
	name := "TestTranslateRawEVEXPackedExtendWeaviateRegression"
	want := "a5c6fa0843f6aa5ce29bd383ca176aa4c00ab8316c9886efb214d3b9ac6447cb"
	if block {
		name = "TestTranslateRawVEXPackedExtendWeaviateBlockRegression"
		want = "c16c5124003d8cf18163d034586f7676da643493c7f4c58fab043a88676e74c3"
	}
	source := x86X87FunctionLiteral(t, "amd64_raw_packed_extend_test.go", name, "source")
	if fmt.Sprintf("%x", sha256.Sum256([]byte(source))) != want {
		t.Fatal("original complete packed-extend block changed", name)
	}
	return source
}

func TestX86RawExtendOriginalMissingInputsRemainContext(t *testing.T) {
	for _, block := range []bool{false, true} {
		source := x86RawExtendOriginalInputSource(t, block)
		file := mustParseX86RawReturn(t, source)
		if len(file.Funcs) != 1 {
			t.Fatal("original packed-extend entry lost its identity")
		}
		name := file.Funcs[0].Sym
		for _, target := range x86RawReturnTargets {
			t.Run(fmt.Sprintf("block=%t/%s", block, target.triple), func(t *testing.T) {
				x86LEAActualTargetGoObject(t, target.arch, target.triple, source)
				for _, api := range []string{"text", "module", "Go"} {
					ir, err := x86ScalarFPBitViewAPI(t, source, "func "+name+"()", target.arch, target.triple, api)
					if ir != "" || !errors.Is(err, ErrProbeNeedsContext) {
						t.Fatal("original bare block invented its native address/vector inputs", api, err)
					}
				}
			})
		}
	}
}

func TestX86RawExtendOriginalInitializedFiveTargetObjects(t *testing.T) {
	llc := findLLVM22Tool("llc")
	if llc == "" {
		t.Fatal("LLVM22 llc required")
	}
	for _, block := range []bool{false, true} {
		original := x86RawExtendOriginalInputSource(t, block)
		header, body, found := strings.Cut(original, "\n")
		if !found || !strings.Contains(header, "$0-0") {
			t.Fatal("original complete packed-extend source lost its TEXT")
		}
		name := mustParseX86RawReturn(t, original).Funcs[0].Sym
		for _, target := range x86RawReturnTargets {
			t.Run(fmt.Sprintf("block=%t/%s", block, target.triple), func(t *testing.T) {
				word, move := 8, "MOVQ"
				if target.arch == "386" {
					word, move = 4, "MOVL"
				}
				// Keep every original opcode and RET. Only this distinct
				// counterpart imports the missing SI/DI pointers, defines the
				// actual AX index and initializes all vector accumulators.
				prefix := fmt.Sprintf("%s\n%s a+0(FP),SI\n%s b+%d(FP),DI\nMOVL $0,AX\nVZEROALL\n",
					strings.Replace(header, "$0-0", fmt.Sprintf("$0-%d", 2*word), 1), move, move, word)
				source := prefix + body
				declaration := "func " + name + "(a,b *[128]byte)"
				x86LEAActualTargetGoObject(t, target.arch, target.triple, source)
				for _, api := range []string{"text", "module", "Go"} {
					ir, err := x86ScalarFPBitViewAPI(t, source, declaration, target.arch, target.triple, api)
					if err != nil || ir == "" {
						t.Fatal("complete original-byte initialized block lost its paired native reader", api, err)
					}
					if block && strings.Count(ir, "asm sideeffect \"vcvtdq2ps") != 4 {
						t.Fatal("original complete block lost its four floating conversions")
					}
					x86ResultInitializationObject(t, llc, target.triple, ir, "extend-original-"+api)
					x86ResultInitializationObject(t, llc, target.triple, x86SignedIntegerFloatOptimize(t, ir), "extend-original-"+api+"-O2")
				}
			})
		}
	}
}
