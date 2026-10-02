package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"io"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"github.com/xgo-dev/plan9asm/internal/gotoolprofile"
)

// Syntax here only avoids unnecessary ordinary-type/CPP queries. No name,
// type, layout or actual missing declaration is inferred by this predicate.
func discoveryTestDeclarationQueryNeeded(plan *discoveryOrdinarySelectionPlan, profile discoveryFeatureProfile, root string, files, tags []string) bool {
	ctx := build.Default
	parts := strings.Split(profile.Observed.Target, "/")
	ctx.GOOS, ctx.GOARCH, ctx.CgoEnabled = parts[0], parts[1], false
	ctx.BuildTags, ctx.ToolTags = tags, profile.Observed.ToolTags
	dirs := make(map[string]bool)
	for _, file := range files {
		dirs[path.Dir(file)] = true
	}
	for _, source := range plan.Sources {
		if !dirs[path.Dir(source.File)] || !strings.HasSuffix(source.File, "_test.go") {
			continue
		}
		filename := filepath.Join(root, filepath.FromSlash(source.File))
		selected, err := ctx.MatchFile(filepath.Dir(filename), filepath.Base(filename))
		if err != nil {
			return true // The real query must report source/tool errors, not this predicate.
		}
		if !selected {
			continue
		}
		syntax, _ := parser.ParseFile(token.NewFileSet(), filename, nil, 0)
		if syntax == nil {
			continue // Unrelated poison tests never affect an adequate ordinary declaration.
		}
		for _, node := range syntax.Decls {
			if decl, ok := node.(*ast.FuncDecl); ok && decl.Body == nil && decl.Recv == nil {
				return true
			}
		}
	}
	return false
}

