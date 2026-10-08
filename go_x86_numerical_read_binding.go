package plan9asm

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"go/format"
	"go/parser"
	"go/token"
	"go/types"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/xgo-dev/plan9asm/internal/gostaticstorage"
)

// GoX86NumericalReadBinding is a live factory token, not a serializable report.
// Its capture context and every source/storage owner must remain live until
// translation completes. Failure/Close revokes the originating source.
type GoX86NumericalReadBinding struct {
	source          *GoX86NumericalReadSource
	originalSource  *GoX86NumericalReadSource
	context         context.Context
	originalContext context.Context
	packageInput    GoPackage
	originalTypes   *types.Package
	originalImports map[string]*types.Package
	file            [32]byte
	options         Options
	symbols         map[string]string
	declarations    map[string]goAMD64FrameDeclaration
	reads           []numericalSBReadFact
	capabilities    map[string]numericalSignatureCapabilities
	seal            [32]byte
}

type numericalSignatureCapabilities struct {
	caller *x86GoCallerRegionProof
	local  *x86FixedLocalBinding
}

func numericalCapabilities(sigs map[string]FuncSig) map[string]numericalSignatureCapabilities {
	result := make(map[string]numericalSignatureCapabilities, len(sigs))
	for name, sig := range sigs {
		result[name] = numericalSignatureCapabilities{sig.x86CallerRegions, sig.x86FixedLocal}
	}
	return result
}

// BindNumericalReads joins complete caller declarations/options to the sealed
// original source and actual native reads. It returns independent Options;
// declarations alone, JSON, pointer types and a numerical owner cannot bind.
func (source *GoX86NumericalReadSource) BindNumericalReads(ctx context.Context, pkg GoPackage, file *File, options Options) (Options, error) {
	if err := source.Verify(ctx); err != nil {
		return Options{}, err
	}
	if file == nil || options.ResolveSym == nil || pkg.Types == nil || pkg.Path != source.request.PackagePath || pkg.Types.Path() != pkg.Path || pkg.SourceSelection != nil {
		return Options{}, numericalSBContext("binding lacks complete original declarations/source/resolver")
	}
	arch := strings.TrimPrefix(source.request.Target, "linux/")
	triple := map[string]string{"386": "i386-unknown-linux-gnu", "amd64": "x86_64-unknown-linux-gnu"}[arch]
	if options.Goarch != arch || options.TargetTriple != triple || options.WASMABI != WASMABIDirect || len(options.X86TailGroups) != 0 || options.NumericalReads != nil {
		return Options{}, numericalSBContext("binding profile or existing authority differs")
	}
	provided, err := x86CallerSourceFingerprint(file)
	if err != nil {
		return Options{}, err
	}
	actual, err := x86CallerSourceFingerprint(source.file)
	if err != nil || provided != actual || len(file.Funcs) != len(options.Sigs) {
		return Options{}, numericalSBContext("binding requires the exact complete original parsed file")
	}
	// Resolve outside observer locks, then verify the complete live source again.
	symbols := make(map[string]string)
	for _, fn := range file.Funcs {
		symbols[fn.Sym] = options.ResolveSym(fn.Sym)
		if symbols[fn.Sym] != numericalNativeName(strings.TrimSuffix(fn.Sym, "(SB)"), pkg.Path) {
			return Options{}, numericalSBContext("function resolver differs from actual ABI0 symbol")
		}
	}
	for _, read := range source.reads {
		base, _, valid := parseSBRef(read.original.Args[read.operand].Sym)
		if !valid || options.ResolveSym(base) != read.variable.NativeName {
			return Options{}, numericalSBContext("storage resolver differs from actual typed data symbol")
		}
		symbols[base] = read.variable.NativeName
		if _, collision := options.Sigs[read.variable.NativeName]; collision {
			return Options{}, numericalSBContext("typed data symbol collides with a function")
		}
	}
	if err := source.Verify(ctx); err != nil {
		return Options{}, err
	}
	provided, err = x86CallerSourceFingerprint(file)
	if err != nil || provided != actual {
		return Options{}, numericalSBContext("binding resolver changed the complete original file")
	}
	source.mu.Lock()
	declarations, err := source.numericalDeclarations(ctx, pkg, arch)
	if err == nil {
		err = numericalDeclarationSyntax(source, pkg)
	}
	source.mu.Unlock()
	if err != nil {
		return Options{}, err
	}
	sigs, err := numericalCloneSignatures(options.Sigs)
	if err != nil {
		return Options{}, err
	}
	for _, fn := range file.Funcs {
		name := symbols[fn.Sym]
		declaration, found := declarations[name]
		sig := sigs[name]
		if !found || sig.arm64FPReadDeclaration != nil || sig.arm64ABI0Declaration != nil || sig.arm64ABI0TailDeclaration != nil || sig.arm64NativeScratch != nil ||
			!reflect.DeepEqual(numericalSignatureShape(sig), declaration.signature) {
			return Options{}, numericalSBContext("signature is not the actual source declaration: %s", name)
		}
	}
	options.Sigs, options.ResolveSym = sigs, nil
	binding := &GoX86NumericalReadBinding{source: source, originalSource: source, context: ctx, originalContext: ctx, packageInput: pkg,
		originalTypes: pkg.Types, originalImports: numericalImportIdentities(pkg.Imports), file: actual,
		options: options, symbols: symbols, declarations: declarations, reads: source.readFacts(), capabilities: numericalCapabilities(sigs)}
	binding.seal, err = binding.digest()
	if err != nil {
		return Options{}, err
	}
	if err := binding.verify(); err != nil {
		return Options{}, err
	}
	bound := options
	bound.Sigs, err = numericalCloneSignatures(options.Sigs)
	if err != nil {
		return Options{}, err
	}
	bound.ResolveSym = binding.resolve
	bound.NumericalReads = binding
	return bound, nil
}

