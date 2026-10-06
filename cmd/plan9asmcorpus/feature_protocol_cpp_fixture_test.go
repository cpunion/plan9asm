package main

import (
	"encoding/json"
	"testing"
)

// These are portable accounting fixtures, not external-module object or
// runtime witnesses. Even synthetic accounting must preserve full CPP inputs.
func TestPortableAccountingFixtureRetainsCurrentOriginalCPP(t *testing.T) {
	candidate := discoveryCandidate{
		Module: "example.com/cpp-accounting", Version: "v1.0.0",
		AsmFiles: []string{"probe.s"},
	}
	result := fixtureOrdinaryPassedResult(t, candidate, []string{"linux/amd64", "linux/arm64"})
	plan := result.OrdinarySelectionPlan
	if plan.CPPSourceProtocol != discoveryReplayCPPInputsProtocol || plan.CPPInputs == nil ||
		plan.CPPInputs.Protocol != discoveryReplayCPPInputsProtocol {
		t.Fatal("fresh accounting fixture silently retained a historical CPP graph")
	}
	for file, source := range plan.CPPInputs.Sources {
		if source.Original == nil || discoveryFeatureBytesSHA256([]byte(*source.Original)) != source.SHA256 {
			t.Fatalf("fresh CPP fixture lost its complete original: %s", file)
		}
	}
	if err := validateOrdinaryProfileResult(result, plan.Targets, plan.GoVersion); err != nil {
		t.Fatalf("fresh original-source accounting failed: %v", err)
	}
	encoded, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	for _, mutation := range []struct {
		name string
		edit func(*discoveryCorpusResult)
	}{
		{
			name: "paired_legacy_downgrade",
			edit: func(copy *discoveryCorpusResult) {
				copy.OrdinarySelectionPlan.CPPSourceProtocol = discoveryCPPInputsProtocol
				copy.OrdinarySelectionPlan.CPPInputs.Protocol = discoveryCPPInputsProtocol
			},
		},
		{
			name: "paired_source_graph_deletion",
			edit: func(copy *discoveryCorpusResult) {
				copy.OrdinarySelectionPlan.CPPSourceProtocol = ""
				copy.OrdinarySelectionPlan.CPPInputs = nil
			},
		},
	} {
		t.Run(mutation.name, func(t *testing.T) {
			var changed discoveryCorpusResult
			if err := json.Unmarshal(encoded, &changed); err != nil {
				t.Fatal(err)
			}
			changed.featureInventory = result.featureInventory
			mutation.edit(&changed)
			if err := validateOrdinaryProfileResult(changed, plan.Targets, plan.GoVersion); err == nil {
				t.Fatal("legacy or deleted original graph was accepted as a current fixture")
			}
		})
	}
}
