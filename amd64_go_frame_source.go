package plan9asm

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/xgo-dev/plan9asm/internal/goabi"
	"github.com/xgo-dev/plan9asm/internal/goasmcmd"
	"golang.org/x/arch/x86/x86asm"
)

// GoAMD64FrameCaptureRequest describes actual input locations, not permissions.
type GoAMD64FrameCaptureRequest struct {
	GoBinary, Directory, AssemblyFile, Target string
	Environment                               []string
}

// GoAMD64FrameSource is a sealed actual-object observation. It does not grant
// a native backing, caller region, memory access or return certificate.
type GoAMD64FrameSource struct {
	mu        sync.Mutex
	ready     bool
	request   GoAMD64FrameCaptureRequest
	selection goAMD64FrameSelection
	env       map[string]string
	inputs    map[string][32]byte
	entries   []string
	owned     string
	functions map[string]goAMD64FrameFunction
	origins   map[string]int
}

type goAMD64FrameFunction struct {
	objectSymbol, listingSymbol string
	locals, args                int64
	code                        []byte
	positions                   map[int]string
	signature                   FuncSig
}

// This is decoded only from the actual driver stdout, never from a proof file.
type goAMD64FrameSelection struct {
	Dir, ImportPath, Name, Export, ForTest                                                 string
	GoFiles, CompiledGoFiles, SFiles, HFiles, CgoFiles, Imports, EmbedFiles, EmbedPatterns []string
	IgnoredGoFiles, IgnoredOtherFiles                                                      []string
	CFiles, CXXFiles, MFiles, FFiles, SwigFiles, SwigCXXFiles, SysoFiles                   []string
	Module                                                                                 *struct {
		Path, Version, Dir, GoMod string
		Main                      bool
		Replace                   json.RawMessage
	}
	Error      json.RawMessage
	DepsErrors []json.RawMessage
}

func (selected goAMD64FrameSelection) nonordinaryRole() bool {
	for _, files := range [][]string{selected.CgoFiles, selected.CFiles, selected.CXXFiles, selected.MFiles,
		selected.FFiles, selected.SwigFiles, selected.SwigCXXFiles, selected.SysoFiles, selected.EmbedFiles, selected.EmbedPatterns} {
		if len(files) != 0 {
			return true
		}
	}
	return false
}

