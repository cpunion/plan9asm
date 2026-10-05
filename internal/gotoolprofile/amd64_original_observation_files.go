package gotoolprofile

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
)

// Files and directory identities are private, in-memory observations. They
// are neither a stored trusted index nor signed archive authentication.
type amd64OriginalObservationFiles struct {
	hashes, paths map[string]string
	entries       map[string][]string
	directories   map[string]os.FileInfo
	cpp           map[string]amd64OriginalCPPObservation
	bytes         int64
}

type amd64OriginalCPPObservation struct {
	Expanded string
	Inputs   []string
}

func newAMD64OriginalObservationFiles() *amd64OriginalObservationFiles {
	return &amd64OriginalObservationFiles{hashes: make(map[string]string), paths: make(map[string]string),
		entries: make(map[string][]string), directories: make(map[string]os.FileInfo), cpp: make(map[string]amd64OriginalCPPObservation)}
}

func (files *amd64OriginalObservationFiles) retain(name string) error {
	if !filepath.IsAbs(name) || len(files.hashes) >= 65536 {
		return fmt.Errorf("original file path is missing or over-bound")
	}
	info, err := os.Lstat(name)
	if err != nil || !info.Mode().IsRegular() || info.Size() > 16<<20 {
		return fmt.Errorf("original source/export is missing, nonregular or over-bound: %s", name)
	}
	physical, err := filepath.EvalSymlinks(name)
	if err != nil {
		return err
	}
	digest, err := FileSHA256(name)
	if err != nil {
		return err
	}
	if prior, present := files.hashes[name]; present {
		if prior != digest || files.paths[name] != physical {
			return fmt.Errorf("original source/export changed during capture")
		}
	} else {
		files.bytes += info.Size()
		if files.bytes > 64<<20 {
			return fmt.Errorf("original source/export inventory exceeds its byte bound")
		}
	}
	files.hashes[name], files.paths[name] = digest, physical
	return nil
}

func (files *amd64OriginalObservationFiles) directory(name string, recursive bool) error {
	physical, err := filepath.EvalSymlinks(name)
	if err != nil {
		return err
	}
	info, err := os.Lstat(physical)
	if err != nil || !info.IsDir() || len(files.directories) >= 4096 {
		return fmt.Errorf("original source directory is missing, nonregular or over-bound")
	}
	if prior, present := files.directories[name]; present && (!os.SameFile(prior, info) || files.paths[name] != physical) {
		return fmt.Errorf("original source directory identity changed")
	}
	entries, err := os.ReadDir(physical)
	if err != nil {
		return err
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		path := filepath.Join(physical, entry.Name())
		names = append(names, entry.Name()+":"+entry.Type().String())
		if entry.Type()&os.ModeSymlink != 0 || !entry.IsDir() && !entry.Type().IsRegular() {
			return fmt.Errorf("original source directory contains an unbound nonregular entry")
		}
		if recursive {
			if entry.IsDir() {
				if err := files.directory(path, true); err != nil {
					return err
				}
			} else if err := files.retain(path); err != nil {
				return err
			}
		}
	}
	if prior, present := files.entries[name]; present && !reflect.DeepEqual(prior, names) {
		return fmt.Errorf("original source directory membership changed during capture")
	}
	files.entries[name], files.paths[name], files.directories[name] = names, physical, info
	return nil
}

func (files *amd64OriginalObservationFiles) verify() error {
	if files == nil {
		return fmt.Errorf("missing original source/export inventory")
	}
	for name, digest := range files.hashes {
		info, err := os.Lstat(name)
		physical, resolveErr := filepath.EvalSymlinks(name)
		if err != nil || resolveErr != nil || !info.Mode().IsRegular() || physical != files.paths[name] {
			return fmt.Errorf("original source/export physical identity changed")
		}
		if actual, err := FileSHA256(name); err != nil || actual != digest {
			return fmt.Errorf("original source/export bytes changed")
		}
	}
	for name := range files.directories {
		physical, err := filepath.EvalSymlinks(name)
		info, statErr := os.Lstat(physical)
		entries, readErr := os.ReadDir(physical)
		if err != nil || statErr != nil || readErr != nil || physical != files.paths[name] || !os.SameFile(files.directories[name], info) {
			return fmt.Errorf("original source directory identity changed")
		}
		names := make([]string, 0, len(entries))
		for _, entry := range entries {
			names = append(names, entry.Name()+":"+entry.Type().String())
		}
		if !reflect.DeepEqual(names, files.entries[name]) {
			return fmt.Errorf("original source directory membership changed")
		}
	}
	return nil
}
