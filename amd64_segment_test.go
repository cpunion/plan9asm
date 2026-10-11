package plan9asm

import (
	"strings"
	"testing"
)

func TestTranslateAMD64SegmentMemory(t *testing.T) {
	llc := findLLVM22Tool("llc")
	if llc == "" {
		t.Fatal("LLVM 22 llc required")
	}
	src := `
TEXT ·loadGS(SB),4,$0-0
	MOVQ 0x30(GS), DI
	MOVQ DI, 0(CX)(GS)
	MOVQ -8(FS), AX
	RET
`
	triple := "x86_64-pc-windows-msvc"
	x86LEAActualTargetGoObject(t, "amd64", triple, src)
	for _, api := range []string{"text", "module", "Go"} {
		t.Run(api, func(t *testing.T) {
			ir, err := x86ScalarFPBitViewAPI(t, src, "func loadGS()", "amd64", triple, api)
			if err != nil {
				t.Fatal(err)
			}
			for _, optimized := range []bool{false, true} {
				output, name := ir, "segment"
				if optimized {
					output, name = x86MulDivRoundOptimize(t, ir), "segment-optimized"
				}
				for _, want := range []string{
					"load volatile i64, ptr addrspace(256)",
					"store i64",
					"load volatile i64, ptr addrspace(257)",
				} {
					if !strings.Contains(output, want) {
						t.Fatalf("segment-relative output missing %q:\n%s", want, output)
					}
				}
				if strings.Contains(output, "0x30(GS)") {
					t.Fatalf("segment-relative memory was emitted as a symbol:\n%s", output)
				}
				compileLLVMToObject(t, llc, triple, name+".ll", name+".o", output)
			}
		})
	}
}

func TestAMD64SegmentMemoryRejectsInvalidUses(t *testing.T) {
	c, _ := newAMD64CtxWithFuncForTest(t, Func{}, FuncSig{Name: "segmentErrors", Ret: Void}, nil)
	if _, err := c.addrFromMem(MemRef{Segment: GS}); err == nil {
		t.Fatal("addrFromMem accepted a segment-relative address")
	}
	if _, _, err := c.ptrFromMem(MemRef{Segment: Reg("CS")}); err == nil {
		t.Fatal("ptrFromMem accepted an unsupported segment")
	}
	if ok, _, err := c.lowerMov(OpMOVQ, Instr{Args: []Operand{
		{Kind: OpMem, Mem: MemRef{Segment: Reg("CS")}},
		{Kind: OpReg, Reg: AX},
	}}); !ok || err == nil {
		t.Fatalf("lowerMov(invalid segment) = (%v, %v), want handled error", ok, err)
	}
}
