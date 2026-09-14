package main

import (
	"archive/zip"
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"go/build"
	"go/build/constraint"
	"go/parser"
	"go/token"
	"io"
	"math/bits"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/xgo-dev/plan9asm/internal/discoverymeta"
	"golang.org/x/mod/modfile"
	"golang.org/x/mod/module"
)

const discoveryReportSchema = 2

const (
	discoveryStatusPassed        = "passed"
	discoveryStatusFailed        = "failed"
	discoveryStatusNotApplicable = "not_applicable"
)

type discoveryRecord struct {
	Kind          string   `json:"kind"`
	Module        string   `json:"module"`
	Version       string   `json:"version"`
	Architectures []string `json:"architectures,omitempty"`
	AsmFiles      []string `json:"asm_files,omitempty"`
}

type discoveryCandidate struct {
	Module        string   `json:"module"`
	Version       string   `json:"version"`
	Architectures []string `json:"architectures"`
	AsmFiles      []string `json:"asm_files"`
}

func (c discoveryCandidate) exactKey() string {
	return c.Module + "@" + c.Version
}

type discoveryCorpusConfig struct {
	LedgerPath       string
	RepoRoot         string
	Translator       string
	LLC              string
	Targets          []string
	FilterTargets    bool
	ShardIndex       int
	ShardCount       int
	CandidateTimeout time.Duration
	ReportPath       string
}

type discoveryExecutionPlan struct {
	ModulePath string
	Patterns   []string
	GoMod      string
}

type discoveryBuildConfiguration struct {
	BuildTags []string `json:"build_tags,omitempty"`
	Targets   []string `json:"targets"`
	AsmFiles  []string `json:"asm_files"`
}

type discoveryPackageGroup struct {
	Pattern  string
	AsmFiles []string
}

const (
	discoverySourceNotApplicableGoAssembler = "go_assembler_rejected_all_supported_targets"
	discoverySourceNotApplicableNoGoPackage = "no_current_go_package"
	discoverySourceNotApplicableGoBuild     = "go_build_rejected_package_target"
	discoverySourceNotApplicableAsmDecl     = "go_vet_asmdecl_rejected_target"
)

type discoverySourceNotApplicableItem struct {
	AsmFile  string   `json:"asm_file,omitempty"`
	AsmFiles []string `json:"asm_files,omitempty"`
	Targets  []string `json:"targets"`
	Kind     string   `json:"kind"`
	Reason   string   `json:"reason"`
}

type moduleDownloadInfo struct {
	Path    string `json:"Path"`
	Version string `json:"Version"`
	Dir     string `json:"Dir"`
	GoMod   string `json:"GoMod"`
	Zip     string `json:"Zip"`
	Error   string `json:"Error"`
}

type discoveryCorpusResult struct {
	Module                    string                             `json:"module"`
	Version                   string                             `json:"version"`
	Status                    string                             `json:"status"`
	DiscoveredAsmFiles        []string                           `json:"discovered_asm_files"`
	ApplicableAsmFiles        []string                           `json:"applicable_asm_files"`
	BuildConfigurations       []discoveryBuildConfiguration      `json:"build_configurations,omitempty"`
	Patterns                  []string                           `json:"patterns,omitempty"`
	Translations              int                                `json:"translations"`
	NotApplicableTranslations int                                `json:"not_applicable_translations,omitempty"`
	NotApplicableItems        []matrixTargetNotApplicableItem    `json:"not_applicable_items,omitempty"`
	SourceNotApplicableItems  []discoverySourceNotApplicableItem `json:"source_not_applicable_items,omitempty"`
	NotApplicableReason       string                             `json:"not_applicable_reason,omitempty"`
	Error                     string                             `json:"error,omitempty"`
}

type discoveryCorpusReport struct {
	SchemaVersion             int                     `json:"schema_version"`
	Targets                   []string                `json:"targets,omitempty"`
	TargetFiltered            bool                    `json:"target_filtered,omitempty"`
	ShardIndex                int                     `json:"shard_index"`
	ShardCount                int                     `json:"shard_count"`
	CandidateTotal            int                     `json:"candidate_total"`
	EligibleCandidates        int                     `json:"eligible_candidates"`
	Selected                  int                     `json:"selected"`
	Passed                    int                     `json:"passed"`
	Failed                    int                     `json:"failed"`
	NotApplicable             int                     `json:"not_applicable"`
	Translations              int                     `json:"translations"`
	NotApplicableTranslations int                     `json:"not_applicable_translations"`
	Results                   []discoveryCorpusResult `json:"results"`
}

func loadDiscoveryCandidates(root string) ([]discoveryCandidate, error) {
	info, err := os.Stat(root)
	if err != nil {
		return nil, fmt.Errorf("stat discovery ledger: %w", err)
	}
	var files []string
	if !info.IsDir() {
		files = []string{root}
	} else {
		files, err = filepath.Glob(filepath.Join(root, "records", "*.jsonl"))
		if err != nil {
			return nil, fmt.Errorf("list discovery ledger records: %w", err)
		}
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("discovery ledger %q contains no JSONL records", root)
	}
	sort.Strings(files)
	type candidateSets struct {
		candidate discoveryCandidate
		arches    map[string]bool
		asmFiles  map[string]bool
	}
	merged := make(map[string]*candidateSets)
	for _, filePath := range files {
		file, err := os.Open(filePath)
		if err != nil {
			return nil, fmt.Errorf("open discovery records %s: %w", filePath, err)
		}
		scanner := bufio.NewScanner(file)
		scanner.Buffer(make([]byte, 64<<10), 32<<20)
		line := 0
		for scanner.Scan() {
			line++
			if len(bytes.TrimSpace(scanner.Bytes())) == 0 {
				continue
			}
			var record discoveryRecord
			if err := json.Unmarshal(scanner.Bytes(), &record); err != nil {
				file.Close()
				return nil, fmt.Errorf("decode discovery record %s:%d: %w", filePath, line, err)
			}
			if record.Kind != "matched" {
				continue
			}
			if err := validateDiscoveryRecord(record); err != nil {
				file.Close()
				return nil, fmt.Errorf("invalid discovery record %s:%d: %w", filePath, line, err)
			}
			key := record.Module + "@" + record.Version
			item := merged[key]
			if item == nil {
				item = &candidateSets{
					candidate: discoveryCandidate{Module: record.Module, Version: record.Version},
					arches:    make(map[string]bool),
					asmFiles:  make(map[string]bool),
				}
				merged[key] = item
			}
			for _, arch := range record.Architectures {
				item.arches[arch] = true
			}
			for _, asmFile := range record.AsmFiles {
				item.asmFiles[asmFile] = true
			}
		}
		if err := scanner.Err(); err != nil {
			file.Close()
			return nil, fmt.Errorf("read discovery records %s: %w", filePath, err)
		}
		if err := file.Close(); err != nil {
			return nil, fmt.Errorf("close discovery records %s: %w", filePath, err)
		}
	}
	candidates := make([]discoveryCandidate, 0, len(merged))
	for _, item := range merged {
		for arch := range item.arches {
			item.candidate.Architectures = append(item.candidate.Architectures, arch)
		}
		for asmFile := range item.asmFiles {
			item.candidate.AsmFiles = append(item.candidate.AsmFiles, asmFile)
		}
		sort.Strings(item.candidate.Architectures)
		sort.Strings(item.candidate.AsmFiles)
		candidates = append(candidates, item.candidate)
	}
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].Module != candidates[j].Module {
			return candidates[i].Module < candidates[j].Module
		}
		return candidates[i].Version < candidates[j].Version
	})
	return candidates, nil
}

func validateDiscoveryRecord(record discoveryRecord) error {
	if record.Module == "" || strings.TrimSpace(record.Module) != record.Module || strings.ContainsAny(record.Module, "\r\n\t ") {
		return fmt.Errorf("invalid module %q", record.Module)
	}
	if record.Version == "" || strings.TrimSpace(record.Version) != record.Version || strings.ContainsAny(record.Version, "\r\n\t ") {
		return fmt.Errorf("invalid version %q", record.Version)
	}
	if len(record.AsmFiles) == 0 {
		return fmt.Errorf("%s@%s has no assembly files", record.Module, record.Version)
	}
	for _, asmFile := range record.AsmFiles {
		if asmFile == "" || strings.Contains(asmFile, "\\") || path.IsAbs(asmFile) || path.Clean(asmFile) != asmFile || strings.HasPrefix(asmFile, "../") || !strings.HasSuffix(asmFile, ".s") {
			return fmt.Errorf("%s@%s has unsafe assembly path %q", record.Module, record.Version, asmFile)
		}
	}
	return nil
}

