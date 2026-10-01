package main

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"reflect"
	"strings"

	"github.com/xgo-dev/plan9asm/internal/gotoolprofile"
)

func validateOrdinaryProfileResult(result discoveryCorpusResult, targets []string, goVersion string) error {
	plan := result.OrdinarySelectionPlan
	if plan == nil || plan.Module != result.Module || plan.Version != result.Version || plan.GoVersion != goVersion ||
		!equalDiscoveryStrings(plan.Targets, uniqueSortedDiscoveryStrings(targets)) {
		return fmt.Errorf("profile-aware ordinary result lacks exact source/module/target/Go inputs")
	}
	if err := validateDiscoverySourceNotApplicableEvidence(result); err != nil {
		return err
	}
	if err := validateDiscoveryProxyGoMod(plan); err != nil {
		return err
	}
	digest, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(plan.ModuleSum, "h1:"))
	if !strings.HasPrefix(plan.ModuleSum, "h1:") || err != nil || len(digest) != sha256.Size || !discoverySHA256Pattern.MatchString(plan.ZipSHA256) {
		return fmt.Errorf("ordinary profile inputs lack exact original ZIP/h1 identity")
	}
	profiles, err := resolveDiscoveryFeatureProfiles(result.featureInventory, result.FeatureProfiles)
	if err != nil {
		return err
	}
	decisions, eligible, err := replayOrdinarySelectionForProfiles(plan, result.DiscoveredAsmFiles, profiles)
	if err != nil {
		return err
	}
	if !equalOrdinaryProfileDecisions(decisions, plan.ProfileDecisions) {
		return fmt.Errorf("source-required actual profile decisions differ from original headers")
	}
	if err := validateDiscoveryGeneratedHeaderQueries(plan, profiles, eligible); err != nil {
		return err
	}
	if len(eligible) != 0 && plan.CPPInputs == nil {
		return fmt.Errorf("selected ordinary profile lacks original CPP source/registration inputs")
	}
	if plan.CPPInputs != nil {
		files := make(map[string]bool)
		for scope := range eligible {
			files[scope.File] = true
		}
		if err := validateDiscoveryCPPInputs(plan.CPPInputs, plan, sortedDiscoverySet(files)); err != nil {
			return err
		}
	}
	executed, err := ordinaryProfileConfigurationKeys(result.BuildConfigurations, eligible)
	if err != nil {
		return err
	}
	if err := validateDiscoveryGeneratedHeaderExecution(plan, executed); err != nil {
		return err
	}
	noPackage := make(map[discoveryProfileScope]bool)
	for _, decision := range decisions {
		if decision.Kind == discoverySourceNotApplicableNoGoPackage {
			for _, file := range decision.AsmFiles {
				for _, target := range decision.Targets {
					noPackage[discoveryProfileScope{File: file, Target: target, ProfileID: decision.ProfileID, Tags: strings.Join(decision.BuildTags, "\x00")}] = true
				}
			}
		}
	}
	rejected := make(map[discoveryProfileScope]bool)
	for _, item := range result.SourceNotApplicableItems {
		files := append(append([]string(nil), item.AsmFiles...), item.AsmFile)
		for _, file := range files {
			if file == "" {
				continue
			}
			for _, target := range item.Targets {
				key := discoveryProfileScope{File: file, Target: target, ProfileID: item.ProfileID, Tags: strings.Join(item.BuildTags, "\x00")}
				valid := eligible[key] && item.Kind != discoverySourceNotApplicableNoGoPackage || noPackage[key] && item.Kind == discoverySourceNotApplicableNoGoPackage
				if !valid || rejected[key] || executed[key] {
					return fmt.Errorf("ordinary source diagnostic claims a missing/duplicate/executed profile scope: %v", key)
				}
				rejected[key] = true
			}
		}
	}
	for key := range eligible {
		if !executed[key] && !rejected[key] {
			return fmt.Errorf("selected source/profile scope was neither consumed nor concretely rejected: %v", key)
		}
	}
	for key := range noPackage {
		if !rejected[key] {
			return fmt.Errorf("ordinary no-package profile lacks its scoped selection evidence: %v", key)
		}
	}
	abi := make(map[discoveryProfileScope]bool)
	targetItems, err := summarizeDiscoveryTargetSkips(discoveryCandidate{Module: result.Module, AsmFiles: result.DiscoveredAsmFiles}, result.NotApplicableItems)
	if err != nil {
		return err
	}
	for _, item := range targetItems {
		key := discoveryProfileScope{File: item.AsmFile, Target: item.Target, ProfileID: item.ProfileID, Tags: strings.Join(item.BuildTags, "\x00")}
		if !executed[key] || abi[key] || item.Kind != targetNotApplicableGoTextArgSize || item.Symbol == "" || item.PkgPath == "" || item.DeclaredArgSize == item.ExpectedArgSize {
			return fmt.Errorf("ordinary ABI exclusion lacks one exact executed profile scope: %v", key)
		}
		abi[key] = true
	}
	profileByID := make(map[string]discoveryFeatureProfile)
	for _, profile := range profiles {
		profileByID[profile.ID] = profile
	}
	consumed := make(map[discoveryProfileScope]bool)
	for _, proof := range result.FeatureConsumption {
		if proof == nil {
			return fmt.Errorf("missing actual compiler-consumption evidence")
		}
		if err := validateOrdinaryProfileCPPConsumption(plan, proof); err != nil {
			return err
		}
		profile, found := profileByID[proof.ProfileID]
		if !found {
			return fmt.Errorf("compiler consumption references an unknown actual profile")
		}
		var files []string
		allowedABI := make(map[string]bool)
		for _, cpp := range proof.CPP {
			key := discoveryProfileScope{File: cpp.File, Target: profile.Observed.Target, ProfileID: proof.ProfileID, Tags: strings.Join(proof.CustomTags, "\x00")}
			if !executed[key] || consumed[key] {
				return fmt.Errorf("compiler consumption claims a duplicate/unexecuted profile scope: %v", key)
			}
			consumed[key] = true
			files = append(files, cpp.File)
			if abi[key] {
				allowedABI[cpp.File] = true
			}
		}
		files = uniqueSortedDiscoveryStrings(files)
		module, err := ordinaryProfileDeclaredModule(plan)
		if err != nil {
			return err
		}
		input := ordinaryProfileConsumerInput(plan, profile, module, "", files, proof.CustomTags)
		if err := gotoolprofile.ValidateSelectionWithABI(input, proof, proof.CustomTags, true, allowedABI); err != nil {
			return err
		}
	}
	if len(consumed) != len(executed) || result.Translations+result.NotApplicableTranslations != len(executed) ||
		result.NotApplicableTranslations != len(abi) || !equalDiscoveryStrings(result.ApplicableAsmFiles, discoveryConfigurationAsmFiles(result.BuildConfigurations)) {
		return fmt.Errorf("ordinary profile consumption/count/file union differs from its four-part scope denominator")
	}
	return nil
}

