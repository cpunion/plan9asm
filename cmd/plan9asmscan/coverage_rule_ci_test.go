package main

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Run the actual CI checker, not a duplicate fingerprint comparison. These
// synthetic reports prove drift detection only, never instruction coverage.
func TestOfficialCoverageCIChecksCompleteEncodingRuleAttributes(t *testing.T) {
	script, err := os.ReadFile("../../scripts/check-go-asm-coverage.sh")
	if err != nil {
		t.Fatal(err)
	}
	_, python, ok := strings.Cut(string(script), "<<'PY'\n")
	if !ok {
		t.Fatal("official coverage checker is not bound to its script")
	}
	python, _, ok = strings.Cut(python, "\nPY\n")
	if !ok {
		t.Fatal("official coverage checker has no complete heredoc")
	}
	tool, err := exec.LookPath("python3")
	if err != nil {
		tool, err = exec.LookPath("python")
	}
	if err != nil {
		t.Fatal("Python 3 is required by the actual official CI checker", err)
	}
	original, err := os.ReadFile("../../testdata/coverage/go-asm-baseline.json")
	if err != nil {
		t.Fatal(err)
	}
	for _, mutation := range []string{
		"unchanged", "changed-fingerprint", "changed-count", "missing-report",
		"missing-baseline", "both-missing",
	} {
		t.Run(mutation, func(t *testing.T) {
			var baseline map[string]interface{}
			if err := json.Unmarshal(original, &baseline); err != nil {
				t.Fatal(err)
			}
			versions := baseline["versions"].(map[string]interface{})
			// Every minor and architecture remains present in the real checker.
			// Add isolated synthetic attributes without changing class buckets.
			for _, release := range versions {
				arches := release.(map[string]interface{})
				for _, arch := range []string{"386", "amd64"} {
					entry := arches[arch].(map[string]interface{})
					entry["encoder_rules"] = float64(3)
					entry["encoder_rule_fingerprint"] = strings.Repeat("a", 64)
				}
			}
			arches := versions["go1.27"].(map[string]interface{})
			reports := make(map[string]map[string]interface{})
			for arch, value := range arches {
				bytes, err := json.Marshal(value)
				if err != nil {
					t.Fatal(err)
				}
				var report map[string]interface{}
				if err := json.Unmarshal(bytes, &report); err != nil {
					t.Fatal(err)
				}
				report["go_version"] = "go1.27.1"
				report["goarch"] = arch
				reports[arch] = report
			}
			families, err := os.ReadFile("../../testdata/coverage/arm64-go-assembler-families.txt")
			if err != nil {
				t.Fatal(err)
			}
			var catalog []map[string]interface{}
			for _, line := range strings.Split(string(families), "\n") {
				line = strings.TrimSpace(line)
				if line != "" && !strings.HasPrefix(line, "#") {
					catalog = append(catalog, map[string]interface{}{
						"opcode": line, "encoder_forms": 1, "observed_in_corpus": true,
					})
				}
			}
			reports["arm64"]["opcode_catalog"] = catalog
			entry := arches["amd64"].(map[string]interface{})
			switch mutation {
			case "changed-fingerprint":
				reports["amd64"]["encoder_rule_fingerprint"] = strings.Repeat("b", 64)
			case "changed-count":
				reports["amd64"]["encoder_rules"] = float64(4)
			case "missing-report", "both-missing":
				delete(reports["amd64"], "encoder_rules")
				delete(reports["amd64"], "encoder_rule_fingerprint")
			}
			if mutation == "missing-baseline" || mutation == "both-missing" {
				delete(entry, "encoder_rules")
				delete(entry, "encoder_rule_fingerprint")
			}
			dir := t.TempDir()
			write := func(name string, value interface{}) {
				bytes, err := json.Marshal(value)
				if err == nil {
					err = os.WriteFile(filepath.Join(dir, name), bytes, 0600)
				}
				if err != nil {
					t.Fatal(err)
				}
			}
			write("baseline.json", baseline)
			reportDir := filepath.Join(dir, "reports")
			if err := os.Mkdir(reportDir, 0700); err != nil {
				t.Fatal(err)
			}
			for arch, report := range reports {
				write(filepath.Join("reports", arch+".json"), report)
			}
			familyPath, err := filepath.Abs("../../testdata/coverage/arm64-go-assembler-families.txt")
			if err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer cancel()
			command := exec.CommandContext(ctx, tool, "-", filepath.Join(dir, "baseline.json"), familyPath, reportDir)
			command.Stdin = strings.NewReader(python)
			output, err := command.CombinedOutput()
			if mutation == "unchanged" {
				if err != nil {
					t.Fatalf("unchanged complete synthetic inventory rejected: %v\n%s", err, output)
				}
			} else if err == nil {
				t.Fatalf("CI accepted %s with identical historical operand-class fingerprint:\n%s", mutation, output)
			} else if ctx.Err() != nil || !strings.Contains(string(output), "encoder_rule") {
				t.Fatalf("CI failure was not an encoding-rule diagnosis: %v\n%s", err, output)
			}
		})
	}
}
