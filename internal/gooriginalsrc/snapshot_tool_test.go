package gooriginalsrc

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"
)

type retainedToolObserver interface {
	RetainTool(string) error
}

func requireRetainedToolObserver(t *testing.T, snapshot *Snapshot) retainedToolObserver {
	t.Helper()
	observer, ok := any(snapshot).(retainedToolObserver)
	if !ok {
		t.Fatal("original selected tool has no distinct bounded physical observation route")
	}
	return observer
}

func makeOriginalToolExtent(t *testing.T, root, base string, extent int64) string {
	t.Helper()
	name := filepath.Join(root, base)
	file, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	if err := file.Truncate(extent); err != nil {
		t.Fatal(err)
	}
	if extent > 0 {
		if _, err := file.WriteAt([]byte{0x52}, 0); err != nil {
			t.Fatal(err)
		}
		if _, err := file.WriteAt([]byte{0xa7}, extent-1); err != nil {
			t.Fatal(err)
		}
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	return name
}

func originalToolFixtureRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	return root
}

func originalToolFullDigest(t *testing.T, name string) string {
	t.Helper()
	file, err := os.Open(name)
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.New()
	_, readErr := io.Copy(digest, file)
	closeErr := file.Close()
	if readErr != nil || closeErr != nil {
		t.Fatal("independent complete tool bytes were not read", readErr, closeErr)
	}
	return fmt.Sprintf("%x", digest.Sum(nil))
}

