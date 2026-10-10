package gooriginalsrc

import (
	"archive/zip"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/mod/module"
	"golang.org/x/mod/sumdb/dirhash"
	modzip "golang.org/x/mod/zip"
)

// Ordinary module archives and non-Go payload members are not individual Go
// source buffers. Stock Go's ZIP validator has a separate bounded contract.
// The fake receipt here proves only origin consistency, never ASM authority.
func TestOriginSetGoAcceptedArchiveAndPayloadExtents(t *testing.T) {
	for _, role := range []string{"archive", "payload member"} {
		t.Run(role, func(t *testing.T) {
			set, receipt, _ := originalZIPConsistencyFixture(t)
			method, members, extent := uint16(zip.Store), 3, int64(8<<20)
			if role == "payload member" {
				method, members, extent = zip.Deflate, 1, fileByteLimit+1
			}
			for index := 0; index < members; index++ {
				name := filepath.Join(receipt.Dir, string(rune('a'+index))+".bin")
				file, err := os.Create(name)
				if err != nil {
					t.Fatal(err)
				}
				_, copyErr := io.CopyN(file, zeroOriginalZIPReader{}, extent)
				if err := errors.Join(copyErr, file.Close()); err != nil {
					t.Fatal(err)
				}
			}
			originalLargeZIPFixture(t, receipt, method)
			if _, err := modzip.CheckZip(module.Version{Path: receipt.Path, Version: receipt.Version}, receipt.Zip); err != nil {
				t.Fatal("official Go ZIP validator rejected the independent positive", err)
			}
			info, err := os.Stat(receipt.Zip)
			if err != nil {
				t.Fatal(err)
			}
			if role == "archive" && info.Size() <= fileByteLimit || role == "payload member" && info.Size() >= fileByteLimit {
				t.Fatal("fixture did not isolate its archive/member boundary", info.Size())
			}
			receipt.Sum, err = dirhash.HashZip(receipt.Zip, dirhash.Hash1)
			if err != nil {
				t.Fatal(err)
			}
			tree, err := dirhash.HashDir(receipt.Dir, receipt.Path+"@"+receipt.Version, dirhash.Hash1)
			if err != nil || tree != receipt.Sum {
				t.Fatal("independent complete original ZIP/tree hashes differ", err)
			}
			observed, err := set.Observe(context.Background(), receipt.Path, receipt.Version, receipt)
			if err != nil || observed.Root != receipt.Dir {
				t.Fatal("Go-accepted complete archive/tree was treated as a small source buffer", err)
			}
			if err := set.Verify(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func originalLargeZIPFixture(t *testing.T, receipt DownloadReceipt, method uint16) {
	t.Helper()
	file, err := os.Create(receipt.Zip)
	if err != nil {
		t.Fatal(err)
	}
	writer := zip.NewWriter(file)
	err = filepath.WalkDir(receipt.Dir, func(name string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil || entry.IsDir() {
			return walkErr
		}
		relative, err := filepath.Rel(receipt.Dir, name)
		if err != nil {
			return err
		}
		member, err := writer.CreateHeader(&zip.FileHeader{
			Name: receipt.Path + "@" + receipt.Version + "/" + filepath.ToSlash(relative), Method: method,
		})
		if err != nil {
			return err
		}
		input, err := os.Open(name)
		if err != nil {
			return err
		}
		_, copyErr := io.Copy(member, input)
		return errors.Join(copyErr, input.Close())
	})
	if err := errors.Join(err, writer.Close(), file.Close()); err != nil {
		t.Fatal(err)
	}
}
