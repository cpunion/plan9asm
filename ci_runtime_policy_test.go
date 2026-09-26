package plan9asm

import (
	"os"
	"strings"
	"testing"
	"time"
)

func TestCIFullSuitesHaveExplicitTimeout(t *testing.T) {
	data, err := os.ReadFile(".github/workflows/go-ci.yml")
	if err != nil {
		t.Fatal(err)
	}
	commands := 0
	for line, command := range strings.Split(string(data), "\n") {
		if !strings.Contains(command, "go test ") || !strings.Contains(command, "./...") {
			continue
		}
		commands++
		var timeout time.Duration
		for _, field := range strings.Fields(command) {
			if strings.HasPrefix(field, "-timeout=") {
				timeout, err = time.ParseDuration(strings.TrimPrefix(field, "-timeout="))
				if err != nil {
					t.Errorf("line %d: invalid timeout: %v", line+1, err)
				}
			}
		}
		if timeout < 20*time.Minute {
			t.Errorf("line %d: full suite requires an explicit timeout of at least 20m: %s", line+1, strings.TrimSpace(command))
		}
	}
	if commands == 0 {
		t.Fatal("no full-suite CI commands found")
	}
}

func TestCICrossRuntimeUsesPinnedQEMU(t *testing.T) {
	data, err := os.ReadFile(".github/workflows/go-ci.yml")
	if err != nil {
		t.Fatal(err)
	}
	_, cross, found := strings.Cut(string(data), "\n  cross-runtime:\n")
	if !found {
		t.Fatal("cross-runtime job not found")
	}
	cross, _, _ = strings.Cut(cross, "\n  test:\n")
	if !strings.Contains(cross, "bash scripts/install-ci-qemu.sh") || strings.Contains(cross, "qemu-user") {
		t.Fatal("cross-runtime must install checksum-pinned QEMU, not Ubuntu 24.04's broken 8.2 dot-product implementation")
	}
	script, err := os.ReadFile("scripts/install-ci-qemu.sh")
	if err != nil {
		t.Fatal(err)
	}
	for _, required := range []string{
		"version=10.2.3",
		"release=deploy/v10.2.3-68",
		"checksum=8e7d8f4c0c7809fc3fea0085199fd6b16f671e7c73d9bf6bec711e1cb535920a",
		"sha256sum --check --status",
		"tools=(qemu-aarch64 qemu-arm qemu-i386)",
		"GITHUB_PATH",
	} {
		if !strings.Contains(string(script), required) {
			t.Errorf("QEMU installer missing %q", required)
		}
	}
}

func TestCIDiscoveredCorpusRetainsAuthenticatedProxyFallback(t *testing.T) {
	data, err := os.ReadFile(".github/workflows/go-ci.yml")
	if err != nil {
		t.Fatal(err)
	}
	_, corpus, found := strings.Cut(string(data), "\n  discovered_library_corpus:\n")
	if !found {
		t.Fatal("discovered corpus job not found")
	}
	corpus, _, _ = strings.Cut(corpus, "\n  discovered_library_corpus_verify:\n")
	if !strings.Contains(corpus, "GOPROXY: https://proxy.golang.org,https://goproxy.cn,direct") {
		t.Fatal("exact-version corpus must try both public module caches before the origin")
	}
	for _, disabled := range []string{"GOSUMDB:", "GONOSUMDB:", "GOPRIVATE:"} {
		if strings.Contains(corpus, disabled) {
			t.Fatalf("public corpus must not bypass checksum-database authentication with %s", disabled)
		}
	}
}
