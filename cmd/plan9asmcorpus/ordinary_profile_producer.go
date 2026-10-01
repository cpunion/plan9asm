package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

type discoveryProfileCaptureOptions struct {
	Markers string
	Cache   *discoveryFeatureObservationCache
}

// First observe source-tag profiles, then capture CPP from their selected file
// union. Capturing only a host/baseline union would lose feature-only assembly
// before its CPP predicates can contribute additional required profiles.
func captureDiscoveryOrdinaryProfiles(ctx context.Context, candidate discoveryCandidate, download moduleDownloadInfo, targets []string, dir string, env []string, options ...discoveryProfileCaptureOptions) (*discoveryOrdinarySelectionPlan, []discoveryFeatureProfile, []discoveryBuildConfiguration, []discoverySourceNotApplicableItem, string, error) {
	if len(options) > 1 || len(options) == 1 && (options[0].Cache == nil || !filepath.IsAbs(options[0].Markers)) {
		return nil, nil, nil, nil, "", fmt.Errorf("ordinary profiles require one owned observation cache")
	}
	plan, err := captureOrdinarySelectionInputs(candidate, download.Dir, targets)
	if err != nil {
		return nil, nil, nil, nil, "", err
	}
	if err := verifyOrdinarySelectionZIP(plan, download.Zip, download.Path, download.Version, download.Sum); err != nil {
		return nil, nil, nil, nil, "", err
	}
	binary, err := exec.LookPath("go")
	if err != nil {
		return nil, nil, nil, nil, "", err
	}
	binary, err = filepath.Abs(binary)
	if err != nil {
		return nil, nil, nil, nil, "", err
	}
	markers := filepath.Join(dir, "actual-feature-markers")
	cache := &discoveryFeatureObservationCache{}
	if len(options) == 1 {
		markers, cache = options[0].Markers, options[0].Cache
	} else if err := os.Mkdir(markers, 0700); err != nil {
		return nil, nil, nil, nil, "", err
	}
	profiles, err := captureDiscoveryFeatureProfiles(ctx, binary, markers, env, plan, candidate.AsmFiles, cache)
	if err != nil {
		return nil, nil, nil, nil, "", err
	}
	_, eligible, err := replayOrdinarySelectionForProfiles(plan, candidate.AsmFiles, profiles)
	if err != nil {
		return nil, nil, nil, nil, "", err
	}
	files := ordinaryProfileEligibleCPPFiles(eligible)
	if len(files) != 0 {
		if err := captureDiscoveryProxyGoMod(ctx, plan, download, dir, env); err != nil {
			return nil, nil, nil, nil, "", err
		}
	}
	var goRoot string
	if len(files) != 0 {
		output, _, err := runDiscoveryMachineCommand(ctx, dir, env, binary, "env", "-json", "GOROOT", "GOVERSION")
		if err != nil {
			return nil, nil, nil, nil, "", err
		}
		var actual map[string]string
		if err := json.Unmarshal(output, &actual); err != nil || len(actual) != 2 || !filepath.IsAbs(actual["GOROOT"]) || actual["GOVERSION"] != plan.GoVersion {
			return nil, nil, nil, nil, "", fmt.Errorf("actual CPP root/Go version differs from the source producer")
		}
		goRoot = actual["GOROOT"]
		plan.CPPInputs, err = captureDiscoveryCPPInputs(plan, download.Dir, goRoot, files, ctx)
		if err != nil {
			return nil, nil, nil, nil, "", err
		}
		if err := verifyDiscoveryCPPModuleZIP(plan.CPPInputs, plan, download.Zip); err != nil {
			return nil, nil, nil, nil, "", err
		}
		profiles, err = captureDiscoveryFeatureProfiles(ctx, binary, markers, env, plan, candidate.AsmFiles, cache)
		if err != nil {
			return nil, nil, nil, nil, "", err
		}
	}
	decisions, eligible, err := replayOrdinarySelectionForProfiles(plan, candidate.AsmFiles, profiles)
	if err != nil {
		return nil, nil, nil, nil, "", err
	}
	if !equalDiscoveryStrings(files, ordinaryProfileEligibleCPPFiles(eligible)) {
		return nil, nil, nil, nil, "", fmt.Errorf("CPP-required profiles changed the captured assembly file union")
	}
	plan.ProfileDecisions = decisions
	configs := ordinaryProfileConfigurations(eligible)
	var rejected []discoverySourceNotApplicableItem
	for _, decision := range decisions {
		if decision.Kind == discoverySourceNotApplicableNoGoPackage {
			rejected = append(rejected, discoverySourceNotApplicableItem{
				ProfileID: decision.ProfileID, BuildTags: decision.BuildTags,
				AsmFiles: decision.AsmFiles, Targets: decision.Targets,
				Kind: decision.Kind, Reason: decision.Diagnostic,
			})
		}
	}
	if err := verifyDiscoveryOrdinaryCPP(plan, download.Dir, goRoot, candidate); err != nil {
		return nil, nil, nil, nil, "", err
	}
	return plan, profiles, configs, rejected, goRoot, nil
}