func validateOrdinaryProfileCPPConsumption(plan *discoveryOrdinarySelectionPlan, proof *gotoolprofile.SelectionProof) error {
	if plan.CPPInputs == nil {
		return fmt.Errorf("actual CPP consumption has no exact original source graph")
	}
	units := make(map[string]discoveryCPPUnit)
	for _, unit := range plan.CPPInputs.Units {
		units[unit.File] = unit
	}
	for _, cpp := range proof.CPP {
		unit, found := units[cpp.File]
		if !found {
			return fmt.Errorf("actual CPP source has no original graph: %s", cpp.File)
		}
		if plan.CPPInputs.Protocol == discoveryDeferredCPPInputsProtocol {
			var selected *gotoolprofile.PackageProof
			for index := range proof.Packages {
				if containsTargetFeature(proof.Packages[index].SFiles, cpp.File) {
					selected = &proof.Packages[index]
				}
			}
			if selected == nil || selected.Macros == nil {
				return fmt.Errorf("deferred CPP consumer lacks its actual selected package/macros")
			}
			var header *gotoolprofile.GeneratedHeaderProof
			if proof.GeneratedHeaders != nil {
				for index := range proof.GeneratedHeaders.Headers {
					if proof.GeneratedHeaders.Headers[index].PackagePath == selected.PackagePath {
						header = &proof.GeneratedHeaders.Headers[index]
					}
				}
			}
			wanted, err := discoveryDeferredCPPConsumption(plan.CPPInputs, unit, header, selected.Macros.Defines)
			if err != nil || !reflect.DeepEqual(wanted, cpp.Inputs) {
				return fmt.Errorf("actual deferred CPP consumption differs from source-order replay: %s: %v", cpp.File, err)
			}
			continue
		}
		wanted := map[string]string{"module/" + unit.File: plan.CPPInputs.Sources["module/"+unit.File].SHA256}
		for _, included := range unit.Includes {
			wanted[included] = plan.CPPInputs.Sources[included].SHA256
		}
		if !reflect.DeepEqual(wanted, cpp.Inputs) {
			return fmt.Errorf("actual CPP consumption omitted or changed an original include graph: %s", cpp.File)
		}
	}
	return nil
}

