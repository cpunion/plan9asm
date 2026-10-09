package main

import (
	"testing"

	"github.com/xgo-dev/plan9asm/internal/gotoolprofile"
)

// This is an original-source planning model, not an actual compiler/header or
// library pass. Its captured namespace and complete CPP inputs come from the
// existing query-only fixture. The callback follows the production preflight
// boundary so an unclosed import cannot silently supply a generated header.
func TestCPPImportedHeaderPlanningKeepsNecessaryCPUProfiles(t *testing.T) {
	const source = `package fixture

import "errors"

var unusedByHeader = errors.New("not a header property")

type Record struct { Value uint64 }

const Truth = true
const Text = "#ifdef GOAMD64_v3"

func Probe()
`
	fixture := newConditionFreePlanningFixture(t, "linux/arm64", "", map[string]string{
		"pkg/decl.go": source,
	})
	requestsSeen := make(map[string]bool)
	fixture.plan.headerPlanningCapture = func(request discoveryFeatureProfileRequest,
		assembly string, tags []string,
	) (*gotoolprofile.GeneratedHeaderProof, error) {
		requestsSeen[discoveryHeaderPlanningKey(request, assembly, tags)] = true
		sources, err := fixture.selectedGoSources(request)
		if err != nil {
			return nil, err
		}
		if err := gotoolprofile.PreflightNativeGeneratedHeaderSource(
			fixture.plan.Module+"/pkg", request.Target, "go1.20", sources,
		); err != nil {
			if gotoolprofile.NativeHeaderSourceDomainUnsupported(err) {
				return nil, &discoveryHeaderPlanningUnsupported{err.Error()}
			}
			return nil, err
		}
		return fixture.header(t, request, assembly)
	}
	initial, found, err := planDiscoveryFeaturePair(fixture.context, fixture.baseline,
		fixture.assembly, "pkg/decl.go")
	if err != nil || !found {
		t.Fatalf("original source pair: found=%t error=%v", found, err)
	}
	requests, err := planDiscoveryCPPFeaturePair(fixture.plan, fixture.context,
		fixture.baseline, fixture.assembly, "pkg/decl.go", initial)
	if err != nil {
		t.Fatal(err)
	}
	if len(requests) != 0 || len(requestsSeen) != 1 {
		t.Fatalf("condition-free imported source expanded into %d extra CPU proposals and %d header requests; want one necessary baseline and no extra CPU proposals",
			len(requests), len(requestsSeen))
	}
}
