package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// These exceptions are exact module versions, never wildcard packages or
// instruction families. A skip is not a translation pass.
type discoveryInvalidMachineCodeEvidence struct {
	AsmFile          string `json:"asm_file"`
	SHA256           string `json:"sha256"`
	SourceExpression string `json:"source_expression"`
	MacroInvocation  string `json:"macro_invocation,omitempty"`
	Word             string `json:"word"`
	Architecture     string `json:"architecture"`
}

type discoveryInvalidMachineCodeSkip struct {
	Module   string                                `json:"module"`
	Version  string                                `json:"version"`
	Reason   string                                `json:"reason"`
	Evidence []discoveryInvalidMachineCodeEvidence `json:"evidence"`
}

type discoveryInvalidMachineCodeManifest struct {
	SchemaVersion int                               `json:"schema_version"`
	Skips         []discoveryInvalidMachineCodeSkip `json:"skips"`
}

var rawWordPattern = regexp.MustCompile(`^WORD\s+\$(0[xX][0-9a-fA-F]+)\b`)
var macroWordPattern = regexp.MustCompile(`^#define\s+(\w+)\(([^)]*)\)\s+WORD\s+\$\((.*)\)$`)
var macroCallPattern = regexp.MustCompile(`^(\w+)\(([^)]*)\)$`)

func loadInvalidMachineCodeSkips(repoRoot string) (map[string]discoveryInvalidMachineCodeSkip, error) {
	path := filepath.Join(repoRoot, "testdata", "corpus", "invalid-machine-code.json")
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var manifest discoveryInvalidMachineCodeManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("decode invalid machine-code manifest: %w", err)
	}
	if manifest.SchemaVersion != 1 {
		return nil, fmt.Errorf("unsupported invalid machine-code manifest schema %d", manifest.SchemaVersion)
	}
	skips := make(map[string]discoveryInvalidMachineCodeSkip, len(manifest.Skips))
	for _, skip := range manifest.Skips {
		key := skip.Module + "@" + skip.Version
		if skip.Module == "" || skip.Version == "" || skips[key].Module != "" {
			return nil, fmt.Errorf("invalid or duplicate machine-code skip %q", key)
		}
		skips[key] = skip
	}
	return skips, nil
}

func verifyInvalidMachineCodeSkip(moduleDir string, candidate discoveryCandidate, skip discoveryInvalidMachineCodeSkip, rejectWord func(string, uint32) error) error {
	if skip.Module != candidate.Module || skip.Version != candidate.Version {
		return fmt.Errorf("skip identity does not match %s", candidate.exactKey())
	}
	if strings.TrimSpace(skip.Reason) == "" || len(skip.Evidence) == 0 {
		return fmt.Errorf("%s: skip needs a reason and evidence", candidate.exactKey())
	}
	discovered := make(map[string]bool, len(candidate.AsmFiles))
	for _, file := range candidate.AsmFiles {
		discovered[file] = true
	}
	for _, item := range skip.Evidence {
		if !discovered[item.AsmFile] || item.Architecture != "arm64" ||
			!discoverySHA256Pattern.MatchString(item.SHA256) ||
			path.Clean(item.AsmFile) != item.AsmFile || path.IsAbs(item.AsmFile) ||
			strings.HasPrefix(item.AsmFile, "../") || strings.Contains(item.AsmFile, "\\") {
			return fmt.Errorf("%s: invalid or unlisted evidence file %q", candidate.exactKey(), item.AsmFile)
		}
		word, err := strconv.ParseUint(item.Word, 0, 32)
		if err != nil || !strings.HasPrefix(item.Word, "0x") {
			return fmt.Errorf("%s: invalid evidence word %q", candidate.exactKey(), item.Word)
		}
		file := filepath.Join(moduleDir, filepath.FromSlash(item.AsmFile))
		data, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("read skip evidence %s: %w", item.AsmFile, err)
		}
		checksum := sha256.Sum256(data)
		if fmt.Sprintf("%x", checksum) != item.SHA256 ||
			!strings.Contains(string(data), item.SourceExpression) || item.SourceExpression == "" ||
			item.MacroInvocation != "" && !strings.Contains(string(data), item.MacroInvocation) {
			return fmt.Errorf("%s: source evidence changed in %s", candidate.exactKey(), item.AsmFile)
		}
		emitted, err := rawWordFromEvidence(item)
		if err != nil || emitted != uint32(word) {
			return fmt.Errorf("%s: source expression does not emit %s: %v", candidate.exactKey(), item.Word, err)
		}
		if err := rejectWord(item.Architecture, uint32(word)); err != nil {
			return fmt.Errorf("%s: word %s is not independently invalid: %w", candidate.exactKey(), item.Word, err)
		}
	}
	return nil
}

