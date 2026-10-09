// Package goasmoracle validates source-bound observations, not ISA/ABI authority.
package goasmoracle

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"text/scanner"
)

func CheckOutcome(source, object string, stdout, stderr []byte, commandErr, contextErr error, wantSuccess bool, focus int, additional func(string, string) bool) error {
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
	streams := []string{string(stdout), string(stderr)}
	for _, marker := range []string{"panic:", "fatal error:", "runtime: failed to", "runtime: out of memory", "signal:", "executable file not found"} {
		for _, text := range streams {
			if strings.Contains(text, marker) {
				return fmt.Errorf("Go assembler oracle infrastructure failure: %s: %w", marker, commandErr)
			}
		}
	}
	// Go prints source positions either directly or in its illegal-combination
	// listing. A footer alone, or a line for another source, cannot reject this
	// exact fixture. Keep positive canonical line/column positions.
	bound := false
	for _, text := range streams {
		bound = bound || SourceDiagnosticLine(source, text, focus) || (additional != nil && additional(source, text))
	}
	if !bound {
		return fmt.Errorf("Go assembler oracle infrastructure failure: no concrete fixture diagnostic: %w", commandErr)
	}
	if wantSuccess {
		return fmt.Errorf("Go assembler rejected accepted forms: %w", commandErr)
	}
	return nil
}

// ReadSource preserves the existing eight-MiB fixture bound. This is an
// independent bounded physical read, not a caller-supplied snapshot/hash.
func ReadSource(source string) ([]byte, bool) {
	before, err := os.Lstat(source)
	if err != nil || !before.Mode().IsRegular() || before.Size() < 0 || before.Size() > 8<<20 {
		return nil, false
	}
	file, err := os.Open(source)
	if err != nil {
		return nil, false
	}
	opened, err := file.Stat()
	if err != nil || !sameInput(before, opened) {
		_ = file.Close()
		return nil, false
	}
	data, readErr := io.ReadAll(io.LimitReader(file, before.Size()+1))
	ended, endErr := file.Stat()
	closeErr := file.Close()
	after, pathErr := os.Lstat(source)
	return data, readErr == nil && endErr == nil && closeErr == nil &&
		pathErr == nil && int64(len(data)) == before.Size() &&
		sameInput(before, ended) && sameInput(before, after)
}

func sameInput(before, after os.FileInfo) bool {
	return os.SameFile(before, after) && before.Mode() == after.Mode() &&
		before.Size() == after.Size() && before.ModTime() == after.ModTime()
}

// The token role is only a diagnostic join. It is not an instruction grammar,
// lowering rule, or permission to execute a source statement.
func statementOpcode(statement string) string {
	var tokens scanner.Scanner
	tokens.Init(strings.NewReader(statement))
	tokens.Mode = scanner.ScanIdents | scanner.ScanComments | scanner.SkipComments
	tokens.Error = func(_ *scanner.Scanner, _ string) {}
	if tokens.Scan() != scanner.Ident {
		return ""
	}
	opcode := tokens.TokenText()
	for tokens.Scan() == '.' {
		if tokens.Scan() != scanner.Ident {
			return ""
		}
		opcode += "." + tokens.TokenText()
	}
	// Consume the complete line. A malformed comment later in an operand is
	// still a scanner failure, never an emitted stderr diagnostic or a match.
	for tokens.Scan() != scanner.EOF {
	}
	for _, digit := range opcode {
		if digit != '.' && (digit < 'A' || digit > 'Z') && (digit < '0' || digit > '9') {
			return ""
		}
	}
	if tokens.ErrorCount != 0 {
		return ""
	}
	return opcode
}

func positivePosition(text string) (int, bool) {
	value, err := strconv.Atoi(text)
	return value, err == nil && value > 0 && strconv.Itoa(value) == text
}

func canonicalPC(pc string) bool {
	address, err := strconv.ParseUint(pc, 10, 64)
	return err == nil && len(pc) >= 5 && fmt.Sprintf("%05d", address) == pc
}

func ListingBindsSource(source, position, pc, instruction string) bool {
	data, ok := ReadSource(source)
	line, valid := positivePosition(position)
	lines := strings.Split(string(data), "\n")
	if !ok || !valid || line > len(lines) || !canonicalPC(strings.TrimSpace(pc)) {
		return false
	}
	fields := strings.Fields(instruction)
	opcode := statementOpcode(lines[line-1])
	return len(fields) > 1 && opcode != "" && strings.EqualFold(opcode, fields[0])
}

func SourceDiagnostic(source, output string) bool {
	return SourceDiagnosticLine(source, output, 0)
}

// Only a real statement in this exact physical fixture can bind a diagnostic.
// A nonzero focus is the script's actual isolated instruction, not its CPP
// header, TEXT setup, RET footer, another source, or an unobserved #line route.
func SourceDiagnosticLine(source, output string, focus int) bool {
	data, ok := ReadSource(source)
	if !ok || focus < 0 {
		return false
	}
	lines := strings.Split(string(data), "\n")
	backendDiagnostic := false
	for _, line := range strings.Split(output, "\n") {
		direct := strings.HasPrefix(line, source+":")
		start, pc := 0, ""
		if !direct {
			marker := strings.Index(line, "("+source+":")
			if marker < 0 {
				if strings.TrimSpace(line) != "" {
					backendDiagnostic = strings.HasPrefix(line, "asm: ") &&
						line != "asm: assembly failed" && line != "asm: "
				}
				continue
			}
			prefix := strings.TrimSpace(line[:marker])
			fields := strings.Fields(prefix)
			if len(fields) == 0 {
				continue
			}
			pc = fields[len(fields)-1]
			if !canonicalPC(pc) {
				continue
			}
			if len(fields) == 1 {
				if !backendDiagnostic {
					continue
				}
			} else if !strings.HasPrefix(prefix, "asm: ") ||
				strings.TrimSpace(strings.TrimSuffix(prefix, pc)) == "asm:" ||
				strings.TrimSpace(strings.TrimSuffix(prefix, pc)) == "asm: assembly failed" {
				continue
			}
			start = marker + 1
		}
		position := line[start+len(source)+1:]
		end := strings.IndexAny(position, ":)")
		if end < 0 {
			continue
		}
		number, valid := positivePosition(position[:end])
		if !valid || number > len(lines) || (focus > 0 && number != focus) ||
			statementOpcode(lines[number-1]) == "" {
			continue
		}
		message := strings.TrimSpace(position[end+1:])
		if position[end] == ')' {
			if !direct && ListingBindsSource(source, position[:end], pc, message) {
				return true
			}
			continue
		}
		if !direct {
			continue
		}
		if column, detail, found := strings.Cut(message, ":"); found &&
			(column == "" || strings.ContainsRune("+-0123456789", rune(column[0]))) {
			if value, valid := positivePosition(column); valid && value <= len([]rune(lines[number-1])) && strings.TrimSpace(detail) != "" {
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
