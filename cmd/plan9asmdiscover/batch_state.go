package main

type scanBatchState struct {
	seen        map[string]seenResult
	sourceSeen  map[string]seenResult
	checkpoints map[string]moduleCheckpoint
}

// The initial adapter deliberately exercises the existing eager helpers.
// The regression tests require this API to retain only the active batch.
func loadBatchState(paths []string, modules []moduleVersion) (scanBatchState, error) {
	seen, sourceSeen, err := loadSeenState(paths)
	if err != nil {
		return scanBatchState{}, err
	}
	checkpoints, err := loadModuleCheckpoints(paths)
	if err != nil {
		return scanBatchState{}, err
	}
	return scanBatchState{seen: seen, sourceSeen: sourceSeen, checkpoints: checkpoints}, nil
}
