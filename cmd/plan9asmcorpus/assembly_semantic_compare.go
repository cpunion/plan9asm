package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/build/constraint"
	"sort"
	"strings"
)

// This private comparison view is never evidence and cannot pass a report or
// ledger reader: physical tool identities are deliberately absent. Call only
// after both original snapshots independently pass requireVerifiedAssemblyLedger.
// All unlisted fields stay intact, so a new proof field is compared by default.
func semanticAssemblyLedgerProgress(original discoveryProgress) (discoveryProgress, error) {
	var result discoveryProgress
	encoded, err := json.Marshal(original)
	if err != nil {
		return result, err
	}
	if err := json.Unmarshal(encoded, &result); err != nil {
		return result, err
	}
	result.Source.Revision = ""
	result.Provenance = &discoveryCorpusProvenance{
		GoVersion:    original.Provenance.GoVersion,
		TranslatorGo: original.Provenance.TranslatorGo,
		// Each original already matched the anchored LLVM 22.x.y identity
		// rule. Only the cross-host contract discards patch differences.
		LLVMVersion: "22",
	}
	identities := make(map[string]string)
	inventory := newDiscoveryFeatureInventory()
	inventory.Protocol = "semantic_go_feature_comparison_v1"
	for physicalID, observed := range result.FeatureInventory.Observations {
		observed.DriverSHA256 = ""
		observed.ToolDirectory, observed.ToolRoutingSHA256 = "", ""
		observed.ToolBinarySHA256, observed.ToolBinaryOrigins = nil, nil
		observed.EnvStderrSHA256, observed.EnvRecheckStderrSHA256, observed.ListStderrSHA256 = "", "", ""
		// Target environment, version, registered sources/dispatcher, marker
		// bytes and selected builtin tags remain part of this semantic identity.
		semanticID := discoveryFeatureProfileID(observed)
		identities[physicalID] = semanticID
		if previous := inventory.Observations[semanticID]; previous != nil {
			before, err := json.Marshal(previous)
			if err != nil {
				return result, err
			}
			after, err := json.Marshal(observed)
			if err != nil || !bytes.Equal(before, after) {
				return result, fmt.Errorf("conflicting semantic profile identity")
			}
		} else {
			inventory.Observations[semanticID] = observed
		}
	}
	result.FeatureInventory = inventory
	replaceID := func(id *string) error {
		if *id == "" {
			return nil // Separately validated exception protocols have no profile.
		}
		mapped, found := identities[*id]
		if !found {
			return fmt.Errorf("profile comparison has an unbound physical identity")
		}
		*id = mapped
		return nil
	}
	for index := range result.Candidates {
		candidate := &result.Candidates[index]
		for index := range candidate.FeatureProfiles {
			if err := replaceID(&candidate.FeatureProfiles[index].ID); err != nil {
				return result, err
			}
		}
		for index := range candidate.BuildConfigurations {
			if err := replaceID(&candidate.BuildConfigurations[index].ProfileID); err != nil {
				return result, err
			}
		}
		if plan := candidate.OrdinarySelectionPlan; plan != nil {
			if metadata := plan.ProxyGoMod; metadata != nil {
				// Each original inclusion/signature was verified first. Tree
				// checkpoints may advance between hosts; exact authenticated
				// record, ZIP/mod sums, bytes and declared origin stay intact.
				metadata.SignedTree, metadata.Inclusion = "", nil
			}
			// This legacy host default is not used by schema10 replay. Its
			// actual target ToolTags remain in every required observation.
			plan.ToolTags = nil
			for index := range plan.ProfileDecisions {
				if err := replaceID(&plan.ProfileDecisions[index].ProfileID); err != nil {
					return result, err
				}
			}
			if err := sortSemanticRows(plan.ProfileDecisions); err != nil {
				return result, err
			}
		}
		if plan := candidate.NativeLayoutPlan; plan != nil {
			if err := keepReferencedNativeToolTags(plan); err != nil {
				return result, err
			}
			for selectionIndex := range plan.Selections {
				for decisionIndex := range plan.Selections[selectionIndex].Decisions {
					decision := &plan.Selections[selectionIndex].Decisions[decisionIndex]
					decision.Reason, err = semanticSourceReason(decision.Reason)
					if err != nil {
						return result, err
					}
				}
			}
		}
		for index := range candidate.SourceNotApplicableItems {
			item := &candidate.SourceNotApplicableItems[index]
			if err := replaceID(&item.ProfileID); err != nil {
				return result, err
			}
			if item.Diagnostic != nil {
				item.Diagnostic.SHA256 = "" // Keep protocol and all source positions.
			}
		}
		for index := range candidate.NotApplicableItems {
			if err := replaceID(&candidate.NotApplicableItems[index].ProfileID); err != nil {
				return result, err
			}
		}
		for _, proof := range candidate.FeatureConsumption {
			if err := replaceID(&proof.ProfileID); err != nil {
				return result, err
			}
			for index := range proof.Packages {
				if err := replaceID(&proof.Packages[index].Macros.FeatureID); err != nil {
					return result, err
				}
			}
			for index := range proof.Outputs {
				proof.Outputs[index].IR, proof.Outputs[index].Object = "", ""
			}
			for index := range proof.CPP {
				if empty := proof.CPP[index].EmptyAssembly; empty != nil {
					if err := replaceID(&empty.ProfileID); err != nil {
						return result, err
					}
					empty.AsmToolSHA256, empty.ObjectSHA256 = "", ""
				}
			}
			if err := sortSemanticRows(proof.Packages); err != nil {
				return result, err
			}
			if err := sortSemanticRows(proof.CPP); err != nil {
				return result, err
			}
			if err := sortSemanticRows(proof.Outputs); err != nil {
				return result, err
			}
		}
		for _, err := range []error{
			sortSemanticRows(candidate.FeatureProfiles), sortSemanticRows(candidate.BuildConfigurations),
			sortSemanticRows(candidate.SourceNotApplicableItems), sortSemanticRows(candidate.NotApplicableItems),
			sortSemanticRows(candidate.FeatureConsumption),
		} {
			if err != nil {
				return result, err
			}
		}
	}
	return result, nil
}

