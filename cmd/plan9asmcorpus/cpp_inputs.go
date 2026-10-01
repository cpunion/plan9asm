package main

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"golang.org/x/mod/sumdb/dirhash"
)

const discoveryCPPInputsProtocol = "exact_go_cpp_sources_v1"
const discoveryDeferredCPPInputsProtocol = "exact_go_cpp_deferred_sources_v2"

type discoveryCPPUnit struct {
	File             string            `json:"file"`
	Includes         map[string]string `json:"includes,omitempty"`
	DeferredIncludes map[string]string `json:"deferred_includes,omitempty"`
}

type discoveryCPPInputs struct {
	Protocol     string                        `json:"protocol"`
	Module       string                        `json:"module"`
	Version      string                        `json:"version"`
	ModuleSum    string                        `json:"module_sum"`
	ZipSHA256    string                        `json:"zip_sha256"`
	Registration *discoveryCPPRegistration     `json:"registration"`
	Sources      map[string]discoveryCPPSource `json:"sources"`
	Units        []discoveryCPPUnit            `json:"units"`
}

func captureDiscoveryCPPInputs(plan *discoveryOrdinarySelectionPlan, moduleDir, goRoot string, files []string, contexts ...context.Context) (*discoveryCPPInputs, error) {
	return captureDiscoveryCPPInputsMode(plan, moduleDir, goRoot, files, false, contexts...)
}

func captureDiscoveryDeferredCPPInputs(plan *discoveryOrdinarySelectionPlan, moduleDir, goRoot string, files []string, contexts ...context.Context) (*discoveryCPPInputs, error) {
	return captureDiscoveryCPPInputsMode(plan, moduleDir, goRoot, files, true, contexts...)
}

func captureDiscoveryCPPInputsMode(plan *discoveryOrdinarySelectionPlan, moduleDir, goRoot string, files []string, deferred bool, contexts ...context.Context) (*discoveryCPPInputs, error) {
	if len(contexts) > 1 || len(contexts) == 1 && contexts[0] == nil {
		return nil, fmt.Errorf("CPP source capture requires at most one live candidate context")
	}
	checkContext := func() error {
		if len(contexts) == 1 {
			return contexts[0].Err()
		}
		return nil
	}
	if err := checkContext(); err != nil {
		return nil, err
	}
	if plan == nil || !filepath.IsAbs(moduleDir) || !filepath.IsAbs(goRoot) || len(files) == 0 || len(files) > 512 {
		return nil, fmt.Errorf("CPP input capture requires bounded exact module/tool sources")
	}
	registration, err := captureDiscoveryCPPRegistration(goRoot, plan.GoVersion)
	if err != nil {
		return nil, err
	}
	inputs := &discoveryCPPInputs{Protocol: discoveryCPPInputsProtocol, Module: plan.Module, Version: plan.Version, ModuleSum: plan.ModuleSum,
		ZipSHA256: plan.ZipSHA256, Registration: registration, Sources: make(map[string]discoveryCPPSource)}
	if deferred {
		inputs.Protocol = discoveryDeferredCPPInputsProtocol
	}
	for _, file := range files {
		if !ordinarySelectionLocalPath(file) {
			return nil, fmt.Errorf("unsafe CPP assembly source %s", file)
		}
		unit := discoveryCPPUnit{File: file, Includes: make(map[string]string), DeferredIncludes: make(map[string]string)}
		active := make(map[string]bool)
		registered := make(map[string]bool)
		var capture func(string, int) error
		capture = func(id string, depth int) error {
			if err := checkContext(); err != nil {
				return err
			}
			if deferred && registered[id] {
				return nil // Registration is a finite graph, not active expansion.
			}
			if depth > 32 || !deferred && active[id] {
				return fmt.Errorf("CPP include nesting/cycle requires an explicit bounded expansion proof: %s", id)
			}
			origin, relative, err := discoveryCPPSourceIdentity(id)
			if err != nil {
				return err
			}
			root := moduleDir
			if origin == "tool" {
				root = goRoot
			}
			name := filepath.Join(root, filepath.FromSlash(relative))
			if err := discoveryCPPRegularSource(root, name); err != nil {
				return err
			}
			reader, err := os.Open(name)
			if err != nil {
				return err
			}
			data, readErr := io.ReadAll(io.LimitReader(reader, 64<<20+1))
			closeErr := reader.Close()
			if readErr != nil || closeErr != nil {
				return fmt.Errorf("read bounded CPP source: %v (close: %v)", readErr, closeErr)
			}
			source, err := discoveryCPPConditionsFromBytes(relative, data)
			if err != nil {
				return err
			}
			if err := checkContext(); err != nil {
				return err
			}
			if before, present := inputs.Sources[id]; present {
				if before.SHA256 != source.SHA256 {
					return fmt.Errorf("CPP source changed between translation units: %s", id)
				}
			} else {
				if len(inputs.Sources) >= 512 {
					return fmt.Errorf("CPP sources exceed the explicit 512-file bound")
				}
				inputs.Sources[id] = source
			}
			active[id] = true
			registered[id] = true
			defer delete(active, id)
			for index, directive := range source.Directives {
				if directive.Kind != "include" {
					continue
				}
				key := id + "#" + strconv.Itoa(index)
				included, kind, err := discoveryCPPResolveRawInclude(moduleDir, goRoot, unit.File, directive.Include, deferred)
				if err != nil {
					return fmt.Errorf("CPP include %s:%d: %w", relative, directive.Line, err)
				}
				if kind != "" {
					unit.DeferredIncludes[key] = kind
					continue
				}
				unit.Includes[key] = included
				if err := capture(included, depth+1); err != nil {
					return err
				}
			}
			return nil
		}
		if err := capture("module/"+file, 0); err != nil {
			return nil, err
		}
		inputs.Units = append(inputs.Units, unit)
	}
	sort.Slice(inputs.Units, func(i, j int) bool { return inputs.Units[i].File < inputs.Units[j].File })
	if err := validateDiscoveryCPPInputs(inputs, plan, files); err != nil {
		return nil, err
	}
	if err := verifyDiscoveryCPPInputsUnchanged(inputs, moduleDir, goRoot); err != nil {
		return nil, err
	}
	return inputs, nil
}