func CaptureGoAMD64FrameSource(ctx context.Context, pkg GoPackage, request GoAMD64FrameCaptureRequest) (_ *GoAMD64FrameSource, captureErr error) {
	if ctx == nil || ctx.Err() != nil || pkg.Types == nil || pkg.Path != pkg.Types.Path() ||
		!filepath.IsAbs(request.GoBinary) || !filepath.IsAbs(request.Directory) ||
		filepath.Base(request.AssemblyFile) != request.AssemblyFile || filepath.Ext(request.AssemblyFile) != ".s" {
		return nil, goAMD64FrameContext("requires live context, actual package and bounded absolute inputs")
	}
	parts := strings.Split(request.Target, "/")
	if len(parts) != 2 || parts[1] != "amd64" || parts[0] != "linux" && parts[0] != "darwin" && parts[0] != "windows" {
		return nil, goAMD64FrameContext("initial object capture supports ordinary AMD64 targets only")
	}
	var err error
	request.Directory, err = filepath.EvalSymlinks(request.Directory)
	if err != nil {
		return nil, err
	}
	source := &GoAMD64FrameSource{
		request: request,
		inputs:  make(map[string][32]byte),
		origins: make(map[string]int),
	}
	source.request.Environment = append([]string(nil), request.Environment...)
	provided := make(map[string]string)
	for _, value := range request.Environment {
		key, setting, valid := strings.Cut(value, "=")
		if !valid || key == "" {
			return nil, goAMD64FrameContext("malformed command environment")
		}
		provided[key] = setting
		if (key == "GOFLAGS" || key == "GOEXPERIMENT") && setting != "" || key == "GOPACKAGESDRIVER" && setting != "" && setting != "off" {
			return nil, goAMD64FrameContext("effective overlay, custom flags, experiments or driver require their own source mapping")
		}
	}
	for _, key := range []string{"GOPATH", "GOMODCACHE", "GOCACHE", "TMPDIR", "GOTMPDIR"} {
		if !filepath.IsAbs(provided[key]) {
			return nil, goAMD64FrameContext("explicit private %s directory is required", key)
		}
	}
	settings := map[string]string{"GOOS": parts[0], "GOARCH": "amd64", "CGO_ENABLED": "0", "GOTOOLCHAIN": "local", "GOENV": "off", "GOWORK": "off", "GOFLAGS": "", "GOPROXY": "off"}
	var env []string
	for _, value := range source.request.Environment {
		key, _, _ := strings.Cut(value, "=")
		if _, replace := settings[key]; !replace {
			env = append(env, value)
		}
	}
	for key, value := range settings {
		env = append(env, key+"="+value)
	}
	source.request.Environment = env
	if err := source.snapshotDirectory(); err != nil {
		return nil, err
	}
	if err := source.observe(ctx); err != nil {
		return nil, err
	}
	selected, err := source.load(ctx)
	if err != nil {
		return nil, err
	}
	source.selection = selected
	if selected.Dir != request.Directory || selected.ImportPath != pkg.Path || selected.ForTest != "" || selected.Module == nil ||
		!selected.Module.Main || selected.Module.Version != "" || selected.Module.Dir != request.Directory || len(selected.Module.Replace) != 0 && string(selected.Module.Replace) != "null" ||
		selected.nonordinaryRole() || len(selected.Imports) != 0 || len(selected.GoFiles) == 0 || len(selected.SFiles) == 0 {
		return nil, goAMD64FrameContext("initial capture needs complete dependency-free ordinary original module sources")
	}
	if selected.Module.GoMod != filepath.Join(request.Directory, "go.mod") {
		return nil, goAMD64FrameContext("actual module metadata is outside the original source directory")
	}
	if err := source.numericDeclarations(pkg); err != nil {
		return nil, err
	}
	source.owned, err = os.MkdirTemp("", "plan9asm-go-frame-object-")
	if err != nil {
		return nil, err
	}
	defer func() {
		if captureErr != nil {
			captureErr = errors.Join(captureErr, source.Close())
		}
	}()
	canonicalOwned, err := filepath.EvalSymlinks(source.owned)
	if err != nil {
		return nil, err
	}
	source.owned = canonicalOwned
	source.functions = make(map[string]goAMD64FrameFunction)
	found := false
	for _, name := range selected.SFiles {
		found = found || name == request.AssemblyFile
		original, err := source.read(filepath.Join(request.Directory, name))
		if err != nil {
			return nil, err
		}
		expanded, err := source.preprocess(name, original)
		if err != nil {
			return nil, err
		}
		file, err := ParseWithDefines(ArchAMD64, expanded, nil)
		if err != nil || len(file.Funcs) == 0 || len(file.Data)+len(file.Globl)+len(file.UnlinkedPrelude) != 0 {
			return nil, goAMD64FrameContext("full original assembly lacks complete TEXT: %v", err)
		}
		object := filepath.Join(source.owned, name+".o")
		listing, err := source.run(ctx, "tool", "asm", "-S", "-p", pkg.Path, "-I", request.Directory, "-I", filepath.Join(source.env["GOROOT"], "pkg", "include"), "-o", object, name)
		if err != nil {
			return nil, err
		}
		if data, err := source.read(object); err != nil || len(data) == 0 {
			if err == nil {
				err = goAMD64FrameContext("actual assembler produced an empty object")
			}
			return nil, err
		}
		if err := os.WriteFile(object+".listing", listing, 0600); err != nil {
			return nil, err
		}
		if _, err := source.read(object + ".listing"); err != nil {
			return nil, err
		}
		functions, err := goAMD64FrameListing(listing, file, pkg, source.origins)
		if err != nil {
			return nil, err
		}
		if err := source.verifyObjectListing(ctx, object, functions); err != nil {
			return nil, err
		}
		for symbol, fn := range functions {
			if _, duplicate := source.functions[symbol]; duplicate {
				return nil, goAMD64FrameContext("duplicate original TEXT symbol")
			}
			source.functions[symbol] = fn
		}
	}
	if !found {
		return nil, goAMD64FrameContext("requested assembly is absent from actual complete selection")
	}
	source.ready = true
	if err := source.Verify(ctx); err != nil {
		return nil, err
	}
	return source, nil
}

