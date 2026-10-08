package gotoolprofile

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAMD64DownloadedObservationRejectsUnprovedInputs(t *testing.T) {
	for _, name := range []string{"latest", "different version", "different module", "owned copy", "changed source", "added ignored source", "changed ZIP", "changed metadata"} {
		t.Run(name, func(t *testing.T) {
			request, path, version := amd64DownloadedObservationFixture(t, "data")
			ctx, cancel := amd64OriginalObservationContext(t)
			defer cancel()
			prefix := filepath.Join(originalObservationEnvironment(request.Environment, "GOMODCACHE"), "cache", "download", path, "@v", version)
			switch name {
			case "latest":
				version = "latest"
			case "different version":
				version = "v1.0.1"
			case "different module":
				path = "example.invalid/unobserved"
			case "owned copy":
				request.Directory = amd64OriginalObservationFixture(t, "data").Directory
			case "changed source":
				mutateAMD64DownloadedFile(t, filepath.Join(request.Directory, "value.go"), "\nconst Changed = 1\n")
			case "added ignored source":
				if err := os.Chmod(request.Directory, 0700); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(request.Directory, "ignored_windows.go"), []byte("package observation\n"), 0600); err != nil {
					t.Fatal(err)
				}
			case "changed ZIP":
				mutateAMD64DownloadedFile(t, prefix+".zip", "different exact archive encoding")
				// HashZip alone intentionally ignores trailing ZIP encoding bytes.
				// Go's actual complete download format must still validate first.
				if err := os.WriteFile(prefix+".zip", []byte("not a ZIP"), 0600); err != nil {
					t.Fatal(err)
				}
			case "changed metadata":
				mutateAMD64DownloadedFile(t, prefix+".mod", "\nrequire example.invalid/unobserved v1.0.0\n")
			}
			if observed, err := captureAMD64DownloadedForTest(ctx, request, path, version); err == nil {
				observed.Close()
				t.Fatal("unobserved or different original source acquired a download observation")
			}
		})
	}
}

func TestAMD64DownloadedObservationRejectsDownloadJSONForgery(t *testing.T) {
	for _, field := range []string{"Path", "Version", "Dir", "Zip", "GoMod", "Sum", "GoModSum", "Error"} {
		t.Run(field, func(t *testing.T) {
			request, path, version := amd64DownloadedObservationFixture(t, "data")
			ctx, cancel := amd64OriginalObservationContext(t)
			defer cancel()
			changed := false
			run := func(ctx context.Context, dir string, env []string, binary string, args ...string) ([]byte, []byte, error) {
				data, stderr, err := RunBounded(ctx, dir, env, binary, args...)
				if err != nil || len(args) < 2 || args[0] != "mod" || args[1] != "download" {
					return data, stderr, err
				}
				var object map[string]json.RawMessage
				if err := json.Unmarshal(data, &object); err != nil {
					return nil, nil, err
				}
				object[field], err = json.Marshal("unobserved")
				if err != nil {
					return nil, nil, err
				}
				changed = true
				data, err = json.Marshal(object)
				return data, stderr, err
			}
			observed, err := CaptureAMD64DownloadedPackageObservation(ctx, AMD64DownloadedObservationRequest{
				Original: request, SourceModule: path, SourceVersion: version,
			}, run)
			if err == nil {
				observed.Close()
				t.Fatal("portable JSON impersonated an actual exact original source")
			}
			if !changed {
				t.Fatal("negative never reached its actual download seam")
			}
		})
	}
}

