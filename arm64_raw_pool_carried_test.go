package plan9asm

import (
	"fmt"
	"testing"
)

func TestARM64PoolCarriedStridedAddress(t *testing.T) {
	for _, step := range []uint64{1, 2, 4, 8, 16} {
		for _, ascending := range []bool{false, true} {
			for _, pointerIncreasing := range []bool{false, true} {
				name := fmt.Sprintf("step=%d/ascending=%v/pointerIncreasing=%v", step, ascending, pointerIncreasing)
				t.Run(name, func(t *testing.T) {
					lines := []string{fmt.Sprintf("and x1,x0,#%d", step), fmt.Sprintf("add x1,x1,#%d", step),
						"lsl x3,x1,#3", "add x3,x3,#32"}
					pointerDelta := -int64(step * 8)
					want := arm64PoolInterval{32 + step*8, 32 + step*16}
					if pointerIncreasing {
						lines = append(lines, "mov x4,#512", "sub x3,x4,x3")
						pointerDelta, want = -pointerDelta, arm64PoolInterval{512 - want.high, 512 - want.low}
					}
					update := fmt.Sprintf("sub x1,x1,#%d", step)
					if ascending {
						lines = append(lines, "neg x1,x1")
						update = fmt.Sprintf("add x1,x1,#%d", step)
					}
					head := len(lines)
					lines = append(lines, fmt.Sprintf("ldr x4,[x3],#%d", pointerDelta), update, "cbnz x1,#-8", "ret")
					flow := arm64PoolTestFlow(t, lines)
					if got := flow.affineInterval(head, arm64PoolRegisterExpression(3)); got != want {
						t.Fatalf("carried address=%+v, want %+v", got, want)
					}
				})
			}
		}
	}
}

func TestARM64PoolCarriedStrideRejectsUnprovedRelation(t *testing.T) {
	for _, change := range []string{
		"sub x3,x3,#7", // Non-integral address/counter ratio.
		"add x3,x3,x5", // Unknown per-iteration delta.
		"ldr x3,[x5]",  // No affine recurrence.
	} {
		flow := arm64PoolTestFlow(t, []string{
			"and x1,x0,#8", "add x1,x1,#8", "lsl x3,x1,#3",
			"nop", change, "sub x1,x1,#8", "cbnz x1,#-12", "ret",
		})
		if got := flow.affineInterval(3, arm64PoolRegisterExpression(3)); got != arm64PoolUnknownInterval {
			t.Fatalf("%s invented a carried bound: %+v", change, got)
		}
	}
}

func TestARM64PoolCarriedAddressCounterRelation(t *testing.T) {
	for _, test := range []struct {
		name          string
		initial, load string
		want          arm64PoolInterval
	}{
		{"decreasing", "mov x3, #64", "ldr x4, [x3], #-8", arm64PoolInterval{8, 64}},
		{"increasing", "mov x3, #0", "ldr x4, [x3], #8", arm64PoolInterval{0, 56}},
	} {
		t.Run(test.name, func(t *testing.T) {
			flow := arm64PoolTestFlow(t, []string{
				test.initial, "mov x1, #0", "mov x2, #8", test.load,
				"add x1, x1, #1", "cmp x1, x2", "b.lt #-12", "ret",
			})
			if got := flow.affineInterval(3, arm64PoolRegisterExpression(3)); got != test.want {
				t.Fatalf("carried address = %+v, want %+v", got, test.want)
			}
		})
	}
	flow := arm64PoolTestFlow(t, []string{
		"mov x3, #64", "mov x1, #0", "mov x2, #8", "ldr x4, [x3], #-8",
		"add x3, x3, x5", "add x1, x1, #1", "cmp x1, x2", "b.lt #-16", "ret",
	})
	if got := flow.affineInterval(3, arm64PoolRegisterExpression(3)); got != arm64PoolUnknownInterval {
		t.Fatalf("nonconstant address recurrence accepted: %+v", got)
	}
}
