package gotoolprofile

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The file proxy contains only this repository's generated first-party code.
// GOSUMDB=off here is not a production source-authentication certificate.
func amd64DownloadedObservationFixture(t *testing.T, variant string) (AMD64OriginalObservationRequest, string, string) {
	t.Helper()
	request := amd64OriginalObservationFixture(t, variant)
	modulePath, version := "example.invalid/observation", "v1.0.0"
	if variant == "alias" {
		modulePath = "example.invalid/archive-observation"
	}
	proxy, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	cache, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	root := request.Directory
	if variant == "nested" {
		root = filepath.Dir(filepath.Dir(root))
	}
	var archive bytes.Buffer
	w := zip.NewWriter(&archive)
	err = filepath.WalkDir(root, func(name string, entry os.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}
		relative, err := filepath.Rel(root, name)
		if err != nil {
			return err
		}
		member, err := w.Create(modulePath + "@" + version + "/" + filepath.ToSlash(relative))
		if err != nil {
			return err
		}
		data, err := os.ReadFile(name)
		if err != nil {
			return err
		}
		_, err = member.Write(data)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	goMod, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(proxy, modulePath, "@v")
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	for suffix, data := range map[string][]byte{
		".zip": archive.Bytes(), ".mod": goMod,
		".info": []byte(`{"Version":"v1.0.0","Time":"2026-10-01T00:00:00Z"}`),
	} {
		if err := os.WriteFile(filepath.Join(dir, version+suffix), data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	t.Cleanup(func() {
		if err := filepath.WalkDir(cache, func(name string, entry os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() {
				return os.Chmod(name, 0700)
			}
			return nil
		}); err != nil {
			t.Error(err)
		}
	})
	request.Environment = replaceEnv(request.Environment, map[string]string{
		"GOPROXY": "file://" + proxy, "GOMODCACHE": cache,
	})
	ctx, cancel := amd64OriginalObservationContext(t)
	defer cancel()
	data, stderr, err := RunBounded(ctx, root, request.Environment, request.GoBinary,
		"mod", "download", "-json", modulePath+"@"+version)
	if err != nil || len(stderr) != 0 {
		t.Fatalf("actual generated module download: %v; %s", err, stderr)
	}
	var downloaded struct{ Path, Version, Dir, Error string }
	if err := json.Unmarshal(data, &downloaded); err != nil || downloaded.Error != "" ||
		downloaded.Path != modulePath || downloaded.Version != version || downloaded.Dir == "" {
		t.Fatalf("actual Go omitted its exact downloaded source identity: %v", err)
	}
	request.Directory = downloaded.Dir
	if variant == "nested" {
		request.Directory = filepath.Join(downloaded.Dir, "internal", "kernel")
	}
	return request, modulePath, version
}

// Keep the source/download identity separate from Go's actual cached-CWD role.
func captureAMD64DownloadedForTest(ctx context.Context, request AMD64OriginalObservationRequest, modulePath, version string) (*AMD64OriginalPackageObservation, error) {
	return CaptureAMD64DownloadedPackageObservation(ctx, AMD64DownloadedObservationRequest{
		Original: request, SourceModule: modulePath, SourceVersion: version,
	}, RunBounded)
}

func TestAMD64DownloadedObservationActualOriginalCache(t *testing.T) {
	for _, variant := range []string{"nested", "data", "sibling", "alias"} {
		t.Run(variant, func(t *testing.T) {
			request, modulePath, version := amd64DownloadedObservationFixture(t, variant)
			ctx, cancel := amd64OriginalObservationContext(t)
			defer cancel()
			if old, err := CaptureAMD64OriginalPackageObservation(ctx, request, RunBounded); err == nil {
				old.Close()
				t.Fatal("downloaded source acquired an owned-main role without an authenticated handoff")
			}
			observed, err := captureAMD64DownloadedForTest(ctx, request, modulePath, version)
			if err != nil {
				t.Fatal(err)
			}
			defer observed.Close()
			pkg := observed.state.packages[request.PackagePath]
			if pkg.Module == nil || !pkg.Module.Main || pkg.Module.Version != "" ||
				observed.state.declarations["Value"] != "func(a uint64) uint64" {
				t.Fatal("actual original cache selection or complete checked declaration was lost")
			}
			if err := observed.Verify(ctx); err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(pkg.Dir, "@"+version) {
				t.Fatal("source was copied into a synthetic owned main instead of observing original cache bytes")
			}
		})
	}
}