// Verify rechecks actual routing/selection and all original/tool/export/artifact
// bytes. Failure permanently revokes this observation; it grants no memory.
func (source *GoAMD64FrameSource) Verify(ctx context.Context) (verifyErr error) {
	if source == nil {
		return goAMD64FrameContext("missing sealed actual-object observation")
	}
	source.mu.Lock()
	defer source.mu.Unlock()
	defer func() {
		if verifyErr != nil {
			source.ready = false
		}
	}()
	if !source.ready || ctx == nil || ctx.Err() != nil {
		return goAMD64FrameContext("closed, uninitialized or expired actual-object observation")
	}
	before := source.env
	if err := source.observe(ctx); err != nil {
		return err
	}
	if !reflect.DeepEqual(before, source.env) {
		return goAMD64FrameContext("actual Go environment changed")
	}
	if err := source.verifyInputs(); err != nil {
		return err
	}
	selected, err := source.load(ctx)
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(source.selection, selected) {
		return goAMD64FrameContext("actual complete package selection/export changed")
	}
	return source.verifyInputs()
}

func (source *GoAMD64FrameSource) verifyInputs() error {
	entries, err := os.ReadDir(source.request.Directory)
	if err != nil {
		return err
	}
	var names []string
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	if !reflect.DeepEqual(names, source.entries) {
		return goAMD64FrameContext("original package directory inventory changed")
	}
	for name, expected := range source.inputs {
		data, err := goAMD64FrameRead(name)
		if err != nil || sha256.Sum256(data) != expected {
			return goAMD64FrameContext("original/tool/export/object input changed: %s: %v", filepath.Base(name), err)
		}
	}
	return nil
}

// Close invalidates the token before removing only its factory-owned artifacts.
// Calling it again is harmless; no source, tool or cache-owned file is removed.
func (source *GoAMD64FrameSource) Close() error {
	if source == nil {
		return nil
	}
	source.mu.Lock()
	defer source.mu.Unlock()
	source.ready = false
	if source.owned == "" {
		return nil
	}
	err := os.RemoveAll(source.owned)
	if err == nil {
		source.owned = ""
	}
	return err
}

func goAMD64FrameContext(format string, args ...any) error {
	return fmt.Errorf("%w: actual Go frame source: %s", ErrProbeNeedsContext, fmt.Sprintf(format, args...))
}

func (source *GoAMD64FrameSource) run(ctx context.Context, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()
	return goasmcmd.Run(ctx, source.request.Directory, source.request.Environment, source.request.GoBinary, args...)
}

func goAMD64FrameRead(name string) (_ []byte, readErr error) {
	resolved, err := filepath.EvalSymlinks(name)
	if err != nil || resolved != name {
		return nil, goAMD64FrameContext("input path is redirected or inaccessible: %s", filepath.Base(name))
	}
	file, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer func() { readErr = errors.Join(readErr, file.Close()) }()
	info, err := file.Stat()
	if err != nil || !info.Mode().IsRegular() || info.Size() > 64<<20 {
		return nil, goAMD64FrameContext("input is not a bounded regular file")
	}
	data, err := io.ReadAll(io.LimitReader(file, 64<<20+1))
	if err != nil || len(data) > 64<<20 {
		return nil, goAMD64FrameContext("input byte bound/read failure: %v", err)
	}
	return data, nil
}

func (source *GoAMD64FrameSource) read(name string) ([]byte, error) {
	data, err := goAMD64FrameRead(name)
	if err != nil {
		return nil, err
	}
	digest := sha256.Sum256(data)
	if before, exists := source.inputs[name]; exists && before != digest {
		return nil, goAMD64FrameContext("input changed during capture: %s", filepath.Base(name))
	}
	source.inputs[name] = digest
	return data, nil
}

