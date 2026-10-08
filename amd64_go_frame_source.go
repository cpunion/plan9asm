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
	"go/importer"
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
	"github.com/xgo-dev/plan9asm/internal/gooriginalsrc"
	"github.com/xgo-dev/plan9asm/internal/gostaticstorage"
	"golang.org/x/arch/x86/x86asm"
)

// GoAMD64FrameCaptureRequest describes actual input locations, not permissions.
type GoAMD64FrameCaptureRequest struct {
	GoBinary, Directory, AssemblyFile, Target string
	Environment                               []string
	// NumericalReads joins one already authenticated original numerical-read
	// source. It observes relocations; it grants no additional memory role.
	NumericalReads *GoX86NumericalReadBinding
}

// GoAMD64FrameSource is a sealed actual-object observation. It does not grant
// a native backing, caller region, memory access or return certificate.
type GoAMD64FrameSource struct {
	mu                 sync.Mutex
	identity           *GoAMD64FrameSource
	ready              bool
	request            GoAMD64FrameCaptureRequest
	selection          goAMD64FrameSelection
	env                map[string]string
	inputs             map[string][32]byte
	entries            []string
	directoryInfo      os.FileInfo
	owned              string
	ownedInfo          os.FileInfo
	functions          map[string]goAMD64FrameFunction
	declarations       map[string]goAMD64FrameDeclaration
	imports            map[string]goAMD64FrameSelection
	origins            map[string]int
	downloadedOrigin   *gooriginalsrc.OriginSet
	downloadedModule   *gooriginalsrc.Module
	numericalBinding   *GoX86NumericalReadBinding
	numericalOrigin    *GoAMD64FrameSource
	numericalSeal      [32]byte
	numericalPackage   GoPackage
	numericalCapturing bool
}

// newGoOriginalInputGuard initializes only private input/cleanup ownership.
// It has no ready object, declaration, source-origin or frame capability.
func newGoOriginalInputGuard(request GoAMD64FrameCaptureRequest) *GoAMD64FrameSource {
	request.Environment = append([]string(nil), request.Environment...)
	source := &GoAMD64FrameSource{request: request,
		inputs: make(map[string][32]byte), origins: make(map[string]int)}
	source.identity = source
	return source
}

// Transport is an observation role, never a memory or alias permission.
type goAMD64FrameTransport uint8

const (
	goAMD64FrameNumericTransport goAMD64FrameTransport = iota + 1
	goAMD64FrameCompositeTransport
)

type goAMD64FrameDeclaration struct {
	signature    FuncSig
	transport    goAMD64FrameTransport
	languageType string // go/types model identity, not a native pointee extent
}

type goAMD64FrameFunction struct {
	objectSymbol, listingSymbol string
	locals, args                int64
	code                        []byte
	positions                   map[int]string
	signature                   FuncSig
	transport                   goAMD64FrameTransport
	languageType                string
}