func TestAMD64DownloadedObservationImmutableSourceAndRevocation(t *testing.T) {
	for _, role := range []string{"source", "unselected source", "foreign source", "ZIP", "metadata", "module directory", "close"} {
		t.Run(role, func(t *testing.T) {
			variant := "data"
			if role == "foreign source" {
				variant = "foreign"
			}
			request, path, version := amd64DownloadedObservationFixture(t, variant)
			ctx, cancel := amd64OriginalObservationContext(t)
			defer cancel()
			observed, err := captureAMD64DownloadedForTest(ctx, request, path, version)
			if err != nil {
				t.Fatal(err)
			}
			defer observed.Close()
			prefix := filepath.Join(originalObservationEnvironment(request.Environment, "GOMODCACHE"), "cache", "download", path, "@v", version)
			name, suffix := filepath.Join(request.Directory, "value.go"), "\nconst Changed = 1\n"
			switch role {
			case "unselected source":
				name, suffix = filepath.Join(request.Directory, "local.h"), "\n#define UNUSED 1\n"
			case "foreign source":
				foreign := observed.downloaded.modules["example.invalid/foreign-observation@"+version]
				if foreign == nil {
					t.Fatal("test lacks its actual separately observed foreign module")
				}
				name = filepath.Join(foreign.root, "shape.go")
			case "ZIP":
				name, suffix = prefix+".zip", "same h1, different retained original encoding"
			case "metadata":
				name = prefix + ".mod"
			case "module directory":
				moved := request.Directory + "-moved"
				if err := os.Chmod(filepath.Dir(request.Directory), 0700); err != nil {
					t.Fatal(err)
				}
				if err := os.Rename(request.Directory, moved); err != nil {
					t.Fatal(err)
				}
				if err := os.Mkdir(request.Directory, 0700); err != nil {
					t.Fatal(err)
				}
				if err := observed.Verify(ctx); err == nil {
					t.Fatal("changed original source directory identity was accepted")
				}
				if err := os.Remove(request.Directory); err != nil {
					t.Fatal(err)
				}
				if err := os.Rename(moved, request.Directory); err != nil {
					t.Fatal(err)
				}
				if err := observed.Verify(ctx); err == nil {
					t.Fatal("restoring the original cache directory resurrected a revoked observation")
				}
				return
			case "close":
				if err := observed.Close(); err != nil {
					t.Fatal(err)
				}
				if _, err := os.Stat(name); err != nil {
					t.Fatal("Close removed caller/Go-owned original module source", err)
				}
				if err := observed.Verify(ctx); err == nil {
					t.Fatal("closed source observation was usable")
				}
				return
			}
			original, err := os.ReadFile(name)
			if err != nil {
				t.Fatal(err)
			}
			mutateAMD64DownloadedFile(t, name, suffix)
			if err := observed.Verify(ctx); err == nil {
				t.Fatal("mutated source/archive was recaptured as a new blessed baseline")
			}
			if err := os.WriteFile(name, original, 0600); err != nil {
				t.Fatal(err)
			}
			if err := observed.Verify(ctx); err == nil {
				t.Fatal("restored bytes resurrected a revoked download observation")
			}
		})
	}
}

func TestAMD64DownloadedObservationCacheRootContinuity(t *testing.T) {
	request, path, version := amd64DownloadedObservationFixture(t, "data")
	ctx, cancel := amd64OriginalObservationContext(t)
	defer cancel()
	observed, err := captureAMD64DownloadedForTest(ctx, request, path, version)
	if err != nil {
		t.Fatal(err)
	}
	defer observed.Close()
	cache := originalObservationEnvironment(request.Environment, "GOMODCACHE")
	moved, replacement := cache+"-original", cache+"-replacement"
	if err := os.Mkdir(replacement, 0700); err != nil {
		t.Fatal(err)
	}
	members, err := os.ReadDir(cache)
	if err != nil {
		t.Fatal(err)
	}
	// Retain every module and metadata subtree inode, byte and path role. Only
	// the actual cache root directory's identity changes at the same pathname.
	for _, member := range members {
		if err := os.Rename(filepath.Join(cache, member.Name()), filepath.Join(replacement, member.Name())); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Rename(cache, moved); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(replacement, cache); err != nil {
		t.Fatal(err)
	}
	// Restore the fixture-owned cache even if the current implementation fails
	// this regression. Never remove a Go/caller-owned module subtree.
	defer func() {
		for _, member := range members {
			if err := os.Rename(filepath.Join(cache, member.Name()), filepath.Join(moved, member.Name())); err != nil {
				t.Error(err)
				return
			}
		}
		if err := os.Remove(cache); err != nil {
			t.Error(err)
			return
		}
		if err := os.Rename(moved, cache); err != nil {
			t.Error(err)
			return
		}
		if err := observed.Verify(ctx); err == nil {
			t.Error("restoring the original cache root resurrected a revoked observation")
		}
	}()
	if err := observed.Verify(ctx); err == nil {
		t.Fatal("replacement cache root preserved all module bytes/inodes but acquired the original cache role")
	}
}

func mutateAMD64DownloadedFile(t *testing.T, name, suffix string) {
	t.Helper()
	if err := os.Chmod(name, 0600); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	if strings.ContainsRune(suffix, '\x00') {
		t.Fatal("fixture source mutation must be explicit readable bytes")
	}
	if err := os.WriteFile(name, append(data, []byte(suffix)...), 0600); err != nil {
		t.Fatal(err)
	}
}
