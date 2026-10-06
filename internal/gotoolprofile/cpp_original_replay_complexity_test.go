package gotoolprofile

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

// Query validation only: this fixture supplies no original-source authority,
// actual compiler object, platform applicability or assembly-ledger success.
func cppOriginalValidationFixture(t testing.TB, units int) *ConsumerInput {
	t.Helper()
	input := &ConsumerInput{
		CPPSourceProtocol: CPPOriginalReplayProtocol,
		Sources:           make(map[string]string), ToolSources: make(map[string]string),
		CPPOriginals: &CPPOriginalReplayInput{
			Sources: make(map[string]string), Units: make(map[string]CPPOriginalUnit),
			RequiredExpandedOperands: make(map[string][]CPPPackageSearchMiss),
		},
	}
	for index := 0; index < units; index++ {
		file := fmt.Sprintf("pkg/native_%03d.s", index)
		text := "TEXT ·Probe(SB),4,$0-0\nRET\n" + strings.Repeat("// complete original bytes\n", 2048)
		digest := sha256.Sum256([]byte(text))
		input.Sources[file] = hex.EncodeToString(digest[:])
		input.CPPOriginals.Sources["module/"+file] = text
		input.CPPOriginals.Units[file] = CPPOriginalUnit{Includes: make(map[string]string)}
		input.AsmFiles = append(input.AsmFiles, file)
	}
	return input
}

func TestCPPOriginalValidationHashesEverySourceOnce(t *testing.T) {
	const units = 16
	input := cppOriginalValidationFixture(t, units)
	allocations := testing.AllocsPerRun(1, func() {
		if err := ValidateCPPOriginalInputs(input); err != nil {
			t.Fatal(err)
		}
	})
	t.Logf("complete source validation allocations: units=%d allocations=%g", units, allocations)
	// Allow three allocations per complete digest and map/key headroom, but
	// reject rehashing each already-validated original from the unit loop.
	if allocations > 3*units+8 {
		t.Fatalf("CPP units rehash independently validated originals: %g > %d", allocations, 3*units+8)
	}
}

func TestCPPOriginalSinglePassValidationRetainsAllResources(t *testing.T) {
	input := cppOriginalValidationFixture(t, 2)
	toolFile, toolText := "pkg/include/inactive.h", "#define INACTIVE 1\n"
	digest := sha256.Sum256([]byte(toolText))
	input.ToolSources[toolFile] = hex.EncodeToString(digest[:])
	input.CPPOriginals.Sources["tool/"+toolFile] = toolText
	encoded, err := json.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateCPPOriginalInputs(input); err != nil {
		t.Fatal(err)
	}
	for _, fixture := range []struct {
		name   string
		mutate func(*ConsumerInput)
	}{
		{"first_original", func(in *ConsumerInput) { in.CPPOriginals.Sources["module/"+in.AsmFiles[0]] += "changed" }},
		{"last_original", func(in *ConsumerInput) { in.CPPOriginals.Sources["module/"+in.AsmFiles[1]] += "changed" }},
		{"hash", func(in *ConsumerInput) { in.Sources[in.AsmFiles[0]] = strings.Repeat("0", 64) }},
		{"missing_original", func(in *ConsumerInput) { delete(in.CPPOriginals.Sources, "module/"+in.AsmFiles[0]) }},
		{"missing_hash", func(in *ConsumerInput) { delete(in.Sources, in.AsmFiles[0]) }},
		{"inactive_tool_original", func(in *ConsumerInput) { in.CPPOriginals.Sources["tool/"+toolFile] += "changed" }},
		{"inactive_tool_hash", func(in *ConsumerInput) { in.ToolSources[toolFile] = strings.Repeat("0", 64) }},
		{"unbound_include", func(in *ConsumerInput) {
			in.CPPOriginals.Units[in.AsmFiles[0]].Includes["missing.h"] = "module/pkg/missing.h"
		}},
	} {
		t.Run(fixture.name, func(t *testing.T) {
			var changed ConsumerInput
			if err := json.Unmarshal(encoded, &changed); err != nil {
				t.Fatal(err)
			}
			fixture.mutate(&changed)
			if err := ValidateCPPOriginalInputs(&changed); err == nil {
				t.Fatal("single-pass validation omitted complete original/resource binding")
			}
		})
	}
}

func BenchmarkCPPOriginalValidation(b *testing.B) {
	input := cppOriginalValidationFixture(b, 16)
	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		if err := ValidateCPPOriginalInputs(input); err != nil {
			b.Fatal(err)
		}
	}
}
