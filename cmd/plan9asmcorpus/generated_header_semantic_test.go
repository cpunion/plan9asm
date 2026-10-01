package main

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/xgo-dev/plan9asm/internal/gotoolprofile"
)

func TestGeneratedHeaderSemanticComparisonRetainsCompleteOriginAndDefinitions(t *testing.T) {
	stored := fixtureSemanticGeneratedProgress(t)
	current := cloneSemanticProgress(t, stored)
	moveSemanticFixtureTools(&current)
	for index := range current.Candidates {
		candidate := &current.Candidates[index]
		if candidate.OrdinarySelectionPlan == nil {
			continue
		}
		for index := range candidate.OrdinarySelectionPlan.GeneratedHeaders {
			query := &candidate.OrdinarySelectionPlan.GeneratedHeaders[index]
			for _, reference := range candidate.FeatureProfiles {
				if reference.Request.Target == query.Target {
					query.ProfileID = reference.ID
				}
			}
			rebindGeneratedSemanticFixture(query.Metadata, query.ProfileID, current.FeatureInventory)
		}
		for _, proof := range candidate.FeatureConsumption {
			rebindGeneratedSemanticFixture(proof.GeneratedHeaders, proof.ProfileID, current.FeatureInventory)
		}
	}
	for _, progress := range []discoveryProgress{stored, current} {
		if err := requireVerifiedAssemblyLedger(progress); err != nil {
			t.Fatalf("both original snapshots must independently authenticate complete metadata: %v", err)
		}
	}
	if err := compareAssemblyLedgerProgress(stored, current); err != nil {
		t.Fatalf("physical compiler/header/object/export bytes must not break validated semantic comparison: %v", err)
	}
	for _, side := range []string{"producer", "consumer"} {
		for _, field := range []string{"export", "compile", "missing header", "missing object", "ImportMap", "value", "presence"} {
			t.Run(side+"/"+field, func(t *testing.T) {
				altered := cloneSemanticProgress(t, stored)
				for index := range altered.Candidates {
					candidate := &altered.Candidates[index]
					if candidate.OrdinarySelectionPlan == nil || len(candidate.OrdinarySelectionPlan.GeneratedHeaders) == 0 {
						continue
					}
					metadata := candidate.OrdinarySelectionPlan.GeneratedHeaders[0].Metadata
					if side == "consumer" {
						metadata = candidate.FeatureConsumption[0].GeneratedHeaders
					}
					header := &metadata.Headers[0]
					switch field {
					case "export":
						header.Imports[0].ExportSHA256 = strings.Repeat("2", 64)
					case "compile":
						header.CompileToolSHA256 = strings.Repeat("2", 64)
					case "missing header":
						header.HeaderSHA256 = ""
					case "missing object":
						header.ObjectSHA256 = ""
					case "ImportMap":
						header.ImportMap["fmt"] = "errors"
					default:
						mutateGeneratedSemanticMeaning(metadata, field)
					}
					break
				}
				if err := requireVerifiedAssemblyLedger(altered); err == nil {
					t.Fatal("one-sided tool/import/full-header alteration passed original-snapshot authentication")
				}
				if err := compareAssemblyLedgerProgress(stored, altered); err == nil {
					t.Fatal("cross-host projection bypassed the altered original's strict validation")
				}
			})
		}
	}
	for _, field := range []string{"value", "presence", "import source", "language", "import mapping"} {
		t.Run(field, func(t *testing.T) {
			altered := cloneSemanticProgress(t, current)
			for index := range altered.Candidates {
				candidate := &altered.Candidates[index]
				if candidate.OrdinarySelectionPlan == nil || len(candidate.OrdinarySelectionPlan.GeneratedHeaders) == 0 {
					continue
				}
				for queryIndex := range candidate.OrdinarySelectionPlan.GeneratedHeaders {
					mutateGeneratedSemanticMeaning(candidate.OrdinarySelectionPlan.GeneratedHeaders[queryIndex].Metadata, field)
				}
				for _, proof := range candidate.FeatureConsumption {
					mutateGeneratedSemanticMeaning(proof.GeneratedHeaders, field)
					canonical, err := gotoolprofile.CanonicalGeneratedHeader(proof.GeneratedHeaders.Headers[0].Definitions)
					if err != nil {
						t.Fatal(err)
					}
					proof.CPP[0].Inputs["generated/"+candidate.Module+"/go_asm.h"] = discoveryFeatureBytesSHA256(canonical)
				}
			}
			if err := requireVerifiedAssemblyLedger(altered); err != nil {
				t.Fatalf("self-consistent alternate semantic fixture should validate locally: %v", err)
			}
			if err := compareAssemblyLedgerProgress(stored, altered); err == nil {
				t.Fatal("cross-host comparison discarded full definition presence/value or dependency source identity")
			}
		})
	}
}