// Raw registration never substitutes an empty generated header. Missing
// ordinary includes remain explicit unresolved edges; active replay/actual Go
// compilation must later discharge them, otherwise the scope fails.
func discoveryCPPResolveRawInclude(moduleDir, goRoot, asm, include string, deferred bool) (string, string, error) {
	if deferred && path.Clean(include) == "go_asm.h" {
		file := path.Clean(path.Join(path.Dir(asm), include))
		if !ordinarySelectionLocalPath(file) || strings.Contains(include, "\\") {
			return "", "", fmt.Errorf("generated include escapes exact module scope")
		}
		name := filepath.Join(moduleDir, filepath.FromSlash(file))
		if _, err := os.Stat(name); os.IsNotExist(err) {
			return "", "generated_go_asm", nil
		} else if err != nil {
			return "", "", err
		}
	}
	file, err := discoveryCPPResolveInclude(moduleDir, goRoot, asm, include)
	if deferred && errors.Is(err, os.ErrNotExist) {
		return "", "unresolved_include", nil
	}
	return file, "", err
}

func discoveryCPPSourceIdentity(id string) (string, string, error) {
	origin, file, found := strings.Cut(id, "/")
	if !found || origin != "module" && origin != "tool" || !ordinarySelectionLocalPath(file) {
		return "", "", fmt.Errorf("invalid CPP source origin/path %s", id)
	}
	return origin, file, nil
}

func discoveryCPPRegularSource(root, name string) error {
	resolvedRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		return err
	}
	resolved, err := filepath.EvalSymlinks(name)
	if err != nil {
		return err
	}
	relative, err := filepath.Rel(resolvedRoot, resolved)
	if err != nil || !filepath.IsLocal(relative) {
		return fmt.Errorf("CPP source escapes its exact source origin")
	}
	info, err := os.Lstat(name)
	if err != nil || !info.Mode().IsRegular() || info.Size() > 64<<20 {
		return fmt.Errorf("CPP source must be a bounded regular file")
	}
	return nil
}

