package gotoolprofile

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"text/scanner"
)

const CPPPackageSearchMissProtocol = "actual_go_fixed_include_search_miss_v1"

// CPPPackageSearchMiss records an unavailable package-directory candidate.
// Only the independently bound canonical GOROOT fallback may replace it. The
// producer and consumer recheck ENOENT before and after their compiler phases.
type CPPPackageSearchMiss struct {
	Protocol         string `json:"protocol"`
	Assembly         string `json:"assembly"`
	Source           string `json:"source"`
	Include          string `json:"include"`
	PackageCandidate string `json:"package_candidate"`
	ToolSource       string `json:"tool_source"`
}

func CPPLocalSource(file string) bool {
	return file != "" && file != "." && !path.IsAbs(file) && path.Clean(file) == file &&
		file != ".." && !strings.HasPrefix(file, "../") && !strings.ContainsAny(file, ":\\\x00\r\n")
}

// CPPIncludeCandidates follows cmd/asm's fixed original package directory and
// GOROOT/pkg/include search. A candidate may escape one origin but not both;
// callers must not read an escaping candidate or accept its existence.
func CPPIncludeCandidates(assembly, include string) (module, tool string, err error) {
	if !CPPLocalSource(assembly) || include == "" || path.IsAbs(include) || strings.ContainsAny(include, ":\\\x00\r\n") {
		return "", "", fmt.Errorf("unsafe fixed Go CPP include")
	}
	module = path.Clean(path.Join(path.Dir(assembly), include))
	tool = path.Clean(path.Join("pkg/include", include))
	if !CPPLocalSource(module) && !CPPLocalSource(tool) {
		return "", "", fmt.Errorf("CPP include escapes both exact source origins")
	}
	return module, tool, nil
}

func ValidateCPPPackageSearchMiss(miss CPPPackageSearchMiss) error {
	module, tool, err := CPPIncludeCandidates(miss.Assembly, miss.Include)
	origin, source, present := strings.Cut(miss.Source, "/")
	if err != nil || miss.Protocol != CPPPackageSearchMissProtocol || !present ||
		(origin != "module" && origin != "tool") || !CPPLocalSource(source) ||
		CPPLocalSource(module) || !CPPLocalSource(tool) ||
		miss.PackageCandidate != module || miss.ToolSource != tool {
		return fmt.Errorf("missing exact outside-package search absence witness")
	}
	return nil
}

func VerifyCPPPackageSearchMiss(miss CPPPackageSearchMiss, moduleRoot, goRoot string) error {
	if err := ValidateCPPPackageSearchMiss(miss); err != nil {
		return err
	}
	preferred, err := CPPPreferredPackageInclude(moduleRoot, miss.Assembly, miss.Include)
	if err != nil {
		return err
	}
	if _, err := os.Lstat(preferred); err == nil {
		return fmt.Errorf("outside-module preferred CPP candidate exists")
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("outside-module preferred CPP candidate has no ENOENT proof: %w", err)
	}
	origin, source, _ := strings.Cut(miss.Source, "/")
	root := moduleRoot
	if origin == "tool" {
		root = goRoot
	}
	name, err := CPPRegularSource(root, source)
	if err != nil {
		return err
	}
	data, err := os.ReadFile(name)
	if err != nil {
		return err
	}
	if !cppLiteralInclude(data, miss.Include) {
		return fmt.Errorf("CPP search absence has no original literal source operand")
	}
	return nil
}

func CPPPreferredPackageInclude(moduleRoot, assembly, include string) (string, error) {
	if _, _, err := CPPIncludeCandidates(assembly, include); err != nil {
		return "", err
	}
	root, err := filepath.EvalSymlinks(moduleRoot)
	if err != nil || !filepath.IsAbs(root) {
		return "", fmt.Errorf("CPP search has no actual canonical module origin")
	}
	packageDir := filepath.Join(root, filepath.FromSlash(path.Dir(assembly)))
	physicalDir, err := filepath.EvalSymlinks(packageDir)
	if err != nil || physicalDir != packageDir {
		return "", fmt.Errorf("CPP package directory is missing or redirected")
	}
	return filepath.Join(physicalDir, filepath.FromSlash(include)), nil
}

