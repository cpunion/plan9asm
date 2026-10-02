package main

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/xgo-dev/plan9asm/internal/gotoolprofile"
)

// This tests only the private comparison projection, not proof acceptance.
// The production CLI tests independently assemble real original data objects.
func TestDataOnlySemanticProjectionRebindsPhysicalNativeIdentity(t *testing.T) {
	progress := fixtureSemanticProgress(t)
	selection := progress.Candidates[0].FeatureConsumption[0]
	cpp := &selection.CPP[0]
	cpp.DataOnlyAssembly = &gotoolprofile.DataOnlyAssemblyProof{
		Native: gotoolprofile.EmptyAssemblyProof{
			Protocol: gotoolprofile.DataOnlyAssemblyProtocol,
			File:     cpp.File, ProfileID: selection.ProfileID,
			SourceSHA256: strings.Repeat("1", 64), ListingSHA256: strings.Repeat("2", 64),
			AsmToolSHA256: strings.Repeat("3", 64), ObjectSHA256: strings.Repeat("4", 64),
		},
		Symbols: []gotoolprofile.DataOnlySymbol{{Name: "pkg.data", Kind: "SDATA", Size: 8}},
	}
	before, _ := json.Marshal(progress)
	projected, err := semanticAssemblyLedgerProgress(progress)
	if err != nil {
		t.Fatal(err)
	}
	after, _ := json.Marshal(progress)
	if string(before) != string(after) {
		t.Fatal("comparison mutated original native evidence")
	}
	witness := projected.Candidates[0].FeatureConsumption[0].CPP[0].DataOnlyAssembly
	if witness.Native.ProfileID == selection.ProfileID || witness.Native.AsmToolSHA256 != "" || witness.Native.ObjectSHA256 != "" {
		t.Fatal("host-bound data-only identity survived the private comparison projection")
	}
	if witness.Native.SourceSHA256 != cpp.DataOnlyAssembly.Native.SourceSHA256 ||
		witness.Native.ListingSHA256 != cpp.DataOnlyAssembly.Native.ListingSHA256 || witness.Symbols[0] != cpp.DataOnlyAssembly.Symbols[0] {
		t.Fatal("comparison discarded data source, listing or symbol semantics")
	}
}