func discoveryCPPResolveInclude(moduleDir, goRoot, asmFile, include string) (string, error) {
	if filepath.IsAbs(include) || strings.Contains(include, "\\") {
		return "", fmt.Errorf("CPP include is outside the explicitly bound module/tool namespace")
	}
	// cmd/asm checks CWD, then its fixed top-level source Dir and -I list.
	// cmd/go runs in the package Dir. Nested headers never change that Dir.
	moduleFile := path.Clean(path.Join(path.Dir(asmFile), include))
	if !ordinarySelectionLocalPath(moduleFile) {
		return "", fmt.Errorf("CPP include escapes exact module source")
	}
	modulePath := filepath.Join(moduleDir, filepath.FromSlash(moduleFile))
	if _, err := os.Stat(modulePath); err == nil {
		if err := discoveryCPPRegularSource(moduleDir, modulePath); err != nil {
			return "", err
		}
		return "module/" + moduleFile, nil
	} else if !os.IsNotExist(err) {
		return "", err
	}
	toolFile := path.Clean(path.Join("pkg/include", include))
	if !ordinarySelectionLocalPath(toolFile) {
		return "", fmt.Errorf("CPP include escapes exact Go source")
	}
	if err := discoveryCPPRegularSource(goRoot, filepath.Join(goRoot, filepath.FromSlash(toolFile))); err != nil {
		return "", fmt.Errorf("unbound CPP/generated include %q (not N/A): %w", include, err)
	}
	return "tool/" + toolFile, nil
}

func validateDiscoveryCPPInputs(inputs *discoveryCPPInputs, plan *discoveryOrdinarySelectionPlan, files []string) error {
	if inputs == nil || plan == nil || inputs.Protocol != discoveryCPPInputsProtocol && inputs.Protocol != discoveryDeferredCPPInputsProtocol || inputs.Module != plan.Module || inputs.Version != plan.Version ||
		inputs.ModuleSum != plan.ModuleSum || inputs.ZipSHA256 != plan.ZipSHA256 || !discoverySHA256Pattern.MatchString(inputs.ZipSHA256) ||
		!strings.HasPrefix(inputs.ModuleSum, "h1:") || len(inputs.Sources) == 0 || len(inputs.Sources) > 512 || len(inputs.Units) != len(files) || len(files) > 512 {
		return fmt.Errorf("missing exact bounded CPP module/source protocol")
	}
	digest, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(inputs.ModuleSum, "h1:"))
	if err != nil || len(digest) != sha256.Size {
		return fmt.Errorf("CPP module proof lacks an exact h1 digest")
	}
	if err := validateDiscoveryCPPRegistration(inputs.Registration, plan.GoVersion); err != nil {
		return err
	}
	rootSHAs, wanted := make(map[string]string), make(map[string]bool)
	for _, source := range plan.Sources {
		rootSHAs[source.File] = source.SHA256
	}
	for _, file := range files {
		if wanted[file] {
			return fmt.Errorf("duplicate CPP required assembly source")
		}
		wanted[file] = true
	}
	for id, source := range inputs.Sources {
		_, file, err := discoveryCPPSourceIdentity(id)
		if err != nil || source.File != file || !discoverySHA256Pattern.MatchString(source.SHA256) || source.Directives == nil || len(source.Directives) > discoveryCPPDirectiveLimit {
			return fmt.Errorf("invalid compact CPP source witness %s", id)
		}
		lastLine := 0
		for _, directive := range source.Directives {
			if directive.Line <= lastLine {
				return fmt.Errorf("CPP control inventory is not in source order")
			}
			lastLine = directive.Line
			switch directive.Kind {
			case "ifdef", "ifndef", "define", "undef":
				if !validDiscoveryCPPMacroName(directive.Name) || directive.Include != "" {
					return fmt.Errorf("missing exact CPP macro name")
				}
			case "include":
				if directive.Include == "" || directive.Name != "" || filepath.IsAbs(directive.Include) || strings.ContainsAny(directive.Include, "\\\r\n") {
					return fmt.Errorf("missing exact CPP include")
				}
			case "else", "endif", "line":
				if directive.Name != "" || directive.Include != "" {
					return fmt.Errorf("unexpected CPP control payload")
				}
			default:
				return fmt.Errorf("unregistered CPP control in source witness")
			}
		}
	}
	used := make(map[string]bool)
	for _, unit := range inputs.Units {
		root, exists := inputs.Sources["module/"+unit.File]
		if !wanted[unit.File] || !exists || root.SHA256 != rootSHAs[unit.File] {
			return fmt.Errorf("CPP root omitted, duplicated or detached from exact ordinary ASM source")
		}
		delete(wanted, unit.File)
		if _, err := discoveryCPPUnitDirectives(inputs, unit, used); err != nil {
			return err
		}
	}
	if len(wanted) != 0 || len(used) != len(inputs.Sources) {
		return fmt.Errorf("CPP input inventory contains missing/unexplained source scopes")
	}
	return nil
}

