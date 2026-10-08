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
	if snapshot == nil || snapshot.hashes == nil || !filepath.IsAbs(name) {
		return fmt.Errorf("original file path is missing or over-bound")
	}
	if _, retained := snapshot.hashes[name]; !retained && len(snapshot.hashes) >= fileLimit {
		return fmt.Errorf("original file inventory exceeds its count bound")
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
		if info.Size() > totalByteLimit-snapshot.bytes {
			return fmt.Errorf("original source/export inventory exceeds its byte bound")
		}
		snapshot.bytes += info.Size()
	}
	snapshot.hashes[name], snapshot.paths[name], snapshot.files[name] = digest, physical, info
	return nil
}

func (snapshot *Snapshot) Directory(name string, recursive bool) error {
	if snapshot == nil || snapshot.directories == nil || !filepath.IsAbs(name) {
		return fmt.Errorf("original source directory requires an absolute initialized role")
	}
	if _, retained := snapshot.directories[name]; !retained && len(snapshot.directories) >= directoryLimit {
		return fmt.Errorf("original source directory inventory exceeds its count bound")
	}
	physical, err := filepath.EvalSymlinks(name)
	if err != nil {
		return err
	}
	info, err := os.Lstat(physical)
	if err != nil || !info.IsDir() {
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
	if snapshot == nil || snapshot.hashes == nil || snapshot.files == nil || snapshot.directories == nil {
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

// ReadRetained reads only an independently retained physical file. It consumes
// the same 16 MiB bound and complete EOF guard as hashing; the returned bytes
// are observations, not frame or source permissions.
func (snapshot *Snapshot) ReadRetained(name string) ([]byte, error) {
	if snapshot == nil || !filepath.IsAbs(name) {
		return nil, fmt.Errorf("missing retained original file role")
	}
	expected := snapshot.files[name]
	digest, retained := snapshot.hashes[name]
	physical, err := filepath.EvalSymlinks(name)
	if !retained || expected == nil || err != nil || physical != snapshot.paths[name] {
		return nil, fmt.Errorf("original file was not retained at its physical role")
	}
	if expected.Size() < 0 || expected.Size() > fileByteLimit {
		return nil, fmt.Errorf("retained original file exceeds its byte bound")
	}
	buffer := &originalFileBuffer{bytes: make([]byte, int(expected.Size()))}
	actual, err := originalFileRead(name, expected, buffer)
	if err != nil {
		return nil, err
	}
	after, err := filepath.EvalSymlinks(name)
	if err != nil || after != physical || actual != digest {
		return nil, fmt.Errorf("retained original file bytes or physical role changed")
	}
	return buffer.bytes, nil
}

// A fixed extent prevents io.Copy from allocating extra storage if an original
// metadata file grows after the observation. Overflow rejects, never truncates.
type originalFileBuffer struct {
	bytes  []byte
	offset int
}

func (buffer *originalFileBuffer) Write(data []byte) (int, error) {
	if len(data) > len(buffer.bytes)-buffer.offset {
		return 0, fmt.Errorf("original source file grew beyond its retained extent")
	}
	count := copy(buffer.bytes[buffer.offset:], data)
	buffer.offset += count
	return count, nil
}

func originalFileSHA256(name string, expected os.FileInfo) (string, error) {
	return originalFileRead(name, expected, nil)
}

func originalFileRead(name string, expected os.FileInfo, output io.Writer) (_ string, result error) {
	if expected == nil || expected.Size() < 0 || expected.Size() > fileByteLimit {
		return "", fmt.Errorf("original source file has no bounded physical extent")
	}
	file, err := os.Open(name)
	if err != nil {
		return "", err
	}
	defer func() { result = errors.Join(result, file.Close()) }()
	opened, err := file.Stat()
	if err != nil || !opened.Mode().IsRegular() || !os.SameFile(expected, opened) || opened.Size() != expected.Size() {
		return "", fmt.Errorf("original source file identity changed while opening")
	}
	hash := sha256.New()
	var sink io.Writer = hash
	if output != nil {
		sink = io.MultiWriter(hash, output)
	}
	buffer := originalHashBuffers.Get().(*[64 << 10]byte)
	defer originalHashBuffers.Put(buffer)
	count, err := io.CopyBuffer(sink, io.LimitReader(file, fileByteLimit+1), buffer[:])
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
