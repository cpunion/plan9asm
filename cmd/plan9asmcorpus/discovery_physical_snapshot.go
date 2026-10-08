package main

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type discoveryPhysicalMember struct {
	name string
	info os.FileInfo
}

// Only the filesystem inventory is retained. Record bytes are streamed into
// the digest, never decoded or kept as another in-memory ledger.
type discoveryPhysicalSnapshot struct {
	root    string
	digest  string
	members []discoveryPhysicalMember
}

func collectDiscoveryPhysicalSnapshot(directory string) (discoveryPhysicalSnapshot, error) {
	var snapshot discoveryPhysicalSnapshot
	root, err := filepath.EvalSymlinks(directory)
	if err != nil {
		return snapshot, err
	}
	root, err = filepath.Abs(root)
	if err != nil {
		return snapshot, err
	}
	snapshot.root = root
	digest := sha256.New()
	err = filepath.WalkDir(root, func(name string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		info, err := os.Lstat(name)
		if err != nil {
			return err
		}
		if !info.IsDir() && !info.Mode().IsRegular() {
			return fmt.Errorf("non-regular physical ledger member")
		}
		relative, err := filepath.Rel(root, name)
		if err != nil {
			return err
		}
		snapshot.members = append(snapshot.members, discoveryPhysicalMember{relative, info})
		fmt.Fprintf(digest, "%q:%d:%d:", filepath.ToSlash(relative), info.Mode(), info.Size())
		if info.IsDir() {
			return nil
		}
		file, err := os.Open(name)
		if err != nil {
			return err
		}
		opened, statErr := file.Stat()
		if statErr != nil || !os.SameFile(info, opened) {
			return errors.Join(fmt.Errorf("physical ledger member changed while opening"), statErr, file.Close())
		}
		n, readErr := io.Copy(digest, file)
		final, statErr := file.Stat()
		closeErr := file.Close()
		if err := errors.Join(readErr, statErr, closeErr); err != nil {
			return err
		}
		if n != info.Size() || !sameDiscoveryPhysicalMember(info, final) {
			return fmt.Errorf("physical ledger member changed while reading")
		}
		return nil
	})
	if err != nil {
		return discoveryPhysicalSnapshot{}, err
	}
	snapshot.digest = fmt.Sprintf("%x", digest.Sum(nil))
	return snapshot, nil
}

func sameDiscoveryPhysicalMember(before, after os.FileInfo) bool {
	return os.SameFile(before, after) && before.Mode() == after.Mode() &&
		before.Size() == after.Size() && before.ModTime().Equal(after.ModTime())
}

func verifyDiscoveryPhysicalSnapshot(directory string, before discoveryPhysicalSnapshot) error {
	after, err := collectDiscoveryPhysicalSnapshot(directory)
	if err != nil {
		return fmt.Errorf("committed assembly ledger changed: %w", err)
	}
	if before.root != after.root || before.digest != after.digest || len(before.members) != len(after.members) {
		return fmt.Errorf("committed assembly ledger changed while verifying reports")
	}
	for index, member := range before.members {
		if member.name != after.members[index].name || !sameDiscoveryPhysicalMember(member.info, after.members[index].info) {
			return fmt.Errorf("committed assembly ledger changed while verifying reports")
		}
	}
	return nil
}

func verifyDiscoveryAggregateInputs(root, scan, stored string, source discoverySourceIdentity,
	scanSHA string, storedSnapshot discoveryPhysicalSnapshot) error {
	finalSource, err := collectDiscoverySource(root)
	if err != nil {
		return err
	}
	if finalSource != source {
		return fmt.Errorf("discovery source changed while verifying reports")
	}
	finalScanSHA, err := discoveryLedgerFingerprint(scan)
	if err != nil {
		return err
	}
	if finalScanSHA != scanSHA {
		return fmt.Errorf("discovery ledger changed while verifying reports")
	}
	if stored != "" {
		return verifyDiscoveryPhysicalSnapshot(stored, storedSnapshot)
	}
	return nil
}
