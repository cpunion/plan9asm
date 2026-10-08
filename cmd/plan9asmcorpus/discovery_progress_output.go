package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type preparedDiscoveryProgress struct {
	destination string
	temporary   string
	owned       os.FileInfo
	parent      os.FileInfo
}

// Prepare and close the complete diagnostic before the final input checks.
// Preparation is not publication and does not create the destination file.
func prepareDiscoveryProgressOutput(filename string, progress discoveryProgress) (*preparedDiscoveryProgress, error) {
	if !filepath.IsAbs(filename) || filepath.Base(filename) != "assembly-progress.json" {
		return nil, fmt.Errorf("discovery progress output requires an absolute new assembly-progress.json path")
	}
	parent, err := filepath.EvalSymlinks(filepath.Dir(filename))
	if err != nil {
		return nil, err
	}
	parentInfo, err := os.Stat(parent)
	if err != nil {
		return nil, err
	}
	file, err := os.CreateTemp(parent, ".assembly-progress-*")
	if err != nil {
		return nil, err
	}
	owned, err := file.Stat()
	if err != nil {
		return nil, errors.Join(err, file.Close())
	}
	prepared := &preparedDiscoveryProgress{
		destination: filepath.Join(parent, filepath.Base(filename)), temporary: file.Name(),
		owned: owned, parent: parentInfo,
	}
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	writeErr := encoder.Encode(progress)
	if err := errors.Join(writeErr, file.Close()); err != nil {
		return nil, errors.Join(err, prepared.discard())
	}
	return prepared, nil
}

// Link only after all frozen input checks succeed. Existing output is never
// replaced, including a file that appeared after the initial preflight.
func (prepared *preparedDiscoveryProgress) publish() error {
	parent, err := os.Stat(filepath.Dir(prepared.destination))
	if err != nil || !os.SameFile(prepared.parent, parent) {
		return errors.Join(fmt.Errorf("progress destination parent changed"), err)
	}
	actual, err := os.Lstat(prepared.temporary)
	if err != nil || !os.SameFile(prepared.owned, actual) {
		return errors.Join(fmt.Errorf("owned progress temporary file changed"), err)
	}
	return os.Link(prepared.temporary, prepared.destination)
}

// Never remove a replacement supplied by another operation. This method runs
// on every ordinary error path before main's strict exit, not in an os.Exit defer.
func (prepared *preparedDiscoveryProgress) discard() error {
	actual, err := os.Lstat(prepared.temporary)
	if err != nil || !os.SameFile(prepared.owned, actual) {
		return errors.Join(fmt.Errorf("owned progress temporary file changed"), err)
	}
	return os.Remove(prepared.temporary)
}
