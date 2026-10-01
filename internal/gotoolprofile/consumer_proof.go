package gotoolprofile

import (
	"fmt"
	"go/build"
	"io"
	"path"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
)

// ValidateSelection checks the compact compiler-consumption evidence against
// the producer's original source/profile inputs. Actual reobservation and ZIP
// byte comparison happen in production, not in this offline replay.
func ValidateSelection(input *ConsumerInput, proof *SelectionProof, tags []string, requireObject bool) error {
	return ValidateSelectionWithABI(input, proof, tags, requireObject, nil)
}

// Only a caller which independently validated exact target ABI exclusions may
// permit their missing LLVM outputs. Default consumers never grant this path.
func ValidateSelectionWithABI(input *ConsumerInput, proof *SelectionProof, tags []string, requireObject bool, abi map[string]bool) error {
	if input == nil || input.Protocol != ConsumerProtocol || input.ID != ProfileID(input.Observed) {
		return fmt.Errorf("missing canonical compiler feature input")
	}
	if err := Validate(input.Observed); err != nil {
		return err
	}
	if proof == nil || proof.Protocol != ConsumerProtocol || proof.ProfileID != input.ID || !equalDiscoveryStrings(proof.CustomTags, tags) || len(proof.Packages) == 0 {
		return fmt.Errorf("compiler consumption differs from the source-required profile/tag scope")
	}
	selectedFiles, filePackages, err := validatePackageSelection(input, proof.Packages, tags)
	if err != nil {
		return err
	}
	if input.GeneratedHeaders != nil || proof.GeneratedHeaders != nil {
		if err := ValidateGeneratedHeaderAgreement(input, input.GeneratedHeaders, proof.GeneratedHeaders, tags); err != nil {
			return err
		}
		if !reflect.DeepEqual(proof.Packages, proof.GeneratedHeaders.Packages) {
			return fmt.Errorf("actual generated header was compiled from a different package selection")
		}
	}
	return validateAssemblyConsumption(input, proof, selectedFiles, filePackages, requireObject, abi)
}

// Metadata-only queries reuse source selection without manufacturing CPP or
// LLVM outputs. Their separate protocol can never satisfy ValidateSelection.
func validatePackageSelection(input *ConsumerInput, packages []PackageProof, tags []string) (map[string]bool, map[string]PackageProof, error) {
	selectedFiles, seenPackages := make(map[string]bool), make(map[string]bool)
	filePackages := make(map[string]PackageProof)
	minor, err := goMinor(input.Observed.GoVersion)
	if err != nil {
		return nil, nil, err
	}
	ctx := build.Default
	ctx.GOOS, ctx.GOARCH = input.Observed.Environment["GOOS"], input.Observed.Environment["GOARCH"]
	ctx.Compiler, ctx.CgoEnabled, ctx.BuildTags = "gc", false, tags
	ctx.ToolTags, ctx.ReleaseTags = input.Observed.ToolTags, nil
	for version := 1; version <= minor; version++ {
		ctx.ReleaseTags = append(ctx.ReleaseTags, fmt.Sprintf("go1.%d", version))
	}
	ctx.OpenFile = func(name string) (io.ReadCloser, error) {
		header, found := input.Headers[filepath.ToSlash(name)]
		if !found {
			return nil, fmt.Errorf("actual Go file has no original selection header: %s", name)
		}
		return io.NopCloser(strings.NewReader(header)), nil
	}
	for _, pkg := range packages {
		if err := ValidatePackageModule(input, pkg); err != nil {
			return nil, nil, err
		}
		if pkg.PackagePath != input.Module && !strings.HasPrefix(pkg.PackagePath, input.Module+"/") || seenPackages[pkg.PackagePath] {
			return nil, nil, fmt.Errorf("actual selected package has a different or duplicate module role")
		}
		seenPackages[pkg.PackagePath] = true
		if err := ValidateOrdinaryAssemblerMacros(pkg.Macros, input.Observed, pkg.PackagePath); err != nil {
			return nil, nil, err
		}
		if len(pkg.GoFiles) == 0 || len(pkg.CompiledGoFiles) == 0 || len(pkg.SFiles) == 0 {
			return nil, nil, fmt.Errorf("compiler consumption lacks actual ordinary Go/ASM selection")
		}
		usedSources := make(map[string]string)
		for _, files := range [][]string{pkg.GoFiles, pkg.CompiledGoFiles, pkg.SFiles} {
			if !sort.StringsAreSorted(files) {
				return nil, nil, fmt.Errorf("noncanonical actual selected source order")
			}
			for index, file := range files {
				if index > 0 && file == files[index-1] || input.Sources[file] == "" || input.Sources[file] != pkg.SourceSHA256[file] {
					return nil, nil, fmt.Errorf("actual selected source has no exact original SHA: %s", file)
				}
				usedSources[file] = pkg.SourceSHA256[file]
				wantedPath := input.Module
				if path.Dir(file) != "." {
					wantedPath += "/" + path.Dir(file)
				}
				if pkg.PackagePath != wantedPath {
					return nil, nil, fmt.Errorf("actual package role differs from the original source directory: %s", file)
				}
				if path.Ext(file) == ".go" || path.Ext(file) == ".s" {
					selected, err := ctx.MatchFile(path.Dir(file), path.Base(file))
					if err != nil || !selected || strings.HasSuffix(file, "_test.go") {
						return nil, nil, fmt.Errorf("actual source selection contradicts its original feature constraints: %s: %v", file, err)
					}
				}
			}
		}
		if !reflect.DeepEqual(usedSources, pkg.SourceSHA256) {
			return nil, nil, fmt.Errorf("actual package source hash map has missing/extra consumption")
		}
		for _, file := range pkg.SFiles {
			if selectedFiles[file] || path.Ext(file) != ".s" {
				return nil, nil, fmt.Errorf("duplicate/non-ASM actual selected file")
			}
			selectedFiles[file] = true
			filePackages[file] = pkg
		}
	}
	return selectedFiles, filePackages, nil
}

