package main

import (
	"fmt"
	"go/build"
	"strings"

	"github.com/xgo-dev/plan9asm"
)

// CPP has a separate feature namespace: for example GOAMD64_v3 is exclusive,
// whereas the amd64.v3 build tag is also true at v4. Enumerate the finite CPU
// environments and replay local macro definedness before proposing profiles.
// Only reachable branch sides require profiles; inactive outer conditions do
// not create a claim of branch execution or full instruction-form coverage.
func planDiscoveryCPPFeaturePair(plan *discoveryOrdinarySelectionPlan, ctx build.Context, baseline *discoveryTargetFeatures, asmFile, goFile string, initial discoveryFeatureProfileRequest) ([]discoveryFeatureProfileRequest, error) {
	if plan.CPPInputs == nil {
		return nil, nil
	}
	var unit *discoveryCPPUnit
	for index := range plan.CPPInputs.Units {
		if plan.CPPInputs.Units[index].File == asmFile {
			unit = &plan.CPPInputs.Units[index]
			break
		}
	}
	if unit == nil {
		return nil, fmt.Errorf("source-required ASM pair lacks its exact CPP unit: %s", asmFile)
	}
	directives, err := discoveryCPPUnitDirectives(plan.CPPInputs, *unit, make(map[string]bool))
	if err != nil {
		return nil, err
	}
	if !discoveryCPPHasFeatureConditions(directives) {
		return nil, nil
	}
	for _, directive := range directives {
		if (directive.Kind == "ifdef" || directive.Kind == "ifndef") && strings.HasPrefix(directive.Name, "GOEXPERIMENT_") {
			return nil, fmt.Errorf("CPP experiment macro requires actual package-role binding, not an ordinary/custom feature assumption")
		}
	}
	minor, err := discoveryGoMinor(baseline.GoVersion)
	if err != nil {
		return nil, err
	}
	cpuKey := discoveryCPPFeatureCPUKey(ctx.GOARCH)
	values := discoveryCPPFeatureCPUValues(ctx.GOARCH, minor, baseline.Environment[cpuKey])
	if err := validateDiscoveryCPPFeatureNamespace(directives, ctx.GOARCH); err != nil {
		return nil, err
	}
	covered := make(map[string]bool)
	cover := func(request discoveryFeatureProfileRequest) (bool, error) {
		env := cloneDiscoveryCPPEnvironment(baseline.Environment)
		for key, value := range request.Overrides {
			env[key] = value
		}
		defines, err := plan9asm.GoAssemblerDefinesForEnvironment(ctx.GOOS, ctx.GOARCH, env)
		if err != nil {
			return false, err
		}
		states, err := replayDiscoveryCPPConditions(directives, defines)
		if err != nil {
			return false, fmt.Errorf("bounded CPP replay requires concrete source diagnosis: %w", err)
		}
		added := false
		for index, state := range states {
			if !state.OuterActive {
				continue
			}
			key := fmt.Sprintf("%d/%t", index, state.Then)
			if !covered[key] {
				covered[key], added = true, true
			}
		}
		return added, nil
	}
	if _, err := cover(initial); err != nil {
		return nil, err
	}
	var requests []discoveryFeatureProfileRequest
	for _, value := range values {
		if err := validateDiscoveryCPUEnvironment(ctx.GOARCH, value, minor); err != nil {
			continue
		}
		// This is only a planning model. captureDiscoveryFeatureProfiles will
		// observe the proposal with the actual Go driver before accepting it.
		model := *baseline
		model.Environment = cloneDiscoveryCPPEnvironment(baseline.Environment)
		model.Environment[cpuKey] = value
		model.ToolTags = discoveryModeledCPUFeatures(baseline.ToolTags, ctx.GOARCH, value, minor)
		modeledContext := ctx
		modeledContext.ToolTags = model.ToolTags
		request, found, err := planDiscoveryFeaturePair(modeledContext, &model, asmFile, goFile)
		if err != nil {
			return nil, err
		}
		if !found {
			continue
		}
		requestedCPU := value
		if override, present := request.Overrides[cpuKey]; present {
			requestedCPU = override
		}
		if requestedCPU == baseline.Environment[cpuKey] {
			delete(request.Overrides, cpuKey)
		} else {
			request.Overrides[cpuKey] = requestedCPU
		}
		request.Baseline = len(request.Overrides) == 0
		request.pairs = [][2]string{{asmFile, goFile}}
		added, err := cover(request)
		if err != nil {
			return nil, err
		}
		if added {
			requests = append(requests, request)
		}
	}
	return requests, nil
}

func discoveryCPPFeatureCPUKey(arch string) string {
	return map[string]string{"386": "GO386", "amd64": "GOAMD64", "arm": "GOARM", "arm64": "GOARM64", "wasm": "GOWASM"}[arch]
}

func discoveryCPPFeatureCPUValues(arch string, minor int, baseline string) []string {
	values := []string{baseline}
	switch arch {
	case "386":
		values = append(values, "sse2", "softfloat")
	case "amd64":
		values = append(values, "v1", "v2", "v3", "v4")
	case "arm":
		values = append(values, "5", "6", "7")
	case "arm64":
		if minor >= 23 {
			// Keep high-level Go constraints. Actual Go registers ,lse but
			// no ,nolse; every level above v8.0 already implies LSE. A high-
			// level Go declaration has no reachable non-LSE CPP variant.
			for _, tag := range discoveryCPUFeatureCandidates() {
				if strings.HasPrefix(tag, "arm64.") {
					level := strings.TrimPrefix(tag, "arm64.")
					values = append(values, level, level+",lse")
				}
			}
		}
	case "wasm":
		values = append(values, "", "satconv", "signext", "satconv,signext")
	}
	return uniqueDiscoveryStrings(values)
}

func cloneDiscoveryCPPEnvironment(environment map[string]string) map[string]string {
	copy := make(map[string]string, len(environment))
	for key, value := range environment {
		copy[key] = value
	}
	return copy
}

func discoveryCPPHasFeatureConditions(directives []discoveryCPPDirective) bool {
	for _, directive := range directives {
		if directive.Kind != "ifdef" && directive.Kind != "ifndef" {
			continue
		}
		for _, prefix := range []string{"GO386_", "GOAMD64_", "GOARM_", "GOARM64_", "GOWASM_", "GOEXPERIMENT_"} {
			if strings.HasPrefix(directive.Name, prefix) {
				return true
			}
		}
	}
	return false
}

func validateDiscoveryCPPFeatureNamespace(directives []discoveryCPPDirective, arch string) error {
	key := discoveryCPPFeatureCPUKey(arch)
	known := map[string][]string{
		"386":   {"GO386_sse2", "GO386_softfloat"},
		"amd64": {"GOAMD64_v1", "GOAMD64_v2", "GOAMD64_v3", "GOAMD64_v4"},
		"arm":   {"GOARM_5", "GOARM_6", "GOARM_7"},
		"arm64": {"GOARM64_LSE"},
		// Go's audited asmArgs does not register these WASM macros. Keep
		// their known source spelling fixed-false, not fake -D or -tags.
		"wasm": {"GOWASM_satconv", "GOWASM_signext"},
	}
	for _, directive := range directives {
		if (directive.Kind == "ifdef" || directive.Kind == "ifndef") && strings.HasPrefix(directive.Name, key+"_") &&
			!containsTargetFeature(known[arch], directive.Name) {
			return fmt.Errorf("unknown target CPU CPP namespace %s (not baseline N/A)", directive.Name)
		}
	}
	return nil
}
