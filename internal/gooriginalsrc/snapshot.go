// Package gooriginalsrc observes bounded original files and exact Go-download
// origins. Its state is byte/role consistency data, never an ASM capability.
package gooriginalsrc

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"hash"
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
	hashes, paths  map[string]string
	entries        map[string][]string
	directories    map[string]os.FileInfo
	files          map[string]os.FileInfo
	fileRoles      map[string]bool
	directoryRoles map[string]bool
	bytes          int64
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
		files:     make(map[string]os.FileInfo),
		fileRoles: make(map[string]bool), directoryRoles: make(map[string]bool),
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
	if !snapshot.fileRoles[physical] && len(snapshot.fileRoles) >= fileLimit {
		return fmt.Errorf("original file member inventory exceeds its shared count bound")
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
	snapshot.fileRoles[physical] = true
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
	entries, info, err := ReadDirectory(name, fileLimit+directoryLimit)
	if err != nil {
		return err
	}
	if prior, present := snapshot.directories[name]; present &&
		(!os.SameFile(prior, info) || snapshot.paths[name] != physical) {
		return fmt.Errorf("original source directory identity changed")
	}
	newFiles, newDirectories := make(map[string]bool), make(map[string]bool)
	if !snapshot.directoryRoles[physical] {
		newDirectories[physical] = true
	}
	for _, entry := range entries {
		child := filepath.Join(physical, entry.Name())
		if entry.Type()&os.ModeSymlink != 0 || !entry.IsDir() && !entry.Type().IsRegular() {
			return fmt.Errorf("original source directory contains an unbound nonregular entry")
		}
		if entry.IsDir() {
			if !snapshot.directoryRoles[child] {
				newDirectories[child] = true
			}
		} else if !snapshot.fileRoles[child] {
			newFiles[child] = true
		}
	}
	if len(snapshot.fileRoles)+len(newFiles) > fileLimit || len(snapshot.directoryRoles)+len(newDirectories) > directoryLimit {
		return fmt.Errorf("complete original directory members exceed the shared graph count bound")
	}
	for child := range newFiles {
		snapshot.fileRoles[child] = true
	}
	for child := range newDirectories {
		snapshot.directoryRoles[child] = true
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
		entries, info, readErr := ReadDirectory(name, fileLimit+directoryLimit)
		if err != nil || readErr != nil || physical != snapshot.paths[name] ||
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
	reader, err := snapshot.OpenRetained(name)
	if err != nil {
		return nil, err
	}
	buffer := &originalFileBuffer{bytes: make([]byte, int(snapshot.files[name].Size()))}
	err = copyOriginalFile(buffer, reader)
	if err != nil {
		return nil, err
	}
	return buffer.bytes, nil
}

// OpenRetained exposes only an observed member's strict physical stream. Its
// complete EOF and close errors must be consumed; it does not mint permissions.
func (snapshot *Snapshot) OpenRetained(name string) (io.ReadCloser, error) {
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
	return openOriginalFile(name, expected, physical, digest)
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
	physical, err := filepath.EvalSymlinks(name)
	if err != nil {
		return "", err
	}
	reader, err := openOriginalFile(name, expected, physical, "")
	if err != nil {
		return "", err
	}
	if err := copyOriginalFile(io.Discard, reader); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", reader.hash.Sum(nil)), nil
}

func copyOriginalFile(output io.Writer, reader io.ReadCloser) error {
	buffer := originalHashBuffers.Get().(*[64 << 10]byte)
	defer originalHashBuffers.Put(buffer)
	_, err := io.CopyBuffer(output, reader, buffer[:])
	return errors.Join(err, reader.Close())
}

type originalFileReader struct {
	file     *os.File
	name     string
	physical string
	expected os.FileInfo
	digest   string
	hash     hash.Hash
	count    int64
	complete bool
	closed   bool
	failure  error
}

func openOriginalFile(name string, expected os.FileInfo, physical, digest string) (*originalFileReader, error) {
	if expected == nil || expected.Size() < 0 || expected.Size() > fileByteLimit {
		return nil, fmt.Errorf("original source file has no bounded physical extent")
	}
	before, err := os.Lstat(name)
	resolved, resolveErr := filepath.EvalSymlinks(name)
	if err != nil || resolveErr != nil || resolved != physical || !before.Mode().IsRegular() ||
		!os.SameFile(expected, before) || before.Size() != expected.Size() {
		return nil, fmt.Errorf("original source file identity or extent changed before opening")
	}
	file, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	opened, err := file.Stat()
	if err != nil || !opened.Mode().IsRegular() || !os.SameFile(expected, opened) || opened.Size() != expected.Size() {
		return nil, errors.Join(fmt.Errorf("original source file identity changed while opening"), file.Close())
	}
	return &originalFileReader{file: file, name: name, physical: physical,
		expected: expected, digest: digest, hash: sha256.New()}, nil
}

func (reader *originalFileReader) Read(data []byte) (int, error) {
	if reader.failure != nil {
		return 0, reader.failure
	}
	if reader.closed {
		return 0, os.ErrClosed
	}
	if reader.complete {
		return 0, io.EOF
	}
	if remaining := reader.expected.Size() - reader.count + 1; int64(len(data)) > remaining {
		data = data[:int(remaining)]
	}
	count, err := reader.file.Read(data)
	reader.count += int64(count)
	_, _ = reader.hash.Write(data[:count])
	if reader.count > reader.expected.Size() {
		err = fmt.Errorf("original source file grew beyond its retained extent")
	} else if err == io.EOF {
		after, statErr := os.Lstat(reader.name)
		physical, resolveErr := filepath.EvalSymlinks(reader.name)
		if statErr != nil || resolveErr != nil || physical != reader.physical || !after.Mode().IsRegular() ||
			!os.SameFile(reader.expected, after) || reader.count != reader.expected.Size() || reader.count != after.Size() ||
			reader.digest != "" && fmt.Sprintf("%x", reader.hash.Sum(nil)) != reader.digest {
			err = fmt.Errorf("original source file identity, bytes or extent changed while reading")
		} else {
			reader.complete = true
		}
	}
	if err != nil && err != io.EOF {
		reader.failure = err
	}
	return count, err
}

func (reader *originalFileReader) Close() error {
	if reader.closed {
		return reader.failure
	}
	reader.closed = true
	if !reader.complete {
		reader.failure = errors.Join(reader.failure, fmt.Errorf("original source member did not reach complete EOF"))
	}
	reader.failure = errors.Join(reader.failure, reader.file.Close())
	return reader.failure
}
