package gotoolprofile

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/mod/module"
)

// CaptureDownloadedProxyGoMod authenticates an actual Go-downloaded metadata
// file which is not in the original ZIP. This grants no source/package role;
// callers still prove ZIP absence, full original inventory and actual imports.
func CaptureDownloadedProxyGoMod(ctx context.Context, binary, directory string, env []string, name, version, zipSum, modSum, metadata string, prior *ProxyGoModProof) (*ProxyGoModProof, string, error) {
	output, _, err := RunBounded(ctx, directory, env, binary, "env", "-json", "GOMODCACHE", "GOPROXY")
	if err != nil {
		return nil, "", err
	}
	var actual map[string]string
	if err := json.Unmarshal(output, &actual); err != nil || len(actual) != 2 || !filepath.IsAbs(actual["GOMODCACHE"]) {
		return nil, "", fmt.Errorf("invalid actual metadata cache observation")
	}
	escapedModule, err := module.EscapePath(name)
	if err != nil {
		return nil, "", err
	}
	escapedVersion, err := module.EscapeVersion(version)
	if err != nil {
		return nil, "", err
	}
	wanted := filepath.Join(actual["GOMODCACHE"], "cache", "download", filepath.FromSlash(escapedModule), "@v", escapedVersion+".mod")
	if filepath.Clean(metadata) != wanted {
		return nil, "", fmt.Errorf("proxy go.mod is not the actual exact-version download metadata")
	}
	info, err := os.Lstat(wanted)
	if err != nil || !info.Mode().IsRegular() {
		return nil, "", fmt.Errorf("proxy metadata is not a regular exact-version download file")
	}
	canonical, err := filepath.EvalSymlinks(wanted)
	if err != nil {
		return nil, "", err
	}
	contents, err := ReadChecksumFile(canonical, 1024)
	if err != nil {
		return nil, "", err
	}
	if prior != nil {
		// Reuse a complete authenticated producer checkpoint, not its claim of
		// success. Reverify its signature/inclusion and actual downloaded bytes;
		// a newer public checkpoint is unnecessary for the same immutable pair.
		if err := ValidateProxyGoMod(prior); err != nil {
			return nil, "", err
		}
		if prior.Module != name || prior.Version != version || prior.ModuleSum != zipSum || prior.GoModSum != modSum ||
			prior.Contents != string(contents) || prior.SHA256 != bytesSHA256(contents) {
			return nil, "", fmt.Errorf("downloaded metadata differs from authenticated producer origin")
		}
		proof := *prior
		proof.Inclusion = append([]string(nil), prior.Inclusion...)
		return &proof, canonical, nil
	}
	reader, err := NewSumDBReader(ctx, actual["GOMODCACHE"], actual["GOPROXY"])
	if err != nil {
		return nil, "", err
	}
	lookup, err := reader.Read("lookup/"+escapedModule+"@"+escapedVersion, 16<<10)
	if err != nil {
		return nil, "", err
	}
	proof, err := CaptureProxyGoMod(name, version, zipSum, modSum, contents, lookup, reader)
	if err != nil {
		return nil, "", err
	}
	return proof, canonical, nil
}
