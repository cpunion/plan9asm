package gotoolprofile

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
)

// go vet always resolves test packages, including with -tests=false: that
// flag controls analyzers, not cmd/go's package loader. Use the registered
// native vet tool's compilation-unit protocol instead. The unit is derived
// independently from actual ordinary Go exports, never a filtered test unit.
func instrumentationOrdinaryAsmdeclArgs() []string {
	return []string{"tool", "vet", "-asmdecl", "ordinary.cfg"}
}

type instrumentationAsmdeclPackage struct {
	Dir, ImportPath, Export          string
	GoFiles, CompiledGoFiles, SFiles []string
	Imports                          []string
	ImportMap                        map[string]string
	Standard, Incomplete             bool
	Error                            json.RawMessage
	DepsErrors                       []json.RawMessage
	Module                           *struct {
		Path, Version, GoMod, GoVersion string
		Main                            bool
		Replace                         json.RawMessage
	}
}

// This is cmd/go's native vet configuration protocol. In particular, neither
// VetxOnly nor SucceedOnTypecheckFailure may be set: type/declaration errors
// must fail. A configuration is a query input, not an ABI or source capability.
type instrumentationAsmdeclConfig struct {
	ID, Compiler, Dir, ImportPath, GoVersion string
	GoFiles, NonGoFiles                      []string
	ImportMap, PackageFile                   map[string]string
	Standard                                 map[string]bool
	VetxOutput                               string
}

// RunInstrumentationOrdinaryAsmdecl checks only the independently selected
// ordinary source role. Supplemental bodyless test declarations retain their
// own actual test-export/typecheck consumer; they are not admitted here.
// The caller must authenticate the original module, selection and exports
// before supplying their expected identities and recheck its outer guards.
func RunInstrumentationOrdinaryAsmdecl(ctx context.Context, binary, invocation string, env []string,
	input *ConsumerInput, plan *InstrumentationOriginalPackagePlan, selected *PackageProof,
	exports map[string]string, run Runner,
) error {
	if ctx == nil || ctx.Err() != nil || run == nil || input == nil || plan == nil || selected == nil ||
		!filepath.IsAbs(binary) || !filepath.IsAbs(invocation) || len(exports) == 0 || len(exports) > 4096 ||
		plan.PackagePath == "" || selected.PackagePath != plan.PackagePath {
		return fmt.Errorf("ordinary asmdecl requires actual bounded source/export identities")
	}
	args := instrumentationOriginalCommand(plan, false)
	args = append([]string{"list", "-deps"}, args[1:]...)
	stdout, stderr, err := run(ctx, invocation, env, binary, args...)
	if err != nil || len(stderr) != 0 || len(stdout) == 0 || len(stdout) > commandOutputLimit {
		return fmt.Errorf("ordinary asmdecl export query failed (not source N/A): %v; %s", err, stderr)
	}
	config := instrumentationAsmdeclConfig{
		ID: plan.PackagePath, Compiler: "gc", ImportPath: plan.PackagePath,
		ImportMap: make(map[string]string), PackageFile: make(map[string]string), Standard: make(map[string]bool),
	}
	guards, seen := make(map[string]string), make(map[string]bool)
	decoder := json.NewDecoder(bytes.NewReader(stdout))
	var root *instrumentationAsmdeclPackage
	for {
		var pkg instrumentationAsmdeclPackage
		if err := decoder.Decode(&pkg); err == io.EOF {
			break
		} else if err != nil {
			return fmt.Errorf("actual ordinary asmdecl export JSON: %w", err)
		}
		if pkg.ImportPath == "" || seen[pkg.ImportPath] || len(seen) >= 4096 || pkg.Incomplete ||
			len(pkg.Error) != 0 && string(pkg.Error) != "null" || len(pkg.DepsErrors) != 0 {
			return fmt.Errorf("ordinary asmdecl export has duplicate, incomplete or failed package")
		}
		seen[pkg.ImportPath] = true
		config.ImportMap[pkg.ImportPath] = pkg.ImportPath
		config.Standard[pkg.ImportPath] = pkg.Standard
		if pkg.ImportPath == "unsafe" {
			continue
		}
		if !filepath.IsAbs(pkg.Export) || exports[pkg.ImportPath] == "" {
			return fmt.Errorf("ordinary asmdecl export lacks independent identity: %s", pkg.ImportPath)
		}
		hash, err := FileSHA256(pkg.Export)
		if err != nil || hash != exports[pkg.ImportPath] {
			return fmt.Errorf("ordinary asmdecl export differs from independent consumption: %s", pkg.ImportPath)
		}
		guards[pkg.Export], config.PackageFile[pkg.ImportPath] = hash, pkg.Export
		if pkg.ImportPath == plan.PackagePath {
			root = &pkg
		}
	}
	if root == nil || len(config.PackageFile) != len(exports) {
		return fmt.Errorf("ordinary asmdecl actual export closure differs from its independently consumed denominator")
	}
	if err := bindInstrumentationAsmdeclSource(input, plan, selected, root, &config, guards); err != nil {
		return err
	}
	for name, resolved := range root.ImportMap {
		if !seen[resolved] {
			return fmt.Errorf("ordinary asmdecl has an unconsumed import mapping: %s", name)
		}
		config.ImportMap[name] = resolved
	}
	for _, imported := range root.Imports {
		if !seen[imported] {
			return fmt.Errorf("ordinary asmdecl has an unconsumed actual import: %s", imported)
		}
	}
	owned, err := os.MkdirTemp("", "plan9asm-ordinary-asmdecl-")
	if err != nil {
		return err
	}
	identity, err := os.Stat(owned)
	if err != nil {
		return err
	}
	defer func() {
		if actual, err := os.Lstat(owned); err == nil && os.SameFile(identity, actual) {
			_ = os.RemoveAll(owned)
		}
	}()
	config.VetxOutput = filepath.Join(owned, "ordinary.vetx")
	encoded, err := json.Marshal(config)
	if err != nil {
		return err
	}
	filename := filepath.Join(owned, "ordinary.cfg")
	if err := os.WriteFile(filename, encoded, 0600); err != nil {
		return err
	}
	guards[filename] = bytesSHA256(encoded)
	if err := recheckInstrumentationAsmdeclGuards(guards); err != nil {
		return err
	}
	stdout, stderr, err = run(ctx, owned, env, binary, instrumentationOrdinaryAsmdeclArgs()...)
	if err != nil || len(stdout) != 0 || len(stderr) != 0 {
		return fmt.Errorf("original ordinary same-mode asmdecl failed or emitted unknown diagnostics (not source N/A): %v; %s; %s", err, stdout, stderr)
	}
	return recheckInstrumentationAsmdeclGuards(guards)
}

