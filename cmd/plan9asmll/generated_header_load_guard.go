package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/xgo-dev/plan9asm/internal/gotoolprofile"
	"golang.org/x/tools/go/packages"
)

// Machine paths stay in this private invocation guard. Portable metadata keeps
// normalized package/module identities and dependency source/export digests.
type generatedHeaderLoadGuard struct {
	Packages map[string]string
	Sources  map[string]string
}

type generatedHeaderLoadSelection struct {
	Name            string
	PackagePath     string
	Module          *packages.Module
	GoFiles         []string
	CompiledGoFiles []string
	OtherFiles      []string
	Imports         map[string]string
}

func captureGeneratedHeaderLoadGuard(ctx context.Context, pkgs []*packages.Package) (*generatedHeaderLoadGuard, error) {
	guard := &generatedHeaderLoadGuard{Packages: make(map[string]string), Sources: make(map[string]string)}
	var captureErr error
	packages.Visit(pkgs, func(pkg *packages.Package) bool {
		if captureErr != nil {
			return false
		}
		if ctx == nil || ctx.Err() != nil {
			captureErr = fmt.Errorf("generated-header source guard requires a live candidate context")
			return false
		}
		if len(pkg.Errors) != 0 {
			var diagnostics []string
			for _, err := range pkg.Errors {
				diagnostics = append(diagnostics, err.Error())
			}
			captureErr = fmt.Errorf("actual generated-header Go source selection:\n%s", strings.Join(diagnostics, "\n"))
			return false
		}
		if pkg.PkgPath == "unsafe" {
			return false // Go's builtin singleton has no source or export file.
		}
		if pkg.ID == "" || len(guard.Packages) >= 4096 || len(pkg.EmbedFiles) != 0 || len(pkg.EmbedPatterns) != 0 || isTestVariantPackage(pkg) {
			captureErr = fmt.Errorf("generated-header query has an unbound embed/test role or exceeds its package bound")
			return false
		}
		canonical := func(files []string) []string {
			result := append([]string(nil), files...)
			sort.Strings(result)
			return result
		}
		selection := generatedHeaderLoadSelection{
			Name: pkg.Name, PackagePath: pkg.PkgPath, Module: pkg.Module,
			GoFiles: canonical(pkg.GoFiles), CompiledGoFiles: canonical(pkg.CompiledGoFiles),
			OtherFiles: canonical(pkg.OtherFiles), Imports: make(map[string]string),
		}
		for name, dep := range pkg.Imports {
			selection.Imports[name] = dep.PkgPath
		}
		data, err := json.Marshal(selection)
		if err != nil {
			captureErr = err
			return false
		}
		guard.Packages[pkg.ID] = featureBytesSHA256(data)
		for _, file := range pkg.CompiledGoFiles {
			if ctx.Err() != nil || len(guard.Sources) >= 32768 {
				captureErr = fmt.Errorf("generated-header source guard exceeded its deadline/file bound")
				return false
			}
			info, err := os.Stat(file)
			if err != nil || !info.Mode().IsRegular() || info.Size() > 64<<20 {
				captureErr = fmt.Errorf("generated-header dependency source must be a bounded regular file")
				return false
			}
			digest, err := gotoolprofile.FileSHA256(file)
			if err != nil {
				captureErr = err
				return false
			}
			guard.Sources[file] = digest
		}
		return true
	}, nil)
	return guard, captureErr
}