func selectDiscoveryShard(candidates []discoveryCandidate, shardIndex, shardCount int) []discoveryCandidate {
	if shardCount <= 0 || shardIndex < 0 || shardIndex >= shardCount {
		return nil
	}
	selected := make([]discoveryCandidate, 0, len(candidates)/shardCount+1)
	for _, candidate := range candidates {
		sum := sha256.Sum256([]byte(candidate.exactKey()))
		if int(binary.BigEndian.Uint64(sum[:8])%uint64(shardCount)) == shardIndex {
			selected = append(selected, candidate)
		}
	}
	return selected
}

func filterDiscoveryCandidatesForTargets(candidates []discoveryCandidate, targets []string) ([]discoveryCandidate, error) {
	filtered := make([]discoveryCandidate, 0, len(candidates))
	for _, candidate := range candidates {
		files, err := discoverymeta.FilterAssemblyFiles(candidate.AsmFiles, targets)
		if err != nil {
			return nil, err
		}
		if len(files) == 0 {
			continue
		}
		candidate.AsmFiles = files
		candidate.Architectures = discoverymeta.ArchitectureHints(files)
		filtered = append(filtered, candidate)
	}
	return filtered, nil
}

func discoveryPackagePatterns(candidate discoveryCandidate) []string {
	return discoveryPackagePatternsForModule(candidate, candidate.Module)
}

func discoveryPackagePatternsForModule(candidate discoveryCandidate, modulePath string) []string {
	groups := discoveryPackageGroupsForModule(candidate, modulePath)
	out := make([]string, 0, len(groups))
	for _, group := range groups {
		out = append(out, group.Pattern)
	}
	return out
}

func discoveryPackageGroupsForModule(candidate discoveryCandidate, modulePath string) []discoveryPackageGroup {
	filesByPattern := make(map[string][]string)
	for _, asmFile := range candidate.AsmFiles {
		dir := path.Dir(asmFile)
		pattern := modulePath
		if dir != "." {
			pattern += "/" + dir
		}
		filesByPattern[pattern] = append(filesByPattern[pattern], asmFile)
	}
	patterns := make([]string, 0, len(filesByPattern))
	for pattern := range filesByPattern {
		patterns = append(patterns, pattern)
	}
	sort.Strings(patterns)
	groups := make([]discoveryPackageGroup, 0, len(patterns))
	for _, pattern := range patterns {
		groups = append(groups, discoveryPackageGroup{
			Pattern:  pattern,
			AsmFiles: uniqueSortedDiscoveryStrings(filesByPattern[pattern]),
		})
	}
	return groups
}

func makeDiscoveryExecutionPlan(candidate discoveryCandidate, declaredModule string) (discoveryExecutionPlan, error) {
	if declaredModule == "" {
		declaredModule = candidate.Module
	}
	if strings.TrimSpace(declaredModule) != declaredModule || strings.ContainsAny(declaredModule, "\r\n\t ") {
		return discoveryExecutionPlan{}, fmt.Errorf("invalid declared module path %q", declaredModule)
	}
	goMod := fmt.Sprintf("module plan9asm.local/discovery\n\ngo 1.27\n\nrequire %s %s\n", declaredModule, candidate.Version)
	if declaredModule != candidate.Module {
		goMod += fmt.Sprintf("\nreplace %s => %s %s\n", declaredModule, candidate.Module, candidate.Version)
	}
	return discoveryExecutionPlan{
		ModulePath: declaredModule,
		Patterns:   discoveryPackagePatternsForModule(candidate, declaredModule),
		GoMod:      goMod,
	}, nil
}

func parseDeclaredModulePath(contents []byte) (string, error) {
	scanner := bufio.NewScanner(bytes.NewReader(contents))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "module") || (len(line) > len("module") && line[len("module")] != ' ' && line[len("module")] != '\t') {
			continue
		}
		rest := strings.TrimSpace(strings.TrimPrefix(line, "module"))
		if rest == "" {
			return "", fmt.Errorf("empty module directive")
		}
		var modulePath string
		if rest[0] == '"' || rest[0] == '`' {
			if _, err := fmt.Sscanf(rest, "%q", &modulePath); err != nil {
				return "", fmt.Errorf("parse module directive %q: %w", line, err)
			}
		} else {
			modulePath = strings.Fields(rest)[0]
		}
		if modulePath == "" || strings.ContainsAny(modulePath, "\r\n\t ") {
			return "", fmt.Errorf("invalid module directive path %q", modulePath)
		}
		return modulePath, nil
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return "", fmt.Errorf("go.mod has no module directive")
}

func applicableDiscoveryAssemblyFiles(candidate discoveryCandidate, moduleDir string, targets []string) ([]string, error) {
	configs, err := discoveryBuildConfigurations(candidate, moduleDir, targets)
	if err != nil {
		return nil, err
	}
	var applicable []string
	for _, config := range configs {
		applicable = append(applicable, config.AsmFiles...)
	}
	sort.Strings(applicable)
	return applicable, nil
}

func discoveryBuildConfigurations(candidate discoveryCandidate, moduleDir string, targets []string) ([]discoveryBuildConfiguration, error) {
	return discoveryBuildConfigurationsWithEvidence(candidate, moduleDir, targets, nil)
}