func discoveryCPPUnitDirectives(inputs *discoveryCPPInputs, unit discoveryCPPUnit, used map[string]bool) ([]discoveryCPPDirective, error) {
	if inputs.Protocol == discoveryDeferredCPPInputsProtocol {
		return discoveryDeferredCPPUnitRegistration(inputs, unit, used)
	}
	if len(unit.DeferredIncludes) != 0 {
		return nil, fmt.Errorf("legacy CPP protocol cannot contain deferred origin edges")
	}
	var directives []discoveryCPPDirective
	active, includeKeys := make(map[string]bool), make(map[string]bool)
	var expand func(string, int) error
	expand = func(id string, depth int) error {
		source, exists := inputs.Sources[id]
		if !exists || depth > 32 || active[id] {
			return fmt.Errorf("missing/unbounded CPP include input")
		}
		active[id], used[id] = true, true
		defer delete(active, id)
		for index, directive := range source.Directives {
			if len(directives) >= 4096 {
				return fmt.Errorf("CPP unit exceeds the explicit 4096-control bound")
			}
			if directive.Kind != "include" {
				directives = append(directives, directive)
				continue
			}
			key := id + "#" + strconv.Itoa(index)
			target, exists := unit.Includes[key]
			if !exists {
				return fmt.Errorf("CPP include has no source-binding proof")
			}
			moduleFile := path.Clean(path.Join(path.Dir(unit.File), directive.Include))
			toolFile := path.Clean(path.Join("pkg/include", directive.Include))
			if !ordinarySelectionLocalPath(moduleFile) || !ordinarySelectionLocalPath(toolFile) ||
				(target != "module/"+moduleFile && target != "tool/"+toolFile) {
				return fmt.Errorf("CPP include binding is outside Go's fixed package/tool search")
			}
			includeKeys[key] = true
			if err := expand(target, depth+1); err != nil {
				return err
			}
		}
		return nil
	}
	if err := expand("module/"+unit.File, 0); err != nil {
		return nil, err
	}
	if len(includeKeys) != len(unit.Includes) {
		return nil, fmt.Errorf("unexplained CPP include binding")
	}
	return directives, nil
}