func numericalImportIdentities(imports map[string]*types.Package) map[string]*types.Package {
	if imports == nil {
		return nil
	}
	identities := make(map[string]*types.Package, len(imports))
	for name, pkg := range imports {
		identities[name] = pkg
	}
	return identities
}

func numericalSameImports(imports, expected map[string]*types.Package) bool {
	if (imports == nil) != (expected == nil) || len(imports) != len(expected) {
		return false
	}
	for name, pkg := range imports {
		if original, found := expected[name]; !found || pkg != original {
			return false
		}
	}
	return true
}

func numericalSignatureShape(sig FuncSig) FuncSig {
	sig.x86CallerRegions, sig.x86FixedLocal = nil, nil
	sig.arm64FPReadDeclaration, sig.arm64ABI0Declaration, sig.arm64ABI0TailDeclaration, sig.arm64NativeScratch = nil, nil, nil, nil
	return sig
}

func numericalCloneSignatures(input map[string]FuncSig) (map[string]FuncSig, error) {
	result, err := cloneFixedLocalSigs(input)
	if err != nil {
		return nil, err
	}
	// Preserve separately established private contracts, not reconstructed JSON.
	for name, old := range input {
		sig := result[name]
		sig.x86CallerRegions, sig.x86FixedLocal = old.x86CallerRegions, old.x86FixedLocal
		sig.arm64FPReadDeclaration, sig.arm64ABI0Declaration, sig.arm64ABI0TailDeclaration, sig.arm64NativeScratch = old.arm64FPReadDeclaration, old.arm64ABI0Declaration, old.arm64ABI0TailDeclaration, old.arm64NativeScratch
		result[name] = sig
	}
	return result, nil
}

func numericalDeclarationSyntax(source *GoX86NumericalReadSource, pkg GoPackage) error {
	if len(pkg.Syntax) != len(source.guard.selection.GoFiles) {
		return numericalSBContext("declaration syntax inventory differs")
	}
	wanted := make(map[string]int)
	for _, name := range source.guard.selection.GoFiles {
		data, err := source.guard.read(filepath.Join(source.request.Directory, name))
		if err != nil {
			return err
		}
		actual, err := parser.ParseFile(token.NewFileSet(), name, data, parser.ParseComments)
		if err != nil {
			return numericalSBContext("declaration syntax is missing or unparseable")
		}
		var a bytes.Buffer
		if format.Node(&a, token.NewFileSet(), actual) != nil {
			return numericalSBContext("original declaration syntax cannot be formatted")
		}
		wanted[a.String()]++
	}
	// File order and absolute token.Pos bases are not declaration semantics.
	// Compare the complete multiset, including comments/directives and bodies.
	for _, supplied := range pkg.Syntax {
		var b bytes.Buffer
		if supplied == nil || format.Node(&b, token.NewFileSet(), supplied) != nil || wanted[b.String()] == 0 {
			return numericalSBContext("supplied complete declaration syntax differs from original source")
		}
		wanted[b.String()]--
	}
	return nil
}

