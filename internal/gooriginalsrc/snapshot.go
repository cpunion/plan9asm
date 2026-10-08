// Package gooriginalsrc observes bounded original files and exact Go-download
// origins. Its state is byte/role consistency data, never an ASM capability.
package gooriginalsrc

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"sync"
)

var originalHashBuffers = sync.Pool{New: func() any { return new([64 << 10]byte) }}

const (
	fileLimit      = 65536
	directoryLimit = 4096
	fileByteLimit  = 16 << 20
	totalByteLimit = 64 << 20
)

// Snapshot retains complete physical files and directory membership. It has no
// serialized constructor, command callback, frame, signature or permission.
type Snapshot struct {
	hashes, paths map[string]string
	entries       map[string][]string
	directories   map[string]os.FileInfo
	files         map[string]os.FileInfo
	bytes         int64
}

// SnapshotView is the existing producer's observation data. These maps expose
// no capability: consumers must retain their own actual Snapshot and Verify it.
// The compatibility producer uses this view rather than duplicating inventories.
type SnapshotView struct {
	Hashes, Paths map[string]string
	Entries       map[string][]string
	Directories   map[string]os.FileInfo
}

func NewSnapshot() *Snapshot {
	return &Snapshot{
		hashes: make(map[string]string), paths: make(map[string]string),
		entries: make(map[string][]string), directories: make(map[string]os.FileInfo),
		files: make(map[string]os.FileInfo),
	}
}

func (snapshot *Snapshot) View() SnapshotView {
	return SnapshotView{snapshot.hashes, snapshot.paths, snapshot.entries, snapshot.directories}
}

func (snapshot *Snapshot) Retain(name string) error {
	if !filepath.IsAbs(name) || len(snapshot.hashes) >= fileLimit {
		return fmt.Errorf("original file path is missing or over-bound")
	}
	info, err := os.Lstat(name)
	if err != nil || !info.Mode().IsRegular() || info.Size() > fileByteLimit {
		return fmt.Errorf("original source/export is missing, nonregular or over-bound: %s", filepath.Base(name))
	}
	physical, err := filepath.EvalSymlinks(name)
	if err != nil {
		return err
	}
	digest, err := originalFileSHA256(name, info)
	if err != nil {
		return err
	}
	if prior, present := snapshot.hashes[name]; present {
		if prior != digest || snapshot.paths[name] != physical || !os.SameFile(snapshot.files[name], info) {
			return fmt.Errorf("original source/export changed during capture")
		}
	} else {
		snapshot.bytes += info.Size()
		if snapshot.bytes > totalByteLimit {
			return fmt.Errorf("original source/export inventory exceeds its byte bound")
		}
	}
	snapshot.hashes[name], snapshot.paths[name], snapshot.files[name] = digest, physical, info
	return nil
}

func (snapshot *Snapshot) Directory(name string, recursive bool) error {
	physical, err := filepath.EvalSymlinks(name)
	if err != nil {
		return err
	}
	info, err := os.Lstat(physical)
	if err != nil || !info.IsDir() || len(snapshot.directories) >= directoryLimit {
		return fmt.Errorf("original source directory is missing, nonregular or over-bound")
	}
	if prior, present := snapshot.directories[name]; present &&
		(!os.SameFile(prior, info) || snapshot.paths[name] != physical) {
		return fmt.Errorf("original source directory identity changed")
	}
	entries, err := os.ReadDir(physical)
	if err != nil {
		return err
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		child := filepath.Join(physical, entry.Name())
		names = append(names, entry.Name()+":"+entry.Type().String())
		if entry.Type()&os.ModeSymlink != 0 || !entry.IsDir() && !entry.Type().IsRegular() {
			return fmt.Errorf("original source directory contains an unbound nonregular entry")
		}
		if recursive {
			if entry.IsDir() {
				if err := snapshot.Directory(child, true); err != nil {
					return err
				}
			} else if err := snapshot.Retain(child); err != nil {
				return err
			}
		}
	}
	if prior, present := snapshot.entries[name]; present && !reflect.DeepEqual(prior, names) {
		return fmt.Errorf("original source directory membership changed during capture")
	}
	snapshot.entries[name], snapshot.paths[name], snapshot.directories[name] = names, physical, info
	return nil
}

func (snapshot *Snapshot) Verify() error {
	if snapshot == nil {
		return fmt.Errorf("missing original source/export inventory")
	}
	for name, digest := range snapshot.hashes {
		info, err := os.Lstat(name)
		physical, resolveErr := filepath.EvalSymlinks(name)
		if err != nil || resolveErr != nil || !info.Mode().IsRegular() ||
			physical != snapshot.paths[name] || !os.SameFile(snapshot.files[name], info) {
			return fmt.Errorf("original source/export physical identity changed")
		}
		actual, err := originalFileSHA256(name, info)
		if err != nil {
			return fmt.Errorf("original source/export bytes changed: %w", err)
		}
		if actual != digest {
			return fmt.Errorf("original source/export bytes changed")
		}
	}
	for name := range snapshot.directories {
		physical, err := filepath.EvalSymlinks(name)
		info, statErr := os.Lstat(physical)
		entries, readErr := os.ReadDir(physical)
		if err != nil || statErr != nil || readErr != nil || physical != snapshot.paths[name] ||
			!os.SameFile(snapshot.directories[name], info) {
			return fmt.Errorf("original source directory identity changed")
		}
		names := make([]string, 0, len(entries))
		for _, entry := range entries {
			names = append(names, entry.Name()+":"+entry.Type().String())
		}
		if !reflect.DeepEqual(names, snapshot.entries[name]) {
			return fmt.Errorf("original source directory membership changed")
		}
	}
	return nil
}

func originalFileSHA256(name string, expected os.FileInfo) (_ string, result error) {
	file, err := os.Open(name)
	if err != nil {
		return "", err
	}
	defer func() { result = errors.Join(result, file.Close()) }()
	opened, err := file.Stat()
	if err != nil || !opened.Mode().IsRegular() || !os.SameFile(expected, opened) {
		return "", fmt.Errorf("original source file identity changed while opening")
	}
	hash := sha256.New()
	buffer := originalHashBuffers.Get().(*[64 << 10]byte)
	defer originalHashBuffers.Put(buffer)
	count, err := io.CopyBuffer(hash, io.LimitReader(file, fileByteLimit+1), buffer[:])
	if err != nil {
		return "", err
	}
	after, err := os.Lstat(name)
	if err != nil || !after.Mode().IsRegular() || !os.SameFile(opened, after) ||
		count != expected.Size() || count != after.Size() || count > fileByteLimit {
		return "", fmt.Errorf("original source file identity or extent changed while reading")
	}
	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}
