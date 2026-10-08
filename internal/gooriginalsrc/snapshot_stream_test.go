package gooriginalsrc

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"
)

type retainedSourceOpener interface {
	OpenRetained(string) (io.ReadCloser, error)
}

func TestSnapshotRetainedStreamPhysicalClosure(t *testing.T) {
	for _, role := range []string{"unchanged", "same inode growth", "replacement", "tail byte", "incomplete close"} {
		t.Run(role, func(t *testing.T) {
			root, err := filepath.EvalSymlinks(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			name := filepath.Join(root, "member")
			original := bytes.Repeat([]byte("original complete source\n"), 6000)
			if err := os.WriteFile(name, original, 0600); err != nil {
				t.Fatal(err)
			}
			snapshot := NewSnapshot()
			if err := snapshot.Retain(name); err != nil {
				t.Fatal(err)
			}
			opener, exists := any(snapshot).(retainedSourceOpener)
			if !exists {
				t.Fatal("complete retained tree has no strict streaming physical read route")
			}
			reader, err := opener.OpenRetained(name)
			if err != nil {
				t.Fatal(err)
			}
			defer reader.Close()
			var prefix [1]byte
			if _, err := io.ReadFull(reader, prefix[:]); err != nil || prefix[0] != original[0] {
				t.Fatal("actual original member could not be read before mutation", err)
			}
			switch role {
			case "same inode growth":
				before, err := os.Lstat(name)
				if err != nil {
					t.Fatal(err)
				}
				if err := os.Truncate(name, fileByteLimit+1); err != nil {
					t.Fatal(err)
				}
				after, err := os.Lstat(name)
				if err != nil || !os.SameFile(before, after) || after.Size() <= fileByteLimit {
					t.Fatal("the actual open member did not grow at the same inode", err)
				}
			case "replacement":
				if err := os.Rename(name, filepath.Join(t.TempDir(), "original")); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(name, original, 0600); err != nil {
					t.Fatal(err)
				}
			case "tail byte":
				changed := append([]byte(nil), original...)
				changed[len(changed)-1] ^= 1
				if err := os.WriteFile(name, changed, 0600); err != nil {
					t.Fatal(err)
				}
			case "incomplete close":
				if err := reader.Close(); err == nil {
					t.Fatal("an incomplete member close impersonated complete EOF")
				}
				return
			}
			count, readErr := io.Copy(io.Discard, reader)
			closeErr := reader.Close()
			if role == "unchanged" {
				if readErr != nil || closeErr != nil || count != int64(len(original)-1) {
					t.Fatal("original member stream did not close complete EOF", readErr, closeErr)
				}
			} else if readErr == nil && closeErr == nil {
				t.Fatal("a changed open physical member yielded a complete retained stream")
			}
		})
	}
}
