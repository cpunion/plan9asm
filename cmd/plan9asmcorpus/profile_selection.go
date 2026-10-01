package main

import (
	"bytes"
	"fmt"
	"io"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

type discoveryProfileScope struct {
	File, Target, ProfileID, Tags string
}

type ordinaryProfileDecision struct {
	ProfileID string `json:"profile_id"`
	ordinarySelectionDecision
}

func replayOrdinarySelectionForProfiles(plan *discoveryOrdinarySelectionPlan, asmFiles []string, profiles []discoveryFeatureProfile) ([]ordinaryProfileDecision, map[discoveryProfileScope]bool, error) {
	if err := validateDiscoveryFeatureProfiles(plan, asmFiles, profiles); err != nil {
		return nil, nil, err
	}
	var decisions []ordinaryProfileDecision
	eligible := make(map[discoveryProfileScope]bool)
	for _, profile := range profiles {
		copy := *plan
		copy.Targets = []string{profile.Observed.Target}
		copy.ToolTags = append([]string(nil), profile.Observed.ToolTags...)
		replayed, keys, err := replayOrdinarySelection(&copy, asmFiles, profile.Observed)
		if err != nil {
			return nil, nil, err
		}
		replayed, err = closeOrdinaryPackageTagScopes(&copy, asmFiles, replayed, keys, profile.Observed)
		if err != nil {
			return nil, nil, err
		}
		for _, decision := range replayed {
			decisions = append(decisions, ordinaryProfileDecision{profile.ID, decision})
		}
		for key := range keys {
			eligible[discoveryProfileScope{key.File, key.Target, profile.ID, key.Tags}] = true
		}
	}
	return decisions, eligible, nil
}

// A tag configuration selects a whole package, not just the file whose
// constraint proposed those tags. Shared assembly must therefore be checked
// again with the declarations selected by every package-local configuration.
// Do not spread one package's custom tags to unrelated packages.
func closeOrdinaryPackageTagScopes(plan *discoveryOrdinarySelectionPlan, asmFiles []string, decisions []ordinarySelectionDecision, eligible map[nativeLayoutPlanKey]bool, actual *discoveryTargetFeatures) ([]ordinarySelectionDecision, error) {
	type packageScope struct {
		Dir, Tags string
	}
	scopeSet := make(map[packageScope]bool)
	for key := range eligible {
		scopeSet[packageScope{Dir: path.Dir(key.File), Tags: key.Tags}] = true
	}
	var scopes []packageScope
	for scope := range scopeSet {
		scopes = append(scopes, scope)
	}
	sort.Slice(scopes, func(i, j int) bool {
		if scopes[i].Dir != scopes[j].Dir {
			return scopes[i].Dir < scopes[j].Dir
		}
		return scopes[i].Tags < scopes[j].Tags
	})
	sources, _, err := ordinarySelectionBytes(plan)
	if err != nil {
		return nil, err
	}
	ctx, err := discoveryContextForFeatureProfile(actual)
	if err != nil {
		return nil, err
	}
	ctx.OpenFile = func(file string) (io.ReadCloser, error) {
		data, found := sources[filepath.ToSlash(file)]
		if !found {
			return nil, fmt.Errorf("uncaptured package/tag selection source %s", file)
		}
		return io.NopCloser(bytes.NewReader(data)), nil
	}
	for _, scope := range scopes {
		ctx.BuildTags = nil
		if scope.Tags != "" {
			ctx.BuildTags = strings.Split(scope.Tags, "\x00")
		}
		for _, file := range asmFiles {
			if path.Dir(file) != scope.Dir {
				continue
			}
			key := nativeLayoutKey(file, actual.Target, ctx.BuildTags)
			if eligible[key] {
				continue
			}
			selected, err := ctx.MatchFile(filepath.FromSlash(scope.Dir), path.Base(file))
			if err != nil {
				return nil, fmt.Errorf("replay package/tag assembly selection for %s: %w", file, err)
			}
			if !selected {
				continue
			}
			eligible[key] = true
			decisions = append(decisions, ordinarySelectionDecision{
				AsmFiles: []string{file}, Targets: []string{actual.Target},
				BuildTags:  append([]string(nil), ctx.BuildTags...),
				Kind:       nativeLayoutSelected,
				Diagnostic: "Go MatchFile selects this shared assembly in a required package/custom-tag configuration; compiler/assembler acceptance is checked separately",
			})
		}
	}
	return compactOrdinarySelectionDecisions(decisions), nil
}