func bindInstrumentationAsmdeclSource(input *ConsumerInput, plan *InstrumentationOriginalPackagePlan, selected *PackageProof,
	pkg *instrumentationAsmdeclPackage, config *instrumentationAsmdeclConfig, guards map[string]string,
) error {
	expectedDir := filepath.Join(input.SourceRoot, filepath.FromSlash(plan.PackageDirectory))
	expectedDir, err := filepath.EvalSymlinks(expectedDir)
	if err != nil {
		return err
	}
	actualDir, err := filepath.EvalSymlinks(pkg.Dir)
	if err != nil || actualDir != expectedDir || pkg.Module == nil || pkg.Module.Path != selected.ModulePath ||
		pkg.Module.Version != selected.ModuleVersion || pkg.Module.Main != (selected.SourceRole == "main") ||
		len(pkg.Module.Replace) != 0 && string(pkg.Module.Replace) != "null" {
		return fmt.Errorf("ordinary asmdecl package/module identity changed")
	}
	config.Dir = actualDir
	config.GoVersion = "go" + pkg.Module.GoVersion
	if pkg.Module.GoVersion == "" {
		config.GoVersion = "go1.16"
	}
	metadataHash := input.Sources["go.mod"]
	if input.ProxyGoMod != nil {
		metadataHash = input.ProxyGoMod.SHA256
	}
	if hash, err := FileSHA256(pkg.Module.GoMod); err != nil || hash != metadataHash || metadataHash == "" {
		return fmt.Errorf("ordinary asmdecl original module metadata changed")
	}
	guards[pkg.Module.GoMod] = metadataHash
	for index, names := range [][]string{pkg.GoFiles, pkg.CompiledGoFiles, pkg.SFiles} {
		var logical, actual []string
		for _, name := range names {
			if filepath.IsAbs(name) {
				name, err = filepath.Rel(pkg.Dir, name)
			}
			if err != nil || path.Base(name) != name || !validInstrumentationSourcePath(name) || strings.HasSuffix(name, "_test.go") {
				return fmt.Errorf("ordinary asmdecl source lacks an exact original role")
			}
			source := path.Join(plan.PackageDirectory, name)
			if selected.SourceSHA256[source] == "" || selected.SourceSHA256[source] != input.Sources[source] {
				return fmt.Errorf("ordinary asmdecl source differs from independently selected original bytes")
			}
			filename := filepath.Join(actualDir, name)
			guards[filename] = selected.SourceSHA256[source]
			logical, actual = append(logical, source), append(actual, filename)
		}
		sort.Strings(logical)
		expected := [][]string{selected.GoFiles, selected.CompiledGoFiles, selected.SFiles}[index]
		if !reflect.DeepEqual(logical, expected) {
			return fmt.Errorf("ordinary asmdecl source selection changed")
		}
		if index == 1 {
			config.GoFiles = actual
		} else if index == 2 {
			config.NonGoFiles = actual
		}
	}
	return nil
}

func recheckInstrumentationAsmdeclGuards(guards map[string]string) error {
	for filename, expected := range guards {
		if actual, err := FileSHA256(filename); err != nil || actual != expected {
			return fmt.Errorf("ordinary asmdecl source/configuration/export changed: %s", filename)
		}
	}
	return nil
}