func captureDiscoveryTestDeclarationQuery(ctx context.Context, plan *discoveryOrdinarySelectionPlan, download moduleDownloadInfo,
	profile discoveryFeatureProfile, tags []string, group discoveryPackageGroup, dir string, env []string, translator string, index int,
) error {
	if !discoveryTestDeclarationQueryNeeded(plan, profile, download.Dir, group.AsmFiles, tags) {
		return nil
	}
	module, err := ordinaryProfileDeclaredModule(plan)
	if err != nil || !filepath.IsAbs(translator) {
		return fmt.Errorf("test-declaration preflight requires actual module/tool identity: %v", err)
	}
	input := ordinaryProfileConsumerInput(plan, profile, module, download.Dir, group.AsmFiles, tags)
	input.TestDependencies = nil
	inputPath := filepath.Join(dir, fmt.Sprintf("test-signature-input-%04d.json", index))
	reportPath := filepath.Join(dir, fmt.Sprintf("test-signature-query-%04d.json", index))
	output := filepath.Join(dir, fmt.Sprintf("test-signature-oracle-%04d", index))
	data, err := json.Marshal(input)
	if err != nil {
		return err
	}
	if err := os.WriteFile(inputPath, data, 0600); err != nil {
		return err
	}
	parts := strings.Split(profile.Observed.Target, "/")
	args := []string{"-signature-source-query", "-goos=" + parts[0], "-goarch=" + parts[1],
		"-module-path=" + module, "-patterns=" + group.Pattern, "-asm-files=" + strings.Join(group.AsmFiles, ","),
		"-feature-profile=" + inputPath, "-out=" + output, "-report=" + reportPath}
	if len(tags) != 0 {
		args = append(args, "-tags="+strings.Join(tags, ","))
	}
	if _, _, err := runDiscoveryMachineCommand(ctx, dir, env, translator, args...); err != nil {
		return err
	}
	data, err = os.ReadFile(reportPath)
	if err != nil || len(data) > 16<<20 {
		return fmt.Errorf("read bounded missing-signature query: %v", err)
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var query gotoolprofile.SignatureQueryProof
	if err := decoder.Decode(&query); err != nil {
		return err
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return fmt.Errorf("signature query contains trailing input")
	}
	if err := gotoolprofile.ValidateSignatureQuery(input, &query, tags); err != nil {
		return err
	}
	if len(query.Requests) == 0 {
		return nil // No test load, regardless of unrelated poison test bodies/imports.
	}
	if len(plan.TestDeclarationQueries) >= 4096 {
		return fmt.Errorf("test declaration queries exceed their explicit bounded scope")
	}
	dependency, err := captureDiscoveryTestDependencySources(ctx, plan, download, input, &query, dir, env, tags, group.Pattern)
	if err != nil {
		return err
	}
	plan.TestDeclarationQueries = append(plan.TestDeclarationQueries, discoveryTestDeclarationQuery{
		Target: profile.Observed.Target, ProfileID: profile.ID, BuildTags: append([]string(nil), tags...),
		AsmFiles: append([]string(nil), group.AsmFiles...), Dependency: dependency,
	})
	input = ordinaryProfileConsumerInput(plan, profile, module, download.Dir, group.AsmFiles, tags)
	if err := gotoolprofile.ValidateTestDependencyPreload(input, dependency, tags); err != nil {
		return err
	}
	return verifyDiscoveryTestDependencySources(plan, download.Dir, download.Zip)
}

func captureDiscoveryTestDependencySources(ctx context.Context, plan *discoveryOrdinarySelectionPlan, download moduleDownloadInfo,
	input *gotoolprofile.ConsumerInput, query *gotoolprofile.SignatureQueryProof, dir string, env, tags []string, pattern string,
) (*gotoolprofile.TestDependencyProof, error) {
	args := []string{"list", "-deps", "-test", "-compiled", "-e", "-json", "-mod=mod"}
	if len(tags) != 0 {
		args = append(args, "-tags="+strings.Join(tags, ","))
	}
	args = append(args, pattern)
	data, _, err := runDiscoveryMachineCommand(ctx, dir, env, "go", args...)
	if err != nil {
		return nil, err
	}
	canonicalRoot, err := filepath.EvalSymlinks(download.Dir)
	if err != nil {
		return nil, err
	}
	type selectedPackage struct {
		ImportPath, Name, ForTest, Dir            string
		GoFiles, CompiledGoFiles, SFiles, Imports []string
		ImportMap                                 map[string]string
		Module                                    *struct {
			Path, Version, Dir string
			Replace            *struct{ Path, Version, Dir string }
		}
		Error      *struct{ Err string }
		DepsErrors []struct{ Err string }
	}
	selected := make(map[string]selectedPackage)
	decoder := json.NewDecoder(bytes.NewReader(data))
	for count := 0; ; count++ {
		var pkg selectedPackage
		if err := decoder.Decode(&pkg); err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
		if count >= 4096 || ctx.Err() != nil || pkg.Error != nil || len(pkg.DepsErrors) != 0 {
			return nil, fmt.Errorf("actual test dependency selection failed or exceeded its bounded scope: %s: %#v %#v", pkg.ImportPath, pkg.Error, pkg.DepsErrors)
		}
		if pkg.Module == nil || pkg.Module.Dir == "" {
			continue
		}
		root, err := filepath.EvalSymlinks(pkg.Module.Dir)
		if err != nil {
			return nil, err
		}
		if root != canonicalRoot {
			continue
		}
		if pkg.Module.Path != input.Module || pkg.Module.Version != input.Version {
			return nil, fmt.Errorf("test dependency uses different actual module/version/replacement source: %s: %#v", pkg.ImportPath, pkg.Module)
		}
		original := query.Selection.Packages[0]
		if replacement := pkg.Module.Replace; replacement != nil {
			if original.SourceRole != "version_replace" || replacement.Path != original.SourceModule || replacement.Version != original.SourceVersion {
				return nil, fmt.Errorf("test dependency changed the exact original replacement role")
			}
			replacedRoot, err := filepath.EvalSymlinks(replacement.Dir)
			if err != nil || replacedRoot != canonicalRoot {
				return nil, fmt.Errorf("test dependency replacement does not use the authenticated original ZIP root: %v", err)
			}
		} else if original.SourceRole != "module" || original.SourceModule != pkg.Module.Path || original.SourceVersion != pkg.Module.Version {
			return nil, fmt.Errorf("test dependency changed the actual exact-module source role")
		}
		if _, duplicate := selected[pkg.ImportPath]; duplicate {
			return nil, fmt.Errorf("actual Go repeated a test dependency package identity")
		}
		selected[pkg.ImportPath] = pkg
	}
	dependency := &gotoolprofile.TestDependencyProof{Protocol: gotoolprofile.TestDependencyProtocol, Query: query}
	roots := make(map[string]bool)
	for _, request := range query.Requests {
		for _, pkg := range query.Selection.Packages {
			if containsTargetFeature(pkg.SFiles, request.File) {
				roots[pkg.PackagePath] = true
			}
		}
	}
	visited := make(map[string]bool)
	var visit func(string, bool) error
	visit = func(packagePath string, test bool) error {
		if visited[packagePath] {
			return nil
		}
		id := packagePath
		if test {
			id += " [" + packagePath + ".test]"
		}
		actual, found := selected[id]
		if !found || test && actual.ForTest != packagePath || !test && actual.ForTest != "" {
			return fmt.Errorf("actual Go omitted the required same-module test dependency: %s", id)
		}
		visited[packagePath] = true
		if err := captureDiscoveryTestSourceDirectory(plan, canonicalRoot, actual.Dir); err != nil {
			return err
		}
		// Source roles come from the ordinary package whose actual Go replacement
		// identity was independently captured in preflight, not invented metadata.
		pkg := query.Selection.Packages[0]
		pkg.PackagePath, pkg.Macros, pkg.SourceSHA256 = packagePath, nil, make(map[string]string)
		capture := func(files []string) ([]string, error) {
			var result []string
			for _, filename := range files {
				if !filepath.IsAbs(filename) {
					filename = filepath.Join(actual.Dir, filename)
				}
				filename, err := filepath.EvalSymlinks(filename)
				if err != nil {
					return nil, err
				}
				if err := discoveryCPPRegularSource(canonicalRoot, filename); err != nil {
					return nil, err
				}
				name, err := filepath.Rel(canonicalRoot, filename)
				name = filepath.ToSlash(name)
				if err != nil || !ordinarySelectionLocalPath(name) {
					return nil, fmt.Errorf("test dependency selected an unbound original source")
				}
				digest, err := discoveryFeatureFileSHA256(filename)
				if err != nil {
					return nil, err
				}
				pkg.SourceSHA256[name] = digest
				result = append(result, name)
			}
			sort.Strings(result)
			return result, nil
		}
		var err error
		if pkg.GoFiles, err = capture(actual.GoFiles); err != nil {
			return err
		}
		if pkg.CompiledGoFiles, err = capture(actual.CompiledGoFiles); err != nil {
			return err
		}
		if pkg.SFiles, err = capture(actual.SFiles); err != nil {
			return err
		}
		if !reflect.DeepEqual(pkg.GoFiles, pkg.CompiledGoFiles) {
			return fmt.Errorf("test dependency generated or replaced original Go source")
		}
		imports := make(map[string]string)
		for _, destination := range actual.Imports {
			matched := false
			for original, target := range actual.ImportMap {
				if target == destination {
					imports[original] = destination
					matched = true
				}
			}
			if !matched {
				imports[destination] = destination
			}
			if destination == input.Module || strings.HasPrefix(destination, input.Module+"/") {
				if err := visit(destination, false); err != nil {
					return err
				}
			}
		}
		dependency.Packages = append(dependency.Packages, gotoolprofile.TestSourcePackage{Package: pkg, VariantID: id, ImportMap: imports})
		return nil
	}
	for root := range roots {
		if err := visit(root, true); err != nil {
			return nil, err
		}
	}
	sort.Slice(dependency.Packages, func(i, j int) bool {
		return dependency.Packages[i].Package.PackagePath < dependency.Packages[j].Package.PackagePath
	})
	return dependency, nil
}

func captureDiscoveryTestSourceDirectory(plan *discoveryOrdinarySelectionPlan, root, directory string) error {
	directory, err := filepath.EvalSymlinks(directory)
	if err != nil {
		return err
	}
	name, err := filepath.Rel(root, directory)
	name = filepath.ToSlash(name)
	if err != nil || name != "." && !ordinarySelectionLocalPath(name) {
		return fmt.Errorf("test dependency directory lies outside the original ZIP")
	}
	if name != "." {
		if err := captureDiscoveryTestSourceDirectory(plan, root, filepath.Dir(directory)); err != nil {
			return err
		}
	}
	entries, err := os.ReadDir(directory)
	if err != nil {
		return err
	}
	captured := ordinarySelectionDirectory{Directory: name, Entries: []ordinarySelectionEntry{}}
	var sources []ordinarySelectionSource
	for _, entry := range entries {
		kind, digest := "directory", ""
		if !entry.IsDir() {
			if !entry.Type().IsRegular() {
				return fmt.Errorf("test dependency contains a non-regular original entry")
			}
			kind = "file"
			filename := filepath.Join(directory, entry.Name())
			digest, err = discoveryFeatureFileSHA256(filename)
			if err != nil {
				return err
			}
			file := path.Join(name, entry.Name())
			if strings.HasSuffix(file, ".go") || strings.HasSuffix(file, ".s") {
				data, err := os.ReadFile(filename)
				if err != nil || len(data) > 64<<20 || len(sources) >= 32768 {
					return fmt.Errorf("test dependency source exceeds its bounded inventory: %v", err)
				}
				original, err := nativeLayoutSourceInputFromBytes(file, data)
				if err != nil {
					return err
				}
				header := compactOrdinarySelectionHeader(original.Header)
				sources = append(sources, ordinarySelectionSource{
					File: file, SHA256: digest, Header: header, HeaderSHA256: discoveryFeatureBytesSHA256([]byte(header)),
				})
			}
		}
		captured.Entries = append(captured.Entries, ordinarySelectionEntry{Name: entry.Name(), Kind: kind, SHA256: digest})
	}
	// An ordinary ancestor already has a complete directory/SHA witness, but
	// may not have any Go selection headers: it need not contain assembly.
	// Recheck that witness and supplement headers without duplicating or
	// replacing its ordinary directory/source role.
	found := false
	for _, existing := range append(append([]ordinarySelectionDirectory(nil), plan.Directories...), plan.TestDeclarationDirectories...) {
		if existing.Directory == name {
			if found || !reflect.DeepEqual(existing, captured) {
				return fmt.Errorf("test dependency original directory membership/bytes changed")
			}
			found = true
		}
	}
	knownSources := make(map[string]ordinarySelectionSource)
	for _, source := range append(append([]ordinarySelectionSource(nil), plan.Sources...), plan.TestDeclarationSources...) {
		if _, duplicate := knownSources[source.File]; duplicate {
			return fmt.Errorf("test dependency has duplicate original source roles")
		}
		knownSources[source.File] = source
	}
	var added []ordinarySelectionSource
	for _, source := range sources {
		if existing, exists := knownSources[source.File]; exists {
			if existing != source {
				return fmt.Errorf("test dependency original source/header changed")
			}
		} else {
			added = append(added, source)
		}
	}
	if len(plan.TestDeclarationSources)+len(added) > 32768 || !found && len(plan.TestDeclarationDirectories) >= 4096 {
		return fmt.Errorf("test dependency source/directory exceeds its bounded inventory")
	}
	if !found {
		plan.TestDeclarationDirectories = append(plan.TestDeclarationDirectories, captured)
	}
	plan.TestDeclarationSources = append(plan.TestDeclarationSources, added...)
	sort.Slice(plan.TestDeclarationDirectories, func(i, j int) bool {
		return plan.TestDeclarationDirectories[i].Directory < plan.TestDeclarationDirectories[j].Directory
	})
	sort.Slice(plan.TestDeclarationSources, func(i, j int) bool { return plan.TestDeclarationSources[i].File < plan.TestDeclarationSources[j].File })
	return nil
}

func verifyDiscoveryTestDependencySources(plan *discoveryOrdinarySelectionPlan, root, zipPath string) error {
	if err := verifyDiscoveryTestDependencySnapshot(plan, root); err != nil {
		return err
	}
	copyPlan := *plan
	copyPlan.Sources = plan.TestDeclarationSources
	copyPlan.Directories = append(append([]ordinarySelectionDirectory(nil), plan.Directories...), plan.TestDeclarationDirectories...)
	before, err := discoveryFeatureFileSHA256(zipPath)
	if err != nil || before != plan.ZipSHA256 {
		return fmt.Errorf("test dependency original ZIP changed before source membership verification")
	}
	if err := verifyOrdinarySelectionZIP(&copyPlan, zipPath, plan.Module, plan.Version, plan.ModuleSum); err != nil {
		return err
	}
	after, err := discoveryFeatureFileSHA256(zipPath)
	if err != nil || before != after || copyPlan.ZipSHA256 != plan.ZipSHA256 || copyPlan.ModuleSum != plan.ModuleSum {
		return fmt.Errorf("test dependency original ZIP changed during source membership verification")
	}
	return nil
}

func verifyDiscoveryTestDependencySnapshot(plan *discoveryOrdinarySelectionPlan, root string) error {
	// Ordinary selection already guards its original inputs separately. Only
	// recheck directories contributing this supplemental closure here, rather
	// than hashing every unrelated assembly package again for each command.
	relevant := make(map[string]bool)
	for _, source := range plan.TestDeclarationSources {
		relevant[path.Dir(source.File)] = true
	}
	for _, dir := range plan.TestDeclarationDirectories {
		relevant[dir.Directory] = true
	}
	for _, dir := range append(append([]ordinarySelectionDirectory(nil), plan.Directories...), plan.TestDeclarationDirectories...) {
		if !relevant[dir.Directory] {
			continue
		}
		entries, err := os.ReadDir(filepath.Join(root, filepath.FromSlash(dir.Directory)))
		if err != nil || len(entries) != len(dir.Entries) {
			return fmt.Errorf("test dependency original directory membership changed: %v", err)
		}
		for index, entry := range entries {
			wanted := dir.Entries[index]
			if entry.Name() != wanted.Name || entry.IsDir() != (wanted.Kind == "directory") {
				return fmt.Errorf("test dependency original directory member changed")
			}
			if wanted.Kind == "file" {
				file := filepath.Join(root, filepath.FromSlash(dir.Directory), entry.Name())
				actual, err := discoveryFeatureFileSHA256(file)
				if err != nil || actual != wanted.SHA256 || !entry.Type().IsRegular() {
					return fmt.Errorf("test dependency original source bytes changed: %s: %v", file, err)
				}
			}
		}
	}
	return nil
}
