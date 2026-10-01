package main

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
		for _, decision := range replayed {
			decisions = append(decisions, ordinaryProfileDecision{profile.ID, decision})
		}
		for key := range keys {
			eligible[discoveryProfileScope{key.File, key.Target, profile.ID, key.Tags}] = true
		}
	}
	return decisions, eligible, nil
}
