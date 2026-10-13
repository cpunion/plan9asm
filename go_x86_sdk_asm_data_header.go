package plan9asm

import (
	"bytes"
	"context"
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"github.com/xgo-dev/plan9asm/internal/goasmheader"
)

type sdkASMDataHeaderObservation struct {
	LanguageVersion string
	Definitions     map[string]string
	Actual, Typed   []byte
	Canonical       []byte
	CompilerObject  []byte
	ImportConfig    []byte
	SymABI          []byte
}

func (source *GoX86SDKASMDataSource) artifact(name string, data []byte) (string, error) {
	file := filepath.Join(source.guard.owned, name)
	if !filepath.IsLocal(name) || !goAssemblyPathWithinRoot(source.guard.owned, file) {
		return "", sdkASMDataContext("artifact escaped the private factory directory")
	}
	if err := os.WriteFile(file, data, 0600); err != nil {
		return "", err
	}
	if _, err := source.guard.read(file); err != nil {
		return "", err
	}
	return file, nil
}

func (source *GoX86SDKASMDataSource) asmArguments(include string) []string {
	args := []string{"tool", "asm", "-p", source.request.PackagePath, "-std", "-I", include,
		"-I", filepath.Join(source.guard.env["GOROOT"], "pkg", "include")}
	if source.guard.env["GOOS"] == "darwin" {
		args = append(args, "-shared")
	}
	for _, define := range GoAssemblerDefinesWithEnv(source.guard.env["GOOS"], source.guard.env["GOARCH"], source.guard.env) {
		args = append(args, "-D", define)
	}
	return args
}

func (source *GoX86SDKASMDataSource) captureHeader(ctx context.Context) error {
	guard := source.guard
	stubDirectory := filepath.Join(guard.owned, "symabis-input")
	if err := os.Mkdir(stubDirectory, 0700); err != nil {
		return err
	}
	if err := source.retainDirectory(stubDirectory); err != nil {
		return err
	}
	if _, err := source.artifact(filepath.Join("symabis-input", "go_asm.h"), nil); err != nil {
		return err
	}
	symabis := filepath.Join(guard.owned, "symabis")
	args := append(source.asmArguments(stubDirectory), "-gensymabis", "-o", symabis)
	for _, name := range guard.selection.SFiles {
		args = append(args, filepath.Join(source.request.Directory, name))
	}
	if _, err := source.run(ctx, args...); err != nil {
		return err
	}
	symABI, err := guard.read(symabis)
	if err != nil {
		return err
	}

	config, err := sdkASMDataImportConfig(guard.selection, guard.imports)
	if err != nil {
		return err
	}
	configFile, err := source.artifact("importcfg", config)
	if err != nil {
		return err
	}
	headerFile, objectFile := filepath.Join(guard.owned, "go_asm.h"), filepath.Join(guard.owned, "go-header.o")
	// Stock SDK packages use the invoked toolchain language, not a caller's
	// go.mod or a self-reported Standard bit. The full actual SDK was bound above.
	language := "go1.27"
	args = []string{"tool", "compile", "-p", source.request.PackagePath, "-std", "-lang=" + language,
		"-goversion", guard.env["GOVERSION"], "-symabis", symabis, "-importcfg", configFile,
		"-asmhdr", headerFile, "-o", objectFile}
	if guard.env["GOOS"] == "darwin" {
		args = append(args, "-shared")
	}
	for _, name := range guard.selection.GoFiles {
		args = append(args, filepath.Join(source.request.Directory, name))
	}
	if _, err := source.run(ctx, args...); err != nil {
		return err
	}
	actual, err := guard.read(headerFile)
	if err != nil {
		return err
	}
	compilerObject, err := guard.read(objectFile)
	if err != nil || len(compilerObject) == 0 {
		return sdkASMDataContext("header compiler did not produce a complete object: %v", err)
	}
	definitions, err := goasmheader.Definitions(actual)
	if err != nil {
		return sdkASMDataContext("actual full generated header is invalid: %v", err)
	}
	typed, err := source.typedHeader(language)
	if err != nil {
		return err
	}
	typedDefinitions, err := goasmheader.Definitions(typed)
	if err != nil || !reflect.DeepEqual(definitions, typedDefinitions) {
		return sdkASMDataContext("complete actual and independently typed header definitions differ: %v", err)
	}
	canonical, err := goasmheader.Canonical(definitions)
	if err != nil {
		return err
	}
	source.observation.Header = sdkASMDataHeaderObservation{LanguageVersion: language, Definitions: definitions,
		Actual: actual, Typed: typed, Canonical: canonical, CompilerObject: compilerObject, ImportConfig: config, SymABI: symABI}
	return nil
}

