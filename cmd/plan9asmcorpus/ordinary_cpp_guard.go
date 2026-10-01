package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
)

// Source/driver proof failures are never Go-source incompatibility evidence.
// Keep the type through joined command errors so package-batch fallback cannot
// turn a changed header into a source N/A after a compiler also rejects it.
type discoverySourceProofError struct{ err error }

func (err *discoverySourceProofError) Error() string {
	return "source proof failure: " + err.err.Error()
}
func (err *discoverySourceProofError) Unwrap() error { return err.err }

func captureDiscoveryOrdinaryCPP(ctx context.Context, candidate discoveryCandidate, download moduleDownloadInfo, targets []string, dir string, env []string) (*discoveryOrdinarySelectionPlan, string, error) {
	plan, err := captureOrdinarySelectionPlan(candidate, download.Dir, targets)
	if err != nil {
		return nil, "", err
	}
	if err := verifyOrdinarySelectionZIP(plan, download.Zip, download.Path, download.Version, download.Sum); err != nil {
		return nil, "", err
	}
	_, eligible, err := replayOrdinarySelection(plan, candidate.AsmFiles)
	if err != nil {
		return nil, "", err
	}
	files := ordinarySelectionEligibleCPPFiles(eligible)
	if len(files) == 0 {
		return plan, "", nil
	}
	output, _, err := runDiscoveryMachineCommand(ctx, dir, env, "go", "env", "-json", "GOROOT", "GOVERSION")
	if err != nil {
		return nil, "", fmt.Errorf("observe actual CPP driver source root: %w", err)
	}
	var actual map[string]string
	if err := json.Unmarshal(output, &actual); err != nil || len(actual) != 2 || !filepath.IsAbs(actual["GOROOT"]) || actual["GOVERSION"] != plan.GoVersion {
		return nil, "", fmt.Errorf("CPP source root/Go version differs from the candidate producer")
	}
	inputs, err := captureDiscoveryCPPInputs(plan, download.Dir, actual["GOROOT"], files, ctx)
	if err != nil {
		return nil, "", err
	}
	if err := verifyDiscoveryCPPModuleZIP(inputs, plan, download.Zip); err != nil {
		return nil, "", err
	}
	plan.CPPInputs = inputs
	return plan, actual["GOROOT"], nil
}

func ordinarySelectionEligibleCPPFiles(eligible map[nativeLayoutPlanKey]bool) []string {
	files := make(map[string]bool)
	for key := range eligible {
		files[key.File] = true
	}
	return sortedDiscoverySet(files)
}

func verifyDiscoveryOrdinaryCPP(plan *discoveryOrdinarySelectionPlan, moduleDir, goRoot string, candidate discoveryCandidate) error {
	if plan == nil {
		return nil // independently verified native/private exception protocol
	}
	if err := verifyOrdinarySelectionUnchanged(plan, moduleDir, candidate); err != nil {
		return &discoverySourceProofError{err}
	}
	if plan.CPPInputs != nil {
		if err := verifyDiscoveryCPPInputsUnchanged(plan.CPPInputs, moduleDir, goRoot); err != nil {
			return &discoverySourceProofError{err}
		}
	}
	return nil
}

func runDiscoveryOrdinaryGuarded(plan *discoveryOrdinarySelectionPlan, moduleDir, goRoot string, candidate discoveryCandidate, run func() error) (runErr error) {
	if err := verifyDiscoveryOrdinaryCPP(plan, moduleDir, goRoot, candidate); err != nil {
		return err
	}
	defer func() {
		runErr = errors.Join(runErr, verifyDiscoveryOrdinaryCPP(plan, moduleDir, goRoot, candidate))
	}()
	return run()
}