// This is decoded only from the actual driver stdout, never from a proof file.
type goAMD64FrameSelection struct {
	Dir, ImportPath, Name, Export, ForTest                                                 string
	GoFiles, CompiledGoFiles, SFiles, HFiles, CgoFiles, Imports, EmbedFiles, EmbedPatterns []string
	IgnoredGoFiles, IgnoredOtherFiles                                                      []string
	Standard                                                                               bool
	ImportMap                                                                              map[string]string
	CFiles, CXXFiles, MFiles, FFiles, SwigFiles, SwigCXXFiles, SysoFiles                   []string
	Module                                                                                 *struct {
		Path, Version, Dir, GoMod, GoVersion string
		Main                                 bool
		Replace                              json.RawMessage
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

func CaptureGoAMD64FrameSource(ctx context.Context, pkg GoPackage, request GoAMD64FrameCaptureRequest) (*GoAMD64FrameSource, error) {
	return captureGoAMD64FrameSource(ctx, pkg, request, nil)
}

func captureGoAMD64FrameSource(ctx context.Context, pkg GoPackage, request GoAMD64FrameCaptureRequest, exact *goAMD64DownloadedIdentity) (_ *GoAMD64FrameSource, captureErr error) {
	if ctx != nil && ctx.Err() != nil {
		if request.NumericalReads != nil {
			return nil, fmt.Errorf("capture numerical frame context: %w", ctx.Err())
		}
		return nil, errors.Join(goAMD64FrameContext("requires live capture context"), ctx.Err())
	}
	if ctx == nil || pkg.Types == nil || pkg.Path != pkg.Types.Path() ||
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
	source := newGoOriginalInputGuard(request)
	if request.NumericalReads != nil {
		if err := verifyGoAMD64FrameNumericalRequest(ctx, pkg, source.request); err != nil {
			return nil, err
		}
		source.numericalBinding = request.NumericalReads
		source.numericalOrigin = source
		source.numericalSeal = request.NumericalReads.seal
		source.numericalPackage = pkg
		source.numericalCapturing = true
	}
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
	if exact == nil {
		if err := source.verifyOwnedSourceDirectory(); err != nil {
			return nil, err
		}
	} else {
		inside, err := gostaticstorage.SourceDirectoryWithinRoot(request.Directory, source.env["GOROOT"])
		if err != nil || inside {
			return nil, errors.Join(goAMD64FrameContext("SDK source cannot acquire an exact downloaded role"), err)
		}
		if err := source.captureDownloadedOrigin(ctx, *exact, provided["GOTMPDIR"]); err != nil {
			return nil, err
		}
	}
	selected, err := source.load(ctx)
	if err != nil {
		return nil, err
	}
	source.selection = selected
	if source.numericalBinding != nil && (len(selected.SFiles) != 1 || selected.SFiles[0] != request.AssemblyFile) {
		return nil, goAMD64FrameContext("numerical frame join requires one complete original assembly selection")
	}
	if selected.Dir != request.Directory || selected.ImportPath != pkg.Path || selected.ForTest != "" || selected.Module == nil ||
		!selected.Module.Main || selected.Module.Version != "" || selected.Module.Dir != request.Directory || len(selected.Module.Replace) != 0 && string(selected.Module.Replace) != "null" ||
		selected.nonordinaryRole() || len(selected.GoFiles) == 0 || len(selected.SFiles) == 0 {
		return nil, goAMD64FrameContext("capture needs complete ordinary original module sources")
	}
	if selected.Module.GoMod != filepath.Join(request.Directory, "go.mod") {
		return nil, goAMD64FrameContext("actual module metadata is outside the original source directory")
	}
	if source.downloadedModule != nil {
		original := source.downloadedModule
		if selected.Module.Path != original.ModulePath || selected.Module.GoVersion != original.GoVersion ||
			selected.Dir != original.Root || len(selected.Imports) != 0 || len(selected.ImportMap) != 0 {
			return nil, goAMD64FrameContext("initial exact downloaded capture requires its import-free primary module root")
		}
	}
	source.imports, err = source.loadStandardDeclarationImports(ctx)
	if err != nil {
		return nil, err
	}
	source.declarations, err = source.typedDeclarations(pkg, exact != nil)
	if err != nil {
		return nil, err
	}
	source.owned, err = os.MkdirTemp(provided["TMPDIR"], "plan9asm-go-frame-object-")
	if err != nil {
		return nil, err
	}
	defer func() {
		if captureErr != nil {
			captureErr = errors.Join(captureErr, source.Close())
		}
	}()
	source.ownedInfo, err = os.Lstat(source.owned)
	if err != nil {
		return nil, err
	}
	canonicalOwned, err := filepath.EvalSymlinks(source.owned)
	if err != nil {
		return nil, err
	}
	source.owned = canonicalOwned
	if err := source.verifyOwned(); err != nil {
		return nil, err
	}
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
		assemblyInput := name
		if source.numericalBinding != nil {
			assemblyInput = filepath.Join(request.Directory, name)
		}
		listing, err := source.run(ctx, "tool", "asm", "-S", "-p", pkg.Path, "-I", request.Directory, "-I", filepath.Join(source.env["GOROOT"], "pkg", "include"), "-o", object, assemblyInput)
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
		var texts []gostaticstorage.TextRecord
		if source.numericalBinding != nil {
			texts, err = source.numericalFrameObject(object, file)
			if err != nil {
				return nil, err
			}
		}
		functions, err := goAMD64FrameListingWithNumericalObject(listing, file, pkg, source.origins, texts)
		if err != nil {
			return nil, err
		}
		if source.numericalBinding != nil && !bytes.Equal(listing, source.numericalBinding.source.native.OriginalListing) {
			return nil, goAMD64FrameContext("independent complete original numerical/frame listing differs")
		}
		for _, text := range texts {
			if !bytes.Equal(functions[text.Name].code, text.Code) {
				return nil, goAMD64FrameContext("independent complete object/listing code bytes differ")
			}
		}
		if err := source.verifyObjectListing(ctx, object, functions); err != nil {
			return nil, err
		}
		for symbol, fn := range functions {
			if _, duplicate := source.functions[symbol]; duplicate {
				return nil, goAMD64FrameContext("duplicate original TEXT symbol")
			}
			declaration, exists := source.declarations[symbol]
			if !exists || !reflect.DeepEqual(fn.signature, declaration.signature) {
				return nil, goAMD64FrameContext("object lacks its independently checked declaration/layout")
			}
			fn.transport, fn.languageType = declaration.transport, declaration.languageType
			source.functions[symbol] = fn
		}
	}
	if !found {
		return nil, goAMD64FrameContext("requested assembly is absent from actual complete selection")
	}
	source.ready = true
	source.numericalCapturing = false
	if err := source.Verify(ctx); err != nil {
		return nil, err
	}
	return source, nil
}

// Verify rechecks actual routing/selection and all original/tool/export/artifact
// bytes. Failure permanently revokes this observation; it grants no memory.
func (source *GoAMD64FrameSource) Verify(ctx context.Context) (verifyErr error) {
	if source == nil || source.identity != source {
		return goAMD64FrameContext("missing sealed actual-object observation")
	}
	if err := source.verifyNumericalFrameInputs(ctx); err != nil {
		source.revokeNumericalFrame()
		return err
	}
	defer func() {
		if err := source.verifyNumericalFrameInputs(ctx); err != nil {
			source.revokeNumericalFrame()
			// The inner observer already revokes on failure. Its own closed
			// generation must not relabel an actual command/cancellation
			// failure as a source exclusion during the outer postcheck.
			if verifyErr == nil || !errors.Is(err, ErrProbeNeedsContext) {
				verifyErr = errors.Join(verifyErr, err)
			}
		}
	}()
	return source.verifyFrameOnly(ctx)
}

func (source *GoAMD64FrameSource) verifyFrameOnly(ctx context.Context) (verifyErr error) {
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
	if ctx != nil && ctx.Err() != nil {
		if source.numericalOrigin != nil {
			return fmt.Errorf("verify numerical frame observer context: %w", ctx.Err())
		}
		return errors.Join(goAMD64FrameContext("expired actual-object observation"), ctx.Err())
	}
	if !source.ready || ctx == nil {
		return goAMD64FrameContext("closed, uninitialized or expired actual-object observation")
	}
	if err := source.verifyDownloadedOrigin(); err != nil {
		return err
	}
	before := source.env
	if err := source.observe(ctx); err != nil {
		return err
	}
	if !reflect.DeepEqual(before, source.env) {
		return goAMD64FrameContext("actual Go environment changed")
	}
	if source.downloadedOrigin == nil {
		if err := source.verifyOwnedSourceDirectory(); err != nil {
			return err
		}
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
	imports, err := source.loadStandardDeclarationImports(ctx)
	if err != nil && source.numericalOrigin != nil {
		return err
	}
	if err != nil || !reflect.DeepEqual(source.imports, imports) {
		return goAMD64FrameContext("actual complete declaration import graph changed: %v", err)
	}
	if err := ctx.Err(); err != nil {
		return errors.Join(goAMD64FrameContext("verification context expired after actual commands"), err)
	}
	return errors.Join(source.verifyInputs(), source.verifyDownloadedOrigin())
}

func (source *GoAMD64FrameSource) verifyOwnedSourceDirectory() error {
	if err := gostaticstorage.ValidateOwnedSourceDirectory(source.request.Directory, source.env["GOROOT"], source.env["GOMODCACHE"]); err != nil {
		return errors.Join(goAMD64FrameContext("owned source needs a distinct physical SDK/cache role"), err)
	}
	return nil
}

func (source *GoAMD64FrameSource) verifyInputs() error {
	if err := source.verifyOwned(); err != nil {
		return err
	}
	entries, directoryInfo, err := gooriginalsrc.ReadDirectory(source.request.Directory, 256)
	if err != nil || source.directoryInfo == nil || !os.SameFile(source.directoryInfo, directoryInfo) {
		return goAMD64FrameContext("original package directory inventory or identity changed: %v", err)
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

func (source *GoAMD64FrameSource) verifyOwned() error {
	info, err := os.Lstat(source.owned)
	if err != nil || source.ownedInfo == nil || !info.IsDir() || !os.SameFile(source.ownedInfo, info) {
		return goAMD64FrameContext("factory artifact directory identity changed or is inaccessible: %v", err)
	}
	return nil
}

// Close invalidates the token before removing only its factory-owned artifacts.
// Calling it again is harmless; no source, tool or cache-owned file is removed.
func (source *GoAMD64FrameSource) Close() error {
	if source == nil {
		return nil
	}
	if source.identity != source {
		return goAMD64FrameContext("copied or uninitialized observation cannot dispose original artifacts")
	}
	source.mu.Lock()
	defer source.mu.Unlock()
	source.ready = false
	if source.numericalOrigin != nil && source.numericalOrigin != source {
		return goAMD64FrameContext("copied numerical frame cannot close the original artifact owner")
	}
	if source.owned == "" {
		return nil
	}
	if err := source.verifyOwned(); err != nil {
		return err
	}
	err := os.RemoveAll(source.owned)
	if err == nil {
		source.owned = ""
		source.ownedInfo = nil
	}
	return err
}

func goAMD64FrameContext(format string, args ...any) error {
	return fmt.Errorf("%w: actual Go frame source: %s", ErrProbeNeedsContext, fmt.Sprintf(format, args...))
}

func (source *GoAMD64FrameSource) run(ctx context.Context, args ...string) ([]byte, error) {
	if ctx == nil {
		return nil, goAMD64FrameContext("actual command requires live context")
	}
	if err := ctx.Err(); err != nil {
		return nil, errors.Join(goAMD64FrameContext("actual command context expired"), err)
	}
	if err := source.verifyDownloadedOrigin(); err != nil {
		return nil, err
	}
	// Capture has not published the source and holds no Frame lock. Verify's
	// existing locked command path is guarded outside the entire round instead.
	if source.numericalCapturing {
		if err := source.numericalBinding.verify(); err != nil {
			return nil, err
		}
	}
	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()
	output, err := goasmcmd.Run(ctx, source.request.Directory, source.request.Environment, source.request.GoBinary, args...)
	if source.numericalCapturing {
		err = errors.Join(err, source.numericalBinding.verify())
	}
	// Preserve both independent origins and cancellation when commands race
	// revocation. Neither observation grants the other's source permission.
	return output, errors.Join(err, ctx.Err(), source.verifyDownloadedOrigin())
}

func goAMD64FrameRead(name string) (_ []byte, readErr error) {
	resolved, err := filepath.EvalSymlinks(name)
	if err != nil || resolved != name {
		return nil, goAMD64FrameContext("input path is redirected or inaccessible: %s", filepath.Base(name))
	}
	before, err := os.Lstat(name)
	if err != nil || !before.Mode().IsRegular() {
		return nil, goAMD64FrameContext("input role changed before opening")
	}
	file, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer func() { readErr = errors.Join(readErr, file.Close()) }()
	info, err := file.Stat()
	if err != nil || !info.Mode().IsRegular() || !os.SameFile(before, info) || info.Size() > 64<<20 {
		return nil, goAMD64FrameContext("input is not a bounded regular file")
	}
	data, err := io.ReadAll(io.LimitReader(file, 64<<20+1))
	if err != nil || len(data) > 64<<20 {
		return nil, goAMD64FrameContext("input byte bound/read failure: %v", err)
	}
	after, err := os.Lstat(name)
	if err != nil || !os.SameFile(info, after) || !after.Mode().IsRegular() || int64(len(data)) != info.Size() {
		return nil, goAMD64FrameContext("input role/extent changed while reading")
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
	entries, directoryInfo, err := gooriginalsrc.ReadDirectory(resolved, 256)
	if err != nil || len(entries) == 0 || len(entries) > 256 {
		return goAMD64FrameContext("package directory is unavailable or exceeds its bound")
	}
	source.directoryInfo = directoryInfo
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
	selected, _, err := source.loadWithRaw(ctx)
	return selected, err
}

func (source *GoAMD64FrameSource) loadWithRaw(ctx context.Context) (goAMD64FrameSelection, []byte, error) {
	var selected goAMD64FrameSelection
	data, err := source.run(ctx, "list", "-json", "-compiled", "-export", "-mod=readonly", ".")
	if err != nil {
		return selected, data, err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(&selected); err != nil {
		return selected, data, err
	}
	var extra any
	if decoder.Decode(&extra) != io.EOF || len(selected.Error) != 0 || len(selected.DepsErrors) != 0 {
		return selected, data, goAMD64FrameContext("incomplete or ambiguous actual package export")
	}
	if selected.Export == "" {
		return selected, data, goAMD64FrameContext("actual package has no compiled export")
	}
	selected.Export, err = filepath.EvalSymlinks(selected.Export)
	cache, cacheErr := filepath.EvalSymlinks(source.env["GOCACHE"])
	if err != nil || cacheErr != nil || !goAssemblyPathWithinRoot(cache, selected.Export) {
		return selected, data, goAMD64FrameContext("compiled export is outside the actual build cache")
	}
	if export, err := source.read(selected.Export); err != nil || len(export) == 0 {
		if err == nil {
			err = goAMD64FrameContext("actual package export is empty")
		}
		return selected, data, err
	}
	return selected, data, nil
}

func (source *GoAMD64FrameSource) numericDeclarations(pkg GoPackage) error {
	_, err := source.typedDeclarations(pkg, true)
	return err
}

// The numeric-only wrapper is retained for fixed-local consumers. Expanding
// capture's observation domain must not expand that consumer's admission.
func (source *GoAMD64FrameSource) typedDeclarations(pkg GoPackage, numericOnly bool) (map[string]goAMD64FrameDeclaration, error) {
	return source.typedDeclarationsForArch(pkg, numericOnly, "amd64")
}

// The numerical SB binding also observes 386 declarations. This helper only
// rechecks language/ABI0 identity; it does not extend the AMD64 frame factory.
func (source *GoAMD64FrameSource) typedDeclarationsForArch(pkg GoPackage, numericOnly bool, arch string) (map[string]goAMD64FrameDeclaration, error) {
	fset := token.NewFileSet()
	var syntax []*ast.File
	var compiled []string
	bodyless := make(map[string]bool)
	for _, name := range source.selection.GoFiles {
		filename := filepath.Join(source.request.Directory, name)
		data, err := source.read(filename)
		if err != nil {
			return nil, err
		}
		file, err := parser.ParseFile(fset, filename, data, parser.ParseComments)
		if err != nil {
			return nil, err
		}
		for _, declaration := range file.Decls {
			switch declaration := declaration.(type) {
			case *ast.FuncDecl:
				if declaration.Body == nil && declaration.Recv == nil {
					bodyless[declaration.Name.Name] = true
				}
			case *ast.GenDecl:
				if declaration.Tok != token.TYPE && declaration.Tok != token.IMPORT &&
					declaration.Tok != token.CONST && declaration.Tok != token.VAR {
					return nil, goAMD64FrameContext("unknown original declaration role")
				}
			default:
				return nil, goAMD64FrameContext("unknown original declaration role")
			}
		}
		syntax = append(syntax, file)
		compiled = append(compiled, filename)
	}
	if len(pkg.Syntax) != len(syntax) {
		return nil, goAMD64FrameContext("supplied syntax omitted complete ordinary source files")
	}
	if len(bodyless) == 0 {
		return nil, goAMD64FrameContext("original source has no bodyless assembly declaration")
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
		return nil, goAMD64FrameContext("effective/generated compiled source mapping is unbound")
	}
	sizes := types.SizesFor("gc", arch)
	actualImporter := importer.ForCompiler(fset, "gc", func(path string) (io.ReadCloser, error) {
		if target, mapped := source.selection.ImportMap[path]; mapped {
			path = target
		}
		dependency, present := source.imports[path]
		if !present || dependency.Export == "" {
			return nil, goAMD64FrameContext("missing captured actual compiler export: %s", path)
		}
		data, err := source.read(dependency.Export)
		if err != nil {
			return nil, err
		}
		return io.NopCloser(bytes.NewReader(data)), nil
	})
	typed, err := (&types.Config{Sizes: sizes, Importer: actualImporter}).Check(pkg.Path, fset, syntax, nil)
	if err != nil {
		return nil, err
	}
	if !reflect.DeepEqual(typed.Scope().Names(), pkg.Types.Scope().Names()) {
		return nil, goAMD64FrameContext("actual selected declarations differ from supplied types")
	}
	declarations := make(map[string]goAMD64FrameDeclaration)
	for _, name := range typed.Scope().Names() {
		if named, ok := typed.Scope().Lookup(name).(*types.TypeName); ok {
			previous, supplied := pkg.Types.Scope().Lookup(name).(*types.TypeName)
			actual, err := goAMD64FrameLanguageType(named.Type(), sizes)
			prior, priorErr := "", error(nil)
			if supplied {
				prior, priorErr = goAMD64FrameLanguageType(previous.Type(), sizes)
			}
			if !supplied || named.IsAlias() != previous.IsAlias() || err != nil || priorErr != nil || actual != prior {
				return nil, goAMD64FrameContext("actual local type declaration differs: %s", name)
			}
			continue
		}
		// Bodies and values are independently typechecked, but only original
		// bodyless assembly declarations enter the frame observation table.
		if !bodyless[name] {
			continue
		}
		fn, ok := typed.Scope().Lookup(name).(*types.Func)
		previous, supplied := pkg.Types.Scope().Lookup(name).(*types.Func)
		if !ok || !supplied {
			return nil, goAMD64FrameContext("missing actual typed declaration")
		}
		signature := fn.Type().(*types.Signature)
		priorSignature, valid := previous.Type().(*types.Signature)
		if !valid || priorSignature == nil || signature.Recv() != nil || signature.Variadic() || signature.TypeParams().Len() != 0 ||
			priorSignature.Recv() != nil || priorSignature.Variadic() || priorSignature.TypeParams().Len() != 0 {
			return nil, goAMD64FrameContext("custom, generic or variadic declaration transport")
		}
		role := goAMD64FrameNumericTransport
		var language strings.Builder
		for _, tuple := range []*types.Tuple{signature.Params(), signature.Results()} {
			if tuple.Len() > 4096 {
				return nil, goAMD64FrameContext("declaration tuple bound")
			}
			language.WriteString("[")
			for i := 0; i < tuple.Len(); i++ {
				typ := tuple.At(i).Type()
				switch underlying := typ.Underlying().(type) {
				case *types.Basic:
					if underlying.Info()&types.IsInteger == 0 {
						return nil, goAMD64FrameContext("observation excludes noninteger scalar transport")
					}
				case *types.Pointer, *types.Slice:
					role = goAMD64FrameCompositeTransport
				default:
					return nil, goAMD64FrameContext("by-value aggregate transport needs a separate observation contract")
				}
				identity, err := goAMD64FrameLanguageType(typ, sizes)
				if err != nil {
					return nil, err
				}
				fmt.Fprintf(&language, "%q:%s;", tuple.At(i).Name(), identity)
				if language.Len() > 64<<10 {
					return nil, goAMD64FrameContext("declaration language byte bound")
				}
			}
			language.WriteString("]")
		}
		if numericOnly && role != goAMD64FrameNumericTransport {
			return nil, goAMD64FrameContext("fixed-local consumer requires numeric transport, not pointer/slice observation")
		}
		// Named types from two independent typechecks have different object
		// identities. Compare their complete language/layout trees, not Ptr or
		// qualified spelling alone; underlying and pointee changes must fail.
		priorLanguage := strings.Builder{}
		for _, tuple := range []*types.Tuple{priorSignature.Params(), priorSignature.Results()} {
			if tuple.Len() > 4096 {
				return nil, goAMD64FrameContext("supplied declaration tuple bound")
			}
			priorLanguage.WriteString("[")
			for i := 0; i < tuple.Len(); i++ {
				identity, err := goAMD64FrameLanguageType(tuple.At(i).Type(), sizes)
				if err != nil {
					return nil, err
				}
				fmt.Fprintf(&priorLanguage, "%q:%s;", tuple.At(i).Name(), identity)
				if priorLanguage.Len() > 64<<10 {
					return nil, goAMD64FrameContext("supplied declaration language byte bound")
				}
			}
			priorLanguage.WriteString("]")
		}
		if language.String() != priorLanguage.String() {
			return nil, goAMD64FrameContext("actual language type/pointee/field identity mismatch: %s", name)
		}
		actual, err := DeriveGoFuncABI0Signature(pkg.Path+"."+name, fn, arch, sizes)
		prior, priorErr := DeriveGoFuncABI0Signature(pkg.Path+"."+name, previous, arch, sizes)
		if err != nil || priorErr != nil || !reflect.DeepEqual(actual, prior) {
			return nil, goAMD64FrameContext("actual typed declaration/layout mismatch: %s", name)
		}
		declarations[actual.Name] = goAMD64FrameDeclaration{signature: actual, transport: role, languageType: language.String()}
	}
	return declarations, nil
}

// A bounded semantic language tree distinguishes equal-sized LLVM pointers,
// named-underlying changes and aggregate element/field models. Sizes belongs
// to the independent go/types model, not an observation of native pointee
// padding (Go1.20 StdSizes differs from its compiler). Only pointer/slice ABI0
// slots are bound to the actual object. No pointee extent, alias or GC grant.
func goAMD64FrameLanguageType(typ types.Type, sizes types.Sizes) (string, error) {
	if sizes == nil {
		return "", goAMD64FrameContext("missing target language layout")
	}
	var output strings.Builder
	active := make(map[types.Type]bool)
	nodes := 0
	var appendType func(types.Type, int) error
	appendType = func(typ types.Type, depth int) error {
		nodes++
		if typ == nil || depth > 32 || nodes > 4096 || active[typ] || output.Len() > 64<<10 {
			return goAMD64FrameContext("recursive or unbounded language type")
		}
		active[typ] = true
		defer delete(active, typ)
		// Check the language role before invoking Sizes: gc Sizes may panic
		// for type parameters. An unknown or generic role is not a layout.
		switch concrete := typ.(type) {
		case *types.Basic:
			if concrete.Info()&types.IsInteger == 0 || concrete.Info()&types.IsUntyped != 0 {
				return goAMD64FrameContext("unsupported noninteger or untyped language type")
			}
		case *types.Named:
			if concrete.TypeParams().Len() != 0 {
				return goAMD64FrameContext("generic language type")
			}
		case *types.Pointer, *types.Slice, *types.Array, *types.Struct:
		default:
			if !goAMD64FrameIsAlias(typ) {
				return goAMD64FrameContext("unknown language type role")
			}
			if method := reflect.ValueOf(typ).MethodByName("TypeParams"); method.IsValid() {
				parameters, ok := method.Call(nil)[0].Interface().(*types.TypeParamList)
				if !ok || parameters.Len() != 0 {
					return goAMD64FrameContext("generic alias language type")
				}
			}
		}
		// Composite identity is written below one node at a time. Asking
		// TypeString to expand an anonymous DAG defeats the node/byte bound.
		spelling := ""
		switch typ.(type) {
		case *types.Basic, *types.Named:
			spelling = types.TypeString(typ, func(pkg *types.Package) string { return pkg.Path() })
		default:
			if goAMD64FrameIsAlias(typ) {
				spelling = types.TypeString(typ, func(pkg *types.Package) string { return pkg.Path() })
			}
		}
		fmt.Fprintf(&output, "(%q:", spelling)
		switch typ := typ.(type) {
		case *types.Basic:
			if typ.Info()&types.IsInteger == 0 {
				return goAMD64FrameContext("unsupported noninteger pointee/element")
			}
			fmt.Fprintf(&output, "integer%d", typ.Kind())
		case *types.Named:
			if typ.TypeParams().Len() != 0 {
				return goAMD64FrameContext("generic language type")
			}
			output.WriteString("named:")
			if err := appendType(typ.Underlying(), depth+1); err != nil {
				return err
			}
		case *types.Pointer:
			output.WriteString("pointer:")
			if err := appendType(typ.Elem(), depth+1); err != nil {
				return err
			}
		case *types.Slice:
			output.WriteString("slice:")
			if err := appendType(typ.Elem(), depth+1); err != nil {
				return err
			}
		case *types.Array:
			fmt.Fprintf(&output, "array%d:", typ.Len())
			if err := appendType(typ.Elem(), depth+1); err != nil {
				return err
			}
			// Go1.20's Sizes can wrap a large array to zero. Its recorded
			// result is not evidence of a representable concrete layout.
			element, alignment := sizes.Sizeof(typ.Elem()), sizes.Alignof(typ.Elem())
			const maxLayout = int64(1<<63 - 1)
			if typ.Len() < 0 || element < 0 || alignment <= 0 || element > maxLayout-(alignment-1) {
				return goAMD64FrameContext("array element layout is unrepresentable")
			}
			stride := (element + alignment - 1) / alignment * alignment
			if typ.Len() > 0 && stride != 0 && typ.Len()-1 > (maxLayout-element)/stride {
				return goAMD64FrameContext("array language layout overflow")
			}
		case *types.Struct:
			if typ.NumFields() > 4096-nodes {
				return goAMD64FrameContext("language field bound")
			}
			fields := make([]*types.Var, typ.NumFields())
			for i := range fields {
				fields[i] = typ.Field(i)
			}
			// Validate every descendant before Sizes traverses the aggregate:
			// generic/incomplete elements are not a concrete layout.
			for _, field := range fields {
				if err := appendType(field.Type(), depth+1); err != nil {
					return err
				}
			}
			offsets := sizes.Offsetsof(fields)
			output.WriteString("struct:")
			for i, field := range fields {
				pkg := ""
				if field.Pkg() != nil {
					pkg = field.Pkg().Path()
				}
				width := sizes.Sizeof(field.Type())
				if offsets[i] < 0 || width < 0 || offsets[i] > (1<<63-1)-width {
					return goAMD64FrameContext("struct language layout overflow")
				}
				fmt.Fprintf(&output, "%q/%q/%q/%t/%d:", pkg, field.Name(), typ.Tag(i), field.Embedded(), offsets[i])
				if output.Len() > 64<<10 {
					return goAMD64FrameContext("language field metadata byte bound")
				}
			}
		default:
			// Alias is newer than Go1.20. Its actual spelling and the local
			// TypeName.IsAlias role are retained alongside the complete RHS.
			// Underlying alone would lose a named RHS's language identity.
			if !goAMD64FrameIsAlias(typ) {
				return goAMD64FrameContext("unknown language type role")
			}
			method := reflect.ValueOf(typ).MethodByName("Rhs")
			if !method.IsValid() || method.Type().NumIn() != 0 || method.Type().NumOut() != 1 {
				return goAMD64FrameContext("unknown alias RHS contract")
			}
			rhs, ok := method.Call(nil)[0].Interface().(types.Type)
			if !ok {
				return goAMD64FrameContext("missing alias RHS language type")
			}
			output.WriteString("alias:")
			if err := appendType(rhs, depth+1); err != nil {
				return err
			}
		}
		size, align := sizes.Sizeof(typ), sizes.Alignof(typ)
		if size < 0 || align <= 0 {
			return goAMD64FrameContext("language type lacks a concrete target layout")
		}
		fmt.Fprintf(&output, ";model_size=%d,model_align=%d)", size, align)
		return nil
	}
	if err := appendType(typ, 0); err != nil {
		return "", err
	}
	if output.Len() > 64<<10 {
		return "", goAMD64FrameContext("language type byte bound")
	}
	return output.String(), nil
}

func goAMD64FrameIsAlias(typ types.Type) bool {
	kind := reflect.TypeOf(typ)
	return kind != nil && kind.Kind() == reflect.Pointer && kind.Elem().PkgPath() == "go/types" && kind.Elem().Name() == "Alias"
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
	return goAMD64FrameListingWithNumericalObject(listing, file, pkg, origins, nil)
}

func goAMD64FrameListingWithNumericalObject(listing []byte, file *File, pkg GoPackage, origins map[string]int, texts []gostaticstorage.TextRecord) (map[string]goAMD64FrameFunction, error) {
	functions := make(map[string]goAMD64FrameFunction)
	scanner := bufio.NewScanner(bytes.NewReader(listing))
	scanner.Buffer(nil, 4<<20)
	name := ""
	var expected int64
	textIndex, relocationIndex := -1, 0
	finishRelocations := func() error {
		if texts != nil && textIndex >= 0 && relocationIndex != len(texts[textIndex].Relocations) {
			return goAMD64FrameContext("missing complete original object/listing relocations")
		}
		return nil
	}
	for scanner.Scan() {
		line := scanner.Text()
		if header := goAMD64FrameHeader.FindStringSubmatch(line); header != nil {
			if err := finishRelocations(); err != nil {
				return nil, err
			}
			if texts != nil {
				textIndex++
				relocationIndex = 0
				if textIndex >= len(texts) || texts[textIndex].Name != header[1] {
					return nil, goAMD64FrameContext("reordered or foreign complete object/listing TEXT")
				}
			}
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
		if strings.HasPrefix(strings.TrimSpace(line), "rel ") && texts != nil {
			relocations := texts[textIndex].Relocations
			if relocationIndex >= len(relocations) || strings.Join(strings.Fields(line), " ") != goAMD64FrameNumericalRelocationLine(relocations[relocationIndex]) {
				return nil, goAMD64FrameContext("foreign, duplicate or reordered actual object/listing relocation")
			}
			relocationIndex++
			continue
		}
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
	if err := finishRelocations(); err != nil {
		return nil, err
	}
	if texts != nil && textIndex+1 != len(texts) {
		return nil, goAMD64FrameContext("missing complete original object/listing TEXT")
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
		var ins x86asm.Inst
		if _, length, aes, err := decodedX86AESInstruction(fn.code[pc:], 64); aes {
			if err != nil || length <= 0 {
				return goAMD64FrameContext("unbound original AES machine instruction")
			}
			// This observer needs exact instruction boundaries, not an AES
			// semantic substitute. The pinned generic decoder predates VAES.
			// Op remains zero: it cannot masquerade as a BP/SP or RET opcode.
			ins.Len = length
		} else {
			var err error
			ins, err = x86asm.Decode(fn.code[pc:], 64)
			if err != nil || ins.Len == 0 || ins.Op == 0 || ins.Op == x86asm.CALL || ins.Op == x86asm.LCALL {
				return goAMD64FrameContext("unknown or calling original machine instruction")
			}
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
