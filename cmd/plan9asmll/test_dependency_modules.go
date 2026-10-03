package main

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"github.com/xgo-dev/plan9asm/internal/gotoolprofile"
	"golang.org/x/mod/module"
	"golang.org/x/tools/go/packages"
)

type testDependencyModuleGuard struct {
	Source gotoolprofile.TestModuleSource
	Root   string
	ZIP    string
}

func (consumer *featureConsumer) foreignTestSourcePackage(actual *packages.Package) (gotoolprofile.TestSourcePackage, error) {
	proof := gotoolprofile.PackageProof{PackagePath: actual.PkgPath, SourceSHA256: make(map[string]string)}
	var root, id, goMod string
	if actual.Module == nil {
		if consumer.Root == "" {
			return gotoolprofile.TestSourcePackage{}, fmt.Errorf("test dependency lacks an actual Go source/tool root")
		}
		var err error
		root, err = filepath.EvalSymlinks(filepath.Join(consumer.Root, "src"))
		if err != nil {
			return gotoolprofile.TestSourcePackage{}, err
		}
		// Root is obtained from the reobserved driver, never from an input
		// JSON path or a caller-supplied stdlib@GoVersion source directory.
		proof.SourceRole, id = "stdlib", "stdlib@"+consumer.Observed.GoVersion
	} else {
		metadata := actual.Module
		if metadata.Main || module.Check(metadata.Path, metadata.Version) != nil || metadata.Path == consumer.Input.Module ||
			actual.PkgPath != metadata.Path && !strings.HasPrefix(actual.PkgPath, metadata.Path+"/") {
			return gotoolprofile.TestSourcePackage{}, fmt.Errorf("foreign test dependency has a different actual module role")
		}
		var err error
		root, err = filepath.EvalSymlinks(metadata.Dir)
		if err != nil || !filepath.IsAbs(root) {
			return gotoolprofile.TestSourcePackage{}, fmt.Errorf("foreign test dependency has an inaccessible actual module directory")
		}
		proof.ModulePath, proof.ModuleVersion = metadata.Path, metadata.Version
		proof.SourceModule, proof.SourceVersion, proof.SourceRole = metadata.Path, metadata.Version, "module"
		goMod = metadata.GoMod
		if replacement := metadata.Replace; replacement != nil {
			if module.Check(replacement.Path, replacement.Version) != nil {
				return gotoolprofile.TestSourcePackage{}, fmt.Errorf("foreign test dependency has an unsupported local/unknown replacement")
			}
			replacedRoot, err := filepath.EvalSymlinks(replacement.Dir)
			if err != nil || replacedRoot != root {
				return gotoolprofile.TestSourcePackage{}, fmt.Errorf("foreign test replacement has a different actual source root")
			}
			proof.SourceModule, proof.SourceVersion, proof.SourceRole = replacement.Path, replacement.Version, "version_replace"
			goMod = replacement.GoMod
		}
		id = proof.SourceModule + "@" + proof.SourceVersion
	}
	var source *gotoolprofile.TestModuleSource
	for index := range consumer.Input.TestDependencies.SourceModules {
		candidate := &consumer.Input.TestDependencies.SourceModules[index]
		if candidate.ID() == id {
			source = candidate
		}
	}
	if source == nil {
		return gotoolprofile.TestSourcePackage{}, fmt.Errorf("actual test dependency lacks its registered foreign source root: %s", actual.PkgPath)
	}
	if err := consumer.bindTestModuleSource(*source, root, goMod); err != nil {
		return gotoolprofile.TestSourcePackage{}, err
	}
	capture := func(files []string, assemblyOnly bool) ([]string, error) {
		var names []string
		for _, filename := range files {
			if assemblyOnly && filepath.Ext(filename) != ".s" {
				continue
			}
			filename, err := filepath.EvalSymlinks(filename)
			if err != nil {
				return nil, err
			}
			name, err := filepath.Rel(root, filename)
			if err != nil {
				return nil, err
			}
			name = filepath.ToSlash(name)
			if _, err := gotoolprofile.CPPRegularSource(root, name); err != nil {
				return nil, err
			}
			digest, err := gotoolprofile.FileSHA256(filename)
			if err != nil || digest != source.Sources[name] {
				return nil, fmt.Errorf("actual test dependency selected an unregistered original source")
			}
			proof.SourceSHA256[name] = digest
			names = append(names, name)
		}
		sort.Strings(names)
		return names, nil
	}
	var err error
	if proof.GoFiles, err = capture(actual.GoFiles, false); err != nil {
		return gotoolprofile.TestSourcePackage{}, err
	}
	if proof.CompiledGoFiles, err = capture(actual.CompiledGoFiles, false); err != nil {
		return gotoolprofile.TestSourcePackage{}, err
	}
	if proof.SFiles, err = capture(actual.OtherFiles, true); err != nil {
		return gotoolprofile.TestSourcePackage{}, err
	}
	builtin := ""
	if actual.PkgPath == "unsafe" {
		if actual.ID != "unsafe" || actual.Name != "unsafe" || actual.Module != nil || actual.ExportFile != "" ||
			len(actual.Imports) != 0 || len(proof.GoFiles) != 1 || proof.GoFiles[0] != "unsafe/unsafe.go" ||
			len(proof.CompiledGoFiles) != 0 || len(proof.SFiles) != 0 {
			return gotoolprofile.TestSourcePackage{}, fmt.Errorf("test dependency changed actual Go unsafe typechecker/documentation selection")
		}
		builtin = "go_typechecker_unsafe_v1"
	}
	if builtin == "" && !reflect.DeepEqual(proof.GoFiles, proof.CompiledGoFiles) {
		return gotoolprofile.TestSourcePackage{}, fmt.Errorf("foreign test dependency uses unproven generated/overlay/cgo source")
	}
	export := ""
	if builtin == "" {
		export, err = gotoolprofile.FileSHA256(actual.ExportFile)
		if err != nil {
			return gotoolprofile.TestSourcePackage{}, fmt.Errorf("foreign test dependency lacks its actual same-profile export: %w", err)
		}
		if consumer.GeneratedGuard == nil {
			consumer.GeneratedGuard = make(map[string]string)
		}
		if before, exists := consumer.GeneratedGuard[actual.ExportFile]; exists && before != export {
			return gotoolprofile.TestSourcePackage{}, fmt.Errorf("foreign test dependency export changed during actual loading")
		}
		consumer.GeneratedGuard[actual.ExportFile] = export
	}
	imports := make(map[string]string)
	for spelling, imported := range actual.Imports {
		imports[spelling] = imported.PkgPath
	}
	return gotoolprofile.TestSourcePackage{
		Package: proof, VariantID: actual.ID, ImportMap: imports, SourceRootID: id, ExportSHA256: export,
		Builtin: builtin,
	}, nil
}