func (binding *GoX86NumericalReadBinding) resolve(symbol string) string {
	if name, found := binding.symbols[symbol]; found {
		return name
	}
	return symbol // No address/call role is created for unknown symbols.
}

func (binding *GoX86NumericalReadBinding) digest() ([32]byte, error) {
	data, err := json.Marshal(struct {
		File         [32]byte
		Target, Arch string
		WASM         WASMABI
		X87          X87Mode
		Annotate     bool
		Sigs         map[string]FuncSig
		TailGroups   []X86TailGroup
		Symbols      map[string]string
		Reads        []numericalSBReadFact
	}{binding.file, binding.options.TargetTriple, binding.options.Goarch, binding.options.WASMABI, binding.options.X87Mode,
		binding.options.AnnotateSource, binding.options.Sigs, binding.options.X86TailGroups, binding.symbols, binding.reads})
	return sha256.Sum256(data), err
}

func (binding *GoX86NumericalReadBinding) verify() error {
	if binding == nil || binding.source == nil || binding.source != binding.originalSource {
		return numericalSBContext("missing live factory binding")
	}
	fail := func(err error) error {
		binding.source.mu.Lock()
		binding.source.ready = false
		binding.source.mu.Unlock()
		return err
	}
	if binding.context == nil || reflect.TypeOf(binding.context) != reflect.TypeOf(binding.originalContext) ||
		!reflect.TypeOf(binding.context).Comparable() || binding.context != binding.originalContext {
		return fail(numericalSBContext("binding live context identity changed"))
	}
	if binding.packageInput.Types == nil || binding.packageInput.Types != binding.originalTypes ||
		binding.packageInput.SourceSelection != nil || !numericalSameImports(binding.packageInput.Imports, binding.originalImports) {
		return fail(numericalSBContext("binding private declaration/import identity changed"))
	}
	if err := binding.source.Verify(binding.context); err != nil {
		return err
	}
	sum, err := binding.digest()
	if err != nil || sum != binding.seal || !numericalSameCapabilities(binding.options.Sigs, binding.capabilities) {
		return fail(numericalSBContext("private binding shape changed"))
	}
	binding.source.mu.Lock()
	declarations, err := binding.source.numericalDeclarations(binding.context, binding.packageInput, binding.options.Goarch)
	if err == nil {
		err = numericalDeclarationSyntax(binding.source, binding.packageInput)
	}
	valid := binding.source.ready && reflect.DeepEqual(binding.declarations, declarations) && reflect.DeepEqual(binding.reads, binding.source.readFacts())
	binding.source.mu.Unlock()
	if err != nil {
		return fail(err)
	}
	if !valid {
		return fail(numericalSBContext("declaration/source/owner read join changed"))
	}
	return binding.source.Verify(binding.context)
}

// Imports are obtained only from the same live, already authenticated owner
// selections. A temporary typecheck guard cannot add new files to the original
// source seal or treat the CPU's assembly namespace as a Go import certificate.
func (source *GoX86NumericalReadSource) numericalDeclarations(ctx context.Context, pkg GoPackage, arch string) (map[string]goAMD64FrameDeclaration, error) {
	guard := *source.guard
	guard.inputs = make(map[string][32]byte)
	guard.imports = make(map[string]goAMD64FrameSelection)
	for _, owner := range source.originalOwners {
		report, err := owner.observation.Readback(ctx)
		if err != nil {
			return nil, numericalStorageFailure(ctx, "declaration owner readback", err)
		}
		if report.Protocol == gostaticstorage.SDKCPUProtocol {
			continue
		}
		decoder := json.NewDecoder(bytes.NewReader(report.RawSelection))
		for decoder.More() {
			var dependency goAMD64FrameSelection
			if err := decoder.Decode(&dependency); err != nil {
				return nil, err
			}
			if dependency.Export != "" {
				guard.imports[dependency.ImportPath] = dependency
			}
		}
	}
	return guard.typedDeclarationsForArch(pkg, false, arch)
}