func discoveryBuildConfigurationsWithEvidence(candidate discoveryCandidate, moduleDir string, targets []string, sourceNotApplicable *[]discoverySourceNotApplicableItem) ([]discoveryBuildConfiguration, error) {
	if moduleDir == "" {
		return nil, fmt.Errorf("%s: downloaded module has no directory", candidate.exactKey())
	}
	contexts := make([]build.Context, 0, len(targets))
	for _, target := range targets {
		goos, goarch, ok := strings.Cut(target, "/")
		if !ok || goos == "" || goarch == "" {
			return nil, fmt.Errorf("invalid discovery target %q", target)
		}
		ctx := build.Default
		ctx.GOOS = goos
		ctx.GOARCH = goarch
		ctx.Compiler = "gc"
		ctx.CgoEnabled = false
		contexts = append(contexts, ctx)
	}

	goFilesByDir := make(map[string][]string)
	invalidGoFilesByDir := make(map[string][]string)
	constraintTagsByFile := make(map[string][]string)
	configsByTargetAndTags := make(map[string]*discoveryBuildConfiguration)
	for _, asmFile := range candidate.AsmFiles {
		dirRel := path.Dir(asmFile)
		if discoveryDirIsIgnored(dirRel) {
			continue
		}
		nested, err := discoveryDirIsNestedModule(moduleDir, dirRel)
		if err != nil {
			return nil, err
		}
		if nested {
			continue
		}
		dir := filepath.Join(moduleDir, filepath.FromSlash(dirRel))
		goFiles, ok := goFilesByDir[dir]
		if !ok {
			entries, err := os.ReadDir(dir)
			if err != nil {
				return nil, fmt.Errorf("read assembly package directory %s: %w", dir, err)
			}
			for _, entry := range entries {
				name := entry.Name()
				if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") || strings.HasPrefix(name, ".") || strings.HasPrefix(name, "_") {
					continue
				}
				if _, parseErr := parser.ParseFile(token.NewFileSet(), filepath.Join(dir, name), nil, parser.PackageClauseOnly); parseErr != nil {
					invalidGoFilesByDir[dir] = append(invalidGoFilesByDir[dir], name)
					continue
				}
				goFiles = append(goFiles, name)
			}
			goFilesByDir[dir] = goFiles
		}
		if len(goFiles) == 0 {
			if sourceNotApplicable != nil {
				invalid := uniqueSortedDiscoveryStrings(invalidGoFilesByDir[dir])
				reason := "assembly directory contains no non-test Go source accepted by Go 1.27"
				if len(invalid) != 0 {
					reason += "; invalid .go files: " + strings.Join(invalid, ", ")
				}
				*sourceNotApplicable = append(*sourceNotApplicable, discoverySourceNotApplicableItem{
					AsmFile: asmFile,
					Targets: uniqueSortedDiscoveryStrings(targets),
					Kind:    discoverySourceNotApplicableNoGoPackage,
					Reason:  reason,
				})
			}
			continue
		}
		fileNames := append([]string(nil), goFiles...)
		fileNames = append(fileNames, path.Base(asmFile))
		var customTags []string
		for _, fileName := range fileNames {
			fullPath := filepath.Join(dir, fileName)
			tags, ok := constraintTagsByFile[fullPath]
			if !ok {
				var err error
				tags, err = discoveryConstraintTags(fullPath)
				if err != nil {
					return nil, err
				}
				constraintTagsByFile[fullPath] = tags
			}
			customTags = append(customTags, tags...)
		}
		customTags = uniqueDiscoveryCustomTags(customTags, contexts)
		if len(customTags) > 16 {
			return nil, fmt.Errorf("%s: assembly package %s has %d custom build tags; refusing to leave the satisfiability search incomplete", candidate.exactKey(), dirRel, len(customTags))
		}
		type contextMatch struct {
			context   build.Context
			target    string
			buildTags []string
		}
		matches := make([]contextMatch, 0, len(contexts))
		for i := range contexts {
			target := contexts[i].GOOS + "/" + contexts[i].GOARCH
			buildTags, ok, err := findDiscoveryBuildTags(contexts[i], dir, path.Base(asmFile), goFiles, customTags)
			if err != nil {
				return nil, fmt.Errorf("classify assembly %s: %w", asmFile, err)
			}
			if !ok {
				continue
			}
			matches = append(matches, contextMatch{context: contexts[i], target: target, buildTags: buildTags})
		}
		matchedContexts := make([]build.Context, 0, len(matches))
		for _, match := range matches {
			matchedContexts = append(matchedContexts, match.context)
		}
		sourceTargets, restrictBySource, reason, err := inferUnsuffixedAssemblyTargetsDetailed(filepath.Join(moduleDir, filepath.FromSlash(asmFile)), matchedContexts)
		if err != nil {
			return nil, fmt.Errorf("classify assembly source %s: %w", asmFile, err)
		}
		if reason != "" && sourceNotApplicable != nil {
			var sourceTargetNames []string
			for _, match := range matches {
				sourceTargetNames = append(sourceTargetNames, match.target)
			}
			*sourceNotApplicable = append(*sourceNotApplicable, discoverySourceNotApplicableItem{
				AsmFile: asmFile,
				Targets: uniqueSortedDiscoveryStrings(sourceTargetNames),
				Kind:    discoverySourceNotApplicableGoAssembler,
				Reason:  reason,
			})
		}
		for _, match := range matches {
			if restrictBySource && !sourceTargets[match.target] {
				continue
			}
			key := match.target + "\x00" + strings.Join(match.buildTags, ",")
			config := configsByTargetAndTags[key]
			if config == nil {
				config = &discoveryBuildConfiguration{
					BuildTags: append([]string(nil), match.buildTags...),
					Targets:   []string{match.target},
				}
				configsByTargetAndTags[key] = config
			}
			config.AsmFiles = append(config.AsmFiles, asmFile)
		}
	}
	// Targets can share one translator invocation only when both their build
	// tags and exact source-inferred assembly allowlist are identical.
	configsByShape := make(map[string]*discoveryBuildConfiguration)
	for _, config := range configsByTargetAndTags {
		config.AsmFiles = uniqueSortedDiscoveryStrings(config.AsmFiles)
		shape := strings.Join(config.BuildTags, ",") + "\x00" + strings.Join(config.AsmFiles, "\x00")
		group := configsByShape[shape]
		if group == nil {
			group = &discoveryBuildConfiguration{
				BuildTags: append([]string(nil), config.BuildTags...),
				AsmFiles:  append([]string(nil), config.AsmFiles...),
			}
			configsByShape[shape] = group
		}
		group.Targets = append(group.Targets, config.Targets...)
	}
	keys := make([]string, 0, len(configsByShape))
	for key := range configsByShape {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	configs := make([]discoveryBuildConfiguration, 0, len(keys))
	for _, key := range keys {
		config := configsByShape[key]
		config.Targets = uniqueSortedDiscoveryStrings(config.Targets)
		configs = append(configs, *config)
	}
	return configs, nil
}

// inferUnsuffixedAssemblyTargets adds a source-level check only when Go's
// filename convention does not already identify an architecture. Go itself
// selects files by name and build tags; external packages sometimes omit both
// while containing unmistakably architecture-specific assembly. Asking the
// same Go assembler that will build the package prevents such x86 sources from
// being sent to ARM or Wasm translators without maintaining a second opcode
// taxonomy here.
func inferUnsuffixedAssemblyTargets(filePath string, contexts []build.Context) (map[string]bool, bool, error) {
	eligible, restrict, _, err := inferUnsuffixedAssemblyTargetsDetailed(filePath, contexts)
	return eligible, restrict, err
}

func inferUnsuffixedAssemblyTargetsDetailed(filePath string, contexts []build.Context) (map[string]bool, bool, string, error) {
	if explicitAssemblyFilenameArchitecture(path.Base(filePath)) != "" {
		return nil, false, "", nil
	}
	eligible := make(map[string]bool, len(contexts))
	probed := make(map[string]struct {
		accepted   bool
		conclusive bool
	})
	acceptedAny := false
	conclusiveAny := false
	for _, ctx := range contexts {
		key := ctx.GOOS + "/" + ctx.GOARCH
		result, ok := probed[key]
		if !ok {
			accepted, conclusive := probeAssemblySourceForTarget(filePath, ctx.GOOS, ctx.GOARCH)
			result = struct {
				accepted   bool
				conclusive bool
			}{accepted: accepted, conclusive: conclusive}
			probed[key] = result
		}
		if result.accepted || !result.conclusive {
			eligible[key] = true
		}
		acceptedAny = acceptedAny || result.accepted
		conclusiveAny = conclusiveAny || result.conclusive
	}
	if !acceptedAny && conclusiveAny && len(eligible) == 0 {
		// The source is a fixture, generated artifact, or assembly for a target
		// outside this repository's supported matrix. The current Go assembler
		// supplies stronger applicability evidence than an opcode guess: keep an
		// explicit empty allowlist instead of failing translation or silently
		// treating it as cross-architecture source.
		var rejectedTargets []string
		for target := range probed {
			rejectedTargets = append(rejectedTargets, target)
		}
		sort.Strings(rejectedTargets)
		return eligible, true, fmt.Sprintf("current Go assembler rejected unsuffixed source for every selected target: %s", strings.Join(rejectedTargets, ", ")), nil
	}
	if !conclusiveAny {
		// Usually a generated go_asm.h (or another build-generated include) was
		// unavailable. Keep every target visible instead of silently excluding it.
		return nil, false, "", nil
	}
	return eligible, true, "", nil
}

func explicitAssemblyFilenameArchitecture(name string) string {
	base := strings.TrimSuffix(name, filepath.Ext(name))
	for _, arch := range []string{
		"386", "amd64", "amd64p32", "arm", "arm64", "armbe", "arm64be",
		"loong64", "mips", "mipsle", "mips64", "mips64le", "mips64p32",
		"mips64p32le", "ppc", "ppc64", "ppc64le", "riscv", "riscv64",
		"s390", "s390x", "sparc", "sparc64", "wasm",
	} {
		if strings.HasSuffix(base, "_"+arch) {
			return arch
		}
	}
	return ""
}

func probeAssemblySourceForTarget(filePath, goos, goarch string) (accepted, conclusive bool) {
	run := func(extraInclude string) ([]byte, error) {
		args := []string{"tool", "asm"}
		if extraInclude != "" {
			args = append(args, "-I", extraInclude)
		}
		args = append(args, "-I", filepath.Dir(filePath), "-I", filepath.Join(runtime.GOROOT(), "pkg", "include"), "-o", os.DevNull, filePath)
		cmd := exec.Command("go", args...)
		cmd.Env = replaceEnv(os.Environ(), map[string]string{
			"GOOS":        goos,
			"GOARCH":      goarch,
			"GOTOOLCHAIN": "local",
			"GOWORK":      "off",
		})
		return cmd.CombinedOutput()
	}
	output, err := run("")
	if err == nil {
		return true, true
	}
	message := strings.ToLower(string(output))
	if missingGoAsmHeader(message) {
		stubDir, stubErr := os.MkdirTemp("", "plan9asm-go-asm-header-")
		if stubErr != nil {
			return false, false
		}
		defer os.RemoveAll(stubDir)
		if stubErr := os.WriteFile(filepath.Join(stubDir, "go_asm.h"), nil, 0o600); stubErr != nil {
			return false, false
		}
		output, err = run(stubDir)
		if err == nil {
			return true, true
		}
		message = strings.ToLower(string(output))
	}
	for _, fragment := range []string{"no such file or directory", "cannot find", "could not find"} {
		if strings.Contains(message, fragment) {
			return false, false
		}
	}
	return false, true
}

