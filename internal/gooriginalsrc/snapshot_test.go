package gooriginalsrc

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestSnapshotFullEOFAndPhysicalRoles(t *testing.T) {
	for _, role := range []string{"unchanged", "tail byte", "same bytes replacement", "symlink", "added member"} {
		t.Run(role, func(t *testing.T) {
			root, err := filepath.EvalSymlinks(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			name := filepath.Join(root, "source")
			bytes := make([]byte, 64<<10+227)
			for i := range bytes {
				bytes[i] = byte(i)
			}
			if err := os.WriteFile(name, bytes, 0600); err != nil {
				t.Fatal(err)
			}
			snapshot := NewSnapshot()
			if err := snapshot.Directory(root, true); err != nil {
				t.Fatal(err)
			}
			if snapshot.View().Hashes[name] != fmt.Sprintf("%x", sha256.Sum256(bytes)) {
				t.Fatal("snapshot did not hash the complete original tail through EOF")
			}
			switch role {
			case "tail byte":
				bytes[len(bytes)-1] ^= 1
				if err := os.WriteFile(name, bytes, 0600); err != nil {
					t.Fatal(err)
				}
			case "same bytes replacement", "symlink":
				backup := filepath.Join(t.TempDir(), "original")
				if err := os.Rename(name, backup); err != nil {
					t.Fatal(err)
				}
				if role == "symlink" {
					if err := os.Symlink(backup, name); err != nil {
						t.Fatal(err)
					}
				} else if err := os.WriteFile(name, bytes, 0600); err != nil {
					t.Fatal(err)
				}
			case "added member":
				if err := os.WriteFile(filepath.Join(root, "ignored"), nil, 0600); err != nil {
					t.Fatal(err)
				}
			}
			err = snapshot.Verify()
			if role == "unchanged" && err != nil {
				t.Fatal(err)
			}
			if role != "unchanged" && err == nil {
				t.Fatal("changed physical source role was accepted")
			}
		})
	}
}

func TestSnapshotOriginalSharedBounds(t *testing.T) {
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	snapshot := NewSnapshot()
	bytes := make([]byte, fileByteLimit)
	for i := 0; i < totalByteLimit/fileByteLimit; i++ {
		name := filepath.Join(root, fmt.Sprintf("member-%d", i))
		if err := os.WriteFile(name, bytes, 0600); err != nil {
			t.Fatal(err)
		}
		if err := snapshot.Retain(name); err != nil {
			t.Fatal("existing graph-wide byte budget rejected its boundary", err)
		}
	}
	if snapshot.bytes != totalByteLimit {
		t.Fatal("graph-wide byte accounting is incomplete")
	}
	if err := snapshot.Verify(); err != nil {
		t.Fatal(err)
	}
	overflow := filepath.Join(root, "extra")
	if err := os.WriteFile(overflow, []byte{1}, 0600); err != nil {
		t.Fatal(err)
	}
	if err := snapshot.Retain(overflow); err == nil {
		t.Fatal("another module/member acquired a fresh byte budget")
	}
}

func TestOriginSetInvalidRoles(t *testing.T) {
	for _, cache := range []string{"", ".", "relative/cache"} {
		if _, err := NewOriginSet(cache, "/nonexistent/tool"); err == nil {
			t.Fatal("nonabsolute or missing source/tool role was accepted")
		}
	}
	if err := (*Snapshot)(nil).Verify(); err == nil {
		t.Fatal("nil source snapshot was accepted")
	}
	if err := (*OriginSet)(nil).Verify(); err == nil {
		t.Fatal("nil exact-source guard was accepted")
	}
}
