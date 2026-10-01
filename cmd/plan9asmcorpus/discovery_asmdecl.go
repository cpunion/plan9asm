package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"text/scanner"

	"github.com/xgo-dev/plan9asm/internal/goabi"
)

var discoveryAsmDeclZeroArgs = regexp.MustCompile(
	`^(.+\.s):([0-9]+)(?::[0-9]+)?: \[([^]]+)\] ([^:]+): wrong argument size 0; expected \$\.\.\.-[1-9][0-9]*$`,
)

var discoveryAsmDeclPaddedArgs = regexp.MustCompile(
	`^(.+\.s):([0-9]+)(?::[0-9]+)?: \[([^]]+)\] ([^:]+): wrong argument size ([1-9][0-9]*); expected \$\.\.\.-([0-9]+)$`,
)

var errDiscoveryAsmDeclSourceProof = errors.New("cannot verify asmdecl source metadata")

// Only literal TEXT declarations are conclusive here. A macro/expression,
// different symbol, stale line number or unselected source must not turn an
// actual ABI failure into an unspecified-argument warning.
var discoveryAsmDeclLiteralText = regexp.MustCompile(
	`^\s*TEXT\s+[^,\s]*·([^\s(]+)\(SB\)\s*,\s*(?:[0-9A-Z|+()]+\s*,\s*)?\$(-?[0-9]+)(?:-([0-9]+))?\s*$`,
)

func filterDiscoveryAsmDeclTextMetadata(
	ctx context.Context,
	dir string,
	env, buildTags, patterns []string,
	modfilePath string,
	err error,
) error {
	if err == nil || isDiscoveryInfrastructureFailure(err) {
		return err
	}
	var failure *discoveryCapturedCommandError
	if !errors.As(err, &failure) {
		return err
	}
	lines := strings.Split(failure.output, "\n")
	hasMetadataWarning := false
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if discoveryAsmDeclZeroArgs.MatchString(line) || discoveryAsmDeclPaddedArgs.MatchString(line) {
			hasMetadataWarning = true
			break
		}
	}
	if !hasMetadataWarning {
		return err
	}

	args := []string{"list", "-json"}
	if modfilePath != "" {
		args = append(args, "-modfile="+modfilePath)
	}
	if len(buildTags) != 0 {
		args = append(args, "-tags="+strings.Join(buildTags, ","))
	}
	args = append(args, patterns...)
	output, listErr := runCapturedCommandOutput(ctx, dir, env, "go", args...)
	if listErr != nil {
		return errors.Join(err, fmt.Errorf("%w: resolve selected sources: %w", errDiscoveryAsmDeclSourceProof, listErr))
	}
	decoder := json.NewDecoder(bytes.NewReader(output))
	var packages []discoveryGoListPackage
	for {
		var pkg discoveryGoListPackage
		if decodeErr := decoder.Decode(&pkg); decodeErr != nil {
			if decodeErr == io.EOF {
				break
			}
			return errors.Join(err, fmt.Errorf("%w: decode selected sources: %w", errDiscoveryAsmDeclSourceProof, decodeErr))
		}
		packages = append(packages, pkg)
	}
	filtered := filterDiscoveryAsmDeclTextMetadataLines(lines, packages)
	if strings.Join(filtered, "\n") == failure.output {
		return err
	}
	// Only package headings may remain when the ignored warning was the whole
	// failure. Other diagnostics retain the original exit and classification.
	onlyHeadings := true
	for _, line := range filtered {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "# ") {
			onlyHeadings = false
			break
		}
	}
	if onlyHeadings {
		return nil
	}
	copy := *failure
	copy.output = strings.Join(filtered, "\n")
	copy.display = limitDiscoveryEvidence(copy.output, 64<<10)
	return &copy
}

