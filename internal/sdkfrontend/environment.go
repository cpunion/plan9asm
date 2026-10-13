package sdkfrontend

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/xgo-dev/plan9asm/gotoolprofile"
	"github.com/xgo-dev/plan9asm/internal/sourcepath"
)

// Resolve the ordinary load environment before packages.Load. Its later source
// observers receive this exact environment, never defaults reconstructed from
// the process after selection. Explicit nonordinary settings remain visible
// and cannot acquire an ordinary source proof.
func OrdinaryEnvironment(goos, goarch string, input []string) ([]string, error) {
	values := make(map[string]string)
	for _, item := range input {
		key, value, ok := strings.Cut(item, "=")
		if !ok || key == "" || strings.ContainsRune(item, '\x00') {
			return nil, fmt.Errorf("malformed package-load environment")
		}
		values[key] = value
	}
	values["GOOS"], values["GOARCH"] = goos, goarch
	if values["GOTOOLCHAIN"] == "" {
		values["GOTOOLCHAIN"] = "local"
	}
	if values["GOPACKAGESDRIVER"] == "" {
		// Match packages.Load's automatic discovery, including its use of
		// the process PATH. A discovered driver is not an ordinary Go load.
		driver, err := exec.LookPath("gopackagesdriver")
		if err != nil {
			driver = "off"
		}
		values["GOPACKAGESDRIVER"] = driver
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
		"env", "-json", "GOCACHE", "GOMODCACHE", "GOPATH", "CGO_ENABLED", "GOENV", "GOWORK", "GOFLAGS")
	if err != nil || len(stderr) != 0 {
		return nil, fmt.Errorf("resolve package-load cache environment: %v (stderr: %s)", err, stderr)
	}
	var actual map[string]string
	if err := json.Unmarshal(stdout, &actual); err != nil {
		return nil, fmt.Errorf("decode package-load cache environment: %w", err)
	}
	// Stock Go selects its cgo default for the requested target. Forcing zero
	// here would silently drop selected assembly (including runtime/cgo).
	// Pure-Go source proofs still require an explicitly compatible profile.
	if actual["CGO_ENABLED"] != "0" && actual["CGO_ENABLED"] != "1" {
		return nil, fmt.Errorf("invalid effective package-load CGO_ENABLED")
	}
	values["CGO_ENABLED"] = actual["CGO_ENABLED"]
	values["GOFLAGS"] = actual["GOFLAGS"]
	goenv, err := GoEnvFile(values["GOENV"], actual["GOENV"])
	if err != nil {
		return nil, fmt.Errorf("resolve package-load GOENV: %w", err)
	}
	actual["GOENV"] = goenv
	for _, key := range []string{"GOENV", "GOWORK"} {
		values[key] = actual[key]
		if values[key] == "" {
			values[key] = "off"
		}
	}
	for _, key := range []string{"GOCACHE", "GOMODCACHE", "GOPATH"} {
		if !filepath.IsAbs(actual[key]) {
			return nil, fmt.Errorf("package-load %s must resolve to an absolute path", key)
		}
		// A local module may never download a dependency, so cmd/go need not
		// create GOPATH or GOMODCACHE. Prepare the selected paths before the
		// load captures their identity; do not repair a later source proof.
		if err := os.MkdirAll(actual[key], 0755); err != nil {
			return nil, fmt.Errorf("prepare package-load %s: %w", key, err)
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
		values[key], err = sourcepath.Resolve(path)
		if err != nil {
			return nil, fmt.Errorf("resolve package-load %s: %w", key, err)
		}
	}
	return encode(), nil
}

func GoEnvFile(requested, actual string) (string, error) {
	if requested != "" || actual == "" || actual == "off" {
		return actual, nil
	}
	// Resolve the absence before packages.Load. An implicit missing config
	// carries Go's defaults; explicitly fixing that choice prevents a later
	// file from changing the compiler invocation observed by source guards.
	// User-selected paths and existing settings keep their original role.
	_, err := os.Lstat(actual)
	if os.IsNotExist(err) {
		return "off", nil
	}
	if err != nil {
		return "", err
	}
	return actual, nil
}