// Retain every true feature referenced by any saved native constraint, even
// when an OR expression leaves selection unchanged. An absent referenced tag
// remains false. Only demonstrably unused host tags may disappear.
func keepReferencedNativeToolTags(plan *discoveryNativeLayoutPlan) error {
	referenced := make(map[string]bool)
	for _, source := range plan.SourceInventory {
		for _, line := range strings.Split(source.Header, "\n") {
			line = strings.TrimSpace(line)
			if !strings.HasPrefix(line, "//go:build ") && !strings.HasPrefix(line, "// +build ") {
				continue
			}
			expression, err := constraint.Parse(line)
			if err != nil {
				return fmt.Errorf("native comparison cannot parse original constraint: %w", err)
			}
			collectDiscoveryConstraintTags(expression, referenced)
		}
	}
	var retained []string
	for _, tag := range plan.ToolTags {
		if referenced[tag] {
			retained = append(retained, tag)
		}
	}
	plan.ToolTags = uniqueSortedDiscoveryStrings(retained)
	return nil
}

func semanticSourceReason(reason string) (string, error) {
	witness := captureDiscoverySourceDiagnostic(reason)
	if witness == nil {
		return reason, nil // Never erase unclassified/selection diagnostics.
	}
	witness.SHA256 = ""
	data, err := json.Marshal(witness)
	return string(data), err
}

func sortSemanticRows[T any](rows []T) error {
	keys := make([][]byte, len(rows))
	indices := make([]int, len(rows))
	for index := range rows {
		encoded, err := json.Marshal(rows[index])
		if err != nil {
			return err
		}
		keys[index], indices[index] = encoded, index
	}
	sort.Slice(indices, func(i, j int) bool { return bytes.Compare(keys[indices[i]], keys[indices[j]]) < 0 })
	ordered := make([]T, len(rows))
	for index, original := range indices {
		ordered[index] = rows[original]
	}
	copy(rows, ordered)
	return nil
}