func TestSnapshotTypedToolLimits(t *testing.T) {
	for _, extent := range []int64{17 << 20, 64 << 20, 64<<20 + 1} {
		t.Run(fmt.Sprintf("extent-%d", extent), func(t *testing.T) {
			name := makeOriginalToolExtent(t, originalToolFixtureRoot(t), "go", extent)
			snapshot := NewSnapshot()
			observer := requireRetainedToolObserver(t, snapshot)
			err := observer.RetainTool(name)
			if extent > 64<<20 {
				if err == nil || snapshot.bytes != 0 || len(snapshot.hashes) != 0 {
					t.Fatal("oversized tool acquired a retained observation")
				}
				return
			}
			if err != nil || snapshot.bytes != extent {
				t.Fatal("existing 64 MiB tool contract was narrowed to source bounds", err)
			}
			if snapshot.View().Hashes[name] != originalToolFullDigest(t, name) {
				t.Fatal("tool observation did not consume the complete independent bytes")
			}
			if err := snapshot.Verify(); err != nil {
				t.Fatal(err)
			}
			if err := observer.RetainTool(name); err != nil || snapshot.bytes != extent {
				t.Fatal("repeat tool retention consumed another byte budget", err)
			}
			if err := NewSnapshot().Retain(name); err == nil {
				t.Fatal("tool limit leaked into ordinary source/export retention")
			}
		})
	}
	for _, role := range []string{"source first", "tool first"} {
		t.Run("shared-budget-"+role, func(t *testing.T) {
			root := originalToolFixtureRoot(t)
			source := makeOriginalToolExtent(t, root, "source", 16<<20)
			tool := makeOriginalToolExtent(t, root, "compile", 48<<20)
			snapshot := NewSnapshot()
			observer := requireRetainedToolObserver(t, snapshot)
			if role == "source first" {
				if err := snapshot.Retain(source); err != nil {
					t.Fatal(err)
				}
			}
			if err := observer.RetainTool(tool); err != nil {
				t.Fatal(err)
			}
			if err := snapshot.Retain(source); err != nil || snapshot.bytes != totalByteLimit {
				t.Fatal("source/tool observations do not share the original total", err)
			}
			extra := makeOriginalToolExtent(t, root, "extra", 1)
			if err := snapshot.Retain(extra); err == nil {
				t.Fatal("new source acquired another graph-wide byte budget")
			}
			if err := observer.RetainTool(extra); err == nil {
				t.Fatal("new tool acquired another graph-wide byte budget")
			}
			if err := snapshot.Verify(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestSnapshotTypedToolRoleIsolation(t *testing.T) {
	for _, first := range []string{"source", "tool"} {
		for _, alias := range []string{"same path", "hard link", "parent alias"} {
			t.Run(first+"-"+alias, func(t *testing.T) {
				root := originalToolFixtureRoot(t)
				name := makeOriginalToolExtent(t, root, "member", 64<<10+1)
				snapshot := NewSnapshot()
				observer := requireRetainedToolObserver(t, snapshot)
				retainFirst, retainSecond := snapshot.Retain, observer.RetainTool
				if first == "tool" {
					retainFirst, retainSecond = retainSecond, retainFirst
				}
				if err := retainFirst(name); err != nil {
					t.Fatal(err)
				}
				other := name
				switch alias {
				case "hard link":
					other = filepath.Join(root, "alias")
					if err := os.Link(name, other); err != nil {
						t.Fatal(err)
					}
				case "parent alias":
					parent := filepath.Join(originalToolFixtureRoot(t), "alias")
					if err := os.Symlink(root, parent); err != nil {
						t.Fatal(err)
					}
					other = filepath.Join(parent, filepath.Base(name))
				}
				before, beforeErr := os.Stat(name)
				after, afterErr := os.Stat(other)
				if beforeErr != nil || afterErr != nil || !os.SameFile(before, after) {
					t.Fatal("cross-role alias does not identify the same physical fixture")
				}
				if err := retainSecond(other); err == nil || snapshot.bytes != before.Size() {
					t.Fatal("same source/tool inode acquired a different retained role", err)
				}
				if first == "tool" {
					if data, err := snapshot.ReadRetained(name); err == nil || data != nil {
						t.Fatal("retained tool yielded bytes through the source read route")
					}
					if reader, err := snapshot.OpenRetained(name); err == nil || reader != nil {
						if reader != nil {
							_ = reader.Close()
						}
						t.Fatal("retained tool yielded a source stream")
					}
					if err := snapshot.Directory(root, true); err == nil {
						t.Fatal("recursive source inventory silently relabeled the tool")
					}
				}
			})
		}
	}
}

func TestSnapshotTypedToolPhysicalClosure(t *testing.T) {
	for _, role := range []string{"unchanged", "tail byte", "same inode growth", "same inode shrink", "same bytes replacement", "symlink", "directory", "added member", "directory replacement"} {
		t.Run(role, func(t *testing.T) {
			root := originalToolFixtureRoot(t)
			name := makeOriginalToolExtent(t, root, "go", 17<<20)
			snapshot := NewSnapshot()
			observer := requireRetainedToolObserver(t, snapshot)
			if err := observer.RetainTool(name); err != nil {
				t.Fatal(err)
			}
			if err := snapshot.Directory(root, false); err != nil {
				t.Fatal(err)
			}
			before, err := os.Lstat(name)
			if err != nil {
				t.Fatal(err)
			}
			switch role {
			case "tail byte":
				file, err := os.OpenFile(name, os.O_WRONLY, 0600)
				if err != nil {
					t.Fatal(err)
				}
				_, writeErr := file.WriteAt([]byte{0xa6}, before.Size()-1)
				closeErr := file.Close()
				if writeErr != nil || closeErr != nil {
					t.Fatal(writeErr, closeErr)
				}
			case "same inode growth", "same inode shrink":
				extent := before.Size() + 1
				if role == "same inode shrink" {
					extent = before.Size() - 1
				}
				if err := os.Truncate(name, extent); err != nil {
					t.Fatal(err)
				}
				after, err := os.Lstat(name)
				if err != nil || !os.SameFile(before, after) || after.Size() != extent {
					t.Fatal("tool fixture did not actually change extent at the same inode", err)
				}
			case "same bytes replacement", "symlink", "directory":
				backup := filepath.Join(originalToolFixtureRoot(t), "original")
				if err := os.Rename(name, backup); err != nil {
					t.Fatal(err)
				}
				switch role {
				case "same bytes replacement":
					makeOriginalToolExtent(t, root, "go", before.Size())
					if originalToolFullDigest(t, name) != originalToolFullDigest(t, backup) {
						t.Fatal("replacement control changed bytes instead of only inode")
					}
				case "symlink":
					if err := os.Symlink(backup, name); err != nil {
						t.Fatal(err)
					}
				case "directory":
					if err := os.Mkdir(name, 0700); err != nil {
						t.Fatal(err)
					}
				}
			case "added member":
				makeOriginalToolExtent(t, root, "foreign", 1)
			case "directory replacement":
				backup := filepath.Join(originalToolFixtureRoot(t), "original")
				if err := os.Rename(root, backup); err != nil {
					t.Fatal(err)
				}
				if err := os.Mkdir(root, 0700); err != nil {
					t.Fatal(err)
				}
				if err := os.Rename(filepath.Join(backup, "go"), name); err != nil {
					t.Fatal(err)
				}
				after, err := os.Lstat(name)
				if err != nil || !os.SameFile(before, after) {
					t.Fatal("directory replacement did not preserve the original tool inode", err)
				}
			}
			err = snapshot.Verify()
			if role == "unchanged" {
				if err != nil {
					t.Fatal(err)
				}
			} else if err == nil {
				t.Fatal("changed tool file or complete directory observation was accepted")
			}
		})
	}
}

func TestOriginSetTypedTool(t *testing.T) {
	root := originalToolFixtureRoot(t)
	cache := filepath.Join(root, "cache")
	if err := os.Mkdir(cache, 0700); err != nil {
		t.Fatal(err)
	}
	tool := makeOriginalToolExtent(t, root, "go", 17<<20)
	set, err := NewOriginSet(cache, tool)
	if err != nil {
		t.Fatal("exact original download classified its selected tool as source/export", err)
	}
	if err := set.Verify(); err != nil {
		t.Fatal(err)
	}
	if data, err := set.files.ReadRetained(tool); err == nil || data != nil {
		t.Fatal("original download tool escaped through a source byte route")
	}
}
