package plan9asm

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

func goAssemblerOracleOutput(t *testing.T, arch, source, experiment string, wantSuccess bool) {
	t.Helper()
	dir := t.TempDir()
	asm, object := filepath.Join(dir, "forms.s"), filepath.Join(dir, "forms.o")
	if err := os.WriteFile(asm, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	stage, cancelStage := runtimeOracleStageContext(t)
	defer cancelStage()
	ctx, cancel := context.WithTimeout(stage, time.Minute)
	defer cancel()
	command := exec.CommandContext(ctx, "go", "tool", "asm", "-p", "example.com/forms", "-o", object, asm)
	command.Env = append(os.Environ(), "GOTOOLCHAIN=local", "GOOS=linux", "GOARCH="+arch)
	if experiment != "" {
		command.Env = append(command.Env, "GOEXPERIMENT="+experiment)
	}
	output, err := runtimeOracleCombinedOutput(ctx, command)
	if failure := goAssemblerOracleOutcome(asm, object, output, err, ctx.Err(), wantSuccess, arch); failure != nil {
		t.Fatalf("%v\n%s\noriginal source:\n%s", failure, output, source)
	}
}

func goAssemblerOracleOutcome(source, object string, output []byte, commandErr, contextErr error, wantSuccess bool, arch string) error {
	if contextErr != nil {
		return fmt.Errorf("Go assembler oracle infrastructure failure: deadline/cancellation: %w", contextErr)
	}
	if commandErr == nil {
		if !wantSuccess {
			return fmt.Errorf("Go assembler accepted a form expected outside the optab")
		}
		info, err := os.Lstat(object)
		if err != nil || !info.Mode().IsRegular() || info.Size() == 0 {
			return fmt.Errorf("Go assembler oracle infrastructure failure: missing/nonregular/empty object: %v", err)
		}
		return nil
	}
	// Cleanup/pipe errors joined to a source exit are still infrastructure
	// failures. A signal, launch error or unknown positive exit is not evidence.
	exit, ok := commandErr.(*exec.ExitError)
	if !ok || exit.ProcessState == nil || !exit.ProcessState.Exited() || exit.ExitCode() != 1 {
		return fmt.Errorf("Go assembler oracle infrastructure failure: unknown process outcome: %w", commandErr)
	}
	text := string(output)
	for _, marker := range []string{"panic:", "fatal error:", "runtime: failed to", "runtime: out of memory", "signal:", "executable file not found"} {
		if strings.Contains(text, marker) {
			return fmt.Errorf("Go assembler oracle infrastructure failure: %s: %w", marker, commandErr)
		}
	}
	// Go prints source positions either directly or in its illegal-combination
	// listing. A footer alone, or a line for another source, cannot reject this
	// exact fixture. Keep positive canonical line/column positions.
	if !goAssemblerOracleSourceDiagnostic(source, text) && !goAssemblerOraclePositionlessDiagnostic(source, arch, text) {
		return fmt.Errorf("Go assembler oracle infrastructure failure: no concrete fixture diagnostic: %w", commandErr)
	}
	if wantSuccess {
		return fmt.Errorf("Go assembler rejected accepted forms: %w", commandErr)
	}
	return nil
}

func goAssemblerOracleSourceDiagnostic(source, output string) bool {
	positivePosition := func(text string) bool {
		value, err := strconv.Atoi(text)
		return err == nil && value > 0 && strconv.Itoa(value) == text
	}
	backendDiagnostic := false
	for _, line := range strings.Split(output, "\n") {
		start := strings.Index(line, source+":")
		if start < 0 {
			if strings.TrimSpace(line) != "" {
				backendDiagnostic = strings.HasPrefix(line, "asm: ") && line != "asm: assembly failed"
			}
			continue
		}
		position := line[start+len(source)+1:]
		end := strings.IndexAny(position, ":)")
		if end < 0 || !positivePosition(position[:end]) {
			continue
		}
		message := strings.TrimSpace(position[end+1:])
		if position[end] == ')' {
			// An illegal-combination listing retains its parenthesized source
			// position and the rejected instruction, not just a tool footer.
			if start > 0 && line[start-1] == '(' && strings.HasPrefix(line, "asm:") && message != "" {
				return true
			}
			if backendDiagnostic && start > 0 && line[start-1] == '(' &&
				goAssemblerOracleListingBindsSource(source, position[:end], line[:start-1], message) {
				return true
			}
			continue
		}
		if column, detail, found := strings.Cut(message, ":"); found &&
			(column == "" || strings.ContainsRune("+-0123456789", rune(column[0]))) {
			// Do not backtrack an explicit invalid column into a line-only
			// diagnostic. Missing columns are different from zero/overflow.
			if positivePosition(column) && strings.TrimSpace(detail) != "" {
				return true
			}
			continue
		}
		if message != "" {
			return true
		}
	}
	return false
}

func TestGoAssemblerOracleActualCommonForms(t *testing.T) {
	for _, arch := range []string{"arm64", "amd64", "386"} {
		t.Run(arch, func(t *testing.T) {
			// These common forms deliberately invoke every actual compatibility
			// assembler, unlike a current-table oracle on an older toolchain.
			goAssemblerOracleOutput(t, arch, "TEXT ·valid(SB),$0-0\n\tRET\n", "", true)
			goAssemblerOracleOutput(t, arch, "TEXT ·invalid(SB),$0-0\n\tNOT_AN_INSTRUCTION\n\tRET\n", "", false)
			invalid := "MOVD (R0),RSP"
			if arch == "amd64" {
				invalid = "MOVQ (AX),(BX)"
			} else if arch == "386" {
				invalid = "MOVL (AX),(BX)"
			}
			goAssemblerOracleOutput(t, arch, "TEXT ·invalid(SB),$0-0\n\t"+invalid+"\n\tRET\n", "", false)
		})
	}
}

func TestGoAssemblerOracleOutcomeRejectsUnboundRejections(t *testing.T) {
	dir := t.TempDir()
	source, object := filepath.Join(dir, "forms.s"), filepath.Join(dir, "forms.o")
	if err := os.WriteFile(source, []byte("TEXT ·invalid(SB),$0-0\n\tNOT_AN_INSTRUCTION\n"), 0600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := runtimeOracleStageContext(t)
	defer cancel()
	command := exec.CommandContext(ctx, "go", "tool", "asm", "-o", object, source)
	command.Env = append(os.Environ(), "GOTOOLCHAIN=local", "GOOS=linux", "GOARCH=arm64")
	output, err := runtimeOracleCombinedOutput(ctx, command)
	if failure := goAssemblerOracleOutcome(source, object, output, err, ctx.Err(), false, "arm64"); failure != nil {
		t.Fatalf("actual source rejection lost: %v\n%s", failure, output)
	}
	for _, test := range []struct {
		name    string
		output  string
		err     error
		context error
	}{
		{name: "no diagnostic", err: err},
		{name: "footer only", output: "asm: assembly failed\n", err: err},
		{name: "foreign source", output: "other.s:2: bad operand\n", err: err},
		{name: "zero line", output: source + ":0: bad operand\n", err: err},
		{name: "negative line", output: source + ":-2: bad operand\n", err: err},
		{name: "overflow line", output: source + ":999999999999999999999999999999: bad operand\n", err: err},
		{name: "zero column", output: source + ":2:0: bad operand\n", err: err},
		{name: "negative column", output: source + ":2:-1: bad operand\n", err: err},
		{name: "empty column", output: source + ":2:: bad operand\n", err: err},
		{name: "empty message", output: source + ":2: \n", err: err},
		{name: "joined cleanup", output: string(output), err: errors.Join(err, errors.New("owned cleanup failure"))},
		{name: "canceled", output: string(output), err: err, context: context.Canceled},
		{name: "deadline", output: string(output), err: err, context: context.DeadlineExceeded},
		{name: "late infrastructure", output: string(output) + "fatal error: out of memory\n", err: err},
		{name: "panic with source diagnostic", output: string(output) + "panic: internal assembler failure\n", err: err},
	} {
		t.Run(test.name, func(t *testing.T) {
			failure := goAssemblerOracleOutcome(source, object, []byte(test.output), test.err, test.context, false, "arm64")
			if failure == nil || !strings.Contains(failure.Error(), "infrastructure failure") {
				t.Fatalf("unbound rejection accepted: %v", failure)
			}
		})
	}
}