func validateInvalidSourceReportEvidence(result discoveryCorpusResult) error {
	if strings.TrimSpace(result.InvalidSourceReason) == "" || len(result.InvalidSourceEvidence) == 0 ||
		result.Translations != 0 || result.NotApplicableTranslations != 0 || result.Error != "" ||
		len(result.ApplicableAsmFiles) != 0 || len(result.BuildConfigurations) != 0 {
		return fmt.Errorf("invalid-source skip must have evidence and no translation outcome")
	}
	files := make(map[string]bool, len(result.DiscoveredAsmFiles))
	for _, file := range result.DiscoveredAsmFiles {
		files[file] = true
	}
	for _, item := range result.InvalidSourceEvidence {
		if !files[item.AsmFile] || !discoverySHA256Pattern.MatchString(item.SHA256) ||
			item.Architecture != "arm64" || item.SourceExpression == "" {
			return fmt.Errorf("invalid source evidence for %q", item.AsmFile)
		}
		word, err := strconv.ParseUint(item.Word, 0, 32)
		if err != nil {
			return fmt.Errorf("invalid source word %q: %w", item.Word, err)
		}
		emitted, err := rawWordFromEvidence(item)
		if err != nil || emitted != uint32(word) {
			return fmt.Errorf("source expression does not emit %q: %v", item.Word, err)
		}
	}
	return nil
}

func rawWordFromEvidence(item discoveryInvalidMachineCodeEvidence) (uint32, error) {
	if item.MacroInvocation == "" {
		match := rawWordPattern.FindStringSubmatch(strings.TrimSpace(item.SourceExpression))
		if match == nil {
			return 0, fmt.Errorf("not a literal WORD directive")
		}
		word, err := strconv.ParseUint(match[1], 0, 32)
		return uint32(word), err
	}
	definition := macroWordPattern.FindStringSubmatch(strings.TrimSpace(item.SourceExpression))
	call := macroCallPattern.FindStringSubmatch(strings.TrimSpace(item.MacroInvocation))
	if definition == nil || call == nil || definition[1] != call[1] {
		return 0, fmt.Errorf("invalid raw-WORD macro evidence")
	}
	params := strings.Split(definition[2], ",")
	args := strings.Split(call[2], ",")
	if len(params) != len(args) {
		return 0, fmt.Errorf("raw-WORD macro arity mismatch")
	}
	values := make(map[string]uint32, len(params))
	for i, param := range params {
		value, err := strconv.ParseUint(strings.TrimSpace(args[i]), 0, 32)
		if err != nil {
			return 0, err
		}
		values[strings.TrimSpace(param)] = uint32(value)
	}
	expr, err := parser.ParseExpr(definition[3])
	if err != nil {
		return 0, err
	}
	return evaluateRawWordExpression(expr, values)
}