func sdkASMDataImportConfig(selected goAMD64FrameSelection, imports map[string]goAMD64FrameSelection) ([]byte, error) {
	var config strings.Builder
	paths := append([]string(nil), selected.Imports...)
	sort.Strings(paths)
	for _, path := range paths {
		if path == "unsafe" {
			continue
		}
		mapped := path
		if target, present := selected.ImportMap[path]; present {
			mapped = target
		}
		dependency, present := imports[mapped]
		// cmd/compile's readImportCfg splits only at the first '='. The
		// filename is literal, not a quoted Go string or import path: native
		// Windows separators, embedded spaces and '=' must remain unchanged.
		// TrimSpace applies to the whole line, so a trailing filename space
		// cannot be represented faithfully; newlines/NUL cannot be inputs.
		if !present || path == "" || mapped == "" || strings.ContainsAny(path+mapped, "\x00\r\n=\\ \t") ||
			dependency.Export == "" || strings.ContainsAny(dependency.Export, "\x00\r\n") ||
			strings.TrimSpace(dependency.Export) != dependency.Export {
			return nil, sdkASMDataContext("header compiler lacks an exact actual dependency export")
		}
		if mapped != path {
			fmt.Fprintf(&config, "importmap %s=%s\n", path, mapped)
		}
		fmt.Fprintf(&config, "packagefile %s=%s\n", mapped, dependency.Export)
	}
	return []byte(config.String()), nil
}

func (source *GoX86SDKASMDataSource) typedHeader(language string) ([]byte, error) {
	fset := token.NewFileSet()
	var syntax []*ast.File
	for _, name := range source.guard.selection.GoFiles {
		file := filepath.Join(source.request.Directory, name)
		data, err := source.guard.read(file)
		if err != nil {
			return nil, err
		}
		parsed, err := parser.ParseFile(fset, file, data, parser.ParseComments)
		if err != nil {
			return nil, err
		}
		syntax = append(syntax, parsed)
	}
	actualImporter := importer.ForCompiler(fset, "gc", func(path string) (io.ReadCloser, error) {
		if target, mapped := source.guard.selection.ImportMap[path]; mapped {
			path = target
		}
		dependency, present := source.guard.imports[path]
		if !present || dependency.Export == "" {
			return nil, sdkASMDataContext("typed header lacks its actual sealed SDK export: %s", path)
		}
		data, err := source.guard.read(dependency.Export)
		if err != nil {
			return nil, err
		}
		return io.NopCloser(bytes.NewReader(data)), nil
	})
	sizes := types.SizesFor("gc", source.guard.env["GOARCH"])
	if sizes == nil {
		return nil, sdkASMDataContext("unknown target compiler layout")
	}
	typed, err := (&types.Config{Sizes: sizes, Importer: actualImporter, GoVersion: language}).Check(source.request.PackagePath, fset, syntax, nil)
	if err != nil {
		return nil, sdkASMDataContext("complete original SDK Go sources cannot be independently typed: %v", err)
	}
	return GoAssemblyHeader(GoPackage{Path: source.request.PackagePath, Types: typed}, source.guard.env["GOARCH"])
}