// Active consumption derives this from the actual including source, even
// when the fallback still has the historical pkg/include prefix. A caller
// cannot delete the witness and borrow that older tool-header allowance.
func RequireCPPPackageSearchMiss(input *ConsumerInput, assembly, source, include string) error {
	module, tool, err := CPPIncludeCandidates(assembly, include)
	if err != nil {
		return err
	}
	if CPPLocalSource(module) {
		return nil
	}
	expected := CPPPackageSearchMiss{
		Protocol: CPPPackageSearchMissProtocol, Assembly: assembly, Source: source,
		Include: include, PackageCandidate: module, ToolSource: tool,
	}
	for _, miss := range input.CPPIncludeSearchMisses {
		if miss == expected {
			return nil
		}
	}
	return fmt.Errorf("active outside-package CPP include lacks its original fixed-search absence")
}

func ValidateCPPIncludeSearch(input *ConsumerInput, proof *SelectionProof) error {
	if input == nil || proof != nil && !reflect.DeepEqual(input.CPPIncludeSearchMisses, proof.CPPIncludeSearchMisses) {
		return fmt.Errorf("actual CPP package search absence differs from its original input")
	}
	seen, tools := make(map[string]bool), make(map[string]bool)
	last := ""
	for _, miss := range input.CPPIncludeSearchMisses {
		key := miss.Assembly + "\x00" + miss.Source + "\x00" + miss.Include
		origin, source, _ := strings.Cut(miss.Source, "/")
		digest := input.Sources[source]
		if origin == "tool" {
			digest = input.ToolSources[source]
		}
		if err := ValidateCPPPackageSearchMiss(miss); err != nil || seen[key] || key <= last ||
			input.Sources[miss.Assembly] == "" || digest == "" || input.ToolSources[miss.ToolSource] == "" {
			return fmt.Errorf("CPP search absence is unbound, duplicated or noncanonical")
		}
		seen[key], tools[miss.ToolSource], last = true, true, key
	}
	for source := range input.ToolSources {
		if !CPPLocalSource(source) || !strings.HasPrefix(source, "pkg/include/") && !tools[source] {
			return fmt.Errorf("tool CPP source lacks the exact fixed-search absence contract")
		}
	}
	return nil
}

func SortCPPIncludeSearch(misses []CPPPackageSearchMiss) {
	sort.Slice(misses, func(i, j int) bool {
		x, y := misses[i], misses[j]
		return x.Assembly+"\x00"+x.Source+"\x00"+x.Include < y.Assembly+"\x00"+y.Source+"\x00"+y.Include
	})
}

// CPPRegularSource binds a canonical regular source within its actual origin.
// In particular GOROOT-relative includes cannot redirect through symlinks.
func CPPRegularSource(root, file string) (string, error) {
	if !filepath.IsAbs(root) || !CPPLocalSource(file) {
		return "", fmt.Errorf("unsafe exact CPP source identity")
	}
	resolvedRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		return "", err
	}
	name := filepath.Join(resolvedRoot, filepath.FromSlash(file))
	resolved, err := filepath.EvalSymlinks(name)
	if err != nil {
		return "", fmt.Errorf("CPP source is missing: %s: %w", file, err)
	}
	if resolved != name {
		return "", fmt.Errorf("CPP source is redirected: %s", file)
	}
	info, err := os.Lstat(name)
	if err != nil || !info.Mode().IsRegular() || info.Size() > 64<<20 {
		return "", fmt.Errorf("CPP source must be a bounded regular file: %s", file)
	}
	return name, nil
}

func cppLiteralInclude(data []byte, wanted string) bool {
	var scan scanner.Scanner
	scan.Init(strings.NewReader(string(data)))
	scan.Mode = scanner.ScanIdents | scanner.ScanStrings | scanner.ScanChars | scanner.ScanInts |
		scanner.ScanFloats | scanner.ScanComments | scanner.SkipComments
	scan.Whitespace = 1<<' ' | 1<<'\t' | 1<<'\r'
	invalid := false
	scan.Error = func(*scanner.Scanner, string) { invalid = true }
	for token := scan.Scan(); token != scanner.EOF; token = scan.Scan() {
		if token != '#' {
			continue
		}
		if scan.Scan() != scanner.Ident || scan.TokenText() != "include" || scan.Scan() != scanner.String {
			continue
		}
		literal, err := strconv.Unquote(scan.TokenText())
		terminator := scan.Scan()
		// This binds an original literal, not a complete input-stack directive.
		// Included EOF may resume a parent's newline; Go and active CPP still
		// require that newline and reject translation-unit EOF independently.
		if (terminator == '\n' || terminator == scanner.EOF) && err == nil && literal == wanted && !invalid {
			return true
		}
	}
	return false
}