func (consumer *featureConsumer) bindTestModuleSource(source gotoolprofile.TestModuleSource, root, metadata string) error {
	if consumer.TestModuleGuards == nil {
		consumer.TestModuleGuards = make(map[string]testDependencyModuleGuard)
	}
	if prior, found := consumer.TestModuleGuards[source.ID()]; found {
		if prior.Root != root || !reflect.DeepEqual(prior.Source, source) {
			return fmt.Errorf("actual dependency source root/identity changed during loading")
		}
		if metadata != "" {
			actual, err := gotoolprofile.FileSHA256(metadata)
			if err != nil || actual != source.GoModSHA256 {
				return fmt.Errorf("actual dependency GoMod metadata changed during loading")
			}
		}
		return nil // The complete roots/ZIPs are rechecked at each final guard.
	}
	guard := testDependencyModuleGuard{Source: source, Root: root}
	if source.SourceModule != "" {
		binary, err := exec.LookPath("go")
		if err != nil {
			return err
		}
		data, _, err := gotoolprofile.RunBounded(consumer.Context, consumer.WorkDir, consumer.Env, binary,
			"mod", "download", "-json", source.ID())
		if err != nil {
			return err
		}
		var download struct {
			Path, Version, Dir, Zip, GoMod, Sum, GoModSum, Error string
		}
		if err := json.Unmarshal(data, &download); err != nil || download.Error != "" ||
			download.Path != source.SourceModule || download.Version != source.SourceVersion ||
			download.Sum != source.ModuleSum || download.GoModSum != source.GoModSum {
			return fmt.Errorf("actual foreign dependency download differs from exact pre-load ZIP/mod identity: %v", err)
		}
		actualRoot, err := filepath.EvalSymlinks(download.Dir)
		if err != nil || actualRoot != root {
			return fmt.Errorf("actual foreign dependency download has a different module directory")
		}
		actualMod, err := gotoolprofile.FileSHA256(metadata)
		proxyMod, proxyErr := gotoolprofile.FileSHA256(download.GoMod)
		if err != nil || proxyErr != nil || actualMod != source.GoModSHA256 || proxyMod != source.GoModSHA256 {
			return fmt.Errorf("actual foreign dependency GoMod differs from original ZIP and proxy metadata")
		}
		if consumer.GeneratedGuard == nil {
			consumer.GeneratedGuard = make(map[string]string)
		}
		for _, filename := range []string{metadata, download.GoMod} {
			if prior, exists := consumer.GeneratedGuard[filename]; exists && prior != source.GoModSHA256 {
				return fmt.Errorf("actual foreign dependency metadata conflicts with earlier source identity")
			}
			consumer.GeneratedGuard[filename] = source.GoModSHA256
		}
		guard.ZIP = download.Zip
	} else {
		actualRoot, err := filepath.EvalSymlinks(filepath.Join(consumer.Root, "src"))
		if err != nil || actualRoot != root || source.GoVersion != consumer.Observed.GoVersion {
			return fmt.Errorf("stdlib dependency must use the actual reobserved Go source/tool root")
		}
	}
	if err := gotoolprofile.VerifyTestModuleSource(consumer.Context, source, root, guard.ZIP); err != nil {
		return fmt.Errorf("actual test dependency source root %s: %w", source.ID(), err)
	}
	consumer.TestModuleGuards[source.ID()] = guard
	return nil
}

func (consumer *featureConsumer) verifyTestModuleSources() error {
	for _, guard := range consumer.TestModuleGuards {
		if err := gotoolprofile.VerifyTestModuleSource(consumer.Context, guard.Source, guard.Root, guard.ZIP); err != nil {
			return fmt.Errorf("final test dependency source root %s: %w", guard.Source.ID(), err)
		}
	}
	return nil
}
