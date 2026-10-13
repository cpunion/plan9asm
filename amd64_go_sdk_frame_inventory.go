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
// child identities are rechecked before/after commands. Existing ordinary/CPU
// flat inventory and the 256-member/16MiB primary bounds stay unchanged.
func (source *GoX86SDKASMDataSource) snapshotSDKPrimaryDirectory() error {
	guard := source.guard
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
			if err := source.retainDirectory(path); err != nil {
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