// The producer compares real exact ZIP bytes to both full SHA and compact
// controls. Offline replay is frozen source/tool-provenance evidence, not an
// independently authenticated ZIP witness without reading those bytes again.
func verifyDiscoveryCPPModuleZIP(inputs *discoveryCPPInputs, plan *discoveryOrdinarySelectionPlan, zipPath string) error {
	if inputs == nil {
		return fmt.Errorf("missing exact CPP source proof")
	}
	var files []string
	for _, unit := range inputs.Units {
		files = append(files, unit.File)
	}
	if err := validateDiscoveryCPPInputs(inputs, plan, files); err != nil {
		return err
	}
	actualSHA, err := discoveryFeatureFileSHA256(zipPath)
	if err != nil || actualSHA != inputs.ZipSHA256 {
		return fmt.Errorf("CPP exact ZIP bytes differ from ordinary source proof")
	}
	sum, err := dirhash.HashZip(zipPath, dirhash.Hash1)
	if err != nil || sum != inputs.ModuleSum {
		return fmt.Errorf("CPP exact ZIP h1 differs from ordinary source proof")
	}
	archive, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer archive.Close()
	zipFiles := make(map[string]*zip.File)
	prefix := inputs.Module + "@" + inputs.Version + "/"
	for _, file := range archive.File {
		if file.FileInfo().IsDir() {
			continue
		}
		if !strings.HasPrefix(file.Name, prefix) {
			return fmt.Errorf("CPP ZIP is outside exact module/version")
		}
		name := strings.TrimPrefix(file.Name, prefix)
		if !ordinarySelectionLocalPath(name) || zipFiles[name] != nil {
			return fmt.Errorf("CPP ZIP has unsafe/duplicate source")
		}
		zipFiles[name] = file
	}
	for id, source := range inputs.Sources {
		origin, relative, _ := discoveryCPPSourceIdentity(id)
		if origin != "module" {
			continue
		}
		entry := zipFiles[relative]
		if entry == nil || entry.UncompressedSize64 > 64<<20 {
			return fmt.Errorf("CPP source missing/unbounded in exact ZIP")
		}
		reader, err := entry.Open()
		if err != nil {
			return err
		}
		data, readErr := io.ReadAll(io.LimitReader(reader, 64<<20+1))
		closeErr := reader.Close()
		if readErr != nil || closeErr != nil {
			return fmt.Errorf("read CPP exact ZIP source: %v (close: %v)", readErr, closeErr)
		}
		actual, err := discoveryCPPConditionsFromBytes(relative, data)
		if err != nil {
			return err
		}
		before, _ := json.Marshal(source)
		after, _ := json.Marshal(actual)
		if !bytes.Equal(before, after) {
			return fmt.Errorf("CPP compact controls/full SHA differ from actual ZIP bytes: %s", relative)
		}
	}
	for _, unit := range inputs.Units {
		for id, source := range inputs.Sources {
			for index, directive := range source.Directives {
				if directive.Kind != "include" {
					continue
				}
				key := id + "#" + strconv.Itoa(index)
				target, selected := unit.Includes[key]
				if kind := unit.DeferredIncludes[key]; kind != "" {
					moduleFile := path.Clean(path.Join(path.Dir(unit.File), directive.Include))
					if zipFiles[moduleFile] != nil {
						return fmt.Errorf("deferred CPP include falsely omitted an original exact ZIP member")
					}
					continue
				}
				if !selected {
					continue
				}
				moduleFile := path.Clean(path.Join(path.Dir(unit.File), directive.Include))
				if !ordinarySelectionLocalPath(moduleFile) {
					return fmt.Errorf("CPP ZIP include escapes module")
				}
				wanted := "tool/" + path.Clean(path.Join("pkg/include", directive.Include))
				if zipFiles[moduleFile] != nil {
					wanted = "module/" + moduleFile
				}
				if target != wanted {
					return fmt.Errorf("CPP include binding differs from actual exact ZIP search")
				}
			}
		}
	}
	return nil
}

func verifyDiscoveryCPPInputsUnchanged(inputs *discoveryCPPInputs, moduleDir, goRoot string) error {
	if inputs == nil || inputs.Registration == nil {
		return fmt.Errorf("missing actual CPP source/registration stability proof")
	}
	for file, wanted := range inputs.Registration.ToolSourceSHA256 {
		actual, err := discoveryFeatureFileSHA256(filepath.Join(goRoot, filepath.FromSlash(file)))
		if err != nil || actual != wanted {
			return fmt.Errorf("CPP registration source changed: %s", file)
		}
	}
	for id, source := range inputs.Sources {
		origin, file, err := discoveryCPPSourceIdentity(id)
		if err != nil {
			return err
		}
		root := moduleDir
		if origin == "tool" {
			root = goRoot
		}
		name := filepath.Join(root, filepath.FromSlash(file))
		if err := discoveryCPPRegularSource(root, name); err != nil {
			return err
		}
		actual, err := discoveryFeatureFileSHA256(name)
		if err != nil || actual != source.SHA256 {
			return fmt.Errorf("consumed CPP source changed: %s", id)
		}
	}
	// A previously missing preferred package header can appear without
	// changing the old tool-header bytes. Recheck the entire resolution map,
	// not only hashes of files which existed at the earlier snapshot.
	for _, unit := range inputs.Units {
		for id, source := range inputs.Sources {
			for index, directive := range source.Directives {
				if directive.Kind != "include" {
					continue
				}
				key := id + "#" + strconv.Itoa(index)
				wanted, selected := unit.Includes[key]
				deferredKind := unit.DeferredIncludes[key]
				if !selected && deferredKind == "" {
					continue
				}
				actual, kind, err := discoveryCPPResolveRawInclude(moduleDir, goRoot, unit.File, directive.Include, inputs.Protocol == discoveryDeferredCPPInputsProtocol)
				if err != nil || actual != wanted || kind != deferredKind {
					return fmt.Errorf("CPP include search selection changed: %s:%d", id, directive.Line)
				}
			}
		}
	}
	return nil
}