func missingGoAsmHeader(message string) bool {
	message = strings.ToLower(message)
	if !strings.Contains(message, "go_asm.h") {
		return false
	}
	for _, fragment := range []string{"no such file or directory", "cannot find", "could not find"} {
		if strings.Contains(message, fragment) {
			return true
		}
	}
	return false
}

func discoveryConstraintTags(filePath string) ([]string, error) {
	if strings.HasSuffix(filePath, ".go") {
		return discoveryGoConstraintTags(filePath)
	}
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("open build constraints %s: %w", filePath, err)
	}
	defer file.Close()
	set := make(map[string]bool)
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 64<<10), 4<<20)
	inBlockComment := false
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if inBlockComment {
			if strings.Contains(line, "*/") {
				inBlockComment = false
			}
			continue
		}
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "/*") {
			inBlockComment = !strings.Contains(line, "*/")
			continue
		}
		if !strings.HasPrefix(line, "//") {
			break
		}
		if !strings.HasPrefix(line, "//go:build ") && !strings.HasPrefix(line, "// +build ") {
			continue
		}
		expr, err := constraint.Parse(line)
		if err != nil {
			return nil, fmt.Errorf("parse build constraint %s: %w", filePath, err)
		}
		collectDiscoveryConstraintTags(expr, set)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read build constraints %s: %w", filePath, err)
	}
	tags := make([]string, 0, len(set))
	for tag := range set {
		tags = append(tags, tag)
	}
	sort.Strings(tags)
	return tags, nil
}

func discoveryGoConstraintTags(filePath string) ([]string, error) {
	parsed, err := parser.ParseFile(token.NewFileSet(), filePath, nil, parser.PackageClauseOnly|parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("parse Go build constraints %s: %w", filePath, err)
	}
	set := make(map[string]bool)
	for _, group := range parsed.Comments {
		for _, comment := range group.List {
			line := strings.TrimSpace(comment.Text)
			if !strings.HasPrefix(line, "//go:build ") && !strings.HasPrefix(line, "// +build ") {
				continue
			}
			expr, err := constraint.Parse(line)
			if err != nil {
				return nil, fmt.Errorf("parse build constraint %s: %w", filePath, err)
			}
			collectDiscoveryConstraintTags(expr, set)
		}
	}
	tags := make([]string, 0, len(set))
	for tag := range set {
		tags = append(tags, tag)
	}
	sort.Strings(tags)
	return tags, nil
}

func collectDiscoveryConstraintTags(expr constraint.Expr, tags map[string]bool) {
	switch expr := expr.(type) {
	case *constraint.TagExpr:
		tags[expr.Tag] = true
	case *constraint.NotExpr:
		collectDiscoveryConstraintTags(expr.X, tags)
	case *constraint.AndExpr:
		collectDiscoveryConstraintTags(expr.X, tags)
		collectDiscoveryConstraintTags(expr.Y, tags)
	case *constraint.OrExpr:
		collectDiscoveryConstraintTags(expr.X, tags)
		collectDiscoveryConstraintTags(expr.Y, tags)
	}
}

func uniqueDiscoveryCustomTags(tags []string, contexts []build.Context) []string {
	reserved := map[string]bool{
		// "ignore" conventionally marks generator inputs and other files that
		// are not part of any build. Enabling it globally also selects Go's own
		// ignored generator programs and produces packages that cannot compile.
		"ignore": true,
		"aix":    true, "android": true, "darwin": true, "dragonfly": true,
		"freebsd": true, "hurd": true, "illumos": true, "ios": true,
		"js": true, "linux": true, "netbsd": true, "openbsd": true,
		"plan9": true, "solaris": true, "unix": true, "wasip1": true,
		"windows": true, "zos": true,
		"386": true, "amd64": true, "amd64p32": true, "arm": true,
		"armbe": true, "arm64": true, "arm64be": true, "loong64": true,
		"mips": true, "mipsle": true, "mips64": true, "mips64le": true,
		"mips64p32": true, "mips64p32le": true, "ppc": true, "ppc64": true,
		"ppc64le": true, "riscv": true, "riscv64": true, "s390": true,
		"s390x": true, "sparc": true, "sparc64": true, "wasm": true,
		"cgo": true, "gc": true, "gccgo": true,
	}
	for _, ctx := range contexts {
		reserved[ctx.GOOS] = true
		reserved[ctx.GOARCH] = true
		reserved[ctx.Compiler] = true
		for _, tag := range ctx.ReleaseTags {
			reserved[tag] = true
		}
		for _, tag := range ctx.ToolTags {
			reserved[tag] = true
		}
	}
	set := make(map[string]bool)
	for _, tag := range tags {
		if tag != "" && !reserved[tag] {
			set[tag] = true
		}
	}
	out := make([]string, 0, len(set))
	for tag := range set {
		out = append(out, tag)
	}
	sort.Strings(out)
	return out
}

func findDiscoveryBuildTags(ctx build.Context, dir, asmFile string, goFiles, customTags []string) ([]string, bool, error) {
	assignments := 1 << len(customTags)
	for wantedCount := 0; wantedCount <= len(customTags); wantedCount++ {
		for mask := 0; mask < assignments; mask++ {
			if bits.OnesCount(uint(mask)) != wantedCount {
				continue
			}
			buildTags := make([]string, 0, wantedCount)
			for i, tag := range customTags {
				if mask&(1<<i) != 0 {
					buildTags = append(buildTags, tag)
				}
			}
			ctx.BuildTags = buildTags
			asmMatches, err := ctx.MatchFile(dir, asmFile)
			if err != nil {
				return nil, false, fmt.Errorf("match assembly for %s/%s with tags %v: %w", ctx.GOOS, ctx.GOARCH, buildTags, err)
			}
			if !asmMatches {
				continue
			}
			for _, goFile := range goFiles {
				matches, err := ctx.MatchFile(dir, goFile)
				if err != nil {
					return nil, false, fmt.Errorf("match Go file %s for %s/%s with tags %v: %w", goFile, ctx.GOOS, ctx.GOARCH, buildTags, err)
				}
				if matches {
					return buildTags, true, nil
				}
			}
		}
	}
	return nil, false, nil
}

func uniqueSortedDiscoveryStrings(values []string) []string {
	out := uniqueDiscoveryStrings(values)
	sort.Strings(out)
	return out
}

func uniqueDiscoveryStrings(values []string) []string {
	set := make(map[string]bool, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		if set[value] {
			continue
		}
		set[value] = true
		out = append(out, value)
	}
	return out
}

func discoveryDirIsIgnored(dir string) bool {
	if dir == "." || dir == "" {
		return false
	}
	for _, elem := range strings.Split(filepath.ToSlash(dir), "/") {
		if elem == "testdata" || elem == "vendor" || strings.HasPrefix(elem, ".") || strings.HasPrefix(elem, "_") {
			return true
		}
	}
	return false
}

func discoveryDirIsNestedModule(moduleDir, dirRel string) (bool, error) {
	for dirRel != "." && dirRel != "" {
		_, err := os.Stat(filepath.Join(moduleDir, filepath.FromSlash(dirRel), "go.mod"))
		switch {
		case err == nil:
			return true, nil
		case os.IsNotExist(err):
			// Continue toward the root module.
		case err != nil:
			return false, fmt.Errorf("stat nested go.mod in %s: %w", dirRel, err)
		}
		dirRel = path.Dir(dirRel)
	}
	return false, nil
}