func (source *GoAMD64FrameSource) snapshotDirectory() error {
	resolved, err := filepath.EvalSymlinks(source.request.Directory)
	if err != nil || resolved != source.request.Directory {
		return goAMD64FrameContext("requires canonical actual package directory")
	}
	entries, err := os.ReadDir(resolved)
	if err != nil || len(entries) == 0 || len(entries) > 256 {
		return goAMD64FrameContext("package directory is unavailable or exceeds its bound")
	}
	var total int
	for _, entry := range entries {
		if entry.IsDir() {
			return goAMD64FrameContext("initial source root requires a flat complete inventory")
		}
		source.entries = append(source.entries, entry.Name())
		data, err := source.read(filepath.Join(resolved, entry.Name()))
		if err != nil {
			return err
		}
		total += len(data)
		if total > 16<<20 {
			return goAMD64FrameContext("original package byte inventory exceeds its bound")
		}
	}
	return nil
}

func (source *GoAMD64FrameSource) observe(ctx context.Context) error {
	if _, err := source.read(source.request.GoBinary); err != nil {
		return err
	}
	data, err := source.run(ctx, "env", "-json", "GOROOT", "GOTOOLDIR", "GOCACHE", "GOMODCACHE", "GOPATH", "GOTMPDIR", "GOVERSION", "GOOS", "GOARCH", "GOFLAGS", "GOEXPERIMENT", "GOAMD64")
	if err != nil {
		return err
	}
	var env map[string]string
	if err := json.Unmarshal(data, &env); err != nil {
		return err
	}
	if env["GOOS"]+"/"+env["GOARCH"] != source.request.Target || env["GOFLAGS"] != "" || env["GOEXPERIMENT"] != "" || env["GOAMD64"] != "v1" ||
		filepath.Join(env["GOROOT"], "bin", "go") != source.request.GoBinary {
		return goAMD64FrameContext("actual driver/target/CPU/source routing needs a separate audited contract")
	}
	version, err := source.read(filepath.Join(env["GOROOT"], "VERSION"))
	if err != nil || strings.Split(string(version), "\n")[0] != env["GOVERSION"] {
		return goAMD64FrameContext("actual driver differs from source VERSION")
	}
	for _, tool := range []string{"asm", "compile", "objdump"} {
		data, err := source.run(ctx, "tool", "-n", tool)
		if err != nil {
			return err
		}
		name := strings.TrimSpace(string(data))
		if !filepath.IsAbs(name) || strings.ContainsAny(name, "\r\n") || tool != "objdump" && name != filepath.Join(env["GOTOOLDIR"], tool) {
			return goAMD64FrameContext("unrecognized actual %s route", tool)
		}
		resolved, err := filepath.EvalSymlinks(name)
		if err != nil {
			return err
		}
		if tool == "objdump" && name != filepath.Join(env["GOTOOLDIR"], tool) {
			cache, err := filepath.EvalSymlinks(env["GOCACHE"])
			if err != nil || !goAssemblyPathWithinRoot(cache, resolved) || filepath.Base(resolved) != tool {
				return goAMD64FrameContext("unknown builtin object observer origin")
			}
		}
		name = resolved
		env["route/"+tool] = name
		if _, err := source.read(name); err != nil {
			return err
		}
	}
	for _, name := range []string{"src/cmd/internal/obj/x86/obj6.go", "src/cmd/internal/obj/x86/asm6.go", "src/cmd/go/internal/work/gc.go", "src/cmd/asm/main.go", "src/cmd/go/internal/tool/tool.go", "src/cmd/internal/objabi/stack.go"} {
		if _, err := source.read(filepath.Join(env["GOROOT"], filepath.FromSlash(name))); err != nil {
			return err
		}
	}
	source.env = env
	return nil
}