func validateAssemblyConsumption(input *ConsumerInput, proof *SelectionProof, selectedFiles map[string]bool, filePackages map[string]PackageProof, requireObject bool, abi map[string]bool) error {
	consumed, outputs := make(map[string]bool), make(map[string]bool)
	for _, cpp := range proof.CPP {
		if consumed[cpp.File] || !selectedFiles[cpp.File] || !discoverySHA256Pattern.MatchString(cpp.ExpandedSHA256) || !discoverySHA256Pattern.MatchString(cpp.TypedExpandedSHA256) || len(cpp.Inputs) == 0 {
			return fmt.Errorf("missing, duplicate or unselected actual CPP consumption")
		}
		consumed[cpp.File] = true
		for id, digest := range cpp.Inputs {
			origin, file, present := strings.Cut(id, "/")
			wanted := input.Sources[file]
			if origin == "tool" {
				wanted = input.ToolSources[file]
			}
			if origin == "generated" {
				wanted = generatedCPPSourceSHA(proof.GeneratedHeaders, filePackages[cpp.File].PackagePath, file)
			}
			if !present || origin != "module" && origin != "tool" && origin != "generated" || wanted == "" || digest != wanted {
				return fmt.Errorf("actual CPP source differs from exact original input: %s", id)
			}
		}
		if cpp.Inputs["module/"+cpp.File] != input.Sources[cpp.File] {
			return fmt.Errorf("actual CPP proof omits its assembly source")
		}
		if err := ValidateEmptyAssembly(input, cpp, filePackages[cpp.File]); err != nil {
			return err
		}
	}
	seenParts := make(map[string]bool)
	for _, output := range proof.Outputs {
		key := output.File + "\x00" + output.Part
		if !consumed[output.File] || abi[output.File] || seenParts[key] || path.Base(output.Part) != output.Part || !strings.HasSuffix(output.Part, ".ll") || !discoverySHA256Pattern.MatchString(output.IR) || requireObject && !discoverySHA256Pattern.MatchString(output.Object) {
			return fmt.Errorf("missing, duplicate or unselected LLVM IR/object consumption")
		}
		seenParts[key], outputs[output.File] = true, true
	}
	for _, file := range input.AsmFiles {
		if !selectedFiles[file] || !consumed[file] || !outputs[file] && !abi[file] {
			return fmt.Errorf("source-required file lacks actual package/CPP/LLVM consumption: %s", file)
		}
	}
	if len(consumed) != len(input.AsmFiles) || len(outputs)+len(abi) != len(input.AsmFiles) {
		return fmt.Errorf("compiler consumption differs from the exact file/profile scope")
	}
	return nil
}

func ValidatePackageModule(input *ConsumerInput, pkg PackageProof) error {
	if input == nil || pkg.ModulePath != input.Module || pkg.ModuleVersion != input.Version {
		return fmt.Errorf("actual Go package module/version differs from the explicit consumer input")
	}
	if metadata := input.ProxyGoMod; metadata != nil {
		if err := ValidateProxyGoMod(metadata); err != nil {
			return err
		}
		if input.Sources["go.mod"] != "" || input.Module != metadata.Module || input.SourceModule != metadata.Module || input.Version != metadata.Version ||
			pkg.GoModOrigin != metadata.Protocol || pkg.GoModSHA256 != metadata.SHA256 || pkg.GoModSum != metadata.GoModSum {
			return fmt.Errorf("actual package lacks the independent authenticated proxy metadata origin")
		}
	} else if pkg.GoModOrigin != "" || pkg.GoModSHA256 != "" || pkg.GoModSum != "" {
		return fmt.Errorf("package proxy metadata has no independently authenticated source proof")
	}
	module := input.SourceModule
	if module == "" {
		module = input.Module
	}
	if pkg.SourceModule != module || pkg.SourceVersion != input.Version {
		return fmt.Errorf("actual Go replacement source module/version differs from exact original inputs")
	}
	if input.SourceModule != "" {
		if pkg.SourceRole != "module" && pkg.SourceRole != "version_replace" {
			return fmt.Errorf("original exact-module proof cannot be relabeled from main/owned-local sources")
		}
	} else if (pkg.SourceRole != "main" && pkg.SourceRole != "owned_local_replace") ||
		(pkg.SourceRole == "main" && input.Version != "") {
		return fmt.Errorf("explicit own-source consumer lacks an exact main/local role")
	}
	return nil
}
