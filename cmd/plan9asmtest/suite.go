package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type config struct {
	Go, Dir, Package, Out string
	Shards                int
	Timeout               time.Duration
	Flags                 []string
	CISelector, CIIsolate string
	CIOwner               int
	CIVerify              string
}

type shardResult struct {
	Index    int               `json:"index"`
	Names    []string          `json:"planned"`
	Command  []string          `json:"command,omitempty"`
	Outcomes map[string]string `json:"outcomes,omitempty"`
	Seconds  float64           `json:"seconds"`
	Status   string            `json:"status"`
	Error    string            `json:"error,omitempty"`
}

type suiteReport struct {
	Schema       int           `json:"schema"`
	Go           string        `json:"go_executable"`
	GoVersion    string        `json:"go_version"`
	Package      string        `json:"package"`
	Flags        []string      `json:"go_flags"`
	Timeout      string        `json:"command_timeout"`
	Names        []string      `json:"inventory"`
	Benchmarks   []string      `json:"benchmarks_not_run_by_default_go_test,omitempty"`
	InventorySHA string        `json:"inventory_sha256"`
	Shards       []shardResult `json:"shards"`
	RootPassed   int           `json:"root_passed"`
	RootSkipped  int           `json:"root_skipped"`
	RootFailed   int           `json:"root_failed"`
	RootMissing  int           `json:"root_incomplete"`
	ChildPassed  int           `json:"subtests_seeds_passed"`
	ChildSkipped int           `json:"subtests_seeds_skipped"`
	ChildFailed  int           `json:"subtests_seeds_failed"`
	ChildMissing int           `json:"subtests_seeds_incomplete"`
	Status       string        `json:"status"`
	Error        string        `json:"error,omitempty"`
}