func (source *GoAMD64FrameSource) load(ctx context.Context) (goAMD64FrameSelection, error) {
	var selected goAMD64FrameSelection
	data, err := source.run(ctx, "list", "-json", "-compiled", "-export", "-mod=readonly", ".")
	if err != nil {
		return selected, err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(&selected); err != nil {
		return selected, err
	}
	var extra any
	if decoder.Decode(&extra) != io.EOF || len(selected.Error) != 0 || len(selected.DepsErrors) != 0 {
		return selected, goAMD64FrameContext("incomplete or ambiguous actual package export")
	}
	if selected.Export == "" {
		return selected, goAMD64FrameContext("actual package has no compiled export")
	}
	selected.Export, err = filepath.EvalSymlinks(selected.Export)
	cache, cacheErr := filepath.EvalSymlinks(source.env["GOCACHE"])
	if err != nil || cacheErr != nil || !goAssemblyPathWithinRoot(cache, selected.Export) {
		return selected, goAMD64FrameContext("compiled export is outside the actual build cache")
	}
	if data, err := source.read(selected.Export); err != nil || len(data) == 0 {
		if err == nil {
			err = goAMD64FrameContext("actual package export is empty")
		}
		return selected, err
	}
	return selected, nil
}

func (source *GoAMD64FrameSource) numericDeclarations(pkg GoPackage) error {
	fset := token.NewFileSet()
	var syntax []*ast.File
	var compiled []string
	for _, name := range source.selection.GoFiles {
		filename := filepath.Join(source.request.Directory, name)
		data, err := source.read(filename)
		if err != nil {
			return err
		}
		file, err := parser.ParseFile(fset, filename, data, parser.ParseComments)
		if err != nil {
			return err
		}
		for _, declaration := range file.Decls {
			fn, ok := declaration.(*ast.FuncDecl)
			if !ok || fn.Body != nil || fn.Recv != nil {
				return goAMD64FrameContext("initial declaration must be a bodyless numeric function")
			}
		}
		syntax = append(syntax, file)
		compiled = append(compiled, filename)
	}
	sort.Strings(compiled)
	actual := append([]string(nil), source.selection.CompiledGoFiles...)
	for i, name := range actual {
		if !filepath.IsAbs(name) {
			actual[i] = filepath.Join(source.request.Directory, name)
		}
	}
	sort.Strings(actual)
	if !reflect.DeepEqual(compiled, actual) {
		return goAMD64FrameContext("effective/generated compiled source mapping is unbound")
	}
	sizes := types.SizesFor("gc", "amd64")
	typed, err := (&types.Config{Sizes: sizes}).Check(pkg.Path, fset, syntax, nil)
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(typed.Scope().Names(), pkg.Types.Scope().Names()) {
		return goAMD64FrameContext("actual selected declarations differ from supplied types")
	}
	for _, name := range typed.Scope().Names() {
		fn, ok := typed.Scope().Lookup(name).(*types.Func)
		previous, supplied := pkg.Types.Scope().Lookup(name).(*types.Func)
		if !ok || !supplied || !types.Identical(fn.Type(), previous.Type()) {
			return goAMD64FrameContext("missing actual numeric declaration")
		}
		signature := fn.Type().(*types.Signature)
		for _, tuple := range []*types.Tuple{signature.Params(), signature.Results()} {
			for i := 0; i < tuple.Len(); i++ {
				basic, ok := tuple.At(i).Type().(*types.Basic)
				if !ok || basic.Info()&types.IsInteger == 0 {
					return goAMD64FrameContext("initial transport needs declared integer scalars")
				}
			}
		}
		actual, err := DeriveGoFuncABI0Signature(pkg.Path+"."+name, fn, "amd64", sizes)
		prior, priorErr := DeriveGoFuncABI0Signature(pkg.Path+"."+name, previous, "amd64", sizes)
		if err != nil || priorErr != nil || !reflect.DeepEqual(actual, prior) {
			return goAMD64FrameContext("actual numeric declaration/layout mismatch: %s", name)
		}
		for _, ty := range append(append([]LLVMType(nil), actual.Args...), actual.Ret) {
			if ty != I8 && ty != I16 && ty != I32 && ty != I64 && ty != Void {
				return goAMD64FrameContext("initial frame transport excludes pointers/aggregates/floats")
			}
		}
	}
	return nil
}

func (source *GoAMD64FrameSource) preprocess(name string, original []byte) (string, error) {
	source.origins[filepath.Join(source.request.Directory, name)] = bytes.Count(original, []byte("\n")) + 1
	// The owned asm command passes this exact basename from the canonical
	// package directory; stock Go can retain that relative spelling in -S.
	source.origins[name] = source.origins[filepath.Join(source.request.Directory, name)]
	var expand func(string, []byte) (string, error)
	expand = func(name string, data []byte) (string, error) {
		for _, line := range strings.Split(string(data), "\n") {
			if strings.HasPrefix(strings.TrimSpace(line), "#if") || strings.HasPrefix(strings.TrimSpace(line), "#else") {
				return "", goAMD64FrameContext("conditional CPP requires actual profile macro registration")
			}
		}
		return PreprocessAssemblySource(string(data), AssemblyPreprocessOptions{FileName: name, ReadInclude: func(parent, include string) (string, []byte, error) {
			if filepath.Base(include) != include || include == "go_asm.h" {
				return "", nil, goAMD64FrameContext("generated/relative CPP origin requires qualified metadata")
			}
			file := filepath.Join(source.request.Directory, include)
			if _, err := os.Stat(file); os.IsNotExist(err) {
				file = filepath.Join(source.env["GOROOT"], "pkg", "include", include)
			}
			data, err := source.read(file)
			if err != nil {
				return "", nil, err
			}
			source.origins[file] = bytes.Count(data, []byte("\n")) + 1
			for _, line := range strings.Split(string(data), "\n") {
				if strings.HasPrefix(strings.TrimSpace(line), "#if") {
					return "", nil, goAMD64FrameContext("conditional header requires profile registration")
				}
			}
			return file, data, nil
		}})
	}
	return expand(name, original)
}

var goAMD64FrameHeader = regexp.MustCompile(`^(\S+) STEXT nosplit(?: asm)? size=([0-9]+) (?:align=0x[0-9a-f]+ )?args=0x([0-9a-f]+) locals=0x([0-9a-f]+)(?: funcid=0x[0-9a-f]+)?(?: align=0x[0-9a-f]+)?$`)
var goAMD64FrameRow = regexp.MustCompile(`^\s*0x([0-9a-f]+) [0-9]+ \((.+)\)\s+(.+)$`)

func goAMD64FrameListing(listing []byte, file *File, pkg GoPackage, origins map[string]int) (map[string]goAMD64FrameFunction, error) {
	functions := make(map[string]goAMD64FrameFunction)
	scanner := bufio.NewScanner(bytes.NewReader(listing))
	scanner.Buffer(nil, 4<<20)
	name := ""
	var expected int64
	for scanner.Scan() {
		line := scanner.Text()
		if header := goAMD64FrameHeader.FindStringSubmatch(line); header != nil {
			if name != "" && int64(len(functions[name].code)) != expected {
				return nil, goAMD64FrameContext("partial native symbol bytes")
			}
			name = header[1]
			if _, duplicate := functions[name]; duplicate {
				return nil, goAMD64FrameContext("duplicate native TEXT")
			}
			expected, _ = strconv.ParseInt(header[2], 10, 64)
			args, _ := strconv.ParseInt(header[3], 16, 64)
			locals, _ := strconv.ParseInt(header[4], 16, 64)
			functions[name] = goAMD64FrameFunction{locals: locals, args: args, positions: make(map[int]string)}
			continue
		}
		if name == "" {
			return nil, goAMD64FrameContext("native listing contains unbound/unknown symbol")
		}
		fn := functions[name]
		if row := goAMD64FrameRow.FindStringSubmatch(line); row != nil {
			pc, err := strconv.ParseInt(row[1], 16, 32)
			if err != nil || pc < 0 || pc >= expected {
				return nil, goAMD64FrameContext("invalid original instruction/source PC")
			}
			separator := strings.LastIndexByte(row[2], ':')
			if separator < 0 {
				return nil, goAMD64FrameContext("unbound original instruction/source path")
			}
			line, err := strconv.Atoi(row[2][separator+1:])
			if err != nil || line <= 0 || line > origins[row[2][:separator]] {
				return nil, goAMD64FrameContext("unbound original instruction/source line: %s", row[2])
			}
			fn.positions[int(pc)] += row[2] + "\t" + row[3] + "\n"
		} else {
			fields := strings.Fields(line)
			if len(fields) < 2 || !strings.HasPrefix(fields[0], "0x") {
				return nil, goAMD64FrameContext("unknown native listing/relocation record: %s", line)
			}
			pc, err := strconv.ParseInt(strings.TrimPrefix(fields[0], "0x"), 16, 32)
			if err != nil || pc != int64(len(fn.code)) {
				return nil, goAMD64FrameContext("native byte gap/overlap")
			}
			count := 0
			for _, word := range fields[1:] {
				if len(word) != 2 {
					break
				}
				value, err := hex.DecodeString(word)
				if err != nil {
					break
				}
				fn.code = append(fn.code, value[0])
				count++
			}
			if count == 0 || count > 16 || int64(len(fn.code)) > expected {
				return nil, goAMD64FrameContext("invalid native byte record")
			}
		}
		functions[name] = fn
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if name == "" || int64(len(functions[name].code)) != expected || len(functions) != len(file.Funcs) {
		return nil, goAMD64FrameContext("missing/incomplete original TEXT inventory")
	}
	bound := make(map[string]goAMD64FrameFunction)
	for _, original := range file.Funcs {
		symbol := pkg.Path + "." + strings.TrimPrefix(original.Sym, "·")
		observed := symbol
		fn, exists := functions[observed]
		if !exists {
			// Older actual Go objects retain PkgIdxSelf's empty-package
			// spelling. Keep those observed bytes/symbols distinct from the
			// package identity independently selected and type-checked above.
			observed = `"".` + strings.TrimPrefix(original.Sym, "·")
			fn, exists = functions[observed]
		}
		declaration, declared := pkg.Types.Scope().Lookup(strings.TrimPrefix(original.Sym, "·")).(*types.Func)
		if !exists || !declared || original.FrameSize <= 0 || original.FrameSize%8 != 0 || fn.locals != original.FrameSize+8 || fn.args != original.ArgSize || strings.Contains(original.Sym, "<") {
			return nil, goAMD64FrameContext("original ABI0 TEXT/declaration/frame identity mismatch")
		}
		var err error
		fn.signature, err = DeriveGoFuncABI0Signature(symbol, declaration, "amd64", types.SizesFor("gc", "amd64"))
		if err != nil {
			return nil, err
		}
		var end int64
		for _, slot := range append(append([]FrameSlot(nil), fn.signature.Frame.Params...), fn.signature.Frame.Results...) {
			if next := slot.Offset + frameTypeSize(slot.Type, 8); next > end {
				end = next
			}
		}
		if !goabi.MatchesABI0TextSize("amd64", fn.args, end) {
			return nil, goAMD64FrameContext("original TEXT args disagree with actual declaration")
		}
		for _, ins := range original.Instrs {
			switch ins.Op {
			case "BYTE", "WORD", "LONG", "QUAD", OpCALL, "JMP", "PCALIGN":
				return nil, goAMD64FrameContext("raw, calling or address-sensitive original frame requires a later source contract")
			case OpRET:
				if len(ins.Args) != 0 {
					return nil, goAMD64FrameContext("nonordinary original RET transport")
				}
			}
		}
		if err := goAMD64FrameMachineShape(fn, original.FrameSize); err != nil {
			return nil, err
		}
		fn.listingSymbol, fn.objectSymbol = observed, symbol
		bound[symbol] = fn
	}
	return bound, nil
}

func goAMD64FrameMachineShape(fn goAMD64FrameFunction, frame int64) error {
	var instructions []x86asm.Inst
	starts := make(map[int]bool)
	for pc := 0; pc < len(fn.code); {
		ins, err := x86asm.Decode(fn.code[pc:], 64)
		if err != nil || ins.Len == 0 || ins.Op == 0 || ins.Op == x86asm.CALL || ins.Op == x86asm.LCALL {
			return goAMD64FrameContext("unknown or calling original machine instruction")
		}
		if len(fn.positions[pc]) == 0 {
			return goAMD64FrameContext("machine instruction lacks its original source position")
		}
		starts[pc] = true
		instructions = append(instructions, ins)
		pc += ins.Len
	}
	for pc := range fn.positions {
		if !starts[pc] {
			return goAMD64FrameContext("original source position is not a machine instruction boundary")
		}
	}
	if len(instructions) < 6 {
		return goAMD64FrameContext("partial automatic BP/SP frame")
	}
	first, second, third := instructions[0], instructions[1], instructions[2]
	push := first.Op == x86asm.PUSH && first.Args[0] == x86asm.RBP &&
		second.Op == x86asm.MOV && second.Args[0] == x86asm.RBP && second.Args[1] == x86asm.RSP &&
		third.Op == x86asm.SUB && third.Args[0] == x86asm.RSP && third.Args[1] == x86asm.Imm(frame)
	slot := x86asm.Mem{Base: x86asm.RSP, Scale: 1, Disp: frame}
	stored := first.Op == x86asm.SUB && first.Args[0] == x86asm.RSP && first.Args[1] == x86asm.Imm(frame+8) &&
		second.Op == x86asm.MOV && second.Args[0] == slot && second.Args[1] == x86asm.RBP && second.DataSize == 64 &&
		third.Op == x86asm.LEA && third.Args[0] == x86asm.RBP && third.Args[1] == slot && third.DataSize == 64
	if !push && !stored {
		return goAMD64FrameContext("unrecognized automatic BP/SP prologue")
	}
	returns := 0
	for i, ins := range instructions {
		if ins.Op == x86asm.RET {
			returns++
			if i < 2 {
				return goAMD64FrameContext("partial original RET epilogue")
			}
			penultimate, previous := instructions[i-2], instructions[i-1]
			pushReturn := push && previous.Op == x86asm.POP && previous.Args[0] == x86asm.RBP &&
				penultimate.Op == x86asm.ADD && penultimate.Args[0] == x86asm.RSP && penultimate.Args[1] == x86asm.Imm(frame)
			storedReturn := stored && penultimate.Op == x86asm.MOV && penultimate.Args[0] == x86asm.RBP && penultimate.Args[1] == slot && penultimate.DataSize == 64 &&
				previous.Op == x86asm.ADD && previous.Args[0] == x86asm.RSP && previous.Args[1] == x86asm.Imm(frame+8)
			if !pushReturn && !storedReturn {
				return goAMD64FrameContext("unrecognized original RET epilogue")
			}
		}
	}
	if returns == 0 || instructions[len(instructions)-1].Op != x86asm.RET {
		return goAMD64FrameContext("original frame has no conventional RET")
	}
	return nil
}

func (source *GoAMD64FrameSource) verifyObjectListing(ctx context.Context, object string, wanted map[string]goAMD64FrameFunction) error {
	data, err := source.run(ctx, "tool", "objdump", object)
	if err != nil {
		return err
	}
	return goAMD64FrameObjectBytes(data, wanted)
}

func goAMD64FrameObjectBytes(data []byte, wanted map[string]goAMD64FrameFunction) error {
	actual := make(map[string][]byte)
	addresses := make(map[string]uint64)
	name := ""
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		if len(fields) == 3 && fields[0] == "TEXT" && strings.HasSuffix(fields[1], "(SB)") {
			name = strings.TrimSuffix(fields[1], "(SB)")
			if _, duplicate := actual[name]; duplicate {
				return goAMD64FrameContext("duplicate actual object symbol")
			}
			actual[name] = nil
			continue
		}
		if len(fields) < 4 || name == "" || !strings.HasPrefix(fields[1], "0x") {
			return goAMD64FrameContext("unknown actual object disassembly record")
		}
		value, err := hex.DecodeString(fields[2])
		if err != nil || len(value) == 0 {
			return goAMD64FrameContext("invalid actual object bytes")
		}
		address, err := strconv.ParseUint(strings.TrimPrefix(fields[1], "0x"), 16, 64)
		if len(actual[name]) == 0 {
			addresses[name] = address
		}
		if err != nil || address != addresses[name]+uint64(len(actual[name])) {
			return goAMD64FrameContext("actual object instruction bytes have a gap/overlap")
		}
		actual[name] = append(actual[name], value...)
	}
	if len(actual) != len(wanted) {
		return goAMD64FrameContext("object/listing symbol inventories differ")
	}
	for _, fn := range wanted {
		if !bytes.Equal(actual[fn.objectSymbol], fn.code) {
			return goAMD64FrameContext("object/listing physical bytes differ for %s", fn.objectSymbol)
		}
	}
	return nil
}
