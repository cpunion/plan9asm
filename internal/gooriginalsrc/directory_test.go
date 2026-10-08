package gooriginalsrc

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestSnapshotNonrecursiveGraphMemberBounds(t *testing.T) {
	for _, role := range []string{"files across directories", "directory members"} {
		t.Run(role, func(t *testing.T) {
			root, err := filepath.EvalSymlinks(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			snapshot := NewSnapshot()
			if role == "directory members" {
				for i := 0; i < directoryLimit; i++ {
					if err := os.Mkdir(filepath.Join(root, fmt.Sprintf("directory-%04d", i)), 0700); err != nil {
						t.Fatal(err)
					}
				}
				if err := snapshot.Directory(root, false); err == nil {
					t.Fatal("nonrecursive original directory members acquired a fresh graph-wide count budget")
				}
				return
			}
			for _, part := range []string{"first", "second"} {
				directory := filepath.Join(root, part)
				if err := os.Mkdir(directory, 0700); err != nil {
					t.Fatal(err)
				}
				for i := 0; i < fileLimit/2+1; i++ {
					file, err := os.OpenFile(filepath.Join(directory, fmt.Sprintf("member-%05d", i)), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
					if err != nil {
						t.Fatal(err)
					}
					if err := file.Close(); err != nil {
						t.Fatal(err)
					}
				}
				if err := snapshot.Directory(directory, false); part == "first" {
					if err != nil {
						t.Fatal("actual first bounded inventory was rejected", err)
					}
				} else if err == nil {
					t.Fatal("nonrecursive original file members acquired a fresh graph-wide count budget")
				}
			}
		})
	}
}