func ordinaryProfileConfigurationKeys(configs []discoveryBuildConfiguration, eligible map[discoveryProfileScope]bool) (map[discoveryProfileScope]bool, error) {
	keys := make(map[discoveryProfileScope]bool)
	for _, config := range configs {
		if !discoverySHA256Pattern.MatchString(config.ProfileID) || len(config.Targets) != 1 || len(config.AsmFiles) == 0 ||
			!equalDiscoveryStrings(config.AsmFiles, uniqueSortedDiscoveryStrings(config.AsmFiles)) ||
			!equalDiscoveryStrings(config.BuildTags, uniqueSortedDiscoveryStrings(config.BuildTags)) {
			return nil, fmt.Errorf("ordinary configuration lacks a canonical exact profile/target/file/tag scope")
		}
		for _, tag := range config.BuildTags {
			if gotoolprofile.ReservedTag(tag) {
				return nil, fmt.Errorf("ordinary custom tag claims a reserved frontend feature: %s", tag)
			}
		}
		for _, file := range config.AsmFiles {
			key := discoveryProfileScope{File: file, Target: config.Targets[0], ProfileID: config.ProfileID, Tags: strings.Join(config.BuildTags, "\x00")}
			if keys[key] || !eligible[key] {
				return nil, fmt.Errorf("ordinary execution claims a duplicate/unselected four-part scope: %v", key)
			}
			keys[key] = true
		}
	}
	return keys, nil
}

func ordinaryProfileConsumerInput(plan *discoveryOrdinarySelectionPlan, profile discoveryFeatureProfile, module, sourceRoot string, files []string, customTags ...[]string) *gotoolprofile.ConsumerInput {
	input := &gotoolprofile.ConsumerInput{
		Protocol: gotoolprofile.ConsumerProtocol, ID: profile.ID, Observed: profile.Observed,
		Module: module, Version: plan.Version, SourceModule: plan.Module,
		SourceRoot: sourceRoot, AsmFiles: append([]string(nil), files...),
		Sources: make(map[string]string), Headers: make(map[string]string),
		ToolSources: make(map[string]string), Directories: make(map[string][]string),
		ProxyGoMod: plan.ProxyGoMod, ProxyGoModPath: plan.proxyGoModPath,
	}
	for _, dir := range plan.Directories {
		for _, entry := range dir.Entries {
			input.Directories[dir.Directory] = append(input.Directories[dir.Directory], entry.Name)
			if entry.Kind == "file" {
				input.Sources[strings.TrimPrefix(dir.Directory+"/"+entry.Name, "./")] = entry.SHA256
			}
		}
	}
	for _, source := range plan.Sources {
		input.Headers[source.File] = source.Header
	}
	for file, digest := range plan.GeneratedGoSources {
		if prior := input.Sources[file]; prior != "" && prior != digest {
			input.Sources[file] = "" // Reject conflicting original/source snapshots.
		} else {
			input.Sources[file] = digest
		}
	}
	var tags []string
	if len(customTags) == 1 {
		tags = customTags[0]
	}
	if metadata, err := ordinaryGeneratedMetadata(plan, profile.ID, files, tags); err == nil {
		input.GeneratedHeaders = metadata
	}
	if plan.CPPInputs != nil {
		for id, source := range plan.CPPInputs.Sources {
			if file, tool := strings.CutPrefix(id, "tool/"); tool {
				input.ToolSources[file] = source.SHA256
			}
		}
	}
	return input
}

func ordinaryProfileDeclaredModule(plan *discoveryOrdinarySelectionPlan) (string, error) {
	if plan.ProxyGoMod != nil {
		if err := validateDiscoveryProxyGoMod(plan); err != nil {
			return "", err
		}
		return plan.ProxyGoMod.Module, nil
	}
	for _, source := range plan.Sources {
		if source.File == "go.mod" {
			return parseDeclaredModulePath([]byte(source.Header))
		}
	}
	return "", fmt.Errorf("ordinary profile consumer lacks the original declared module path; legacy metadata needs a separate exact proof")
}

func equalOrdinaryProfileDecisions(left, right []ordinaryProfileDecision) bool {
	return reflect.DeepEqual(left, right)
}