func runSuite(ctx context.Context, cfg config) error {
	if os.Getenv("GOFLAGS") != "" {
		return fmt.Errorf("GOFLAGS must be empty; pass supported build/instrumentation flags explicitly after --")
	}
	if cfg.Timeout <= 0 || cfg.Package == "" || strings.ContainsAny(cfg.Package, " \t\r\n*?[") || strings.Contains(cfg.Package, "...") {
		return fmt.Errorf("positive timeout and one exact package required")
	}
	forwarded, profile, err := validateFlags(cfg.Flags)
	if err != nil {
		return err
	}
	if profile != "" {
		forwarded = append(forwarded, "-cover")
	}
	if cfg.Go, err = exec.LookPath(cfg.Go); err != nil {
		return err
	}
	cfg.Go, err = filepath.Abs(cfg.Go)
	if err != nil {
		return err
	}
	cfg.Dir, err = filepath.Abs(cfg.Dir)
	if err != nil {
		return err
	}
	cfg.Out, err = filepath.Abs(cfg.Out)
	if err != nil {
		return err
	}
	if profile != "" {
		if !filepath.IsAbs(profile) {
			profile = filepath.Join(cfg.Dir, profile)
		}
		if _, err := os.Lstat(profile); !os.IsNotExist(err) {
			return fmt.Errorf("coverage destination must not exist: %s (%v)", profile, err)
		}
	}
	if err := os.Mkdir(cfg.Out, 0755); err != nil {
		return fmt.Errorf("artifact directory must be new (parent must exist): %w", err)
	}
	report := suiteReport{Schema: 1, Go: cfg.Go, Flags: forwarded, Timeout: cfg.Timeout.String(), Status: "failed"}
	publish := func() error {
		report.RootMissing = len(report.Names) - report.RootPassed - report.RootSkipped - report.RootFailed
		data, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(cfg.Out, "report.json"), append(data, '\n'), 0600)
	}
	fail := func(err error) error {
		report.Error = err.Error()
		return errors.Join(err, publish())
	}
	command := func(stem string, args []string) error {
		_, err := captureCommand(ctx, cfg.Timeout, cfg.Dir, cfg.Go, args, filepath.Join(cfg.Out, stem))
		return err
	}
	if err := command("version", []string{"version"}); err != nil {
		return fail(err)
	}
	version, err := os.ReadFile(filepath.Join(cfg.Out, "version.stdout"))
	if err != nil {
		return fail(err)
	}
	report.GoVersion = strings.TrimSpace(string(version))
	listArgs := append([]string{"test"}, forwarded...)
	listArgs = append(listArgs, "-count=1", "-timeout="+cfg.Timeout.String(), "-list=.", cfg.Package)
	if err := command("inventory", listArgs); err != nil {
		return fail(fmt.Errorf("actual Go inventory failed: %w", err))
	}
	listing, err := os.ReadFile(filepath.Join(cfg.Out, "inventory.stdout"))
	if err != nil {
		return fail(err)
	}
	report.Package, report.Names, report.Benchmarks, err = parseList(string(listing))
	if err != nil {
		return fail(err)
	}
	shards, err := partition(report.Names, cfg.Shards)
	if err != nil {
		return fail(err)
	}
	hash := sha256.Sum256([]byte(strings.Join(report.Names, "\n") + "\n"))
	report.InventorySHA = hex.EncodeToString(hash[:])
	for i, names := range shards {
		result := shardResult{Index: i, Names: names, Status: "pending"}
		if len(names) == 0 {
			result.Status = "empty"
		} else {
			pattern, err := runPattern(names)
			if err != nil {
				return fail(err)
			}
			result.Command = append([]string{cfg.Go, "test"}, forwarded...)
			if profile != "" {
				result.Command = append(result.Command, "-coverprofile="+filepath.Join(cfg.Out, fmt.Sprintf("shard-%04d.cover", i)))
			}
			result.Command = append(result.Command, "-count=1", "-timeout="+cfg.Timeout.String(), "-json", "-run="+pattern, cfg.Package)
		}
		report.Shards = append(report.Shards, result)
	}
	if err := publish(); err != nil {
		return err
	}
	var failures []error
	var profiles []string
	for i := range report.Shards {
		result := &report.Shards[i]
		if result.Status == "empty" {
			continue
		}
		stem := filepath.Join(cfg.Out, fmt.Sprintf("shard-%04d", i))
		result.Seconds, err = captureCommand(ctx, cfg.Timeout, cfg.Dir, cfg.Go, result.Command[1:], stem)
		var validationErr error
		file, openErr := os.Open(stem + ".stdout")
		if openErr == nil {
			result.Outcomes, validationErr = validateEvents(file, report.Package, result.Names)
			validationErr = errors.Join(validationErr, file.Close())
		} else {
			validationErr = openErr
		}
		if err = errors.Join(err, validationErr); err != nil {
			result.Status, result.Error = "failed", err.Error()
			failures = append(failures, fmt.Errorf("shard %d: %w", i, err))
		} else {
			result.Status = "passed"
		}
		for name, state := range result.Outcomes {
			child := strings.Contains(name, "/")
			switch state {
			case "pass":
				if child {
					report.ChildPassed++
				} else {
					report.RootPassed++
				}
			case "skip":
				if child {
					report.ChildSkipped++
				} else {
					report.RootSkipped++
				}
			case "fail":
				if child {
					report.ChildFailed++
				} else {
					report.RootFailed++
				}
			case "run":
				if child {
					report.ChildMissing++
				}
			}
		}
		if profile != "" {
			profiles = append(profiles, stem+".cover")
		}
		fmt.Printf("shard %d/%d: %s (%d roots, %.3fs)\n", i+1, len(shards), result.Status, len(result.Names), result.Seconds)
		if err := publish(); err != nil {
			return err
		}
	}
	if len(failures) != 0 {
		return fail(errors.Join(failures...))
	}
	if profile != "" {
		merged, err := mergeCoverage(profiles)
		if err != nil {
			return fail(err)
		}
		file, err := os.OpenFile(profile, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
		if err != nil {
			return fail(err)
		}
		_, writeErr := file.Write(merged)
		if err := errors.Join(writeErr, file.Close()); err != nil {
			return fail(err)
		}
	}
	report.Status = "passed"
	fmt.Printf("complete root inventory: %d passed, %d skipped, %d failed; subtests/seeds: %d passed, %d skipped, %d failed\n", report.RootPassed, report.RootSkipped, report.RootFailed, report.ChildPassed, report.ChildSkipped, report.ChildFailed)
	return publish()
}

type boundedOutput struct {
	mu        sync.Mutex
	Remaining int64
	Cancel    context.CancelFunc
	Err       error
}

type outputWriter struct {
	File  *os.File
	Bound *boundedOutput
}

func (writer outputWriter) Write(data []byte) (int, error) {
	writer.Bound.mu.Lock()
	defer writer.Bound.mu.Unlock()
	if int64(len(data)) > writer.Bound.Remaining {
		writer.Bound.Err = fmt.Errorf("command output exceeds 64 MiB capture bound")
		writer.Bound.Cancel()
		return 0, writer.Bound.Err
	}
	n, err := writer.File.Write(data)
	writer.Bound.Remaining -= int64(n)
	return n, err
}

func captureCommand(parent context.Context, timeout time.Duration, dir, executable string, args []string, stem string) (float64, error) {
	ctx, cancel := context.WithTimeout(parent, timeout)
	defer cancel()
	out, err := os.OpenFile(stem+".stdout", os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return 0, err
	}
	defer out.Close()
	errOut, err := os.OpenFile(stem+".stderr", os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return 0, err
	}
	defer errOut.Close()
	bound := &boundedOutput{Remaining: 64 << 20, Cancel: cancel}
	cmd := exec.CommandContext(ctx, executable, args...)
	cmd.Dir, cmd.WaitDelay = dir, 2*time.Second
	cmd.Stdout, cmd.Stderr = outputWriter{out, bound}, outputWriter{errOut, bound}
	started := time.Now()
	err = runOwnedCommand(cmd)
	err = errors.Join(err, ctx.Err(), bound.Err, out.Close(), errOut.Close())
	return time.Since(started).Seconds(), err
}

var _ io.Writer = outputWriter{}
