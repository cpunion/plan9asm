package main

import "github.com/xgo-dev/plan9asm/internal/gotoolprofile"

func validateDiscoveryTargetFeatures(features *discoveryTargetFeatures) error {
	return gotoolprofile.Validate(features)
}
