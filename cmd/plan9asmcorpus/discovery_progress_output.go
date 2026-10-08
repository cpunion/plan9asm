package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// Publication is diagnostic output, never a reusable report-validation token.
// A complete audited JSON file may survive a failed completion/ledger gate.
// Link the closed temporary file atomically, without overwriting old evidence.
func writeDiscoveryProgressOutput(filename string, progress discoveryProgress) (result error) {
	if !filepath.IsAbs(filename) || filepath.Base(filename) != "assembly-progress.json" {
		return fmt.Errorf("discovery progress output requires an absolute new assembly-progress.json path")
	}
	parent, err := filepath.EvalSymlinks(filepath.Dir(filename))
	if err != nil {
		return err
	}
	file, err := os.CreateTemp(parent, ".assembly-progress-*")
	if err != nil {
		return err
	}
	owned, err := file.Stat()
	if err != nil {
		return errors.Join(err, file.Close())
	}
	defer func() {
		actual, err := os.Lstat(file.Name())
		if err != nil || !os.SameFile(owned, actual) {
			result = errors.Join(result, fmt.Errorf("owned progress temporary file changed: %v", err))
			return
		}
		result = errors.Join(result, os.Remove(file.Name()))
	}()
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	writeErr := encoder.Encode(progress)
	if err := errors.Join(writeErr, file.Close()); err != nil {
		return err
	}
	return os.Link(file.Name(), filepath.Join(parent, filepath.Base(filename)))
}