func runDiscoveryCorpus(cfg discoveryCorpusConfig) error {
	if cfg.ShardCount <= 0 || cfg.ShardIndex < 0 || cfg.ShardIndex >= cfg.ShardCount {
		return fmt.Errorf("invalid discovery shard %d/%d", cfg.ShardIndex, cfg.ShardCount)
	}
	if cfg.CandidateTimeout <= 0 {
		return fmt.Errorf("candidate timeout must be positive")
	}
	allCandidates, err := loadDiscoveryCandidates(cfg.LedgerPath)
	if err != nil {
		return err
	}
	candidates := allCandidates
	if cfg.FilterTargets {
		candidates, err = filterDiscoveryCandidatesForTargets(allCandidates, cfg.Targets)
		if err != nil {
			return err
		}
	}
	selected := selectDiscoveryShard(candidates, cfg.ShardIndex, cfg.ShardCount)
	report := discoveryCorpusReport{
		SchemaVersion:      discoveryReportSchema,
		Targets:            append([]string(nil), cfg.Targets...),
		TargetFiltered:     cfg.FilterTargets,
		ShardIndex:         cfg.ShardIndex,
		ShardCount:         cfg.ShardCount,
		CandidateTotal:     len(allCandidates),
		EligibleCandidates: len(candidates),
		Selected:           len(selected),
		Results:            make([]discoveryCorpusResult, 0, len(selected)),
	}
	tmpRoot, err := os.MkdirTemp("", fmt.Sprintf("plan9asm-discovery-%02d-", cfg.ShardIndex))
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpRoot)
	for i, candidate := range selected {
		fmt.Printf("[%d/%d] %s\n", i+1, len(selected), candidate.exactKey())
		result := discoveryCorpusResult{
			Module:             candidate.Module,
			Version:            candidate.Version,
			DiscoveredAsmFiles: append([]string(nil), candidate.AsmFiles...),
		}
		candidateDir := filepath.Join(tmpRoot, fmt.Sprintf("candidate-%04d", i))
		matrix, patterns, buildConfigurations, runErr := runDiscoveryCandidate(cfg, candidate, candidateDir)
		applicableAsmFiles := discoveryConfigurationAsmFiles(buildConfigurations)
		result.Patterns = patterns
		result.ApplicableAsmFiles = applicableAsmFiles
		result.BuildConfigurations = buildConfigurations
		result.NotApplicableItems = append(result.NotApplicableItems, matrix.NotApplicableItems...)
		result.SourceNotApplicableItems = append(result.SourceNotApplicableItems, matrix.SourceNotApplicableItems...)
		if runErr != nil {
			result.Status = discoveryStatusFailed
			result.Error = runErr.Error()
			report.Failed++
			fmt.Fprintf(os.Stderr, "FAIL %s: %v\n", candidate.exactKey(), runErr)
		} else if len(applicableAsmFiles) == 0 || matrix.Success == 0 {
			result.Status = discoveryStatusNotApplicable
			if len(applicableAsmFiles) == 0 {
				result.NotApplicableReason = "no discovered assembly file belongs to a buildable Go package on the supported target matrix"
			} else {
				result.NotApplicableReason = "every target-selected assembly source has evidence-backed current-Go source or ABI incompatibility"
			}
			result.NotApplicableTranslations = matrix.NotApplicable
			report.NotApplicableTranslations += matrix.NotApplicable
			report.NotApplicable++
			fmt.Printf("NOT_APPLICABLE %s: discovered_files=%d\n", candidate.exactKey(), len(candidate.AsmFiles))
		} else {
			result.Status = discoveryStatusPassed
			result.Translations = matrix.Success
			result.NotApplicableTranslations = matrix.NotApplicable
			report.Passed++
			report.Translations += matrix.Success
			report.NotApplicableTranslations += matrix.NotApplicable
			fmt.Printf("PASS %s: applicable_files=%d translations=%d not_applicable_translations=%d targets=%d\n", candidate.exactKey(), len(applicableAsmFiles), matrix.Success, matrix.NotApplicable, matrix.TotalTargets)
		}
		report.Results = append(report.Results, result)
	}
	if err := validateDiscoveryCorpusAccounting(report); err != nil {
		return err
	}
	if err := writeDiscoveryCorpusReport(cfg.ReportPath, report); err != nil {
		return err
	}
	if report.Failed != 0 {
		return fmt.Errorf("discovery shard %d/%d failed: passed=%d failed=%d selected=%d", cfg.ShardIndex, cfg.ShardCount, report.Passed, report.Failed, report.Selected)
	}
	fmt.Printf("discovery shard %d/%d passed: applicable=%d not_applicable=%d translations=%d not_applicable_translations=%d\n", cfg.ShardIndex, cfg.ShardCount, report.Passed, report.NotApplicable, report.Translations, report.NotApplicableTranslations)
	return nil
}

func validateDiscoveryCorpusAccounting(report discoveryCorpusReport) error {
	if report.Selected != report.Passed+report.Failed+report.NotApplicable {
		return fmt.Errorf("discovery report accounting mismatch: selected=%d passed=%d failed=%d not_applicable=%d", report.Selected, report.Passed, report.Failed, report.NotApplicable)
	}
	if len(report.Results) != 0 && len(report.Results) != report.Selected {
		return fmt.Errorf("discovery report result count mismatch: selected=%d results=%d", report.Selected, len(report.Results))
	}
	for _, result := range report.Results {
		if err := validateDiscoverySourceNotApplicableEvidence(result); err != nil {
			return fmt.Errorf("%s@%s: %w", result.Module, result.Version, err)
		}
	}
	return nil
}

func validateDiscoverySourceNotApplicableEvidence(result discoveryCorpusResult) error {
	discovered := make(map[string]bool, len(result.DiscoveredAsmFiles))
	for _, asmFile := range result.DiscoveredAsmFiles {
		discovered[asmFile] = true
	}
	allowedKinds := map[string]bool{
		discoverySourceNotApplicableGoAssembler: true,
		discoverySourceNotApplicableNoGoPackage: true,
		discoverySourceNotApplicableGoBuild:     true,
		discoverySourceNotApplicableAsmDecl:     true,
	}
	for _, item := range result.SourceNotApplicableItems {
		files := append([]string(nil), item.AsmFiles...)
		if item.AsmFile != "" {
			files = append(files, item.AsmFile)
		}
		files = uniqueSortedDiscoveryStrings(files)
		valid := allowedKinds[item.Kind] && len(files) != 0 && len(item.Targets) != 0 && item.Reason != ""
		for _, asmFile := range files {
			valid = valid && discovered[asmFile]
		}
		for _, target := range item.Targets {
			goos, goarch, ok := strings.Cut(target, "/")
			valid = valid && ok && goos != "" && goarch != ""
		}
		if !valid {
			return fmt.Errorf("invalid source not-applicable evidence: kind=%q files=%v targets=%v", item.Kind, files, item.Targets)
		}
	}
	return nil
}

func discoveryConfigurationAsmFiles(configs []discoveryBuildConfiguration) []string {
	var files []string
	for _, config := range configs {
		files = append(files, config.AsmFiles...)
	}
	return uniqueSortedDiscoveryStrings(files)
}

