package main

import "fmt"

// A subordinate scheduling role, never a test selector or a package admission.
// The complete module owner numbers and selected-native protocol stay intact.
func validateModuleRootScheduling(cfg config) error {
	fail := func() error {
		return fmt.Errorf("root partition scheduling requires an explicit module root part or exhaustive aggregate, with unchanged package ownership")
	}
	if cfg.ModuleRootPartitions < 1 || cfg.ModuleRootPartitions > cfg.Shards {
		return fail()
	}
	if cfg.ModuleRootPartitions == 1 {
		if cfg.ModuleRootPartition != -1 {
			return fail()
		}
		return nil
	}
	if !cfg.Module || cfg.Package != "." || cfg.CISelector != "" || cfg.CIIsolate != "" ||
		cfg.CIVerify != "" || cfg.CIOwner != -1 {
		return fail()
	}
	if cfg.ModuleVerify != "" {
		if cfg.ModuleOwner != -1 || cfg.ModuleRootPartition != -1 {
			return fail()
		}
		return nil
	}
	switch cfg.ModuleOwner {
	case 0:
		if cfg.ModuleRootPartition < 0 || cfg.ModuleRootPartition >= cfg.ModuleRootPartitions {
			return fail()
		}
	case 1, 2:
		if cfg.ModuleRootPartition != -1 {
			return fail()
		}
	default:
		return fail()
	}
	return nil
}

// Reuse the ordinary name-to-shard plan unchanged. Parts own whole global
// shard indices, so child/seed terminals cannot move between CI jobs.
func moduleRootShardParts(shards, parts int) ([][]int, error) {
	if shards < 1 || shards > 1024 || parts < 1 || parts > shards {
		return nil, fmt.Errorf("root partition scheduling requires 1..1024 shards and 1..shards parts")
	}
	result := make([][]int, parts)
	for index := 0; index < shards; index++ {
		part := index % parts
		result[part] = append(result[part], index)
	}
	return result, nil
}