// This portable protocol fixture is deliberately synthetic, not Go/LLVM
// execution evidence. Actual producer/consumer compilation has separate tests.
func fixtureSemanticGeneratedProgress(t *testing.T) discoveryProgress {
	t.Helper()
	progress := fixtureSemanticProgress(t)
	for index, candidate := range progress.Candidates {
		if candidate.Status != discoveryStatusPassed || candidate.OrdinarySelectionPlan == nil {
			continue
		}
		const assembly = "#include \"go_asm.h\"\nTEXT ·Probe(SB),$0-0\nRET\n"
		scanned := discoveryCandidate{Module: candidate.Module, Version: candidate.Version, AsmFiles: []string{"probe.s"}}
		plan := fixtureOrdinarySelectionForCandidate(t, scanned, progress.Targets, map[string]string{
			"go.mod":  "module " + candidate.Module + "\n\ngo 1.20\n",
			"decl.go": "package fixture\nimport \"fmt\"\nvar _ = fmt.Sprintf\nconst Answer = 7\nfunc Probe()\n",
			"probe.s": assembly,
		})
		result := discoveryCorpusResult{
			Module: candidate.Module, Version: candidate.Version, Status: discoveryStatusPassed,
			DiscoveredAsmFiles: scanned.AsmFiles, ApplicableAsmFiles: scanned.AsmFiles,
			Translations: len(progress.Targets), OrdinarySelectionPlan: plan,
			BuildConfigurations: []discoveryBuildConfiguration{{AsmFiles: scanned.AsmFiles, Targets: progress.Targets}},
		}
		fixtureProfileEvidence(t, &result, progress.FeatureInventory)
		plan.CPPInputs.Protocol = discoveryDeferredCPPInputsProtocol
		plan.CPPInputs.Units[0].DeferredIncludes = map[string]string{"module/probe.s#0": "generated_go_asm"}
		for _, proof := range result.FeatureConsumption {
			observed := progress.FeatureInventory.Observations[proof.ProfileID]
			definitions := map[string]string{"const_Answer": "7"}
			metadata := &gotoolprofile.MetadataProof{
				Protocol: gotoolprofile.MetadataProtocol, ProfileID: proof.ProfileID, Packages: proof.Packages,
				Headers: []gotoolprofile.GeneratedHeaderProof{{
					Protocol: gotoolprofile.GeneratedHeaderProtocol, PackagePath: candidate.Module,
					ProfileID: proof.ProfileID, Target: observed.Target, GoVersion: observed.GoVersion, LanguageVersion: "go1.20",
					CompileToolSHA256: observed.ToolBinarySHA256["compile"], HeaderSHA256: strings.Repeat("a", 64), ObjectSHA256: strings.Repeat("b", 64),
					Definitions: definitions, DefinitionsSHA256: gotoolprofile.HeaderDefinitionSHA256(definitions),
					CompiledGoSHA256: map[string]string{"decl.go": proof.Packages[0].SourceSHA256["decl.go"]},
					ImportMap:        map[string]string{"fmt": "fmt"}, Imports: []gotoolprofile.ImportExport{{
						ImportPath: "fmt", PackagePath: "fmt", SourceRole: "stdlib", ExportSHA256: strings.Repeat("c", 64),
						GoSHA256: map[string]string{"fmt/print.go": strings.Repeat("d", 64)},
					}},
				}},
			}
			canonical, err := gotoolprofile.CanonicalGeneratedHeader(definitions)
			if err != nil {
				t.Fatal(err)
			}
			proof.GeneratedHeaders = metadata
			proof.CPP[0].Inputs["generated/"+candidate.Module+"/go_asm.h"] = discoveryFeatureBytesSHA256(canonical)
			// The producer and independent consumer do not share mutable maps.
			data, _ := json.Marshal(metadata)
			var original gotoolprofile.MetadataProof
			if err := json.Unmarshal(data, &original); err != nil {
				t.Fatal(err)
			}
			plan.GeneratedHeaders = append(plan.GeneratedHeaders, discoveryGeneratedHeaderQuery{
				Target: observed.Target, ProfileID: proof.ProfileID, AsmFiles: scanned.AsmFiles, Metadata: &original,
			})
		}
		progress.Translations += result.Translations - candidate.Translations
		progress.Candidates[index] = discoveryCandidateProgress{
			Module: result.Module, Version: result.Version, Status: result.Status, Translations: result.Translations,
			DiscoveredAsmFiles: result.DiscoveredAsmFiles, ApplicableAsmFiles: result.ApplicableAsmFiles,
			BuildConfigurations: result.BuildConfigurations, OrdinarySelectionPlan: plan,
			FeatureProfiles: result.FeatureProfiles, FeatureConsumption: result.FeatureConsumption,
		}
		if err := requireVerifiedAssemblyLedger(progress); err != nil {
			t.Fatalf("portable full generated metadata fixture: %v", err)
		}
		return progress
	}
	t.Fatal("no ordinary pass to bind the portable metadata fixture")
	return progress
}

func rebindGeneratedSemanticFixture(metadata *gotoolprofile.MetadataProof, id string, inventory *discoveryFeatureInventory) {
	if metadata == nil {
		return
	}
	metadata.ProfileID = id
	for index := range metadata.Packages {
		metadata.Packages[index].Macros.FeatureID = id
	}
	for index := range metadata.Headers {
		header := &metadata.Headers[index]
		header.ProfileID, header.CompileToolSHA256 = id, inventory.Observations[id].ToolBinarySHA256["compile"]
		header.HeaderSHA256, header.ObjectSHA256 = strings.Repeat("e", 64), strings.Repeat("f", 64)
		for index := range header.Imports {
			header.Imports[index].ExportSHA256 = strings.Repeat("a", 64)
		}
	}
}

func mutateGeneratedSemanticMeaning(metadata *gotoolprofile.MetadataProof, field string) {
	for index := range metadata.Headers {
		header := &metadata.Headers[index]
		switch field {
		case "value":
			header.Definitions["const_Answer"] = "9"
		case "presence":
			header.Definitions["const_Additional"] = "1"
		case "import source":
			header.Imports[0].GoSHA256["fmt/print.go"] = strings.Repeat("1", 64)
		case "language":
			header.LanguageVersion = "go1.21"
		case "import mapping":
			header.ImportMap = map[string]string{"errors": "errors"}
			header.Imports[0].ImportPath, header.Imports[0].PackagePath = "errors", "errors"
			header.Imports[0].GoSHA256 = map[string]string{"errors/errors.go": strings.Repeat("d", 64)}
		}
		header.DefinitionsSHA256 = gotoolprofile.HeaderDefinitionSHA256(header.Definitions)
	}
}