func evaluateRawWordExpression(expr ast.Expr, values map[string]uint32) (uint32, error) {
	switch expr := expr.(type) {
	case *ast.ParenExpr:
		return evaluateRawWordExpression(expr.X, values)
	case *ast.Ident:
		value, ok := values[expr.Name]
		if !ok {
			return 0, fmt.Errorf("unbound macro parameter %q", expr.Name)
		}
		return value, nil
	case *ast.BasicLit:
		value, err := strconv.ParseUint(expr.Value, 0, 32)
		return uint32(value), err
	case *ast.BinaryExpr:
		left, err := evaluateRawWordExpression(expr.X, values)
		if err != nil {
			return 0, err
		}
		right, err := evaluateRawWordExpression(expr.Y, values)
		if err != nil {
			return 0, err
		}
		switch expr.Op {
		case token.OR:
			return left | right, nil
		case token.SHL:
			if right >= 32 {
				return 0, fmt.Errorf("raw-WORD macro shift out of range")
			}
			return left << right, nil
		}
	}
	return 0, fmt.Errorf("unsupported raw-WORD macro expression %T", expr)
}

func rejectInvalidMachineCodeWithLLVM(ctx context.Context, llvmMC, arch string, word uint32) error {
	if arch != "arm64" {
		return fmt.Errorf("unsupported independent decoder architecture %q", arch)
	}
	input := fmt.Sprintf("0x%02x 0x%02x 0x%02x 0x%02x\n", byte(word), byte(word>>8), byte(word>>16), byte(word>>24))
	cmd := exec.CommandContext(ctx, llvmMC, "-disassemble", "-triple=aarch64", "-mattr=+neon,+sm4")
	cmd.Stdin = strings.NewReader(input)
	output, err := cmd.CombinedOutput()
	if err != nil || !bytes.Contains(output, []byte("warning: invalid instruction encoding")) {
		return fmt.Errorf("LLVM 22 did not reject ARM64 word %#08x: %s: %v", word, strings.TrimSpace(string(output)), err)
	}
	return nil
}

func discoveryLLVMDecoder(llc string) (string, error) {
	base := filepath.Base(llc)
	name := ""
	switch base {
	case "llc":
		name = "llvm-mc"
	case "llc-22":
		name = "llvm-mc-22"
	case "llc.exe":
		name = "llvm-mc.exe"
	}
	if name == "" {
		return "", fmt.Errorf("cannot find matching LLVM 22 decoder for %q", llc)
	}
	decoder := filepath.Join(filepath.Dir(llc), name)
	output, err := exec.Command(decoder, "--version").CombinedOutput()
	if err != nil || !discoveryLLVMVersionPattern.Match(output) {
		return "", fmt.Errorf("matching LLVM 22 decoder unavailable at %q: %v: %s", decoder, err, output)
	}
	return decoder, nil
}

func verifyInvalidMachineCodeCandidate(cfg discoveryCorpusConfig, candidate discoveryCandidate, workDir string, skip discoveryInvalidMachineCodeSkip) (runErr error) {
	decoder, err := discoveryLLVMDecoder(cfg.LLC)
	if err != nil {
		return err
	}
	if err := os.Mkdir(workDir, 0755); err != nil {
		return err
	}
	defer func() {
		if err := removeDiscoveryCandidateWorkspace(workDir); err != nil {
			runErr = errors.Join(runErr, fmt.Errorf("clean skip evidence workspace: %w", err))
		}
	}()
	if err := os.WriteFile(filepath.Join(workDir, "go.mod"), []byte("module plan9asm.local/discovery\n\ngo 1.27\n"), 0644); err != nil {
		return err
	}
	return runDiscoveryOperation(cfg.CandidateTimeout, func(ctx context.Context) error {
		env, err := discoveryCandidateEnvironment(ctx, workDir, os.Environ())
		if err != nil {
			return err
		}
		if cfg.buildCache != "" {
			env = replaceEnv(env, map[string]string{"GOCACHE": cfg.buildCache})
		}
		metadata, commandErr := runCapturedCommandOutput(
			ctx, workDir, env, "go", "mod", "download", "-json",
			candidate.exactKey(),
		)
		download, err := resolveModuleDownload(metadata, commandErr, filepath.Join(workDir, "module-source"))
		if err != nil {
			return fmt.Errorf("download skip evidence module: %w", err)
		}
		return verifyInvalidMachineCodeSkip(download.Dir, candidate, skip, func(arch string, word uint32) error {
			decodeCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()
			return rejectInvalidMachineCodeWithLLVM(decodeCtx, decoder, arch, word)
		})
	})
}
