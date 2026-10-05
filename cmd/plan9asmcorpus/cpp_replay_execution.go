package main

import (
	"encoding/json"
	"fmt"
	"path"
	"reflect"
	"strconv"

	"github.com/xgo-dev/plan9asm"
	"github.com/xgo-dev/plan9asm/internal/gotoolprofile"
)

// Replay full originals with the same executor as the actual CLI. Resource
// registration includes inactive edges; only this execution decides active
// inputs, branches and output. Defines retain their exact ordered -D text.
func replayDiscoveryOriginalCPP(inputs *discoveryCPPInputs, unit discoveryCPPUnit, header *gotoolprofile.GeneratedHeaderProof, defines []string) (plan9asm.AssemblyCPPObservation, map[string]string, error) {
	if inputs == nil || inputs.Protocol != discoveryReplayCPPInputsProtocol {
		return plan9asm.AssemblyCPPObservation{}, nil, fmt.Errorf("missing full original CPP replay protocol")
	}
	read := func(id string) ([]byte, error) {
		source, found := inputs.Sources[id]
		if !found || source.Original == nil || discoveryFeatureBytesSHA256([]byte(*source.Original)) != source.SHA256 {
			return nil, fmt.Errorf("CPP execution lacks its independently captured original: %s", id)
		}
		return []byte(*source.Original), nil
	}
	original := "module/" + unit.File
	source, err := read(original)
	if err != nil {
		return plan9asm.AssemblyCPPObservation{}, nil, err
	}
	used := map[string]string{original: inputs.Sources[original].SHA256}
	observation, err := plan9asm.ObserveAssemblyCPP(string(source), plan9asm.AssemblyPreprocessOptions{
		FileName: original, Defines: defines,
		ReadInclude: func(parent, name string) (string, []byte, error) {
			module, tool, err := gotoolprofile.CPPIncludeCandidates(unit.File, name)
			if err != nil {
				return "", nil, err
			}
			// Search is rooted at the original assembly package CWD, never
			// the including header's directory or an expanded logical #line.
			id := "tool/" + tool
			if gotoolprofile.CPPLocalSource(module) {
				if _, present := inputs.Sources["module/"+module]; present {
					id = "module/" + module
				}
			} else {
				if err := gotoolprofile.RequireCPPPackageSearchMiss(&gotoolprofile.ConsumerInput{
					CPPIncludeSearchMisses: discoveryReplaySearchMisses(unit),
				}, unit.File, parent, name); err != nil {
					return "", nil, err
				}
			}
			bound := false
			// A macro token's runtime parent is its invocation reader, not
			// necessarily its physical definition. The full executor binds
			// that relationship; the raw graph only registers resources.
			for origin, registered := range inputs.Sources {
				for index, directive := range registered.Directives {
					if directive.Kind != "include" || directive.Include != name {
						continue
					}
					key := origin + "#" + strconv.Itoa(index)
					if unit.Includes[key] == id {
						bound = true
					}
					if unit.DeferredIncludes[key] == "generated_go_asm" && path.Clean(name) == "go_asm.h" && header != nil {
						data, err := gotoolprofile.CanonicalGeneratedHeader(header.Definitions)
						if err != nil {
							return "", nil, err
						}
						id = "generated/" + header.PackagePath + "/go_asm.h"
						used[id] = discoveryFeatureBytesSHA256(data)
						return id, data, nil
					}
				}
			}
			if !bound {
				return "", nil, fmt.Errorf("active CPP operand has no complete original registration: %s from %s", name, parent)
			}
			data, err := read(id)
			if err == nil {
				used[id] = inputs.Sources[id].SHA256
			}
			return id, data, err
		},
	})
	if err != nil {
		return plan9asm.AssemblyCPPObservation{}, nil, err
	}
	return observation, used, nil
}

func discoveryReplaySearchMisses(unit discoveryCPPUnit) []gotoolprofile.CPPPackageSearchMiss {
	var misses []gotoolprofile.CPPPackageSearchMiss
	for _, miss := range unit.PackageSearchMisses {
		misses = append(misses, miss)
	}
	gotoolprofile.SortCPPIncludeSearch(misses)
	return misses
}

func validateDiscoveryOriginalCPPConsumption(inputs *discoveryCPPInputs, unit discoveryCPPUnit, header *gotoolprofile.GeneratedHeaderProof, defines []string, proof gotoolprofile.CPPProof) error {
	observation, used, err := replayDiscoveryOriginalCPP(inputs, unit, header, defines)
	if err != nil || !reflect.DeepEqual(used, proof.Inputs) || discoveryFeatureBytesSHA256([]byte(observation.Output)) != proof.ExpandedSHA256 {
		return fmt.Errorf("actual CPP inputs/output differ from full original source replay: %s: %v", unit.File, err)
	}
	return nil
}

// Identity is source/expansion lineage, not a profile-dependent event ordinal.
// Both branch sides of the same invocation receive the same key. Payloads
// distinguish repeated generated controls and definition versions.
func discoveryCPPBranchCoverage(observation plan9asm.AssemblyCPPObservation) (map[string]bool, error) {
	events := make(map[int]plan9asm.AssemblyCPPEvent, len(observation.Events))
	for _, event := range observation.Events {
		events[event.Ordinal] = event
	}
	var identity func(int, int) (string, error)
	identity = func(ordinal, depth int) (string, error) {
		if ordinal == 0 {
			return "", nil
		}
		event, found := events[ordinal]
		if !found || depth > 32 {
			return "", fmt.Errorf("CPP branch has missing/unbounded expansion lineage")
		}
		parent, err := identity(event.Expansion, depth+1)
		if err != nil {
			return "", err
		}
		include, err := identity(event.Include, depth+1)
		if err != nil {
			return "", err
		}
		definition, err := identity(event.Definition, depth+1)
		if err != nil {
			return "", err
		}
		data, err := json.Marshal(struct {
			Position                                           plan9asm.AssemblyCPPObservedPosition
			Kind, Name, Directive, Parent, Include, Definition string
			Tokens                                             []plan9asm.AssemblyCPPObservedToken
		}{event.Position, event.Kind, event.Name, event.Directive, parent, include, definition, event.Tokens})
		return string(data), err
	}
	result := make(map[string]bool)
	for _, event := range observation.Events {
		if event.Kind != "control" || event.Directive != "ifdef" && event.Directive != "ifndef" || !event.EnabledBefore {
			continue
		}
		key, err := identity(event.Ordinal, 0)
		if err != nil {
			return nil, err
		}
		result[key+"/"+strconv.FormatBool(event.EnabledAfter)] = true
	}
	return result, nil
}
