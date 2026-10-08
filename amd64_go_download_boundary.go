package plan9asm

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"

	"github.com/xgo-dev/plan9asm/internal/gooriginalsrc"
)

const goAMD64DownloadBoundaryModule = "module example.invalid/plan9asm-frame-download\ngo 1.20\n"

// This concrete owner holds temporary IO only, never source/frame authority.
type goAMD64DownloadBoundary struct {
	directory, name         string
	directoryInfo, fileInfo os.FileInfo
	file                    *os.File
	written                 []byte
	fileClosed, fileRemoved bool
	directoryRemoved        bool
	closeErr                error
}

func newGoAMD64DownloadBoundary(temporary string) (*goAMD64DownloadBoundary, error) {
	created, err := os.MkdirTemp(temporary, "plan9asm-frame-download-")
	if err != nil {
		return nil, err
	}
	boundary := &goAMD64DownloadBoundary{directory: created}
	boundary.directoryInfo, err = os.Lstat(created)
	if err != nil {
		return nil, err
	}
	boundary.directory, err = filepath.EvalSymlinks(created)
	if err != nil {
		return nil, err
	}
	boundary.name = filepath.Join(boundary.directory, "go.mod")
	boundary.file, err = os.OpenFile(boundary.name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return nil, err
	}
	boundary.fileInfo, err = boundary.file.Stat()
	if err != nil {
		return nil, errors.Join(err, boundary.closeFile())
	}
	return boundary, nil
}

func (boundary *goAMD64DownloadBoundary) prepare() error {
	n, err := boundary.file.WriteString(goAMD64DownloadBoundaryModule)
	boundary.written = []byte(goAMD64DownloadBoundaryModule[:n])
	if err := errors.Join(err, boundary.closeFile()); err != nil {
		return err
	}
	return boundary.verify()
}

func (boundary *goAMD64DownloadBoundary) closeFile() error {
	if boundary.file != nil && !boundary.fileClosed {
		boundary.fileClosed = true
		boundary.closeErr = boundary.file.Close()
	}
	return boundary.closeErr
}

func (boundary *goAMD64DownloadBoundary) verify() error {
	entries, info, err := gooriginalsrc.ReadDirectory(boundary.directory, 1)
	if err != nil || boundary.directoryInfo == nil || !os.SameFile(boundary.directoryInfo, info) {
		return errors.Join(goAMD64FrameContext("download command boundary directory changed"), err)
	}
	if len(entries) != 1 || entries[0].Name() != "go.mod" || boundary.fileInfo == nil {
		return goAMD64FrameContext("download command boundary acquired unexpected or missing members")
	}
	actual, err := os.Lstat(boundary.name)
	if err != nil || !actual.Mode().IsRegular() || !os.SameFile(boundary.fileInfo, actual) || actual.Size() != int64(len(boundary.written)) {
		return errors.Join(goAMD64FrameContext("download boundary metadata role/extent changed"), err)
	}
	snapshot := gooriginalsrc.NewSnapshot()
	if err := snapshot.Retain(boundary.name); err != nil {
		return err
	}
	data, err := snapshot.ReadRetained(boundary.name)
	if err != nil || !bytes.Equal(data, boundary.written) {
		return errors.Join(goAMD64FrameContext("download boundary metadata bytes changed"), err)
	}
	after, err := os.Lstat(boundary.name)
	if err != nil || !os.SameFile(boundary.fileInfo, after) {
		return errors.Join(goAMD64FrameContext("download boundary metadata identity changed"), err)
	}
	if err := snapshot.Verify(); err != nil {
		return err
	}
	entries, info, err = gooriginalsrc.ReadDirectory(boundary.directory, 1)
	if err != nil || !os.SameFile(boundary.directoryInfo, info) || len(entries) != 1 || entries[0].Name() != "go.mod" {
		return errors.Join(goAMD64FrameContext("download boundary final inventory changed"), err)
	}
	return nil
}

func (boundary *goAMD64DownloadBoundary) Close() error {
	if boundary == nil {
		return nil
	}
	result := boundary.closeFile()
	if boundary.directoryRemoved {
		return result
	}
	if boundary.fileInfo != nil && !boundary.fileRemoved {
		if err := boundary.verify(); err != nil {
			return errors.Join(result, err)
		}
		if err := os.Remove(boundary.name); err != nil {
			return errors.Join(result, err)
		}
		boundary.fileRemoved = true
	}
	entries, info, err := gooriginalsrc.ReadDirectory(boundary.directory, 0)
	if err != nil || boundary.directoryInfo == nil || !os.SameFile(boundary.directoryInfo, info) || len(entries) != 0 {
		return errors.Join(result, goAMD64FrameContext("download boundary cannot remove a foreign or nonempty directory"), err)
	}
	err = os.Remove(boundary.directory)
	if err == nil {
		boundary.directoryRemoved = true
	}
	return errors.Join(result, err)
}