func ordinaryProfileEligibleCPPFiles(scopes map[discoveryProfileScope]bool) []string {
	files := make(map[string]bool)
	for scope := range scopes {
		files[scope.File] = true
	}
	return sortedDiscoverySet(files)
}

func ordinaryProfileConfigurations(scopes map[discoveryProfileScope]bool) []discoveryBuildConfiguration {
	byKey := make(map[discoveryProfileScope][]string)
	for scope := range scopes {
		file := scope.File
		scope.File = ""
		byKey[scope] = append(byKey[scope], file)
	}
	keys := make([]discoveryProfileScope, 0, len(byKey))
	for key := range byKey {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		a, b := keys[i], keys[j]
		if a.Target != b.Target {
			return a.Target < b.Target
		}
		if a.ProfileID != b.ProfileID {
			return a.ProfileID < b.ProfileID
		}
		return a.Tags < b.Tags
	})
	var configs []discoveryBuildConfiguration
	for _, key := range keys {
		var tags []string
		if key.Tags != "" {
			tags = strings.Split(key.Tags, "\x00")
		}
		configs = append(configs, discoveryBuildConfiguration{
			ProfileID: key.ProfileID, BuildTags: tags, Targets: []string{key.Target},
			AsmFiles: uniqueSortedDiscoveryStrings(byKey[key]),
		})
	}
	return configs
}

func ordinaryProfileEnvironment(base []string, observed *discoveryTargetFeatures) []string {
	replacements := map[string]string{
		"GOTOOLCHAIN": "local", "GOWORK": "off", "GOFLAGS": "", "GOPACKAGESDRIVER": "off",
	}
	for key, value := range observed.Environment {
		if key != "GOVERSION" {
			replacements[key] = value
		}
	}
	return replaceEnv(base, replacements)
}

func verifyDiscoveryProfileCurrent(ctx context.Context, markers string, env []string, profile discoveryFeatureProfile, cache *discoveryFeatureObservationCache) error {
	binary, err := exec.LookPath("go")
	if err != nil {
		return &discoverySourceProofError{err}
	}
	binary, err = filepath.Abs(binary)
	if err != nil {
		return &discoverySourceProofError{err}
	}
	current, err := cache.observe(ctx, binary, markers, env, profile.Request.Target, profile.Request.Overrides)
	if err != nil {
		return &discoverySourceProofError{err}
	}
	if discoveryFeatureProfileID(current) != profile.ID {
		return &discoverySourceProofError{fmt.Errorf("actual driver/source/tools/environment changed for profile %s", profile.ID)}
	}
	return nil
}

func profileObservation(profile *discoveryFeatureProfile) *discoveryTargetFeatures {
	if profile == nil {
		return nil
	}
	return profile.Observed
}
