package main

import "testing"

func TestCPPConditionFreeIndexAvoidsRepeatedLexing(t *testing.T) {
	plan, _, _ := cppPlanningSourceIndexFixture(t, 1, 1)
	unit := plan.CPPInputs.Units[0]
	measure := func(index *discoveryCPPPlanningSourceIndex) float64 {
		return testing.AllocsPerRun(1, func() {
			free, err := discoveryCPPConditionFreeWithIndex(plan, unit, nil, index)
			if err != nil || !free {
				t.Fatalf("complete source query changed: free=%t error=%v", free, err)
			}
		})
	}
	uncached := measure(nil)
	warm := measure(&discoveryCPPPlanningSourceIndex{})
	t.Logf("complete condition-free query allocations: uncached=%g warm=%g", uncached, warm)
	// AllocsPerRun first warms the private index. Hashes, public projections
	// and graph relationships must still be checked, but repeated complete
	// lexical scans must not dominate that already-observed immutable text.
	if uncached == 0 || warm*4 >= uncached {
		t.Fatalf("warm source index still repeats lexical inspection: %g versus %g", warm, uncached)
	}
}