type x86NumericalReadCertificate struct {
	binding     *GoX86NumericalReadBinding
	read        int
	instruction [32]byte
}

func verifyNumericalReadInputs(file *File, options Options) (result error) {
	binding := options.NumericalReads
	if binding == nil {
		return nil
	}
	defer func() {
		if result != nil && binding.source != nil {
			binding.source.mu.Lock()
			binding.source.ready = false
			binding.source.mu.Unlock()
		}
	}()
	if err := binding.verify(); err != nil {
		return err
	}
	check := func() error {
		if file == nil || options.ResolveSym == nil {
			return numericalSBContext("consumer lost complete source/resolver")
		}
		shape := options
		shape.ResolveSym, shape.NumericalReads = nil, nil
		fingerprint, err := x86CallerSourceFingerprint(file)
		if err != nil || fingerprint != binding.file || !reflect.DeepEqual(shape, binding.options) || !numericalSameCapabilities(options.Sigs, binding.capabilities) {
			return numericalSBContext("consumer original file/declarations/options changed")
		}
		return nil
	}
	if err := check(); err != nil {
		return err
	}
	for symbol, expected := range binding.symbols {
		if options.ResolveSym(symbol) != expected {
			return numericalSBContext("consumer resolver changed exact storage/function identity")
		}
	}
	if err := check(); err != nil {
		return err
	}
	return binding.verify() // Resolver callbacks cannot mutate a live source/owner.
}

func freezeNumericalReadInputs(file *File, options Options) (*File, Options, error) {
	if err := verifyNumericalReadInputs(file, options); err != nil {
		return nil, Options{}, err
	}
	binding := options.NumericalReads
	if binding == nil {
		return file, options, nil
	}
	encoded, err := json.Marshal(file)
	if err != nil {
		return nil, Options{}, err
	}
	var frozen File
	if err := json.Unmarshal(encoded, &frozen); err != nil {
		return nil, Options{}, err
	}
	options = binding.options
	options.Sigs, err = numericalCloneSignatures(binding.options.Sigs)
	if err != nil {
		return nil, Options{}, err
	}
	options.ResolveSym, options.NumericalReads = binding.resolve, binding
	for i, read := range binding.reads {
		found := false
		for f := range frozen.Funcs {
			fn := &frozen.Funcs[f]
			if binding.resolve(fn.Sym) != read.Function || read.Instruction < 0 || read.Instruction >= len(fn.Instrs) {
				continue
			}
			ins := &fn.Instrs[read.Instruction]
			if !reflect.DeepEqual(*ins, read.Original) || ins.x86NumericalRead != nil {
				return nil, Options{}, numericalSBContext("original read ordinal/instruction differs")
			}
			digest, err := fixedLocalInstructionDigest(*ins)
			if err != nil {
				return nil, Options{}, err
			}
			ins.x86NumericalRead = &x86NumericalReadCertificate{binding, i, digest}
			found = true
		}
		if !found {
			return nil, Options{}, numericalSBContext("original read has no unique complete source function")
		}
	}
	return &frozen, options, verifyNormalizedNumericalReads(&frozen, options)
}

func numericalCertifiedRead(ins Instr, arch string) (numericalSBReadFact, x86NumericalSBReadForm, bool) {
	cert := ins.x86NumericalRead
	if cert == nil || cert.binding == nil || cert.binding.source == nil || cert.read < 0 || cert.read >= len(cert.binding.reads) || arch != cert.binding.options.Goarch {
		return numericalSBReadFact{}, x86NumericalSBReadForm{}, false
	}
	read := cert.binding.reads[cert.read]
	digest, err := fixedLocalInstructionDigest(ins)
	form, valid := x86NumericalSBReadProperties(ins, arch)
	if err != nil || digest != cert.instruction || !valid || form.operand != read.Operand || form.spec.sourceBytes != read.Width || !numericalOwnerContains(read.Variable, read.Offset, read.Width) {
		return numericalSBReadFact{}, x86NumericalSBReadForm{}, false
	}
	return read, form, true
}