func runDiscoveryCandidate(cfg discoveryCorpusConfig, candidate discoveryCandidate, workDir string) (matrixReport, []string, []discoveryBuildConfiguration, error) {
	if err := os.MkdirAll(workDir, 0755); err != nil {
		return matrixReport{}, nil, nil, err
	}
	if err := os.WriteFile(filepath.Join(workDir, "go.mod"), []byte("module plan9asm.local/discovery\n\ngo 1.27\n"), 0644); err != nil {
		return matrixReport{}, nil, nil, fmt.Errorf("write temporary go.mod: %w", err)
	}
	env := replaceEnv(os.Environ(), map[string]string{
		"GOFLAGS":     "-mod=mod",
		"GOTOOLCHAIN": "local",
		"GOWORK":      "off",
	})
	ctx, cancel := context.WithTimeout(context.Background(), cfg.CandidateTimeout)
	defer cancel()
	downloadJSON, commandErr := runCapturedCommandOutput(ctx, workDir, env, "go", "mod", "download", "-json", candidate.Module+"@"+candidate.Version)
	download, err := resolveModuleDownload(downloadJSON, commandErr, filepath.Join(workDir, "module-source"))
	if err != nil {
		return matrixReport{}, nil, nil, fmt.Errorf("download module: %w", err)
	}
	var sourceNotApplicable []discoverySourceNotApplicableItem
	buildConfigurations, err := discoveryBuildConfigurationsWithEvidence(candidate, download.Dir, cfg.Targets, &sourceNotApplicable)
	if err != nil {
		return matrixReport{}, nil, nil, fmt.Errorf("classify discovered assembly: %w", err)
	}
	if len(buildConfigurations) == 0 {
		return matrixReport{SourceNotApplicableItems: sourceNotApplicable}, nil, buildConfigurations, nil
	}
	declaredModule := candidate.Module
	if download.GoMod != "" {
		goModContents, err := os.ReadFile(download.GoMod)
		if err != nil {
			return matrixReport{}, nil, buildConfigurations, fmt.Errorf("read downloaded go.mod: %w", err)
		}
		declaredModule, err = parseDeclaredModulePath(goModContents)
		if err != nil {
			return matrixReport{}, nil, buildConfigurations, fmt.Errorf("read declared module path: %w", err)
		}
	}
	patternSet := make(map[string]bool)
	aggregate := matrixReport{SourceNotApplicableItems: sourceNotApplicable}
	runTargets := make(map[string]bool)
	executedBuildConfigurations := make([]discoveryBuildConfiguration, 0, len(buildConfigurations))
	invocationIndex := 0
	for _, buildConfiguration := range buildConfigurations {
		applicableCandidate := candidate
		applicableCandidate.AsmFiles = buildConfiguration.AsmFiles
		plan, err := makeDiscoveryExecutionPlan(applicableCandidate, declaredModule)
		if err != nil {
			return matrixReport{}, nil, buildConfigurations, err
		}
		if err := os.WriteFile(filepath.Join(workDir, "go.mod"), []byte(plan.GoMod), 0644); err != nil {
			return matrixReport{}, nil, buildConfigurations, fmt.Errorf("write exact module mapping: %w", err)
		}
		packageGroups := discoveryPackageGroupsForModule(applicableCandidate, plan.ModulePath)
		for _, target := range buildConfiguration.Targets {
			var targetAsmFiles, targetPatterns []string
			for _, group := range packageGroups {
				if err := runDiscoveryGoBuild(ctx, workDir, env, target, buildConfiguration.BuildTags, group.Pattern); err != nil {
					if isDiscoveryGoBuildInfrastructureFailure(err.Error()) {
						return matrixReport{}, sortedDiscoverySet(patternSet), executedBuildConfigurations, fmt.Errorf("verify current Go package %s for %s with build tags %v: %w", group.Pattern, target, buildConfiguration.BuildTags, err)
					}
					aggregate.SourceNotApplicableItems = append(aggregate.SourceNotApplicableItems, discoverySourceNotApplicableItem{
						AsmFiles: append([]string(nil), group.AsmFiles...),
						Targets:  []string{target},
						Kind:     discoverySourceNotApplicableGoBuild,
						Reason:   limitDiscoveryEvidence(err.Error(), 8192),
					})
					continue
				}
				validAsmFiles := append([]string(nil), group.AsmFiles...)
				if err := runDiscoveryAsmDecl(ctx, workDir, env, target, buildConfiguration.BuildTags, []string{group.Pattern}); err != nil {
					rejected := discoveryAsmDeclRejectedFiles(validAsmFiles, err.Error())
					if len(rejected) == 0 {
						rejected = append(rejected, validAsmFiles...)
					}
					aggregate.SourceNotApplicableItems = append(aggregate.SourceNotApplicableItems, discoverySourceNotApplicableItem{
						AsmFiles: append([]string(nil), rejected...),
						Targets:  []string{target},
						Kind:     discoverySourceNotApplicableAsmDecl,
						Reason:   limitDiscoveryEvidence(err.Error(), 8192),
					})
					validAsmFiles = subtractDiscoveryStrings(validAsmFiles, rejected)
				}
				if len(validAsmFiles) == 0 {
					continue
				}
				targetAsmFiles = append(targetAsmFiles, validAsmFiles...)
				targetPatterns = append(targetPatterns, group.Pattern)
			}
			if len(targetAsmFiles) == 0 {
				continue
			}
			targetAsmFiles = uniqueSortedDiscoveryStrings(targetAsmFiles)
			targetPatterns = uniqueSortedDiscoveryStrings(targetPatterns)
			for _, pattern := range targetPatterns {
				patternSet[pattern] = true
			}
			targetCandidate := candidate
			targetCandidate.AsmFiles = targetAsmFiles
			executed := discoveryBuildConfiguration{
				BuildTags: append([]string(nil), buildConfiguration.BuildTags...),
				Targets:   []string{target},
				AsmFiles:  append([]string(nil), targetAsmFiles...),
			}
			executedBuildConfigurations = append(executedBuildConfigurations, executed)
			reportPath := filepath.Join(workDir, fmt.Sprintf("matrix-report-%04d.json", invocationIndex))
			invocation := makeTranslatorInvocationForTargetsAndTags(
				workDir,
				plan.ModulePath,
				targetPatterns,
				buildConfiguration.BuildTags,
				[]string{target},
				targetAsmFiles,
				filepath.Join(workDir, "out", fmt.Sprintf("config-%04d", invocationIndex)),
				cfg.RepoRoot,
				cfg.LLC,
				reportPath,
			)
			invocationIndex++
			if err := runCapturedCommand(ctx, invocation.Dir, env, cfg.Translator, invocation.Args...); err != nil {
				return matrixReport{}, sortedDiscoverySet(patternSet), executedBuildConfigurations, fmt.Errorf("translate and compile target %s with build tags %v: %w", target, buildConfiguration.BuildTags, err)
			}
			report, err := loadReport(reportPath)
			if err != nil {
				return matrixReport{}, sortedDiscoverySet(patternSet), executedBuildConfigurations, err
			}
			if err := validateDiscoveryReport([]string{target}, targetCandidate, report); err != nil {
				return matrixReport{}, sortedDiscoverySet(patternSet), executedBuildConfigurations, fmt.Errorf("target %s build tags %v: %w", target, buildConfiguration.BuildTags, err)
			}
			runTargets[target] = true
			aggregate.TotalAsm += report.TotalAsm
			aggregate.Success += report.Success
			aggregate.NotApplicable += report.NotApplicable
			aggregate.Failed += report.Failed
			aggregate.NotApplicableItems = append(aggregate.NotApplicableItems, collectMatrixNotApplicableItems(report)...)
		}
	}
	aggregate.TotalTargets = len(runTargets)
	return aggregate, sortedDiscoverySet(patternSet), executedBuildConfigurations, nil
}

func resolveModuleDownload(downloadJSON []byte, commandErr error, extractDir string) (moduleDownloadInfo, error) {
	var download moduleDownloadInfo
	if err := json.Unmarshal(downloadJSON, &download); err != nil {
		if commandErr != nil {
			return moduleDownloadInfo{}, commandErr
		}
		return moduleDownloadInfo{}, fmt.Errorf("decode module download metadata: %w", err)
	}
	if download.Dir != "" {
		if info, err := os.Stat(download.Dir); err == nil && info.IsDir() {
			return download, nil
		}
	}
	if download.Zip != "" {
		if err := extractModuleZip(download.Zip, extractDir, download.Path, download.Version); err == nil {
			download.Dir = extractDir
			if download.GoMod == "" {
				download.GoMod = filepath.Join(extractDir, "go.mod")
			}
			return download, nil
		} else if commandErr == nil && download.Error == "" {
			return moduleDownloadInfo{}, fmt.Errorf("extract cached module ZIP: %w", err)
		}
	}
	if download.Error != "" {
		return moduleDownloadInfo{}, errors.New(download.Error)
	}
	if commandErr != nil {
		return moduleDownloadInfo{}, commandErr
	}
	return moduleDownloadInfo{}, errors.New("module download metadata contains neither a directory nor a usable ZIP")
}

func extractModuleZip(zipPath, destination, modulePath, version string) error {
	escapedPath, err := module.EscapePath(modulePath)
	if err != nil {
		return fmt.Errorf("escape module path: %w", err)
	}
	escapedVersion, err := module.EscapeVersion(version)
	if err != nil {
		return fmt.Errorf("escape module version: %w", err)
	}
	prefix := escapedPath + "@" + escapedVersion + "/"
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer reader.Close()
	if err := os.MkdirAll(destination, 0755); err != nil {
		return err
	}
	for _, file := range reader.File {
		name := filepath.ToSlash(file.Name)
		if !strings.HasPrefix(name, prefix) {
			return fmt.Errorf("module ZIP entry %q is outside expected prefix %q", file.Name, prefix)
		}
		relative := strings.TrimPrefix(name, prefix)
		if relative == "" {
			continue
		}
		clean := path.Clean(relative)
		if clean == "." || clean == ".." || strings.HasPrefix(clean, "../") || path.IsAbs(clean) {
			return fmt.Errorf("unsafe module ZIP path %q", file.Name)
		}
		target := filepath.Join(destination, filepath.FromSlash(clean))
		rel, err := filepath.Rel(destination, target)
		if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			return fmt.Errorf("unsafe module ZIP path %q", file.Name)
		}
		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
			continue
		}
		if !file.Mode().IsRegular() {
			return fmt.Errorf("unsupported non-regular module ZIP entry %q", file.Name)
		}
		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return err
		}
		source, err := file.Open()
		if err != nil {
			return err
		}
		output, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
		if err == nil {
			_, err = io.Copy(output, source)
		}
		closeOutputErr := error(nil)
		if output != nil {
			closeOutputErr = output.Close()
		}
		closeSourceErr := source.Close()
		if err != nil {
			return err
		}
		if closeOutputErr != nil {
			return closeOutputErr
		}
		if closeSourceErr != nil {
			return closeSourceErr
		}
	}
	return nil
}

