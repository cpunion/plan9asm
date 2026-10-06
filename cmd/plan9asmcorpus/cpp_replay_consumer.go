package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/xgo-dev/plan9asm"
	"github.com/xgo-dev/plan9asm/internal/gotoolprofile"
)

// The plan owns the original graph once. Private consumer inputs receive the
// needed immutable bytes; this transport is never portable source authority.
func discoveryCPPOriginalConsumerInput(inputs *discoveryCPPInputs) (*gotoolprofile.ConsumerInput, error) {
	if inputs == nil || inputs.Protocol != discoveryReplayCPPInputsProtocol {
		return nil, fmt.Errorf("missing full original CPP graph")
	}
	input := &gotoolprofile.ConsumerInput{
		CPPSourceProtocol: gotoolprofile.CPPOriginalReplayProtocol,
		Sources:           make(map[string]string), ToolSources: make(map[string]string),
		CPPOriginals: &gotoolprofile.CPPOriginalReplayInput{
			Sources: make(map[string]string), Units: make(map[string]gotoolprofile.CPPOriginalUnit),
			RequiredExpandedOperands: make(map[string][]gotoolprofile.CPPPackageSearchMiss),
		},
	}
	for id, source := range inputs.Sources {
		if source.Original == nil {
			return nil, fmt.Errorf("CPP consumer omitted original source bytes")
		}
		origin, file, err := discoveryCPPSourceIdentity(id)
		if err != nil {
			return nil, err
		}
		input.CPPOriginals.Sources[id] = *source.Original
		if origin == "tool" {
			input.ToolSources[file] = source.SHA256
		} else {
			input.Sources[file] = source.SHA256
		}
	}
	for _, unit := range inputs.Units {
		input.AsmFiles = append(input.AsmFiles, unit.File)
		registered := gotoolprofile.CPPOriginalUnit{Includes: make(map[string]string)}
		for id, source := range inputs.Sources {
			for index, directive := range source.Directives {
				if directive.Kind != "include" {
					continue
				}
				key := id + "#" + strconv.Itoa(index)
				if target := unit.Includes[key]; target != "" {
					if previous := registered.Includes[directive.Include]; previous != "" && previous != target {
						return nil, fmt.Errorf("CPP fixed include origins conflict between raw controls")
					}
					registered.Includes[directive.Include] = target
				}
				if unit.DeferredIncludes[key] == "generated_go_asm" {
					if len(registered.GeneratedHeaders) == 0 {
						registered.GeneratedHeaders = []string{"go_asm.h"}
					}
				}
			}
		}
		input.CPPOriginals.Units[unit.File] = registered
	}
	input.CPPIncludeSearchMisses = discoveryCPPPhysicalSearchMisses(inputs)
	return input, gotoolprofile.ValidateCPPOriginalInputs(input)
}

func bindDiscoveryCPPOriginalConsumer(input *gotoolprofile.ConsumerInput, inputs *discoveryCPPInputs) {
	graphInput, err := discoveryCPPOriginalConsumerInput(inputs)
	input.CPPSourceProtocol = gotoolprofile.CPPOriginalReplayProtocol
	if err != nil {
		return // The mandatory protocol without a graph is fail-closed.
	}
	input.CPPOriginals = graphInput.CPPOriginals
	defines, err := plan9asm.GoAssemblerDefinesForEnvironment(input.Observed.Environment["GOOS"], input.Observed.Environment["GOARCH"], input.Observed.Environment)
	if err != nil {
		return
	}
	for _, assembly := range input.AsmFiles {
		packagePath := gotoolprofile.CPPOriginalPackagePath(input.Module, assembly)
		var header []byte
		if input.GeneratedHeaders != nil {
			for _, actual := range input.GeneratedHeaders.Headers {
				if actual.PackagePath == packagePath {
					header, err = gotoolprofile.CanonicalGeneratedHeader(actual.Definitions)
					if err != nil {
						return
					}
				}
			}
		}
		_, _, required, err := gotoolprofile.ReplayCPPOriginal(input, assembly, packagePath, defines, header)
		if err != nil {
			continue // Query inputs cannot manufacture a missing native header.
		}
		input.CPPOriginals.RequiredExpandedOperands[assembly] = required
		input.CPPIncludeSearchMisses = append(input.CPPIncludeSearchMisses, required...)
	}
	gotoolprofile.SortCPPIncludeSearch(input.CPPIncludeSearchMisses)
}

func discoveryCPPPhysicalSearchMisses(inputs *discoveryCPPInputs) []gotoolprofile.CPPPackageSearchMiss {
	var misses []gotoolprofile.CPPPackageSearchMiss
	seen := make(map[string]bool)
	for _, unit := range inputs.Units {
		for _, miss := range unit.PackageSearchMisses {
			if miss.Protocol != gotoolprofile.CPPPackageSearchMissProtocol {
				continue // Raw generated resources are not active receipts.
			}
			key := strings.Join([]string{miss.Assembly, miss.Source, miss.Include}, "\x00")
			if !seen[key] {
				seen[key] = true
				misses = append(misses, miss)
			}
		}
	}
	gotoolprofile.SortCPPIncludeSearch(misses)
	return misses
}
