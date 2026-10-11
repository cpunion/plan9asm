package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/xgo-dev/plan9asm/gotoolprofile"
)

// Resolve the ordinary pure-Go profile before packages.Load. Its later source
// observers receive this exact environment, never defaults reconstructed from
// the process after selection. Explicit nonordinary settings remain visible
// and cannot acquire an ordinary source proof.
func ordinaryPackageEnvironment(goos, goarch string, input []string) ([]string, error) {
	values := make(map[string]string)
	for _, item := range input {
		key, value, ok := strings.Cut(item, "=")
		if !ok || key == "" || strings.ContainsRune(item, '\x00') {
			return nil, fmt.Errorf("malformed package-load environment")
		}
		values[key] = value
	}
	values["GOOS"], values["GOARCH"] = goos, goarch
	for key, value := range map[string]string{
		"GOENV": "off", "GOWORK": "off", "GOTOOLCHAIN": "local",
		"CGO_ENABLED": "0", "GOPACKAGESDRIVER": "off", "GOFLAGS": "",
	} {
		if values[key] == "" {
			values[key] = value
		}
	}
	encode := func() []string {
		result := make([]string, 0, len(values))
		for key, value := range values {
			result = append(result, key+"="+value)
		}
		sort.Strings(result)
		return result
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	stdout, stderr, err := gotoolprofile.RunBounded(ctx, "", encode(), "go",
		"env", "-json", "GOCACHE", "GOMODCACHE", "GOPATH")
	if err != nil || len(stderr) != 0 {
		return nil, fmt.Errorf("resolve package-load cache environment: %v (stderr: %s)", err, stderr)
	}
	var actual map[string]string
	if err := json.Unmarshal(stdout, &actual); err != nil {
		return nil, fmt.Errorf("decode package-load cache environment: %w", err)
	}
	for _, key := range []string{"GOCACHE", "GOMODCACHE", "GOPATH"} {
		if !filepath.IsAbs(actual[key]) {
			return nil, fmt.Errorf("package-load %s must resolve to an absolute path", key)
		}
		values[key] = actual[key]
	}
	for _, key := range []string{"TMPDIR", "GOTMPDIR"} {
		if values[key] == "" {
			values[key] = values["TMPDIR"]
			if values[key] == "" {
				values[key] = os.TempDir()
			}
		}
		path, err := filepath.Abs(values[key])
		if err != nil {
			return nil, err
		}
		values[key], err = filepath.EvalSymlinks(path)
		if err != nil {
			return nil, fmt.Errorf("resolve package-load %s: %w", key, err)
		}
	}
	return encode(), nil
}
