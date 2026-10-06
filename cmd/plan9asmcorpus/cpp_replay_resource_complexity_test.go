package main

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

// This fixture measures query-graph construction, not source authentication or
// assembly coverage. Increasing independent units must not multiply the lexer
// work for every other unit's complete original source.
func cppResourceComplexityFixture(t testing.TB, units int) *discoveryCPPInputs {
	t.Helper()
	inputs := &discoveryCPPInputs{
		Protocol: discoveryReplayCPPInputsProtocol,
		Sources:  make(map[string]discoveryCPPSource),
	}
	for index := 0; index < units; index++ {
		file := fmt.Sprintf("pkg/native_%03d.s", index)
		text := fmt.Sprintf("TEXT ·Probe%d(SB),4,$0-0\nRET\n", index) +
			strings.Repeat("DATA ·table(SB)/8,$0x0123456789abcdef\n", 1024)
		source, err := discoveryReplayCPPSource(file, []byte(text))
		if err != nil {
			t.Fatal(err)
		}
		inputs.Sources["module/"+file] = source
		inputs.Units = append(inputs.Units, discoveryCPPUnit{File: file})
	}
	return inputs
}

func TestCPPOriginalConsumerResourcePlanningScalesLinearly(t *testing.T) {
	small := cppResourceComplexityFixture(t, 2)
	large := cppResourceComplexityFixture(t, 16)
	measure := func(inputs *discoveryCPPInputs) float64 {
		return testing.AllocsPerRun(1, func() {
			consumer, err := discoveryCPPOriginalConsumerInput(inputs)
			if err != nil {
				t.Fatal(err)
			}
			if len(consumer.CPPOriginals.Sources) != len(inputs.Sources) ||
				len(consumer.CPPOriginals.Units) != len(inputs.Units) ||
				len(consumer.AsmFiles) != len(inputs.Units) {
				t.Fatal("resource planning discarded complete sources or units")
			}
		})
	}
	smallAllocs, largeAllocs := measure(small), measure(large)
	t.Logf("query graph allocations: 2 units=%g; 16 units=%g", smallAllocs, largeAllocs)
	// Eight times the source/units allows twelve times the allocations. This
	// has headroom for maps, but rejects per-unit rescans (quadratic growth).
	if smallAllocs == 0 || largeAllocs > 12*smallAllocs {
		t.Fatalf("CPP resource lexer work is multiplied by every unit: %g > 12 * %g", largeAllocs, smallAllocs)
	}
}

func TestCPPOriginalConsumerResourcePlanningRemainsFreshAndIndependent(t *testing.T) {
	inputs := cppResourceComplexityFixture(t, 2)
	before, err := discoveryCPPOriginalConsumerInput(inputs)
	if err != nil {
		t.Fatal(err)
	}
	again, err := discoveryCPPOriginalConsumerInput(inputs)
	if err != nil || !reflect.DeepEqual(before, again) {
		t.Fatalf("unchanged complete query graph changed: %v", err)
	}
	file := inputs.Units[0].File
	changedText := *inputs.Sources["module/"+file].Original + "DATA ·name(SB)/8,$\"missing.h\"\n"
	changed, err := discoveryReplayCPPSource(file, []byte(changedText))
	if err != nil {
		t.Fatal(err)
	}
	inputs.Sources["module/"+file] = changed
	after, err := discoveryCPPOriginalConsumerInput(inputs)
	if err != nil || after.CPPOriginals.Sources["module/"+file] != changedText ||
		before.CPPOriginals.Sources["module/"+file] == changedText {
		t.Fatalf("resource planning reused stale source or aliased an earlier consumer: %v", err)
	}
	after.CPPOriginals.Sources["module/"+file] = "forged"
	if *inputs.Sources["module/"+file].Original != changedText {
		t.Fatal("consumer mutation changed the original query graph")
	}
	missing := inputs.Sources["module/"+file]
	missing.Original = nil
	inputs.Sources["module/"+file] = missing
	if _, err := discoveryCPPOriginalConsumerInput(inputs); err == nil {
		t.Fatal("previous graph construction hid missing complete original bytes")
	}
}

func BenchmarkCPPOriginalConsumerResourcePlanning(b *testing.B) {
	inputs := cppResourceComplexityFixture(b, 16)
	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		if _, err := discoveryCPPOriginalConsumerInput(inputs); err != nil {
			b.Fatal(err)
		}
	}
}
