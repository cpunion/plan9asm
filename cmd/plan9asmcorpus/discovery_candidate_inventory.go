package main

import (
	"sort"

	"github.com/xgo-dev/plan9asm/internal/discoverymeta"
)

type discoveryCandidateInventory struct {
	// Identities have no assembly payload. Exact-request and alias validation
	// need the complete ledger membership, not every shard's source-file lists.
	Identities []discoveryCandidate
	Eligible   int
	FullShard  []discoveryCandidate
}

// loadDiscoveryCandidateInventory validates every record while retaining full
// assembly payloads only for the executing shard. Global membership and target
// eligibility remain exact, including duplicate matched records across files.
func loadDiscoveryCandidateInventory(root string, cfg discoveryCorpusConfig) (discoveryCandidateInventory, error) {
	type identity struct {
		candidate discoveryCandidate
		eligible  bool
	}
	type selectedPayload struct {
		candidate discoveryCandidate
		arches    map[string]bool
		files     map[string]bool
	}
	identities := make(map[string]*identity)
	selected := make(map[string]*selectedPayload)
	err := walkDiscoveryRecords(root, func(record discoveryRecord) error {
		if record.Kind != "matched" {
			return nil
		}
		key := record.Module + "@" + record.Version
		item := identities[key]
		if item == nil {
			item = &identity{candidate: discoveryCandidate{Module: record.Module, Version: record.Version}}
			identities[key] = item
		}
		files := record.AsmFiles
		if cfg.FilterTargets {
			var err error
			files, err = discoverymeta.FilterAssemblyFiles(files, cfg.Targets)
			if err != nil {
				return err
			}
		}
		if len(files) == 0 {
			return nil
		}
		item.eligible = true
		if discoveryCandidateShard(item.candidate, cfg.ShardCount) != cfg.ShardIndex {
			return nil
		}
		payload := selected[key]
		if payload == nil {
			payload = &selectedPayload{candidate: item.candidate,
				arches: make(map[string]bool), files: make(map[string]bool)}
			selected[key] = payload
		}
		for _, file := range files {
			payload.files[file] = true
		}
		if !cfg.FilterTargets {
			for _, arch := range record.Architectures {
				payload.arches[arch] = true
			}
		}
		return nil
	})
	if err != nil {
		return discoveryCandidateInventory{}, err
	}
	result := discoveryCandidateInventory{
		Identities: make([]discoveryCandidate, 0, len(identities)),
		FullShard:  make([]discoveryCandidate, 0, len(selected)),
	}
	for _, item := range identities {
		result.Identities = append(result.Identities, item.candidate)
		if item.eligible {
			result.Eligible++
		}
	}
	for _, payload := range selected {
		for file := range payload.files {
			payload.candidate.AsmFiles = append(payload.candidate.AsmFiles, file)
		}
		sort.Strings(payload.candidate.AsmFiles)
		if cfg.FilterTargets {
			payload.candidate.Architectures = discoverymeta.ArchitectureHints(payload.candidate.AsmFiles)
		} else {
			for arch := range payload.arches {
				payload.candidate.Architectures = append(payload.candidate.Architectures, arch)
			}
			sort.Strings(payload.candidate.Architectures)
		}
		result.FullShard = append(result.FullShard, payload.candidate)
	}
	for _, candidates := range [][]discoveryCandidate{result.Identities, result.FullShard} {
		sort.Slice(candidates, func(i, j int) bool {
			if candidates[i].Module != candidates[j].Module {
				return candidates[i].Module < candidates[j].Module
			}
			return candidates[i].Version < candidates[j].Version
		})
	}
	return result, nil
}