func runDiscoveryGoBuild(ctx context.Context, dir string, env []string, target string, buildTags []string, pattern string) error {
	goos, goarch, ok := strings.Cut(target, "/")
	if !ok || goos == "" || goarch == "" {
		return fmt.Errorf("invalid discovery target %q", target)
	}
	args := []string{"build"}
	if len(buildTags) != 0 {
		args = append(args, "-tags="+strings.Join(buildTags, ","))
	}
	args = append(args, pattern)
	targetEnv := replaceEnv(env, map[string]string{
		"CGO_ENABLED": "0",
		"GOOS":        goos,
		"GOARCH":      goarch,
	})
	return runCapturedCommand(ctx, dir, targetEnv, "go", args...)
}

func isDiscoveryGoBuildInfrastructureFailure(diagnostic string) bool {
	diagnostic = strings.ToLower(diagnostic)
	// "unexpected EOF" is also the Go assembler's deterministic diagnostic
	// for a truncated source file. Do not retain that source failure as a
	// transient network retry merely because proxies use the same phrase.
	if strings.Contains(diagnostic, "unexpected eof") &&
		strings.Contains(diagnostic, ".s:") &&
		strings.Contains(diagnostic, "asm: assembly of") {
		return false
	}
	for _, marker := range []string{
		"context deadline exceeded",
		"i/o timeout",
		"tls handshake timeout",
		"temporary failure",
		"no such host",
		"connection refused",
		"connection reset",
		"network is unreachable",
		"proxyconnect",
		"unexpected eof",
		"bad gateway",
		"service unavailable",
		"gateway timeout",
		"no space left on device",
		"signal: killed",
	} {
		if strings.Contains(diagnostic, marker) {
			return true
		}
	}
	return false
}

func runDiscoveryAsmDecl(ctx context.Context, dir string, env []string, target string, buildTags, patterns []string) error {
	goos, goarch, ok := strings.Cut(target, "/")
	if !ok || goos == "" || goarch == "" {
		return fmt.Errorf("invalid discovery target %q", target)
	}
	args := []string{"vet", "-asmdecl"}
	if len(buildTags) != 0 {
		args = append(args, "-tags="+strings.Join(buildTags, ","))
	}
	args = append(args, patterns...)
	targetEnv := replaceEnv(env, map[string]string{
		"CGO_ENABLED": "0",
		"GOOS":        goos,
		"GOARCH":      goarch,
	})
	_, err := runCapturedCommandOutput(ctx, dir, targetEnv, "go", args...)
	if err == nil {
		return nil
	}
	if isDiscoveryAsmDeclABIMismatch(err.Error()) {
		return err
	}
	// go vet type-checks package tests before running asmdecl. Old modules can
	// have tests that no longer compile even though their production package
	// builds; retry against temporary module copies with only *_test.go files
	// emptied so unrelated diagnostics cannot hide an assembly ABI mismatch.
	// Go forbids overlays beneath GOMODCACHE, hence the explicit module copies.
	retryErr := runDiscoveryAsmDeclWithTestlessModuleCopies(ctx, dir, targetEnv, args, buildTags, patterns)
	if retryErr != nil && isDiscoveryAsmDeclABIMismatch(retryErr.Error()) {
		return retryErr
	}
	return nil
}

type discoveryGoListPackage struct {
	Dir          string
	TestGoFiles  []string
	XTestGoFiles []string
	Module       *struct {
		Path  string
		Dir   string
		GoMod string
	}
}

func runDiscoveryAsmDeclWithTestlessModuleCopies(ctx context.Context, dir string, env, vetArgs, buildTags, patterns []string) error {
	listArgs := []string{"list", "-json"}
	if len(buildTags) != 0 {
		listArgs = append(listArgs, "-tags="+strings.Join(buildTags, ","))
	}
	listArgs = append(listArgs, patterns...)
	output, err := runCapturedCommandOutput(ctx, dir, env, "go", listArgs...)
	if err != nil {
		return err
	}

	goModPath := filepath.Join(dir, "go.mod")
	goModData, err := os.ReadFile(goModPath)
	if err != nil {
		return err
	}
	parsedGoMod, err := modfile.Parse(goModPath, goModData, nil)
	if err != nil {
		return err
	}

	decoder := json.NewDecoder(bytes.NewReader(output))
	packages := make([]discoveryGoListPackage, 0)
	for {
		var pkg discoveryGoListPackage
		if err := decoder.Decode(&pkg); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		packages = append(packages, pkg)
	}

	stages := make(map[string]string)
	hasTests := false
	for _, pkg := range packages {
		if len(pkg.TestGoFiles)+len(pkg.XTestGoFiles) == 0 {
			continue
		}
		hasTests = true
		if pkg.Module == nil || pkg.Module.Path == "" || pkg.Module.Dir == "" {
			return fmt.Errorf("cannot create testless asmdecl copy for package %s without module metadata", pkg.Dir)
		}
		stage := stages[pkg.Module.Dir]
		if stage == "" {
			stage, err = os.MkdirTemp("", "plan9asm-vet-module-")
			if err != nil {
				return err
			}
			defer os.RemoveAll(stage)
			if err := copyDiscoveryModuleTree(pkg.Module.Dir, stage); err != nil {
				return err
			}
			stagedGoMod := filepath.Join(stage, "go.mod")
			if _, err := os.Stat(stagedGoMod); os.IsNotExist(err) {
				moduleGoMod, readErr := os.ReadFile(pkg.Module.GoMod)
				if readErr != nil {
					return readErr
				}
				if err := os.WriteFile(stagedGoMod, moduleGoMod, 0644); err != nil {
					return err
				}
			} else if err != nil {
				return err
			}
			if err := parsedGoMod.AddReplace(pkg.Module.Path, "", stage, ""); err != nil {
				return err
			}
			stages[pkg.Module.Dir] = stage
		}
		for _, fileName := range append(append([]string(nil), pkg.TestGoFiles...), pkg.XTestGoFiles...) {
			sourcePath := filepath.Join(pkg.Dir, fileName)
			relativePath, err := filepath.Rel(pkg.Module.Dir, sourcePath)
			if err != nil || relativePath == ".." || strings.HasPrefix(relativePath, ".."+string(filepath.Separator)) {
				return fmt.Errorf("test file %s is outside module %s", sourcePath, pkg.Module.Dir)
			}
			parsed, err := parser.ParseFile(token.NewFileSet(), sourcePath, nil, parser.PackageClauseOnly)
			if err != nil {
				return err
			}
			if err := os.WriteFile(filepath.Join(stage, relativePath), []byte("package "+parsed.Name.Name+"\n"), 0644); err != nil {
				return err
			}
		}
	}
	if !hasTests {
		return nil
	}
	formattedGoMod, err := parsedGoMod.Format()
	if err != nil {
		return err
	}
	modfilePath := filepath.Join(dir, ".plan9asm-vet.mod")
	if err := os.WriteFile(modfilePath, formattedGoMod, 0644); err != nil {
		return err
	}
	args := append([]string{"vet", "-modfile=" + modfilePath}, vetArgs[1:]...)
	_, err = runCapturedCommandOutput(ctx, dir, env, "go", args...)
	return err
}

func copyDiscoveryModuleTree(source, destination string) error {
	return filepath.Walk(source, func(sourcePath string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relativePath, err := filepath.Rel(source, sourcePath)
		if err != nil {
			return err
		}
		destinationPath := filepath.Join(destination, relativePath)
		if info.IsDir() {
			return os.MkdirAll(destinationPath, info.Mode().Perm()|0700)
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("unsupported non-regular module file %s", sourcePath)
		}
		contents, err := os.ReadFile(sourcePath)
		if err != nil {
			return err
		}
		return os.WriteFile(destinationPath, contents, info.Mode().Perm()|0600)
	})
}