func filterDiscoveryAsmDeclTextMetadataLines(lines []string, packages []discoveryGoListPackage) []string {
	retained := make([]string, 0, len(lines))
	packagePath := ""
	for _, line := range lines {
		if strings.HasPrefix(line, "# ") {
			packagePath = strings.TrimSpace(strings.TrimPrefix(line, "# "))
			// Vet prints both '# package' and '# [package]' headings.
			packagePath = strings.Trim(packagePath, "[]")
		}
		match := discoveryAsmDeclZeroArgs.FindStringSubmatch(strings.TrimSpace(line))
		if match != nil && discoveryAsmDeclHasLiteralUnspecifiedArgs(match, packagePath, packages) {
			continue
		}
		match = discoveryAsmDeclPaddedArgs.FindStringSubmatch(strings.TrimSpace(line))
		if match != nil && discoveryAsmDeclHasLiteralPaddedArgs(match, packagePath, packages) {
			continue
		}
		retained = append(retained, line)
	}
	return retained
}

func discoveryAsmDeclHasLiteralUnspecifiedArgs(match []string, packagePath string, packages []discoveryGoListPackage) bool {
	text := discoveryAsmDeclSelectedText(match[1], match[2], packagePath, packages)
	if text == nil {
		return false
	}
	symbol := text[1]
	if beforeABI, _, ok := strings.Cut(symbol, "<ABI"); ok {
		symbol = beforeABI
	}
	if text[3] != "" {
		args, err := strconv.ParseInt(text[3], 10, 64)
		if err != nil || args != 0 {
			return false
		}
	}
	return symbol == match[4]
}

func discoveryAsmDeclHasLiteralPaddedArgs(match []string, packagePath string, packages []discoveryGoListPackage) bool {
	text := discoveryAsmDeclSelectedText(match[1], match[2], packagePath, packages)
	if text == nil || text[3] == "" || strings.Contains(text[1], "<ABIInternal>") {
		return false
	}
	symbol := strings.TrimSuffix(text[1], "<ABI0>")
	declared, declaredErr := strconv.ParseInt(match[5], 10, 64)
	expected, expectedErr := strconv.ParseInt(match[6], 10, 64)
	actual, actualErr := strconv.ParseInt(text[3], 10, 64)
	return symbol == match[4] && declaredErr == nil && expectedErr == nil && actualErr == nil &&
		actual == declared && goabi.MatchesABI0TextSize(match[3], declared, expected)
}

func discoveryAsmDeclSelectedText(filename, reportedLine, packagePath string, packages []discoveryGoListPackage) []string {
	reported := filepath.Clean(filepath.FromSlash(filename))
	var selected string
	for _, pkg := range packages {
		if packagePath != "" && pkg.ImportPath != packagePath {
			continue
		}
		for _, name := range pkg.SFiles {
			file := filepath.Join(pkg.Dir, name)
			if reported != file && reported != name {
				continue
			}
			if selected != "" && selected != file {
				return nil
			}
			selected = file
		}
	}
	if selected == "" {
		return nil
	}
	source, err := os.ReadFile(selected)
	if err != nil {
		return nil
	}
	lineNumber, err := strconv.Atoi(reportedLine)
	lines := discoveryAsmDeclUncommentedLines(source)
	if err != nil || lineNumber < 1 || lineNumber > len(lines) {
		return nil
	}
	return discoveryAsmDeclLiteralText.FindStringSubmatch(lines[lineNumber-1])
}

// Go's assembly lexer uses text/scanner comments. Mask them without moving
// physical line numbers, including block-comment licenses before TEXT, while
// keeping comment delimiters inside DATA strings intact.
func discoveryAsmDeclUncommentedLines(source []byte) []string {
	var lexer scanner.Scanner
	lexer.Init(bytes.NewReader(source))
	lexer.Mode = scanner.ScanComments | scanner.ScanStrings | scanner.ScanRawStrings | scanner.ScanChars
	valid := true
	lexer.Error = func(*scanner.Scanner, string) { valid = false }
	uncommented := bytes.Clone(source)
	for token := lexer.Scan(); token != scanner.EOF; token = lexer.Scan() {
		if token != scanner.Comment {
			continue
		}
		start := lexer.Position.Offset
		end := start + len(lexer.TokenText())
		for index := start; index < end; index++ {
			if uncommented[index] != '\n' && uncommented[index] != '\r' {
				uncommented[index] = ' '
			}
		}
	}
	if !valid {
		return nil
	}
	return strings.Split(string(uncommented), "\n")
}
