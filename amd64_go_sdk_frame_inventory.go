package plan9asm

import (
	"crypto/sha256"
	"path/filepath"

	"github.com/xgo-dev/plan9asm/internal/gooriginalsrc"
	"github.com/xgo-dev/plan9asm/internal/sourcepath"
)

// This distinct SDK-primary role retains every direct member, including all
// inactive candidates and child-directory identities. Child contents belong
// to separate packages/generators, not the selected compiler package. Any
// selected dependency retains its independent complete original source/export
// graph; active relative CPP includes remain unsupported. Root membership and
// child identities are rechecked before/after commands. The ordinary owned
// flat inventory and the 256-member/16MiB SDK bounds stay unchanged.
func (source *GoX86SDKASMDataSource) snapshotSDKPrimaryDirectory() error {
	return captureSDKPackageDirectoryInputs(source.guard, source.retainDirectory)
}

// SDK assembly, numerical reads and literal alignment share only this input
// inventory. It grants neither child-package source/include authority nor any
// frame, native entry or initialized-storage role. Each caller authenticates
// its own complete SDK selection and retains/rechecks every child identity.
func captureSDKPackageDirectoryInputs(guard *GoAMD64FrameSource, retainDirectory func(string) error) error {
	resolved, err := sourcepath.Resolve(guard.request.Directory)
	if err != nil || resolved != guard.request.Directory {
		return sdkASMDataContext("SDK-primary package directory is not canonical")
	}
	entries, info, err := gooriginalsrc.ReadDirectory(resolved, 256)
	if err != nil || len(entries) == 0 || len(entries) > 256 {
		return sdkASMDataContext("SDK-primary complete package inventory exceeds its bound")
	}
	guard.directoryInfo = info
	remaining := int64(16 << 20)
	for _, entry := range entries {
		guard.entries = append(guard.entries, entry.Name())
		path := filepath.Join(resolved, entry.Name())
		if entry.IsDir() {
			if err := retainDirectory(path); err != nil {
				return err
			}
			continue
		}
		data, err := goAMD64FrameReadBound(path, remaining)
		if err != nil {
			return err
		}
		digest := sha256.Sum256(data)
		if previous, present := guard.inputs[path]; present && previous != digest {
			return sdkASMDataContext("SDK-primary original inventory changed while capturing")
		}
		remaining -= int64(len(data))
		guard.inputs[path] = digest
	}
	return nil
}