func isDiscoveryAsmDeclABIMismatch(diagnostic string) bool {
	diagnostic = strings.ToLower(diagnostic)
	if strings.Contains(diagnostic, "wrong argument size") || strings.Contains(diagnostic, "invalid offset") {
		return true
	}
	// asmdecl reports an invalid load/store width against an FP operand when
	// the declared Go parameter or result has a different ABI width. Keep the
	// FP requirement so generic assembler, dependency, and source diagnostics
	// continue into translation instead of being hidden as not applicable.
	return strings.Contains(diagnostic, ": invalid ") && strings.Contains(diagnostic, "(fp)")
}

func discoveryAsmDeclRejectedFiles(asmFiles []string, diagnostic string) []string {
	diagnostic = filepath.ToSlash(diagnostic)
	var rejected []string
	for _, asmFile := range asmFiles {
		asmFile = filepath.ToSlash(asmFile)
		if strings.Contains(diagnostic, asmFile+":") || strings.Contains(diagnostic, "/"+asmFile+":") {
			rejected = append(rejected, asmFile)
			continue
		}
		// The Go command sometimes shortens diagnostics to a package-local
		// basename. Only use that fallback when it identifies this file.
		if strings.Contains(diagnostic, path.Base(asmFile)+":") {
			rejected = append(rejected, asmFile)
		}
	}
	return uniqueSortedDiscoveryStrings(rejected)
}

func subtractDiscoveryStrings(values, removed []string) []string {
	removeSet := make(map[string]bool, len(removed))
	for _, value := range removed {
		removeSet[value] = true
	}
	out := make([]string, 0, len(values))
	for _, value := range values {
		if !removeSet[value] {
			out = append(out, value)
		}
	}
	return out
}

func limitDiscoveryEvidence(message string, maxBytes int) string {
	message = strings.TrimSpace(message)
	if maxBytes <= 0 || len(message) <= maxBytes {
		return message
	}
	return message[:maxBytes] + "\n... evidence truncated ..."
}

func collectMatrixNotApplicableItems(report matrixReport) []matrixTargetNotApplicableItem {
	var items []matrixTargetNotApplicableItem
	for _, target := range report.Targets {
		targetName := target.Goos + "/" + target.Goarch
		for _, item := range target.NotApplicableItems {
			items = append(items, matrixTargetNotApplicableItem{Target: targetName, targetNotApplicableItem: item})
		}
	}
	return items
}

func sortedDiscoverySet(set map[string]bool) []string {
	values := make([]string, 0, len(set))
	for value := range set {
		values = append(values, value)
	}
	sort.Strings(values)
	return values
}

func runCapturedCommand(ctx context.Context, dir string, env []string, name string, args ...string) error {
	_, err := runCapturedCommandOutput(ctx, dir, env, name, args...)
	return err
}

func runCapturedCommandOutput(ctx context.Context, dir string, env []string, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	cmd.Env = env
	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output
	err := cmd.Run()
	if ctx.Err() != nil {
		return nil, fmt.Errorf("%s: %w", strings.Join(append([]string{name}, args...), " "), ctx.Err())
	}
	if err == nil {
		return output.Bytes(), nil
	}
	message := strings.TrimSpace(output.String())
	if len(message) > 64<<10 {
		message = "... output truncated ...\n" + message[len(message)-(64<<10):]
	}
	if message == "" {
		return nil, fmt.Errorf("%s: %w", strings.Join(append([]string{name}, args...), " "), err)
	}
	return output.Bytes(), fmt.Errorf("%s: %w\n%s", strings.Join(append([]string{name}, args...), " "), err, message)
}

func replaceEnv(base []string, replacements map[string]string) []string {
	out := make([]string, 0, len(base)+len(replacements))
	for _, item := range base {
		key, _, ok := strings.Cut(item, "=")
		if ok {
			if _, replace := replacements[key]; replace {
				continue
			}
		}
		out = append(out, item)
	}
	keys := make([]string, 0, len(replacements))
	for key := range replacements {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		out = append(out, key+"="+replacements[key])
	}
	return out
}

func validateDiscoveryReport(targets []string, candidate discoveryCandidate, report matrixReport) error {
	if report.TotalTargets != len(targets) || len(report.Targets) != len(targets) {
		return fmt.Errorf("%s: target coverage changed: got total=%d reports=%d, want %d", candidate.exactKey(), report.TotalTargets, len(report.Targets), len(targets))
	}
	seen := make(map[string]bool, len(targets))
	totalAsm := 0
	totalSuccess := 0
	totalNotApplicable := 0
	for _, actual := range report.Targets {
		target := actual.Goos + "/" + actual.Goarch
		if seen[target] {
			return fmt.Errorf("%s: duplicate target report %s", candidate.exactKey(), target)
		}
		seen[target] = true
		if actual.Failed != 0 || actual.Success+actual.NotApplicable != actual.TotalAsm {
			return fmt.Errorf("%s: %s failed: success=%d not_applicable=%d failed=%d total=%d", candidate.exactKey(), target, actual.Success, actual.NotApplicable, actual.Failed, actual.TotalAsm)
		}
		if err := validateTargetNotApplicableEvidence(actual); err != nil {
			return fmt.Errorf("%s: %s: %w", candidate.exactKey(), target, err)
		}
		totalAsm += actual.TotalAsm
		totalSuccess += actual.Success
		totalNotApplicable += actual.NotApplicable
	}
	for _, target := range targets {
		if !seen[target] {
			return fmt.Errorf("%s: target %s was silently skipped", candidate.exactKey(), target)
		}
	}
	if totalAsm == 0 {
		return fmt.Errorf("%s: no assembly entered translation on the target matrix", candidate.exactKey())
	}
	observed := make(map[string]bool)
	for _, target := range report.Targets {
		for _, actualPath := range target.AsmFiles {
			actualPath = filepath.ToSlash(actualPath)
			for _, expectedPath := range candidate.AsmFiles {
				if actualPath == expectedPath || strings.HasSuffix(actualPath, "/"+expectedPath) {
					observed[expectedPath] = true
				}
			}
		}
	}
	var missing []string
	for _, expectedPath := range candidate.AsmFiles {
		if !observed[expectedPath] {
			missing = append(missing, expectedPath)
		}
	}
	if len(missing) != 0 {
		return fmt.Errorf("%s: %d supported assembly files were not exercised: %v", candidate.exactKey(), len(missing), missing)
	}
	if report.TotalAsm != totalAsm || report.Success != totalSuccess || report.NotApplicable != totalNotApplicable || report.Failed != 0 || report.Success+report.NotApplicable != report.TotalAsm {
		return fmt.Errorf("%s: inconsistent matrix totals: success=%d not_applicable=%d failed=%d total=%d", candidate.exactKey(), report.Success, report.NotApplicable, report.Failed, report.TotalAsm)
	}
	return nil
}

func validateTargetNotApplicableEvidence(report targetReport) error {
	if report.NotApplicable != len(report.NotApplicableItems) {
		return fmt.Errorf("not-applicable accounting mismatch: count=%d items=%d", report.NotApplicable, len(report.NotApplicableItems))
	}
	asmFiles := make(map[string]bool, len(report.AsmFiles))
	for _, asmFile := range report.AsmFiles {
		asmFiles[asmFile] = true
	}
	for _, item := range report.NotApplicableItems {
		if item.Kind != targetNotApplicableGoTextArgSize || item.PkgPath == "" || item.AsmFile == "" || !asmFiles[item.AsmFile] || item.Symbol == "" || item.DeclaredArgSize == item.ExpectedArgSize || item.Reason == "" {
			return fmt.Errorf("invalid not-applicable evidence for %q: kind=%q symbol=%q declared=%d expected=%d", item.AsmFile, item.Kind, item.Symbol, item.DeclaredArgSize, item.ExpectedArgSize)
		}
	}
	return nil
}

func writeDiscoveryCorpusReport(reportPath string, report discoveryCorpusReport) error {
	if reportPath == "" {
		return nil
	}
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("encode discovery corpus report: %w", err)
	}
	data = append(data, '\n')
	if err := os.MkdirAll(filepath.Dir(reportPath), 0755); err != nil {
		return fmt.Errorf("create discovery report directory: %w", err)
	}
	if err := os.WriteFile(reportPath, data, 0644); err != nil {
		return fmt.Errorf("write discovery corpus report: %w", err)
	}
	return nil
}
