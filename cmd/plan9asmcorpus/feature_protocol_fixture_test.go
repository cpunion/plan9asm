package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"

	"github.com/xgo-dev/plan9asm/internal/gotoolprofile"
)

// These Go 1.27.1 baseline/registration bytes were captured with the real
// driver. Synthetic objects below test portable accounting only; they are not
// library, LLVM or runtime passes. Keeping this fixture portable also lets
// Go 1.20 test the schema reader without pretending to run the Go 1.27 gate.
type featureProtocolFixture struct {
	Observed map[string]*discoveryTargetFeatures       `json:"observed"`
	Macros   map[string]*gotoolprofile.AssemblerMacros `json:"macros"`
	CPP      *discoveryCPPRegistration                 `json:"cpp"`
}

var ordinaryFixtureRoots sync.Map

func fixtureProfileEvidence(t *testing.T, result *discoveryCorpusResult, inventory *discoveryFeatureInventory) {
	t.Helper()
	data, err := os.ReadFile("testdata/go127_feature_protocol_fixture.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture featureProtocolFixture
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	plan := result.OrdinarySelectionPlan
	var profiles []discoveryFeatureProfile
	byTarget := make(map[string]discoveryFeatureProfile)
	for _, target := range plan.Targets {
		observed := fixture.Observed[target]
		if observed == nil {
			t.Fatalf("no actually captured protocol baseline fixture for %s", target)
		}
		profile := discoveryFeatureProfile{ID: discoveryFeatureProfileID(observed), Observed: observed,
			Request: discoveryFeatureProfileRequest{Target: target, Baseline: true}}
		profiles = append(profiles, profile)
		byTarget[target] = profile
	}
	sort.Slice(profiles, func(i, j int) bool {
		return discoveryFeatureRequestKey(profiles[i].Request) < discoveryFeatureRequestKey(profiles[j].Request)
	})
	result.FeatureProfiles, err = registerDiscoveryFeatureProfiles(inventory, profiles)
	if err != nil {
		t.Fatal(err)
	}
	result.featureInventory = inventory
	decisions, eligible, err := replayOrdinarySelectionForProfiles(plan, result.DiscoveredAsmFiles, profiles)
	if err != nil {
		t.Fatal(err)
	}
	plan.ProfileDecisions = decisions
	files := ordinaryProfileEligibleCPPFiles(eligible)
	if len(files) != 0 {
		root, found := ordinaryFixtureRoots.Load(plan)
		if !found {
			t.Fatal("synthetic protocol fixture lost its original source directory")
		}
		inputs := &discoveryCPPInputs{Protocol: discoveryCPPInputsProtocol,
			Module: plan.Module, Version: plan.Version, ModuleSum: plan.ModuleSum, ZipSHA256: plan.ZipSHA256,
			Registration: fixture.CPP, Sources: make(map[string]discoveryCPPSource)}
		for _, file := range files {
			contents, err := os.ReadFile(filepath.Join(root.(string), filepath.FromSlash(file)))
			if err != nil {
				t.Fatal(err)
			}
			source, err := discoveryCPPConditionsFromBytes(file, contents)
			if err != nil {
				t.Fatal(err)
			}
			inputs.Sources["module/"+file] = source
			inputs.Units = append(inputs.Units, discoveryCPPUnit{File: file})
		}
		plan.CPPInputs = inputs
	}
	var configs []discoveryBuildConfiguration
	for _, config := range result.BuildConfigurations {
		for _, target := range config.Targets {
			configs = append(configs, discoveryBuildConfiguration{
				ProfileID: byTarget[target].ID, BuildTags: config.BuildTags,
				Targets: []string{target}, AsmFiles: config.AsmFiles,
			})
		}
	}
	result.BuildConfigurations = configs
	var sourceItems []discoverySourceNotApplicableItem
	for _, item := range result.SourceNotApplicableItems {
		for _, target := range item.Targets {
			copy := item
			copy.ProfileID, copy.Targets = byTarget[target].ID, []string{target}
			sourceItems = append(sourceItems, copy)
		}
	}
	result.SourceNotApplicableItems = sourceItems
	result.FeatureConsumption = nil
	for _, config := range configs {
		profile := byTarget[config.Targets[0]]
		module, err := ordinaryProfileDeclaredModule(plan)
		if err != nil {
			t.Fatal(err)
		}
		input := ordinaryProfileConsumerInput(plan, profile, module, "", config.AsmFiles)
		proof := &gotoolprofile.SelectionProof{Protocol: gotoolprofile.ConsumerProtocol, ProfileID: profile.ID, CustomTags: config.BuildTags}
		ctx, err := discoveryContextForFeatureProfile(profile.Observed)
		if err != nil {
			t.Fatal(err)
		}
		ctx.BuildTags = config.BuildTags
		byDirectory := make(map[string][]string)
		for _, file := range config.AsmFiles {
			byDirectory[path.Dir(file)] = append(byDirectory[path.Dir(file)], file)
			proof.CPP = append(proof.CPP, gotoolprofile.CPPProof{File: file,
				ExpandedSHA256: strings.Repeat("1", 64), TypedExpandedSHA256: strings.Repeat("2", 64),
				Inputs: map[string]string{"module/" + file: input.Sources[file]}})
			proof.Outputs = append(proof.Outputs, gotoolprofile.OutputProof{
				File: file, Part: path.Base(file) + ".ll", IR: strings.Repeat("3", 64), Object: strings.Repeat("4", 64),
			})
		}
		for _, dir := range sortedFixtureDirectories(byDirectory) {
			pkgPath := module
			if dir != "." {
				pkgPath += "/" + dir
			}
			macros := *fixture.Macros[profile.Observed.Target]
			macros.PackagePath = pkgPath
			pkg := gotoolprofile.PackageProof{
				PackagePath: pkgPath, ModulePath: module, ModuleVersion: plan.Version,
				SourceModule: plan.Module, SourceVersion: plan.Version, SourceRole: "module",
				SFiles: byDirectory[dir], SourceSHA256: make(map[string]string), Macros: &macros,
			}
			for _, source := range plan.Sources {
				if path.Dir(source.File) != dir || !strings.HasSuffix(source.File, ".go") || strings.HasSuffix(source.File, "_test.go") {
					continue
				}
				ctx.OpenFile = fixtureHeaderOpener(plan)
				selected, err := ctx.MatchFile(dir, path.Base(source.File))
				if err != nil {
					t.Fatal(err)
				}
				if selected {
					pkg.GoFiles = append(pkg.GoFiles, source.File)
					pkg.SourceSHA256[source.File] = input.Sources[source.File]
				}
			}
			pkg.CompiledGoFiles = append([]string(nil), pkg.GoFiles...)
			for _, file := range pkg.SFiles {
				pkg.SourceSHA256[file] = input.Sources[file]
			}
			proof.Packages = append(proof.Packages, pkg)
		}
		result.FeatureConsumption = append(result.FeatureConsumption, proof)
	}
}

func sortedFixtureDirectories(dirs map[string][]string) []string {
	set := make(map[string]bool)
	for dir := range dirs {
		set[dir] = true
	}
	return sortedDiscoverySet(set)
}

func fixtureHeaderOpener(plan *discoveryOrdinarySelectionPlan) func(string) (io.ReadCloser, error) {
	return func(file string) (io.ReadCloser, error) {
		for _, source := range plan.Sources {
			if source.File == filepath.ToSlash(file) {
				return io.NopCloser(strings.NewReader(source.Header)), nil
			}
		}
		return nil, fmt.Errorf("uncaptured fixture source %s", file)
	}
}

func fixtureProfileMatrix(t *testing.T, result discoveryCorpusResult) matrixReport {
	t.Helper()
	profiles, err := resolveDiscoveryFeatureProfiles(result.featureInventory, result.FeatureProfiles)
	if err != nil {
		t.Fatal(err)
	}
	return matrixReport{Success: result.Translations, TotalTargets: len(result.OrdinarySelectionPlan.Targets),
		OrdinarySelectionPlan: result.OrdinarySelectionPlan,
		FeatureProfiles:       profiles, FeatureConsumption: result.FeatureConsumption}
}