func verifyNormalizedNumericalReads(file *File, options Options) error {
	if options.NumericalReads == nil {
		return nil
	}
	seen := make(map[int]bool)
	for _, fn := range file.Funcs {
		for ordinal, ins := range fn.Instrs {
			if ins.x86NumericalRead == nil {
				continue
			}
			cert := ins.x86NumericalRead
			read, _, valid := numericalCertifiedRead(ins, options.Goarch)
			if !valid || cert.binding != options.NumericalReads || seen[cert.read] || options.ResolveSym(fn.Sym) != read.Function || ordinal != read.Instruction {
				return numericalSBContext("normalized function/ordinal/operand/native read identity changed")
			}
			seen[cert.read] = true
		}
	}
	if len(seen) != len(options.NumericalReads.reads) || len(file.Data)+len(file.Globl) != 0 {
		return numericalSBContext("normalized read inventory changed or storage collides with local data")
	}
	return nil
}

func numericalSameCapabilities(sigs map[string]FuncSig, expected map[string]numericalSignatureCapabilities) bool {
	if len(sigs) != len(expected) {
		return false
	}
	for name, sig := range sigs {
		cap, found := expected[name]
		if !found || sig.x86CallerRegions != cap.caller || sig.x86FixedLocal != cap.local {
			return false
		}
	}
	return true
}

func finishNumericalReads(file *File, options Options) error {
	if err := verifyNumericalReadInputs(file, options); err != nil {
		return fmt.Errorf("numerical read post-consumption guard: %w", err)
	}
	return nil
}

func verifyGoNumericalReadInputs(binding *GoX86NumericalReadBinding, pkg GoPackage, opt GoModuleOptions) (result error) {
	defer func() {
		if result != nil && binding != nil && binding.source != nil {
			binding.source.mu.Lock()
			binding.source.ready = false
			binding.source.mu.Unlock()
		}
	}()
	if err := binding.verify(); err != nil {
		return err
	}
	if pkg.Path != binding.packageInput.Path || pkg.Types != binding.originalTypes || pkg.SourceSelection != nil ||
		!numericalSameImports(pkg.Imports, binding.originalImports) ||
		opt.GOOS != "linux" || opt.GOARCH != binding.options.Goarch || opt.TargetTriple != binding.options.TargetTriple ||
		opt.WASMABI != binding.options.WASMABI || opt.X87Mode != binding.options.X87Mode || opt.AnnotateSource != binding.options.AnnotateSource ||
		opt.Sizes != nil || opt.FrameSizes != nil ||
		opt.ResolveSym == nil || opt.KeepFunc != nil || opt.ManualSig != nil || len(opt.ARM64ClosureABIs) != 0 ||
		opt.FileName != binding.source.request.AssemblyFile && opt.FileName != filepath.Join(binding.source.request.Directory, binding.source.request.AssemblyFile) {
		return numericalSBContext("Go wrapper declaration/profile/source role differs")
	}
	binding.source.mu.Lock()
	err := numericalDeclarationSyntax(binding.source, pkg)
	binding.source.mu.Unlock()
	if err != nil {
		return err
	}
	for symbol, expected := range binding.symbols {
		if opt.ResolveSym(symbol) != expected {
			return numericalSBContext("Go wrapper resolver changed exact source/storage identity")
		}
	}
	return binding.verify()
}

func numericalGoSignatures(binding *GoX86NumericalReadBinding, supplied map[string]FuncSig) (map[string]FuncSig, error) {
	if len(supplied) != len(binding.options.Sigs) {
		return nil, numericalSBContext("Go wrapper signature inventory differs from sealed declarations")
	}
	for name, signature := range supplied {
		original, found := binding.options.Sigs[name]
		if !found || !reflect.DeepEqual(numericalSignatureShape(signature), numericalSignatureShape(original)) {
			return nil, numericalSBContext("Go wrapper signature differs from sealed original declaration")
		}
	}
	// Only the opaque factory transports separately verified private contracts;
	// a ManualSig callback or a new source selection cannot recreate them.
	return numericalCloneSignatures(binding.options.Sigs)
}
