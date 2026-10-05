//go:build go1.27

package gotoolprofile

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"
)

// This proxy is a first-party mechanics fixture, NOT official SumDB source
// authentication. Production capture must still use authenticated downloads.
func instrumentationForeignFixtureProxy(t *testing.T, root, module, version, declaration string) {
	t.Helper()
	files := map[string]string{
		"go.mod":          "module example.invalid/layoutdep\n\ngo 1.20\n",
		"model/layout.go": declaration,
	}
	var buffer bytes.Buffer
	writer := zip.NewWriter(&buffer)
	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		entry, err := writer.Create(module + "@" + version + "/" + name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := io.WriteString(entry, files[name]); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	directory := filepath.Join(root, filepath.FromSlash(module), "@v")
	for suffix, contents := range map[string][]byte{
		".mod":  []byte(files["go.mod"]),
		".info": []byte(fmt.Sprintf("{\"Version\":%q,\"Time\":\"2026-01-01T00:00:00Z\"}\n", version)),
		".zip":  buffer.Bytes(),
	} {
		writeInstrumentationFixture(t, directory, version+suffix, contents)
	}
}

func TestInstrumentationForeignTypeSourceActualDualModule(t *testing.T) {
	binary, err := exec.LookPath("go")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	for _, replaced := range []bool{false, true} {
		t.Run(fmt.Sprintf("version_replacement_%t", replaced), func(t *testing.T) {
			proxy := instrumentationMarkerDirectory(t)
			cache := instrumentationMarkerDirectory(t)
			t.Cleanup(func() {
				// Go intentionally extracts this owned fixture read-only. Restore
				// only its directory modes before testing removes the private tree.
				if err := filepath.Walk(cache, func(name string, info os.FileInfo, err error) error {
					if err == nil && info.IsDir() {
						return os.Chmod(name, 0700)
					}
					return err
				}); err != nil {
					t.Error(err)
				}
			})
			archive := "example.invalid/layoutdep"
			mod := "module " + instrumentationExportModule + "\n\ngo 1.20\nrequire example.invalid/layoutdep v1.0.0\n"
			if replaced {
				archive = "example.invalid/layout-archive"
				mod += "replace example.invalid/layoutdep v1.0.0 => " + archive + " v1.0.0\n"
			}
			instrumentationForeignFixtureProxy(t, proxy, archive, "v1.0.0",
				"package model\nconst Words=2\ntype Header struct { Lo uint32; Pad [Words]byte; Hi uint16 }\n")
			env := replaceEnv(os.Environ(), map[string]string{
				"GOPROXY": "file://" + proxy, "GOSUMDB": "off", "GONOSUMDB": "", "GOPRIVATE": "",
				"GOMODCACHE": cache, "GOPATH": instrumentationMarkerDirectory(t),
			})
			marker, err := CaptureInstrumentationExportObservation(ctx, binary, instrumentationMarkerDirectory(t), env,
				"darwin/arm64", InstrumentationRace, "0", nil, RunBounded)
			if err != nil {
				t.Fatal(err)
			}
			original := instrumentationOriginalFixture(t, marker.Early)
			for name, source := range map[string]string{
				"go.mod":      mod,
				"base.go":     "package markers\nimport shape \"example.invalid/layoutdep/model\"\nfunc marker(value shape.Header) uint64\n",
				"mode_race.s": "//go:build race\n\n#include \"textflag.h\"\nTEXT ·marker(SB),NOSPLIT,$0-16\nMOVD $0,R0\nMOVD R0,ret+8(FP)\nRET\n",
			} {
				writeInstrumentationFixture(t, original.SourceRoot, name, []byte(source))
				original.Sources[name], original.Headers[name] = bytesSHA256([]byte(source)), source
			}
			env = replaceEnv(env, marker.Early.Environment)
			download, downloadStderr, downloadErr := RunBounded(ctx, original.SourceRoot, env, binary,
				"mod", "download", "-json", archive+"@v1.0.0")
			if downloadErr != nil {
				t.Fatalf("first-party fixture download: %v; %s", downloadErr, downloadStderr)
			}
			t.Logf("actual first-party download (not official authentication): %s; stderr=%q", download, downloadStderr)
			// Keep raw stock-Go records before capture fails. These are actual
			// Module/Replace/ImportMap records, not reconstructed expected roots.
			args := []string{"list", "-mod=mod", "-deps", "-compiled", "-export", "-json", "-race", original.Module}
			stdout, stderr, commandErr := RunBounded(ctx, original.SourceRoot, env, binary, args...)
			if commandErr != nil || len(stderr) != 0 {
				t.Fatalf("actual dual-module graph setup: %v; %s", commandErr, stderr)
			}
			// Freeze actual invocation metadata only after stock Go has written
			// its first-party checksums. Capture may not mutate the inventory.
			sum, err := os.ReadFile(filepath.Join(original.SourceRoot, "go.sum"))
			if err != nil {
				t.Fatal(err)
			}
			original.Sources["go.sum"], original.Headers["go.sum"] = bytesSHA256(sum), string(sum)
			original.Directories["."] = append(original.Directories["."], "go.sum")
			sort.Strings(original.Directories["."])
			inventory, err := CaptureInstrumentationOriginalSourceInventory(original, ".")
			if err != nil {
				t.Fatal(err)
			}
			plan, err := PlanInstrumentationOriginalPackage(original, inventory, marker.Early, ".", nil)
			if err != nil {
				t.Fatal(err)
			}
			input, err := NewInstrumentationConsumerInput(original, inventory, marker, plan)
			if err != nil {
				t.Fatal(err)
			}
			decoder := json.NewDecoder(bytes.NewReader(stdout))
			found := false
			for {
				var pkg instrumentationTypeListPackage
				if err := decoder.Decode(&pkg); err == io.EOF {
					break
				} else if err != nil {
					t.Fatal(err)
				}
				if pkg.ImportPath == plan.PackagePath {
					record, err := json.Marshal(pkg)
					if err != nil {
						t.Fatal(err)
					}
					t.Logf("actual root package: %s", record)
				}
				if pkg.ImportPath == "example.invalid/layoutdep/model" {
					found = true
					if pkg.Module == nil || pkg.Module.Path != "example.invalid/layoutdep" || pkg.Module.Version != "v1.0.0" ||
						(pkg.Module.Replace != nil) != replaced || pkg.Export == "" {
						t.Fatalf("actual dependency identity lost: %+v", pkg)
					}
					record, err := json.Marshal(pkg)
					if err != nil {
						t.Fatal(err)
					}
					t.Logf("actual foreign package: %s", record)
				}
			}
			if !found {
				t.Fatal("stock Go omitted the foreign package")
			}
			t.Logf("actual tool=%s sha256=%s graph_sha256=%s argv=%q; first-party file proxy is not a SumDB certificate",
				binary, marker.Early.BaselineFeatures.DriverSHA256, bytesSHA256(stdout), args)
			proof, err := CaptureInstrumentationTypeSources(ctx, binary, original.SourceRoot, env, input, RunBounded)
			if err != nil {
				if !strings.Contains(err.Error(), "authenticated foreign module source closure") {
					t.Fatalf("unexpected actual graph failure: %v", err)
				}
				t.Fatalf("actual dual-module source role remains unavailable: %v", err)
			}
			if proof == nil {
				t.Fatal("capture returned no independent source proof")
			}
		})
	}
}
